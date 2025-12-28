# Testing Guidelines

This document outlines the testing approach for Splice.

## Testing Philosophy

Splice uses a **layered testing strategy** that balances speed, maintainability, and confidence:

1. **Unit tests** for fast feedback on individual components
2. **Golden file tests** for visual correctness of TUI rendering
3. **E2E tests** for critical user workflows

## Unit Testing

### Test Organization

Tests are co-located with source files using the `*_test.go` naming convention:

```
internal/ui/states/
├── log_state.go
├── log_view.go
├── log_update.go
├── log_view_test.go      # Tests View() method
├── log_update_test.go    # Tests Update() method
└── testhelpers_test.go   # Package-level test utilities
```

### What to Test

**View tests** verify rendering logic:
- UI component layout and formatting
- Text truncation and wrapping at various terminal sizes
- Cursor and selection highlighting
- Viewport scrolling and pagination
- Edge cases (empty lists, single items, boundaries)

**Update tests** verify event handling:
- Navigation (j/k/g/G keys, arrow keys)
- State transitions between screens
- Cursor boundary conditions
- Command generation for async operations
- Message handling (success/error responses)
- Stale response handling

### Naming Conventions

**Test functions**: `Test<StateName>_<MethodName>_<Scenario>`
```go
func TestLogState_View_RendersCommits(t *testing.T) { ... }
func TestFilesState_Update_NavigationDown(t *testing.T) { ... }
```

**Table-driven tests** for multiple scenarios:
```go
tests := []struct {
    name       string
    cursor     int
    goldenFile string
}{
    {"first commit selected", 0, "selection_first.golden"},
    {"second commit selected", 1, "selection_second.golden"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test using tt fields
    })
}
```

### Mocking and Test Data

**Mock Context Pattern** (`internal/ui/states/testhelpers_test.go`):
```go
ctx := mockContext{width: 80, height: 24}
output := state.View(ctx)
```

The `mockContext` implements the `Context` interface, providing controlled terminal dimensions and stub implementations of git operations.

**Shared Test Utilities** (`internal/ui/testutils/helpers.go`):
```go
// Create test commits
commits := testutils.CreateTestCommits(50)
commits := testutils.CreateTestCommitsWithMessages([]string{"msg1", "msg2"})

// Mock functions for dependency injection
m := ui.NewModel(
    ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
    ui.WithFetchFileChanges(testutils.MockFetchFileChanges(files, nil)),
    ui.WithFetchFullFileDiff(testutils.MockFetchFullFileDiff(result, nil)),
)
```

Mock functions support both success and error cases by accepting optional error parameters.

### Best Practices

- Use `t.Helper()` in all helper functions for better error reporting
- Use fixed timestamps for deterministic output
- Test edge cases systematically (empty, single item, boundaries)
- Prefer table-driven tests for testing multiple similar scenarios

## Golden File Testing

### What Are Golden Files?

Golden files are **snapshot tests** that compare rendered output against saved reference files. This is essential for TUI applications where visual correctness is critical.

### How Golden Files Work

1. **Test execution**: Call `View()` method and capture the rendered string
2. **Comparison**: Compare output against the saved `.golden` file
3. **Update mode**: Use `-update` flag to regenerate golden files when output intentionally changes

### Directory Structure

**Unit test golden files**: `internal/ui/states/testdata/`
```
testdata/
├── log_view/
│   ├── renders_commits.golden
│   ├── selection_first.golden
│   └── split_view_wide.golden
├── files_view/
│   ├── renders_header.golden
│   └── binary_files.golden
└── diff_view/
    ├── all_line_types.golden
    └── inline_diff_rendering.golden
```

**E2E test golden files**: `e2e/testdata/`
```
testdata/
├── basic_navigation/
│   ├── 1_initial.golden
│   ├── 2_after_down_once.golden
│   └── 3_after_down_twice.golden
├── viewport_scrolling/
├── window_resize/
└── error_state/
```

Golden files are organized by feature/state with descriptive names.

### Using Golden Files

**In unit tests**:
```go
func TestLogState_View_RendersCommits(t *testing.T) {
    commits := createTestCommits(5)
    s := LogState{Commits: commits, Cursor: 0}
    ctx := mockContext{width: 80, height: 24}

    output := s.View(ctx)
    assertLogViewGolden(t, output, "renders_commits.golden")
}
```

**In E2E tests**:
```go
func TestBasicNavigation(t *testing.T) {
    runner := NewE2ETestRunner(t, model)

    runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
    runner.AssertGolden("basic_navigation/1_initial.golden")

    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("basic_navigation/2_after_down_once.golden")
}
```

### ANSI Code Handling

Golden files contain **plain text only** (no ANSI escape codes for colors/cursor control). The E2E test framework:
1. **Strips ANSI codes** using `stripAnsiCodes()` function
2. **Extracts latest frame** from accumulated output (Bubbletea redraws by moving cursor up)
3. **Saves readable text** for code review

This makes golden files human-readable and suitable for version control.

### Updating Golden Files

**Command**:
```bash
# Update all golden files
go test ./... -update

# Update specific package
go test ./internal/ui/states -update
go test ./e2e -update

# Update specific test
go test ./e2e -run TestBasicNavigation -update
```

**Workflow**:
1. Run tests to see failures
2. Review the changes (are they intentional?)
3. Update golden files: `go test ./... -update`
4. Verify tests pass with new golden files
5. Review changes: `git diff testdata/`
6. Commit updated golden files

**⚠️ Important**: Only update golden files when changes are **intentional**. Don't use `-update` as a shortcut to fix failing tests caused by bugs.

### When to Use Golden Files vs Regular Assertions

**Use golden files for**:
- TUI rendering (View methods)
- Complex multi-line formatted output
- Testing exact formatting (alignment, spacing, truncation)
- E2E full-screen tests
- Catching unexpected visual regressions

**Use regular assertions for**:
- Simple values (numbers, booleans, short strings)
- State changes (cursor position, indices)
- Logic (calculations, data transformations)
- Update tests (event handling)

Example:
```go
// Golden file - for visual output
output := s.View(ctx)
assertLogViewGolden(t, output, "renders_commits.golden")

// Regular assertion - for state changes
newState, cmd := s.Update(msg, ctx)
if newState.(LogState).Cursor != 1 {
    t.Errorf("Expected cursor at 1, got %d", newState.(LogState).Cursor)
}
```

## E2E Testing

### What E2E Tests Cover

E2E (End-to-End) tests verify complete user workflows by running the full Bubbletea application with mocked git operations.

**Current E2E tests** (`e2e/` directory):
1. **Basic Navigation** - j/k/g/G navigation, cursor movement
2. **Viewport Scrolling** - Scrolling through large lists
3. **Error State** - Handling git command failures
4. **Empty Repository** - Handling repositories with no commits
5. **Window Resize** - Responsive behavior at different terminal sizes

### E2ETestRunner

The `E2ETestRunner` (`e2e/helpers_test.go`) wraps a Bubbletea program for testing:

```go
runner := NewE2ETestRunner(t, model)

// Send messages to the program
runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

// Assert rendered output matches golden file
runner.AssertGolden("basic_navigation/1_initial.golden")

// Quit the program
runner.Quit()
```

**Key features**:
- Runs program asynchronously in a goroutine
- Captures input/output via buffers
- Polls output with timeout for assertions
- Strips ANSI codes and extracts latest frame
- Provides clean API for testing user interactions

### When to Write E2E Tests

**Write E2E tests for**:
- Full user journeys (load → navigate → select → view diff → quit)
- State transitions between screens
- Async behavior (loading, commands)
- Window resize and responsiveness
- Error scenarios affecting the entire app
- Critical workflows that must never break

**Write unit tests for**:
- Individual View/Update methods
- Pure logic and data transformation
- Internal helper functions
- Edge cases hard to trigger in E2E
- Fast iteration during development

### E2E vs Unit Tests

| Aspect | Unit Tests | E2E Tests |
|--------|-----------|-----------|
| **Scope** | Single method | Full program |
| **Setup** | Create state directly | Create model via `ui.NewModel()` |
| **Execution** | Call method | Send messages to running program |
| **Timing** | Synchronous | Asynchronous with polling |
| **Speed** | Very fast (μs) | Slower (ms) |
| **Granularity** | Specific functionality | Complete workflows |
| **Failure clarity** | Pinpoints exact method | Shows where journey breaks |

**Rule of thumb**: Unit tests are fast and focused; E2E tests are slower but test real user experience.

## Running Tests

### Basic Commands

```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./internal/ui/states
go test ./e2e

# Run specific test
go test ./e2e -run TestBasicNavigation

# Run with verbose output
go test ./... -v

# Run with coverage
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Development Workflows

**When adding a new feature**:
1. Write unit tests first (TDD): `go test ./internal/ui/states -run TestNewFeature -v`
2. Create golden files: `go test ./internal/ui/states -run TestNewFeature -update`
3. Add E2E test: `go test ./e2e -run TestNewFeature -update`
4. Run all tests: `go test ./...`

**When fixing a bug**:
1. Write failing test that reproduces the bug
2. Fix the bug in source code
3. Verify test passes: `go test ./internal/ui/states -run TestBugFix`
4. Run full suite: `go test ./...`

**When refactoring**:
1. Run tests before changes: `go test ./... -v`
2. Make refactoring changes
3. Run tests after changes: `go test ./... -v`
4. Update golden files if visual changes are intentional: `go test ./... -update`

## CI/CD Integration

Tests run automatically on:
- Every push to `main`
- Every pull request

Golden files must match exactly in CI (no `-update` flag). Failing tests block PR merging.

See `.github/workflows/ci.yml` for configuration.

## Tips and Tricks

### Debugging Test Failures

**View actual vs expected output**:
```bash
# Test will print both expected and actual in error message
go test ./internal/ui/states -run TestLogState_View_RendersCommits -v
```

**Update and diff golden files**:
```bash
# Update golden files
go test ./internal/ui/states -update

# Review changes
git diff internal/ui/states/testdata/

# Revert if changes are wrong
git restore internal/ui/states/testdata/
```

**Debug E2E tests**:
```go
// Add prints in test (visible with -v flag)
t.Logf("Output so far:\n%s", runner.out.String())

// Increase timeout for assertions
runner.AssertGolden("file.golden", 5*time.Second)
```

### Test Determinism

- **Always use fixed timestamps** in test data
- **Don't rely on system time** or random values
- **Keep golden files in version control** for reproducibility
- **Test at specific terminal sizes** (width/height)

### Performance Considerations

- **Unit tests are fast** - run them frequently during development
- **E2E tests are slower** - reserve for critical workflows
- **Balance coverage vs speed** - not everything needs E2E tests
- **Use mocking** to avoid real git operations

## Summary

Splice's testing strategy:
- ✅ **Unit tests** for most functionality (fast, focused)
- ✅ **Golden files** for visual correctness (TUI rendering)
- ✅ **E2E tests** for critical workflows (user journeys)
- ✅ **Clear separation** between test types
- ✅ **CI/CD integration** for continuous validation

This approach enables confident refactoring while maintaining visual correctness of the TUI.
