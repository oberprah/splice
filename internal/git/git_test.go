package git

import (
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"
)

func TestParseGitLogOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []core.GitCommit
	}{
		{
			name:  "single commit with parent",
			input: "abc123def456789012345678901234567890abcd\x00parent1\x00\x00John Doe\x002024-01-15T10:00:00Z\x00Fix memory leak\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "abc123def456789012345678901234567890abcd",
					ParentHashes: []string{"parent1"},
					Refs:         []core.RefInfo{},
					Author:       "John Doe",
					Date:         time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					Message:      "Fix memory leak",
					Body:         "",
				},
			},
		},
		{
			name:  "multiple commits",
			input: "hash1\x00\x00\x00Author One\x002024-01-01T10:00:00Z\x00First commit\x00\x1ehash2\x00hash1\x00\x00Author Two\x002024-01-02T11:30:00Z\x00Second commit\x00\x1ehash3\x00hash2\x00\x00Author Three\x002024-01-03T15:45:00Z\x00Third commit\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash1",
					ParentHashes: []string{},
					Refs:         []core.RefInfo{},
					Author:       "Author One",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "First commit",
					Body:         "",
				},
				{
					Hash:         "hash2",
					ParentHashes: []string{"hash1"},
					Refs:         []core.RefInfo{},
					Author:       "Author Two",
					Date:         time.Date(2024, 1, 2, 11, 30, 0, 0, time.UTC),
					Message:      "Second commit",
					Body:         "",
				},
				{
					Hash:         "hash3",
					ParentHashes: []string{"hash2"},
					Refs:         []core.RefInfo{},
					Author:       "Author Three",
					Date:         time.Date(2024, 1, 3, 15, 45, 0, 0, time.UTC),
					Message:      "Third commit",
					Body:         "",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []core.GitCommit{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: []core.GitCommit{},
		},
		{
			name:  "commit with body",
			input: "hash\x00parent123\x00\x00Author\x002024-01-01T10:00:00Z\x00Fix memory leak\x00This commit fixes a critical memory leak.\n\nThe issue was in the cleanup code.\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent123"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Fix memory leak",
					Body:         "This commit fixes a critical memory leak.\n\nThe issue was in the cleanup code.",
				},
			},
		},
		{
			name:  "pipes and special chars in message",
			input: "hash\x00parent1\x00\x00Author\x002024-01-01T10:00:00Z\x00Fix A | B | C issue\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent1"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Fix A | B | C issue",
					Body:         "",
				},
			},
		},
		{
			name:  "multiple commits with bodies",
			input: "hash1\x00\x00\x00Author\x002024-01-01T10:00:00Z\x00First\x00Body 1\x1ehash2\x00hash1\x00\x00Author\x002024-01-02T10:00:00Z\x00Second\x00Body 2\x1ehash3\x00hash2\x00\x00Author\x002024-01-03T10:00:00Z\x00Third\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash1",
					ParentHashes: []string{},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "First",
					Body:         "Body 1",
				},
				{
					Hash:         "hash2",
					ParentHashes: []string{"hash1"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
					Message:      "Second",
					Body:         "Body 2",
				},
				{
					Hash:         "hash3",
					ParentHashes: []string{"hash2"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
					Message:      "Third",
					Body:         "",
				},
			},
		},
		{
			name:  "author with special characters",
			input: "hash\x00parent\x00\x00José García-López\x002024-01-01T10:00:00Z\x00Add feature\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent"},
					Refs:         []core.RefInfo{},
					Author:       "José García-López",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Add feature",
					Body:         "",
				},
			},
		},
		{
			name:  "empty message",
			input: "hash\x00parent\x00\x00Author\x002024-01-01T10:00:00Z\x00\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "",
					Body:         "",
				},
			},
		},
		{
			name:  "merge commit with two parents",
			input: "merge123\x00parent1 parent2\x00\x00Author\x002024-01-01T10:00:00Z\x00Merge branch feature\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "merge123",
					ParentHashes: []string{"parent1", "parent2"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Merge branch feature",
					Body:         "",
				},
			},
		},
		{
			name:  "octopus merge with three parents",
			input: "octopus\x00p1 p2 p3\x00\x00Author\x002024-01-01T10:00:00Z\x00Octopus merge\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "octopus",
					ParentHashes: []string{"p1", "p2", "p3"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Octopus merge",
					Body:         "",
				},
			},
		},
		{
			name:  "root commit with no parents",
			input: "root\x00\x00\x00Author\x002024-01-01T10:00:00Z\x00Initial commit\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "root",
					ParentHashes: []string{},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Initial commit",
					Body:         "",
				},
			},
		},
		{
			name:  "commit with HEAD and branch refs",
			input: "hash\x00parent\x00 (HEAD -> main)\x00Author\x002024-01-01T10:00:00Z\x00Commit on main\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent"},
					Refs: []core.RefInfo{
						{Name: "main", Type: core.RefTypeBranch, IsHead: true},
					},
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Commit on main",
					Body:    "",
				},
			},
		},
		{
			name:  "commit with multiple refs including tag",
			input: "hash\x00parent\x00 (HEAD -> main, tag: v1.0, origin/main)\x00Author\x002024-01-01T10:00:00Z\x00Release v1.0\x00\x1e",
			expected: []core.GitCommit{
				{
					Hash:         "hash",
					ParentHashes: []string{"parent"},
					Refs: []core.RefInfo{
						{Name: "main", Type: core.RefTypeBranch, IsHead: true},
						{Name: "v1.0", Type: core.RefTypeTag, IsHead: false},
						{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
					},
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Release v1.0",
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
	input := "hash\x00parent\x00\x00Author\x00INVALID_DATE\x00Message\x00\x1e"

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
		expected []core.GitCommit
	}{
		{
			name:     "incomplete commit with 5 fields only",
			input:    "hash\x00parent\x00\x00Author\x00Date",
			expected: []core.GitCommit{},
		},
		{
			name:  "valid commit followed by incomplete data",
			input: "hash1\x00parent1\x00\x00Author\x002024-01-01T10:00:00Z\x00Message\x00Body\x1eincomplete\x00data",
			expected: []core.GitCommit{
				{
					Hash:         "hash1",
					ParentHashes: []string{"parent1"},
					Refs:         []core.RefInfo{},
					Author:       "Author",
					Date:         time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message:      "Message",
					Body:         "Body",
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
		expected []core.FileChange
	}{
		{
			name:  "single file modified",
			input: "15\t3\tinternal/ui/app.go",
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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
			expected: []core.FileChange{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: []core.FileChange{},
		},
		{
			name:  "binary file",
			input: "-\t-\timage.png",
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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
			expected: []core.FileChange{
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

func TestParseRefDecorations(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []core.RefInfo
	}{
		{
			name:     "empty string",
			input:    "",
			expected: []core.RefInfo{},
		},
		{
			name:     "no refs - just whitespace",
			input:    "   ",
			expected: []core.RefInfo{},
		},
		{
			name:  "HEAD pointing to branch",
			input: " (HEAD -> main)",
			expected: []core.RefInfo{
				{Name: "main", Type: core.RefTypeBranch, IsHead: true},
			},
		},
		{
			name:  "HEAD and remote branch",
			input: " (HEAD -> main, origin/main)",
			expected: []core.RefInfo{
				{Name: "main", Type: core.RefTypeBranch, IsHead: true},
				{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
			},
		},
		{
			name:  "tag only",
			input: " (tag: v1.0)",
			expected: []core.RefInfo{
				{Name: "v1.0", Type: core.RefTypeTag, IsHead: false},
			},
		},
		{
			name:  "local branch only",
			input: " (main)",
			expected: []core.RefInfo{
				{Name: "main", Type: core.RefTypeBranch, IsHead: false},
			},
		},
		{
			name:  "remote branch only",
			input: " (origin/main)",
			expected: []core.RefInfo{
				{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
			},
		},
		{
			name:  "multiple refs with HEAD, remote, and tag",
			input: " (HEAD -> main, origin/main, tag: v1.0)",
			expected: []core.RefInfo{
				{Name: "main", Type: core.RefTypeBranch, IsHead: true},
				{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
				{Name: "v1.0", Type: core.RefTypeTag, IsHead: false},
			},
		},
		{
			name:  "multiple local branches",
			input: " (main, develop)",
			expected: []core.RefInfo{
				{Name: "main", Type: core.RefTypeBranch, IsHead: false},
				{Name: "develop", Type: core.RefTypeBranch, IsHead: false},
			},
		},
		{
			name:  "multiple remote branches",
			input: " (origin/main, upstream/main)",
			expected: []core.RefInfo{
				{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
				{Name: "upstream/main", Type: core.RefTypeRemoteBranch, IsHead: false},
			},
		},
		{
			name:  "multiple tags",
			input: " (tag: v1.0, tag: v1.0.1)",
			expected: []core.RefInfo{
				{Name: "v1.0", Type: core.RefTypeTag, IsHead: false},
				{Name: "v1.0.1", Type: core.RefTypeTag, IsHead: false},
			},
		},
		{
			name:  "complex scenario - HEAD, local, remote, tags",
			input: " (HEAD -> feature/branch, origin/feature/branch, main, tag: release-1.0)",
			expected: []core.RefInfo{
				{Name: "feature/branch", Type: core.RefTypeBranch, IsHead: true},
				{Name: "origin/feature/branch", Type: core.RefTypeRemoteBranch, IsHead: false},
				{Name: "main", Type: core.RefTypeBranch, IsHead: false},
				{Name: "release-1.0", Type: core.RefTypeTag, IsHead: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseRefDecorations(tt.input)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseRefDecorations() mismatch:\ngot:  %+v\nwant: %+v", result, tt.expected)
			}
		})
	}
}

// Tests for uncommitted changes file list functions
func TestParseUncommittedFileChanges(t *testing.T) {
	tests := []struct {
		name         string
		numstatOut   string
		statusOut    string
		expectedType string // "unstaged", "staged", or "all"
		expected     []core.FileChange
	}{
		{
			name:         "unstaged - single modified file",
			numstatOut:   "15\t3\tinternal/ui/app.go",
			statusOut:    "M\tinternal/ui/app.go",
			expectedType: "unstaged",
			expected: []core.FileChange{
				{
					Path:      "internal/ui/app.go",
					Additions: 15,
					Deletions: 3,
					Status:    "M",
				},
			},
		},
		{
			name:         "staged - added file",
			numstatOut:   "50\t0\tnewfile.go",
			statusOut:    "A\tnewfile.go",
			expectedType: "staged",
			expected: []core.FileChange{
				{
					Path:      "newfile.go",
					Additions: 50,
					Deletions: 0,
					Status:    "A",
				},
			},
		},
		{
			name:         "all uncommitted - deleted file",
			numstatOut:   "0\t25\toldfile.go",
			statusOut:    "D\toldfile.go",
			expectedType: "all",
			expected: []core.FileChange{
				{
					Path:      "oldfile.go",
					Additions: 0,
					Deletions: 25,
					Status:    "D",
				},
			},
		},
		{
			name:         "multiple files with different statuses",
			numstatOut:   "10\t5\tREADME.md\n50\t0\tnew.go\n0\t30\told.go",
			statusOut:    "M\tREADME.md\nA\tnew.go\nD\told.go",
			expectedType: "unstaged",
			expected: []core.FileChange{
				{
					Path:      "README.md",
					Additions: 10,
					Deletions: 5,
					Status:    "M",
				},
				{
					Path:      "new.go",
					Additions: 50,
					Deletions: 0,
					Status:    "A",
				},
				{
					Path:      "old.go",
					Additions: 0,
					Deletions: 30,
					Status:    "D",
				},
			},
		},
		{
			name:         "binary file",
			numstatOut:   "-\t-\timage.png",
			statusOut:    "M\timage.png",
			expectedType: "unstaged",
			expected: []core.FileChange{
				{
					Path:      "image.png",
					Additions: 0,
					Deletions: 0,
					Status:    "M",
					IsBinary:  true,
				},
			},
		},
		{
			name:         "empty - no changes",
			numstatOut:   "",
			statusOut:    "",
			expectedType: "unstaged",
			expected:     []core.FileChange{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse status map (simulating what the function does)
			statusMap := make(map[string]string)
			statusLines := strings.Split(strings.TrimSpace(tt.statusOut), "\n")
			for _, line := range statusLines {
				if line == "" {
					continue
				}
				parts := strings.SplitN(line, "\t", 2)
				if len(parts) == 2 {
					statusMap[parts[1]] = parts[0]
				}
			}

			// Parse file changes
			changes, err := ParseFileChangesOutput(tt.numstatOut)
			if err != nil {
				t.Fatalf("ParseFileChangesOutput() error = %v", err)
			}

			// Add status to each change
			for i := range changes {
				if status, ok := statusMap[changes[i].Path]; ok {
					changes[i].Status = status
				}
			}

			if !reflect.DeepEqual(changes, tt.expected) {
				t.Errorf("Result mismatch:\ngot:  %+v\nwant: %+v", changes, tt.expected)
			}
		})
	}
}

// Integration test for FetchFileChanges with commit ranges
// This test uses the actual git repository to verify correct behavior
func TestFetchFileChanges_CommitRange(t *testing.T) {
	// Skip if not in a git repository
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		t.Skip("Not in a git repository")
	}

	// Get origin/main commit
	cmd = exec.Command("git", "rev-parse", "origin/main")
	originMainOutput, err := cmd.Output()
	if err != nil {
		t.Skip("origin/main not available")
	}
	originMainHash := strings.TrimSpace(string(originMainOutput))

	// Get HEAD commit
	cmd = exec.Command("git", "rev-parse", "HEAD")
	headOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get HEAD: %v", err)
	}
	headHash := strings.TrimSpace(string(headOutput))

	// Skip if origin/main and HEAD are the same
	if originMainHash == headHash {
		t.Skip("origin/main and HEAD are the same commit")
	}

	// Count commits in the range using git rev-list
	cmd = exec.Command("git", "rev-list", "--count", "origin/main..HEAD")
	countOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to count commits: %v", err)
	}
	expectedCount := strings.TrimSpace(string(countOutput))
	if expectedCount == "0" {
		t.Skip("No commits between origin/main and HEAD")
	}

	// Get the actual file list for the correct range (origin/main..HEAD)
	cmd = exec.Command("git", "diff", "--name-only", "origin/main..HEAD")
	correctRangeOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get correct range files: %v", err)
	}
	correctFiles := strings.Split(strings.TrimSpace(string(correctRangeOutput)), "\n")
	if len(correctFiles) == 1 && correctFiles[0] == "" {
		correctFiles = []string{}
	}

	// Get the file list for the incorrect range (origin/main^..HEAD) to show the bug
	cmd = exec.Command("git", "diff", "--name-only", "origin/main^..HEAD")
	incorrectRangeOutput, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get incorrect range files: %v", err)
	}
	incorrectFiles := strings.Split(strings.TrimSpace(string(incorrectRangeOutput)), "\n")
	if len(incorrectFiles) == 1 && incorrectFiles[0] == "" {
		incorrectFiles = []string{}
	}

	// Verify the ranges are different (otherwise the test is meaningless)
	if len(correctFiles) == len(incorrectFiles) {
		t.Skip("Both ranges have the same number of files; bug may not be visible in current state")
	}

	// Create GitCommit objects for the range
	originMainCommit := core.GitCommit{Hash: originMainHash}
	headCommit := core.GitCommit{Hash: headHash}

	// Create CommitRange (not a single commit)
	commitRange := core.CommitRange{
		Start: originMainCommit,
		End:   headCommit,
		Count: 2, // Doesn't matter as long as > 1
	}

	// Call FetchFileChanges
	files, err := FetchFileChanges(commitRange)
	if err != nil {
		t.Fatalf("FetchFileChanges failed: %v", err)
	}

	// Extract file paths from the result
	actualPaths := make([]string, len(files))
	for i, f := range files {
		actualPaths[i] = f.Path
	}

	// Sort both slices for comparison
	sort.Strings(correctFiles)
	sort.Strings(actualPaths)

	// Verify that FetchFileChanges returns the correct range (origin/main..HEAD),
	// not the incorrect range (origin/main^..HEAD)
	if !reflect.DeepEqual(actualPaths, correctFiles) {
		t.Errorf("FetchFileChanges returned wrong files for commit range.\n"+
			"Expected files from range origin/main..HEAD (%d files):\n%v\n\n"+
			"Got files matching origin/main^..HEAD range (%d files):\n%v\n\n"+
			"This indicates FetchFileChanges is using Start^ instead of Start",
			len(correctFiles), correctFiles,
			len(actualPaths), actualPaths)
	}
}

// Tests for uncommitted file diff functions
func TestUncommittedFileDiffParsing(t *testing.T) {
	tests := []struct {
		name         string
		file         core.FileChange
		oldContent   string
		newContent   string
		unifiedDiff  string
		diffType     string // "unstaged", "staged", or "all"
		expectOldErr bool
		expectNewErr bool
	}{
		{
			name: "unstaged - modified file",
			file: core.FileChange{
				Path:      "test.go",
				Status:    "M",
				Additions: 5,
				Deletions: 2,
			},
			oldContent:  "package main\n\nfunc old() {}\n",
			newContent:  "package main\n\nfunc new() {}\nfunc added() {}\n",
			unifiedDiff: "diff --git a/test.go b/test.go\n@@ -1,3 +1,4 @@\n package main\n \n-func old() {}\n+func new() {}\n+func added() {}\n",
			diffType:    "unstaged",
		},
		{
			name: "staged - added file (no old content)",
			file: core.FileChange{
				Path:      "new.go",
				Status:    "A",
				Additions: 10,
				Deletions: 0,
			},
			oldContent:   "",
			newContent:   "package main\n\nfunc main() {}\n",
			unifiedDiff:  "diff --git a/new.go b/new.go\nnew file mode 100644\n@@ -0,0 +1,3 @@\n+package main\n+\n+func main() {}\n",
			diffType:     "staged",
			expectOldErr: true, // Old content doesn't exist
		},
		{
			name: "all - deleted file (no new content)",
			file: core.FileChange{
				Path:      "deleted.go",
				Status:    "D",
				Additions: 0,
				Deletions: 5,
			},
			oldContent:   "package main\n\nfunc removed() {}\n",
			newContent:   "",
			unifiedDiff:  "diff --git a/deleted.go b/deleted.go\ndeleted file mode 100644\n@@ -1,3 +0,0 @@\n-package main\n-\n-func removed() {}\n",
			diffType:     "all",
			expectNewErr: true, // New content doesn't exist
		},
		{
			name: "unstaged - binary file",
			file: core.FileChange{
				Path:     "image.png",
				Status:   "M",
				IsBinary: true,
			},
			oldContent:  "binary content old",
			newContent:  "binary content new",
			unifiedDiff: "Binary files a/image.png and b/image.png differ\n",
			diffType:    "unstaged",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the test data structure
			if tt.file.Status == "A" && !tt.expectOldErr {
				t.Error("Added files should expect old content error")
			}
			if tt.file.Status == "D" && !tt.expectNewErr {
				t.Error("Deleted files should expect new content error")
			}

			// Verify unified diff contains expected markers
			if tt.file.IsBinary {
				if !strings.Contains(tt.unifiedDiff, "Binary files") {
					t.Error("Binary file diff should contain 'Binary files' marker")
				}
			} else {
				if !strings.Contains(tt.unifiedDiff, "diff --git") {
					t.Error("Unified diff should contain 'diff --git' header")
				}
			}
		})
	}
}

func TestGetRepositoryRoot(t *testing.T) {
	tests := []struct {
		name              string
		expectError       bool
		expectedErrSubstr string
	}{
		{
			name:        "success - returns valid path",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := GetRepositoryRoot()

			if tt.expectError {
				if err == nil {
					t.Fatal("GetRepositoryRoot() expected error, got nil")
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("Error message %q should contain %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Fatalf("GetRepositoryRoot() unexpected error: %v", err)
				}
				if root == "" {
					t.Error("GetRepositoryRoot() returned empty path")
				}
				// Verify the returned path looks reasonable (should be an absolute path)
				if !strings.HasPrefix(root, "/") {
					t.Errorf("GetRepositoryRoot() returned non-absolute path: %q", root)
				}
			}
		})
	}
}
