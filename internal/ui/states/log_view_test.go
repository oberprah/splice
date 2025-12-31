package states

import (
	"flag"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/graph"
)

var update = flag.Bool("update", false, "update golden files")

func assertLogViewGolden(t *testing.T, output, filename string) {
	t.Helper()
	assertGolden(t, output, "log_view/"+filename, *update)
}

// createLogStateWithGraph creates a LogState with computed GraphLayout
func createLogStateWithGraph(commits []git.GitCommit) LogState {
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

	return LogState{
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

	assertLogViewGolden(t, output, "renders_commits.golden")
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

			assertLogViewGolden(t, output, tt.goldenFile)
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

	assertLogViewGolden(t, output, "viewport_limits.golden")
}

func TestLogState_View_EmptyViewport(t *testing.T) {
	commits := createTestCommits(5)

	s := createLogStateWithGraph(commits)
	s.Cursor = 0
	s.ViewportStart = 10 // Beyond end
	s.Preview = PreviewNone{}
	ctx := mockContext{width: 80, height: 10}

	output := s.View(ctx)

	assertLogViewGolden(t, output, "empty_viewport.golden")
}

func TestLogState_View_LineTruncation(t *testing.T) {
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

	assertLogViewGolden(t, output, "line_truncation.golden")
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

	assertLogViewGolden(t, output, "split_view_wide.golden")
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

	assertLogViewGolden(t, output, "split_view_narrow.golden")
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

	assertLogViewGolden(t, output, "split_view_loading.golden")
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

	assertLogViewGolden(t, output, "split_view_error.golden")
}

// Helper function tests - testing internal formatting logic
func TestLogState_formatFileEntry(t *testing.T) {
	s := LogState{}

	tests := []struct {
		name     string
		file     git.FileChange
		width    int
		expected string
	}{
		{
			name:     "modified file",
			file:     git.FileChange{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
			width:    80,
			expected: "M + 10 - 5  src/main.go",
		},
		{
			name:     "added file",
			file:     git.FileChange{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
			width:    80,
			expected: "A + 20 - 0  README.md",
		},
		{
			name:     "deleted file",
			file:     git.FileChange{Path: "old.txt", Status: "D", Additions: 0, Deletions: 15},
			width:    80,
			expected: "D + 0 - 15  old.txt",
		},
		{
			name:     "binary file",
			file:     git.FileChange{Path: "image.png", Status: "M", IsBinary: true},
			width:    80,
			expected: "M (binary)  image.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.formatFileEntry(tt.file, tt.width)

			if result != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// Helper function tests - testing internal text wrapping logic
func TestLogState_wrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected []string
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    20,
			expected: []string{"Hello world"},
		},
		{
			name:  "long text",
			text:  "This is a very long line that should be wrapped into multiple lines",
			width: 20,
			expected: []string{
				"This is a very long",
				"line that should be",
				"wrapped into",
				"multiple lines",
			},
		},
		{
			name:     "zero width",
			text:     "Hello",
			width:    0,
			expected: []string{"Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d\nExpected: %v\nGot: %v",
					len(tt.expected), len(result), tt.expected, result)
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d mismatch\nExpected: %q\nGot: %q", i, tt.expected[i], line)
				}
			}
		})
	}
}
