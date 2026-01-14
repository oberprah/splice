package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/oberprah/splice/internal/core"
)

// MockContext is a test helper that implements the core.Context interface.
// Use with named fields: testutils.MockContext{W: 80, H: 24}
type MockContext struct {
	W                     int
	H                     int
	MockFetchFileChanges  core.FetchFileChangesFunc
	MockFetchFullFileDiff core.FetchFullFileDiffFunc
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
	return func(commitRange core.CommitRange) ([]core.FileChange, error) {
		return []core.FileChange{}, nil
	}
}

func (m MockContext) FetchFullFileDiff() core.FetchFullFileDiffFunc {
	if m.MockFetchFullFileDiff != nil {
		return m.MockFetchFullFileDiff
	}
	return func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
		return &core.FullFileDiffResult{}, nil
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

// AssertGolden compares the output against a golden file.
// If update is true, it updates the golden file instead.
// The goldenPath should be relative to the caller's testdata directory.
// ANSI escape codes are stripped from output before comparison/storage.
func AssertGolden(t *testing.T, output, goldenPath string, update bool) {
	t.Helper()

	// Strip ANSI codes from output for clean, readable golden files
	strippedOutput := stripAnsiCodes(output)

	if update {
		dir := filepath.Dir(goldenPath)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		err = os.WriteFile(goldenPath, []byte(strippedOutput), 0644)
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

	if string(expected) != strippedOutput {
		t.Errorf("Output does not match golden file %s.\nRun with -update to update golden files.\n\nExpected:\n%s\n\nGot:\n%s",
			goldenPath, string(expected), strippedOutput)
	}
}

// CreateTestCommits generates n mock git commits for testing
// Uses fixed dates that are exactly 1 year old to ensure deterministic formatting
// Creates linear commit history in display order (newest to oldest)
// Each commit has the next commit in the array as its parent (linear history)
func CreateTestCommits(count int) []core.GitCommit {
	commits := make([]core.GitCommit, count)
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

		commits[i] = core.GitCommit{
			Hash:         fmt.Sprintf("%040d", i), // Full 40-char hash
			ParentHashes: parentHashes,
			Refs:         []core.RefInfo{}, // No refs by default
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
func CreateTestCommitsWithMessages(messages []string) []core.GitCommit {
	commits := make([]core.GitCommit, len(messages))
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

		commits[i] = core.GitCommit{
			Hash:         fmt.Sprintf("%040d", i),
			ParentHashes: parentHashes,
			Refs:         []core.RefInfo{}, // No refs by default
			Message:      msg,
			Body:         "",
			Author:       "Test Author",
			Date:         baseTime.Add(time.Duration(-i) * time.Hour),
		}
	}

	return commits
}

// MockFetchCommits creates a mock function that returns test commits
func MockFetchCommits(commits []core.GitCommit, err error) func(int) ([]core.GitCommit, error) {
	return func(limit int) ([]core.GitCommit, error) {
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
func MockFetchFileChanges(files []core.FileChange, err error) func(core.CommitRange) ([]core.FileChange, error) {
	return func(commitRange core.CommitRange) ([]core.FileChange, error) {
		if err != nil {
			return nil, err
		}
		return files, nil
	}
}

// MockFetchFullFileDiff creates a mock function that returns full file diff result
func MockFetchFullFileDiff(result *core.FullFileDiffResult, err error) func(core.CommitRange, core.FileChange) (*core.FullFileDiffResult, error) {
	return func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}
