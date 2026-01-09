# AI Agent Testing for Splice

## Problem Statement

AI agents need a way to test the Splice TUI application to verify functionality and catch regressions. The challenge: TUI apps require interactive terminals (`/dev/tty`), which aren't available to AI agents running through CLI tools.

## Current State

✅ **Splice already has excellent AI-agent-testable infrastructure**:
- E2E tests using Bubbletea's `WithInput()`/`WithOutput()` with byte buffers
- No PTY required - tests run in-process
- Golden file snapshot testing
- All git dependencies mocked via functional options
- Works headless in CI/CD

**The current approach already solves the problem.**

## Options Evaluated

### 1. Current E2ETestRunner (Keep) ⭐⭐⭐⭐⭐

**Status**: Production ready, no changes needed

**How it works**: In-process testing with byte buffers, no real terminal needed

**Pros**: Fast, deterministic, state access, already working
**Cons**: Not black-box, requires Go code generation

### 2. VHS (Add Optionally) ⭐⭐⭐⭐

**Status**: Recommended as complement

**How it works**: Script terminal sessions with `.tape` files, output text for golden files

**Pros**: Black-box testing, simple DSL, generates demo GIFs, AI-friendly
**Cons**: Slower, external dependencies (ttyd, ffmpeg), timing-sensitive, **continuous capture only**

**Repository**: https://github.com/charmbracelet/vhs

**Note**: VHS cannot do snapshot-only mode for text testing (see [research/vhs-snapshot-mode.md](research/vhs-snapshot-mode.md) for detailed investigation of Hide/Show and Screenshot commands)

### 3. teatest (Skip) ⭐⭐⭐

**Status**: Not recommended - buggy, experimental

**How it works**: Official Bubbletea testing library (experimental)

**Pros**: Official, fast, async helpers
**Cons**: User reported bugs, experimental API

### 4. PTY Solutions (Skip) ⭐⭐⭐

**Status**: Viable but unnecessarily complex

**How it works**: Real pseudo-terminals with subprocess spawning

**Pros**: Black-box, real terminal behavior
**Cons**: Slow, complex, harder debugging

### 5. Lazygit-Style Framework (Skip) ⭐⭐

**Status**: Not applicable - overkill for Splice's scale

**How it works**: Custom test framework with GuiDriver abstractions

**Pros**: Intent-driven, maintainable
**Cons**: Significant upfront work, only practical for large projects

## Recommendation

### Primary Strategy: Keep Current Approach ✅

Your E2ETestRunner is already perfect for AI agents. No changes needed.

**AI agent workflow**:
1. Generate Go test code using E2ETestRunner API
2. Specify input sequence (key messages)
3. Define golden file checkpoints
4. Run `go test`
5. Parse results

### Optional Addition: VHS for Integration Tests 🎬

Add VHS for:
- End-to-end smoke tests
- Testing compiled binary
- Demo GIF generation for README
- CI/CD integration validation

**AI agent workflow**:
1. Generate `.tape` file (simple DSL)
2. Run `vhs test.tape` (outputs `test.txt`)
3. Compare against golden file with `git diff`

### Combined Testing Strategy

```bash
# Fast unit/integration (current) - <1s
go test ./...

# Slow E2E smoke tests (VHS) - ~5s
vhs smoke-test.tape
git diff --exit-code smoke-test.txt
```

## VHS Quick Reference

### .tape File Format

```tape
Output test.txt          # Text output for golden files
Output demo.gif          # GIF for documentation
Set Width 80
Set Height 24

Type "./splice"
Enter
Sleep 1s                 # Wait for app load
Type "jj"               # Navigate down twice
Type "q"                # Quit
Wait                    # Wait for process exit
```

### Output Formats

- `.gif`, `.mp4`, `.webm` - Video/image (documentation)
- `.txt` - Plain text (testing) - likely ANSI stripped
- `.ascii` - ASCII text (testing) - likely ANSI preserved

**Note**: Difference between `.txt` and `.ascii` not officially documented

### Installation

```bash
# macOS
brew install vhs

# Docker
docker run --rm -v $PWD:/vhs ghcr.io/charmbracelet/vhs test.tape

# GitHub Actions
- uses: charmbracelet/vhs-action@v2
  with:
    path: 'test.tape'
```

## Why VHS is AI-Friendly

1. **Simple DSL** - easier to generate than Go code
2. **Declarative** - describes behavior, not implementation
3. **No compilation** - AI doesn't need Go syntax knowledge
4. **Text output** - easy comparison with golden files
5. **Dual purpose** - testing + documentation generation

## Next Steps

1. **Do nothing** - current tests already work for AI agents ✅
2. **Optional**: Experiment with VHS
   - Install: `brew install vhs`
   - Create test tape: `test/e2e/smoke.tape`
   - Generate baseline: `vhs test/e2e/smoke.tape`
   - Add to CI for integration testing

## References

See [research/testing-options.md](research/testing-options.md) for detailed analysis of all options.

Key resources:
- [VHS Repository](https://github.com/charmbracelet/vhs)
- [Splice E2E Tests](../../test/e2e/)
- [Testing Guidelines](../../guidelines/testing-guidelines.md)
