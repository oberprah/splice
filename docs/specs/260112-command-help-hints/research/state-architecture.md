# Research: Splice State Architecture

**Date:** 2026-01-12

## Core State Interface

The `State` interface is the foundation of Splice's architecture (`internal/core/state.go`):

```go
type State interface {
    View(ctx Context) ViewRenderer
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
```

**Key Points:**
- States are immutable values - methods return `(State, tea.Cmd)` pairs
- States receive a `core.Context` interface for dependency injection
- Returns `tea.Cmd` functions for async operations and navigation messages

## Update() Pattern

All states follow a consistent message-handling pattern:

```go
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return s, func() tea.Msg { return core.PopScreenMsg{} }
        case "j", "down":
            s.Cursor++
            return s, nil
        case "enter":
            return s, loadSomething(ctx)
        }
    }
    return s, nil
}
```

**Key Patterns:**
- Type-safe message handling with `switch msg := msg.(type)`
- Direct string matching for key bindings (no keybinding library)
- Navigation via typed messages (`PushLogScreenMsg`, `PopScreenMsg`, etc.)
- Inline state mutations for simple updates
- Return `tea.Cmd` functions for async operations

## View() Pattern

All states follow a consistent rendering pattern:

```go
func (s State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    lines := s.renderSection()
    for _, line := range lines {
        vb.AddLine(line)
    }

    if ctx.Width() >= threshold {
        leftVb := s.buildLeftColumn(ctx)
        rightVb := s.buildRightColumn(ctx)
        vb.AddSplitView(leftVb, rightVb)
    }

    return vb
}
```

**Key Patterns:**
- Return `core.ViewRenderer` (implemented by `ViewBuilder`)
- Responsive layout based on `ctx.Width()` and `ctx.Height()`
- Split views using `vb.AddSplitView(left, right)`
- Line-by-line building with `vb.AddLine()`

## State Composition

**Sum Types for State Variants:**
```go
type PreviewState interface {
    isPreviewState()
}

type PreviewNone struct{}
type PreviewLoading struct{ ForHash string }
type PreviewLoaded struct{ ForHash string; Files []core.FileChange }
type PreviewError struct{ ForHash string; Err error }
```

Benefits: Makes illegal states unrepresentable, avoids boolean explosion

**Component Reuse:**
- States call shared component functions from `internal/ui/components`
- Components accept primitives and return rendered strings or ViewBuilders
- Example: `components.CommitInfo()`, `components.FileSection()`

## Key Binding Patterns

Direct string matching:
- `j/k` or `down/up` - Navigate
- `g/G` - Jump to start/end
- `enter` - Select/open
- `q` or `ctrl+c` - Back/quit
- `esc` - Clear mode

**No global keybinding registry** - each state independently handles its own keys

## Navigation Architecture

**Navigation Messages:**
- `PushLogScreenMsg`, `PushFilesScreenMsg`, `PushDiffScreenMsg`
- `PopScreenMsg`

**Flow:**
1. State returns command: `func() tea.Msg { return core.Push*ScreenMsg{...} }`
2. Bubbletea executes command
3. `app.Model.Update()` handles navigation message
4. Model pushes/pops state on navigation stack

## State File Organization

Each state package:
```
internal/ui/states/<name>/
├── state.go       # Struct definition, constructor, helpers
├── update.go      # Update() method
└── view.go        # View() method
```

## Context Interface

States receive `core.Context` for dependency injection:
```go
type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}
```

Benefits: Decouples states from app model, enables easy testing
