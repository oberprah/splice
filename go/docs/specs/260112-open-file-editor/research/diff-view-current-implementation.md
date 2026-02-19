# Research: Diff View Current Implementation

## Overview

The diff view is implemented in `/home/user/splice/internal/ui/states/diff/` following the standard state pattern used throughout Splice.

## State Structure

**Location:** `/home/user/splice/internal/ui/states/diff/state.go:9-35`

```go
type State struct {
	CommitRange      core.CommitRange        // Start and end commits for the diff
	File             core.FileChange         // The file being viewed
	Diff             *diff.AlignedFileDiff   // The actual diff data
	ViewportStart    int                     // Current scroll position
	CurrentChangeIdx int                     // Current position in ChangeIndices
	ChangeIndices    []int                   // Line indices with changes (for "n"/"N" navigation)
}
```

**Key fields for editor integration:**
- `File.Path` - File path relative to repository root
- `File.Status` - Git status (M, A, D, R, etc.)
- `File.IsBinary` - Whether file is binary
- `ViewportStart` - Current scroll position (needed for cursor positioning)
- `Diff.Right.Lines` - The new version of the file with line content

## Current Keybindings

**Location:** `/home/user/splice/internal/ui/states/diff/update.go:13-80`

| Key | Action |
|-----|--------|
| `q` | Go back (pop navigation stack) |
| `ctrl+c`, `Q` | Quit application |
| `j`, `down` | Scroll down one line |
| `k`, `up` | Scroll up one line |
| `ctrl+d` | Scroll down half page |
| `ctrl+u` | Scroll up half page |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `n` | Jump to next change |
| `N` | Jump to previous change |

**Available keys:** `o`, `e`, `Enter`, and many others are currently unassigned.

## File Tracking

**FileChange Type:** `/home/user/splice/internal/core/git_types.go:32-39`

```go
type FileChange struct {
	Path      string // Relative to repository root
	Status    string // M, A, D, R, etc.
	Additions int
	Deletions int
	IsBinary  bool
}
```

**Navigation:** Files are passed to diff state via `PushDiffScreenMsg` (`/home/user/splice/internal/core/navigation.go:25-31`)

## Diff Data Structure

**AlignedFileDiff:** `/home/user/splice/internal/domain/diff/alignment.go:99-107`

```go
type AlignedFileDiff struct {
	Left       FileContent // Old version
	Right      FileContent // New version
	Alignments []Alignment // One per display row
}

type FileContent struct {
	Path  string
	Lines []AlignedLine  // With syntax highlighting
}
```

**Alignment Types:**
- `UnchangedAlignment` - Identical lines
- `ModifiedAlignment` - Changed lines (with inline diff)
- `RemovedAlignment` - Only in old version
- `AddedAlignment` - Only in new version

**Change Navigation:** `ChangeIndices` array stores alignment indices with actual changes, used by `jumpToNextChange()`/`jumpToPreviousChange()` methods.

## External Command Execution

**Current Pattern:** `/home/user/splice/internal/git/git.go:6`

- Uses `os/exec.Command()` for git commands
- Example: `cmd := exec.Command("git", "log", ...)`
- No existing editor integration in the codebase

**Bubbletea Async Pattern:**
- Return `tea.Cmd` that executes async and returns a message
- Example: diff loading returns `DiffLoadedMsg`
- States never block in Update methods

## Architecture Patterns

**Navigation:**
- Typed messages: `PushDiffScreenMsg`, `PopScreenMsg`, etc.
- Model handles routing: `/home/user/splice/internal/app/model.go:72-106`

**Dependency Injection:**
- `core.Context` interface provides injected dependencies
- Allows clean testing with mocks

**Error Handling:**
- `PushErrorScreenMsg` can be used to show error states

## Key Insights for Editor Feature

1. **File path available:** `s.File.Path` (relative to repo root)
2. **File status available:** `s.File.Status` (can detect deleted files)
3. **Binary detection:** `s.File.IsBinary` flag
4. **Current line:** Can be calculated from `s.ViewportStart` and alignment data
5. **Async pattern ready:** Can use `tea.Cmd` for launching editor
6. **Error handling:** Can use error messages or error screen for failures
7. **Clean architecture:** Can add new keybinding without breaking existing code
