# AI Agent Testing Examples

Practical examples showing how AI agents can test Splice using both the current E2ETestRunner and VHS.

---

## Current Approach: E2ETestRunner

### Example 1: Basic Navigation Test

```go
func TestNavigationDown(t *testing.T) {
    // Setup test data
    commits := testutils.CreateTestCommits(10)
    graphLayout := testutils.CreateBasicGraphLayout(commits)

    // Create model with mocked dependencies
    model := app.NewModel(
        app.WithInitialState(log.New(commits, graphLayout)),
        app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
        app.WithNow(func() time.Time {
            return time.Date(2026, 1, 9, 12, 0, 0, 0, time.UTC)
        }),
    )

    // Create test runner
    runner := NewE2ETestRunner(t, model)
    defer runner.Quit()

    // Set terminal size
    runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
    runner.AssertGolden("navigation_initial.golden")

    // Navigate down
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("navigation_down_1.golden")

    // Navigate down again
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("navigation_down_2.golden")
}
```

### Example 2: Multi-Key Sequence Test

```go
func TestVisualModeSelection(t *testing.T) {
    commits := testutils.CreateTestCommits(10)
    graphLayout := testutils.CreateBasicGraphLayout(commits)

    model := app.NewModel(
        app.WithInitialState(log.New(commits, graphLayout)),
        app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
    )

    runner := NewE2ETestRunner(t, model)
    defer runner.Quit()

    runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

    // Enter visual mode
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("V")})
    runner.AssertGolden("visual_mode_entered.golden")

    // Select multiple commits
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("jjj")})
    runner.AssertGolden("visual_mode_selected_3.golden")

    // Exit visual mode
    runner.Send(tea.KeyMsg{Type: tea.KeyEsc})
    runner.AssertGolden("visual_mode_exited.golden")
}
```

### Example 3: Screen Navigation Test

```go
func TestFilePreviewNavigation(t *testing.T) {
    commits := testutils.CreateTestCommits(5)
    graphLayout := testutils.CreateBasicGraphLayout(commits)

    files := []*domain.FileChange{
        {Path: "main.go", ChangeType: domain.Modified},
        {Path: "README.md", ChangeType: domain.Modified},
    }

    model := app.NewModel(
        app.WithInitialState(log.New(commits, graphLayout)),
        app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
        app.WithFetchFileChanges(testutils.MockFetchFileChanges(files, nil)),
    )

    runner := NewE2ETestRunner(t, model)
    defer runner.Quit()

    runner.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

    // Open file preview
    runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
    // Wait for async load (TestRunner handles this)
    runner.AssertGolden("files_preview_opened.golden")

    // Navigate in files list
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("files_preview_second_file.golden")

    // Go back to log view
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
    runner.AssertGolden("back_to_log.golden")
}
```

### AI Agent Task Template

When an AI agent needs to test a feature:

```
1. Identify required test data (commits, files, etc.)
2. Create model with mocked dependencies
3. Create E2ETestRunner
4. Define interaction sequence:
   - Window size
   - Key presses
   - Golden file checkpoints
5. Run test: go test ./path/to/test
6. If test fails:
   - Parse error output
   - Check golden file diff
   - Determine if bug or expected change
7. If expected change:
   - Run: go test ./path/to/test -update
   - Review changes with git diff
   - Commit updated golden files
```

---

## VHS Approach: .tape Scripts

### Example 1: Basic Smoke Test

**File**: `test/e2e/vhs/smoke.tape`

```tape
Output test/e2e/vhs/smoke.txt
Set Shell bash
Set Width 80
Set Height 24

# Build binary
Hide
Type "go build -o splice ."
Enter
Wait
Show

# Run splice
Type "./splice"
Enter
Sleep 1s

# Basic navigation
Type "j"
Sleep 200ms
Type "j"
Sleep 200ms
Type "k"
Sleep 200ms

# Quit
Type "q"
Wait
```

**Run**:
```bash
vhs test/e2e/vhs/smoke.tape
git diff --exit-code test/e2e/vhs/smoke.txt
```

### Example 2: Multi-Commit Selection

**File**: `test/e2e/vhs/visual_mode.tape`

```tape
Output test/e2e/vhs/visual_mode.txt
Output test/e2e/vhs/visual_mode.gif
Set Width 120
Set Height 30
Set Theme "Dracula"

# Build and run
Hide
Type "go build -o splice ."
Enter
Wait
Show

Type "./splice"
Enter
Sleep 1s

# Enter visual mode
Type "V"
Sleep 300ms
Screenshot test/e2e/vhs/visual_mode_enter.png

# Select commits
Type "jjj"
Sleep 500ms
Screenshot test/e2e/vhs/visual_mode_selected.png

# Exit visual mode
Escape
Sleep 300ms

# Quit
Type "q"
Wait
```

**Run**:
```bash
vhs test/e2e/vhs/visual_mode.tape
# Generates:
# - visual_mode.txt (for testing)
# - visual_mode.gif (for docs)
# - visual_mode_enter.png (screenshot)
# - visual_mode_selected.png (screenshot)
```

### Example 3: File Preview Navigation

**File**: `test/e2e/vhs/file_preview.tape`

```tape
Output test/e2e/vhs/file_preview.txt
Set Width 120
Set Height 40

# Setup
Hide
Type "go build -o splice ."
Enter
Wait
Show

# Run
Type "./splice"
Enter
Sleep 1s

# Open file preview
Enter
Sleep 500ms

# Navigate files
Down 2
Sleep 300ms
Down
Sleep 300ms

# Open diff view
Enter
Sleep 500ms

# Scroll diff
Down 5
Sleep 300ms

# Back to files
Type "q"
Sleep 300ms

# Back to log
Type "q"
Sleep 300ms

# Quit
Type "q"
Wait
```

### Example 4: Full User Workflow

**File**: `test/e2e/vhs/full_workflow.tape`

```tape
Output test/e2e/vhs/full_workflow.txt
Output docs/demo.gif
Set Width 120
Set Height 40
Set Theme "Catppuccin Mocha"
Set TypingSpeed 50ms

# Build
Hide
Type "go build -o splice ."
Enter
Wait
Show

# Start
Type "./splice"
Enter
Sleep 1.5s

# Explore commits
Type "Navigate log view"
Sleep 1s
Backspace 100
Down 3
Sleep 500ms
Up
Sleep 500ms

# View files for commit
Enter
Sleep 800ms

# Navigate files
Down 2
Sleep 500ms

# View diff
Enter
Sleep 800ms

# Scroll through diff
Down 5
Sleep 500ms
Down 5
Sleep 500ms

# Go back
Type "q"
Sleep 300ms
Type "q"
Sleep 300ms

# Visual mode selection
Type "V"
Sleep 300ms
Down 3
Sleep 500ms
Escape
Sleep 300ms

# Exit
Type "q"
Wait
```

This generates both:
- `full_workflow.txt` - for automated testing
- `docs/demo.gif` - for README documentation

### AI Agent VHS Template

When an AI agent needs to create a VHS test:

```
1. Determine test scenario
2. Create .tape file with structure:
   Output <name>.txt          # For testing
   Output <name>.gif          # Optional: for docs
   Set Width <W>
   Set Height <H>

   Hide                       # Hide build commands
   Type "go build -o splice ."
   Enter
   Wait
   Show                       # Show app interaction

   Type "./splice"
   Enter
   Sleep 1s

   <interaction sequence>

   Type "q"
   Wait

3. Run: vhs <file>.tape
4. Verify: git diff --exit-code <name>.txt
5. On failure:
   - If expected: git add <name>.txt
   - If bug: investigate and fix
6. Commit updated golden files
```

---

## Comparison: When to Use Which

### Use E2ETestRunner When:

- **Fast iteration needed** - in-process is much faster
- **Testing internal state** - need to inspect model
- **Unit/integration tests** - testing specific components
- **Deterministic tests** - no timing issues
- **Most tests** - this should be the default

### Use VHS When:

- **End-to-end smoke tests** - verify binary works
- **Creating documentation** - need GIFs for README
- **Integration testing** - test real binary in real terminal
- **Visual verification** - want screenshots at key points
- **CI/CD validation** - ensure binary runs in clean environment

---

## CI/CD Integration

### GitHub Actions: E2ETestRunner

```yaml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'
      - name: Run tests
        run: go test ./...
      - name: Run E2E tests
        run: go test ./test/e2e/...
```

### GitHub Actions: VHS

```yaml
name: VHS Tests
on: [push, pull_request]
jobs:
  vhs-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      - name: Build binary
        run: go build -o splice .

      - name: Run VHS smoke test
        uses: charmbracelet/vhs-action@v2
        with:
          path: 'test/e2e/vhs/smoke.tape'

      - name: Verify output unchanged
        run: git diff --exit-code test/e2e/vhs/smoke.txt

      - name: Upload artifacts on failure
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: vhs-output
          path: test/e2e/vhs/*.txt
```

---

## Golden File Management

### Updating Golden Files

**E2ETestRunner**:
```bash
# Update all golden files
go test ./... -update

# Review changes
git diff **/*.golden

# Commit if intentional
git add **/*.golden
git commit -m "Update golden files for new UI"
```

**VHS**:
```bash
# Re-run tape to regenerate
vhs test/e2e/vhs/smoke.tape

# Review changes
git diff test/e2e/vhs/smoke.txt

# Commit if intentional
git add test/e2e/vhs/smoke.txt
git commit -m "Update VHS golden files"
```

### Golden File Best Practices

1. **Always review diffs** before committing
2. **Understand why changed** - bug fix or expected UI change?
3. **Keep golden files in git** - track changes over time
4. **Use descriptive names** - `visual_mode_selected.golden` not `test1.golden`
5. **One assertion per golden file** - easier to debug failures
6. **Configure git attributes** for golden files:

```gitattributes
# .gitattributes
*.golden -text
*.tape -text
test/e2e/vhs/*.txt -text
```

---

## Debugging Test Failures

### E2ETestRunner Failures

```bash
# Run single test with verbose output
go test -v ./test/e2e -run TestNavigationDown

# View golden file diff
git diff test/e2e/testdata/navigation_down_1.golden

# Update if expected
go test ./test/e2e -run TestNavigationDown -update
```

### VHS Failures

```bash
# Run tape manually
vhs test/e2e/vhs/smoke.tape

# View output
cat test/e2e/vhs/smoke.txt

# Compare with expected
git diff test/e2e/vhs/smoke.txt

# Debug: increase sleep times if timing issue
# Edit .tape file:
Sleep 1s -> Sleep 2s

# Re-run
vhs test/e2e/vhs/smoke.tape
```

---

## Summary

AI agents can effectively test Splice using:

1. **Primary: E2ETestRunner** - Fast, deterministic, full control
2. **Secondary: VHS** - Smoke tests, documentation, integration

Both approaches are AI-friendly and work in CI/CD pipelines. The current E2ETestRunner is already excellent and requires no changes.
