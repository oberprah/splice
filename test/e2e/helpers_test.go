package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var update = flag.Bool("update", false, "update golden files")

const defaultAssertTimeout = 1 * time.Second
const defaultQuitTimeout = 2 * time.Second

// E2ETestRunner wraps a Bubbletea program for e2e testing
type E2ETestRunner struct {
	t       *testing.T
	program *tea.Program
	in      *bytes.Buffer
	out     *bytes.Buffer
	done    chan struct{}
}

// NewE2ETestRunner creates a new test runner and starts the program in a goroutine
func NewE2ETestRunner(t *testing.T, model tea.Model) *E2ETestRunner {
	t.Helper()

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	p := tea.NewProgram(
		model,
		tea.WithInput(in),
		tea.WithOutput(out),
	)

	done := make(chan struct{})
	go func() {
		_, err := p.Run()
		if err != nil {
			t.Errorf("program error: %v", err)
		}
		close(done)
	}()

	return &E2ETestRunner{
		t:       t,
		program: p,
		in:      in,
		out:     out,
		done:    done,
	}
}

// Send sends a message to the program
func (r *E2ETestRunner) Send(msg tea.Msg) {
	r.program.Send(msg)
}

// Quit sends ctrl+c to quit the program and waits for it to finish.
// Uses ctrl+c instead of 'q' because 'q' exits visual mode when active.
func (r *E2ETestRunner) Quit(timeout ...time.Duration) {
	r.t.Helper()

	quitTimeout := defaultQuitTimeout
	if len(timeout) > 0 {
		quitTimeout = timeout[0]
	}

	r.program.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	select {
	case <-r.done:
		// Success
	case <-time.After(quitTimeout):
		r.t.Fatal("test timed out waiting for program to quit")
	}
}

// AssertGolden asserts that the output matches the golden file
// Timeout can be optionally provided, defaults to 1 second
func (r *E2ETestRunner) AssertGolden(goldenFile string, timeout ...time.Duration) {
	r.t.Helper()

	assertTimeout := defaultAssertTimeout
	if len(timeout) > 0 {
		assertTimeout = timeout[0]
	}

	assertEventually(r.t, r.out, goldenFile, assertTimeout)
}

// stripAnsiCodes removes ANSI escape sequences from text, leaving only readable content
func stripAnsiCodes(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Found start of ANSI escape sequence "\x1b["
			// Skip until we find the terminating character (a letter)
			j := i + 2
			for j < len(s) {
				ch := s[j]
				// Check if this is the terminating character
				if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
					// Skip the entire sequence including terminator
					i = j + 1
					break
				}
				// Continue scanning (digits, semicolons, question marks, etc.)
				j++
			}
			if j >= len(s) {
				// Malformed sequence at end of string, just skip to end
				break
			}
		} else {
			// Regular character, keep it
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// extractLatestFrame extracts the final rendered frame from accumulated terminal output
func extractLatestFrame(output string) string {
	// Strategy: Bubbletea redraws by sending "\x1b[<N>A" (cursor up N lines) then re-rendering
	// We find the LAST cursor-up sequence to identify the start of the final frame

	// Find all cursor-up movements "\x1b[<N>A"
	var frameStarts []int
	for i := 0; i < len(output)-3; i++ {
		if output[i] == '\x1b' && output[i+1] == '[' {
			// Look for digits followed by 'A'
			j := i + 2
			for j < len(output) && output[j] >= '0' && output[j] <= '9' {
				j++
			}
			if j < len(output) && output[j] == 'A' && j > i+2 {
				// Found "\x1b[<N>A" pattern - this marks a frame start
				frameStarts = append(frameStarts, i)
			}
		}
	}

	// Start from the last frame boundary (or beginning if none found)
	lastFrameIdx := 0
	if len(frameStarts) > 0 {
		lastFrameIdx = frameStarts[len(frameStarts)-1]
	}

	frame := output[lastFrameIdx:]

	// Strip cleanup codes that appear when the program quits
	cleanupPatterns := []string{
		"\x1b[?2004l", // disable bracketed paste mode
		"\x1b[?25h",   // show cursor
	}

	for _, pattern := range cleanupPatterns {
		if idx := strings.Index(frame, pattern); idx != -1 {
			// Find the last newline before cleanup codes
			lastNewline := strings.LastIndex(frame[:idx], "\n")
			if lastNewline != -1 {
				frame = frame[:lastNewline+1]
			} else {
				frame = frame[:idx]
			}
			break
		}
	}

	// Strip all ANSI escape codes for clean, readable golden files
	frame = stripAnsiCodes(frame)

	return frame
}

// assertEventually polls the output buffer until the condition is met or timeout
func assertEventually(t *testing.T, out *bytes.Buffer, goldenFile string, timeout time.Duration) {
	t.Helper()

	// In update mode, wait a bit to ensure render is complete before capturing
	if *update {
		time.Sleep(100 * time.Millisecond)
	}

	deadline := time.Now().Add(timeout)
	var lastOutput string

	for time.Now().Before(deadline) {
		rawOutput := out.String()
		lastOutput = extractLatestFrame(rawOutput)

		if *update {
			// Update mode: write the final output to golden file
			if time.Now().Add(50 * time.Millisecond).After(deadline) {
				// We're near the end of timeout, write the golden file
				goldenPath := filepath.Join("testdata", goldenFile)
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("failed to create golden file directory: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(lastOutput), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				t.Logf("Updated golden file: %s", goldenPath)
				return
			}
		} else {
			// Test mode: compare with golden file
			goldenPath := filepath.Join("testdata", goldenFile)
			expected, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
			}

			if lastOutput == string(expected) {
				// Match found!
				return
			}
		}

		// Small sleep to avoid busy-waiting and give Bubbletea time to render
		time.Sleep(10 * time.Millisecond)
	}

	if !*update {
		// Timeout reached without match
		goldenPath := filepath.Join("testdata", goldenFile)
		expected, _ := os.ReadFile(goldenPath)
		t.Errorf("assertEventually timeout: output did not match golden file %s\nExpected:\n%s\n\nGot:\n%s",
			goldenFile, string(expected), lastOutput)
	}
}
