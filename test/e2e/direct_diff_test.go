package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/states/directdiff"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestDirectDiff_UnstagedChanges tests the workflow for viewing unstaged changes
func TestDirectDiff_UnstagedChanges(t *testing.T) {
	// Create test data
	files := []core.FileChange{
		{Path: "main.go", Status: "M", Additions: 10, Deletions: 5, IsBinary: false},
		{Path: "test.go", Status: "M", Additions: 2, Deletions: 1, IsBinary: false},
	}

	// Create DiffSource for unstaged changes
	diffSource := core.UncommittedChangesDiffSource{
		Type: core.UncommittedTypeUnstaged,
	}

	// Mock function to return files for any DiffSource
	mockFetchForSource := func(source core.DiffSource) ([]core.FileChange, error) {
		return files, nil
	}

	// Fixed time for deterministic output
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with DirectDiffLoadingState
	m := app.NewModel(
		app.WithInitialState(directdiff.New(diffSource)),
		app.WithFetchFileChangesForSource(mockFetchForSource),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("direct_diff/unstaged_loading.golden")

	// Should transition to FilesState
	runner.AssertGolden("direct_diff/unstaged_files.golden")

	// Quit (should exit app due to ExitOnPop=true)
	runner.Quit()
}

// TestDirectDiff_StagedChanges tests the workflow for viewing staged changes
func TestDirectDiff_StagedChanges(t *testing.T) {
	// Create test data
	files := []core.FileChange{
		{Path: "auth.go", Status: "A", Additions: 50, Deletions: 0, IsBinary: false},
		{Path: "config.go", Status: "M", Additions: 8, Deletions: 3, IsBinary: false},
	}

	// Create DiffSource for staged changes
	diffSource := core.UncommittedChangesDiffSource{
		Type: core.UncommittedTypeStaged,
	}

	// Mock function to return files
	mockFetchForSource := func(source core.DiffSource) ([]core.FileChange, error) {
		return files, nil
	}

	// Fixed time for deterministic output
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with DirectDiffLoadingState
	m := app.NewModel(
		app.WithInitialState(directdiff.New(diffSource)),
		app.WithFetchFileChangesForSource(mockFetchForSource),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("direct_diff/staged_loading.golden")

	// Should transition to FilesState
	runner.AssertGolden("direct_diff/staged_files.golden")

	// Quit
	runner.Quit()
}

// TestDirectDiff_AllUncommitted tests the workflow for viewing all uncommitted changes
func TestDirectDiff_AllUncommitted(t *testing.T) {
	// Create test data (combination of staged and unstaged)
	files := []core.FileChange{
		{Path: "main.go", Status: "M", Additions: 15, Deletions: 8, IsBinary: false},
		{Path: "auth.go", Status: "A", Additions: 50, Deletions: 0, IsBinary: false},
		{Path: "old.go", Status: "D", Additions: 0, Deletions: 30, IsBinary: false},
	}

	// Create DiffSource for all uncommitted changes
	diffSource := core.UncommittedChangesDiffSource{
		Type: core.UncommittedTypeAll,
	}

	// Mock function to return files
	mockFetchForSource := func(source core.DiffSource) ([]core.FileChange, error) {
		return files, nil
	}

	// Fixed time for deterministic output
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with DirectDiffLoadingState
	m := app.NewModel(
		app.WithInitialState(directdiff.New(diffSource)),
		app.WithFetchFileChangesForSource(mockFetchForSource),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("direct_diff/all_loading.golden")

	// Should transition to FilesState
	runner.AssertGolden("direct_diff/all_files.golden")

	// Quit
	runner.Quit()
}

// TestDirectDiff_CommitRange tests the workflow for viewing a commit range
func TestDirectDiff_CommitRange(t *testing.T) {
	// Create test commits
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add feature",
		"Fix bug",
		"Initial commit",
	})

	// Create test data
	files := []core.FileChange{
		{Path: "feature.go", Status: "A", Additions: 100, Deletions: 0, IsBinary: false},
		{Path: "main.go", Status: "M", Additions: 20, Deletions: 5, IsBinary: false},
	}

	// Create DiffSource for commit range (commits[2]..commits[0])
	diffSource := core.CommitRangeDiffSource{
		Start: commits[2], // Initial commit (older)
		End:   commits[0], // Add feature (newer)
		Count: 3,
	}

	// Mock function to return files
	mockFetchForSource := func(source core.DiffSource) ([]core.FileChange, error) {
		return files, nil
	}

	// Fixed time for deterministic output
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with DirectDiffLoadingState
	m := app.NewModel(
		app.WithInitialState(directdiff.New(diffSource)),
		app.WithFetchFileChangesForSource(mockFetchForSource),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("direct_diff/range_loading.golden")

	// Should transition to FilesState
	runner.AssertGolden("direct_diff/range_files.golden")

	// Quit
	runner.Quit()
}

// TestDirectDiff_ExitOnPopBehavior tests that pressing 'q' in FilesState
// quits the app when ExitOnPop=true (direct diff view), rather than
// returning to a previous screen like in the log view workflow
func TestDirectDiff_ExitOnPopBehavior(t *testing.T) {
	// Create test data
	files := []core.FileChange{
		{Path: "test.go", Status: "M", Additions: 5, Deletions: 2, IsBinary: false},
	}

	// Create DiffSource for unstaged changes
	diffSource := core.UncommittedChangesDiffSource{
		Type: core.UncommittedTypeUnstaged,
	}

	// Mock function to return files
	mockFetchForSource := func(source core.DiffSource) ([]core.FileChange, error) {
		return files, nil
	}

	// Fixed time for deterministic output
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with DirectDiffLoadingState
	m := app.NewModel(
		app.WithInitialState(directdiff.New(diffSource)),
		app.WithFetchFileChangesForSource(mockFetchForSource),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Should transition to FilesState
	runner.AssertGolden("direct_diff/exit_on_pop_files.golden")

	// Press 'q' - should quit the app (ExitOnPop=true), not return to a previous screen
	// The Quit() call will verify that the program actually exits
	runner.Quit()
}
