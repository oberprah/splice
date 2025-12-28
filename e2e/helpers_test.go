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

// Quit sends a quit key message and waits for the program to finish
func (r *E2ETestRunner) Quit(timeout ...time.Duration) {
	r.t.Helper()

	quitTimeout := defaultQuitTimeout
	if len(timeout) > 0 {
		quitTimeout = timeout[0]
	}

	r.program.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

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

// extractLatestFrame attempts to extract the final rendered frame from accumulated terminal output
func extractLatestFrame(output string) string {
	// Strategy: Look for cursor movement patterns that indicate a redraw
	// Common pattern: "[<N>A" means move cursor up N lines, often marks frame start

	// Find the last occurrence of a cursor up movement
	// This heuristic assumes redraws start with cursor repositioning
	lastFrameIdx := -1
	for i := len(output) - 1; i >= 0; i-- {
		if i > 0 && output[i] == 'A' && output[i-1] >= '0' && output[i-1] <= '9' {
			// Found potential "[<N>A" pattern, look backwards for '['
			for j := i - 1; j >= 0; j-- {
				if output[j] == '[' {
					lastFrameIdx = j
					break
				}
				if output[j] < '0' || output[j] > '9' {
					break
				}
			}
			if lastFrameIdx != -1 {
				break
			}
		}
	}

	if lastFrameIdx == -1 {
		// No cursor movement found, return full output
		lastFrameIdx = 0
	}

	frame := output[lastFrameIdx:]

	// Strip cleanup codes at the end
	// Common cleanup patterns that indicate program termination:
	// - "[?2004l" - disable bracketed paste mode
	// - "[?25h" - show cursor
	// - "[?1002l", "[?1003l", "[?1006l" - disable mouse tracking
	// These typically appear at the very end when the program quits
	cleanupPatterns := []string{
		"[?2004l", // Most reliable indicator of cleanup
		"[?25h",   // Show cursor (end of program)
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

	return frame
}

// assertEventually polls the output buffer until the condition is met or timeout
func assertEventually(t *testing.T, out *bytes.Buffer, goldenFile string, timeout time.Duration) {
	t.Helper()

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

		time.Sleep(50 * time.Millisecond)
	}

	if !*update {
		// Timeout reached without match
		goldenPath := filepath.Join("testdata", goldenFile)
		expected, _ := os.ReadFile(goldenPath)
		t.Errorf("assertEventually timeout: output did not match golden file %s\nExpected:\n%s\n\nGot:\n%s",
			goldenFile, string(expected), lastOutput)
	}
}
