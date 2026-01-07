# Design: Architecture Restructure

## Executive Summary

This design restructures the codebase into clear architectural layers (`domain/`, `app/`, `ui/`) and introduces a navigation stack to decouple states from each other. The key insight is that states currently embed parent state data for back-navigation, creating tight coupling. With a navigation stack, the Model preserves previous states directly, eliminating this coupling and enabling states to live in separate packages without circular imports.

The State and Context interfaces live in `core/` (a neutral package that breaks import cycles). Navigation happens through typed messages (`PushLogScreenMsg`, `PushFilesScreenMsg`, etc.) that the Model intercepts, rather than states constructing each other directly. The Model uses a single stack where the current state is always `stack[len-1]`.

Implementation proceeded in phases: (1) navigation stack with typed messages, (2) package reorganization, (3) state subfolders and component extraction, (4) stack simplification (removing `currentState` and `firstPush` fields).

## Context & Problem Statement

The current architecture has grown organically and now exhibits several structural issues:

1. **Flat hierarchy**: `git/`, `diff/`, `graph/`, `highlight/`, `ui/` all at the same level with no layering
2. **UI scope creep**: `internal/ui/` contains both presentation code and application orchestration (Model)
3. **State coupling**: States directly construct each other (e.g., `FilesState` creates both `LogState` and `DiffState`)
4. **Manual state preservation**: `DiffState` carries 7 fields just to recreate `FilesState` on back-navigation

This design addresses all four issues while preserving the Elm Architecture pattern.

## Current State

### Package Structure
```
internal/
├── git/           # ~480 lines - Git CLI adapter
├── diff/          # ~620 lines - Diff processing
├── graph/         # ~500 lines - Commit graph layout
├── highlight/     # ~130 lines - Syntax highlighting
└── ui/
    ├── app.go, model.go   # Application core
    ├── states/            # 20+ files, 5 states
    ├── messages/          # Message definitions
    ├── styles/            # Lipgloss styles
    └── format/            # Display formatting
```

### State Transitions (Current)

States directly construct each other and pass data via messages:

```go
// log_update.go - LogState creates FilesState
return &FilesState{
    Commit: msg.Commit,
    Files: msg.Files,
    // Must preserve LogState data for back-navigation
    ListCommits: msg.ListCommits,
    ListCursor: msg.ListCursor,
    ListViewportStart: msg.ListViewportStart,
}, nil

// files_update.go - FilesState creates LogState on back
return &LogState{
    Commits: s.ListCommits,  // Retrieved from preserved fields
    Cursor: s.ListCursor,
    // ...
}, nil
```

This creates import cycles if states are in separate packages.

## Solution

### 1. Package Hierarchy

> **Decision:** Keep `git/` at root level, group processing packages under `domain/`, separate application core from UI.

Rationale: `git/` is the external adapter and conceptually distinct. Domain processing packages (`diff/`, `graph/`, `highlight/`) are peers that can be grouped. Separating `app/` from `ui/` clarifies that Model orchestrates the application, not just UI.

```
internal/
├── git/              # External adapter (special case)
├── domain/           # Domain processing
│   ├── diff/
│   ├── graph/
│   └── highlight/
├── core/             # Shared interfaces (breaks import cycles)
│   ├── state.go      # State, Context, ViewRenderer interfaces
│   ├── navigation.go # Typed navigation messages (PushLogScreenMsg, etc.)
│   └── messages.go   # Domain messages (CommitsLoadedMsg, etc.)
├── app/              # Application core
│   ├── model.go      # Model struct (Bubbletea lifecycle)
│   ├── options.go    # Functional options (WithFetchCommits, etc.)
│   └── accessors.go  # Context interface implementation
└── ui/               # Presentation
    ├── states/       # State implementations
    │   ├── loading/
    │   ├── log/
    │   ├── files/
    │   ├── diff/
    │   └── error/
    ├── components/   # Shared view components
    ├── styles/
    └── format/
```

### 2. Navigation Stack

> **Decision:** Model maintains a single stack of states where current state is always `stack[len-1]`. States signal navigation via typed messages, Model intercepts and manages transitions.

#### Stack Mechanism

```go
// app/model.go
type Model struct {
    stack []core.State  // Current state is always stack[len-1]
    // ...other fields (width, height, fetchers)
}

// current returns the current state (top of stack)
func (m *Model) current() core.State {
    return m.stack[len(m.stack)-1]
}

// pushState adds a new state. LoadingState is transient - replaced, not stacked.
func (m *Model) pushState(newState core.State) {
    if _, isLoading := m.current().(loading.State); isLoading {
        m.stack[len(m.stack)-1] = newState  // Replace LoadingState
    } else {
        m.stack = append(m.stack, newState)  // Normal push
    }
}
```

#### Navigation Messages

Typed messages provide compile-time safety (no `any` type assertions):

```go
// core/navigation.go
type PushLogScreenMsg struct {
    Commits     []git.GitCommit
    GraphLayout *graph.Layout
    InitCmd     tea.Cmd
}

type PushFilesScreenMsg struct {
    Commit git.GitCommit
    Files  []git.FileChange
}

type PushDiffScreenMsg struct {
    Commit        git.GitCommit
    File          git.FileChange
    Diff          *diff.AlignedFileDiff
    ChangeIndices []int
}

type PushErrorScreenMsg struct {
    Err error
}

type PopScreenMsg struct{}
```

#### Data Flow Diagram

```
┌───────────────────────────────────────────────────────────────┐
│                          Model                                │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ stack: [LogState, FilesState, DiffState]               │   │
│  │         ↑                      ↑                       │   │
│  │      history              current()                    │   │
│  └────────────────────────────────────────────────────────┘   │
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    Model.Update(msg)                    │  │
│  │  ┌──────────────────────────────────────────────────┐   │  │
│  │  │ switch msg.(type) {                              │   │  │
│  │  │ case core.PushFilesScreenMsg:                    │   │  │
│  │  │     m.pushState(&files.State{...})               │   │  │
│  │  │ case core.PopScreenMsg:                          │   │  │
│  │  │     m.stack = m.stack[:len-1]                    │   │  │
│  │  │ default:                                         │   │  │
│  │  │     m.current().Update(msg, &m)                  │   │  │
│  │  │ }                                                │   │  │
│  │  └──────────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

#### State Transitions (New)

States return commands that produce typed navigation messages:

```go
// states/log/update.go - NO import of files package
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    case core.FilesLoadedMsg:
        // Return command that sends typed PushFilesScreenMsg
        return s, func() tea.Msg {
            return core.PushFilesScreenMsg{
                Commit: msg.Commit,
                Files:  msg.Files,
            }
        }
}

// states/files/update.go - NO import of log or diff packages
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return s, func() tea.Msg {
                return core.PopScreenMsg{}
            }
        }
}
```

#### Direct State Creation (No Factory)

> **Decision:** Model creates states directly in Update handlers. This is simpler than factory registration and still avoids import cycles because typed messages carry all data needed.

```go
// app/model.go - imports all state packages directly
case core.PushFilesScreenMsg:
    newState := &files.State{
        Commit: msg.Commit,
        Files:  msg.Files,
    }
    m.pushState(newState)
    return m, nil
```

Import cycle is broken by the `core` package:
- `core/` defines interfaces and messages (no imports of `app/` or states)
- `app/` imports `core/` and all state packages
- State packages import `core/` (not `app/`)

### 3. Interface Locations

> **Decision:** State and Context interfaces live in `core/` package.

Rationale: The `core/` package is a neutral ground that breaks import cycles. Both `app/` and state packages can import `core/` without creating cycles.

```go
// core/state.go
type State interface {
    View(ctx Context) ViewRenderer
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}

type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}

type ViewRenderer interface {
    String() string
}

type FetchFileChangesFunc func(commitHash string) ([]git.FileChange, error)
type FetchFullFileDiffFunc func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error)
```

### 4. Screen Data Types

> **Decision:** Screen data is embedded directly in typed navigation messages. No separate data types needed.

Each `Push*ScreenMsg` contains exactly the data needed for that screen:

```go
// core/navigation.go - data is part of the message
type PushLogScreenMsg struct {
    Commits     []git.GitCommit
    GraphLayout *graph.Layout
    InitCmd     tea.Cmd  // Command to start preview loading
}

type PushFilesScreenMsg struct {
    Commit git.GitCommit
    Files  []git.FileChange
}
// etc.
```

This is simpler than separate `*ScreenData` types - no indirection, no type assertions.

### 5. Message Consolidation

> **Decision:** Messages live in `core/` package, split by purpose.

- `core/navigation.go` - Navigation messages (`PushLogScreenMsg`, `PushFilesScreenMsg`, `PushDiffScreenMsg`, `PushErrorScreenMsg`, `PopScreenMsg`)
- `core/messages.go` - Domain messages (`CommitsLoadedMsg`, `FilesLoadedMsg`, `FilesPreviewLoadedMsg`, `DiffLoadedMsg`)

### 6. Special State Handling

#### LoadingState

> **Decision:** LoadingState is transient - it gets replaced instead of stacked.

LoadingState is set as the initial state in `main.go`:

```go
// main.go
initialModel := app.NewModel(
    app.WithInitialState(loading.State{}),
)
```

When any Push message arrives while LoadingState is current, Model replaces it instead of stacking:

```go
// app/model.go
func (m *Model) pushState(newState core.State) {
    // LoadingState is transient - replace instead of stacking
    if _, isLoading := m.current().(loading.State); isLoading {
        m.stack[len(m.stack)-1] = newState
    } else {
        m.stack = append(m.stack, newState)
    }
}
```

This ensures:
- LoadingState never appears in navigation history
- User can't "go back" to the loading screen
- Simple type check, no special flags needed

#### ErrorState

> **Decision:** ErrorState is pushed onto stack. User can go back.

If an error occurs while loading files, push ErrorState. User presses 'q' → PopScreenMsg → returns to previous state.

Exception: If error occurs during initial loading (LoadingState → Error), the ErrorState replaces LoadingState (same transient behavior).

### 7. Shared Components

> **Decision:** Extract `viewbuilder.go`, `commit_info.go`, `file_section.go` to `ui/components/`.

```
ui/components/
├── viewbuilder.go    # ViewBuilder type
├── commit_info.go    # CommitInfo() function
└── file_section.go   # FileSection() function
```

`log_line_format.go` stays with LogState (only user) or moves to `ui/components/` if we want consistency.

### 8. Dependency Graph (Final)

```
                    main.go
                       │
         ┌─────────────┼─────────────┐
         │             │             │
         ▼             ▼             ▼
       app/      states/loading   (initial state)
         │
         ├─────────────────────────────────────┐
         │             │             │         │
         ▼             ▼             ▼         ▼
    states/log   states/files  states/diff  states/error
         │             │             │         │
         └─────────────┴─────────────┴─────────┘
                       │
                       ▼
                     core/  ←─────────────────── app/ also imports
                       │
         ┌─────────────┼─────────────┐
         │             │             │
         ▼             ▼             ▼
   ui/components   ui/styles    ui/format
         │
         ▼
      domain/*
         │
         ▼
       git/
```

- `main.go` imports `app/` and `loading` (to set initial state)
- `app/` imports `core/` and all state packages (to create states in Update)
- State packages import `core/` (interfaces, messages) - NOT `app/`
- State packages import `ui/components/`, `ui/styles/`, `ui/format/`
- State packages import `domain/*` as needed
- `core/` is the neutral ground that breaks import cycles
- No circular imports

### 9. State Struct Changes

With the navigation stack, states no longer need to preserve parent data:

**Before (FilesState):**
```go
type FilesState struct {
    Commit        git.GitCommit
    Files         []git.FileChange
    Cursor        int
    ViewportStart int
    // Preserved LogState data (7 fields in DiffState!)
    ListCommits       []git.GitCommit
    ListCursor        int
    ListViewportStart int
}
```

**After:**
```go
type State struct {
    Commit        git.GitCommit
    Files         []git.FileChange
    Cursor        int
    ViewportStart int
    // No preserved parent data - stack handles this
}
```

**DiffState** goes from 14 fields to 7 fields.

### 10. Migration Strategy (Completed)

The migration was completed incrementally:

**Phase 1: Navigation Stack with Typed Messages**
1. Added stack to Model with `firstPush` flag for LoadingState handling
2. Created typed navigation messages (`PushLogScreenMsg`, etc.) instead of generic `PushScreenMsg{Screen, Data any}`
3. Created `core/` package to break import cycles (interfaces + messages)
4. Updated states to use typed messages via `core/`
5. Removed preserved parent fields from state structs

**Phase 2: Package Reorganization**
1. Created `internal/domain/` and moved `diff/`, `graph/`, `highlight/`
2. Created `internal/app/` with Model, options, accessors
3. Created `internal/core/` with State, Context interfaces and all messages
4. Updated all import paths

**Phase 3: State Subfolders + Components**
1. Created `ui/components/` and extracted shared view utilities
2. Moved each state to its subfolder (`states/log/`, etc.)
3. Renamed files (`log_state.go` → `state.go`, etc.)

**Phase 4: Stack Simplification**
1. Removed `currentState` field - current is now `stack[len-1]`
2. Removed `firstPush` flag - replaced with LoadingState type check
3. Added `current()` and `pushState()` helper methods
4. Split `model.go` into `model.go`, `options.go`, `accessors.go`

## Summary

| Aspect | Decision |
|--------|----------|
| Message location | `core/navigation.go` (nav), `core/messages.go` (domain) |
| Interfaces | `core/state.go` (State, Context, ViewRenderer) |
| Screen data | Embedded in typed messages (no separate types) |
| Stack model | Single `stack []core.State`, current = `stack[len-1]` |
| LoadingState | Transient - replaced via type check, not stacked |
| Import cycles | Broken by `core/` package as neutral ground |
