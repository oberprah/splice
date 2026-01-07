package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/git"
	_ "github.com/oberprah/splice/internal/ui/states"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestViewportScrolling tests that viewport adjusts when navigating beyond visible area
// Uses 50 commits with a small 10-line screen to force scrolling behavior
func TestViewportScrolling(t *testing.T) {
	// Create more commits than can fit on screen (50 commits, 10-line screen)
	commits := testutils.CreateTestCommits(50)

	// Fixed time for deterministic date formatting (commits are exactly 1 year old)
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with mocked commits and file changes
	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges([]git.FileChange{}, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size (small height to force scrolling) and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 10})
	runner.AssertGolden("viewport_scrolling/1_initial.golden")

	// Navigate down 15 times to trigger viewport scrolling
	for i := 0; i < 15; i++ {
		runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	runner.AssertGolden("viewport_scrolling/2_after_scroll_down.golden")

	// Jump to bottom to test large jump
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	runner.AssertGolden("viewport_scrolling/3_at_bottom.golden")

	// Jump back to top
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	runner.AssertGolden("viewport_scrolling/4_back_to_top.golden")

	// Quit
	runner.Quit()
}
