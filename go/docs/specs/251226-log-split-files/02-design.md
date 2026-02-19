# Design: Log View Split Files Panel

## Executive Summary

This design adds a split panel to the log view that displays commit details and changed files for the currently selected commit. The panel appears on the right side when terminal width exceeds 160 characters, using `lipgloss.JoinHorizontal` for layout (the same pattern used in the diff view).

The implementation introduces eager loading: when the user moves the cursor, a `tea.Cmd` fetches file data asynchronously. The panel shows "Loading..." until data arrives, then displays the commit message and file list. A hash check ensures stale responses are discarded when the user navigates faster than data loads.

This approach requires changes only to the log state files—no new states or complex state machine modifications.

## Context & Problem Statement

When viewing the commit log, users must press Enter and navigate to a separate screen to see which files changed in a commit. This interrupts the browsing flow for a quick overview.

**Scope:** This design covers the split panel display in the log view. It does not cover interactive file selection from the panel (that remains on the full-screen files view).

## Current State

The log view (`internal/ui/states/log_*.go`) renders a simple list:

```
> a4c3a8a Fix memory leak - John Doe (4 min ago)
  e5b6d7f Add tests - Jane Doe (2 hours ago)
  ...
```

Key characteristics:
- Navigation (j/k) is synchronous—cursor updates immediately with no data fetching
- Pressing Enter triggers async `git.FetchFileChanges()` and transitions to FilesState
- View receives terminal width/height via `Context` interface
- No loading indicators during navigation

The files view (`internal/ui/states/files_view.go`) already has rendering code for:
- Commit header (hash, author, time, stats)
- Commit message (subject + body)
- File list with status indicators and line counts

The diff view (`internal/ui/states/diff_view.go`) demonstrates side-by-side layout using `lipgloss.JoinHorizontal`.

## Solution

### Layout Strategy

```
┌─────────────────── Log ────────────────────┬────── Details ──────┐
│ > a4c3a8a Fix memory leak - John Doe (4m)  │ Fix memory leak     │
│   e5b6d7f Add tests - Jane Doe (2h)        │                     │
│   ...                                      │ Fixes issue #42...  │
│                                            │ ─────────────────── │
│                                            │ M +17 -13 parser.go │
│                                            │ A +45 -0  tests.go  │
└────────────────────────────────────────────┴─────────────────────┘
```

> **Decision:** Give the details panel a fixed 80 characters; the log gets the remaining width.

Rationale:
- 80 characters gives the panel room for full file paths and commit message text
- The log can still display useful commit info in narrower space (hash, truncated message, author)
- Log gets `width - 80 - 3` (3 for separator " │ ")

> **Decision:** Display the panel only when terminal width >= 160 characters. Below this threshold, the log view renders exactly as it does today.

Rationale: With 80 chars for panel and 3 for separator, the log needs at least ~77 chars for readable commit lines. Setting the threshold at 160 ensures both sides have adequate space (log gets 77 chars).

### State Changes

Add a `Preview` field to `LogState` using a sum type to represent the preview panel state:

```go
type LogState struct {
    Commits       []git.GitCommit
    Cursor        int
    ViewportStart int
    Preview       PreviewState  // New: sum type for preview panel
}

// PreviewState variants
type PreviewNone struct{}
type PreviewLoading struct{ ForHash string }
type PreviewLoaded struct{ ForHash string; Files []git.FileChange }
type PreviewError struct{ ForHash string; Err error }
```

> **Decision:** Model preview state as a sum type rather than flat fields with boolean flags.

Rationale: The sum type makes states explicit and ensures all cases are handled. The `ForHash` field in each variant allows detecting stale responses when the user navigates faster than data loads.

Note: Commit message data (subject, body) comes directly from `Commits[Cursor]`—no need to store it in the preview state.

### Data Flow

```
User presses j/k
       │
       ▼
┌──────────────────────────────┐
│ Update cursor position       │
│ Set Preview = PreviewLoading │
│ Return loadPreview Cmd       │
└───────────┬──────────────────┘
            │
            ▼ (async)
┌──────────────────────────────┐
│ git.FetchFileChanges()       │
│ Return FilesPreviewLoadedMsg │
└───────────┬──────────────────┘
            │
            ▼
┌──────────────────────────────┐
│ Check hash matches cursor    │
│ If yes: Preview = Loaded/Err │
│ If no: discard stale message │
└──────────────────────────────┘
```

Cursor movement remains instant. The preview updates asynchronously. Reuses `git.FetchFileChanges()` which is already used for the files view transition.

### Panel Rendering

When terminal width meets the threshold, the view renders in split mode:
- Log list on the left (remaining width)
- Details panel on the right (fixed 80 chars)
- Joined with " │ " separator using `lipgloss.JoinHorizontal`

Preview panel content (top to bottom):
1. **Commit message**: Subject line, then body if present (wrapped to panel width)
2. **Separator**: Horizontal line
3. **File list**: Status, stats, path per file
4. **Overflow indicator**: "... and N more files" if files exceed available height

> **Decision:** Truncate file paths from the left (show filename + parent dir) rather than from the right.

Rationale: The filename is the most important part. `...onents/Button.tsx` is more useful than `src/components/Button....`

### Loading and Error States

The commit message section always renders immediately (data comes from `Commits[Cursor]`). The file list section depends on the `PreviewState` variant:

- **PreviewLoading**: Shows "Loading..." in place of file list
- **PreviewLoaded**: Shows file list
- **PreviewError**: Shows "Unable to load files" in place of file list

Errors are non-fatal—the log list continues to work normally.

### File Entry Format

```
M +17 -13 parser.go
A +45 -0  Button.tsx
D +0  -89 ...old/legacy.ts
```

Format: `Status +add -del  path`

- No selection indicator (panel is non-interactive)
- Stats right-aligned for visual alignment
- Path truncated from left with "..." prefix if needed

> **Decision:** Show full path but truncate from left when necessary.

Alternative considered: Always show just filename. Rejected because directory context matters when multiple files have the same name.

### Commit Message Display

- Subject line (first line) rendered prominently
- Body wrapped to panel width, truncated to ~5 lines with "..." indicator

> **Decision:** Limit commit body to 5 lines with truncation indicator.

Rationale: The panel should give a quick overview, not replace `git show`. Users who want the full message can press Enter to go to the files view.

## Open Questions

None. All design decisions have been made and documented above.
