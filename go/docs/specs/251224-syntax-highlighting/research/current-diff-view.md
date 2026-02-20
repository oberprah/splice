# Current Diff View Implementation

## Overview

The diff view in Splice displays a side-by-side view of file changes. The left column shows the old file version, the right column shows the new file version.

## Key Files

- `internal/ui/states/diff_state.go` - State struct definition
- `internal/ui/states/diff_view.go` - Rendering logic
- `internal/diff/merge.go` - Merges full file content with diff info
- `internal/ui/styles/styles.go` - Style definitions

## Data Structures

### FullFileDiff (`internal/diff/merge.go`)
```go
type FullFileLine struct {
    LeftLineNo   int        // Line number in old file (0 if not present)
    RightLineNo  int        // Line number in new file (0 if not present)
    LeftContent  string     // Content from old file
    RightContent string     // Content from new file
    Change       ChangeType // Unchanged, Added, or Removed
}

type FullFileDiff struct {
    OldPath       string
    NewPath       string
    Lines         []FullFileLine
    ChangeIndices []int  // For change navigation
}
```

### DiffState (`internal/ui/states/diff_state.go`)
Holds the `FullFileDiff` plus viewport control for scrolling.

## Current Rendering

The diff view (`diff_view.go`) already:
- Shows the **entire file** (not just diff hunks)
- Uses a **side-by-side layout** (left = old, right = new)
- Applies **background colors** for changed lines:
  - `DiffAdditionsStyle` - subtle green background for added lines
  - `DiffDeletionsStyle` - subtle red/pink background for removed lines
- Shows line numbers on each side
- Handles truncation with ellipsis

## Current Styling

From `styles.go`:
```go
// Diff line addition style (green background)
DiffAdditionsStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
    Light: "#e8f5e9", // Very subtle pale green
    Dark:  "#1e3a1e", // Very subtle dark green
})

// Diff line deletion style (red background)
DiffDeletionsStyle = lipgloss.NewStyle().Background(lipgloss.AdaptiveColor{
    Light: "#ffebee", // Very subtle pale pink
    Dark:  "#3a1e1e", // Very subtle dark red
})
```

## What's Missing

**No syntax highlighting** - all code content is rendered in the same color (gray for unchanged lines, or with background color for changes). There's no language-aware coloring of keywords, strings, comments, etc.

## Technology Stack

The project uses:
- **Lip Gloss** (from Charm) for terminal styling
- **Bubbletea** for the TUI framework

These work with ANSI escape codes for colors.
