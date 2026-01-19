package commands

import (
	"strings"
)

// FetchFileContent retrieves the content of a file at a specific commit.
// Returns empty string without error if the file doesn't exist at that commit.
func FetchFileContent(commitHash, filePath string) (string, error) {
	stdout, stderr, err := execGit("show", commitHash+":"+filePath)
	if err != nil {
		stderrStr := strings.TrimSpace(stderr)
		// Check if this is a "file not found" error - return empty string, no error
		if strings.Contains(stderrStr, "does not exist") ||
			strings.Contains(stderrStr, "exists on disk, but not in") ||
			strings.Contains(stderrStr, "fatal: path") {
			return "", nil
		}
		// Check for invalid commit
		if strings.Contains(stderrStr, "unknown revision") ||
			strings.Contains(stderrStr, "bad revision") ||
			strings.Contains(stderrStr, "not a valid object name") {
			return "", checkGitError(stderr, err, "git show")
		}
		return "", checkGitError(stderr, err, "git show")
	}

	return stdout, nil
}

// FetchFileDiffFromCommit retrieves the unified diff for a specific file in a commit.
// The filePath should be relative to the repository root.
func FetchFileDiffFromCommit(commitHash, filePath string) (string, error) {
	// Use :(top) pathspec to ensure path is relative to repo root regardless of cwd
	stdout, stderr, err := execGit("show", commitHash, "--format=", "--", ":(top)"+filePath)
	if err != nil {
		return "", checkGitError(stderr, err, "git show")
	}

	return stdout, nil
}

// FetchFileDiffForRange retrieves the unified diff for a specific file in a commit range.
// The rangeSpec should be in the format "fromHash..toHash".
// The filePath should be relative to the repository root.
func FetchFileDiffForRange(rangeSpec, filePath string) (string, error) {
	// Use :(top) pathspec to ensure path is relative to repo root regardless of cwd
	stdout, stderr, err := execGit("diff", rangeSpec, "--", ":(top)"+filePath)
	if err != nil {
		return "", checkGitError(stderr, err, "git diff")
	}

	return stdout, nil
}

// FetchIndexFileContent retrieves the content of a file from the index (staging area).
// Returns empty string without error if the file doesn't exist in the index.
func FetchIndexFileContent(filePath string) (string, error) {
	stdout, stderr, err := execGit("show", ":"+filePath)
	if err != nil {
		stderrStr := strings.TrimSpace(stderr)
		// Check if this is a "file not found" error - return empty string, no error
		if strings.Contains(stderrStr, "does not exist") ||
			strings.Contains(stderrStr, "exists on disk, but not in") ||
			strings.Contains(stderrStr, "fatal: path") {
			return "", nil
		}
		return "", checkGitError(stderr, err, "git show")
	}

	return stdout, nil
}
