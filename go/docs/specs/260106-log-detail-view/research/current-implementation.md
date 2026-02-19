# Current Detail View Implementation

Research into the log view detail panel (right pane shown in split view).

## Rendering Structure

The detail view is rendered in `internal/ui/states/log_view.go`:

1. **Entry**: `LogState.View()` checks terminal width (≥160 for split view)
2. **Split composition**: `renderSplitView()` builds two columns:
   - Left: commit list (80 chars)
   - Right: detail panel (80 chars)
   - Separator: " │ " (3 chars)

## Detail Panel Content Order

From `renderDetailsPanel()` (log_view.go:149-179):

```
1. Metadata line (if PreviewLoaded for current commit)
2. Blank line (if metadata present)
3. Subject line (always, truncated to panel width)
4. Blank line (if body exists)
5. Body lines (max 5 lines, wrapped to panel width)
6. Separator line (─────────...)
7. File list (based on Preview state)
8. Padding (blank lines to fill height)
```

## Metadata Line

Rendered by `RenderCommitMetadata()` in `commit_render.go:15-42`:
- Format: `abc123d · John Doe committed 2 hours ago · 3 files · +45 -12`
- Only shown when `PreviewLoaded.ForHash` matches current commit
- Truncated with "..." if exceeds panel width

## Separator Line

From `log_view.go:171`:
```go
separator := strings.Repeat("─", width)
lines = append(lines, styles.HeaderStyle.Render(separator))
```
- Unicode dash character `─` repeated to panel width
- Styled with `HeaderStyle` (gray)

## Refs Handling

**Current state**: Refs are NOT shown in detail panel
- Refs are displayed in left commit list
- Complex formatting logic exists in `log_line_format.go`
- Detail panel ignores refs entirely (no code to display them)

## Preview State Flow

State machine in `log_state.go:51-62`:
```
PreviewNone → PreviewLoading → PreviewLoaded/PreviewError
```

Update flow in `log_update.go`:
1. User navigates (j/k/g/G keys)
2. Cursor updates
3. Preview set to `PreviewLoading{ForHash}`
4. Async command `loadPreview()` dispatched
5. Result returns as `FilesPreviewLoadedMsg`
6. Preview updated to `PreviewLoaded` or `PreviewError`
7. View re-renders

## Constants

From `log_view.go:14-19`:
```go
splitPanelWidth    = 80   // Fixed width for details panel
splitThreshold     = 160  // Minimum terminal width
separatorWidth     = 3    // " │ "
commitBodyMaxLines = 5    // Max body lines shown
```

## Potential Flickering Sources

1. **Preview state transitions**: `PreviewLoading` → `PreviewLoaded` causes re-renders
2. **Metadata loading timing**: Metadata line appears only after preview loads
3. **ViewBuilder rebuilds**: Both columns rebuilt on every render
4. **Stale response detection**: Race conditions with rapid navigation

## Key Files

- `internal/ui/states/log_view.go` - Main rendering logic
- `internal/ui/states/log_state.go` - State definitions
- `internal/ui/states/log_update.go` - State transitions
- `internal/ui/states/commit_render.go` - Metadata formatting
- `internal/ui/states/viewbuilder.go` - Split view composition
