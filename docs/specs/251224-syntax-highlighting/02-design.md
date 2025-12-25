# Design: Syntax Highlighting in Diff View

## Executive Summary

This design adds syntax highlighting to the diff view by integrating the Chroma library for tokenization and ANSI output. Chroma outputs foreground colors for syntax tokens, and Lip Gloss applies background colors for change indicators—these work independently in ANSI, so both display correctly together.

The integration point is `formatColumnContent()` in `diff_view.go`. Before rendering, content is passed through a new highlighter module that uses Chroma with terminal256 output. Language detection uses file extension matching. A new `highlight` package provides a simple API: `Highlight(content, filename) string`.

Light/dark theme adaptation uses two Chroma styles selected based on Lip Gloss's color profile detection, coordinated with existing Splice color palette.

## Context & Problem Statement

The diff view displays file content without syntax highlighting. All code appears in a single color (gray for unchanged, with subtle green/red backgrounds for changes), making it harder to read and understand code structure. Developers expect language-aware coloring similar to their editors.

**Scope:** This design covers syntax highlighting in the diff view only. It does not cover the log list or file list views.

## Current State

The diff view (`internal/ui/states/diff_view.go`) renders side-by-side columns:

```
  1   func main() {        │   1   func main() {
  2 - 	oldCode()          │
                           │   2 + 	newCode()
  3   }                    │   3   }
```

**Rendering flow:**
1. `renderFullFileLine()` calls `formatColumnContent()` for each column
2. `formatColumnContent()` builds string: `lineNo + indicator + content`
3. Entire string receives one Lip Gloss style (background color for changes)

**Current styles** (`internal/ui/styles/styles.go`):
- `DiffAdditionsStyle`: Very subtle green background (`#e8f5e9` light / `#1e3a1e` dark)
- `DiffDeletionsStyle`: Very subtle red background (`#ffebee` light / `#3a1e1e` dark)
- `TimeStyle`: Gray foreground for unchanged lines

## Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        diff_view.go                             │
│                                                                 │
│  formatColumnContent(lineNo, indicator, content, ...)           │
│         │                                                       │
│         ▼                                                       │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │  highlight.Highlight(content, filepath)                  │   │
│  │  └─> Returns ANSI-colored string (foreground only)       │   │
│  └─────────────────────────────────────────────────────────┘   │
│         │                                                       │
│         ▼                                                       │
│  lineNo + indicator + highlightedContent                        │
│         │                                                       │
│         ▼                                                       │
│  style.Render(...)  ← Lip Gloss adds background color           │
│         │                                                       │
│         ▼                                                       │
│  Final output: syntax colors + change background                │
└─────────────────────────────────────────────────────────────────┘
```

### Component Design

#### New Package: `internal/highlight`

A focused module for syntax highlighting with a minimal API:

```go
// Highlight returns the content with ANSI syntax highlighting.
// If the language cannot be detected or highlighting fails, returns content unchanged.
func Highlight(content, filename string) string

// SetDarkMode configures whether to use dark or light theme.
// Should be called once at startup based on terminal background detection.
func SetDarkMode(dark bool)
```

**Internal structure:**
- Uses Chroma's `lexers.Match(filename)` for language detection by extension
- Uses Chroma's `terminal256` formatter for ANSI output
- Caches formatter and style instances (they're reusable)
- Falls back to unmodified content on any error

> **Decision:** Create a new `internal/highlight` package rather than adding Chroma calls directly to `diff_view.go`. This keeps the diff view focused on layout/rendering and makes the highlighting logic testable in isolation.

#### Integration in diff_view.go

Modify `formatColumnContent()` to highlight content before applying background style:

```go
func (s *DiffState) formatColumnContent(lineNo int, indicator, content string,
    lineNoWidth, contentWidth int, style lipgloss.Style) string {

    // ... existing line number formatting ...

    expandedContent := expandTabs(content, 4)
    truncated := truncateWithEllipsis(expandedContent, contentWidth)

    // NEW: Apply syntax highlighting to content only
    highlighted := highlight.Highlight(truncated, s.File.Path)

    // Build column with highlighted content
    columnStr := lineNoStr + " " + indicator + " " + highlighted

    // Lip Gloss background wraps the whole thing
    return style.Render(columnStr)
}
```

> **Decision:** Highlight after truncation, not before. Truncation operates on visible characters and would corrupt ANSI escape sequences if applied to already-highlighted content. Chroma handles short content efficiently.

### Why ANSI Foreground + Lip Gloss Background Works

ANSI escape codes support independent foreground and background colors:
- Chroma outputs: `\033[38;5;123mfunc\033[0m` (foreground color for "func")
- Lip Gloss adds: `\033[48;5;22m...\033[0m` (background color wrapping content)

Terminal displays both: blue "func" text on green background. Lip Gloss is ANSI-aware and preserves existing escape sequences when applying new styles.

### Theme Selection

> **Decision:** Use Chroma's built-in `monokai` (dark) and `github` (light) styles. These are well-tested, widely recognized, and provide good contrast. Custom themes add complexity without clear benefit given the non-goal of user-configurable themes.

**Light/dark detection:**
```go
func init() {
    // Lip Gloss uses termenv which detects terminal background
    profile := lipgloss.DefaultRenderer().ColorProfile()
    // Use light theme if terminal reports light background
    // This integrates with existing Splice adaptive color patterns
}
```

The `SetDarkMode()` function allows explicit override if detection is unreliable.

### Color Palette Considerations

Syntax colors must remain readable on the diff backgrounds:
- Green background (`#1e3a1e` dark / `#e8f5e9` light) — avoid green syntax colors
- Red background (`#3a1e1e` dark / `#ffebee` light) — avoid red syntax colors

Monokai and GitHub styles use blue, purple, orange, and yellow for most tokens, which have good contrast against both backgrounds.

> **Decision:** Accept standard Chroma themes without modification. The diff backgrounds are subtle enough that standard syntax colors remain readable. If specific colors prove problematic in practice, we can create custom Chroma styles in a future iteration.

### Language Detection

Chroma's `lexers.Match(filename)` handles extension mapping:
- `.go` → Go lexer
- `.js`, `.mjs` → JavaScript lexer
- `.ts`, `.tsx` → TypeScript lexer
- `.py` → Python lexer
- `.rs` → Rust lexer
- `.java` → Java lexer
- `.c`, `.h` → C lexer
- `.cpp`, `.hpp`, `.cc` → C++ lexer
- 200+ other languages

**Fallback:** If no lexer matches, return content unchanged (current behavior).

### Data Flow Diagram

```
User opens diff for "main.go"
           │
           ▼
┌──────────────────────────────────────────────────────────────────┐
│ DiffState.View()                                                 │
│   for each visible line:                                         │
│     renderFullFileLine(line, ...)                                │
│       │                                                          │
│       ▼                                                          │
│     formatColumnContent(lineNo=5, indicator="+",                 │
│                         content="func main() {", ...)            │
│       │                                                          │
│       ├─1─▶ expandTabs + truncate                                │
│       │       content = "func main() {"                          │
│       │                                                          │
│       ├─2─▶ highlight.Highlight(content, "main.go")              │
│       │       │                                                  │
│       │       ├─▶ lexers.Match("main.go") → Go lexer             │
│       │       ├─▶ lexer.Tokenise(content)                        │
│       │       ├─▶ formatter.Format(tokens, style)                │
│       │       └─▶ "[blue]func[reset] main() {"                   │
│       │                                                          │
│       ├─3─▶ columnStr = "  5 + [blue]func[reset] main() {"       │
│       │                                                          │
│       └─4─▶ DiffAdditionsStyle.Render(columnStr)                 │
│               → "[green-bg][blue]func[reset] main() {[reset]"    │
└──────────────────────────────────────────────────────────────────┘
           │
           ▼
    Terminal renders: blue "func" on green background
```

### Error Handling

The highlight module should never cause rendering failures:
- Lexer not found → return original content
- Tokenization error → return original content
- Formatter error → return original content

No errors or warnings displayed to users for unsupported file types.

## Open Questions

None. All decisions have been made based on research findings.

**Decisions summary:**
1. **Library:** Chroma (proven in production, excellent language/theme support)
2. **Integration point:** After truncation in `formatColumnContent()`
3. **Package structure:** New `internal/highlight` package
4. **Themes:** Monokai (dark) / GitHub (light), no customization
5. **Detection:** File extension via Chroma's lexer matching
