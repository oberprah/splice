package commands

import (
	"fmt"
	"strings"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

// FetchFileChangesWithFlags executes git diff with custom flags and returns file changes.
// This is the base function that eliminates duplication across unstaged/staged/all functions.
//
// Common usage:
//   - Unstaged: FetchFileChangesWithFlags("diff")
//   - Staged: FetchFileChangesWithFlags("diff", "--staged")
//   - All: FetchFileChangesWithFlags("diff", "HEAD")
func FetchFileChangesWithFlags(args ...string) ([]core.FileChange, error) {
	// First, get file statuses (A/M/D/R)
	statusArgs := append([]string{}, args...)
	statusArgs = append(statusArgs, "--name-status")

	statusOut, statusErr, err := execGit(statusArgs...)
	if err != nil {
		return nil, checkGitError(statusErr, err, "git diff")
	}

	// Parse status information into a map
	statusMap := parseStatusMap(statusOut)

	// Now get numstat information
	numstatArgs := append([]string{}, args...)
	numstatArgs = append(numstatArgs, "--numstat")

	numstatOut, numstatErr, err := execGit(numstatArgs...)
	if err != nil {
		return nil, checkGitError(numstatErr, err, "git diff")
	}

	// Parse file changes
	changes, err := git.ParseFileChangesOutput(numstatOut)
	if err != nil {
		return nil, err
	}

	// Add status to each change
	return addStatusToChanges(changes, statusMap), nil
}

// parseStatusMap parses git diff --name-status output into a map of path -> status.
func parseStatusMap(output string) map[string]string {
	statusMap := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			status := parts[0]
			path := parts[1]
			statusMap[path] = status
		}
	}

	return statusMap
}

// addStatusToChanges adds status information to file changes.
// If a file has no status in the map, defaults to "M" (modified).
func addStatusToChanges(changes []core.FileChange, statusMap map[string]string) []core.FileChange {
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}
	return changes
}

// FetchFileDiffWithFlags executes git diff with custom flags for a specific file.
// This is the base function for fetching unified diffs with various flags.
//
// Common usage:
//   - Unstaged: FetchFileDiffWithFlags(filePath, "diff")
//   - Staged: FetchFileDiffWithFlags(filePath, "diff", "--staged")
//   - All: FetchFileDiffWithFlags(filePath, "diff", "HEAD")
func FetchFileDiffWithFlags(filePath string, args ...string) (string, error) {
	// Append file path to arguments
	diffArgs := append([]string{}, args...)
	diffArgs = append(diffArgs, "--", filePath)

	stdout, stderr, err := execGit(diffArgs...)
	if err != nil {
		return "", checkGitError(stderr, err, "git diff")
	}

	return stdout, nil
}

// FetchFileChangesForRange executes git diff for a commit range and returns file changes.
// The rangeSpec should be in the format "fromHash..toHash".
func FetchFileChangesForRange(rangeSpec string) ([]core.FileChange, error) {
	// First, get file statuses (A/M/D/R)
	statusOut, statusErr, err := execGit("diff", "--name-status", rangeSpec)
	if err != nil {
		if strings.Contains(statusErr, "unknown revision") || strings.Contains(statusErr, "bad revision") {
			return nil, fmt.Errorf("invalid commit range: %s", rangeSpec)
		}
		return nil, checkGitError(statusErr, err, "git diff")
	}

	// Parse status information into a map
	statusMap := parseStatusMap(statusOut)

	// Now get numstat information
	numstatOut, numstatErr, err := execGit("diff", "--numstat", rangeSpec)
	if err != nil {
		if strings.Contains(numstatErr, "unknown revision") || strings.Contains(numstatErr, "bad revision") {
			return nil, fmt.Errorf("invalid commit range: %s", rangeSpec)
		}
		return nil, checkGitError(numstatErr, err, "git diff")
	}

	// Parse file changes and add status
	changes, err := git.ParseFileChangesOutput(numstatOut)
	if err != nil {
		return nil, err
	}

	return addStatusToChanges(changes, statusMap), nil
}
