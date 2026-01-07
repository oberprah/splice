package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestVisualModeMultiCommitSelection tests the full visual mode workflow:
// 1. Start in LogState with multiple commits
// 2. Press 'v' to enter visual mode (cursor changes to █)
// 3. Press 'j' to move down and extend selection (selected commits show ▌)
// 4. Press Enter to navigate to FilesState
// 5. Verify FilesState shows the range header like "abc123d..def456e (N commits)"
// 6. Verify FilesState shows combined file changes from all commits
func TestVisualModeMultiCommitSelection(t *testing.T) {
	// Create mock commits for testing
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add authentication feature",
		"Fix login validation",
		"Update user model",
		"Refactor auth flow",
		"Initial auth setup",
	})

	// Fixed time for deterministic date formatting
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create mock file changes that will be returned for the commit range
	// These simulate combined changes from multiple commits
	mockFiles := []core.FileChange{
		{Path: "auth/login.go", Status: "M", Additions: 45, Deletions: 12},
		{Path: "auth/validation.go", Status: "A", Additions: 28, Deletions: 0},
		{Path: "models/user.go", Status: "M", Additions: 15, Deletions: 8},
		{Path: "routes/auth.go", Status: "M", Additions: 22, Deletions: 5},
	}

	// Create a mock fetch function that returns our test files for any range
	mockFetchFileChanges := testutils.MockFetchFileChanges(mockFiles, nil)

	// Create model with mocked commits and file changes
	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(mockFetchFileChanges),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size and wait for loading to complete
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("visual_mode/1_initial.golden")

	// Press 'v' to enter visual mode
	// Cursor should change from → to █
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	runner.AssertGolden("visual_mode/2_enter_visual_mode.golden")

	// Press 'j' to move down and extend selection
	// First commit should show ▌ (selected), second commit should show █ (visual cursor)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("visual_mode/3_extend_selection_once.golden")

	// Press 'j' again to extend selection further
	// First and second commits show ▌, third commit shows █
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("visual_mode/4_extend_selection_twice.golden")

	// Press Enter to navigate to FilesState
	// Should show range header like "0000000..0000000 (3 commits)"
	// and combined file changes from all selected commits
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.AssertGolden("visual_mode/5_files_state_with_range.golden")

	// Go back to log (q from files view)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Quit (q from log view quits the app)
	runner.Quit()
}

// TestVisualModeUpwardSelection tests selecting commits by moving up in visual mode
func TestVisualModeUpwardSelection(t *testing.T) {
	// Create mock commits
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Third commit",
		"Second commit",
		"First commit",
	})

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	mockFiles := []core.FileChange{
		{Path: "file.go", Status: "M", Additions: 10, Deletions: 5},
	}

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("visual_mode_upward/1_initial.golden")

	// Move cursor down to second commit
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("visual_mode_upward/2_cursor_at_second.golden")

	// Enter visual mode
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	runner.AssertGolden("visual_mode_upward/3_visual_mode_at_second.golden")

	// Move up with 'k' to extend selection upward
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	runner.AssertGolden("visual_mode_upward/4_extend_upward.golden")

	// Quit
	runner.Quit()
}

// TestVisualModeEscape tests exiting visual mode with Escape key
func TestVisualModeEscape(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"First commit",
		"Second commit",
		"Third commit",
	})

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges([]core.FileChange{}, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("visual_mode_escape/1_initial.golden")

	// Enter visual mode
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	runner.AssertGolden("visual_mode_escape/2_visual_mode.golden")

	// Extend selection
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("visual_mode_escape/3_selection_extended.golden")

	// Press Escape to exit visual mode
	runner.Send(tea.KeyMsg{Type: tea.KeyEsc})
	runner.AssertGolden("visual_mode_escape/4_after_escape.golden")

	// Quit
	runner.Quit()
}
