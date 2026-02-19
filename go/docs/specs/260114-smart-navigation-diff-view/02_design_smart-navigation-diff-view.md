# Design: Smart Navigation in Diff View

## Executive Summary

This design enables smart keyboard navigation in the diff view with two capabilities: (1) scrolling through multi-screen changes before jumping to the next change, and (2) navigating between files without returning to the files view.

The key architectural change is replacing the current flat `Alignment` data model with a **block-based structure**. Currently, changes are tracked as individual line indices (`ChangeIndices []int`) pointing into a flat `[]Alignment` slice, making it impossible to know where a logical "change" begins and ends. The new `FileDiff` structure groups lines into `UnchangedBlock` and `ChangeBlock` types, with blocks containing line data directly (not indices). This matches how users mentally model diffs and makes navigation logic straightforward.

For file-to-file navigation, the diff state will receive the full file list when created, giving it the context needed to navigate without returning to the files view.

## Context & Problem Statement

The diff view currently has basic change navigation (`n`/`N`) that jumps between individual changed lines within a single file. Three problems exist:

1. **Multi-screen changes**: When a change spans multiple screens, pressing `n` skips to the next change, potentially missing content
2. **No file navigation**: Users must press `q` to return to the files view to see the next file
3. **No smart scrolling**: The system doesn't know when the user has "finished viewing" a change

This design covers smart navigation within the diff view. It does not cover visual indicators, configuration options, or navigation history.

## Current State

### Data Model

```
AlignedFileDiff
├── Left: FileContent (old file lines with syntax tokens)
├── Right: FileContent (new file lines with syntax tokens)
└── Alignments: []Alignment (one per display row)
    ├── UnchangedAlignment (line same on both sides)
    ├── ModifiedAlignment (line changed, has inline diff)
    ├── RemovedAlignment (line only in old file)
    └── AddedAlignment (line only in new file)
```

Navigation uses `ChangeIndices []int` - a list of indices pointing to non-unchanged alignments. This is a flat list of individual lines with no grouping.

The `Alignment` types were designed for a flat structure where each represents a single display row. They embed indices into `FileContent.Lines` rather than containing the line data directly.

### State Isolation

The diff state only knows about the single file being viewed. It receives:
- `Source` (CommitRangeDiffSource or UncommittedChangesDiffSource)
- `File` (single FileChange)
- `Diff` (AlignedFileDiff for this file)
- `ChangeIndices` (line indices)

It has no access to the file list or knowledge of other files in the diff source.

## Solution

### 1. Block-Based Data Model

Replace the flat `[]Alignment` with a block-based structure. The `Alignment` types are removed entirely - blocks contain line data directly rather than indices.

> **Decision:** Remove `Alignment` types entirely rather than reusing them inside blocks. The alignment types were designed for a flat structure and carry unnecessary indirection (indices into FileContent). Blocks should own their line data directly, making the model cleaner and removing the need for FileContent lookup during rendering.

```
FileDiff
└── Blocks: []Block
    ├── UnchangedBlock
    │   └── Lines: []LinePair (left + right line with tokens)
    └── ChangeBlock
        └── Lines: []ChangeLine
            ├── ModifiedLine (left + right + inline diff)
            ├── RemovedLine (left only)
            └── AddedLine (right only)
```

**Type definitions:**

```go
// Block is a sealed interface for diff content blocks
type Block interface {
    block()          // Sealed marker
    LineCount() int  // Display lines in this block
}

// UnchangedBlock contains consecutive lines identical on both sides
type UnchangedBlock struct {
    Lines []LinePair
}

// LinePair holds matching lines from both file versions
type LinePair struct {
    LeftLineNo  int               // 1-indexed line number in old file
    RightLineNo int               // 1-indexed line number in new file
    Tokens      []highlight.Token // Shared tokens (content is identical)
}

// ChangeBlock contains consecutive changed lines (a "hunk")
type ChangeBlock struct {
    Lines []ChangeLine
}

// ChangeLine is a sealed interface for individual changed lines
type ChangeLine interface {
    changeLine()
}

// ModifiedLine: line exists in both files but content differs
type ModifiedLine struct {
    LeftLineNo  int
    RightLineNo int
    LeftTokens  []highlight.Token
    RightTokens []highlight.Token
    InlineDiff  []diffmatchpatch.Diff
}

// RemovedLine: line exists only in old file
type RemovedLine struct {
    LeftLineNo int
    Tokens     []highlight.Token
}

// AddedLine: line exists only in new file
type AddedLine struct {
    RightLineNo int
    Tokens      []highlight.Token
}
```

> **Decision:** `LinePair` shares tokens between sides since content is identical for unchanged lines. This saves memory and simplifies rendering. Changed lines store tokens separately since they differ.

**FileDiff replaces AlignedFileDiff:**

```go
// FileDiff is the top-level structure for a parsed file diff
type FileDiff struct {
    Path   string   // File path (for display)
    Blocks []Block
}
```

> **Decision:** Remove `Left`/`Right` `FileContent` fields. The old design stored full file content separately and used indices. The new design embeds line data directly in blocks, eliminating the need for separate file content storage.

### 2. Expanded Diff State

The diff state needs file context for navigation:

```go
type State struct {
    // File context for navigation
    Source    DiffSource
    Files     []FileChange  // All files in current diff source
    FileIndex int           // Position of current file in Files

    // Current file's diff
    Diff *FileDiff

    // Viewport
    ViewportStart int  // Display line offset (across all blocks)

    // Navigation state
    CurrentBlockIdx int  // Index into Blocks for current position
}
```

> **Decision:** Pass the full file list to diff state rather than using messages back through the navigation stack. This keeps the diff state self-sufficient for navigation decisions without complex coordination. The files list is small (just metadata, not diff content) so duplication is acceptable.

### 3. Navigation Behavior

#### Smart Change Navigation (`n`)

```
┌─────────────────────────────────────────────────────────────┐
│                    Press 'n' (next change)                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌─────────────────┐                                        │
│  │ In ChangeBlock? │                                        │
│  └────────┬────────┘                                        │
│           │                                                 │
│     ┌─────┴─────┐                                           │
│     │           │                                           │
│    Yes          No                                          │
│     │           │                                           │
│     ▼           ▼                                           │
│  ┌──────────┐  Jump to next ChangeBlock ────────────────┐   │
│  │ Block end│                                           │   │
│  │ visible? │                                           │   │
│  └────┬─────┘                                           │   │
│       │                                                 │   │
│   ┌───┴───┐                                             │   │
│   │       │                                             │   │
│  Yes      No                                            │   │
│   │       │                                             │   │
│   ▼       ▼                                             │   │
│ Jump to  Scroll down                                    │   │
│ next     half page                                      │   │
│ Change   (stay in block)                                │   │
│ Block                                                   │   │
│   │                                                     │   │
│   ▼                                                     │   │
│ ┌────────────────┐                                      │   │
│ │ More blocks in │                                      │   │
│ │ current file?  │◄─────────────────────────────────────┘   │
│ └───────┬────────┘                                          │
│         │                                                   │
│     ┌───┴───┐                                               │
│     │       │                                               │
│    Yes      No                                              │
│     │       │                                               │
│     ▼       ▼                                               │
│   Jump    ┌──────────────┐                                  │
│   to it   │ More files?  │                                  │
│           └──────┬───────┘                                  │
│                  │                                          │
│              ┌───┴───┐                                      │
│              │       │                                      │
│             Yes      No                                     │
│              │       │                                      │
│              ▼       ▼                                      │
│           Load     Stay in                                  │
│           next     place                                    │
│           file                                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

#### Smart Change Navigation (`p`)

Mirror of `n`: scrolls up through current change if above viewport, otherwise jumps to previous change or previous file.

#### File Navigation (`]` and `[`)

Direct file jumps:
1. Calculate target file index (current ± 1)
2. If valid: trigger async diff load, position at first ChangeBlock
3. If boundary: stay in place

### 4. State Transitions

```
┌─────────────────────────────────────────────────────────────┐
│                      FilesState                             │
│  (has full file list and loads diffs)                       │
└────────────────────────┬────────────────────────────────────┘
                         │ Enter on file
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                       DiffState                             │
│  (receives file list + current file's diff)                 │
│                                                             │
│  Navigation:                                                │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐        │
│  │    n    │  │    p    │  │    ]    │  │    [    │        │
│  │  next   │  │  prev   │  │  next   │  │  prev   │        │
│  │ change  │  │ change  │  │  file   │  │  file   │        │
│  └────┬────┘  └────┬────┘  └────┬────┘  └────┬────┘        │
│       │           │            │            │               │
│       ▼           ▼            ▼            ▼               │
│  Smart scroll  Smart scroll  Load diff   Load diff         │
│  or jump       or jump       for next    for prev          │
│  within/across within/across file        file              │
│  files         files                                       │
│                                                             │
│  Loading triggers async command, returns to DiffState       │
│  with new file's diff                                       │
└─────────────────────────────────────────────────────────────┘
```

### 5. Async Loading for File Navigation

When navigating to a different file, the diff needs to be loaded. Reuse the existing pattern from files state:

1. `]` pressed in diff state
2. State creates command to fetch diff for `Files[FileIndex+1]`
3. Command returns `DiffLoadedMsg`
4. State updates with new diff, resets viewport to first ChangeBlock

> **Decision:** Handle file navigation within diff state rather than popping back to files state. This avoids visual flicker and keeps navigation smooth. The diff state will have its own `loadDiff` function similar to files state.

## Changes Required

### Domain Layer (`internal/domain/diff/`)

1. **New file**: `block.go` with `Block`, `UnchangedBlock`, `ChangeBlock`, `LinePair`, `ChangeLine`, `ModifiedLine`, `RemovedLine`, `AddedLine`
2. **New type**: `FileDiff` (replaces `AlignedFileDiff`)
3. **Modified**: `BuildAlignedFileDiff` → `BuildFileDiff`, returns `*FileDiff` with blocks
4. **Removed**: `alignment.go` types (`Alignment`, `UnchangedAlignment`, `ModifiedAlignment`, `RemovedAlignment`, `AddedAlignment`, `AlignedLine`, `FileContent`, `AlignedFileDiff`)
5. **Removed**: `ChangeIndices` return value (replaced by block structure)

### UI Layer (`internal/ui/states/diff/`)

1. **Modified**: `State` struct:
   - Gains `Files []FileChange`, `FileIndex int` fields
   - `Diff` type changes from `*AlignedFileDiff` to `*FileDiff`
   - Removes `ChangeIndices`, `CurrentChangeIdx` fields
   - Adds `CurrentBlockIdx int`
2. **Modified**: View rendering iterates blocks, then lines within blocks
3. **Modified**: `Update` handles `n`, `p` with smart scrolling logic
4. **New**: Handlers for `]`, `[` file navigation
5. **New**: `loadDiff` function for async file loading

### Navigation (`internal/core/`)

1. **Modified**: `PushDiffScreenMsg`:
   - Gains `Files []FileChange`, `FileIndex int`
   - `Diff` type changes to `*FileDiff`
   - Removes `ChangeIndices` field
2. **Modified**: `DiffLoadedMsg` same changes as above

### Files State (`internal/ui/states/files/`)

1. **Modified**: Pass `Files` and `FileIndex` when creating `PushDiffScreenMsg`
2. **Modified**: `loadDiff` returns new `FileDiff` type

## Open Questions

None - ready for implementation.

## References

- [Codebase analysis](research/codebase-analysis.md)
- [Data structure options](research/data-structure-options.md)
- [Navigation logic exploration](research/navigation-logic.md)
