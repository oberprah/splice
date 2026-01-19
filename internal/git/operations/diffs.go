package operations

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// contentSource identifies where to fetch file content from
type contentSource int

const (
	sourceIndex       contentSource = iota // Git index (staging area)
	sourceWorkingTree                      // Working tree (filesystem)
	sourceHEAD                             // HEAD commit
)

// fetchContent retrieves file content from the specified source
func fetchContent(source contentSource, filePath string) (string, error) {
	switch source {
	case sourceIndex:
		return commands.FetchIndexFileContent(filePath)
	case sourceWorkingTree:
		return commands.FetchWorkingTreeFileContent(filePath)
	case sourceHEAD:
		return commands.FetchFileContent("HEAD", filePath)
	default:
		return "", fmt.Errorf("unknown content source: %d", source)
	}
}

// fetchUncommittedFileDiff is the consolidated implementation for all uncommitted diff types.
// It fetches old content, new content, and diff output based on the specified sources.
func fetchUncommittedFileDiff(
	file core.FileChange,
	oldSource contentSource,
	newSource contentSource,
	diffFlags ...string,
) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content
	switch file.Status {
	case "A": // Added file - no old content
		result.OldContent = ""
	default:
		oldContent, err := fetchContent(oldSource, file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content
	switch file.Status {
	case "D": // Deleted file - no new content
		result.NewContent = ""
	default:
		newContent, err := fetchContent(newSource, file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	args := append([]string{"diff"}, diffFlags...)
	diffOutput, err := commands.FetchFileDiffWithFlags(file.Path, args...)
	if err != nil {
		return nil, err
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchUncommittedFileDiff fetches the full file diff for any uncommitted change type.
// This is the unified interface that dispatches to the correct sources based on type.
func FetchUncommittedFileDiff(uncommittedType core.UncommittedType, file core.FileChange) (*core.FullFileDiffResult, error) {
	switch uncommittedType {
	case core.UncommittedTypeUnstaged:
		return fetchUncommittedFileDiff(file, sourceIndex, sourceWorkingTree)
	case core.UncommittedTypeStaged:
		return fetchUncommittedFileDiff(file, sourceHEAD, sourceIndex, "--staged")
	case core.UncommittedTypeAll:
		return fetchUncommittedFileDiff(file, sourceHEAD, sourceWorkingTree, "HEAD")
	default:
		return nil, fmt.Errorf("unknown uncommitted type: %v", uncommittedType)
	}
}

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
