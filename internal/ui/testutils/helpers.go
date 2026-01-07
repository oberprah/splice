package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

// MockContext is a test helper that implements the core.Context interface.
// Use with named fields: testutils.MockContext{W: 80, H: 24}
type MockContext struct {
	W                    int
	H                    int
	MockFetchFileChanges core.FetchFileChangesFunc
}

func (m MockContext) Width() int {
	return m.W
}

func (m MockContext) Height() int {
	return m.H
}

func (m MockContext) FetchFileChanges() core.FetchFileChangesFunc {
	if m.MockFetchFileChanges != nil {
		return m.MockFetchFileChanges
	}
	return func(fromHash, toHash string) ([]git.FileChange, error) {
		return []git.FileChange{}, nil
	}
}

func (m MockContext) FetchFullFileDiff() core.FetchFullFileDiffFunc {
	return func(fromHash, toHash string, change git.FileChange) (*git.FullFileDiffResult, error) {
		return &git.FullFileDiffResult{}, nil
	}
}

func (m MockContext) Now() time.Time {
	// Return fixed time for deterministic tests (commits are exactly 1 year old)
	return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
}

// SetupColorProfile enables TrueColor for deterministic test output.
// Call at the start of individual tests that need consistent color rendering.
func SetupColorProfile() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

// AssertGolden compares the output against a golden file.
// If update is true, it updates the golden file instead.
// The goldenPath should be relative to the caller's testdata directory.
func AssertGolden(t *testing.T, output, goldenPath string, update bool) {
	t.Helper()

	if update {
		dir := filepath.Dir(goldenPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		err = os.WriteFile(goldenPath, []byte(output), 0644)
		if err != nil {
			t.Fatalf("Failed to write golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v\nRun with -update to create it", err)
	}

	if string(expected) != output {
		t.Errorf("Output does not match golden file %s.\nRun with -update to update golden files.\n\nExpected:\n%s\n\nGot:\n%s",
			goldenPath, string(expected), output)
	}
}

// CreateTestCommits generates n mock git commits for testing
// Uses fixed dates that are exactly 1 year old to ensure deterministic formatting
// Creates linear commit history in display order (newest to oldest)
// Each commit has the next commit in the array as its parent (linear history)
func CreateTestCommits(count int) []git.GitCommit {
	commits := make([]git.GitCommit, count)
	// Fixed date exactly 1 year ago from test "now" (2024-01-01 from 2025-01-01)
	// This ensures "1 year ago" formatting consistently in tests
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for i := range count {
		var parentHashes []string
		if i < count-1 {
			// Each commit has the next commit in array as its parent (older commit)
			// Array is in display order: newest first, so parent is i+1
			parentHashes = []string{fmt.Sprintf("%040d", i+1)}
		} else {
			// Last commit in array (oldest) has no parents
			parentHashes = []string{}
		}

		commits[i] = git.GitCommit{
			Hash:         fmt.Sprintf("%040d", i), // Full 40-char hash
			ParentHashes: parentHashes,
			Refs:         []git.RefInfo{}, // No refs by default
			Message:      fmt.Sprintf("Commit message %d", i),
			Body:         "",
			Author:       fmt.Sprintf("Author %d", i%3),               // Vary authors
			Date:         baseTime.Add(time.Duration(-i) * time.Hour), // Reverse chronological
		}
	}

	return commits
}

// CreateTestCommitsWithMessages generates commits with specific messages
// Uses fixed dates that are exactly 1 year old to ensure deterministic formatting
// Creates linear commit history in display order (newest to oldest)
// Each commit has the next commit in the array as its parent (linear history)
func CreateTestCommitsWithMessages(messages []string) []git.GitCommit {
	commits := make([]git.GitCommit, len(messages))
	// Fixed date exactly 1 year ago from test "now" (2024-01-01 from 2025-01-01)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for i, msg := range messages {
		var parentHashes []string
		if i < len(messages)-1 {
			// Each commit has the next commit in array as its parent (older commit)
			// Array is in display order: newest first, so parent is i+1
			parentHashes = []string{fmt.Sprintf("%040d", i+1)}
		} else {
			// Last commit in array (oldest) has no parents
			parentHashes = []string{}
		}

		commits[i] = git.GitCommit{
			Hash:         fmt.Sprintf("%040d", i),
			ParentHashes: parentHashes,
			Refs:         []git.RefInfo{}, // No refs by default
			Message:      msg,
			Body:         "",
			Author:       "Test Author",
			Date:         baseTime.Add(time.Duration(-i) * time.Hour),
		}
	}

	return commits
}

// MockFetchCommits creates a mock function that returns test commits
func MockFetchCommits(commits []git.GitCommit, err error) func(int) ([]git.GitCommit, error) {
	return func(limit int) ([]git.GitCommit, error) {
		if err != nil {
			return nil, err
		}
		if limit < len(commits) {
			return commits[:limit], nil
		}
		return commits, nil
	}
}

// MockFetchFileChanges creates a mock function that returns file changes for a commit range
func MockFetchFileChanges(files []git.FileChange, err error) func(string, string) ([]git.FileChange, error) {
	return func(fromHash, toHash string) ([]git.FileChange, error) {
		if err != nil {
			return nil, err
		}
		return files, nil
	}
}

// MockFetchFullFileDiff creates a mock function that returns full file diff result
func MockFetchFullFileDiff(result *git.FullFileDiffResult, err error) func(string, string, git.FileChange) (*git.FullFileDiffResult, error) {
	return func(fromHash, toHash string, change git.FileChange) (*git.FullFileDiffResult, error) {
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}
