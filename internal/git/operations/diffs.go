package operations

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// FetchFullFileDiff fetches the complete file content before and after a change,
// along with the diff output. This enables showing the full file with changes highlighted.
func FetchFullFileDiff(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
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

	result := &core.FullFileDiffResult{
		NewPath: change.Path,
		OldPath: change.Path,
	}

	// TODO: Handle renames properly (status starts with "R")
	// For renames, the path contains "old -> new" format, but we get OldPath from git
	// In our FileChange struct, we only have Path (the new path)
	// We need to handle this differently - for now assume same path

	// Fetch new content (at toHash)
	switch change.Status {
	case "D": // Deleted file - no new content
		result.NewContent = ""
	default:
		newContent, err := commands.FetchFileContent(toHash, change.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch old content (at fromHash)
	switch change.Status {
	case "A": // Added file - no old content
		result.OldContent = ""
	default:
		oldContent, err := commands.FetchFileContent(fromHash, result.OldPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch the diff
	rangeSpec := fromHash + ".." + toHash
	diffOutput, err := commands.FetchFileDiffForRange(rangeSpec, change.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch diff: %w", err)
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchUnstagedFileDiff fetches the complete file content before and after an unstaged change,
// along with the diff output. This enables showing the full file with changes highlighted.
// Unstaged changes compare the index (old) with the working tree (new).
func FetchUnstagedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from index)
	switch file.Status {
	case "A": // Added file - no old content in index
		result.OldContent = ""
	default:
		oldContent, err := commands.FetchIndexFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from working tree)
	switch file.Status {
	case "D": // Deleted file - no new content in working tree
		result.NewContent = ""
	default:
		newContent, err := commands.FetchWorkingTreeFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffOutput, err := commands.FetchFileDiffWithFlags(file.Path, "diff")
	if err != nil {
		return nil, err
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchStagedFileDiff fetches the complete file content before and after a staged change,
// along with the diff output. This enables showing the full file with changes highlighted.
// Staged changes compare HEAD (old) with the index (new).
func FetchStagedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from HEAD)
	switch file.Status {
	case "A": // Added file - no old content in HEAD
		result.OldContent = ""
	default:
		oldContent, err := commands.FetchFileContent("HEAD", file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from index)
	switch file.Status {
	case "D": // Deleted file - no new content in index
		result.NewContent = ""
	default:
		newContent, err := commands.FetchIndexFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffOutput, err := commands.FetchFileDiffWithFlags(file.Path, "diff", "--staged")
	if err != nil {
		return nil, err
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchAllUncommittedFileDiff fetches the complete file content before and after all uncommitted changes,
// along with the diff output. This enables showing the full file with changes highlighted.
// All uncommitted changes compare HEAD (old) with the working tree (new).
func FetchAllUncommittedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from HEAD)
	switch file.Status {
	case "A": // Added file - no old content in HEAD
		result.OldContent = ""
	default:
		oldContent, err := commands.FetchFileContent("HEAD", file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from working tree)
	switch file.Status {
	case "D": // Deleted file - no new content in working tree
		result.NewContent = ""
	default:
		newContent, err := commands.FetchWorkingTreeFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffOutput, err := commands.FetchFileDiffWithFlags(file.Path, "diff", "HEAD")
	if err != nil {
		return nil, err
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchFileContent retrieves the content of a file at a specific commit.
// Returns empty string without error if the file doesn't exist at that commit.
func FetchFileContent(commitHash, filePath string) (string, error) {
	return commands.FetchFileContent(commitHash, filePath)
}

// FetchFileDiff retrieves the unified diff for a specific file in a commit.
// The filePath should be relative to the repository root.
func FetchFileDiff(commitHash, filePath string) (string, error) {
	return commands.FetchFileDiffFromCommit(commitHash, filePath)
}

// FetchFileDiffRange retrieves the unified diff for a specific file in a commit range.
// The rangeSpec should be in the format "fromHash..toHash".
// The filePath should be relative to the repository root.
func FetchFileDiffRange(rangeSpec, filePath string) (string, error) {
	return commands.FetchFileDiffForRange(rangeSpec, filePath)
}

// FetchIndexFileContent retrieves the content of a file from the index (staging area).
// Returns empty string without error if the file doesn't exist in the index.
func FetchIndexFileContent(filePath string) (string, error) {
	return commands.FetchIndexFileContent(filePath)
}

// FetchWorkingTreeFileContent retrieves the content of a file from the working tree.
// Returns empty string without error if the file doesn't exist.
func FetchWorkingTreeFileContent(filePath string) (string, error) {
	return commands.FetchWorkingTreeFileContent(filePath)
}
