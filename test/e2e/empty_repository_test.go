package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestEmptyRepository tests handling of empty commit list
func TestEmptyRepository(t *testing.T) {
	// Create model with empty commits (returns empty slice, not an error)
	emptyCommits := []git.GitCommit{}
	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(emptyCommits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges([]git.FileChange{}, nil)),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size and wait for error state to display
	// The app treats empty commits as an error case showing "no commits found"
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("empty_repository/1_no_commits_error.golden")

	// Quit from the error state
	runner.Quit()
}
