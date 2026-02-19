# Design: Open File in Editor from Diff View

## Executive Summary

This feature adds an `o` keyboard shortcut to the diff view that opens the current file in the user's configured terminal editor (`$EDITOR` or `$VISUAL`) with the cursor positioned at the current viewport line. The implementation leverages Bubbletea's built-in `ExecProcess()` function for TUI suspension/resumption, adds a new `git.GetRepositoryRoot()` function for path resolution, and uses the existing alignment structure to map viewport positions to file line numbers. The design requires minimal changes to existing code—primarily adding a key handler in `diff/update.go` and a helper function in the git package. Error cases (no editor configured, deleted files, binary files) are handled with user-friendly error messages.

## Context & Problem Statement

When reviewing code in Splice's diff view, users frequently want to edit the file they're viewing. Currently, they must exit Splice (or switch terminal tabs), manually open their editor, navigate to the file, and find the relevant line. This context switching breaks flow and adds friction to the development workflow.

This design adds a seamless editor integration that:
- Opens files directly from the diff view with a single keypress
- Positions the cursor at the line being viewed
- Maintains state when returning to Splice

**Scope:** This design covers opening the current file from the diff view in a terminal editor. It does not cover GUI editors, editing historical file versions, or configuring editors beyond environment variables.

## Current State

### Diff View Architecture

The diff view (`internal/ui/states/diff/`) follows Splice's standard state pattern:
- **State struct** (`state.go:9-35`): Holds `FileChange`, `AlignedFileDiff`, and `ViewportStart` (scroll position)
- **Update method** (`update.go`): Handles keyboard input and returns new state + commands
- **View method** (`view.go`): Renders the diff

Current keybindings: `q` (back), `j/k` (scroll), `g/G` (jump to top/bottom), `n/N` (next/previous change).

### File Path Handling

- `FileChange.Path` is relative to repository root (`internal/core/git_types.go:32-39`)
- Git commands use `:(top)` pathspec magic to work with relative paths
- **Repository root is not currently stored or calculated anywhere in the application**

### Alignment Structure

The diff uses an aligned representation (`internal/domain/diff/alignment.go`):
- `Alignments[]` array: one entry per display row
- Each alignment is one of: `UnchangedAlignment`, `ModifiedAlignment`, `AddedAlignment`, `RemovedAlignment`
- Alignments with `RightIdx` map to lines in the new file version
- `ViewportStart` is an index into the `Alignments` array

### External Command Execution

Splice uses Go's `os/exec` package only for git commands. There is no existing pattern for launching interactive external programs or suspending the TUI.

## Solution

### Overview

Add an `o` key handler in the diff view that:
1. Validates preconditions (editor configured, file exists, not binary)
2. Resolves the file path to an absolute path
3. Maps viewport position to file line number
4. Launches the editor using Bubbletea's `ExecProcess()`
5. Handles errors gracefully with user-friendly messages

### Component Interaction

```
User presses 'o' in diff view
         ↓
DiffState.Update() validates preconditions
         ↓
Calculate line number from ViewportStart + Alignments
         ↓
Resolve relative path → absolute path via git.GetRepositoryRoot()
         ↓
Return tea.Cmd that calls tea.ExecProcess()
         ↓
Bubbletea suspends TUI, launches editor
         ↓
User edits, exits editor
         ↓
EditorFinishedMsg returned to Update()
         ↓
Handle error (if any) or resume seamlessly
```

### Key Decisions

#### Decision 1: Use Bubbletea's ExecProcess()

**Choice:** Use `tea.ExecProcess()` for TUI suspension and editor launch.

**Rationale:**
- Bubbletea provides built-in support specifically for this use case
- Automatically handles terminal state save/restore (`ReleaseTerminal()`/`RestoreTerminal()`)
- Callback-based API fits Bubbletea's message-passing architecture
- Well-documented pattern with official examples

**Alternative considered:** Manual terminal state management with `ReleaseTerminal()`/`RestoreTerminal()`. Rejected because it's more error-prone and `ExecProcess()` is the recommended approach.

**Tradeoff:** Requires Bubbletea v0.23.0+ (Splice already uses v0.25.0+).

#### Decision 2: Add git.GetRepositoryRoot() Function

**Choice:** Add a new `GetRepositoryRoot()` function to the git package that calls `git rev-parse --show-toplevel`.

**Rationale:**
- Repository root is not currently available in the application
- `git rev-parse --show-toplevel` is the standard, reliable way to get the repo root
- Fits existing pattern of git commands in `internal/git/git.go`
- Minimal, focused function with single responsibility

**Alternative considered:** Add repository root to `app.Model` and thread it through the context. Rejected because it adds complexity for a single use case and requires changes to multiple layers.

**Tradeoff:** Adds a git command execution per editor launch (negligible performance impact).

#### Decision 3: Handle RemovedAlignment by Finding Adjacent Line

**Choice:** When viewport is at a `RemovedAlignment` (deleted line), search forward for the next alignment with a `RightIdx`, or fall back to line 1.

**Rationale:**
- Removed lines don't exist in the new file (no `RightIdx`)
- Opening at the next available line provides reasonable UX (user sees context around the change)
- Simple, predictable behavior

**Alternative considered:** Always open at line 1 for removed lines. Rejected because it's less helpful—user loses context about where the change was.

**Tradeoff:** Slightly more complex logic, but better UX.

#### Decision 4: Standard +line Syntax for All Editors

**Choice:** Use `editor +lineNumber filePath` command format for all editors.

**Rationale:**
- All common terminal editors (vim, nvim, vi, nano, emacs) support `+line` syntax
- Simple, uniform implementation
- If an editor doesn't support it, it gracefully ignores the flag and opens at top

**Alternative considered:** Detect editor and use editor-specific syntax (e.g., emacs can use `file:line`). Rejected as over-engineering—the `+line` syntax is universal enough.

**Tradeoff:** Uncommon editors that don't support `+line` will open at the top (acceptable degradation).

#### Decision 5: Show Error Messages, Don't Silently Fail

**Choice:** When the `o` key can't open the file (no editor, deleted file, binary file), show a brief error message.

**Rationale:**
- Users need feedback to understand why nothing happened
- Clear error messages are more user-friendly than silent failures
- Consistent with Splice's existing error handling patterns

**Alternative considered:** Silently disable the shortcut. Rejected because users won't understand why the key doesn't work.

**Tradeoff:** Requires displaying error messages (can use existing error screen or inline message).

### Data Flow and State Changes

**Pre-launch validation:**
```
Check $EDITOR/$VISUAL → error if unset
Check File.IsBinary → error if binary
Check File.Status != "D" → error if deleted
Check file exists on disk → error if missing
```

**Line number calculation:**
```
ViewportStart (int) → Alignments[ViewportStart] (Alignment)
                   ↓
              Type switch on alignment
                   ↓
         Extract RightIdx (or find adjacent)
                   ↓
            lineNumber = RightIdx + 1
```

**Path resolution:**
```
File.Path (relative) → git rev-parse --show-toplevel → repoRoot
                                    ↓
                    filepath.Join(repoRoot, File.Path) → absolute path
```

**Editor launch:**
```
tea.Cmd → tea.ExecProcess(cmd, callback) → EditorFinishedMsg
                                          ↓
                               Handle error or resume
```

### Type Definitions

**New message type in `diff/update.go`:**
```go
type EditorFinishedMsg struct {
    err error
}
```

**New function in `internal/git/git.go`:**
```go
func GetRepositoryRoot() (string, error)
```

**New helper methods in `diff/update.go`:**
```go
func (s *State) openFileInEditor() tea.Cmd
func (s *State) getCurrentFileLineNumber() (int, error)
func (s *State) canOpenInEditor() error
```

### Error Handling

| Error Case | Detection | User Message |
|------------|-----------|--------------|
| No editor configured | `os.Getenv("EDITOR")` and `os.Getenv("VISUAL")` both empty | "No editor configured (set $EDITOR or $VISUAL)" |
| File deleted | `File.Status == "D"` | "Cannot open: file has been deleted" |
| File not found | `os.Stat(absolutePath)` returns error | "Cannot open: file not found" |
| Binary file | `File.IsBinary == true` | "Cannot open binary file in editor" |
| Editor launch fails | `EditorFinishedMsg.err != nil` | "Failed to launch editor: \<error\>" |
| Repo root error | `git.GetRepositoryRoot()` returns error | "Failed to determine repository root: \<error\>" |

All errors can be displayed using a simple status message or by pushing an error screen (decision deferred to implementation).

### Testing Strategy

**Unit tests for line number mapping:**
- Test each alignment type (Unchanged, Modified, Added, Removed)
- Test RemovedAlignment fallback logic
- Test edge cases (empty diff, viewport out of range)

**Unit tests for validation:**
- Test each error condition
- Test successful validation path

**Integration test with mock editor:**
- Mock `tea.ExecProcess` to avoid actually launching editor
- Verify correct command is constructed (editor path, +line, file path)
- Verify state is preserved after "editor" returns

**Manual testing:**
- Test with various editors (vim, nvim, nano, emacs)
- Test error cases in real environment
- Verify TUI suspend/resume is clean

## Open Questions

None. All design decisions have been made and documented.

## References

- Requirements: `01_requirements_open-file-editor.md`
- Research documents:
  - `research/diff-view-current-implementation.md`
  - `research/bubbletea-exec-patterns.md`
  - `research/repository-root-resolution.md`
  - `research/line-number-mapping.md`
