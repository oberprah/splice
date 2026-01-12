# Research: UI Components

**Date:** 2026-01-12

## Components Inventory

`internal/ui/components/` contains:

- **ViewBuilder** - Base rendering utility
- **CommitInfo** - Commit metadata and body rendering
- **CommitInfoFromRange** - Handles single commits and ranges
- **FileSection** - File list with stats rendering
- **FormatCommitLine** - Progressive truncation for commit lines
- **LineDisplayState** - Display state enumeration

## Component Structure

**Pure Functions:**
- Most components are pure functions
- No internal state management
- State passed via parameters

**Return Types:**
- Some return `[]string` (lines to display)
- Others use `ViewBuilder` and return `core.ViewRenderer`

Example: `CommitInfo(commit, width, bodyMaxLines, ctx) []string`

## Typical Component Signatures

**Line-based components:**
```go
CommitInfo(commit core.GitCommit, width int, bodyMaxLines int, ctx core.Context) []string
FileSection(files []core.FileChange, width int, cursor *int) []string
```

**Complex formatting:**
```go
FormatCommitLine(components CommitLineComponents, availableWidth int) string

type CommitLineComponents struct {
    DisplayState LineDisplayState
    Graph        string
    Hash         string
    Refs         []core.RefInfo
    Message      string
    Author       string
    Time         string
}
```

**Builder pattern:**
```go
vb := components.NewViewBuilder()
vb.AddLine(styledLine)
vb.AddSplitView(leftVb, rightVb)
return vb
```

## Styling Architecture

**Styles applied internally, not passed as parameters.**

Pre-defined Lip Gloss styles in `internal/ui/styles/styles.go`:
- `HashStyle`, `MessageStyle`, `AuthorStyle`, `TimeStyle`
- `Selected*Style` variants (bold)
- `AdditionsStyle`, `DeletionsStyle`
- `FilePathStyle`

Usage:
```go
line := styles.HashStyle.Render(hash)
line += styles.MessageStyle.Render(message)
```

Components directly use module-level styles.

## Component Usage Pattern

**From states:**
```go
// Build components separately
lineComponents := s.buildCommitLineComponents(commit, i, false, ctx)

// Call pure function for formatting
line := components.FormatCommitLine(lineComponents, width)

// Add to ViewBuilder
vb.AddLine(line)
```

## Special Features

**Progressive Truncation:**
- FormatCommitLine implements 10-level progressive truncation
- Gracefully degrades for narrow terminals
- Priority: hash > time > author > message

**Composite Rendering:**
- ViewBuilder.AddSplitView() joins two columns with " │ " separator
- Used in log split view (commit list + files preview)

**Display States (Enum):**
```go
const (
    LineStateNone         // Regular line
    LineStateCursor       // Normal mode cursor "→ "
    LineStateSelected     // Visual mode selected "▌ "
    LineStateVisualCursor // Visual mode cursor "█ "
)
```

## No Overlay/Modal Components

**None exist currently.** All rendering is full-screen, state-based.

## Testing Support

Mock context and helpers in `internal/ui/testutils/`:
- `MockContext` - Test implementation
- `CreateTestCommits()` - Test fixtures
- `AssertGolden()` - Golden file testing
- Mock fetch functions

## Design Principles

1. **Pure Functions** - Deterministic, no hidden state
2. **Explicit Parameters** - All inputs passed explicitly
3. **Separation of Concerns** - Data prep (impure) in states, formatting (pure) in components
4. **Composability** - Components combine naturally
5. **Internal Styling** - Styles not configurable per-call

## Recommendations for Help Overlay

- Follow pure function pattern (no state in component)
- Return `[]string` for lines or use ViewBuilder
- Define overlay styling in `internal/ui/styles/styles.go`
- Accept `width` and `ctx` parameters
- Use existing style patterns
