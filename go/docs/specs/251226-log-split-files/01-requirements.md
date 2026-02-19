# Requirements: Log View Split Files Panel

## Problem Statement

When viewing the commit log, users currently have to press Enter and navigate to a separate screen to see which files changed in a commit. This requires extra navigation steps to get a quick overview of commit contents.

## Goals

- **Quick file preview**: Show changed files alongside the commit log without requiring navigation
- **Commit context**: Display the full commit message for better understanding
- **Non-intrusive**: Only show the panel when there's sufficient screen space
- **Responsive navigation**: File preview must not slow down commit navigation

## Non-Goals

- Interactive file selection from the split panel
- Replacing the existing full-screen files view
- Keyboard navigation within the files panel
- Showing file diffs in the split view

## User Impact

Users with wide terminals will see a panel on the right side of the log view displaying the commit message and changed files for the currently selected commit. This provides immediate context about what each commit contains without interrupting the browsing flow.

## Key Requirements

### Functional Requirements

1. **Split layout**: When terminal width is sufficient, display the log on the left and a details panel on the right

2. **Dynamic content**: The panel shows details for the currently highlighted commit, updating as the user navigates up/down through the log

3. **Commit message display**: Show the commit message similar to the files view:
   - Subject line (first line of commit message)
   - Body (remaining lines, if present)

4. **File information**: Each file entry displays:
   - File path (truncated if needed)
   - Change type indicator (added/modified/deleted/renamed)
   - Line statistics (lines added/removed, e.g., `+10 -5`)

5. **Overflow handling**: When a commit has more files than fit in the panel height, show as many as fit with an indicator at the bottom (e.g., "... and 12 more files")

6. **Loading state**: Show a loading indicator in the panel while file data is being fetched

7. **Existing behavior preserved**: Pressing Enter continues to navigate to the full-screen files view (current behavior unchanged)

### Non-Functional Requirements

1. **Non-blocking loading**: File data is fetched eagerly when the cursor moves, but fetching must not block navigation. Users can continue moving through commits; the panel updates when data arrives.

2. **Width threshold**: The split view only appears when there's enough width for the log to display comfortably. The panel has a fixed width; the log uses remaining width.

3. **Graceful degradation**: On narrow terminals, the view behaves exactly as it does today (no panel)

## Visual Mockup

### Wide terminal (split view active)

```
┌─ Log ─────────────────────────────────────────────────────────┬─ Details ────────────────────────┐
│                                                               │                                  │
│   a1b2c3d Add user authentication - Jane Doe (2 hours ago)    │  Add GitHub Actions CI pipeline  │
│   e4f5g6h Fix memory leak in parser - John Smith (5 hours)    │                                  │
│ > ac5fb5e Add GitHub Actions CI pipeline - Alice (1 day ago)  │  This commit adds a CI pipeline  │
│   a9026b7 Add implementation tracking doc - Bob (1 day ago)   │  that runs tests and linting on  │
│   eb8d1e3 Use Go tool directive - Carol (2 days ago)          │  every push and PR.              │
│   c491b99 Add GitHub Actions CI pipeline - Dave (2 days ago)  │                                  │
│   7434605 Add GitHub Actions pipeline design - Eve (3 days)   │  ─────────────────────────────── │
│                                                               │                                  │
│                                                               │  M .github/workflows/ci.yml +120 │
│                                                               │  A docs/ci-pipeline.md       +45 │
│                                                               │  M go.mod                     +2 │
│                                                               │  M internal/config.go        +15 │
│                                                               │                                  │
│                                                               │  ... and 3 more files            │
│                                                               │                                  │
└───────────────────────────────────────────────────────────────┴──────────────────────────────────┘
```

### While loading

```
┌─ Log ─────────────────────────────────────────────────────────┬─ Details ────────────────────────┐
│                                                               │                                  │
│   a1b2c3d Add user authentication - Jane Doe (2 hours ago)    │  Loading...                      │
│ > e4f5g6h Fix memory leak in parser - John Smith (5 hours)    │                                  │
│   ac5fb5e Add GitHub Actions CI pipeline - Alice (1 day ago)  │                                  │
│   ...                                                         │                                  │
└───────────────────────────────────────────────────────────────┴──────────────────────────────────┘
```

### Narrow terminal (no split view)

```
┌─ Log ──────────────────────────────────────────────────────────────────┐
│                                                                        │
│   a1b2c3d Add user authentication - Jane Doe (2 hours ago)             │
│ > e4f5g6h Fix memory leak in parser - John Smith (5 hours ago)         │
│   ac5fb5e Add GitHub Actions CI pipeline - Alice (1 day ago)           │
│   ...                                                                  │
└────────────────────────────────────────────────────────────────────────┘
```

### Change type indicators

- `A` = Added (green)
- `M` = Modified (yellow)
- `D` = Deleted (red)
- `R` = Renamed (blue/cyan)

## Open Questions for Design Phase

- What is the minimum comfortable width for the log portion?
- How should the panel width be allocated (fixed, percentage, or remaining space)?
- Visual styling: border/separator between panels, header for panel?
- Should file paths show full path or just filename?
- How to handle very long commit messages (truncate, wrap, scroll)?

## References

- Current log view implementation: `internal/ui/states/log_*.go`
- Files view header format: `internal/ui/states/files_view.go:44-91`
