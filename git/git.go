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
	Hash         string    // Full 40-char hash
	ShortHash    string    // First 7 chars
	Message      string    // First line of commit message
	Author       string    // Author name (not email)
	Date         time.Time // Commit timestamp
	RelativeTime string    // "4 min ago", computed
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
			Hash:         hash,
			ShortHash:    hash[:7],
			Message:      message,
			Author:       author,
			Date:         date,
			RelativeTime: FormatRelativeTime(date),
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// FormatRelativeTime converts a timestamp into a human-readable relative time
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	seconds := int(diff.Seconds())
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		return "just now"
	case minutes < 60:
		if minutes == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 7:
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case days < 30:
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case days < 365:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "1 year ago"
		}
		// For old commits, show absolute date
		return t.Format("Jan 2, 2006")
	}
}
