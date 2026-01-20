package operations

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

// TestFetchFileChanges_MultiCommitRange tests that FetchFileChanges correctly
// returns combined file changes for a multi-commit range, not just the files
// from the newest commit.
//
// This is a regression test for the bug where selecting multiple commits in
// visual mode showed only files from the newest commit instead of combined files.
func TestFetchFileChanges_MultiCommitRange(t *testing.T) {
	// This test requires being run in the splice git repository
	// If not in a git repository, FetchFileChanges will fail and the test will fail

	// Use two consecutive commits from this repository's history
	// Commit 951443e is the parent of commit 1da9976
	newerCommit := core.GitCommit{Hash: "1da9976"}
	olderCommit := core.GitCommit{Hash: "951443e"}

	// Test 1: Get files for newer commit alone (single commit)
	singleCommitRange := core.NewCommitRange(newerCommit, newerCommit, 1)
	singleFiles, err := FetchFileChanges(singleCommitRange)
	if err != nil {
		t.Fatalf("FetchFileChanges for single commit failed: %v", err)
	}

	// Test 2: Get files for the multi-commit range (both commits)
	multiCommitRange := core.NewCommitRange(olderCommit, newerCommit, 2)
	multiFiles, err := FetchFileChanges(multiCommitRange)
	if err != nil {
		t.Fatalf("FetchFileChanges for range failed: %v", err)
	}

	// The multi-commit range should have MORE files than the single commit
	// because it includes changes from both commits
	if len(multiFiles) <= len(singleFiles) {
		t.Errorf("Multi-commit range returned %d files, expected MORE than single commit's %d files",
			len(multiFiles), len(singleFiles))
		t.Errorf("Single commit files: %v", fileNames(singleFiles))
		t.Errorf("Multi-commit files: %v", fileNames(multiFiles))
	}

	// Verify the multi-commit range actually has the expected number of files
	// Based on manual verification: git diff 951443e^..1da9976 returns 21 files
	expectedMinFiles := 20 // Allow some tolerance
	if len(multiFiles) < expectedMinFiles {
		t.Errorf("Multi-commit range returned %d files, expected at least %d files",
			len(multiFiles), expectedMinFiles)
	}
}

// fileNames extracts file paths from file changes for easier debugging
func fileNames(files []core.FileChange) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Path
	}
	return names
}
