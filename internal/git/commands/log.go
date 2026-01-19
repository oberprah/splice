package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

// FetchLog executes git log with custom format and returns parsed commits.
func FetchLog(limit int) ([]core.GitCommit, error) {
	// Use git log with custom format using NULL separator: hash\0parents\0refs\0author\0date\0subject\0body
	// NULL character is used as field delimiter since it won't appear in commit messages
	// ASCII Record Separator (0x1e) is used as commit record separator
	// %P outputs parent hashes (space-separated for merges, empty for root commits)
	// %d outputs ref decorations (e.g., " (HEAD -> main, tag: v1.0)")
	stdout, stderr, err := execGit("log",
		"--pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e",
		"--date=iso-strict",
		fmt.Sprintf("-n %d", limit))

	if err != nil {
		return nil, checkGitError(stderr, err, "git log")
	}

	return git.ParseGitLogOutput(stdout)
}

// ResolveRefToCommit resolves a git ref (like "HEAD", "main", "abc123") to a GitCommit.
func ResolveRefToCommit(ref string) (core.GitCommit, error) {
	stdout, stderr, err := execGit("log", "-1", "--format=%H%n%s%n%an%n%aI%n%P", ref)
	if err != nil {
		return core.GitCommit{}, checkGitError(stderr, err, "git log")
	}

	// Parse the output (lines: hash, message, author, date, parents)
	commit, err := parseSimpleCommit(stdout)
	if err != nil {
		return core.GitCommit{}, err
	}

	return commit, nil
}

// FindMergeBase finds the merge base of two refs.
func FindMergeBase(ref1, ref2 string) (string, error) {
	stdout, stderr, err := execGit("merge-base", ref1, ref2)
	if err != nil {
		return "", checkGitError(stderr, err, "git merge-base")
	}

	return strings.TrimSpace(stdout), nil
}

// CountCommitsInRange counts commits between two refs.
func CountCommitsInRange(startRef, endRef string) (int, error) {
	stdout, stderr, err := execGit("rev-list", "--count", startRef+".."+endRef)
	if err != nil {
		return 0, checkGitError(stderr, err, "git rev-list")
	}

	var count int
	if _, err := fmt.Sscanf(stdout, "%d", &count); err != nil {
		return 0, fmt.Errorf("error parsing commit count: %v", err)
	}
	return count, nil
}

// parseSimpleCommit parses the output of git log -1 --format=%H%n%s%n%an%n%aI%n%P
// into a GitCommit struct.
func parseSimpleCommit(output string) (core.GitCommit, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 4 {
		return core.GitCommit{}, fmt.Errorf("unexpected git log output")
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[3]))
	if err != nil {
		return core.GitCommit{}, fmt.Errorf("error parsing date: %v", err)
	}

	// Parse parent hashes
	var parents []string
	if len(lines) >= 5 && lines[4] != "" {
		parents = strings.Fields(lines[4])
	}

	return core.GitCommit{
		Hash:         lines[0],
		Message:      lines[1],
		Author:       lines[2],
		Date:         date,
		ParentHashes: parents,
		Refs:         []core.RefInfo{},
	}, nil
}
