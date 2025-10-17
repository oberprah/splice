package main

import (
	"fmt"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestBasicNavigation tests the full user journey: load commits, navigate, and quit
func TestBasicNavigation(t *testing.T) {
	// Create mock commits
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Fix authentication bug",
		"Add dark mode toggle",
		"Update dependencies",
		"Refactor git parsing",
		"Initial commit",
	})

	// Create model with mocked commits
	m := ui.NewModel(ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)))

	// Create test model with fixed terminal size
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Wait for loading to complete
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate down (j key)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Navigate up (k key)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})

	// Jump to bottom (G)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})

	// Jump to top (g)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	// Verify final model state
	fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	// Verify output matches golden file
	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestViewportScrolling tests that viewport adjusts when navigating beyond visible area
func TestViewportScrolling(t *testing.T) {
	// Create more commits than can fit on screen
	commits := testutils.CreateTestCommits(50)

	m := ui.NewModel(ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 10))

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 10})

	// Navigate down many times to trigger scrolling
	for range 15 {
		tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}

	// Jump to bottom to test large jump
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})

	// Jump back to top
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestErrorState tests error handling when git command fails
func TestErrorState(t *testing.T) {
	// Create model with error-returning mock
	m := ui.NewModel(ui.WithFetchCommits(
		testutils.MockFetchCommits(nil, fmt.Errorf("not a git repository")),
	))

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Try to quit from error state
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestEmptyRepository tests handling of empty commit list
func TestEmptyRepository(t *testing.T) {
	// Create model with empty commits (returns empty slice)
	emptyCommits := []git.GitCommit{}
	m := ui.NewModel(ui.WithFetchCommits(testutils.MockFetchCommits(emptyCommits, nil)))

	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Try to quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestWindowResize tests that the UI adapts to terminal size changes
func TestWindowResize(t *testing.T) {
	commits := testutils.CreateTestCommits(10)

	m := ui.NewModel(ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	// Send initial size
	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate a bit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Resize window
	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Navigate again
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Resize to smaller
	tm.Send(tea.WindowSizeMsg{Width: 60, Height: 15})

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}

// TestQuitWithCtrlC tests quitting with Ctrl+C
func TestQuitWithCtrlC(t *testing.T) {
	commits := testutils.CreateTestCommits(5)

	m := ui.NewModel(ui.WithFetchCommits(testutils.MockFetchCommits(commits, nil)))
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Navigate
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

	// Quit with Ctrl+C
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})

	fm := tm.FinalModel(t, teatest.WithFinalTimeout(3*time.Second))
	if _, ok := fm.(ui.Model); !ok {
		t.Fatalf("expected final model to be ui.Model, got %T", fm)
	}

	outReader := tm.FinalOutput(t, teatest.WithFinalTimeout(3*time.Second))
	out, err := io.ReadAll(outReader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	teatest.RequireEqualOutput(t, out)
}
