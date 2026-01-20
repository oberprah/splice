package commands

import (
	"testing"
)

// TestFetchFileChangesForRange_ParentBehavior documents the correct behavior
// for multi-commit ranges.
//
// When fetching file changes for a range of commits, we want to see the
// COMBINED changes from all commits in the range, not just the newest commit.
//
// Git diff behavior:
// - "A..B" shows changes in B relative to A (only files changed in B if A is parent of B)
// - "A^..B" shows changes from parent of A to B (combined changes if A and B are consecutive)
//
// For a multi-commit range, we must use the parent of the start commit.
func TestFetchFileChangesForRange_ParentBehavior(t *testing.T) {
	// Skip if not in a git repository
	_, _, err := execGit("rev-parse", "--git-dir")
	if err != nil {
		t.Skip("Not in a git repository")
	}

	// Use two consecutive commits: 951443e is parent of 1da9976
	olderCommit := "951443e"
	newerCommit := "1da9976"

	// Test 1: Without parent (current buggy behavior)
	// This only shows files from the newer commit
	filesWithoutParent, err := FetchFileChangesForRange(olderCommit + ".." + newerCommit)
	if err != nil {
		t.Fatalf("FetchFileChangesForRange(%s..%s) failed: %v", olderCommit, newerCommit, err)
	}

	// Test 2: With parent (correct behavior)
	// This shows combined files from both commits
	filesWithParent, err := FetchFileChangesForRange(olderCommit + "^.." + newerCommit)
	if err != nil {
		t.Fatalf("FetchFileChangesForRange(%s^..%s) failed: %v", olderCommit, newerCommit, err)
	}

	// The version with parent should return MORE files
	if len(filesWithParent) <= len(filesWithoutParent) {
		t.Errorf("Range with parent returned %d files, expected MORE than without parent's %d files",
			len(filesWithParent), len(filesWithoutParent))
		t.Logf("Without parent (%s..%s): %d files", olderCommit, newerCommit, len(filesWithoutParent))
		t.Logf("With parent (%s^..%s): %d files", olderCommit, newerCommit, len(filesWithParent))
	}

	// Based on manual verification of this repo's history
	expectedWithoutParent := 19 // Only files from 1da9976
	expectedWithParent := 21    // Combined files from both 951443e and 1da9976

	if len(filesWithoutParent) != expectedWithoutParent {
		t.Errorf("Without parent: got %d files, expected %d", len(filesWithoutParent), expectedWithoutParent)
	}

	if len(filesWithParent) < expectedWithParent {
		t.Errorf("With parent: got %d files, expected at least %d", len(filesWithParent), expectedWithParent)
	}
}
