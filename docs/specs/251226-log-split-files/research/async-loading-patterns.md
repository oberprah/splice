# Async Data Loading Patterns

## Tea.Cmd and Message Pattern

The codebase uses the Elm Architecture pattern:

1. User action in `Update()` returns `(newState, tea.Cmd)`
2. `tea.Cmd` is a closure: `func() tea.Msg`
3. Command executes asynchronously
4. Message routes back to `Update()` on completion

Example from `log_update.go`:
```go
case "enter":
    return s, func() tea.Msg {
        fileChanges, err := git.FetchFileChanges(selectedCommit.Hash)
        return messages.FilesLoadedMsg{
            Commit:            selectedCommit,
            Files:             fileChanges,
            Err:               err,
            ListCommits:       s.Commits,
            ListCursor:        s.Cursor,
            ListViewportStart: s.ViewportStart,
        }
    }
```

Key characteristics:
- Commands are simple closures capturing current state
- State transitions only happen on message receipt
- No separate state management needed

## Loading States

Current implementation is minimal:
- `LoadingState` shows "Loading commits..." on startup
- File/diff loading has no loading indicator
- UI stays on current state until message arrives

## Eager Loading - Not Currently Implemented

Cursor movement (j/k keys) is synchronous with no data fetching:
```go
case "j", "down":
    if s.Cursor < len(s.Commits)-1 {
        s.Cursor++
        s.updateViewport(ctx.Height())
    }
    return s, nil  // No command, instant update
```

Data only loads when user presses Enter.

## Cancellation - Not Currently Implemented

No concurrency control mechanisms:
- No `context.Context` usage
- No goroutine management
- Commands execute sequentially
- Messages processed in order of arrival

## Implications for Split Panel

To implement non-blocking preview loading:

1. **New message type**: `FilesPreviewLoadedMsg`
2. **Trigger on cursor movement**: Return a Cmd from j/k handlers
3. **Store in LogState**: Add `PreviewFiles`, `PreviewLoading`, `PreviewErr`
4. **Last-write-wins**: New cursor position overwrites pending preview
5. **Match by hash**: Verify message matches current selection before updating

Example approach:
```go
// Add to LogState
PreviewFiles   []git.FileChange
PreviewLoading bool
PreviewForHash string  // Track which commit the preview is for

// On cursor movement
case "j", "down":
    if s.Cursor < len(s.Commits)-1 {
        s.Cursor++
        s.updateViewport(ctx.Height())
        s.PreviewLoading = true
        return s, s.loadPreview(s.Commits[s.Cursor])
    }
    return s, nil

// On preview loaded
case messages.FilesPreviewLoadedMsg:
    // Only update if this matches current selection
    if msg.Hash == s.Commits[s.Cursor].Hash {
        s.PreviewFiles = msg.Files
        s.PreviewLoading = false
    }
    return s, nil
```

This fits the existing architecture without new concurrency primitives.
