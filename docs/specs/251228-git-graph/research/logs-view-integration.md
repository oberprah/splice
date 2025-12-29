# Research: Logs View Integration Points

## Overview

This document identifies the specific integration points in Splice's logs view where git graph functionality will be added.

## Current Rendering Flow

```
LogState.View(ctx)
    ↓
    if ctx.Width() < 160:
        renderSimpleView(ctx)
    else:
        renderSplitView(ctx)
    ↓
formatCommitLine(commit, isSelected, width, ctx)
    ↓
Styled output: [Selector][Hash][Message][Author][Time]
```

## Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `internal/ui/states/log_state.go` | State struct definition | ~46 |
| `internal/ui/states/log_view.go` | Rendering logic | ~398 |
| `internal/ui/states/log_update.go` | Event handling | ~154 |
| `internal/git/git.go` | Git data fetching | ~200 |
| `internal/ui/styles/styles.go` | Color palette | ~100 |

## Integration Point 1: GitCommit Struct

**File**: `internal/git/git.go`

**Current**:
```go
type GitCommit struct {
    Hash    string
    Message string
    Body    string
    Author  string
    Date    time.Time
}
```

**Required Changes**:
- Add `ParentHashes []string`
- Add `Refs []RefInfo`

## Integration Point 2: FetchCommits()

**File**: `internal/git/git.go`

**Current Git Command**:
```bash
git log --pretty=format:%H%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

**Required Changes**:
- Add `%P` for parent hashes
- Add `%d` for decorations
- Parse additional fields

## Integration Point 3: LogState

**File**: `internal/ui/states/log_state.go`

**Current**:
```go
type LogState struct {
    Commits       []git.GitCommit
    Cursor        int
    ViewportStart int
    Preview       PreviewState
}
```

**Required Changes**:
- Add `GraphLayout *graph.Layout` (computed once, reused for rendering)

## Integration Point 4: formatCommitLine()

**File**: `internal/ui/states/log_view.go` (lines 99-169)

**Current Output Structure**:
```
[Selector 2ch][Hash 7ch][Message var][Author var][Time var]
```

**Modified Output Structure**:
```
[Graph var][Selector 2ch][Hash 7ch][Refs var][Message var][Author var][Time var]
```

**Width Calculation Changes**:
```go
// Current
availableWidth = max(width - fixedWidth, 10)

// Modified
graphWidth := calculateGraphWidth(layout)
availableWidth = max(width - fixedWidth - graphWidth - 1, 10)
```

## Integration Point 5: renderSimpleView() / renderSplitView()

**File**: `internal/ui/states/log_view.go`

**Changes Needed**:
- Pass graph layout to formatCommitLine()
- Adjust width calculations for graph column

**Split View Specific**:
```go
// Current
logWidth := ctx.Width() - splitPanelWidth - separatorWidth

// Modified - graph width must fit in log panel
graphWidth := min(calculateGraphWidth(layout), maxGraphWidth)
logWidth := ctx.Width() - splitPanelWidth - separatorWidth
// formatCommitLine receives (logWidth - graphWidth - 1)
```

## Integration Point 6: Styles

**File**: `internal/ui/styles/styles.go`

**Current Palette**:
| Element | Light | Dark |
|---------|-------|------|
| Hash | 172 | 214 |
| Message | 237 | 252 |
| Author | 36 | 86 |
| Time | 242 | 244 |

**New Styles Needed**:
- Branch colors (6-8 distinct colors for different branches)
- Ref decoration styles (tag vs branch vs HEAD)

## Width Budget Analysis

### Simple View (80 char terminal)

```
Current:
  Selector: 2
  Hash: 7
  Space: 1
  Separator: 3
  Time prefix: 1
  Time: ~12
  Total fixed: ~26
  Available for message+author: ~54

With Graph (assuming 6 chars):
  Graph: 6
  Space: 1
  Total fixed: ~33
  Available for message+author: ~47
```

### Split View (160 char terminal)

```
Current:
  Log panel: 77 chars (160 - 80 - 3)

With Graph (assuming 10 chars max):
  Graph: 10
  Space: 1
  Available for commit line: 66 chars
```

## Testing Impact

**Files**:
- `internal/ui/states/log_view_test.go`
- `internal/ui/states/testdata/log_view/*.golden`

**Required Updates**:
- Golden files will need to include graph column
- New test cases for:
  - Linear history (simple graph)
  - Merge commits
  - Multiple branches
  - Ref decorations

## Message Flow (if async loading needed)

Current async pattern:
1. User action → tea.Cmd
2. Cmd executes → returns message
3. Update() processes message

**Recommendation**: Compute graph layout synchronously during initial commit load, not async. Graph computation is fast enough (< 10ms for 500 commits).
