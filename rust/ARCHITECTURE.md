# Rust Implementation Architecture

Port of the Go implementation (`../go/`) with Rust-idiomatic patterns. The Go codebase is the source of truth for behavior.

## Architecture

### Layered Structure

```
src/
├── core/           # Shared types (Commit, FileChange, DiffSource enums)
├── domain/         # Pure business logic (diff, graph, filetree) - NO git, NO UI
├── git/            # Git operations (commands, operations, parsing)
└── ui/             # UI layer (views, components, theme)
```

### Import Direction

```
app → ui → domain → git
        ↑
     core/ (shared)
```

**Rules:**
- `domain/` has NO dependencies on `git/` or `ui/`
- `git/` only depends on `core/` and `domain/`
- `ui/` depends on everything

## Go → Rust Pattern Mapping

| Go Pattern | Rust Equivalent |
|------------|-----------------|
| Sealed interfaces (`DiffSource`) | Native enums |
| `core.State` interface | `View` enum with state variants |
| Functional options | Builder pattern |
| `core.Context` interface | Pass dependencies directly or via struct |
| Navigation stack | `Vec<View>` for history |

**Key insight:** Rust enums are more powerful than Go's sealed interfaces. Use them for sum types.

## Testing Strategy

### Test Categories

| Layer | Test Type | Git Strategy |
|-------|-----------|--------------|
| `domain/` | Unit tests | No git needed (pure functions) |
| `git/` | Integration tests | **Real git repos** with `tempfile` |
| `ui/` | Snapshot tests | Mocked git via trait |
| Full app | E2E tests | Real git repos |

### Why Real Git Repos for Integration Tests?

- Tests actual git behavior (edge cases, version differences)
- Catches parsing issues with real output
- `tempfile` creates isolated test repos quickly
- Go implementation uses mocking; we can do better with Rust's test isolation

### Snapshot Testing for UI

Use `insta` for deterministic UI tests. Mock git operations to ensure consistent output.

## Implementation Order

1. **Core types** - Commit, FileChange, DiffSource enums
2. **Domain layer** - Port from Go (pure functions, easy to test)
3. **Git layer** - parsing → commands → operations
4. **UI infrastructure** - theme, components
5. **UI states** - loading → log → files → diff
6. **Polish** - Error handling, edge cases

## Go Reference Files

| Rust Module | Go Reference |
|-------------|--------------|
| `core/types` | `go/internal/core/git_types.go`, `diff_source.go` |
| `domain/diff` | `go/internal/domain/diff/` |
| `domain/graph` | `go/internal/domain/graph/` |
| `domain/filetree` | `go/internal/domain/filetree/` |
| `git/commands` | `go/internal/git/commands/` |
| `git/operations` | `go/internal/git/operations/` |
| `git/parsing` | `go/internal/git/parsing/parsing.go` |
| `ui/log` | `go/internal/ui/states/log/` |
| `ui/files` | `go/internal/ui/states/files/` |
| `ui/diff` | `go/internal/ui/states/diff/` |
| `ui/theme` | `go/internal/ui/styles/styles.go` |
| `ui/components` | `go/internal/ui/components/` |

## Coding Principles

- **Deep functions**: Fewer, more capable functions over many shallow ones
- **Pure functions**: Isolate side effects at boundaries
- **Minimal comments**: Code should be self-explanatory
- **Make illegal states unrepresentable**: Use enums to prevent invalid states
- **Enums over multiple booleans**: Related booleans → single enum
- **TDD**: Write tests first
