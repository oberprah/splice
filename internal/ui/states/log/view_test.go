package log

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/components"
)

func assertLogViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), ""+filename, *update)
}

// createLogStateWithGraph creates a LogState with computed GraphLayout
func createLogStateWithGraph(commits []git.GitCommit) State {
	// Convert to graph.Commits
	graphCommits := make([]graph.Commit, len(commits))
	for i, commit := range commits {
		graphCommits[i] = graph.Commit{
			Hash:    commit.Hash,
			Parents: commit.ParentHashes,
		}
	}

	// Compute layout
	layout := graph.ComputeLayout(graphCommits)

	return State{
		Commits:     commits,
		GraphLayout: layout,
	}
}

func TestLogState_View_RendersCommits(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{"def456"}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
		{Hash: "def456", ParentHashes: []string{}, Message: "Second commit", Body: "", Author: "Bob", Date: fixedTime.Add(time.Hour)},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "renders_commits.golden")
}

func TestLogState_View_SelectionIndicator(t *testing.T) {
	commits := createTestCommits(3)

	tests := []struct {
		name       string
		cursor     int
		goldenFile string
	}{
		{"first commit selected", 0, "selection_first.golden"},
		{"second commit selected", 1, "selection_second.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createLogStateWithGraph(commits)
			s.Cursor = tt.cursor
			s.ViewportStart = 0
			s.Preview = PreviewNone{}
			ctx := mockContext{width: 80, height: 24}

			output := s.View(ctx)

			assertLogViewGolden(t, output.(*components.ViewBuilder), tt.goldenFile)
		})
	}
}

func TestLogState_View_ViewportLimits(t *testing.T) {
	commits := createTestCommits(20)

	s := createLogStateWithGraph(commits)
	s.Cursor = 10
	s.ViewportStart = 5
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 10}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "viewport_limits.golden")
}

func TestLogState_View_EmptyViewport(t *testing.T) {
	commits := createTestCommits(5)

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 10 // Beyond end
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 10}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "empty_viewport.golden")
}

func TestLogState_View_LineTruncation(t *testing.T) {
	// Force no color output for consistent golden file testing
	lipgloss.SetColorProfile(termenv.Ascii)
	defer lipgloss.SetColorProfile(termenv.TrueColor) // Reset after test

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{
			Hash:         "abc123def456",
			ParentHashes: []string{},
			Message:      "This is a very long commit message that should be truncated when the terminal is narrow",
			Body:         "",
			Author:       "VeryLongAuthorNameThatShouldAlsoGetTruncated",
			Date:         fixedTime,
		},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}

	ctx := mockContext{width: 40, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "line_truncation.golden")
}

func TestLogState_View_SplitView_WideTerminal(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{"def456"}, Message: "First commit", Body: "This is the body", Author: "Alice", Date: fixedTime},
		{Hash: "def456", ParentHashes: []string{}, Message: "Second commit", Body: "", Author: "Bob", Date: fixedTime.Add(time.Hour)},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoaded{ForHash: "abc123", Files: files}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "split_view_wide.golden")
}

func TestLogState_View_SplitView_NarrowTerminal(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "This is the body", Author: "Alice", Date: fixedTime},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoaded{ForHash: "abc123", Files: files}

	ctx := mockContext{width: 100, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "split_view_narrow.golden")
}

func TestLogState_View_SplitView_PreviewLoading(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewLoading{ForHash: "abc123"}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "split_view_loading.golden")
}

func TestLogState_View_SplitView_PreviewError(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	commits := []git.GitCommit{
		{Hash: "abc123", ParentHashes: []string{}, Message: "First commit", Body: "", Author: "Alice", Date: fixedTime},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewError{ForHash: "abc123", Err: nil}

	ctx := mockContext{width: 160, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "split_view_error.golden")
}

func TestLogState_View_MergeBranchGraph(t *testing.T) {
	// Create a simple feature branch merge scenario:
	// E (merge commit) <- merges B and D
	// D (feature branch)
	// C (feature branch)
	// B (main branch)
	// A (initial commit)
	//
	// Expected graph:
	// ├─╮  E: Merge feature-x
	// │ ├  D: Add feature X part 2
	// │ ├  C: Add feature X part 1
	// ├ │  B: Fix bug on main
	// ├─╯  A: Initial commit

	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commits := []git.GitCommit{
		{
			Hash:         "eeeeeeee",
			ParentHashes: []string{"bbbbbbbb", "dddddddd"}, // Merge commit
			Message:      "Merge feature-x",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime,
		},
		{
			Hash:         "dddddddd",
			ParentHashes: []string{"cccccccc"},
			Message:      "Add feature X part 2",
			Body:         "",
			Author:       "Bob",
			Date:         fixedTime.Add(time.Hour),
		},
		{
			Hash:         "cccccccc",
			ParentHashes: []string{"aaaaaaaa"},
			Message:      "Add feature X part 1",
			Body:         "",
			Author:       "Bob",
			Date:         fixedTime.Add(2 * time.Hour),
		},
		{
			Hash:         "bbbbbbbb",
			ParentHashes: []string{"aaaaaaaa"},
			Message:      "Fix bug on main",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime.Add(3 * time.Hour),
		},
		{
			Hash:         "aaaaaaaa",
			ParentHashes: []string{},
			Message:      "Initial commit",
			Body:         "",
			Author:       "Alice",
			Date:         fixedTime.Add(4 * time.Hour),
		},
	}

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 0
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLogViewGolden(t, output.(*components.ViewBuilder), "merge_branch_graph.golden")
}
