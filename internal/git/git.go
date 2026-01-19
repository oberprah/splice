package git

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/operations"
	"github.com/oberprah/splice/internal/git/parsing"
)

// FetchCommits executes git log and returns a slice of commits
func FetchCommits(limit int) ([]core.GitCommit, error) {
	return operations.FetchCommits(limit)
}

// FetchFileChanges executes git diff and returns a slice of file changes for a commit range.
func FetchFileChanges(commitRange core.CommitRange) ([]core.FileChange, error) {
	return operations.FetchFileChanges(commitRange)
}

// FetchFileContent retrieves the content of a file at a specific commit.
// Returns empty string without error if the file doesn't exist at that commit.
func FetchFileContent(commitHash, filePath string) (string, error) {
	return operations.FetchFileContent(commitHash, filePath)
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

// FetchFileDiff retrieves the unified diff for a specific file in a commit.
// The filePath should be relative to the repository root.
func FetchFileDiff(commitHash, filePath string) (string, error) {
	return operations.FetchFileDiff(commitHash, filePath)
}

// FetchFileDiffRange retrieves the unified diff for a specific file in a commit range.
// The rangeSpec should be in the format "fromHash..toHash".
// The filePath should be relative to the repository root.
func FetchFileDiffRange(rangeSpec, filePath string) (string, error) {
	return operations.FetchFileDiffRange(rangeSpec, filePath)
}

// FetchUnstagedFileChanges executes git diff and returns a slice of file changes
// for unstaged changes (working tree vs index).
func FetchUnstagedFileChanges() ([]core.FileChange, error) {
	return operations.FetchUnstagedFileChanges()
}

// FetchStagedFileChanges executes git diff --staged and returns a slice of file changes
// for staged changes (index vs HEAD).
func FetchStagedFileChanges() ([]core.FileChange, error) {
	return operations.FetchStagedFileChanges()
}

// FetchAllUncommittedFileChanges executes git diff HEAD and returns a slice of file changes
// for all uncommitted changes (working tree vs HEAD).
func FetchAllUncommittedFileChanges() ([]core.FileChange, error) {
	return operations.FetchAllUncommittedFileChanges()
}

// FetchIndexFileContent retrieves the content of a file from the index (staging area).
// Returns empty string without error if the file doesn't exist in the index.
func FetchIndexFileContent(filePath string) (string, error) {
	return operations.FetchIndexFileContent(filePath)
}

// FetchWorkingTreeFileContent retrieves the content of a file from the working tree.
// Returns empty string without error if the file doesn't exist.
func FetchWorkingTreeFileContent(filePath string) (string, error) {
	return operations.FetchWorkingTreeFileContent(filePath)
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

// ResolveRef resolves a git ref (like "HEAD", "main", "abc123") to a GitCommit.
func ResolveRef(ref string) (core.GitCommit, error) {
	return operations.ResolveRef(ref)
}

// GetRepositoryRoot executes git rev-parse --show-toplevel to get the absolute path of the repository root.
func GetRepositoryRoot() (string, error) {
	return operations.GetRepositoryRoot()
}

// ParseGitLogOutput parses git log output into GitCommit structs.
// Input format: "hash\0parents\0refs\0author\0date\0subject\0body\x1e" (NULL-separated fields, record separator between commits).
func ParseGitLogOutput(output string) ([]core.GitCommit, error) {
	return parsing.ParseGitLogOutput(output)
}

// ParseFileChangesOutput parses git diff output into FileChange structs.
// Input format: "additions\tdeletions\tfilepath" (one file per line).
func ParseFileChangesOutput(output string) ([]core.FileChange, error) {
	return parsing.ParseFileChangesOutput(output)
}
