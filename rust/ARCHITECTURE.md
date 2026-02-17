# Rust Implementation Architecture

Port of the Go implementation (`../go/`) with Rust-idiomatic patterns. The Go codebase is the source of truth for behavior.

## Architecture

### Layered Structure

```
src/
├── core/           # Shared types (Commit, FileChange, FileStatus)
├── app/            # App state + view states (LogView, FilesView)
├── git/            # Git operations (commands, parsing)
└── ui/             # UI layer (views, theme)
```

### Import Direction

```
app → ui → git
         ↑
      core/ (shared)
```

**Rules:**
- `git/` only depends on `core/`
- `ui/` depends on `core/`, `app/` (for view types), and `git/`
- `app/` depends on `core/`, `git/`, and `input/`

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

1. **Core types** - Commit, FileChange, FileStatus
2. **Git layer** - parsing → commands → operations
3. **App layer** - View enum, LogView, FilesView states
4. **UI layer** - theme → log view → files view
5. **Polish** - Error handling, edge cases

## Go Reference Files

| Rust Module | Go Reference |
|-------------|--------------|
| `core/types` | `go/internal/core/git_types.go`, `diff_source.go` |
| `git/log` | `go/internal/git/parsing/parsing.go` (log parsing) |
| `git/file_changes` | `go/internal/git/parsing/parsing.go` (file changes) |
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
