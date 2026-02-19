# Requirements: Open File in Editor from Diff View

## Problem Statement

When viewing a diff in Splice, users often want to edit the file directly. Currently, they must:
1. Exit Splice or remember the file path
2. Open their editor manually
3. Navigate to the correct file and line

This breaks the flow and adds friction to the development workflow.

## Goals

- Allow users to quickly open the current file in their preferred editor directly from the diff view
- Position the cursor at the line corresponding to their current position in the diff
- Maintain a seamless terminal experience by properly suspending/resuming the TUI

## Non-Goals

- Support for GUI editors (only terminal-based editors via `$EDITOR`/`$VISUAL`)
- Editing historical versions of files (only current working directory version)
- Choosing between left/right versions in the diff (always open current version)
- Custom editor configuration beyond environment variables

## User Impact

**Who:** All Splice users who need to make quick edits while reviewing diffs

**Benefit:** Seamless transition from review to edit without context switching or manual file navigation

## Functional Requirements

### FR1: Keyboard Shortcut
- Pressing `o` in the diff view opens the current file in the user's editor
- The shortcut is only active when opening is possible (see FR5)

### FR2: Editor Selection
- Use `$EDITOR` environment variable to determine which editor to launch
- Fall back to `$VISUAL` if `$EDITOR` is not set
- If neither is set, show error message: "No editor configured (set $EDITOR or $VISUAL)"

### FR3: Cursor Positioning
- Open the file with the cursor positioned at the line corresponding to the current viewport position in the diff
- Use editor-specific syntax when supported (e.g., `vim +42 file.go`, `nano +42 file.go`)
- Common editors with `+line` syntax: vim, nvim, nano, emacs, vi
- For editors without known line positioning syntax, open at the top of the file

### FR4: TUI Suspension
- Properly suspend the Splice TUI when launching the editor (clear screen, restore terminal state)
- Resume the Splice TUI when the editor exits
- Preserve the diff state and scroll position after returning from the editor

### FR5: Edge Case Handling
- **No editor configured:** Show error message "No editor configured (set $EDITOR or $VISUAL)"
- **Deleted files (status `D`):** Show error message "Cannot open: file has been deleted"
- **File doesn't exist in working directory:** Show error message "Cannot open: file not found"
- **Binary files:** Show error message "Cannot open binary file in editor"
- **Editor launch fails:** Show error message with the underlying error (e.g., "Failed to launch editor: <error>")

### FR6: File Path Resolution
- The file path in `FileChange.Path` is relative to repository root
- Resolve to absolute path by combining with repository root before launching editor
- Map the diff line number (which may include alignment gaps) to the actual file line number in the new version

## Non-Functional Requirements

### NFR1: Performance
- Editor launch should be near-instantaneous (no noticeable delay beyond editor startup time)
- TUI suspension/resumption should be smooth with no visual artifacts

### NFR2: Compatibility
- Must work with common terminal editors (vim, nvim, nano, emacs, vi, etc.)
- Should handle editors that don't support line positioning gracefully

### NFR3: Testability
- Editor launch command should be mockable for unit tests
- TUI suspension should be testable without requiring a real terminal

## Open Questions

None - all requirements clarified.

## References

- Research: `research/diff-view-current-implementation.md` (see agent findings)
- Current diff view keybindings: `/home/user/splice/internal/ui/states/diff/update.go:13-80`
- File tracking: `/home/user/splice/internal/core/git_types.go:32-39`
- State structure: `/home/user/splice/internal/ui/states/diff/state.go:9-35`
