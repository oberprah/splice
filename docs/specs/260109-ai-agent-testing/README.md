# AI Agent Testing for Splice - Documentation Index

This directory contains comprehensive research and findings on enabling AI agents to test the Splice TUI application.

---

## Quick Summary

**Conclusion**: Your E2E tests are already perfect for AI agent testing. VHS is a complementary tool for integration testing and documentation, but **cannot** do snapshot-only mode for text testing.

**Recommendation**:
- ✅ Keep E2E tests as primary testing approach
- ✅ Add VHS optionally for integration tests and demo GIFs

---

## Documentation Files

### Main Documents

#### [01-overview.md](01-overview.md)
High-level overview of AI agent testing options for Splice.

**Contents**:
- Problem statement
- Current state (E2E tests already work!)
- All options evaluated with ratings
- Recommendations
- Quick reference for VHS

**Read this first** for a quick understanding.

#### [02-examples.md](02-examples.md)
Practical code examples for both E2E tests and VHS.

**Contents**:
- E2E test examples (navigation, visual mode, screen changes)
- VHS `.tape` file examples (smoke tests, workflows)
- When to use which approach
- CI/CD integration examples
- Golden file management
- Debugging test failures

**Read this** when you want to write tests.

---

### Research Documents

#### [research/testing-options.md](research/testing-options.md)
Detailed analysis of all testing options investigated.

**Contents**:
- Current E2ETestRunner approach (⭐⭐⭐⭐⭐)
- VHS with full `.tape` format documentation (⭐⭐⭐⭐)
- teatest experimental library (⭐⭐⭐)
- PTY-based solutions (⭐⭐⭐)
- Lazygit-style framework (⭐⭐)
- Comparison matrix
- Pros/cons/tradeoffs for each option
- Real-world examples

**Read this** for comprehensive evaluation of all options.

#### [research/vhs-snapshot-mode.md](research/vhs-snapshot-mode.md) ⭐ **NEW**
Deep investigation into whether VHS can do snapshot-only mode.

**Contents**:
- What `Hide`/`Show` actually does (corrected understanding)
- What `Screenshot` actually does (PNG-only snapshots)
- Test results from 7 different approaches
- Why VHS can't do text-based snapshots
- Architectural limitations
- Recommendations for different use cases
- Alternative tools for snapshot testing

**Read this** if you're wondering about VHS snapshot capabilities.

---

## Experimental Results

### Testing Files in Repository Root

During research, several test files were created in the repository root:

- `test_basic.tape` - Basic smoke test
- `test_navigation.tape` - Navigation with screenshots
- `test_verify_behavior.tape` - Full workflow test
- `test_capture_timing.tape` - Capture timing investigation
- `test_framerate_*.tape` - Framerate impact tests
- `test_hide_show.tape` - Hide/Show behavior test
- `test_snapshot_attempt.tape` - Snapshot-only attempt
- `demo.gif` - Generated GIF demo
- `demo_dual.gif` / `demo_dual.txt` - Dual-output example
- `verify_test.sh` - Automated verification script
- `analyze_vhs_output.sh` - Output analysis script

### Additional Documentation

- `VHS_AGENT_TESTING_RESULTS.md` - Initial VHS testing results
- `VHS_GIF_TESTING.md` - GIF generation testing results
- `VHS_CAPTURE_MECHANISM.md` - How VHS captures output
- `VHS_VS_SNAPSHOT_TESTING.md` - Comparison with snapshot testing

---

## Key Findings

### 1. E2E Tests (Your Current Approach) ✅

**Perfect for AI agent testing!**

```go
runner := NewE2ETestRunner(t, model)
runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
runner.AssertGolden("state1.golden")  // Discrete snapshot!
```

**Characteristics**:
- ⚡ Fast (<1s per test)
- ✅ Discrete snapshots at exact points
- ✅ Full model state access
- ✅ Deterministic (mocked dependencies)
- ✅ Plain text golden files
- ✅ Already implemented and working

**Use for**: Most testing (primary approach)

### 2. VHS (Complementary Tool) ✅

**Good for integration testing and documentation!**

```tape
Output demo.gif       # For documentation
Output demo.txt       # For CI verification

Type "./splice"
Enter
Type "jjj"
Type "q"
```

**Characteristics**:
- 🐌 Slower (~5s per test)
- 🎬 Continuous frame capture
- ✅ Tests actual compiled binary
- ✅ Generates demo GIFs
- ✅ AI-friendly `.tape` DSL
- ❌ Cannot do snapshot-only for text

**Use for**: Integration smoke tests, documentation GIFs

### 3. VHS Snapshot Mode ❌

**Cannot do snapshot-only mode for text testing.**

**What was investigated**:
- `Hide`/`Show` commands - Reduces frames but still continuous
- `Screenshot` command - PNG-only snapshots (not plain text)
- Low framerate - Fewer frames but still continuous
- Various combinations - None achieve true text snapshots

**Why it doesn't work**: VHS is architecturally designed for continuous video capture, not discrete snapshots.

**Alternative**: Use `Screenshot` for PNG-based visual regression testing (separate from text testing)

---

## Recommendations

### For Different Use Cases

| Use Case | Tool | Why |
|----------|------|-----|
| **Precise state verification** | E2E Tests | Exact snapshots at specific points |
| **Fast iteration** | E2E Tests | <1s execution, in-process |
| **Internal state inspection** | E2E Tests | Full model access |
| **Text-based golden files** | E2E Tests | Plain text snapshots |
| **Most testing** | E2E Tests | Already perfect for this |
| | | |
| **Integration/smoke tests** | VHS | Tests real binary end-to-end |
| **Documentation GIFs** | VHS | Beautiful animated demos |
| **Visual regression** | VHS Screenshot | PNG snapshots for pixel-perfect testing |
| **Black-box validation** | VHS | No knowledge of internals needed |

### Testing Strategy

```
Splice Testing Approach:
├── Primary: E2E Tests (90% of tests)
│   └── Fast, precise, snapshot-based
├── Secondary: VHS (10% of tests)
│   ├── Smoke tests (integration)
│   └── Demo generation (documentation)
└── Optional: VHS Screenshots
    └── Visual regression (PNG comparison)
```

---

## Questions Answered

### Q: Can VHS do snapshot-only mode?

**A**: ❌ No, not for text-based testing.

- `Hide`/`Show` reduces frames but doesn't eliminate continuous capture
- `Screenshot` can do PNG-only snapshots (not plain text)
- VHS is designed for continuous video recording

**See**: [research/vhs-snapshot-mode.md](research/vhs-snapshot-mode.md)

### Q: What's the difference between .txt and .ascii output?

**A**: ⚠️ Not officially documented by VHS.

Likely:
- `.txt` - Plain text, ANSI codes stripped
- `.ascii` - Text with ANSI codes preserved

Both contain all captured frames concatenated.

### Q: Should I use VHS or E2E tests?

**A**: ✅ Use **both** - they're complementary!

- **E2E tests**: Primary testing (fast, precise, snapshots)
- **VHS**: Integration testing + documentation (slow, continuous, GIFs)

### Q: Can AI agents test Splice?

**A**: ✅ Yes! Your E2E tests already enable this.

AI agents can:
- Generate Go test code using E2ETestRunner
- Run tests with `go test`
- Parse results
- Update golden files when needed

VHS adds:
- Generate `.tape` files (simple DSL)
- Run with `vhs test.tape`
- Verify output with `grep` or `git diff`

---

## Next Steps

### If You Want to Try VHS

1. **Install VHS**:
   ```bash
   brew install vhs
   ```

2. **Create a smoke test**:
   ```bash
   cat > test/vhs/smoke.tape << 'EOF'
   Output test/vhs/smoke.txt
   Set Width 120
   Set Height 40

   Type "./splice"
   Enter
   Sleep 2s
   Type "jjj"
   Type "q"
   Wait
   EOF
   ```

3. **Run it**:
   ```bash
   vhs test/vhs/smoke.tape
   ```

4. **Verify**:
   ```bash
   grep "expected pattern" test/vhs/smoke.txt
   ```

### If You Want Visual Documentation

Create demo GIFs for your README:

```tape
Output docs/demo.gif
Set Width 1200
Set Height 600
Set Theme "Dracula"

Type "./splice"
Enter
Sleep 2s
Type "jjjj"
Sleep 1s
Enter
Sleep 1s
Type "qq"
Wait
```

Then in README.md:
```markdown
![Splice Demo](docs/demo.gif)
```

---

## Timeline of Research

1. **Initial investigation** - Discovered E2E tests already work for AI agents
2. **VHS exploration** - Tested VHS for complementary black-box testing
3. **GIF generation** - Confirmed VHS creates great documentation GIFs
4. **Capture mechanism** - Understood continuous frame capture model
5. **Snapshot investigation** - Deep dive into Hide/Show/Screenshot commands
6. **Final conclusion** - VHS cannot do snapshot-only for text, E2E tests are optimal

---

## References

### VHS Resources
- [VHS GitHub Repository](https://github.com/charmbracelet/vhs)
- [VHS README](https://github.com/charmbracelet/vhs/blob/main/README.md)
- [VHS Issue #130 - Hide Command Behavior](https://github.com/charmbracelet/vhs/issues/130)

### Bubbletea Testing
- [Writing Bubble Tea Tests](https://carlosbecker.com/posts/teatest/)
- [teatest Package](https://pkg.go.dev/github.com/charmbracelet/x/exp/teatest)

### Related
- [Testing TUI Applications](https://blog.waleedkhan.name/testing-tui-apps/)
- [Lazygit Integration Testing](https://jesseduffield.com/IntegrationTests/)

---

## Contact

For questions or feedback about this research:
- Review the documentation in this directory
- Check the test files in repository root
- Consult the Splice CLAUDE.md for testing guidelines
