# Design: Architecture Restructure

## Executive Summary

This design restructures the codebase into clear architectural layers (`domain/`, `app/`, `ui/`) and introduces a navigation stack to decouple states from each other. The key insight is that states currently embed parent state data for back-navigation, creating tight coupling. With a navigation stack, the Model preserves previous states directly, eliminating this coupling and enabling states to live in separate packages without circular imports.

The State and Context interfaces move to `app/`, and states register themselves via init-time registration to avoid import cycles. Navigation happens through messages (`PushScreenMsg`, `PopScreenMsg`) that the Model intercepts, rather than states constructing each other directly.

Implementation proceeds in three phases: (1) navigation stack with factory registration, (2) package reorganization, (3) state subfolders and component extraction.

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
├── app/              # Application core
│   ├── state.go      # State, Context interfaces
│   ├── navigation.go # Navigation messages, Screen enum
│   ├── messages.go   # Domain messages (merged)
│   ├── model.go      # Model struct
│   └── factory.go    # State factory with registration
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

> **Decision:** Model maintains a stack of states. States signal navigation via messages, Model intercepts and manages transitions.

#### Stack Mechanism

```go
// app/model.go
type Model struct {
    stack        []State   // Previous states (preserved exactly)
    currentState State
    // ...existing fields
}
```

#### Navigation Messages

```go
// app/navigation.go
type Screen int

const (
    LogScreen Screen = iota
    FilesScreen
    DiffScreen
    ErrorScreen
)

type PushScreenMsg struct {
    Screen Screen
    Data   any  // Screen-specific data
}

type PopScreenMsg struct{}
```

#### Data Flow Diagram

```
┌───────────────────────────────────────────────────────────────┐
│                          Model                                │
│  ┌────────────────────────────────────────────────────────┐   │
│  │ stack: [LogState, FilesState]                          │   │
│  │ current: DiffState                                     │   │
│  └────────────────────────────────────────────────────────┘   │
│                              │                                │
│                              ▼                                │
│  ┌─────────────────────────────────────────────────────────┐  │
│  │                    Model.Update(msg)                    │  │
│  │  ┌──────────────────────────────────────────────────┐   │  │
│  │  │ switch msg.(type) {                              │   │  │
│  │  │ case PushScreenMsg:                              │   │  │
│  │  │     stack = append(stack, currentState)          │   │  │
│  │  │     currentState = createState(msg.Screen, data) │   │  │
│  │  │ case PopScreenMsg:                               │   │  │
│  │  │     currentState = stack[len-1]                  │   │  │
│  │  │     stack = stack[:len-1]                        │   │  │
│  │  │ default:                                         │   │  │
│  │  │     delegate to currentState.Update()            │   │  │
│  │  │ }                                                │   │  │
│  │  └──────────────────────────────────────────────────┘   │  │
│  └─────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

#### State Transitions (New)

States return commands that produce navigation messages:

```go
// states/log/update.go - NO import of files package
func (s *State) Update(msg tea.Msg, ctx app.Context) (app.State, tea.Cmd) {
    case messages.FilesLoadedMsg:
        // Return command that sends PushScreenMsg
        return s, func() tea.Msg {
            return app.PushScreenMsg{
                Screen: app.FilesScreen,
                Data: app.FilesScreenData{
                    Commit: msg.Commit,
                    Files:  msg.Files,
                },
            }
        }
}

// states/files/update.go - NO import of log or diff packages
func (s *State) Update(msg tea.Msg, ctx app.Context) (app.State, tea.Cmd) {
    case "q":
        // Return command that sends PopScreenMsg
        return s, func() tea.Msg {
            return app.PopScreenMsg{}
        }
}
```

#### State Factory (Registration Pattern)

> **Decision:** Use init-time registration to avoid Model importing state packages directly.

```go
// app/factory.go
var stateFactories = map[Screen]func(any) State{}

func RegisterStateFactory(screen Screen, factory func(any) State) {
    stateFactories[screen] = factory
}

func CreateState(screen Screen, data any) State {
    factory, ok := stateFactories[screen]
    if !ok {
        panic(fmt.Sprintf("no factory registered for screen %v", screen))
    }
    return factory(data)
}
```

```go
// states/log/state.go
func init() {
    app.RegisterStateFactory(app.LogScreen, func(data any) app.State {
        d := data.(app.LogScreenData)
        return New(d.Commits, d.GraphLayout)
    })
}
```

This breaks the import cycle:
- `app/` defines interfaces, doesn't import states
- State packages import `app/` and register at init
- `main.go` imports state packages (triggering init registration)
- Model uses `app.CreateState()` at runtime

### 3. Interface Locations

> **Decision:** Move State and Context interfaces to `app/` package.

Rationale: The Model implements Context and consumes State. Moving interfaces to `app/` makes the dependency direction clear: states depend on `app/`, not the other way around.

```go
// app/state.go
type State interface {
    View(ctx Context) *ViewBuilder
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}

type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}

type FetchFileChangesFunc func(commitHash string) ([]git.FileChange, error)
type FetchFullFileDiffFunc func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error)
```

### 4. Screen Data Types

> **Decision:** Define screen data types in `app/navigation.go` alongside Screen enum.

```go
// app/navigation.go
type LogScreenData struct {
    Commits     []git.GitCommit
    GraphLayout *graph.Layout
}

type FilesScreenData struct {
    Commit git.GitCommit
    Files  []git.FileChange
}

type DiffScreenData struct {
    Commit        git.GitCommit
    File          git.FileChange
    Diff          *diff.AlignedFileDiff
    ChangeIndices []int
}

type ErrorScreenData struct {
    Err error
}
```

### 5. Message Consolidation

> **Decision:** Merge `messages/messages.go` into `app/messages.go`.

Currently messages are in `internal/ui/messages/`. With the new structure, they belong in `app/` since they're part of application orchestration.

The navigation messages (`PushScreenMsg`, `PopScreenMsg`) go in `app/navigation.go`.
Domain messages (`FilesLoadedMsg`, `DiffLoadedMsg`, `FilesPreviewLoadedMsg`) go in `app/messages.go`.

Note: `CommitsLoadedMsg` is currently defined in `loading_update.go`. It should move to `app/messages.go` for consistency.

### 6. Special State Handling

#### LoadingState

> **Decision:** LoadingState is the initial state, never pushed to stack.

```go
func NewModel(opts ...ModelOption) Model {
    m := Model{
        currentState: loading.New(),  // Initial state
        stack:        nil,            // Empty stack
        // ...
    }
}
```

When LoadingState completes, it returns `PushScreenMsg{LogScreen, ...}`. The Model handles this specially: since there's nothing to push (LoadingState shouldn't be on stack), it replaces currentState directly.

```go
case PushScreenMsg:
    if len(m.stack) == 0 && isLoadingState(m.currentState) {
        // First transition from LoadingState - replace, don't push
        m.currentState = CreateState(msg.Screen, msg.Data)
    } else {
        // Normal push
        m.stack = append(m.stack, m.currentState)
        m.currentState = CreateState(msg.Screen, msg.Data)
    }
```

#### ErrorState

> **Decision:** ErrorState is pushed onto stack. User can go back.

If an error occurs while loading files, push ErrorState. User presses 'q' → PopScreenMsg → returns to previous state.

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
    states/log    states/files   states/diff ...
         │             │             │
         └─────────────┼─────────────┘
                       │
                       ▼
                     app/
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

- `main.go` imports all state packages (triggers init registration)
- State packages import `app/` (interfaces, navigation, messages)
- State packages import `ui/components/`, `ui/styles/`, `ui/format/`
- State packages import `domain/*` as needed
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

### 10. Migration Strategy

> **Decision:** Incremental migration in three phases.

**Phase 1: Navigation Stack**
1. Add stack to Model
2. Add navigation messages to existing `messages/` package
3. Add factory registration pattern
4. Update states one-by-one to use `PushScreenMsg`/`PopScreenMsg`
5. Remove preserved parent fields from state structs
6. All tests pass at each step

**Phase 2: Package Reorganization**
1. Create `internal/domain/` and move `diff/`, `graph/`, `highlight/`
2. Create `internal/app/` and move Model, messages, navigation
3. Update all import paths
4. Verify no circular imports

**Phase 3: State Subfolders + Components**
1. Create `ui/components/` and extract shared view utilities
2. Move each state to its subfolder (`states/log/`, etc.)
3. Rename files (`log_state.go` → `state.go`, etc.)

Rationale: Phase 1 is the riskiest (changes behavior). Doing it first while structure is familiar reduces risk. Phases 2-3 are mechanical refactors.

## Open Questions

None. All decisions have been made:

| Question | Decision |
|----------|----------|
| Message location | `app/messages.go` (domain), `app/navigation.go` (nav) |
| Context interface | `app/state.go` |
| Screen data types | `app/navigation.go` |
| Migration strategy | Incremental: stack → packages → subfolders |
