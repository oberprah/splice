package commands

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

// TestFetchFileChangesWithFlags tests the base function with various flag combinations.
// This is an integration test that requires a real git repository.
func TestFetchFileChangesWithFlags(t *testing.T) {
	// Skip if not in a git repository
	_, _, err := execGit("rev-parse", "--git-dir")
	if err != nil {
		t.Skip("Not in a git repository")
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "unstaged changes (may be empty)",
			args:        []string{"diff"},
			expectError: false,
		},
		{
			name:        "staged changes (may be empty)",
			args:        []string{"diff", "--staged"},
			expectError: false,
		},
		{
			name:        "all uncommitted changes (may be empty)",
			args:        []string{"diff", "HEAD"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := FetchFileChangesWithFlags(tt.args...)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Result may be empty (no changes), which is valid
			if changes == nil {
				t.Fatal("expected non-nil slice, got nil")
			}

			// Verify structure of returned changes
			for i, change := range changes {
				if change.Path == "" {
					t.Errorf("change %d has empty path", i)
				}
				// Status may be empty if not set, but that's ok for this test
			}
		})
	}
}

// TestParseStatusMap tests status parsing logic
func TestParseStatusMap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single modified file",
			input: "M\tREADME.md",
			expected: map[string]string{
				"README.md": "M",
			},
		},
		{
			name:  "multiple files with different statuses",
			input: "M\tREADME.md\nA\tnew.go\nD\told.go",
			expected: map[string]string{
				"README.md": "M",
				"new.go":    "A",
				"old.go":    "D",
			},
		},
		{
			name:  "file with spaces in path",
			input: "M\tpath with spaces/file.go",
			expected: map[string]string{
				"path with spaces/file.go": "M",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStatusMap(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}

			for path, expectedStatus := range tt.expected {
				actualStatus, ok := result[path]
				if !ok {
					t.Errorf("missing entry for path %q", path)
					continue
				}
				if actualStatus != expectedStatus {
					t.Errorf("path %q: expected status %q, got %q", path, expectedStatus, actualStatus)
				}
			}
		})
	}
}

// TestAddStatusToChanges tests adding status information to file changes
func TestAddStatusToChanges(t *testing.T) {
	statusMap := map[string]string{
		"file1.go": "M",
		"file2.go": "A",
		"file3.go": "D",
	}

	changes := []core.FileChange{
		{Path: "file1.go", Additions: 10, Deletions: 5},
		{Path: "file2.go", Additions: 20, Deletions: 0},
		{Path: "file3.go", Additions: 0, Deletions: 15},
		{Path: "file4.go", Additions: 5, Deletions: 2}, // No status in map
	}

	result := addStatusToChanges(changes, statusMap)

	// Check that statuses were added correctly
	if result[0].Status != "M" {
		t.Errorf("file1.go: expected status M, got %q", result[0].Status)
	}
	if result[1].Status != "A" {
		t.Errorf("file2.go: expected status A, got %q", result[1].Status)
	}
	if result[2].Status != "D" {
		t.Errorf("file3.go: expected status D, got %q", result[2].Status)
	}
	if result[3].Status != "M" {
		t.Errorf("file4.go: expected default status M, got %q", result[3].Status)
	}
}
