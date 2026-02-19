# Current Diff View Implementation

## Overview

The diff view (`internal/ui/states/diff/`) displays side-by-side diffs with syntax highlighting and inline diff highlighting for modified lines.

## State Structure

```go
type State struct {
    Source           core.DiffSource    // Where the diff came from (commit range or uncommitted)
    File             core.FileChange    // Metadata about the file
    Diff             *diff.AlignedFileDiff  // Parsed diff with alignments
    ViewportStart    int                // Current scroll position (index into Alignments)
    CurrentChangeIdx int                // Index into ChangeIndices for navigation
    ChangeIndices    []int              // Indices of alignments that have changes
}
```

## Key Concepts

- **Alignments**: The diff content is represented as a slice of `Alignment` objects, one per line. Each alignment can be:
  - `UnchangedAlignment`: Line appears on both sides
  - `ModifiedAlignment`: Line changed between sides (has inline diff)
  - `AddedAlignment`: Line only on right side
  - `RemovedAlignment`: Line only on left side

- **ChangeIndices**: A pre-computed list of indices into the Alignments slice that point to lines with actual changes (Modified, Added, or Removed alignments). This is used for jump-to-change navigation.

## Current Keyboard Shortcuts

### Scrolling
- `j`/`down`: Scroll down one line
- `k`/`up`: Scroll up one line
- `ctrl+d`: Scroll down half page
- `ctrl+u`: Scroll up half page
- `g`: Jump to top
- `G`: Jump to bottom

### Change Navigation
- `n`: Jump to next change (`jumpToNextChange`)
- `N`: Jump to previous change (`jumpToPreviousChange`)

### Other
- `o`: Open file in editor at current line
- `q`: Go back to previous screen (files view)
- `Q`/`ctrl+c`: Quit application

## Current Jump-to-Change Behavior

The existing `jumpToNextChange` and `jumpToPreviousChange` functions (lines 108-150 in `update.go`):

1. Find the next/previous index in `ChangeIndices` after/before the current `ViewportStart`
2. Set `ViewportStart` to that change index
3. Clamp to max viewport to prevent scrolling past the end
4. Do NOT wrap around (stay at current position if no more changes found)

**Limitations**:
- Only jumps within the current file
- Does not handle multi-screen changes (if a change is taller than the viewport)
- Does not automatically navigate to next/previous file when reaching end of current file

## Navigation Context

- Users can press `q` in diff view to go back to the files view
- Files view shows a tree of all changed files in the current diff source
- Users can navigate files in the files view and press Enter to view a diff
- There is currently NO direct file-to-file navigation from within the diff view

## Source Information

The diff view can show diffs from different sources:
- **CommitRangeDiffSource**: Diff between commits (single commit or range)
- **UncommittedChangesDiffSource**: Unstaged, staged, or all uncommitted changes

The files view has access to `s.Files` (slice of all `FileChange` objects in the diff source), so it knows about all files in the current context.
