# Current Log View Implementation

## Structure

The log view follows the standard state pattern with three files:
- `log_state.go` - State struct definition
- `log_view.go` - Rendering logic (View method)
- `log_update.go` - Event handling (Update method)

## Log State Data Structure

```go
type LogState struct {
    Commits       []git.GitCommit  // List of commits from git log
    Cursor        int              // Index of selected commit
    ViewportStart int              // Starting index for visible viewport
}
```

Available commit data (`git.GitCommit`):
- `Hash` - Full 40-char commit hash
- `Message` - First line of commit message (subject)
- `Body` - Commit message body
- `Author` - Author name (not email)
- `Date` - Commit timestamp (time.Time)

## Rendering

Each commit renders as:
```
> a4c3a8a Fix memory leak - John Doe (4 min ago)
```

Key rendering aspects:
- Selection indicator: `"> "` for selected, `"  "` for unselected
- Dynamic truncation based on terminal width (2/3 for message, 1/3 for author)
- Styled components: hash, message, author, time each have Lip Gloss styles
- No header - just a compact list format

## Navigation

Vim-style navigation:
- `j/down` - Move cursor down
- `k/up` - Move cursor up
- `g` - Jump to first commit
- `G` - Jump to last commit
- `q/ctrl+c` - Quit

The `updateViewport()` function keeps the cursor visible by scrolling the viewport.

## Transition to Files View

When user presses `enter`:
1. Calls `git.FetchFileChanges(selectedCommit.Hash)` via tea.Cmd
2. Command executes async
3. Returns `FilesLoadedMsg` with commit data, files, and state preservation data
4. State transitions: LogState → FilesState

The original log state (commits, cursor, viewport) is preserved in FilesState for navigation back.

## Key Observations

- Pure state machine: each view is independent
- Async loading via Bubbletea Cmd pattern
- Context interface provides width/height for responsive rendering
- State preservation via messages enables back navigation
