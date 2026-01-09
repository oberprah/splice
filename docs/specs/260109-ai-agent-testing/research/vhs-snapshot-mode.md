# VHS Snapshot-Only Mode Investigation

**Date**: 2026-01-09
**Question**: Can VHS capture at dedicated steps (snapshots) instead of continuously?
**Answer**: Partially, but not for text-based golden file testing

---

## Executive Summary

**For text-based testing**: ❌ VHS cannot do true snapshot-only mode
- `Hide`/`Show` reduces frames but doesn't eliminate continuous capture
- Commands during `Hide` still appear in output

**For visual testing**: ✅ `Screenshot` command works independently
- Can create PNG snapshots without continuous video
- But PNGs aren't usable for text golden files

**Recommendation**: Keep E2E tests for snapshot-based testing, use VHS for demos/integration

---

## Hide/Show Commands

### Official Documentation

From VHS README:
> Hide: Hide the typing animation from the output
> Show: Show the typing animation in the output

### What Hide/Show Actually Does

**Implementation** (from VHS source code):
```go
Hide:
- Calls PauseRecording()
- Stops capturing frames
- Commands continue executing in PTY

Show:
- Calls ResumeRecording()
- Resumes capturing frames from current terminal state
```

### Critical Misunderstanding

Many users (including initial AI analysis) misunderstand `Hide`:

**What people think Hide does**:
```tape
Hide
Type "secret command"
Enter
Show
```
❌ "Secret command won't appear in output"

**What Hide actually does**:
```tape
Hide
Type "secret command"  # Stops frame capture
Enter                  # Still executes in terminal
Show                   # Resumes - terminal state includes "secret command" output!
```
✅ "Secret command" WILL appear in output because terminal state contains it

### Evidence: GitHub Issue #130

From VHS issue tracker:
> "The Hide command appears to not hide the command output, only the typing itself"

Multiple users report this confusion about Hide behavior.

### Test Results: Hide/Show

**Test Pattern**:
```tape
Output test_hide_show.txt
Set Width 120
Set Height 120

Type "echo 'VISIBLE 1'"
Enter
Sleep 500ms

Hide
Type "echo 'HIDDEN'"
Enter
Sleep 500ms
Show

Type "echo 'VISIBLE 2'"
Enter
```

**Result**:
```
# From test_hide_show.txt:
> echo 'VISIBLE 1'
VISIBLE 1
> echo 'HIDDEN'    # ← Command appears!
HIDDEN              # ← Output appears!
> echo 'VISIBLE 2'
VISIBLE 2
```

**Frames Captured**: 11 frames (vs 15-20 without Hide/Show)

**Conclusion**: Hide/Show reduces frame count but does NOT hide command results

---

## Screenshot Command

### Official Documentation

From VHS README:
> Screenshot: Take a screenshot of the terminal

### How Screenshot Works

**Behavior**:
- Creates PNG image of current terminal state
- Can be used with OR without `Output` directive
- Works independently of continuous capture
- PNG includes styling, margins, theme

### Screenshot-Only Mode

**Test Pattern**:
```tape
# No Output directive!

Type "./splice"
Enter
Sleep 2s
Screenshot snapshot1.png

Type "jj"
Sleep 300ms
Screenshot snapshot2.png

Type "q"
Wait
```

**Result**:
- ✅ Creates `snapshot1.png` and `snapshot2.png`
- ✅ NO continuous frame capture
- ✅ NO txt/gif files generated
- ✅ True snapshot-only mode!

**File Output**:
```bash
ls -lh snapshot*.png
# -rw-r--r-- snapshot1.png  # PNG image
# -rw-r--r-- snapshot2.png  # PNG image

ls *.txt *.gif
# (no files found)
```

### The Limitation for Testing

**Problem**: PNG is not plain text

```
Screenshot output:
┌─────────────────────────┐
│  Styled PNG Image       │
│  - Colors               │
│  - Fonts                │
│  - Borders              │
│  - Margins              │
│  Not parseable as text! │
└─────────────────────────┘
```

**Cannot be used for**:
- ❌ Traditional golden file testing (text diff)
- ❌ `grep` patterns
- ❌ Simple text assertions
- ❌ Version control friendly diffs

**Could be used for**:
- ✅ Visual regression testing (image comparison)
- ✅ Documentation images
- ✅ Manual review
- ✅ Pixel-perfect UI verification

---

## Combined Approaches Tested

### Test 1: Hide/Show + Screenshot

**Pattern**:
```tape
Output test_combined.txt

Hide
Type "./splice"
Enter
Sleep 2s
Show
Sleep 10ms     # Minimal capture
Screenshot state1.png
Hide

Type "jj"
Sleep 300ms
Show
Sleep 10ms
Screenshot state2.png
Hide

Type "q"
Wait
```

**Result**:
- Frames captured: 14
- TXT file: Still has continuous frames
- Screenshots: Created successfully

**Conclusion**: Combination doesn't solve text-based snapshot problem

### Test 2: Low Framerate + Hide/Show

**Pattern**:
```tape
Output test_low_fps.txt
Set Framerate 1      # Very low framerate

Hide
Type "./splice"
Enter
Sleep 2s
Show
Sleep 100ms
Hide

Type "jj"
Sleep 300ms
Show
Sleep 100ms
Hide

Type "q"
Wait
```

**Result**:
- Frames captured: 10
- TXT file: Fewer frames but still continuous
- Execution time: Slightly faster

**Conclusion**: Reduces frames but doesn't eliminate continuous capture

### Test 3: Minimal Show Duration

**Pattern**:
```tape
Output test_minimal.txt
Set Width 120
Set Height 120

Hide
Type "./splice"
Enter
Sleep 2s
Show
Sleep 10ms     # Capture just 1 frame?
Hide

Type "jj"
Show
Sleep 10ms
Hide

Type "q"
Wait
```

**Result**:
- Frames captured: 7 (reduced from ~15-20)
- TXT file: Still multiple frames at each Show point
- Best compromise but not true snapshot

**Conclusion**: Closest to snapshot mode but still captures multiple frames

---

## Why VHS Can't Do Text Snapshots

### Architectural Design

VHS is fundamentally a **video recorder**:

```
┌─────────────────────────────────────┐
│  VHS Architecture                    │
├─────────────────────────────────────┤
│                                      │
│  1. Start PTY + Recording Goroutine │
│  2. Capture frames at interval      │
│     (framerate = 60 FPS default)    │
│  3. Store frames in buffer          │
│  4. Generate outputs:               │
│     - GIF: Frames as animation      │
│     - TXT: All frames concatenated  │
│                                      │
└─────────────────────────────────────┘
```

**Key constraint**: TXT output is ALL captured frames, no way to select "just one"

### Recording Logic

From VHS source code:

```go
// Recording goroutine captures continuously
for recording {
    if !paused {
        captureFrame()  // Capture terminal state
    }
    sleep(frameInterval)  // Based on framerate
}
```

**Implications**:
- Recording runs continuously until session ends
- `Hide` sets `paused = true` (stops capture)
- `Show` sets `paused = false` (resumes capture)
- Multiple frames captured during any `Show` period
- ALL frames dumped to TXT output

### Why Screenshot Doesn't Help

**Screenshot implementation**:
```go
Screenshot:
- Render current terminal to PNG
- Apply styling, borders, theme
- Write PNG file
- Independent of frame buffer
```

**Problems for text testing**:
- PNG is binary image format
- Not human-readable text
- Can't diff easily in git
- Requires image comparison tools
- Not compatible with golden file pattern

---

## Comparison: VHS vs E2E Tests

### What Each Tool Provides

| Feature | VHS | E2E Tests |
|---------|-----|-----------|
| **Capture Model** | Continuous frames | Discrete snapshots |
| **Output Format** | TXT (all frames) or PNG (visual) | Plain text golden files |
| **Frame Control** | Hide/Show (reduces, doesn't eliminate) | Exact control per assertion |
| **Text-Based Testing** | ❌ Still continuous in TXT | ✅ True snapshots |
| **Visual Testing** | ✅ Screenshot PNGs | ❌ Text only |
| **Speed** | 🐌 ~5-10s | ⚡ <1s |
| **Binary Testing** | ✅ Tests real binary | ❌ Tests model |

### Use Case Fit

**E2E Tests are better for**:
```go
// Precise state verification at specific points
runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
runner.AssertGolden("step1.golden")  // Exact snapshot

runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
runner.AssertGolden("step2.golden")  // Exact snapshot
```

**VHS is better for**:
```tape
# Visual documentation
Output demo.gif
Type "./splice"
Enter
Type "jjj"
# Smooth animation showing full interaction
```

---

## Detailed Test Results

### Test Suite Summary

| Test | Pattern | Frames | TXT Output | Usable for Text Testing? |
|------|---------|--------|-----------|-------------------------|
| 1. Normal | Continuous capture | 15-20 | All frames | ❌ Too many frames |
| 2. Hide/Show | Hide → Show → Hide | 11 | Reduced frames | ❌ Still continuous |
| 3. Screenshot + TXT | Both outputs | 15 | All frames | ❌ PNG separate from TXT |
| 4. Minimal Show | 10ms Show periods | 7 | Reduced frames | ⚠️ Best compromise but not true snapshot |
| 5. Screenshot only | No Output | 0 | No TXT | ❌ PNGs only |
| 6. Low FPS | 1 FPS + Hide/Show | 10 | Reduced frames | ❌ Still continuous |

### Best Compromise: Minimal Show Pattern

**If you must use VHS for text testing**, this is the closest to snapshots:

```tape
Output test.txt
Set Width 120
Set Height 120
Set Framerate 1    # Low framerate

Hide
# Setup/navigation
Type "./splice"
Enter
Sleep 2s
Show
Sleep 100ms        # Brief capture
Hide

# Next step
Type "jj"
Sleep 300ms
Show
Sleep 100ms        # Brief capture
Hide

Type "q"
Wait
```

**Result**: 7-10 frames vs 15-20 frames

**Still not true snapshots**: Multiple frames per Show period

---

## Recommendations

### For Text-Based Golden File Testing

✅ **Use E2E Tests** (what you already have):

```go
func TestFeature(t *testing.T) {
    runner := NewE2ETestRunner(t, model)

    // Snapshot 1
    runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    runner.AssertGolden("state1.golden")

    // Snapshot 2
    runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
    runner.AssertGolden("state2.golden")
}
```

**Why**:
- True discrete snapshots
- Plain text output
- Fast execution
- Exact control

### For Visual Regression Testing

✅ **Use VHS Screenshot-only mode**:

```tape
# No Output directive!
Type "./splice"
Enter
Sleep 2s
Screenshot baseline_initial.png

Type "jj"
Sleep 300ms
Screenshot baseline_after_nav.png

Type "q"
Wait
```

**Why**:
- True snapshot-only (no continuous capture)
- Visual verification
- Pixel-perfect UI testing

**Setup needed**:
- Image comparison tool (pixelmatch, resemblejs, etc.)
- CI/CD integration for image diffs

### For Documentation and Demos

✅ **Use VHS continuous mode**:

```tape
Output README_demo.gif
Output demo.txt     # For CI verification
Set Theme "Dracula"

Type "./splice"
Enter
Sleep 2s
Type "jjj"
Enter
Sleep 1s
Type "q"
Wait
```

**Why**:
- Beautiful animated GIFs
- Shows full interaction flow
- Smooth transitions
- Dual-purpose (docs + CI)

### For Integration Testing

✅ **Use VHS with grep assertions**:

```tape
Output integration_test.txt

Type "./splice"
Enter
Sleep 2s
Type "jjj"
Enter
Sleep 1s
Type "q"
Wait
```

```bash
# Verify key states are present
grep "expected commit hash" integration_test.txt
grep "files ·" integration_test.txt
```

**Why**:
- Tests real binary
- Black-box validation
- CI/CD friendly
- Don't need exact frame matching

---

## Alternative Tools for Snapshot Testing

If you need text-based snapshot testing of actual binary (not in-process tests):

### 1. expect/pexpect

**Concept**: Send input, wait for output, capture snapshot

```python
import pexpect

child = pexpect.spawn('./splice')
child.expect('.*')
snapshot1 = child.before

child.send('j')
child.expect('.*')
snapshot2 = child.before
```

**Pros**: True snapshots at exact points
**Cons**: Requires Python, more complex

### 2. Custom Test Harness

**Concept**: Simple stdin/stdout capture

```go
func TestBinarySnapshot(t *testing.T) {
    cmd := exec.Command("./splice")
    stdin, _ := cmd.StdinPipe()
    stdout, _ := cmd.StdoutPipe()

    cmd.Start()
    time.Sleep(2 * time.Second)

    // Snapshot 1
    buf := make([]byte, 4096)
    n, _ := stdout.Read(buf)
    snapshot1 := string(buf[:n])

    // Send input
    stdin.Write([]byte("j\n"))
    time.Sleep(300 * time.Millisecond)

    // Snapshot 2
    n, _ = stdout.Read(buf)
    snapshot2 := string(buf[:n])

    // Compare against golden files
    compareGolden(t, snapshot1, "state1.golden")
    compareGolden(t, snapshot2, "state2.golden")
}
```

**Pros**: Full control, snapshots exactly when needed
**Cons**: More code, need to handle PTY properly

### 3. tmux + capture-pane

**Concept**: Run in tmux session, capture pane content

```bash
tmux new-session -d -s test "./splice"
sleep 2
tmux capture-pane -t test -p > snapshot1.txt

tmux send-keys -t test "j" Enter
sleep 0.3
tmux capture-pane -t test -p > snapshot2.txt

tmux kill-session -t test
```

**Pros**: Simple, real terminal behavior
**Cons**: Requires tmux, timing sensitive, platform-specific

---

## Conclusion

### Direct Answer to Original Question

> "Would it be possible to specify dedicated steps where we want to take a snapshot instead of continuously?"

**For text-based testing**: ❌ **No**, VHS cannot do true snapshot-only mode for plain text output

**For visual testing**: ✅ **Yes**, VHS Screenshot command can create PNG snapshots without continuous capture

### Why VHS Isn't Right for Text Snapshots

1. **Architectural**: VHS is a video recorder, not a snapshot tool
2. **TXT output**: Contains ALL captured frames, no way to select specific frames
3. **Hide/Show**: Reduces frames but doesn't eliminate continuous capture
4. **Screenshot**: Works for PNGs but not plain text

### What You Should Do

**Keep your current approach**:
- ✅ E2E tests for snapshot-based testing (perfect for this!)
- ✅ VHS for demos and integration testing (complementary)

**Don't try to force VHS into snapshot-only text mode** - it's not designed for that.

Your E2E tests already provide exactly what you're looking for:
```go
runner.Send(action)           // Execute
runner.AssertGolden("snap")   // Snapshot!
```

This is the right tool for the job! 🎯

---

## References

- [VHS GitHub Repository](https://github.com/charmbracelet/vhs)
- [VHS README - Hide Command](https://github.com/charmbracelet/vhs?tab=readme-ov-file#hide)
- [VHS README - Screenshot Command](https://github.com/charmbracelet/vhs?tab=readme-ov-file#screenshot)
- [VHS Issue #130 - Hide Command Behavior](https://github.com/charmbracelet/vhs/issues/130)
- [VHS Source Code - command.go](https://github.com/charmbracelet/vhs/blob/main/command.go)
- [VHS Source Code - vhs.go](https://github.com/charmbracelet/vhs/blob/main/vhs.go)
- Agent Research: Task a1d6c33 - Deep investigation of Hide/Show and Screenshot commands

---

## Test Files Generated

All test files from this research are in the repository root:

1. `test_hide_show.tape` - Hide/Show behavior test
2. `test_combined.tape` - Hide/Show + Screenshot combination
3. `test_minimal.tape` - Minimal Show duration pattern
4. `test_low_fps.tape` - Low framerate + Hide/Show
5. `test_screenshot_only.tape` - Screenshot-only mode (PNG)

Run any of these to reproduce findings:
```bash
vhs test_hide_show.tape
```
