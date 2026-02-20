# AI Agent Testing: Tape Runner

## Problem

AI agents need a way to test the compiled Splice binary in a real terminal environment. Running `./splice` directly fails because it requires an interactive terminal (`/dev/tty`), which isn't available in most test environments.

## Solution

The **tape-runner** tool (`test/agent/tape-runner/`) enables AI agents to test Splice by:
1. Building the splice binary from current source
2. Running it in a tmux session (real terminal environment)
3. Sending keyboard input via tape file commands
4. Capturing snapshots (text, ANSI, or PNG)

## Tape File Format

Tape files use a simple DSL similar to VHS:

```tape
# Configuration
Width 120
Height 40

# Actions
Sleep 1s
Textshot initial      # Plain text snapshot (.txt)
Ansishot initial      # Text with ANSI codes (.ansi)
Snapshot initial      # PNG screenshot (.png)

Send jjj              # Send keys
Sleep 200ms
Textshot after-nav

Send <enter>
Sleep 200ms
Textshot files-view
```

**Commands:**
- `Width <cols>` / `Height <rows>` - Terminal size
- `Send <keys>` - Send keyboard input (supports `<enter>`, `<esc>`, `<ctrl-c>`, etc.)
- `Sleep <duration>` - Wait (e.g., 200ms, 1s)
- `Textshot [name]` - Capture plain text without ANSI codes
- `Ansishot [name]` - Capture text with ANSI color codes
- `Snapshot [name]` - Capture PNG image (requires freeze)

## Usage

```bash
./run-tape --help           # View documentation
./run-tape my-test.tape     # Run a test
```

Output is saved to `.test-output/<timestamp>/` with numbered snapshots:
- `001-initial.txt`
- `002-initial.png`
- `003-after-nav.txt`

## Why This Approach?

### What We Tried

**VHS** (https://github.com/charmbracelet/vhs): Popular terminal recorder with similar tape format
- ✅ Great for demo GIFs and documentation
- ✅ Can output text files for verification
- ❌ **Cannot do text snapshots at specific points** - only continuous frame capture
- ❌ `Screenshot` command only produces PNG images, not text
- ❌ `Hide`/`Show` commands reduce frames but don't enable discrete snapshots

VHS is designed for continuous video recording, not discrete snapshot testing.

### Why Tape Runner?

**Advantages:**
- ✅ Discrete snapshots at exact moments (not continuous capture)
- ✅ Multiple output formats (text, ANSI, PNG)
- ✅ Simple DSL that AI agents can easily generate
- ✅ Tests the actual compiled binary end-to-end
- ✅ Minimal dependencies (just tmux, optionally freeze for PNG)

**Current Limitations:**
- ⚠️ Requires tmux to be installed
- ⚠️ Requires freeze for PNG snapshots
- ⚠️ Not as polished as VHS

## Future Improvements

A **pure Go implementation** would be better:
- No external dependencies (tmux, freeze)
- Easier installation for users
- More portable across platforms
- Could use PTY library (github.com/creack/pty) to create pseudo-terminal
- Could use internal rendering for PNG generation

The current tmux-based approach was faster to implement and validates the concept. Migration to a Go implementation can happen later without changing the tape file format.

## For AI Agents

This tool is specifically designed for AI agent testing:

1. **Simple DSL**: Easy to generate tape files from natural language
2. **Deterministic**: Same input → same output (with proper sleep delays)
3. **Verifiable**: Text snapshots can be compared with git diff
4. **Visual**: PNG snapshots for visual regression testing
5. **Self-contained**: Builds from source, doesn't require pre-built binary

### Example Workflow

```bash
# AI agent creates tape file
cat > test.tape <<EOF
Width 120
Height 40
Sleep 1s
Textshot start
Send jjj
Textshot after-navigation
EOF

# Run test
./run-tape test.tape

# Verify output
git diff .test-output/*/001-start.txt
git diff .test-output/*/002-after-navigation.txt
```

## Complementary Testing

Tape-runner is for **integration testing the compiled binary**. For unit/component testing, continue using the existing E2E test framework (test/e2e/) which uses Bubbletea's in-process testing with byte buffers.

**Testing Strategy:**
- **E2E tests** (primary): Fast, precise, in-process - for most testing
- **Tape-runner** (secondary): Slower, black-box - for integration smoke tests

## Dependencies

- **tmux**: Required for all tests
  - Install: `brew install tmux`
- **freeze**: Required only for PNG snapshots
  - Install: `brew install charmbracelet/tap/freeze`

The tool checks dependencies upfront and fails fast with clear error messages.

## References

- Tape runner source: `test/agent/tape-runner/main.go`
- Helper script: `./run-tape`
- CLAUDE.md: Usage instructions for AI agents
