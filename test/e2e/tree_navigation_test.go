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

// TestTreeNavigation_BasicStructure tests that files are displayed in a tree structure
// with proper box-drawing characters, folders first, and alphabetical sorting
// Verifies FR1 (tree structure), FR2 (item ordering), and FR7 (visual selection)
func TestTreeNavigation_BasicStructure(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add feature with nested files",
	})

	// Create files in multiple folders at different depths
	mockFiles := []core.FileChange{
		{Path: "src/components/App.tsx", Status: "M", Additions: 17, Deletions: 13},
		{Path: "src/utils/helper.ts", Status: "A", Additions: 42, Deletions: 0},
		{Path: "src/index.ts", Status: "M", Additions: 5, Deletions: 2},
		{Path: "old/deprecated.ts", Status: "D", Additions: 0, Deletions: 15},
		{Path: "README.md", Status: "M", Additions: 3, Deletions: 1},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for loading
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.AssertGolden("tree_navigation/basic_structure_log.golden")

	// Press Enter to navigate to FilesState
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show tree structure with proper box-drawing characters
	// - src/ folder at top (folders first)
	// - old/ folder next
	// - README.md at bottom (files after folders)
	// - Items within folders alphabetically sorted
	runner.AssertGolden("tree_navigation/basic_structure_files.golden")

	// Quit
	runner.Quit()
}

// TestTreeNavigation_UpDown tests cursor navigation through the tree
// Verifies FR6 (navigation controls) and FR7 (visual selection)
func TestTreeNavigation_UpDown(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Update multiple files",
	})

	mockFiles := []core.FileChange{
		{Path: "src/app.ts", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.ts", Status: "A", Additions: 20, Deletions: 0},
		{Path: "test/app.test.ts", Status: "A", Additions: 30, Deletions: 0},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Initial state: cursor on first item (src/ folder)
	runner.AssertGolden("tree_navigation/updown_initial.golden")

	// Navigate down with j - cursor moves to next item
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("tree_navigation/updown_after_j.golden")

	// Navigate down again
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	runner.AssertGolden("tree_navigation/updown_after_jj.golden")

	// Navigate up with k
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	runner.AssertGolden("tree_navigation/updown_after_jjk.golden")

	// Jump to bottom with G
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	runner.AssertGolden("tree_navigation/updown_at_bottom.golden")

	// Jump to top with g
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	runner.AssertGolden("tree_navigation/updown_back_to_top.golden")

	runner.Quit()
}

// TestTreeNavigation_CollapseExpand tests folder collapse and expand functionality
// Verifies FR4 (default expanded), FR5 (collapse/expand), and FR6 (navigation controls)
func TestTreeNavigation_CollapseExpand(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add nested components",
	})

	mockFiles := []core.FileChange{
		{Path: "src/components/Header.tsx", Status: "A", Additions: 50, Deletions: 0},
		{Path: "src/components/Footer.tsx", Status: "A", Additions: 30, Deletions: 0},
		{Path: "src/utils/format.ts", Status: "A", Additions: 20, Deletions: 0},
		{Path: "src/index.ts", Status: "M", Additions: 5, Deletions: 2},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	// Set window size and wait for log view to render
	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	time.Sleep(50 * time.Millisecond) // Give time for initial render

	// Navigate to files view
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Initial: all folders expanded by default (FR4)
	runner.AssertGolden("tree_navigation/collapse_initial.golden")

	// Cursor on src/ folder - press Enter to collapse
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show "src/ +105 -2 (4 files)" and hide children
	runner.AssertGolden("tree_navigation/collapse_after_enter.golden")

	// Press Enter again to expand
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show children again
	runner.AssertGolden("tree_navigation/collapse_after_expand.golden")

	// Navigate down to src/ folder again, then test Space key
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}) // jump to top
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")}) // Space to toggle
	// Should collapse with Space (FR6)
	runner.AssertGolden("tree_navigation/collapse_with_space.golden")

	runner.Quit()
}

// TestTreeNavigation_ArrowKeys tests left/right arrow keys for expand/collapse
// Verifies FR6 (navigation controls - right expands, left collapses)
func TestTreeNavigation_ArrowKeys(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Reorganize code",
	})

	mockFiles := []core.FileChange{
		{Path: "lib/core/engine.ts", Status: "M", Additions: 40, Deletions: 10},
		{Path: "lib/utils/helper.ts", Status: "A", Additions: 25, Deletions: 0},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Initial: cursor on lib/ folder (expanded)
	runner.AssertGolden("tree_navigation/arrows_initial.golden")

	// Press left arrow to collapse
	runner.Send(tea.KeyMsg{Type: tea.KeyLeft})
	// lib/ folder should collapse
	runner.AssertGolden("tree_navigation/arrows_after_left.golden")

	// Press left arrow again (should be no-op when already collapsed)
	runner.Send(tea.KeyMsg{Type: tea.KeyLeft})
	// Should remain collapsed
	runner.AssertGolden("tree_navigation/arrows_after_left_noop.golden")

	// Press right arrow to expand
	runner.Send(tea.KeyMsg{Type: tea.KeyRight})
	// lib/ folder should expand
	runner.AssertGolden("tree_navigation/arrows_after_right.golden")

	// Press right arrow again (should be no-op when already expanded)
	runner.Send(tea.KeyMsg{Type: tea.KeyRight})
	// Should remain expanded
	runner.AssertGolden("tree_navigation/arrows_after_right_noop.golden")

	runner.Quit()
}

// TestTreeNavigation_CollapsedPaths tests that single-child folder chains are collapsed
// Verifies FR3 (path collapsing)
func TestTreeNavigation_CollapsedPaths(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add deeply nested file",
	})

	// Create a deep chain of single-child folders
	mockFiles := []core.FileChange{
		{Path: "src/components/nested/deep/VeryDeep.tsx", Status: "A", Additions: 100, Deletions: 0},
		{Path: "src/index.ts", Status: "M", Additions: 2, Deletions: 1},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show "src/components/nested/deep/" as a single collapsed path
	runner.AssertGolden("tree_navigation/collapsed_path_initial.golden")

	// Cursor is on src/ - collapse the entire src folder
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter}) // collapse src/
	// Should show "src/ +102 -1 (2 files)" hiding all nested content
	runner.AssertGolden("tree_navigation/collapsed_path_collapsed.golden")

	// Expand src/ again
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should reveal the collapsed path and all files
	runner.AssertGolden("tree_navigation/collapsed_path_expanded.golden")

	runner.Quit()
}

// TestTreeNavigation_EnterFile tests entering a file to view its diff
// Verifies FR6 (Enter on file opens diff) and back navigation
func TestTreeNavigation_EnterFile(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Update components",
	})

	mockFiles := []core.FileChange{
		{Path: "src/App.tsx", Status: "M", Additions: 25, Deletions: 10},
		{Path: "src/utils/helper.ts", Status: "A", Additions: 15, Deletions: 0},
	}

	// Mock diff content for the file
	mockDiffResult := &core.FullFileDiffResult{
		OldContent: "import React from 'react'\n\nfunction App() {\n  return (\n    <div>",
		NewContent: "import React from 'react'\nimport { useState } from 'react'\n\nfunction App() {\n  const [count, setCount] = useState(0)\n  return (\n    <div>",
		DiffOutput: "@@ -1,5 +1,7 @@\n import React from 'react'\n+import { useState } from 'react'\n \n function App() {\n+  const [count, setCount] = useState(0)\n   return (\n     <div>",
		OldPath:    "src/App.tsx",
		NewPath:    "src/App.tsx",
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithFetchFullFileDiff(testutils.MockFetchFullFileDiff(mockDiffResult, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Initial files view
	runner.AssertGolden("tree_navigation/enter_file_initial.golden")

	// Navigate to a file (src/App.tsx)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // move to src/ folder
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}) // move to src/App.tsx file

	// Press Enter on file to view diff
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show diff view
	runner.AssertGolden("tree_navigation/enter_file_diff.golden")

	// Press 'q' to go back to tree
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	// Should return to files view with cursor at same position
	runner.AssertGolden("tree_navigation/enter_file_back_to_tree.golden")

	runner.Quit()
}

// TestTreeNavigation_MixedStatuses tests tree rendering with different file statuses
// Verifies that all file status types (M, A, D, R) are rendered correctly in tree
func TestTreeNavigation_MixedStatuses(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Refactor and cleanup",
	})

	mockFiles := []core.FileChange{
		{Path: "src/new.ts", Status: "A", Additions: 50, Deletions: 0},
		{Path: "src/modified.ts", Status: "M", Additions: 20, Deletions: 10},
		{Path: "src/deleted.ts", Status: "D", Additions: 0, Deletions: 30},
		{Path: "src/renamed.ts", Status: "R", Additions: 5, Deletions: 3},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show all status types in tree with correct colors/indicators
	runner.AssertGolden("tree_navigation/mixed_statuses.golden")

	runner.Quit()
}

// TestTreeNavigation_BinaryFiles tests that binary files are displayed correctly in tree
func TestTreeNavigation_BinaryFiles(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add images",
	})

	mockFiles := []core.FileChange{
		{Path: "assets/logo.png", Status: "A", Additions: 0, Deletions: 0, IsBinary: true},
		{Path: "assets/icon.svg", Status: "M", Additions: 5, Deletions: 2, IsBinary: false},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show binary file with "(binary)" marker
	runner.AssertGolden("tree_navigation/binary_files.golden")

	runner.Quit()
}

// TestTreeNavigation_SingleFileFolder tests that single-file folders display correctly
// (not collapsed, since FR3 only collapses single-child folder chains)
func TestTreeNavigation_SingleFileFolder(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add utility",
	})

	mockFiles := []core.FileChange{
		{Path: "utils/helper.ts", Status: "A", Additions: 30, Deletions: 0},
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show "utils/" folder separately (not collapsed)
	// with the file as a child
	runner.AssertGolden("tree_navigation/single_file_folder.golden")

	runner.Quit()
}

// TestTreeNavigation_EmptyTree tests edge case of no files
func TestTreeNavigation_EmptyTree(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Empty commit",
	})

	mockFiles := []core.FileChange{}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 80, Height: 24})
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	// Should show "0 files · +0 -0" with empty tree
	runner.AssertGolden("tree_navigation/empty_tree.golden")

	runner.Quit()
}
