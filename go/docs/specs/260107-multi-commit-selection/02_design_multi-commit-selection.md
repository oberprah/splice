# Design: Multi-Commit Selection

## Executive Summary

Multi-commit selection adds a "visual mode" to LogState, allowing users to select a range of commits and view their combined diff. Two key type changes enable this:

1. **`CursorState` sum type** - Replaces the `Cursor int` field. Either `CursorNormal{Pos}` or `CursorVisual{Pos, Anchor}`, making illegal states unrepresentable.

2. **`CommitRange` type** - Replaces single `Commit` in navigation messages. Contains `Start` and `End` commits (same for single commit). All downstream components (FilesState, DiffState, git functions) work with ranges.

The detail panel reuses existing components (`CommitInfo`, `FileSection`) for both single commits and ranges. Git functions are unified to always use range syntax.

## Context & Problem Statement

Splice currently only allows viewing one commit at a time. Users reviewing feature branches or understanding a series of related changes must either view each commit individually (tedious) or fall back to external tools like `git diff branch1..branch2`.

**Goal:** Enable users to select a range of commits in LogState and view their combined diff.

**Scope:** Selection interaction in LogState, visual feedback, data flow to FilesState/DiffState. Does not cover multi-file selection within FilesState.

## Current State

### Component Structure

```
LogState
├── Commit list (left panel)
│   └── FormatCommitLine() renders each line
└── Detail panel (right panel, split view only)
    ├── CommitInfo component (hash, author, time, refs, subject, body)
    └── FileSection component (stats + file list)

FilesState
├── CommitInfo component (same as above)
└── FileSection component (with cursor for navigation)

DiffState
└── Diff rendering (file content with changes)
```

`CommitInfo` and `FileSection` are shared components in `internal/ui/components/`.

### Navigation Messages

```go
type PushFilesScreenMsg struct {
    Commit git.GitCommit
    Files  []git.FileChange
}
```

### Git Functions

`FetchFileChanges(commitHash)` internally compares `commitHash^..commitHash` - already range syntax.

## Solution

### Key Types

#### CommitRange

> **Decision:** Introduce `CommitRange` to represent either a single commit or a range. All downstream components work with this type.
>
> **Rationale:** Unifies the data model. FilesState, DiffState, and git functions don't need separate code paths for single vs. range.

```go
// CommitRange represents either a single commit or a range.
// Start is always the older commit, End is the newer commit.
// For single commits, Start.Hash == End.Hash.
type CommitRange struct {
    Start git.GitCommit
    End   git.GitCommit
}
```

#### CursorState

> **Decision:** Replace `Cursor int` with a `CursorState` sum type combining position and optional anchor.
>
> **Rationale:** Makes it impossible to have an anchor without visual mode, or visual mode without an anchor.

```go
type CursorState interface {
    Position() int
}

type CursorNormal struct {
    Pos int
}

type CursorVisual struct {
    Pos    int
    Anchor int
}
```

#### LineDisplayState (for rendering)

> **Decision:** Use an enum for commit line display state instead of multiple booleans.
>
> **Rationale:** Three booleans would create 8 possible states, but only 4 are valid. An enum makes illegal states unrepresentable.

```go
type LineDisplayState int

const (
    LineStateNone         LineDisplayState = iota  // Not selected, not cursor
    LineStateCursor                                 // Normal mode cursor (→)
    LineStateSelected                               // Visual mode selected (▌)
    LineStateVisualCursor                           // Visual mode cursor (█)
)
```

Used in `CommitLineComponents` (replaces `IsSelected bool`), passed to `FormatCommitLine`.

### State Changes

```go
type State struct {
    Commits       []git.GitCommit
    Cursor        CursorState      // CursorNormal or CursorVisual
    ViewportStart int
    Preview       PreviewState
    GraphLayout   *graph.Layout
}
```

### Selection Behavior

| Current State | Key | Result |
|---------------|-----|--------|
| `CursorNormal{Pos: n}` | `v` | `CursorVisual{Pos: n, Anchor: n}` |
| `CursorNormal{Pos: n}` | `Enter` | View single commit diff |
| `CursorNormal{Pos: n}` | `j/k` | Update Pos, load preview |
| `CursorVisual{Pos: m, Anchor: n}` | `j/k` | Update Pos, selection = Anchor..Pos |
| `CursorVisual{Pos: m, Anchor: n}` | `Enter` | View combined diff of range |
| `CursorVisual{Pos: m, Anchor: n}` | `Escape` or `v` | `CursorNormal{Pos: m}` |

### Visual Feedback

**Selector characters:**

| State | Selector |
|-------|----------|
| `LineStateNone` | `  ` |
| `LineStateCursor` | `→ ` |
| `LineStateSelected` | `▌ ` |
| `LineStateVisualCursor` | `█ ` |

> **Decision:** Change normal cursor from `> ` to `→ `.
>
> **Rationale:** Visual consistency - all indicators are single-character symbols.

### Commit Info Display (Split View + FilesState)

Both the split view detail panel and FilesState use `CommitInfo` + `FileSection` components. Both need to handle `CommitRange`.

> **Decision:** Components show range info when `CommitRange` spans multiple commits.
>
> **Rationale:** Unified display. Users see consistent information whether previewing in split view or viewing in FilesState.

**Single commit:**
```
fc5a4c3 · author committed 9 hours ago
main, origin/main

Fix authentication bug
...

2 files · +39 -51
M +10 -10  internal/auth/login.go
M +29 -41  internal/auth/session.go
```

**Range (multiple commits):**
```
abc123d..def456e (3 commits)

5 files · +120 -45
M +45 -12  internal/auth/login.go
M +50 -20  internal/ui/theme.go
...
```

For ranges: no subject/body (ranges don't have a single message), just the range identifier and file stats.

> **Decision:** Debounce preview loading in split view during navigation.
>
> **Rationale:** Prevents excessive loading during rapid cursor movement. Exact delay TBD during implementation.

### Navigation Data Flow

**Updated message:**

```go
type PushFilesScreenMsg struct {
    Range CommitRange
    Files []git.FileChange
}
```

**Commit ordering:** When creating `CommitRange`, commits are normalized to chronological order (older..newer) regardless of selection direction. This ensures `git diff` produces intuitive results (additions shown as additions).

### Git Functions

> **Decision:** Unify git functions to use range syntax. Single commits become `parent..commit`.
>
> **Rationale:** Existing code already uses parent reference internally. Unifying eliminates parallel code paths.

**Updated signatures:**

```go
func FetchFileChanges(fromHash, toHash string) ([]FileChange, error)
func FetchFullFileDiff(fromHash, toHash string, change FileChange) (*FullFileDiffResult, error)
```

### Component Updates

**CommitInfo:** Add `CommitInfoFromRange(CommitRange, ...)` that delegates to existing `CommitInfo` for single commits, or renders range header for ranges.

**FileSection:** No changes needed - already works with `[]FileChange`.

**FormatCommitLine:** Accept `LineDisplayState` enum instead of `IsSelected` bool.

## Design Decisions Summary

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Cursor state | `CursorState` sum type | Makes illegal states unrepresentable |
| Commit representation | `CommitRange` for all flows | Unified model, no separate code paths |
| Line display state | `LineDisplayState` enum | Enum prevents invalid boolean combinations |
| Detail panel | Reuse CommitInfo + FileSection | Shows affected files, no new components |
| Git functions | Unified range syntax | Single code path |
| Commit ordering | Normalized in CommitRange creation | Intuitive diffs |

## Open Questions

None.

## References

- Requirements: `01_requirements_multi-commit-selection.md`
- Shared components: `internal/ui/components/commit_info.go`, `file_section.go`
- Navigation: `internal/core/navigation.go`
