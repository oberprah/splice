package git

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/operations"
)

// FetchCommits executes git log and returns a slice of commits
func FetchCommits(limit int) ([]core.GitCommit, error) {
	return operations.FetchCommits(limit)
}

// FetchFileChanges executes git diff and returns a slice of file changes for a commit range.
func FetchFileChanges(commitRange core.CommitRange) ([]core.FileChange, error) {
	return operations.FetchFileChanges(commitRange)
}

// FetchFileChangesForSource fetches file changes based on DiffSource.
// This is the unified interface for getting file changes from any source.
func FetchFileChangesForSource(source core.DiffSource) ([]core.FileChange, error) {
	switch src := source.(type) {
	case core.CommitRangeDiffSource:
		return operations.FetchFileChanges(src.ToCommitRange())
	case core.UncommittedChangesDiffSource:
		return operations.FetchUncommittedFileChanges(src.Type)
	default:
		return nil, fmt.Errorf("unknown diff source type: %T", source)
	}
}

// FetchFullFileDiff fetches the complete file content before and after a change,
// along with the diff output. This enables showing the full file with changes highlighted.
func FetchFullFileDiff(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
	return operations.FetchFullFileDiff(commitRange, change)
}

// FetchFullFileDiffForSource fetches the full file diff based on DiffSource.
// For commit ranges, it delegates to the injected fetchFullFileDiff.
// For uncommitted changes, it uses the git package helpers.
func FetchFullFileDiffForSource(
	source core.DiffSource,
	change core.FileChange,
	fetchFullFileDiff core.FetchFullFileDiffFunc,
) (*core.FullFileDiffResult, error) {
	switch src := source.(type) {
	case core.CommitRangeDiffSource:
		commitRange := src.ToCommitRange()
		return fetchFullFileDiff(commitRange, change)

	case core.UncommittedChangesDiffSource:
		return operations.FetchUncommittedFileDiff(src.Type, change)

	default:
		return nil, fmt.Errorf("unknown diff source type: %T", source)
	}
}

// ValidateDiffHasChanges checks if a diff specification has any changes.
// For uncommitted changes, checks the appropriate git diff.
// For commit ranges, checks if the range has any diff.
// Returns nil if there are changes, error if no changes or invalid spec.
func ValidateDiffHasChanges(rawSpec string, uncommittedType *core.UncommittedType) error {
	return operations.ValidateDiffHasChanges(rawSpec, uncommittedType)
}

// ResolveCommitRange parses a commit range spec (like "main..feature" or "HEAD~5")
// and resolves refs to GitCommit objects.
func ResolveCommitRange(spec string) (core.CommitRangeDiffSource, error) {
	return operations.ResolveCommitRange(spec)
}

// GetRepositoryRoot executes git rev-parse --show-toplevel to get the absolute path of the repository root.
func GetRepositoryRoot() (string, error) {
	return operations.GetRepositoryRoot()
}
