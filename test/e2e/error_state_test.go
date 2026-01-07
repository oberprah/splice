package main

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/git"
	_ "github.com/oberprah/splice/internal/ui/states"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestErrorState tests error handling when git command fails
func TestErrorState(t *testing.T) {
	// Create model with error-returning mock
	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(nil, fmt.Errorf("not a git repository"))),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges([]git.FileChange{}, nil)),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size and wait for error state to display
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("error_state/1_error_displayed.golden")

	// Quit from error state
	runner.Quit()
}
