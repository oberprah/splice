# Testing Guidelines

## Testing Strategy

1. **Unit tests** for most functionality (fast, focused)
2. **Golden files** for TUI rendering (visual correctness)
3. **E2E tests** for critical user workflows only

## Mocking External Dependencies

**Never call git commands directly in tests.** All external dependencies are injected via functional options, allowing tests to use mocks instead. This keeps tests fast, reliable, and deterministic.

```go
// Production: real git (default)
m := ui.NewModel()

// Tests: inject mocks
m := ui.NewModel(
    ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
)
```

For pure functions without external dependencies, write standard Go tests.

## Naming Convention

```
Unit tests:  Test<Function>_<Scenario>       e.g. TestParseUnifiedDiff_SimpleHunk
State tests: Test<State>_<Method>_<Scenario> e.g. TestLogState_View_RendersCommits
E2E tests:   Test<Feature>                   e.g. TestBasicNavigation
```

## Test Utilities

Use helpers from `internal/ui/testutils/helpers.go`:

```go
// Test data
commits := testutils.CreateTestCommits(50)
commits := testutils.CreateTestCommitsWithMessages([]string{"msg1", "msg2"})

// Error scenarios - pass an error to mock functions
ui.WithFetchCommits(testutils.MockFetchCommits(nil, errors.New("git failed")))
```

## Golden File Testing

Golden files are snapshot tests comparing rendered output against saved `.golden` files. Use them for all View tests.

Each state package has helpers in `testhelpers_test.go`: `mockContext` for terminal dimensions and `assert*Golden` functions for comparisons.

```go
func TestLogState_View_RendersCommits(t *testing.T) {
    commits := []git.GitCommit{
        {Hash: "abc123", Message: "First commit", Author: "Alice"},
    }
    s := LogState{Commits: commits, Cursor: 0}
    ctx := mockContext{width: 80, height: 24}

    output := s.View(ctx)
    assertLogViewGolden(t, output, "renders_commits.golden")
}
```

Update golden files when output intentionally changes:
```bash
go test ./... -update
```

## E2E Testing

Use E2E tests for full user journeys across multiple states. Use unit tests for everything else.

E2ETestRunner API (`e2e/helpers_test.go`):

```go
func TestNavigateToFilesView(t *testing.T) {
    commits := testutils.CreateTestCommits(10)
    m := ui.NewModel(
        ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
        ui.WithFetchFileChanges(testutils.MockFetchFileChanges([]git.FileChange{}, nil)),
    )

    runner := NewE2ETestRunner(t, m)

    runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
    runner.AssertGolden("navigate_to_files/1_initial.golden")

    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("navigate_to_files/2_after_down.golden")

    runner.Quit()
}
```
