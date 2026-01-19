package commands

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// execGit executes a git command with the provided arguments.
// Returns stdout, stderr, and error.
func execGit(args ...string) (string, string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// checkGitError examines stderr and returns a more descriptive error.
func checkGitError(stderr string, err error, context string) error {
	if err == nil {
		return nil
	}

	stderrStr := strings.TrimSpace(stderr)

	// Check for common git errors
	if strings.Contains(stderrStr, "not a git repository") {
		return fmt.Errorf("not a git repository")
	}

	if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
		return fmt.Errorf("invalid revision in %s", context)
	}

	// Return generic error with context
	return fmt.Errorf("%s failed: %v - %s", context, err, stderrStr)
}
