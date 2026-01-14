# Navigation Logic Design

## Problem: File-to-File Navigation

The diff state currently only knows about the single file being viewed. To navigate between files, we need:

1. Access to the full file list
2. Knowledge of current file's position in the list
3. Ability to load diffs for other files

### Option A: Pass File List to Diff State

Expand `PushDiffScreenMsg` and `diff.State` to include the full context:

```go
type PushDiffScreenMsg struct {
    Source        DiffSource
    Files         []FileChange  // Full list
    FileIndex     int           // Current file's position
    File          FileChange
    Diff          *AlignedFileDiff
    ChangeIndices []int
}
```

**Pros:**
- Diff state has all info needed for navigation
- Single source of truth for file ordering

**Cons:**
- Diff state becomes larger
- File list duplicated from files state

### Option B: Navigation Messages Back Through Stack

Diff state emits messages that app.Model handles:

```go
type NavigateToNextFileMsg struct{}
type NavigateToPrevFileMsg struct{}
```

App.Model would need to:
1. Pop diff state
2. Look at files state to find next/prev file
3. Load diff for that file
4. Push new diff state

**Pros:**
- Diff state stays focused on single file
- Reuses existing navigation patterns

**Cons:**
- Complex coordination between app and states
- Files state needs to expose methods or app needs file-ordering logic
- Visual "flicker" from pop/push cycle

### Option C: Shared Context Through Files State

Files state manages navigation, diff state is "child" that delegates file changes up:

**Cons:**
- Against the current navigation stack pattern
- Would require significant architectural changes

## Recommendation

**Option A** is most straightforward. The diff state needs to know about files anyway to navigate, and passing the list is explicit and testable.

## Problem: Smart Change Navigation

Need to scroll through multi-screen changes before jumping to next change.

### State Tracking

With the two-level block structure, the state needs:

```go
type State struct {
    // File context
    Files     []FileChange  // All files in diff
    FileIndex int           // Current file index

    // Diff content
    Source DiffSource
    Diff   *AlignedFileDiff  // Contains Blocks

    // Viewport
    ViewportStart int  // Line offset into current view

    // Navigation tracking
    CurrentBlockIdx int  // Which block we're in (for change navigation)
}
```

### Navigation Algorithm for `n` (next change)

```
1. Find current position in block structure
2. Determine current block (may be unchanged or change block)
3. If in a ChangeBlock:
   a. Calculate block's end position
   b. If end of block is below viewport bottom:
      - Scroll down half page
      - Return (stay in same block)
   c. Else (entire block visible or scrolled past):
      - Move to next ChangeBlock
4. If in UnchangedBlock or no current block:
   - Find next ChangeBlock and jump to it
5. If no more ChangeBlocks in file:
   - If more files exist: navigate to next file's first ChangeBlock
   - Else: stay in place (boundary)
```

### Navigation Algorithm for `p` (previous change)

```
1. Find current position in block structure
2. If at the start of a ChangeBlock:
   - Jump to previous ChangeBlock's start
3. If inside a ChangeBlock (scrolled down):
   a. If start of block is above viewport top:
      - Scroll up half page
      - Return
   b. Else:
      - Jump to this block's start
4. If in UnchangedBlock:
   - Jump to previous ChangeBlock's start
5. If no previous ChangeBlocks:
   - If previous files exist: navigate to prev file's first ChangeBlock
   - Else: stay in place (boundary)
```

### File Navigation (`]` and `[`)

Simpler - just loads the next/previous file's diff and positions at first change:

```
1. Calculate target file index
2. If valid index:
   a. Trigger async diff load for target file
   b. On load complete: update state with new diff, position at first ChangeBlock
3. If invalid (boundary): stay in place
```
