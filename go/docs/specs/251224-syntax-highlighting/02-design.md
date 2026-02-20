# Design: Syntax Highlighting in Diff View

## Executive Summary

This design adds syntax highlighting to the diff view using Chroma for tokenization and styling. The entire file is tokenized once when loading the diff, then tokens are stored per line. During rendering, we look up colors from Chroma's built-in styles (monokai/github) and apply them via Lip Gloss, giving us control over both foreground (syntax) and background (diff changes) colors.

Key insight: By using Chroma's token output rather than its ANSI output, we avoid conflicts between Chroma's color resets and our diff background colors. Chroma handles lexing and provides color schemes, we handle the actual ANSI rendering.

## Context & Problem Statement

The diff view displays file content without syntax highlighting. All code appears in a single color (gray for unchanged, with subtle green/red backgrounds for changes), making it harder to read and understand code structure.

**Scope:** Syntax highlighting in the diff view only. Does not cover log list or file list views.

## Current State

The diff view renders side-by-side columns showing the entire file:

```
  1   func main() {        │   1   func main() {
  2 -     oldCode()        │
                           │   2 +     newCode()
  3   }                    │   3   }
```

**Data flow:**
1. `MergeFullFile(oldContent, newContent, parsedDiff)` → `FullFileDiff`
2. `FullFileDiff.Lines` contains `FullFileLine` structs with `LeftContent`/`RightContent`
3. `renderFullFileLine()` calls `formatColumnContent()` for each column
4. `formatColumnContent()` truncates content and applies Lip Gloss style

## Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Load Diff                                                      │
│                                                                 │
│  oldContent, newContent ──► MergeFullFile() ──► FullFileDiff    │
│                                                     │           │
│                                                     ▼           │
│                            ApplySyntaxHighlighting(diff, path)  │
│                                                     │           │
│                              ┌──────────────────────┴─────┐     │
│                              ▼                            ▼     │
│                    Tokenize(oldContent)      Tokenize(newContent)
│                              │                            │     │
│                              ▼                            ▼     │
│                    Split tokens by line      Split tokens by line
│                              │                            │     │
│                              └────────────┬───────────────┘     │
│                                           ▼                     │
│                           Store in FullFileLine structs         │
└─────────────────────────────────────────────────────────────────┘
                                            │
                                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  Render (diff_view.go)                                          │
│                                                                 │
│  formatColumnContent()                                          │
│       │                                                         │
│       ├─► Render tokens with syntax styles (foreground)         │
│       ├─► Truncate if needed (track visible width)              │
│       └─► Apply diff background style                           │
└─────────────────────────────────────────────────────────────────┘
```

### Why Per-File Tokenization

Tokenizing line-by-line would fail for multi-line constructs:
- Block comments: `/* ... */` spanning lines
- Multi-line strings: Go backtick strings, Python triple quotes

A line containing `continued comment */` needs context from previous lines. Per-file tokenization provides this.

### Why Token Output (Not ANSI)

Chroma can output ANSI-colored strings directly, but those include reset codes (`\033[0m`) that would clear our diff background colors mid-line. By using token output:
- Chroma handles lexing (the hard part)
- We look up colors from Chroma styles, apply them via Lip Gloss
- Full control over foreground + background together

### Data Model Changes

**Token type:**
```go
// Token represents a syntax-highlighted token
type Token struct {
    Type  chroma.TokenType // e.g., chroma.Keyword, chroma.String, chroma.Text
    Value string           // the actual text
}
```

**Extended FullFileLine:**
```go
type FullFileLine struct {
    LeftLineNo   int
    RightLineNo  int
    LeftTokens   []highlight.Token  // Syntax tokens (always populated)
    RightTokens  []highlight.Token  // Syntax tokens (always populated)
    Change       ChangeType
}
```

> **Decision:** Store only tokens, not plain text. For unsupported languages, tokens contain a single `Text` token per line. This gives one data representation and one rendering path.

### New Package: `internal/highlight`

```go
package highlight

// TokenizeFile tokenizes file content and returns tokens grouped by line.
// Always returns valid tokens - uses Text tokens for unsupported languages.
func TokenizeFile(content, filename string) [][]Token
```

**Internal behavior:**
- Uses `lexers.Match(filename)` for language detection
- Iterates Chroma's token stream, splitting on newlines
- For unrecognized file types: returns `Text` tokens (one per line)

### Syntax Styles

Use Chroma's built-in styles rather than defining our own. The highlight package manages the active style:

```go
package highlight

var activeStyle *chroma.Style

func init() {
    // Select style based on terminal background
    if isDarkTerminal() {
        activeStyle = styles.Get("monokai")
    } else {
        activeStyle = styles.Get("github")
    }
}

// StyleForToken returns a Lip Gloss style for the given token type.
func StyleForToken(tokenType chroma.TokenType) lipgloss.Style {
    entry := activeStyle.Get(tokenType)
    style := lipgloss.NewStyle()
    if entry.Colour.IsSet() {
        style = style.Foreground(lipgloss.Color(entry.Colour.String()))
    }
    return style
}
```

> **Decision:** Use Chroma's built-in styles (monokai for dark, github for light). These are battle-tested color schemes. We can customize later if they don't match Splice's aesthetic.

Chroma styles handle token type hierarchy automatically - if `KeywordDeclaration` has no entry, it inherits from `Keyword`.

### Integration: ApplySyntaxHighlighting

```go
// ApplySyntaxHighlighting adds tokens to a FullFileDiff.
// Called after MergeFullFile(), before rendering.
func ApplySyntaxHighlighting(diff *FullFileDiff, oldContent, newContent, filepath string)
```

**Process:**
1. `highlight.TokenizeFile(oldContent, filepath)` → `[][]Token` (tokens per line)
2. `highlight.TokenizeFile(newContent, filepath)` → `[][]Token`
3. For each `FullFileLine`, set `LeftTokens`/`RightTokens` by line number

### Rendering Changes

Modify `formatColumnContent()` to accept tokens instead of plain string content:

1. Iterate tokens, applying `highlight.StyleForToken(tok.Type)` to each
2. Track visible width, truncate with "…" if exceeds column width
3. Concatenate styled tokens into `renderedContent`
4. Build line: `lineNo + indicator + renderedContent`
5. Wrap with `bgStyle.Render()` for diff background

### Theme Selection

> **Decision:** Use `monokai` for dark terminals, `github` for light terminals. Detect terminal background using Lip Gloss/termenv. These are widely used, well-tested themes.

### Language Detection

Chroma's `lexers.Match(filename)` handles extension mapping:
- `.go` → Go, `.js` → JavaScript, `.py` → Python, etc.
- 200+ languages supported
- Unknown extensions → `Text` tokens (plain rendering)

### Error Handling

The highlight module never causes rendering failures:
- Lexer not found → return `Text` tokens
- Tokenization error → return `Text` tokens

No errors shown for unsupported file types.

## Data Flow Diagram

```
User opens diff for "main.go"
           │
           ▼
┌──────────────────────────────────────────────────────────────────┐
│ loadDiff()                                                       │
│                                                                  │
│   fullDiff = diff.MergeFullFile(oldContent, newContent, parsed)  │
│                       │                                          │
│                       ▼                                          │
│   diff.ApplySyntaxHighlighting(fullDiff, old, new, "main.go")    │
│         │                                                        │
│         ├──► highlight.TokenizeFile(oldContent, "main.go")       │
│         │      └──► [line1: [{chroma.Keyword,"func"},...]]       │
│         │                                                        │
│         └──► Populate LeftTokens/RightTokens per line            │
└──────────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────────┐
│ DiffState.View()                                                 │
│                                                                  │
│   formatColumnContent(tokens=[{chroma.Keyword,"func"},...], ...) │
│         │                                                        │
│         ├──► renderTokens() applies syntax styles (foreground)   │
│         │      └──► highlight.StyleForToken(tok.Type).Render()   │
│         │                                                        │
│         └──► bgStyle.Render() applies diff background            │
│               └──► Green background wraps entire line            │
└──────────────────────────────────────────────────────────────────┘
           │
           ▼
    Terminal: "func" in purple on green background
```

## New Dependencies

```go
import (
    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
)
```

Note: We use Chroma's styles for colors but not its formatters - we handle ANSI rendering ourselves via Lip Gloss.

## Open Questions

None. All design decisions have been made.

**Decisions summary:**
1. **Per-file tokenization** - Preserves context for multi-line constructs
2. **Token output** - We control rendering, avoids ANSI reset conflicts
3. **Data model** - Store `[]Token` per line, no plain text (Text tokens for unsupported languages)
4. **Styles** - Use Chroma's built-in styles (monokai/github), customize later if needed
5. **Integration** - After `MergeFullFile()`, before rendering
