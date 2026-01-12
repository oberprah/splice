# Solution Approaches: Generic Help Overlay

**Date:** 2026-01-12

## Overview

This document analyzes four different architectural approaches for implementing a reusable help overlay system in Splice. The goal is to avoid code duplication while maintaining consistency with Splice's architectural patterns.

## Requirements Recap

- **Generic**: Reusable across LogState, FilesState, and DiffState
- **Context-sensitive**: Each screen shows different commands
- **Modal overlay**: Rendered as centered box over current view
- **Toggle behavior**: `?` key shows/hides, also dismissed by `q` or `esc`
- **No code duplication**: Single implementation, not per-state solutions

## Approach 1: Pure Component Function

### Description

Create a pure rendering function in `internal/ui/components/help_overlay.go`:

```go
type Command struct {
    Keys        []string  // e.g., ["j", "down"]
    Description string
}

func HelpOverlay(commands []Command, width, height int) string {
    // Returns a centered modal box with commands listed
}
```

Each state:
- Adds a `showHelp bool` field
- Handles `?` key to toggle the field
- Defines static command list (e.g., `logCommands = []Command{...}`)
- In `View()`, conditionally renders overlay over content

```go
// In log/state.go
type State struct {
    Commits       []core.GitCommit
    Cursor        core.CursorState
    ViewportStart int
    Preview       PreviewState
    GraphLayout   *graph.Layout
    showHelp      bool  // New field
}

// In log/update.go
case "?":
    s.showHelp = !s.showHelp
    return s, nil

// In log/view.go
func (s State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // ... normal rendering ...

    if s.showHelp {
        overlay := components.HelpOverlay(logCommands, ctx.Width(), ctx.Height())
        vb.AddOverlay(overlay)  // New ViewBuilder method
    }

    return vb
}
```

### Analysis

**How well does it meet the "generic" requirement?**
- ✅ Rendering logic is fully generic and reusable
- ⚠️ Each state must manually integrate (field + key handler + view call)
- ⚠️ Command lists are defined per-state (acceptable - they ARE different)

**Code duplication across states?**
- ❌ Each state needs: `showHelp bool` field (3 duplications)
- ❌ Each state needs: `case "?"` handler (3 duplications)
- ❌ Each state needs: conditional overlay rendering (3 duplications)
- ✅ Rendering logic is not duplicated
- ✅ Command definitions are not duplication (they differ per state)

**Complexity vs. clarity tradeoff**
- ✅ Very clear - no abstraction layers
- ✅ Easy to understand what each state does
- ✅ Direct control over behavior
- ❌ Boilerplate across states

**Consistency with existing Splice patterns**
- ✅ Pure component function matches `CommitInfo`, `FileSection` pattern
- ✅ States manage their own fields and behavior
- ✅ No new architectural concepts
- ✅ ViewBuilder pattern (would need `AddOverlay` method)

**Ease of maintenance and future extensions**
- ✅ Updating overlay styling: change component function (1 place)
- ❌ Adding new dismiss key: update all 3 states
- ❌ Changing help toggle behavior: update all 3 states
- ⚠️ Adding help to new state: copy pattern (clear but repetitive)

### Pros & Cons Summary

**Pros:**
- Simple, straightforward implementation
- Matches existing component patterns exactly
- Pure function is easily testable
- States have full control over help behavior
- No new architectural concepts to learn

**Cons:**
- Key handling duplicated 3 times
- `showHelp` field duplicated 3 times
- View integration duplicated 3 times
- Changes to help behavior require 3 updates
- New states must remember to implement pattern

---

## Approach 2: Help State Wrapper

### Description

Create a `HelpWrapper` type that wraps any state and intercepts help-related keys:

```go
// In internal/ui/components/help_wrapper.go
type HelpWrapper struct {
    wrapped  core.State
    commands []Command
    visible  bool
}

func NewHelpWrapper(wrapped core.State, commands []Command) *HelpWrapper {
    return &HelpWrapper{
        wrapped:  wrapped,
        commands: commands,
        visible:  false,
    }
}

func (h *HelpWrapper) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    if keyMsg, ok := msg.(tea.KeyMsg); ok {
        switch keyMsg.String() {
        case "?":
            h.visible = !h.visible
            return h, nil
        case "q", "esc":
            if h.visible {
                h.visible = false
                return h, nil
            }
        }
    }

    // If help is visible, don't pass other keys to wrapped state
    if h.visible {
        return h, nil
    }

    // Delegate to wrapped state
    newWrapped, cmd := h.wrapped.Update(msg, ctx)
    h.wrapped = newWrapped
    return h, cmd
}

func (h *HelpWrapper) View(ctx core.Context) core.ViewRenderer {
    wrappedView := h.wrapped.View(ctx)

    if !h.visible {
        return wrappedView
    }

    // Render overlay on top
    overlay := HelpOverlay(h.commands, ctx.Width(), ctx.Height())
    return RenderWithOverlay(wrappedView, overlay)
}
```

Usage in app.Model:

```go
case core.PushLogScreenMsg:
    wrapped := log.New(msg.Commits, msg.GraphLayout)
    withHelp := components.NewHelpWrapper(wrapped, logCommands)
    m.pushState(withHelp)
```

### Analysis

**How well does it meet the "generic" requirement?**
- ✅ Completely generic - works with any state
- ✅ Single implementation of help logic
- ✅ Command lists passed during wrapper creation

**Code duplication across states?**
- ✅ No duplication in states themselves
- ✅ Single help toggle implementation
- ⚠️ Each state creation needs wrapper call
- ⚠️ Command lists defined at app level

**Complexity vs. clarity tradeoff**
- ❌ Wrapper pattern not used elsewhere in Splice
- ❌ Adds indirection - hard to trace message flow
- ❌ State wrapping feels unnatural for the architecture
- ❌ Wrapper lifecycle management unclear
- ❌ Does wrapped state get replaced or does wrapper get replaced?

**Consistency with existing Splice patterns**
- ❌ States don't wrap other states in Splice
- ❌ No existing decorator/wrapper patterns
- ❌ Violates the simple State interface expectations
- ❌ App.Model would need to know about wrappers
- ❌ Navigation messages would reference wrappers, not actual states

**Ease of maintenance and future extensions**
- ✅ Updating help behavior: change wrapper (1 place)
- ✅ Adding help to new state: wrap it during creation
- ❌ Debugging becomes harder (extra layer)
- ❌ State type assertions break (is it LogState or HelpWrapper?)
- ❌ Testing becomes more complex

### Pros & Cons Summary

**Pros:**
- Zero duplication across states
- Completely generic wrapper
- Single source of truth for help logic
- Easy to add help to new states

**Cons:**
- Foreign pattern to Splice architecture
- Adds significant conceptual complexity
- State wrapping is unusual and confusing
- Breaks state type identity
- Navigation stack becomes unclear
- No precedent for this pattern in codebase
- Hard to debug wrapped message flow

---

## Approach 3: State Composition with Embedded Helper

### Description

Create a `HelpOverlay` struct that can be embedded in states, with helper functions for common operations:

```go
// In internal/ui/components/help_overlay.go
type HelpOverlay struct {
    Visible  bool
    Commands []Command
}

func NewHelpOverlay(commands []Command) HelpOverlay {
    return HelpOverlay{
        Visible:  false,
        Commands: commands,
    }
}

// Helper function for Update()
func (h *HelpOverlay) HandleHelpKey(key string) (handled bool, shouldReturn bool) {
    switch key {
    case "?":
        h.Visible = !h.Visible
        return true, true  // handled, should return
    case "q", "esc":
        if h.Visible {
            h.Visible = false
            return true, true
        }
    }
    return false, false
}

// Helper function for View()
func (h *HelpOverlay) RenderIfVisible(vb *components.ViewBuilder, width, height int) {
    if h.Visible {
        overlay := HelpOverlay(h.Commands, width, height)
        vb.AddOverlay(overlay)
    }
}
```

Usage in states:

```go
// In log/state.go
type State struct {
    Commits       []core.GitCommit
    Cursor        core.CursorState
    ViewportStart int
    Preview       PreviewState
    GraphLayout   *graph.Layout
    HelpOverlay   components.HelpOverlay  // Embedded
}

func New(commits []core.GitCommit, layout *graph.Layout) *State {
    return &State{
        // ... other fields ...
        HelpOverlay: components.NewHelpOverlay(logCommands),
    }
}

// In log/update.go
case tea.KeyMsg:
    if handled, shouldReturn := s.HelpOverlay.HandleHelpKey(msg.String()); handled {
        if shouldReturn {
            return s, nil
        }
    }

    // Don't process other keys when help is visible
    if s.HelpOverlay.Visible {
        return s, nil
    }

    switch msg.String() {
    // ... normal key handling ...
    }

// In log/view.go
func (s State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // ... normal rendering ...

    s.HelpOverlay.RenderIfVisible(vb, ctx.Width(), ctx.Height())

    return vb
}
```

### Analysis

**How well does it meet the "generic" requirement?**
- ✅ Shared logic via helper functions
- ✅ Help toggle logic centralized
- ⚠️ States still need to embed and call helpers
- ✅ Command lists defined once per state

**Code duplication across states?**
- ⚠️ Embedding declaration duplicated (3 times)
- ⚠️ Helper function calls duplicated (3 times in Update, 3 in View)
- ✅ Help toggle logic not duplicated
- ✅ Rendering logic not duplicated
- Better than Approach 1, but still some duplication

**Complexity vs. clarity tradeoff**
- ⚠️ Helper functions add a layer of abstraction
- ✅ Helper functions are straightforward
- ⚠️ Two-step integration (embed + call helpers)
- ✅ Explicit control flow (clear where helpers are called)

**Consistency with existing Splice patterns**
- ⚠️ Embedding is a Go idiom, but not heavily used in Splice
- ⚠️ Helper functions pattern exists but not common
- ❌ States don't typically embed shared behavior
- ⚠️ Feels like partial abstraction (why not go full component?)

**Ease of maintenance and future extensions**
- ✅ Updating help behavior: change helper functions
- ✅ Updating overlay styling: change component
- ⚠️ Adding new dismiss key: update `HandleHelpKey` (1 place)
- ⚠️ New states must embed and call helpers (clear pattern)

### Pros & Cons Summary

**Pros:**
- Reduced duplication compared to Approach 1
- Helper functions are testable
- States maintain control
- Clear where help logic is invoked

**Cons:**
- Still requires embedding in each state
- Still requires calling helpers in Update/View
- Feels like incomplete abstraction
- Embedding pattern uncommon in Splice
- Not fully "generic" - manual integration needed
- Helper functions add indirection

---

## Approach 4: Separate Help Screen State

### Description

Create a `HelpState` that is pushed onto the navigation stack like any other screen:

```go
// In internal/ui/states/help/state.go
type StateType int

const (
    StateTypeLog StateType = iota
    StateTypeFiles
    StateTypeDiff
)

type State struct {
    ParentType StateType
}

func New(parentType StateType) *State {
    return &State{ParentType: parentType}
}

// In help/update.go
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "?", "q", "esc":
            return s, func() tea.Msg { return core.PopScreenMsg{} }
        }
    }
    return s, nil
}

// In help/view.go
func (s State) View(ctx core.Context) core.ViewRenderer {
    commands := s.getCommandsForType()
    return renderHelpScreen(commands, ctx.Width(), ctx.Height())
}

func (s State) getCommandsForType() []Command {
    switch s.ParentType {
    case StateTypeLog:
        return logCommands
    case StateTypeFiles:
        return filesCommands
    case StateTypeDiff:
        return diffCommands
    }
}
```

Navigation message:

```go
// In internal/core/navigation.go
type PushHelpScreenMsg struct {
    ParentType help.StateType
}
```

Usage in states:

```go
// In log/update.go
case "?":
    return s, func() tea.Msg {
        return core.PushHelpScreenMsg{ParentType: help.StateTypeLog}
    }
```

### Analysis

**How well does it meet the "generic" requirement?**
- ✅ Single HelpState implementation
- ✅ Fully generic help screen
- ⚠️ Requires parent type enum (not truly generic)
- ❌ States must know about help state and navigation

**Code duplication across states?**
- ✅ No duplication of help logic
- ✅ No duplication of rendering
- ⚠️ Each state needs `case "?"` handler (3 times)
- ⚠️ Each state needs to create PushHelpScreenMsg (3 times)

**Complexity vs. clarity tradeoff**
- ❌ Adds new state type (more complexity)
- ❌ Adds navigation message type
- ❌ Help state needs to know about all parent types
- ❌ Parent type enum maintenance burden
- ✅ Uses familiar navigation pattern

**Consistency with existing Splice patterns**
- ⚠️ Uses standard push/pop navigation (familiar)
- ❌ But help is a modal overlay, not a screen transition
- ❌ Violates the UX requirement (overlay, not new screen)
- ❌ Adds to navigation stack when it shouldn't be a "screen"
- ❌ Dimmed background is harder (current screen not in view)

**Ease of maintenance and future extensions**
- ✅ Help rendering in one place
- ❌ Adding new state type: update enum, update switch
- ⚠️ Navigation overhead for simple modal
- ❌ Stack depth increases unnecessarily
- ❌ "Back" navigation feels wrong for help

### Pros & Cons Summary

**Pros:**
- Uses familiar push/pop pattern
- Help logic completely centralized
- Full screen control for complex help (if needed)
- Standard state implementation

**Cons:**
- Help is an overlay, not a navigation destination
- Violates UX requirement (modal vs screen)
- Adds to navigation stack unnecessarily
- Requires parent type enum and maintenance
- New navigation message type
- More complex than needed for simple overlay
- Dimmed background harder to implement
- "Back" semantics feel wrong

---

## Comparison Matrix

| Criterion | Approach 1: Pure Component | Approach 2: State Wrapper | Approach 3: Embedded Helper | Approach 4: Help Screen |
|-----------|---------------------------|---------------------------|----------------------------|------------------------|
| **Zero duplication** | ❌ Some duplication | ✅ Minimal | ⚠️ Reduced | ⚠️ Reduced |
| **Truly generic** | ⚠️ Manual integration | ✅ Yes | ⚠️ Manual integration | ⚠️ Requires enum |
| **Architectural consistency** | ✅ Perfect match | ❌ Foreign pattern | ⚠️ Uncommon | ⚠️ Wrong abstraction |
| **Conceptual simplicity** | ✅ Very simple | ❌ Complex | ⚠️ Moderate | ❌ Overhead |
| **UX requirement match** | ✅ Modal overlay | ✅ Modal overlay | ✅ Modal overlay | ❌ Screen navigation |
| **Ease of testing** | ✅ Pure function | ⚠️ Wrapper complexity | ✅ Testable helpers | ✅ Standard state |
| **Maintainability** | ⚠️ Multiple update points | ✅ Single source | ✅ Single source | ⚠️ Enum maintenance |
| **Learning curve** | ✅ Zero (familiar) | ❌ New pattern | ⚠️ Helper pattern | ✅ Familiar navigation |
| **Future flexibility** | ✅ Easy to extend | ❌ Locked in wrapper | ✅ Can extend helpers | ⚠️ Limited by enum |

## Recommendation

### Primary Recommendation: **Approach 1 (Pure Component Function)**

Despite the code duplication, Approach 1 is the best choice for Splice because:

1. **Architectural Consistency**: Perfectly matches existing component patterns (`CommitInfo`, `FileSection`). No new concepts to learn.

2. **Simplicity Over Cleverness**: The duplication is straightforward and visible. Each state explicitly declares "I have help" rather than having help magically injected.

3. **Explicit is Better**: Each state handles its own `?` key, making the behavior obvious when reading code. No indirection through wrappers or helpers.

4. **Low Maintenance Burden**: The "duplicated" code is ~5 lines per state. The complexity savings from avoiding abstraction outweigh the copy-paste cost.

5. **YAGNI Principle**: The duplication might enable future flexibility (e.g., LogState could add a `shift+?` for advanced help, without affecting other states).

6. **Testing Simplicity**: Pure component function is trivial to test. State integration is obvious and doesn't require wrapper/helper tests.

### Implementation Notes for Approach 1

**Minimize duplication where practical:**
- Create `HelpOverlay(commands, width, height)` component function (fully reusable)
- Add `ViewBuilder.AddOverlay(overlay string)` method (one implementation)
- Consider a comment/doc pointing to the pattern for new states

**Accept duplication where it's simple:**
- `showHelp bool` field (1 line per state)
- `case "?"` handler (2 lines per state)
- `if s.showHelp` check in View() (3 lines per state)

**Total per-state cost:** ~6 lines of simple, obvious code.

### Why Not The Others?

- **Approach 2 (State Wrapper)**: Too clever. Introduces architectural complexity foreign to Splice. State wrapping breaks type identity and makes debugging harder.

- **Approach 3 (Embedded Helper)**: Halfway abstraction. Still requires manual integration but adds helper indirection. Doesn't eliminate enough duplication to justify the complexity.

- **Approach 4 (Help Screen)**: Wrong abstraction. Help is a modal overlay, not a navigation destination. Violates UX requirement and adds unnecessary stack depth.

### Alternative: Hybrid Approach 1 + 3

If the duplication becomes truly burdensome (e.g., adding 5+ more states with help), consider:
- Keep Approach 1's structure (pure component, explicit integration)
- Add small helper for common `case "?"` logic if the pattern evolves

But for 3 states with simple help, pure Approach 1 is the pragmatic choice.

---

## Conclusion

**Winner: Approach 1 (Pure Component Function)**

The best solution embraces Splice's existing patterns and values clarity over cleverness. The small amount of duplication is acceptable given the simplicity and maintainability benefits.

The pure component function approach:
- Matches existing architecture perfectly
- Requires no new concepts
- Is obvious when reading state code
- Keeps each state self-contained
- Allows future per-state customization

For a 3-state application with a simple help overlay, avoiding premature abstraction is the right choice.
