# Diff View Rendering Analysis

## Overview

Analysis of how Splice currently renders the diff view, to inform syntax highlighting integration.

## Data Flow

1. Raw git diff → `parse.ParseUnifiedDiff()` → `FileDiff` with `Line` structs
2. Old/new file content + parsed diff → `diff.MergeFullFile()` → `FullFileDiff` with `FullFileLine` structs
3. `FullFileLine` structs → `DiffState.View()` → rendered as side-by-side columns

## Key Files

- `internal/ui/states/diff_view.go` - Main rendering logic
- `internal/ui/states/diff_state.go` - State structure
- `internal/ui/styles/styles.go` - Style definitions
- `internal/diff/merge.go` - Diff data structures

## Current Rendering Process

### Per-Line Rendering (`diff_view.go`)

Loop through visible diff lines from viewport:
1. Call `renderFullFileLine()` for each line
2. Returns (left_column_string, right_column_string)
3. Use `lipgloss.JoinHorizontal()` to combine: left column + separator + right column
4. Append newline after each row

### Style Application (Critical Finding)

**Styling happens in `formatColumnContent()` at a single point:**

```go
columnStr := lineNoStr + " " + indicator + " " + truncated
return style.Render(columnStr)  // ONE style covers everything
```

The entire line (line number + indicator + content) receives a single monolithic style.

**Styles applied:**
- Unchanged lines: `styles.TimeStyle` (gray foreground)
- Removed lines: `styles.DiffDeletionsStyle` (red background)
- Added lines: `styles.DiffAdditionsStyle` (green background)

### Code Paths by Change Type

**Unchanged lines:**
```
FullFileLine.Change == diff.Unchanged
→ formatColumnContent(lineNo, " ", content, ..., styles.TimeStyle)
→ Displayed: "  1   content" | "  1   content"
```

**Removed lines:**
```
FullFileLine.Change == diff.Removed
→ formatColumnContent(lineNo, "-", LeftContent, ..., styles.DiffDeletionsStyle)
→ Displayed: "  1 - content" | (empty)
```

**Added lines:**
```
FullFileLine.Change == diff.Added
→ formatColumnContent(lineNo, "+", RightContent, ..., styles.DiffAdditionsStyle)
→ Displayed: (empty) | "  1 + content"
```

## Existing Style Patterns

### Adaptive Colors

All styles use `lipgloss.AdaptiveColor` for light/dark terminal support:

```go
Background(lipgloss.AdaptiveColor{
    Light: "#e8f5e9",  // Light terminals
    Dark:  "#1e3a1e",  // Dark terminals
})
```

### Diff Line Styles (from `styles.go`)

Very subtle background colors:
- `DiffAdditionsStyle`: `#e8f5e9` (light) / `#1e3a1e` (dark) - pale green
- `DiffDeletionsStyle`: `#ffebee` (light) / `#3a1e1e` (dark) - pale red/pink

### Content Width

```go
contentWidth = columnWidth - lineNoWidth - 4
// 4 = space + indicator + space (e.g., " - ")
```

Content truncated with ellipsis if exceeds width. Tabs expanded to 4 spaces before truncation.

## Integration Points for Syntax Highlighting

### Option A: Pre-render Content (Recommended)

Add syntax highlighting before `formatColumnContent()`:
1. Extend `FullFileLine` with `HighlightedLeftContent` and `HighlightedRightContent` fields
2. Apply Chroma highlighting when loading diff data
3. `formatColumnContent()` uses highlighted content if present, falls back to raw content

Benefits: Clean separation, graceful fallback, highlighting computed once.

### Option B: Content Preprocessing

Add highlighting layer between `MergeFullFile()` and rendering:
1. New function: `ApplySyntaxHighlighting(fullDiff, fileType)` → modified `FullFileDiff`
2. Returns `FullFileDiff` with ANSI-colored content strings
3. Rendering logic unchanged

Benefits: Single point of integration, testable in isolation.

### Key Constraint

Current design applies ONE style to entire line. For syntax highlighting:
- Content portion needs per-token styling (Chroma output)
- Line number and indicator need separate styling
- Background style needs to wrap the whole line

Solution: Build content as `lineNoStyled + " " + indicator + " " + chromaHighlighted`, then wrap with background style.
