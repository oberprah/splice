package operations

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// FetchUncommittedFileChanges fetches file changes for any uncommitted change type.
// This is the unified interface that dispatches based on type.
func FetchUncommittedFileChanges(uncommittedType core.UncommittedType) ([]core.FileChange, error) {
	switch uncommittedType {
	case core.UncommittedTypeUnstaged:
		return commands.FetchFileChangesWithFlags("diff")
	case core.UncommittedTypeStaged:
		return commands.FetchFileChangesWithFlags("diff", "--staged")
	case core.UncommittedTypeAll:
		return commands.FetchFileChangesWithFlags("diff", "HEAD")
	default:
		return nil, fmt.Errorf("unknown uncommitted type: %v", uncommittedType)
	}
}

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
