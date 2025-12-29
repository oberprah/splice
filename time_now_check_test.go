package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoDirectTimeNowUsage ensures time.Now() is not used directly in production code.
// All date formatting should use ctx.Now() for testability.
func TestNoDirectTimeNowUsage(t *testing.T) {
	// Allowed files where time.Now() is acceptable
	allowedFiles := map[string]bool{
		"e2e/helpers_test.go":    true, // Test timeout/deadline logic
		"internal/ui/app.go":     true, // Default initialization: nowFunc: time.Now
		"time_now_check_test.go": true, // This test file
	}

	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-Go files, vendor, and testdata
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.Contains(path, "vendor/") || strings.Contains(path, "testdata/") {
			return nil
		}

		// Normalize path separators for cross-platform compatibility
		normalizedPath := filepath.ToSlash(path)

		// Skip allowed files
		if allowedFiles[normalizedPath] {
			return nil
		}

		// Read file contents
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Check for time.Now() usage
		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			// Skip comments
			if strings.Contains(line, "//") {
				commentIdx := strings.Index(line, "//")
				line = line[:commentIdx]
			}

			if strings.Contains(line, "time.Now()") {
				violations = append(violations,
					fmt.Sprintf("%s:%d: contains time.Now() - use ctx.Now() instead", normalizedPath, i+1))
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if len(violations) > 0 {
		t.Errorf("Found %d violation(s) of time.Now() usage:\n%s\n\n"+
			"All date formatting should use ctx.Now() for deterministic testing.\n"+
			"If this is intentional (e.g., timeout logic), add the file to allowedFiles in time_now_check_test.go",
			len(violations), strings.Join(violations, "\n"))
	}
}
