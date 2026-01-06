# Current Architecture Research

## High-Level Overview

Splice is a terminal-based git diff viewer (~14k lines Go) using Bubbletea (Elm Architecture).

```
main.go → ui.NewModel() → Bubbletea Program
```

## Package Structure

| Package | Responsibility |
|---------|----------------|
| `internal/ui` | Main Model, Init/Update/View |
| `internal/ui/states` | State machine (5 states) |
| `internal/ui/messages` | Message definitions |
| `internal/ui/styles` | UI styling |
| `internal/ui/format` | Hash/time formatting |
| `internal/git` | Git command execution & parsing |
| `internal/diff` | Diff parsing & alignment |
| `internal/graph` | Commit graph visualization |
| `internal/highlight` | Syntax highlighting |

## State Machine

```
LoadingState → LogState → FilesState → DiffState
                  ↑______________|__________|
                  (q/esc returns to previous)
```

### States

| State | Purpose | Key Data |
|-------|---------|----------|
| `LoadingState` | Initial load | (stateless) |
| `LogState` | Commit log | Commits, Cursor, GraphLayout, Preview |
| `FilesState` | Files in commit | Files, Cursor + saved LogState data |
| `DiffState` | Side-by-side diff | Diff, Alignments + saved FilesState data |
| `ErrorState` | Error display | Err |

### State Interface (`internal/ui/states/state.go:25-33`)

```go
type State interface {
    View(ctx Context) *ViewBuilder
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
```

### Context Interface (`internal/ui/states/state.go:16-23`)

```go
type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}
```

## Message Flow (Elm Architecture)

1. **Init**: `Model.Init()` → `fetchCommits()` → `CommitsLoadedMsg`
2. **Update**: `Model.Update(msg)` → `currentState.Update(msg, ctx)` → `(newState, cmd)`
3. **View**: `Model.View()` → `currentState.View(ctx)` → rendered string

### Messages (`internal/ui/messages/messages.go`)

| Message | Triggers Transition |
|---------|---------------------|
| `CommitsLoadedMsg` | Loading → Log |
| `FilesPreviewLoadedMsg` | Updates LogState preview |
| `FilesLoadedMsg` | Log → Files |
| `DiffLoadedMsg` | Files → Diff |

## State Preservation Pattern

Each state stores data needed to restore previous state on back-navigation:

- `FilesState` stores: `ListCommits`, `ListCursor`, `ListViewportStart`
- `DiffState` stores: `FilesCommit`, `FilesFiles`, `FilesCursor`, etc.

## Dependency Injection (`internal/ui/app.go`)

Functional options pattern:

```go
NewModel(
    WithFetchCommits(fn),
    WithFetchFileChanges(fn),
    WithFetchFullFileDiff(fn),
    WithNow(fn),
)
```

Enables testing with mocks, real git functions as defaults.

## Key Subsystems

### Git Integration (`internal/git/git.go`)

- `FetchCommits(limit)` - Get commit log
- `FetchFileChanges(commitHash)` - Get files in commit
- `FetchFullFileDiff(commitHash, change)` - Get full diff

### Diff Processing (`internal/diff/`)

Pipeline: `ParseUnifiedDiff` → `BuildFileContent` (with syntax tokens) → `BuildAlignments`

Alignment types (sum type):
- `UnchangedAlignment` - Line in both versions
- `ModifiedAlignment` - Line changed (with inline diff)
- `RemovedAlignment` - Line only in old
- `AddedAlignment` - Line only in new

### Graph Visualization (`internal/graph/`)

Computes ASCII graph layout for commit history display.

### View Rendering (`internal/ui/states/viewbuilder.go`)

`ViewBuilder` utility for building views. Split view for wide terminals (≥160 chars).

## File Organization per State

```
internal/ui/states/
├── {state}_state.go   # Struct definition
├── {state}_view.go    # View() method
└── {state}_update.go  # Update() method
```

## Design Patterns Used

1. **State Pattern** - Each screen is a State implementation
2. **Message/Command Pattern** - Elm Architecture for async
3. **Sum Types** - Sealed interfaces (Alignment, PreviewState)
4. **Functional Options** - DI for Model
5. **Viewport/Cursor Pattern** - Universal scrolling
