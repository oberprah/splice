# Requirements: Architecture Restructure

## Problem Statement

The current architecture has several issues:

1. **Flat package hierarchy** - Domain packages (`diff`, `graph`, `highlight`), data access (`git`), and UI (`ui`) all sit at the same level with no hierarchy expressing their relationships

2. **Unclear `ui/` scope** - The `ui/` package contains both presentation code and application orchestration (Model), making "UI" a misnomer

3. **Large `states/` folder** - 20+ files in a flat structure, with shared utilities mixed alongside state implementations

4. **Tight state coupling** - States directly construct each other for navigation, creating implicit dependencies and requiring manual preservation of parent state data for back-navigation

## Goals

1. **Clearer package hierarchy** - Structure that reflects architectural layers
2. **Separated concerns** - Application core distinct from UI presentation
3. **Decoupled states** - States don't know about each other's existence
4. **Automatic back-navigation** - No manual field copying to restore previous state
5. **Organized state files** - States in subfolders, shared components extracted

## Non-Goals

- Changing the Elm Architecture (Model-Update-View) pattern
- Modifying domain logic in `git`, `diff`, `graph`, `highlight`
- Changing the visual appearance or UX

## Key Requirements

### 1. Package Hierarchy

Restructure `internal/` to reflect layers:

```
internal/
├── git/              # External adapter (stays at root - special)
├── domain/           # Domain processing
│   ├── diff/
│   ├── graph/
│   └── highlight/
├── app/              # Application core
│   ├── model.go      # Model, Init/Update/View
│   ├── navigation.go # Navigation messages, Screen enum
│   └── messages/     # Message definitions
└── ui/               # Presentation
    ├── states/       # State implementations (subfolders)
    ├── components/   # Shared view components
    ├── styles/       # Style definitions
    └── format/       # Display formatting
```

### 2. Navigation Stack

Replace direct state construction with a stack-based navigation system:

- **Model manages stack**: `stack []State` preserves previous states
- **Navigation via messages**: States return `PushScreenMsg` or `PopScreenMsg`
- **Factory in Model**: Model creates states, states never import each other
- **Automatic restoration**: Pop returns exact previous state (no field copying)

### 3. State Handling

- **LoadingState**: Initial state, never on stack. Transitions to LogState.
- **ErrorState**: Pushed onto stack. User can go back to previous state.
- **LogState, FilesState, DiffState**: Normal stack behavior (push forward, pop back).

### 4. Shared Components Extraction

Move from `ui/states/` to `ui/components/`:
- `viewbuilder.go` - View output builder
- `commit_info.go` - Commit metadata rendering
- `file_section.go` - File list rendering

Keep with LogState (or move to components):
- `log_line_format.go` - Only used by LogState

### 5. State Subfolders

Organize states into subfolders:
```
ui/states/
├── interfaces.go     # State, Context interfaces
├── loading/
│   ├── state.go
│   ├── view.go
│   └── update.go
├── log/
│   └── ...
├── files/
│   └── ...
├── diff/
│   └── ...
└── error/
    └── ...
```

## Dependency Rules

After restructuring:
- `app/` imports all state packages (for factory)
- State packages import `app/` (for navigation messages)
- State packages import `domain/` packages as needed
- State packages import `ui/components/`
- **State packages never import each other**

```
        app/
       ↙ ↓  ↘
    log files diff ...
      ↘  ↓  ↙
    ui/components
         ↓
      domain/*
```

## Open Questions for Design Phase

1. **Message definitions location** - Should `messages/` stay under `app/` or move elsewhere?
2. **Context interface location** - Should it live in `app/` or `ui/states/`?
3. **Screen data types** - Where should `FilesScreenData`, `DiffScreenData` etc. be defined?
4. **Migration strategy** - Big bang or incremental refactoring?

## References

- [Research: Current Architecture](research/current-architecture.md)
- [Research: Package Boundaries](research/package-boundaries.md)
