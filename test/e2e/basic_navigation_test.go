package main

import (
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestBasicNavigation tests the full user journey: load commits, navigate with j/k/g/G keys, and quit
func TestBasicNavigation(t *testing.T) {
	// Create mock commits
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Fix authentication bug",
		"Add dark mode toggle",
		"Update dependencies",
		"Refactor git parsing",
		"Initial commit",
	})

	// Fixed time for deterministic date formatting (commits are exactly 1 year old)
	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	// Create model with mocked commits and file changes
	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges([]core.FileChange{}, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set initial window size and wait for loading to complete
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("basic_navigation/1_initial.golden")

	// Navigate down twice with j key
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("basic_navigation/2_after_down_once.golden")

	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("basic_navigation/3_after_down_twice.golden")

	// Navigate up with k key
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	runner.AssertGolden("basic_navigation/4_after_up.golden")

	// Jump to bottom with G
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	runner.AssertGolden("basic_navigation/5_at_bottom.golden")

	// Jump to top with g
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	runner.AssertGolden("basic_navigation/6_back_to_top.golden")

	// Quit
	runner.Quit()
}
