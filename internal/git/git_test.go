package git

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseGitLogOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []GitCommit
	}{
		{
			name:  "single commit",
			input: "abc123def456789012345678901234567890abcd\x00John Doe\x002024-01-15T10:00:00Z\x00Fix memory leak\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "abc123def456789012345678901234567890abcd",
					Author:  "John Doe",
					Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					Message: "Fix memory leak",
					Body:    "",
				},
			},
		},
		{
			name:  "multiple commits",
			input: "hash1\x00Author One\x002024-01-01T10:00:00Z\x00First commit\x00\x1ehash2\x00Author Two\x002024-01-02T11:30:00Z\x00Second commit\x00\x1ehash3\x00Author Three\x002024-01-03T15:45:00Z\x00Third commit\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash1",
					Author:  "Author One",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "First commit",
					Body:    "",
				},
				{
					Hash:    "hash2",
					Author:  "Author Two",
					Date:    time.Date(2024, 1, 2, 11, 30, 0, 0, time.UTC),
					Message: "Second commit",
					Body:    "",
				},
				{
					Hash:    "hash3",
					Author:  "Author Three",
					Date:    time.Date(2024, 1, 3, 15, 45, 0, 0, time.UTC),
					Message: "Third commit",
					Body:    "",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []GitCommit{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: []GitCommit{},
		},
		{
			name:  "commit with body",
			input: "hash\x00Author\x002024-01-01T10:00:00Z\x00Fix memory leak\x00This commit fixes a critical memory leak.\n\nThe issue was in the cleanup code.\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Fix memory leak",
					Body:    "This commit fixes a critical memory leak.\n\nThe issue was in the cleanup code.",
				},
			},
		},
		{
			name:  "pipes and special chars in message",
			input: "hash\x00Author\x002024-01-01T10:00:00Z\x00Fix A | B | C issue\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Fix A | B | C issue",
					Body:    "",
				},
			},
		},
		{
			name:  "multiple commits with bodies",
			input: "hash1\x00Author\x002024-01-01T10:00:00Z\x00First\x00Body 1\x1ehash2\x00Author\x002024-01-02T10:00:00Z\x00Second\x00Body 2\x1ehash3\x00Author\x002024-01-03T10:00:00Z\x00Third\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash1",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "First",
					Body:    "Body 1",
				},
				{
					Hash:    "hash2",
					Author:  "Author",
					Date:    time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
					Message: "Second",
					Body:    "Body 2",
				},
				{
					Hash:    "hash3",
					Author:  "Author",
					Date:    time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
					Message: "Third",
					Body:    "",
				},
			},
		},
		{
			name:  "author with special characters",
			input: "hash\x00José García-López\x002024-01-01T10:00:00Z\x00Add feature\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "José García-López",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Add feature",
					Body:    "",
				},
			},
		},
		{
			name:  "empty message",
			input: "hash\x00Author\x002024-01-01T10:00:00Z\x00\x00\x1e",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "",
					Body:    "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits, err := ParseGitLogOutput(tt.input)

			if err != nil {
				t.Fatalf("ParseGitLogOutput() error = %v", err)
			}

			if !reflect.DeepEqual(commits, tt.expected) {
				t.Errorf("ParseGitLogOutput() mismatch:\ngot:  %+v\nwant: %+v", commits, tt.expected)
			}
		})
	}
}

func TestParseGitLogOutput_InvalidDate(t *testing.T) {
	input := "hash\x00Author\x00INVALID_DATE\x00Message\x00\x1e"

	commits, err := ParseGitLogOutput(input)

	if err == nil {
		t.Fatal("ParseGitLogOutput() expected error for invalid date, got nil")
	}

	if commits != nil {
		t.Errorf("ParseGitLogOutput() expected nil commits on error, got %d commits", len(commits))
	}

	// Verify error message contains the invalid date
	expectedSubstring := "INVALID_DATE"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message %q should contain %q", err.Error(), expectedSubstring)
	}
}

func TestParseGitLogOutput_IncompleteData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []GitCommit
	}{
		{
			name:     "incomplete commit with 3 fields only",
			input:    "hash\x00Author\x00Date",
			expected: []GitCommit{},
		},
		{
			name:  "valid commit followed by incomplete data",
			input: "hash1\x00Author\x002024-01-01T10:00:00Z\x00Message\x00Body\x1eincomplete\x00data",
			expected: []GitCommit{
				{
					Hash:    "hash1",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Message",
					Body:    "Body",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits, err := ParseGitLogOutput(tt.input)

			if err != nil {
				t.Fatalf("ParseGitLogOutput() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(commits, tt.expected) {
				t.Errorf("ParseGitLogOutput() mismatch:\ngot:  %+v\nwant: %+v", commits, tt.expected)
			}
		})
	}
}

func TestParseFileChangesOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []FileChange
	}{
		{
			name:  "single file modified",
			input: "15\t3\tinternal/ui/app.go",
			expected: []FileChange{
				{
					Path:      "internal/ui/app.go",
					Additions: 15,
					Deletions: 3,
				},
			},
		},
		{
			name:  "multiple files",
			input: "45\t12\tinternal/ui/app.go\n3\t1\tinternal/ui/model.go\n120\t0\tinternal/git/git.go",
			expected: []FileChange{
				{
					Path:      "internal/ui/app.go",
					Additions: 45,
					Deletions: 12,
				},
				{
					Path:      "internal/ui/model.go",
					Additions: 3,
					Deletions: 1,
				},
				{
					Path:      "internal/git/git.go",
					Additions: 120,
					Deletions: 0,
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []FileChange{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: []FileChange{},
		},
		{
			name:  "binary file",
			input: "-\t-\timage.png",
			expected: []FileChange{
				{
					Path:      "image.png",
					Additions: 0,
					Deletions: 0,
					IsBinary:  true,
				},
			},
		},
		{
			name:  "mixed binary and text files",
			input: "10\t5\tREADME.md\n-\t-\tlogo.png\n3\t2\tmain.go",
			expected: []FileChange{
				{
					Path:      "README.md",
					Additions: 10,
					Deletions: 5,
				},
				{
					Path:      "logo.png",
					Additions: 0,
					Deletions: 0,
					IsBinary:  true,
				},
				{
					Path:      "main.go",
					Additions: 3,
					Deletions: 2,
				},
			},
		},
		{
			name:  "new file",
			input: "50\t0\tnewfile.go",
			expected: []FileChange{
				{
					Path:      "newfile.go",
					Additions: 50,
					Deletions: 0,
				},
			},
		},
		{
			name:  "deleted file",
			input: "0\t25\tdeletedfile.go",
			expected: []FileChange{
				{
					Path:      "deletedfile.go",
					Additions: 0,
					Deletions: 25,
				},
			},
		},
		{
			name:  "file with spaces in path",
			input: "5\t2\tpath with spaces/file.go",
			expected: []FileChange{
				{
					Path:      "path with spaces/file.go",
					Additions: 5,
					Deletions: 2,
				},
			},
		},
		{
			name:  "empty lines between files ignored",
			input: "10\t5\tfile1.go\n\n20\t3\tfile2.go\n\n\n15\t8\tfile3.go",
			expected: []FileChange{
				{
					Path:      "file1.go",
					Additions: 10,
					Deletions: 5,
				},
				{
					Path:      "file2.go",
					Additions: 20,
					Deletions: 3,
				},
				{
					Path:      "file3.go",
					Additions: 15,
					Deletions: 8,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := ParseFileChangesOutput(tt.input)

			if err != nil {
				t.Fatalf("ParseFileChangesOutput() error = %v", err)
			}

			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("ParseFileChangesOutput() mismatch:\ngot:  %+v\nwant: %+v", changes, tt.expected)
			}
		})
	}
}

func TestParseFileChangesOutput_InvalidInput(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedErrSubstr string
	}{
		{
			name:              "line with only 2 fields",
			input:             "15\t3",
			expectedErrSubstr: "malformed line",
		},
		{
			name:              "line with only 1 field",
			input:             "singlefield",
			expectedErrSubstr: "malformed line",
		},
		{
			name:              "invalid additions number",
			input:             "abc\t3\tfile.go",
			expectedErrSubstr: "invalid additions",
		},
		{
			name:              "invalid deletions number",
			input:             "15\txyz\tfile.go",
			expectedErrSubstr: "invalid deletions",
		},
		{
			name:              "malformed line after valid file",
			input:             "15\t3\tfile1.go\nMALFORMED_LINE",
			expectedErrSubstr: "malformed line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := ParseFileChangesOutput(tt.input)

			if err == nil {
				t.Fatal("ParseFileChangesOutput() expected error for invalid input, got nil")
			}

			if changes != nil {
				t.Errorf("ParseFileChangesOutput() expected nil changes on error, got %d changes", len(changes))
			}

			if !strings.Contains(err.Error(), tt.expectedErrSubstr) {
				t.Errorf("Error message %q should contain %q", err.Error(), tt.expectedErrSubstr)
			}
		})
	}
}

func TestFetchFileChanges_Integration(t *testing.T) {
	// This test requires a git repository with at least one commit
	// Skip if not in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Get the latest commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD commit: %v", err)
	}
	commitHash := strings.TrimSpace(string(out))

	// Fetch file changes for the commit
	changes, err := FetchFileChanges(commitHash)
	if err != nil {
		t.Fatalf("FetchFileChanges() error = %v", err)
	}

	// Basic validation - we should have at least some structure
	// (We can't make strong assertions about content since it depends on the actual commit)
	if changes == nil {
		t.Error("FetchFileChanges() returned nil changes")
	}

	// Each file change should have a valid path
	for i, change := range changes {
		if change.Path == "" {
			t.Errorf("Change %d has empty path", i)
		}
		if change.Additions < 0 {
			t.Errorf("Change %d (%s) has negative additions: %d", i, change.Path, change.Additions)
		}
		if change.Deletions < 0 {
			t.Errorf("Change %d (%s) has negative deletions: %d", i, change.Path, change.Deletions)
		}
		// Validate Status field is populated
		if change.Status == "" {
			t.Errorf("Change %d (%s) has empty Status field", i, change.Path)
		}
		// Status should be one of the valid git status codes
		validStatuses := []string{"M", "A", "D", "R", "C", "T", "U", "X"}
		validStatus := false
		for _, valid := range validStatuses {
			if strings.HasPrefix(change.Status, valid) {
				validStatus = true
				break
			}
		}
		if !validStatus {
			t.Errorf("Change %d (%s) has invalid Status: %q (expected one of M/A/D/R/C/T/U/X)",
				i, change.Path, change.Status)
		}
	}
}

func TestFetchFileChanges_InvalidCommit(t *testing.T) {
	// This test requires a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Try to fetch file changes for an invalid commit hash
	changes, err := FetchFileChanges("invalid_commit_hash_12345")

	if err == nil {
		t.Fatal("FetchFileChanges() expected error for invalid commit, got nil")
	}

	if changes != nil {
		t.Errorf("FetchFileChanges() expected nil changes on error, got %d changes", len(changes))
	}
}

func TestFetchFileDiff_Integration(t *testing.T) {
	// This test requires a git repository with at least one commit
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	// Get the latest commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD commit: %v", err)
	}
	commitHash := strings.TrimSpace(string(out))

	// Get file changes for this commit to find a file to test
	changes, err := FetchFileChanges(commitHash)
	if err != nil {
		t.Fatalf("FetchFileChanges() error = %v", err)
	}

	if len(changes) == 0 {
		t.Skip("No file changes in HEAD commit, skipping test")
	}

	// Fetch diff for the first file
	filePath := changes[0].Path
	diff, err := FetchFileDiff(commitHash, filePath)
	if err != nil {
		t.Fatalf("FetchFileDiff() error = %v", err)
	}

	// Basic validation - diff should contain the file path
	if !strings.Contains(diff, filePath) {
		t.Errorf("FetchFileDiff() output should contain file path %q", filePath)
	}

	// Diff should contain diff header markers
	if !strings.Contains(diff, "diff --git") {
		t.Error("FetchFileDiff() output should contain 'diff --git' header")
	}
}

func TestFetchFileDiff_InvalidCommit(t *testing.T) {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("Not in a git repository, skipping integration test")
	}

	diff, err := FetchFileDiff("invalid_commit_hash_12345", "some/file.go")

	if err == nil {
		t.Fatal("FetchFileDiff() expected error for invalid commit, got nil")
	}

	if diff != "" {
		t.Errorf("FetchFileDiff() expected empty diff on error, got: %s", diff)
	}
}
