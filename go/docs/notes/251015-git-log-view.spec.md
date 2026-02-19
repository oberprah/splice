# Git Log View Specification

**Date**: 2025-10-15
**Status**: Draft
**Feature**: Git Log List View

## Overview

Implement a minimal git log viewer as the first major feature of Splice. The view displays a scrollable list of commits with essential information in a clean, distraction-free interface.

## UI Design

### Layout

Minimal design with no header or footer. Each commit displayed as a single line:

```
  a4c3a8a Fix memory leak in parser - John Doe (4 min ago)
> d807e4d Add authentication middleware - Jane Smith (2 hours ago)
  aede6d9 Update dependencies to latest - Bob Wilson (3 hours ago)
```

**Format**: `hash message - author (time ago)`

**Selection Indicator**: `>` prefix on selected line

### Color Scheme (Lip Gloss)

- **Hash**: Amber/Yellow (`lipgloss.Color("214")`)
- **Message**: Bright white/default (`lipgloss.Color("252")`)
- **Author**: Cyan (`lipgloss.Color("86")`)
- **Time**: Muted gray (`lipgloss.Color("244")`)
- **Selected line**: Inverted/highlighted background (`lipgloss.Color("255")` bg, `lipgloss.Color("0")` fg)

### Truncation Rules

When terminal width is limited:
1. Hash always visible (7 chars)
2. Message truncates with "..." if needed
3. Author truncates if critically short
4. Time shows in parentheses (shortest form)

Priority: hash > message > author > time

### Loading and Error States

**Loading**:
```
  Loading commits...
```

**Error**:
```
  Error: not a git repository
```

## Data Model

### Core Structures

```go
type GitCommit struct {
    Hash         string    // Full 40-char hash
    ShortHash    string    // First 7 chars
    Message      string    // First line of commit message
    Author       string    // Author name (not email)
    Date         time.Time // Commit timestamp
    RelativeTime string    // "4 min ago", computed
}

type ViewState int
const (
    ListView ViewState = iota
    LoadingView
    ErrorView
)

type Model struct {
    // Data
    commits      []GitCommit

    // Navigation
    cursor       int
    viewportStart int
    viewportHeight int

    // View
    state        ViewState
    width        int
    height       int

    // Status
    loading      bool
    error        error
}
```

## Technical Implementation

### Git Data Source

**Decision**: Use `os/exec` with git commands (NOT go-git library)

**Rationale**:
- go-git is 10-180x slower than native git for most operations
- Direct access to battle-tested git implementation
- Simpler parsing of structured output

**Command**:
```bash
git log --pretty=format:"%H|%an|%ad|%s" --date=iso-strict -n 500
```

**Output format**: `hash|author|date|message` (pipe-separated)

### Relative Time Calculation

Smart relative times for recent commits, absolute dates for old ones:

- `< 1 min`: "just now"
- `< 1 hour`: "N min ago" / "N mins ago"
- `< 24 hours`: "N hour ago" / "N hours ago"
- `< 7 days`: "N day ago" / "N days ago"
- `< 30 days`: "N week ago" / "N weeks ago"
- `< 365 days`: "N month ago" / "N months ago"
- `>= 365 days`: "Jan 2, 2006" (absolute date)

### Viewport Management

- Track `viewportStart` (first visible commit index)
- Track `viewportHeight` (number of visible lines)
- Auto-scroll when cursor moves beyond visible area
- Render only visible commits for performance

### Navigation

**Keyboard Shortcuts**:
- `j` / `↓`: Move down
- `k` / `↑`: Move up
- `g`: Jump to top
- `G`: Jump to bottom
- `q` / `Ctrl+C`: Quit

**Future** (not in initial scope):
- `/`: Search
- `Enter`: View commit details
- `n`/`N`: Next/previous search result

## Implementation Phases

### Phase 1: Basic List (Current Scope)
- Load commits via git command
- Display formatted list
- Basic navigation (j/k, arrows)
- Cursor selection
- Color highlighting

### Phase 2: Enhancements (Future)
- Search/filter functionality
- Commit detail view
- Branch selection
- Date range filtering
- Author filtering

## File Structure

```
splice/
├── main.go           # Entry point
├── model.go          # Model struct, state
├── update.go         # Event handling
├── view.go           # Rendering logic
├── git/
│   └── git.go       # Git command execution
└── styles/
    └── styles.go    # Lip Gloss styles
```

## Success Criteria

- [ ] Displays git log commits in terminal
- [ ] Smooth scrolling with keyboard navigation
- [ ] Proper color highlighting
- [ ] Selected line clearly visible
- [ ] Relative timestamps display correctly
- [ ] Handles large repositories (500+ commits)
- [ ] Graceful error handling (not a git repo, etc.)
- [ ] Responsive to terminal resize

## Notes

- Initially load 500 commits (configurable)
- No pagination in Phase 1 (load all at once)
- Detailed commit view deferred to Phase 2
- Focus on minimal, fast, clean UX
