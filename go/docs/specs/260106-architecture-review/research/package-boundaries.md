# Package Boundaries Analysis

## Dependency Graph

```
main.go
    └── internal/ui
            ├── internal/ui/states
            │       ├── internal/git
            │       ├── internal/diff ──► internal/highlight
            │       ├── internal/graph
            │       ├── internal/ui/styles
            │       ├── internal/ui/format
            │       └── internal/ui/messages
            ├── internal/ui/styles
            ├── internal/ui/format
            └── internal/ui/messages
```

**No circular dependencies** - graph is acyclic.

## Package Cohesion Summary

| Package | Lines | Files | Responsibility | Cohesion |
|---------|-------|-------|----------------|----------|
| `internal/git` | ~480 | 2 | Git CLI execution + output parsing | Clean |
| `internal/diff` | ~620 | 6 | Diff parsing, alignment, pairing | Clean |
| `internal/highlight` | ~130 | 2 | Syntax tokenization + styling | Clean |
| `internal/graph` | ~500 | 8 | Commit graph layout computation | Clean |
| `internal/ui/states` | ~2500 | 20+ | State machine (5 states) | Large but organized |
| `internal/ui/format` | ~70 | 2 | Display formatting (hash, time) | Clean |
| `internal/ui/styles` | ~120 | 1 | Lipgloss style definitions | Clean |
| `internal/ui/messages` | ~42 | 1 | Message type definitions | Clean |

## Boundary Clarity Assessment

### Domain Packages (Clean Separation)

- **`internal/git`**: Pure git concerns - executes CLI, parses output
- **`internal/diff`**: Pure diff concerns - parsing, alignment, pairing algorithm
- **`internal/graph`**: Pure layout concerns - lane computation, symbol generation
- **`internal/highlight`**: Pure highlighting - Chroma lexer, Lipgloss styling

### UI Package Structure

```
internal/ui/
├── app.go              # Model, Init/Update/View, DI options
├── model.go            # Model implements Context
├── states/
│   ├── state.go        # State & Context interfaces
│   ├── viewbuilder.go  # View output utility
│   ├── commit_info.go  # Shared: commit metadata rendering
│   ├── file_section.go # Shared: file list rendering
│   ├── log_line_format.go  # Log-specific: progressive truncation
│   ├── loading_*.go    # LoadingState (3 files)
│   ├── log_*.go        # LogState (4 files)
│   ├── files_*.go      # FilesState (3 files)
│   ├── diff_*.go       # DiffState (4 files)
│   └── error_*.go      # ErrorState (3 files)
├── messages/
│   └── messages.go     # Message definitions
├── styles/
│   └── styles.go       # Style definitions
└── format/
    ├── hash.go         # ToShortHash()
    └── time.go         # ToRelativeTimeFrom()
```

## Observations

### Strengths

1. **Clear domain boundaries** - git/diff/graph/highlight are independent
2. **No god files** - largest files (~400 lines) are justified
3. **Testability** - DI via functional options works well
4. **File naming convention** - `{state}_state.go`, `{state}_view.go`, `{state}_update.go` is consistent

### Potential Areas for Discussion

1. **Shared view utilities in states package**
   - `commit_info.go`, `file_section.go`, `viewbuilder.go` are view components
   - Could extract to `internal/ui/components/` if more emerge

2. **State preservation pattern**
   - `DiffState` stores 7 fields from parent states for back-navigation
   - `FilesState` stores 3 fields from LogState
   - Common in Elm-style state machines, but adds coupling

3. **`log_line_format.go` placement**
   - 397 lines of progressive truncation algorithm
   - Only used by LogState
   - Could argue it belongs in `internal/ui/format/` or stays where it is

4. **Message definitions location**
   - Currently in `internal/ui/messages/messages.go`
   - Some message types embed domain types (`git.GitCommit`, `diff.AlignedFileDiff`)
   - Creates coupling between messages and domain packages
