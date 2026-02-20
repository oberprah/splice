package architecture_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGitCommandsOnlyInGitPackage ensures git commands are only executed in internal/git package.
// All git operations should go through the git package API for proper abstraction and testability.
func TestGitCommandsOnlyInGitPackage(t *testing.T) {
	// Get project root (two levels up from test/architecture/)
	projectRoot := filepath.Join("..", "..")

	var violations []string

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, vendor, testdata
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor/") || strings.Contains(path, "testdata/") {
			return nil
		}

		// Get path relative to project root and normalize separators
		relPath, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		normalizedPath := filepath.ToSlash(relPath)

		// Allow git commands only in internal/git package (including subpackages)
		if strings.HasPrefix(normalizedPath, "internal/git/") {
			return nil
		}

		// Allow git commands in test files (they may set up test fixtures)
		if strings.HasPrefix(normalizedPath, "test/") {
			return nil
		}

		// Read file contents
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check for exec.Command("git") usage
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Skip comments
			if strings.Contains(line, "//") {
				commentIdx := strings.Index(line, "//")
				line = line[:commentIdx]
			}

			// Check for exec.Command("git"...)
			if strings.Contains(line, `exec.Command("git"`) {
				violations = append(violations,
					fmt.Sprintf("%s:%d: git command execution outside internal/git package", normalizedPath, i+1))
			}

			// Also check for exec.Command(`git`...) (backtick strings)
			if strings.Contains(line, "exec.Command(`git`") {
				violations = append(violations,
					fmt.Sprintf("%s:%d: git command execution outside internal/git package", normalizedPath, i+1))
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("Found %d violation(s) of git command execution policy:\n%s\n\n"+
			"Git commands should only be executed in internal/git package.\n"+
			"Use the git package API instead (e.g., git.FetchCommits, git.FetchFileChanges).\n"+
			"This ensures proper abstraction, testability, and centralized git operations.",
			len(violations), strings.Join(violations, "\n"))
	}
}
