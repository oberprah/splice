package testutils

import (
	"fmt"
	"time"

	"github.com/oberprah/splice/internal/git"
)

// CreateTestCommits generates n mock git commits for testing
func CreateTestCommits(count int) []git.GitCommit {
	commits := make([]git.GitCommit, count)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for i := range count {
		commits[i] = git.GitCommit{
			Hash:    fmt.Sprintf("%040d", i), // Full 40-char hash
			Message: fmt.Sprintf("Commit message %d", i),
			Body:    "",
			Author:  fmt.Sprintf("Author %d", i%3),               // Vary authors
			Date:    baseTime.Add(time.Duration(-i) * time.Hour), // Reverse chronological
		}
	}

	return commits
}

// CreateTestCommitsWithMessages generates commits with specific messages
func CreateTestCommitsWithMessages(messages []string) []git.GitCommit {
	commits := make([]git.GitCommit, len(messages))
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for i, msg := range messages {
		commits[i] = git.GitCommit{
			Hash:    fmt.Sprintf("%040d", i),
			Message: msg,
			Body:    "",
			Author:  "Test Author",
			Date:    baseTime.Add(time.Duration(-i) * time.Hour),
		}
	}

	return commits
}

// MockFetchCommits creates a mock function that returns test commits
func MockFetchCommits(commits []git.GitCommit, err error) func(int) ([]git.GitCommit, error) {
	return func(limit int) ([]git.GitCommit, error) {
		if err != nil {
			return nil, err
		}
		if limit < len(commits) {
			return commits[:limit], nil
		}
		return commits, nil
	}
}

// MockFetchFileChanges creates a mock function that returns file changes for a commit
func MockFetchFileChanges(files []git.FileChange, err error) func(string) ([]git.FileChange, error) {
	return func(commitHash string) ([]git.FileChange, error) {
		if err != nil {
			return nil, err
		}
		return files, nil
	}
}

// MockFetchFullFileDiff creates a mock function that returns full file diff result
func MockFetchFullFileDiff(result *git.FullFileDiffResult, err error) func(string, git.FileChange) (*git.FullFileDiffResult, error) {
	return func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error) {
		if err != nil {
			return nil, err
		}
		return result, nil
	}
}
