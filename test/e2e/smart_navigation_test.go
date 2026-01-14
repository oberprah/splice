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

// TestSmartNavigation_JumpsBetweenChanges tests that n/p keys navigate between changes
func TestSmartNavigation_JumpsBetweenChanges(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Update components",
	})

	// Use root-level file to simplify tree navigation
	mockFiles := []core.FileChange{
		{Path: "app.go", Status: "M", Additions: 10, Deletions: 5},
	}

	// Create diff with multiple change blocks separated by unchanged lines
	// Structure: 5 unchanged, 3 changed, 5 unchanged, 3 changed, 5 unchanged
	mockDiffResult := &core.FullFileDiffResult{
		OldContent: `line 1
line 2
line 3
line 4
line 5
old line 6
old line 7
old line 8
line 9
line 10
line 11
line 12
line 13
old line 14
old line 15
old line 16
line 17
line 18
line 19
line 20
line 21`,
		NewContent: `line 1
line 2
line 3
line 4
line 5
new line 6
new line 7
new line 8
line 9
line 10
line 11
line 12
line 13
new line 14
new line 15
new line 16
line 17
line 18
line 19
line 20
line 21`,
		DiffOutput: `@@ -3,17 +3,17 @@ line 1
 line 3
 line 4
 line 5
-old line 6
-old line 7
-old line 8
+new line 6
+new line 7
+new line 8
 line 9
 line 10
 line 11
 line 12
 line 13
-old line 14
-old line 15
-old line 16
+new line 14
+new line 15
+new line 16
 line 17
 line 18`,
		OldPath: "app.go",
		NewPath: "app.go",
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

	// Set window size and wait for log view
	runner.Send(tea.WindowSizeMsg{Width: 100, Height: 30})
	runner.WaitForContent("Update components")

	// Navigate to files view
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("app.go")

	// Navigate to diff view (file is already selected since it's root level)
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("new line")

	// Should start at first change block
	runner.AssertGolden("smart_navigation/01_initial_diff.golden")

	// Press n to jump to next change
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	runner.AssertGolden("smart_navigation/02_after_n.golden")

	// Press p to go back to first change
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	runner.AssertGolden("smart_navigation/03_after_p.golden")

	runner.Quit()
}

// TestSmartNavigation_ScrollsThroughMultiScreenChange tests that n scrolls through large changes
func TestSmartNavigation_ScrollsThroughMultiScreenChange(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Add large feature",
	})

	// Use root-level file to simplify tree navigation
	mockFiles := []core.FileChange{
		{Path: "large.go", Status: "A", Additions: 40, Deletions: 0},
	}

	// Create a diff with a large change block (40 lines)
	// This exceeds typical viewport height, so n should scroll first
	oldContent := `line 1
line 2
line 3`

	var newContent string
	newContent = `line 1
line 2
line 3
`
	// Add 40 new lines
	for i := 1; i <= 40; i++ {
		newContent += "new added line number " + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "\n"
	}
	newContent += `line 4
line 5`

	diffOutput := `@@ -1,3 +1,45 @@ line 1
 line 1
 line 2
 line 3
+new added line number 01
+new added line number 02
+new added line number 03
+new added line number 04
+new added line number 05
+new added line number 06
+new added line number 07
+new added line number 08
+new added line number 09
+new added line number 10
+new added line number 11
+new added line number 12
+new added line number 13
+new added line number 14
+new added line number 15
+new added line number 16
+new added line number 17
+new added line number 18
+new added line number 19
+new added line number 20
+new added line number 21
+new added line number 22
+new added line number 23
+new added line number 24
+new added line number 25
+new added line number 26
+new added line number 27
+new added line number 28
+new added line number 29
+new added line number 30
+new added line number 31
+new added line number 32
+new added line number 33
+new added line number 34
+new added line number 35
+new added line number 36
+new added line number 37
+new added line number 38
+new added line number 39
+new added line number 40
+line 4
+line 5`

	mockDiffResult := &core.FullFileDiffResult{
		OldContent: oldContent,
		NewContent: newContent,
		DiffOutput: diffOutput,
		OldPath:    "large.go",
		NewPath:    "large.go",
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

	// Use small viewport to ensure scrolling is needed
	runner.Send(tea.WindowSizeMsg{Width: 100, Height: 20})
	runner.WaitForContent("Add large feature")

	// Navigate to files view
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("large.go")

	// Navigate to diff view (file is already selected since it's root level)
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("added line")

	// Should start at first change block
	runner.AssertGolden("smart_navigation/multi_screen_01_initial.golden")

	// Press n - should scroll down (not jump) since change is multi-screen
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	runner.AssertGolden("smart_navigation/multi_screen_02_after_n.golden")

	// Press n again - should continue scrolling
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	runner.AssertGolden("smart_navigation/multi_screen_03_after_n2.golden")

	runner.Quit()
}

// TestFileNavigation_NextAndPrevious tests ] and [ for file navigation
func TestFileNavigation_NextAndPrevious(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Update multiple files",
	})

	// Use root-level files to simplify tree navigation
	mockFiles := []core.FileChange{
		{Path: "first.go", Status: "M", Additions: 5, Deletions: 3},
		{Path: "second.go", Status: "A", Additions: 10, Deletions: 0},
		{Path: "third.go", Status: "M", Additions: 2, Deletions: 1},
	}

	// Create different diff results for each file
	mockDiffFunc := func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
		switch change.Path {
		case "first.go":
			return &core.FullFileDiffResult{
				OldContent: "old content first",
				NewContent: "new content first",
				DiffOutput: "@@ -1 +1 @@\n-old content first\n+new content first",
				OldPath:    "first.go",
				NewPath:    "first.go",
			}, nil
		case "second.go":
			return &core.FullFileDiffResult{
				OldContent: "",
				NewContent: "new file second content",
				DiffOutput: "@@ -0,0 +1 @@\n+new file second content",
				OldPath:    "",
				NewPath:    "second.go",
			}, nil
		case "third.go":
			return &core.FullFileDiffResult{
				OldContent: "old third",
				NewContent: "new third",
				DiffOutput: "@@ -1 +1 @@\n-old third\n+new third",
				OldPath:    "third.go",
				NewPath:    "third.go",
			}, nil
		default:
			return &core.FullFileDiffResult{}, nil
		}
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithFetchFullFileDiff(mockDiffFunc),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 100, Height: 24})
	runner.WaitForContent("Update multiple files")

	// Navigate to files view
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("first.go")

	// Open first file diff (first file is already selected)
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("content first")
	runner.AssertGolden("smart_navigation/file_nav_01_first.golden")

	// Press ] to go to next file (second.go)
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	runner.WaitForContent("second content")
	runner.AssertGolden("smart_navigation/file_nav_02_second.golden")

	// Press ] again to go to third file
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	runner.WaitForContent("new third")
	runner.AssertGolden("smart_navigation/file_nav_03_third.golden")

	// Press ] at last file - should stay in place
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	// Still showing third file
	runner.AssertGolden("smart_navigation/file_nav_04_third_boundary.golden")

	// Press [ to go back to second file
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	runner.WaitForContent("second content")
	runner.AssertGolden("smart_navigation/file_nav_05_back_to_second.golden")

	// Press [ again to go to first file
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	runner.WaitForContent("content first")
	runner.AssertGolden("smart_navigation/file_nav_06_back_to_first.golden")

	// Press [ at first file - should stay in place
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")})
	// Still showing first file
	runner.AssertGolden("smart_navigation/file_nav_07_first_boundary.golden")

	runner.Quit()
}

// TestSmartNavigation_CrossFileNavigation tests n/p crossing file boundaries
func TestSmartNavigation_CrossFileNavigation(t *testing.T) {
	commits := testutils.CreateTestCommitsWithMessages([]string{
		"Multi-file change",
	})

	mockFiles := []core.FileChange{
		{Path: "alpha.go", Status: "M", Additions: 1, Deletions: 1},
		{Path: "beta.go", Status: "M", Additions: 1, Deletions: 1},
	}

	// Create diff with single change in each file
	mockDiffFunc := func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
		switch change.Path {
		case "alpha.go":
			return &core.FullFileDiffResult{
				OldContent: "line 1\nold alpha\nline 3",
				NewContent: "line 1\nnew alpha\nline 3",
				DiffOutput: "@@ -1,3 +1,3 @@\n line 1\n-old alpha\n+new alpha\n line 3",
				OldPath:    "alpha.go",
				NewPath:    "alpha.go",
			}, nil
		case "beta.go":
			return &core.FullFileDiffResult{
				OldContent: "line 1\nold beta\nline 3",
				NewContent: "line 1\nnew beta\nline 3",
				DiffOutput: "@@ -1,3 +1,3 @@\n line 1\n-old beta\n+new beta\n line 3",
				OldPath:    "beta.go",
				NewPath:    "beta.go",
			}, nil
		default:
			return &core.FullFileDiffResult{}, nil
		}
	}

	fixedNow := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	m := app.NewModel(
		app.WithInitialState(loading.State{}),
		app.WithFetchCommits(testutils.MockFetchCommits(commits, nil)),
		app.WithFetchFileChanges(testutils.MockFetchFileChanges(mockFiles, nil)),
		app.WithFetchFullFileDiff(mockDiffFunc),
		app.WithNow(func() time.Time { return fixedNow }),
	)

	runner := NewE2ETestRunner(t, m)

	runner.Send(tea.WindowSizeMsg{Width: 100, Height: 24})
	runner.WaitForContent("Multi-file change")

	// Navigate to files view
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("alpha")

	// Open first file diff
	runner.Send(tea.KeyMsg{Type: tea.KeyEnter})
	runner.WaitForContent("alpha")
	runner.AssertGolden("smart_navigation/cross_file_01_alpha.golden")

	// Press n - at last change in first file, should go to next file
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	runner.WaitForContent("beta")
	runner.AssertGolden("smart_navigation/cross_file_02_beta.golden")

	// Press p - at first change in second file, should go back to first file
	runner.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	runner.WaitForContent("alpha")
	runner.AssertGolden("smart_navigation/cross_file_03_back_to_alpha.golden")

	runner.Quit()
}
