# Design: Type-Safe View Rendering with ViewBuilder

## Executive Summary

We'll introduce a `ViewBuilder` type that encapsulates the logic for building terminal views without trailing newlines. State `View()` methods will return `*ViewBuilder` instead of `string`, making it architecturally impossible to reintroduce the trailing newline bug. This change affects the `State` interface and all five state implementations, but the refactoring is straightforward and mechanical.

## Context & Problem Statement

When fixing the "first line cutoff" bug, we discovered that all view rendering methods had to be updated to avoid adding a trailing newline (which causes scrolling in Bubbletea's altscreen mode). The fix involved adding conditional logic in four places:

```go
if i < viewportEnd-1 {
    b.WriteString("\n")
}
```

This pattern is duplicated and error-prone. Future view implementations could easily reintroduce the bug by forgetting this conditional.

**Scope:** This design covers refactoring view rendering to be type-safe. It does not change any rendering logic or visual output - only how views are constructed internally.

## Current State

State implementations use `strings.Builder` directly and return `string`:

```go
type State interface {
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
    View(ctx Context) string  // Returns string
}

func (s *DiffState) View(ctx Context) string {
    var b strings.Builder
    // ... build content
    for i := start; i < end; i++ {
        b.WriteString(line)
        if i < end-1 {  // Error-prone conditional
            b.WriteString("\n")
        }
    }
    return b.String()
}
```

All five states follow this pattern: `LoadingState`, `ErrorState`, `LogState`, `FilesState`, `DiffState`.

## Solution

### ViewBuilder Type

Create a new type that handles newline logic automatically:

```go
// File: internal/ui/states/viewbuilder.go
package states

import "strings"

// ViewBuilder builds view output with automatic newline handling.
// It ensures views never end with a trailing newline by storing
// lines separately and joining them only when String() is called.
type ViewBuilder struct {
    lines []string
}

// NewViewBuilder creates a new ViewBuilder.
func NewViewBuilder() *ViewBuilder {
    return &ViewBuilder{}
}

// AddLine adds a line to the view. Newlines are automatically
// added between lines (via strings.Join) but not after the last line.
func (vb *ViewBuilder) AddLine(line string) {
    vb.lines = append(vb.lines, line)
}

// String returns the final view output without trailing newline.
// Lines are joined with "\n" separator.
func (vb *ViewBuilder) String() string {
    return strings.Join(vb.lines, "\n")
}
```

### Updated State Interface

Change the return type from `string` to `*ViewBuilder`:

```go
type State interface {
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
    View(ctx Context) *ViewBuilder  // Changed from string
}
```

### State Implementation Pattern

Each state's View method changes from:

```go
func (s *DiffState) View(ctx Context) string {
    var b strings.Builder
    // ... header
    for i := start; i < end; i++ {
        b.WriteString(line)
        if i < end-1 {
            b.WriteString("\n")
        }
    }
    return b.String()
}
```

To:

```go
func (s *DiffState) View(ctx Context) *ViewBuilder {
    vb := NewViewBuilder()
    // ... header
    for i := start; i < end; i++ {
        vb.AddLine(line)  // Newline handling is automatic
    }
    return vb
}
```

### Bubbletea Integration

The top-level `Model.View()` converts to string:

```go
// File: internal/ui/app.go
func (m Model) View() string {
    return m.currentState.View(&m).String()
}
```

### Data Flow Diagram

```
┌─────────────────────────────────────────────────────┐
│ Bubbletea (requires string)                         │
│                                                     │
│  Model.View() string                                │
│    ↓                                                │
│    └─→ currentState.View(&m).String()               │
└─────────────────────────┬───────────────────────────┘
                          │
                          ↓
┌─────────────────────────────────────────────────────┐
│ State Layer (returns *ViewBuilder)                  │
│                                                     │
│  LogState.View(ctx) *ViewBuilder                    │
│  FilesState.View(ctx) *ViewBuilder                  │
│  DiffState.View(ctx) *ViewBuilder                   │
│  ErrorState.View(ctx) *ViewBuilder                  │
│  LoadingState.View(ctx) *ViewBuilder                │
│                                                     │
│  Each creates ViewBuilder and uses AddLine()        │
└─────────────────────────┬───────────────────────────┘
                          │
                          ↓
┌─────────────────────────────────────────────────────┐
│ ViewBuilder (encapsulates newline logic)            │
│                                                     │
│  • AddLine() - automatically handles newlines       │
│  • String() - returns final output                  │
│                                                     │
│  Guarantees: no trailing newline ever               │
└─────────────────────────────────────────────────────┘
```

### Changes Required

1. **Create ViewBuilder** (`internal/ui/states/viewbuilder.go`)
   - New file with `ViewBuilder` type and methods

2. **Update State Interface** (`internal/ui/states/context.go`)
   - Change `View(ctx Context) string` to `View(ctx Context) *ViewBuilder`

3. **Refactor 5 State Implementations**
   - `loading_view.go`, `error_view.go`, `log_view.go`, `files_view.go`, `diff_view.go`
   - Change return type and use `ViewBuilder.AddLine()`
   - Remove all conditional newline logic

4. **Update Model.View()** (`internal/ui/app.go`)
   - Call `.String()` on result: `return m.currentState.View(&m).String()`

5. **Update Tests**
   - All view tests will need minor updates to call `.String()`

> **Decision:** We chose to return `*ViewBuilder` (pointer) rather than `ViewBuilder` (value) because:
> - The builder accumulates state (growing slice of lines)
> - Pointer semantics make it clear this is a mutable builder
> - Avoids copying the entire line slice on return
> - Consistent with builder patterns in Go

> **Decision:** We're placing this in `internal/ui/states/` rather than a separate package because:
> - It's only used by state view methods
> - Keeping it in the same package avoids import cycles
> - It's tightly coupled to the State interface

### Benefits

1. **Architectural prevention** - Impossible to forget newline handling
2. **Simpler code** - No conditionals in loops, just `AddLine()`
3. **Self-documenting** - `AddLine()` clearly expresses intent
4. **Type safety** - Compiler enforces correct usage
5. **Future-proof** - New states automatically get correct behavior
6. **Clean implementation** - Using `[]string` with `strings.Join` is idiomatic and explicit about building lines

### Risks & Tradeoffs

**Risk:** Large diff touching all view code
- *Mitigation:* The changes are mechanical and well-tested. Golden files verify no visual changes.

**Tradeoff:** Added abstraction layer
- *Accept:* The benefit of preventing bugs outweighs the small complexity increase. The API is simple (just `AddLine()`).

**Tradeoff:** Slice allocations vs. strings.Builder direct writes
- *Accept:* Terminal views are bounded by screen height (~50-200 lines max). The memory difference is negligible, and the cleaner code is worth it.

## Open Questions

None - the design is straightforward and all decisions have been made.
