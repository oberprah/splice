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

	// Parse the output
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	commits := make([]GitCommit, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue // Skip malformed lines
		}

		hash := parts[0]
		author := parts[1]
		dateStr := parts[2]
		message := parts[3]

		// Parse the date
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			// If parsing fails, use current time as fallback
			date = time.Now()
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
