package operations

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// FetchFileChanges executes git diff and returns a slice of file changes for a commit range.
func FetchFileChanges(commitRange core.CommitRange) ([]core.FileChange, error) {
	// Determine the from and to hashes based on whether this is a single commit or range
	var fromHash, toHash string
	if commitRange.IsSingleCommit() {
		// Single commit: compare commit with its parent
		fromHash = commitRange.End.Hash + "^"
		toHash = commitRange.End.Hash
	} else {
		// Range: compare Start commit with End commit
		fromHash = commitRange.Start.Hash
		toHash = commitRange.End.Hash
	}

	rangeSpec := fromHash + ".." + toHash
	return commands.FetchFileChangesForRange(rangeSpec)
}

// FetchUnstagedFileChanges executes git diff and returns a slice of file changes
// for unstaged changes (working tree vs index).
func FetchUnstagedFileChanges() ([]core.FileChange, error) {
	return commands.FetchFileChangesWithFlags("diff")
}

// FetchStagedFileChanges executes git diff --staged and returns a slice of file changes
// for staged changes (index vs HEAD).
func FetchStagedFileChanges() ([]core.FileChange, error) {
	return commands.FetchFileChangesWithFlags("diff", "--staged")
}

// FetchAllUncommittedFileChanges executes git diff HEAD and returns a slice of file changes
// for all uncommitted changes (working tree vs HEAD).
func FetchAllUncommittedFileChanges() ([]core.FileChange, error) {
	return commands.FetchFileChangesWithFlags("diff", "HEAD")
}
