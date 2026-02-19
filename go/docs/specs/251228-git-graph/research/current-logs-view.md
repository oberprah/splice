# Research: Current Logs View Implementation

## Overview

The logs view in Splice displays git commit history in a terminal-based list format. This document captures the current implementation to inform the git graph feature design.

## Current Display Format

Each commit line shows:
```
> a4c3a8a Fix memory leak - John Doe (4 min ago)
```

**Components:**
- Selection indicator: `"> "` (selected) or `"  "` (unselected)
- Commit hash: 7-character short hash (amber/yellow color)
- Commit message: Subject line only (gray text)
- Separator: ` - ` literal
- Author name: Author only, no email (cyan color)
- Relative time: Human-readable time ago format (gray text)

## Display Modes

1. **Simple View** (terminal width < 160 chars): Traditional single-column list
2. **Split View** (terminal width ≥ 160 chars):
   - Log list on left (80 chars width)
   - Details panel on right (80 chars width)
   - 3-char separator between panels

## Data Structures

### GitCommit
```go
type GitCommit struct {
    Hash    string    // Full 40-character hash
    Message string    // First line (subject) of commit message
    Body    string    // Everything after first line
    Author  string    // Author name only (no email)
    Date    time.Time // Commit timestamp
}
```

### FileChange
```go
type FileChange struct {
    Path      string // File path relative to repository root
    Status    string // Git status: M, A, D, R, etc.
    Additions int    // Lines added
    Deletions int    // Lines deleted
    IsBinary  bool   // True for binary files
}
```

### LogState
```go
type LogState struct {
    Commits       []git.GitCommit  // All loaded commits
    Cursor        int              // Index of selected commit
    ViewportStart int              // Starting line of visible area
    Preview       PreviewState     // One of: PreviewNone, PreviewLoading, PreviewLoaded, PreviewError
}
```

## Git Commands Used

1. **Fetch commits**:
```bash
git log --pretty=format:%H%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

2. **Fetch file changes**:
```bash
git diff-tree --no-commit-id --name-status -r <commitHash>
git diff-tree --no-commit-id --numstat -r <commitHash>
```

3. **Fetch file content**:
```bash
git show <commitHash>:<filePath>
```

4. **Fetch unified diff**:
```bash
git show <commitHash> --format= -- :(top)<filePath>
```

## File Structure

**Location:** `internal/ui/states/`

- `log_state.go` (46 lines): Struct definitions
- `log_view.go` (398 lines): Rendering logic
- `log_update.go` (154 lines): Event handling
- `commit_render.go` (176 lines): Shared formatting utilities

## UI Patterns

### Elm Architecture
- States implement `State` interface with `View()` and `Update()` methods
- Message-driven async data loading
- Context interface provides terminal dimensions and data fetching

### Navigation
- Vim-style keys: j/k, g/G, Enter
- Viewport management keeps cursor visible
- Cursor position tracked in state

### Async Loading Pattern
1. User action → return `tea.Cmd`
2. Cmd executes async → returns message
3. Message processed in `Update()` → potentially transitions states

### Styling
- `lipgloss.Style` for colors and formatting
- Adaptive light/dark terminal colors
- Color palette:
  - Amber: hash
  - Cyan: author
  - Green: additions
  - Red: deletions
  - Gray: message/time

## Key Architectural Observations

1. **Pure State Machine**: No side effects outside `tea.Cmd` closures
2. **Responsive UI**: Adapts to terminal width with threshold-based branching
3. **Eager Preview Loading**: Files preview loads on cursor movement
4. **Single Responsibility**: Each state owns its view and update logic
5. **Reusable Components**: Shared rendering functions used by multiple states
6. **Context Pattern**: Loose coupling via interface

## Implications for Git Graph Feature

The current architecture would support a git graph by:
- Adding graph data to commit structure
- Extending render functions with ASCII art line drawing
- Maintaining the responsive two-column layout
- Using existing async loading patterns if needed
- Following established color palette and styling patterns
