package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// GitCommit represents a single git commit with all necessary display information
type GitCommit struct {
	Hash    string    // Full 40-char hash
	Message string    // First line of commit message
	Author  string    // Author name (not email)
	Date    time.Time // Commit timestamp
}

// FileChange represents a file that was changed in a commit
type FileChange struct {
	Path      string // File path relative to repository root
	Status    string // Git status: M (modified), A (added), D (deleted), R (renamed), etc.
	Additions int    // Number of lines added
	Deletions int    // Number of lines deleted
	IsBinary  bool   // True if the file is binary
}

// ParseGitLogOutput parses git log output into GitCommit structs.
// Input format: "hash|author|date|message" (one commit per line).
func ParseGitLogOutput(output string) ([]GitCommit, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	commits := make([]GitCommit, 0, len(lines))

	for i, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("malformed line %d: expected 4 pipe-separated fields, got %d: %q", i+1, len(parts), line)
		}

		hash := parts[0]
		author := parts[1]
		dateStr := parts[2]
		message := parts[3]

		// Parse the date
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", dateStr, err)
		}

		commit := GitCommit{
			Hash:    hash,
			Message: message,
			Author:  author,
			Date:    date,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// FetchCommits executes git log and returns a slice of commits
func FetchCommits(limit int) ([]GitCommit, error) {
	// Use git log with custom format: hash|author|date|message
	cmd := exec.Command("git", "log",
		"--pretty=format:%H|%an|%ad|%s",
		"--date=iso-strict",
		fmt.Sprintf("-n %d", limit))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if this is not a git repository
		if strings.Contains(stderr.String(), "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git log failed: %v - %s", err, stderr.String())
	}

	return ParseGitLogOutput(out.String())
}

// ParseFileChangesOutput parses git diff output into FileChange structs.
// Input format: "additions\tdeletions\tfilepath" (one file per line).
func ParseFileChangesOutput(output string) ([]FileChange, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	changes := make([]FileChange, 0, len(lines))

	for i, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("malformed line %d: expected 3 tab-separated fields, got %d: %q", i+1, len(parts), line)
		}

		additionsStr := parts[0]
		deletionsStr := parts[1]
		path := parts[2]

		var additions, deletions int
		var isBinary bool

		// Check if this is a binary file (git shows "-" for both additions and deletions)
		if additionsStr == "-" && deletionsStr == "-" {
			isBinary = true
			additions = 0
			deletions = 0
		} else {
			// Parse additions
			var err error
			_, err = fmt.Sscanf(additionsStr, "%d", &additions)
			if err != nil {
				return nil, fmt.Errorf("invalid additions count on line %d: %q", i+1, additionsStr)
			}

			// Parse deletions
			_, err = fmt.Sscanf(deletionsStr, "%d", &deletions)
			if err != nil {
				return nil, fmt.Errorf("invalid deletions count on line %d: %q", i+1, deletionsStr)
			}
		}

		change := FileChange{
			Path:      path,
			Additions: additions,
			Deletions: deletions,
			IsBinary:  isBinary,
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// FetchFileChanges executes git diff and returns a slice of file changes for a commit
func FetchFileChanges(commitHash string) ([]FileChange, error) {
	// First, get file statuses (A/M/D/R)
	statusCmd := exec.Command("git", "diff-tree", "--no-commit-id", "--name-status", "-r", commitHash)
	var statusOut bytes.Buffer
	var statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr

	err := statusCmd.Run()
	if err != nil {
		stderrStr := statusErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return nil, fmt.Errorf("invalid commit: %s", commitHash)
		}
		return nil, fmt.Errorf("git diff-tree failed: %v - %s", err, stderrStr)
	}

	// Parse status information into a map
	statusMap := make(map[string]string)
	statusLines := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	for _, line := range statusLines {
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

	// Now get numstat information
	cmd := exec.Command("git", "diff-tree", "--no-commit-id", "--numstat", "-r", commitHash)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return nil, fmt.Errorf("invalid commit: %s", commitHash)
		}
		return nil, fmt.Errorf("git diff-tree failed: %v - %s", err, stderrStr)
	}

	// Parse file changes and add status
	changes, err := ParseFileChangesOutput(out.String())
	if err != nil {
		return nil, err
	}

	// Add status to each change
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}

	return changes, nil
}
