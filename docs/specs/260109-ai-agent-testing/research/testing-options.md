# AI Agent Testing Options for Splice TUI

## Summary

Research into enabling AI agents to test the Splice TUI application. The current E2E testing approach using Bubbletea's `WithInput()`/`WithOutput()` is already excellent for AI agent interaction. VHS can be added as a complementary black-box testing solution.

**Recommendation**: Keep current approach, optionally add VHS for integration testing and documentation generation.

---

## Current Approach: E2ETestRunner ⭐⭐⭐⭐⭐

### How It Works

Uses Bubbletea's official testing pattern with byte buffers instead of real terminals:
- `tea.WithInput()` - inject input messages
- `tea.WithOutput()` - capture output to buffer
- Mocked git dependencies via functional options
- Golden file snapshot testing

### Example

```go
runner := NewE2ETestRunner(t, model)
runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
runner.AssertGolden("1_initial.golden")
runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
runner.AssertGolden("2_after_down.golden")
runner.Quit()
```

### AI Agent Workflow

1. Generate Go test code using E2ETestRunner API
2. Specify input sequence (key presses, window sizes)
3. Define golden file checkpoints
4. Run `go test`
5. Parse test output for pass/fail

### Pros

- ✅ **Already implemented** - zero additional setup
- ✅ **Fastest** - in-process, no external dependencies
- ✅ **State access** - can inspect model internals
- ✅ **Deterministic** - mocked dependencies, no timing issues
- ✅ **Clean abstractions** - well-designed E2ETestRunner
- ✅ **CI/CD ready** - works headless
- ✅ **AI-friendly** - structured Go code

### Cons

- ❌ **Not black-box** - doesn't test compiled binary
- ❌ **Requires Go code** - AI must generate valid Go syntax
- ❌ **Less visual** - text-based assertions only

### Tradeoffs

- **Speed vs Coverage**: Fastest but doesn't catch binary-level issues
- **Precision vs Simplicity**: Full control but more verbose

### Current Status

✅ **Production ready** - no changes needed for AI agent testing

---

## Option 1: VHS (Charmbracelet) ⭐⭐⭐⭐

### How It Works

Script-driven terminal session recording and testing:
- Write `.tape` files (simple DSL) describing terminal interactions
- VHS spawns binary in real PTY via `ttyd` (terminal daemon)
- Outputs to GIF/MP4 (demos) or TXT/ASCII (testing)
- Text output used as golden files

**Repository**: https://github.com/charmbracelet/vhs

### .tape File Format

```tape
# Output configuration
Output demo.gif      # GIF for documentation
Output test.txt      # Plain text for testing

# Settings (must appear at top)
Set Shell bash
Set Width 80
Set Height 24
Set TypingSpeed 100ms
Set Theme "Catppuccin Mocha"

# Commands
Type "./splice"
Enter
Sleep 1s             # Wait for app to load
Type "jj"            # Navigate down twice
Sleep 500ms
Type "q"             # Quit
Wait                 # Wait for process to exit
```

### Available Commands

**Input**:
- `Type "text"` - Emulate typing
- `Enter`, `Space`, `Tab`, `Backspace`
- `Down`, `Up`, `Left`, `Right` (with optional count: `Down 3`)
- `Ctrl+C`, `Ctrl+D`, etc.

**Flow Control**:
- `Sleep 1s` - Fixed pause
- `Wait` - Wait for prompt (regex match, default 15s timeout)
- `Wait +Screen /text/` - Wait for text anywhere on screen

**Output**:
- `Screenshot frame.png` - Capture PNG
- `Hide` / `Show` - Stop/resume capture

### Output Formats

| Format | Purpose | Content |
|--------|---------|---------|
| `.gif` | Documentation | Animated GIF |
| `.mp4` | Documentation | MP4 video |
| `.webm` | Documentation | WebM video |
| `.txt` | **Testing** | Plain text output (likely ANSI stripped) |
| `.ascii` | **Testing** | Text output (likely ANSI preserved) |

**Note**: `.txt` vs `.ascii` difference is not officially documented. Testing suggests `.txt` may strip ANSI codes while `.ascii` preserves them, but this needs verification.

### AI Agent Workflow

1. Generate `.tape` file from test specification
2. Run `vhs test.tape` (outputs `test.txt`)
3. Compare `test.txt` against golden file with `git diff`
4. Report diff if different

**Example**:
```bash
# Generate golden file
vhs test.tape
git add test.txt
git commit -m "Add test baseline"

# Subsequent runs
vhs test.tape
git diff --exit-code test.txt || echo "Test failed!"
```

### Pros

- ✅ **Black-box testing** - tests actual compiled binary
- ✅ **Simple DSL** - AI can easily generate `.tape` files
- ✅ **Text output** - perfect for golden file testing
- ✅ **CI/CD ready** - official GitHub Action available
- ✅ **Bonus: documentation** - same scripts generate GIF demos
- ✅ **Active maintenance** - 17.8k stars, Charmbracelet official
- ✅ **Multiple outputs** - single tape generates multiple formats

### Cons

- ❌ **External dependencies** - requires `ttyd` and `ffmpeg` in PATH
- ❌ **Slower** - spawns actual process (~5s vs <1s)
- ❌ **Timing-dependent** - needs `Sleep`/`Wait` tuning
- ❌ **No state access** - can't inspect internal model state
- ❌ **Limited assertions** - only regex matching on output
- ❌ **Continuous capture only** - cannot do snapshot-only mode for text testing

**Note on Snapshot Mode**: VHS uses continuous frame capture by design. `Hide`/`Show` reduces frames but doesn't eliminate continuous capture. `Screenshot` can create PNG-only snapshots but not plain text. See [vhs-snapshot-mode.md](vhs-snapshot-mode.md) for detailed investigation.

### Tradeoffs

- **Speed vs Reality**: Slower but tests actual binary behavior
- **Simplicity vs Power**: Easy to write but limited introspection

### Installation & Setup

```bash
# macOS
brew install vhs

# Docker (no local install)
docker run --rm -v $PWD:/vhs ghcr.io/charmbracelet/vhs test.tape

# GitHub Actions
- uses: charmbracelet/vhs-action@v2
  with:
    path: 'test.tape'
```

### When to Use

- **Integration testing** - verify end-to-end workflows
- **Demo generation** - create GIFs for README
- **Smoke tests** - quick validation that binary runs
- **Regression testing** - ensure UI output doesn't change unexpectedly

---

## Option 2: teatest (Experimental) ⭐⭐⭐

### How It Works

Official Charmbracelet testing library (experimental):
- Wraps `tea.Model` in `TestModel`
- Simulates terminal without real I/O
- Provides `WaitFor` helpers for async operations
- Built-in golden file support

**Package**: `github.com/charmbracelet/x/exp/teatest`

### Example

```go
import "github.com/charmbracelet/x/exp/teatest"

func TestApp(t *testing.T) {
    model := app.NewModel()
    tm := teatest.NewTestModel(t, model,
        teatest.WithInitialTermSize(80, 24))

    // Wait for specific output
    teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
        return bytes.Contains(bts, []byte("ready"))
    })

    // Send input
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

    // Get final output
    output := tm.FinalOutput(t)
    teatest.RequireEqualOutput(t, output)  // Golden file
}
```

### Pros

- ✅ **Official** - from Bubbletea maintainers
- ✅ **Fast** - in-process like current approach
- ✅ **Async helpers** - `WaitFor` simplifies timing
- ✅ **Golden files** - built-in snapshot testing
- ✅ **No external dependencies** - pure Go

### Cons

- ❌ **Experimental** - API may change
- ❌ **Buggy** - user reported issues in testing
- ❌ **Less mature** - newer, less battle-tested
- ❌ **Color profile issues** - requires workarounds for CI

### Status

⚠️ **Not recommended** - user tried and found it buggy

---

## Option 3: PTY-based Solutions ⭐⭐⭐

### Libraries

1. **termtest** (ActiveState) - `github.com/ActiveState/termtest`
2. **tui-tester** (Aschey) - `github.com/aschey/tui-tester`

### How It Works

- Create real pseudo-terminal (PTY) using OS APIs
- Spawn binary as subprocess
- Send input programmatically
- Capture output with ANSI parsing
- Assert on terminal state (text, colors, cursor position)

### termtest Example

```go
suite := termtest.NewSuite()
defer suite.TearDown()

opts := termtest.Options{
    CmdName: "./splice",
    Args: []string{},
}

cp, err := termtest.NewTest(t, opts)
require.NoError(t, err)
defer cp.Close()

// Send input
cp.SendLine("j")
cp.SendLine("j")

// Wait for output
cp.Expect("expected text")
cp.SendLine("q")
cp.ExpectExitCode(0)
```

### tui-tester Example

```go
suite := tuitest.NewSuite()
tester, _ := suite.NewTester("./splice")
console, _ := tester.CreateConsole()

// Send input
console.SendString("j")
console.SendString(tuitest.KeyEnter)

// Wait for condition
state, _ := console.WaitFor(func(s tuitest.TermState) bool {
    return strings.Contains(s.Output(), "expected")
})

// Inspect terminal state
lines := state.OutputLines()
fgColor := state.ForegroundColor(0, 0)
```

### Pros

- ✅ **Black-box testing** - tests actual binary
- ✅ **Real terminal** - catches PTY-specific issues
- ✅ **Cross-platform** - Windows (ConPTY), macOS, Linux
- ✅ **Rich assertions** - can check colors, cursor position
- ✅ **Key constants** - `KeyEnter`, `KeyCtrlC`, etc.

### Cons

- ❌ **Slower** - spawns process, heavier than in-memory
- ❌ **More complex** - PTY management, ANSI parsing
- ❌ **tui-tester pre-1.0** - API may change
- ❌ **No state access** - can't inspect model internals
- ❌ **Harder debugging** - process boundaries complicate errors

### Tradeoffs

- **Reality vs Speed**: Real terminal but slower tests
- **Flexibility vs Complexity**: More capabilities but harder to set up

### When to Use

- Need to test terminal-specific behavior (colors, cursor)
- Want to verify binary behavior
- Need cross-platform terminal testing

### Status

⚠️ **Viable but complex** - only use if current approach insufficient

---

## Option 4: Lazygit-Style Code-Driven Tests ⭐⭐

### Overview

Lazygit evolved from keystroke recording to code-driven tests with abstractions:
- **GuiDriver** - low-level GUI manipulation (stable interface)
- **Input** - high-level action abstractions
- **Assert** - validation helpers
- **Shell** - git command execution during setup

### Philosophy

> "A big JSON blob tells you nothing about intention"

Tests written entirely in code with clear abstractions that express intent, not just keystrokes.

### Example

```go
var Commit = types.NewTest(types.NewTestArgs{
    Description: "Staging a couple files and committing",
    SetupRepo: func(shell types.Shell) {
        shell.CreateFile("myfile", "myfile content")
        shell.CreateFile("myfile2", "myfile2 content")
    },
    Run: func(shell types.Shell, input types.Input, assert types.Assert, keys config.KeybindingConfig) {
        assert.CommitCount(0)
        input.Select()
        input.NextItem()
        input.Select()
        input.PushKeys(keys.Files.CommitChanges)
        input.Type("my commit message")
        input.Confirm()
        assert.CommitCount(1)
        assert.HeadCommitMessage("my commit message")
    },
})
```

### Key Lessons

1. **Stable Interface** - GuiDriver minimal and stable
2. **Staged Assertions** - assert after each action, not just at end
3. **Intent-Driven** - abstractions express "what" not "how"
4. **Refactoring-Resistant** - keybinding changes don't break tests

### Pros

- ✅ **Intent-driven** - maintainable tests
- ✅ **Refactoring-resistant** - stable driver interface
- ✅ **Fast feedback** - staged assertions
- ✅ **Keybinding-independent** - can change bindings
- ✅ **Excellent error messages**

### Cons

- ❌ **Significant upfront work** - requires framework building
- ❌ **Application must be architected for testability**
- ❌ **Not a library** - must implement yourself
- ❌ **Only practical for large projects** - hundreds of integration tests

### Applicability to Splice

⚠️ **Not recommended** - Splice is smaller scale than Lazygit. Current approach is more appropriate.

However, **principles can be adopted**:
- Stable test interface (already have with Bubbletea)
- High-level helpers (E2ETestRunner is similar)
- Staged assertions (golden file checks)

### Status

⚠️ **Not applicable** - overkill for Splice's size

---

## Comparison Matrix

| Feature | Current E2E | VHS | teatest | PTY Solutions | Lazygit-style |
|---------|-------------|-----|---------|---------------|---------------|
| **Speed** | ⚡⚡⚡ Fastest | ⚡ Slow | ⚡⚡⚡ Fast | ⚡ Slow | ⚡⚡ Medium |
| **Setup** | ✅ None | ⚠️ External deps | ✅ Go only | ⚠️ Go deps | ❌ Framework build |
| **Black-box** | ❌ No | ✅ Yes | ❌ No | ✅ Yes | ❌ No |
| **State access** | ✅ Yes | ❌ No | ✅ Yes | ❌ No | ✅ Yes |
| **AI-friendly** | ✅ Go code | ⭐ .tape DSL | ✅ Go code | ⚠️ Go code | ⚠️ Complex |
| **Reliability** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ (exp) | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Golden files** | ✅ Manual | ✅ Built-in | ✅ Built-in | ⚠️ Manual | ✅ Custom |
| **CI/CD** | ✅ Easy | ✅ Action | ✅ Easy | ✅ Easy | ✅ Easy |
| **Maintenance** | ✅ None | ⚠️ External | ⚠️ Experimental | ⚠️ Pre-1.0 | ❌ High |

---

## Recommendations

### Primary: Keep Current Approach ✅

**E2ETestRunner** is already excellent for AI agents:
- Already working
- Fast iteration
- Full control
- AI can generate Go test code easily
- No additional dependencies

**No changes needed for AI agent testing.**

### Secondary: Add VHS for Integration Tests 🎬

Complement current approach with VHS for:
- End-to-end smoke tests
- Testing compiled binary
- Documentation (GIFs + testing)
- CI/CD integration testing

### Combined Workflow

```bash
# Fast unit/integration tests (current)
go test ./...              # <1s, AI generates Go code

# Slow E2E smoke tests (VHS)
vhs smoke-test.tape        # ~5s, AI generates .tape files
git diff --exit-code smoke-test.txt
```

### For AI Agent Testing Specifically

**VHS is optimal for AI generation** because:
1. **Simplest format** - `.tape` DSL easier than Go code
2. **No compilation** - AI doesn't need Go syntax knowledge
3. **Declarative** - describes behavior, not implementation
4. **Text output** - easy diff comparison
5. **One-time setup** - `brew install vhs` or Docker

**Example AI workflow**:
```
AI Prompt: "Test navigating down twice then quitting"

AI Generates test.tape:
---
Output test.txt
Type "./splice"
Enter
Sleep 1s
Type "jj"
Type "q"
---

Execute: vhs test.tape
Compare: diff test.txt expected.txt
```

---

## Decision

> **Keep current E2ETestRunner approach** as primary testing strategy. It's already AI-agent-friendly, fast, and reliable.

> **Optionally add VHS** for integration testing and demo generation. The dual-purpose nature (testing + documentation) provides good value, and the `.tape` DSL is extremely AI-friendly.

---

## References

- [VHS GitHub Repository](https://github.com/charmbracelet/vhs)
- [Writing Bubble Tea Tests](https://carlosbecker.com/posts/teatest/)
- [Lazygit Integration Testing Evolution](https://jesseduffield.com/IntegrationTests/)
- [Testing TUI Applications](https://blog.waleedkhan.name/testing-tui-apps/)
- [teatest Package Documentation](https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest)
- [termtest (ActiveState)](https://github.com/ActiveState/termtest)
- [tui-tester Package](https://pkg.go.dev/github.com/aschey/tui-tester)
