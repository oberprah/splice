# Codebase Analysis for Smart Navigation

## Current Architecture

### Data Flow: Files → Diff View

1. **Files State** (`internal/ui/states/files/state.go`):
   - Holds `Source` (DiffSource), `Files` ([]FileChange), and file tree structure
   - When user selects a file and presses Enter:
     - Calls `loadDiff()` which fetches full file content and builds `AlignedFileDiff`
     - Returns `DiffLoadedMsg` containing the diff and change indices
   - On `DiffLoadedMsg`, emits `PushDiffScreenMsg` to navigate to diff view

2. **Diff State** (`internal/ui/states/diff/state.go`):
   - Holds `Source`, `File`, `Diff`, `ViewportStart`, `CurrentChangeIdx`, `ChangeIndices`
   - Knows only about the single file being viewed
   - Has **no access** to the full file list or files state

3. **Navigation** (`internal/app/model.go`):
   - `PushDiffScreenMsg` creates a new `diff.State` and pushes it onto stack
   - `PopScreenMsg` removes current state, returning to files view
   - States are isolated - diff state cannot "reach back" to files state

### Current Data Structures

**ChangeIndices** (`[]int`):
- Pre-computed list of indices into `Alignments` that represent changes
- Every single changed line gets its own entry
- For a hunk with 10 changed lines, there are 10 separate indices

**Limitations:**
- No grouping of consecutive changes into logical hunks
- Can't easily determine "current hunk" or "end of current hunk"
- Navigation jumps line-by-line through changes, not hunk-by-hunk

### User's Input on Data Structure

The user suggested changing the underlying data structure to have a "list of Foo" where each element represents a logical group that can be:
- Unchanged (same on both sides)
- Addition
- Deletion
- Modification

This aligns with how git diff actually works - it produces "hunks" of changes, not individual lines.

## Key Observations

1. **Isolation Problem**: Diff state doesn't know about other files - need to pass file list when navigating to diff
2. **Granularity Problem**: ChangeIndices are per-line, not per-hunk - makes smart scrolling difficult
3. **Existing Pattern**: The `Alignment` type is already a sum type with the right cases (Unchanged, Modified, Removed, Added)
4. **Loading Pattern**: Diffs are loaded on-demand via async commands - file-to-file navigation needs to trigger loading
