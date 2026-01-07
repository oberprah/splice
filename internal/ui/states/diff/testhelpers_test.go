package diff

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

var update = flag.Bool("update", false, "update golden files")

// mockContext is a test helper that implements the Context interface
type mockContext struct {
	width  int
	height int
}

func (m mockContext) Width() int {
	return m.width
}

func (m mockContext) Height() int {
	return m.height
}

func (m mockContext) FetchFileChanges() core.FetchFileChangesFunc {
	// Return a mock function that returns empty file changes
	return func(commitHash string) ([]git.FileChange, error) {
		return []git.FileChange{}, nil
	}
}

func (m mockContext) FetchFullFileDiff() core.FetchFullFileDiffFunc {
	// Return a mock function that returns an empty diff result
	return func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error) {
		return &git.FullFileDiffResult{}, nil
	}
}

func (m mockContext) Now() time.Time {
	// Return fixed time for deterministic tests (commits are exactly 1 year old)
	return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
}

// assertGolden compares the output against a golden file.
// If the -update flag is set, it updates the golden file instead.
// The goldenFileName should be relative to the testdata directory and can include subdirectories
// (e.g., "log_view/renders_commits.golden")
func assertGolden(t *testing.T, output, goldenFileName string, updateFlag bool) {
	t.Helper()

	goldenPath := filepath.Join("testdata", goldenFileName)

	if updateFlag {
		// Create parent directories if they don't exist
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
