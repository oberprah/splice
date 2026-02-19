# ADR-003: Git Package Restructuring

* **Status**: Accepted
* **Date**: 2026-01-19
* **Problem**: Git package is a single 1117-line file that's hard to navigate and contains significant code duplication
* **Decision**: Restructure into layered architecture with commands/operations separation and eliminate duplication

## Context

The `internal/git` package currently consists of a single 1117-line `git.go` file with 20+ functions. While it has excellent test coverage and proper abstraction boundaries (functions are injectable via app.Model), the file organization creates several problems:

### Current Structure Issues

1. **Hard to navigate**: Single 1117-line file requires constant scrolling/searching
2. **Code duplication**:
   - Uncommitted file operations duplicated 3x (unstaged/staged/all)
   - Uncommitted diff operations duplicated 3x
   - Status parsing duplicated across 6 functions
   - ~300 lines of duplicated code
3. **Mixed abstraction levels**: Low-level parsing mixed with high-level operations
4. **Poor grouping**: Functions only separated by comments

### What Works Well (Keep)

1. **Abstraction boundary**: Git operations injectable via functional options in app.Model
2. **Pure parsing**: Parsing separated from I/O, easy to test
3. **Test coverage**: Comprehensive edge case coverage (1017-line test file)
4. **Error handling**: Consistent descriptive errors
5. **No external dependencies**: Only stdlib
6. **E2E test isolation**: E2E tests use mocks, not real git package (proper dependency injection)

## Decision

Restructure the package using layered architecture:

```
internal/git/
├── git.go              - Public API only (exported functions)
├── types.go            - Shared types and parsing helpers
├── commands/           - Low-level git command execution
│   ├── log.go          - git log execution + parsing
│   ├── diff.go         - git diff base function (eliminates duplication)
│   ├── show.go         - git show execution
│   └── common.go       - Shared command utilities
└── operations/         - High-level operations composing commands
    ├── commits.go      - Commit operations (Fetch, Resolve, etc.)
    ├── files.go        - File change operations
    └── diffs.go        - Diff operations
```

### Key Improvements

1. **Eliminate duplication**: Extract common patterns (e.g., `fetchFileChanges(flags...)`)
2. **Clear layering**: commands → operations → public API
3. **Package encapsulation**: `commands/` and `operations/` are internal packages
4. **Composable primitives**: Base functions can be composed for new features
5. **Maintain compatibility**: Public API in `git.go` unchanged (zero breaking changes)

## Alternatives Considered

### Option 1: Domain-Driven (commits/, files/, diffs/ packages)
- **Pro**: Clear domain separation, natural extension points
- **Con**: Over-engineered (6 packages for 20 functions), longer import paths
- **Rejected**: Too complex for current size

### Option 2: Parser/Executor Separation
- **Pro**: Clear layering, parsers testable in isolation
- **Con**: Doesn't address duplication, some files still large
- **Rejected**: Doesn't solve core problems

### Option 3: Git Object Model (commits.go, workdir.go, index.go, tree.go)
- **Pro**: Maps to Git's conceptual model, intuitive for Git users
- **Con**: Some files still large, requires Git internals knowledge
- **Rejected**: Good long-term but overkill now; consider for future growth

### Option 4: Minimal Split (6-8 files in same package)
- **Pro**: Minimal disruption, zero breaking changes, immediate value
- **Con**: Doesn't add abstractions or reduce duplication
- **Rejected**: Too superficial, doesn't address duplication

### Option 5: Layered Architecture (Chosen)
- **Pro**: Eliminates duplication, clear layering, composable, maintainable
- **Con**: More complex refactoring, higher initial effort
- **Accepted**: Best long-term solution, worth the upfront investment

## Implementation Plan

### Phase 1: Establish Structure
1. Create package structure (commands/, operations/, types.go)
2. Move parsing helpers to types.go
3. Split tests to match structure
4. Ensure all tests still pass (TDD)

### Phase 2: Extract Commands Layer
1. Create `commands/common.go` with shared exec helpers
2. Extract `commands/log.go` (git log execution)
3. Extract `commands/diff.go` with base function (eliminates duplication)
4. Extract `commands/show.go` (git show execution)
5. Ensure all tests pass after each extraction

### Phase 3: Extract Operations Layer
1. Create `operations/commits.go` (uses commands/log.go)
2. Create `operations/files.go` (composes commands)
3. Create `operations/diffs.go` (composes commands, eliminates 3x duplication)
4. Ensure all tests pass

### Phase 4: Finalize Public API
1. Keep only exported functions in `git.go`
2. Wire operations layer to public API
3. Verify no breaking changes in consumers (main.go, app/, ui/states/)
4. Run full test suite including E2E tests

### Phase 5: Add Enforcement Test
1. Add test to ensure `exec.Command("git"` only called in git package
2. Prevents future violations of package boundaries

## Consequences

### Positive
- **Reduced duplication**: ~300 lines of duplicated code eliminated
- **Better organization**: Clear file/package boundaries, easy to locate functions
- **Easier maintenance**: Changes to git command execution in one place
- **Composable**: Base commands can be reused for new features
- **Encapsulated**: Internal packages prevent misuse
- **Enforced boundaries**: Test prevents git calls outside git package

### Negative
- **More files**: 1 file → ~10 files (but smaller, focused files)
- **More complexity**: 3-layer architecture vs flat structure
- **Refactoring effort**: Requires careful TDD to avoid regressions

### Neutral
- **Import path unchanged**: Still `internal/git` for public API
- **Public API unchanged**: No breaking changes to consumers
- **Test compatibility**: Existing mocks/injection patterns continue to work

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Break existing functionality | Use TDD: ensure all tests pass after each step |
| Break E2E tests | E2E tests use mocks, unaffected by internal restructuring |
| Miss git command calls outside package | Add enforcement test using grep/static analysis |
| Over-engineer for current needs | Keep it simple: only 3 layers, no interfaces yet |

## Validation

Success criteria:
- [ ] All unit tests pass
- [ ] All E2E tests pass
- [ ] Zero breaking changes to consumers (main.go, app/, ui/)
- [ ] ~300 lines of duplication eliminated
- [ ] Each file < 300 lines
- [ ] Enforcement test prevents git calls outside package
- [ ] Code review confirms improved maintainability
