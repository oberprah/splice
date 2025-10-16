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
