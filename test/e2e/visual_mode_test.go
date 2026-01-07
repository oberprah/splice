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

	// Create a mock fetch function that returns DIFFERENT files depending on the commit range
	// This allows us to verify that the correct files are loaded after escape
	mockFetchFileChanges := func(commitRange core.CommitRange) ([]core.FileChange, error) {
		if commitRange.IsSingleCommit() {
			// Single commit - return files based on which commit
			switch commitRange.End.Hash {
			case commits[0].Hash:
				// First commit
				return []core.FileChange{
					{Path: "first.go", Status: "M", Additions: 10, Deletions: 5},
					{Path: "shared.go", Status: "A", Additions: 20, Deletions: 0},
				}, nil
			case commits[1].Hash:
				// Second commit - different files to distinguish from first commit
				return []core.FileChange{
					{Path: "second.go", Status: "M", Additions: 15, Deletions: 3},
					{Path: "another.go", Status: "A", Additions: 8, Deletions: 2},
				}, nil
			default:
				// Third commit or other
				return []core.FileChange{
					{Path: "third.go", Status: "M", Additions: 5, Deletions: 1},
				}, nil
			}
		}
		// Multi-commit range - return combined files
		return []core.FileChange{
			{Path: "combined1.go", Status: "M", Additions: 25, Deletions: 8},
			{Path: "combined2.go", Status: "A", Additions: 28, Deletions: 2},
		}, nil
	}

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(mockFetchFileChanges),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size (wide enough to show split view with preview)
	// Step 0: Detail view needs to be visible (wide terminal, split view)
	runner.Send(tea.WindowSizeMsg{Width: 160, Height: 24})
	// Step 1: Start with files already loaded in the preview for the initial commit
	runner.AssertGolden("visual_mode_escape/1_initial.golden")

	// Step 2: Enter visual mode with 'v'
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	runner.AssertGolden("visual_mode_escape/2_visual_mode.golden")

	// Step 3: Select at least 2 commits by pressing 'j'
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	// Step 4: Verify files of both commits (combined) are visible in preview
	runner.AssertGolden("visual_mode_escape/3_selection_extended.golden")

	// Step 5: Exit visual mode with Escape
	runner.Send(tea.KeyMsg{Type: tea.KeyEsc})
	// Step 6 & 7: Verify commit message of the selected commit is visible
	// AND files show the actual files for that single commit (not "loading files..." or combined files)
	runner.AssertGolden("visual_mode_escape/4_after_escape.golden")

	// Quit
	runner.Quit()
}

// TestVisualModeToggleOut tests exiting visual mode by pressing 'v' again
func TestVisualModeToggleOut(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"First commit",
		"Second commit",
		"Third commit",
	})

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create a mock fetch function that returns DIFFERENT files depending on the commit range
	mockFetchFileChanges := func(commitRange core.CommitRange) ([]core.FileChange, error) {
		if commitRange.IsSingleCommit() {
			// Single commit - return files based on which commit
			switch commitRange.End.Hash {
			case commits[0].Hash:
				// First commit
				return []core.FileChange{
					{Path: "first.go", Status: "M", Additions: 10, Deletions: 5},
					{Path: "shared.go", Status: "A", Additions: 20, Deletions: 0},
				}, nil
			case commits[1].Hash:
				// Second commit - different files to distinguish from first commit
				return []core.FileChange{
					{Path: "second.go", Status: "M", Additions: 15, Deletions: 3},
					{Path: "another.go", Status: "A", Additions: 8, Deletions: 2},
				}, nil
			default:
				// Third commit or other
				return []core.FileChange{
					{Path: "third.go", Status: "M", Additions: 5, Deletions: 1},
				}, nil
			}
		}
		// Multi-commit range - return combined files
		return []core.FileChange{
			{Path: "combined1.go", Status: "M", Additions: 25, Deletions: 8},
			{Path: "combined2.go", Status: "A", Additions: 28, Deletions: 2},
		}, nil
	}

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(mockFetchFileChanges),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size (wide enough to show split view with preview)
	runner.Send(tea.WindowSizeMsg{Width: 160, Height: 24})
	runner.AssertGolden("visual_mode_toggle_out/1_initial.golden")

	// Enter visual mode with 'v'
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	runner.AssertGolden("visual_mode_toggle_out/2_visual_mode.golden")

	// Select 2 commits by pressing 'j'
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	// Verify files of both commits (combined) are visible in preview
	runner.AssertGolden("visual_mode_toggle_out/3_selection_extended.golden")

	// Exit visual mode by pressing 'v' again (toggle out)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	// Verify: back in normal mode with files for single commit at position 1
	runner.AssertGolden("visual_mode_toggle_out/4_after_toggle_out.golden")

	// Quit
	runner.Quit()
}
