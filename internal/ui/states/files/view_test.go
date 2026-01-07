package files

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/testutils"
)

var update = flag.Bool("update", false, "update golden files")

// Per-file helper that adds subdirectory prefix
func assertFilesViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", filename)
	testutils.AssertGolden(t, output.String(), goldenPath, *update)
}

func createTestCommit() git.GitCommit {
	return git.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add automatic light/dark theme support",
		Body:    "",
		Author:  "John Doe",
		Date:    time.Date(2024, 10, 15, 10, 0, 0, 0, time.UTC),
	}
}

func createTestFileChanges(count int) []git.FileChange {
	changes := make([]git.FileChange, count)
	statuses := []string{"M", "A", "D", "M", "M"} // Cycle through some statuses
	for i := range count {
		status := "M"
		if i < len(statuses) {
			status = statuses[i]
		}
		changes[i] = git.FileChange{
			Path:      "file" + string(rune('0'+i)) + ".go",
			Status:    status,
			Additions: i * 10,
			Deletions: i * 2,
		}
	}
	return changes
}

func TestFilesState_View_RendersHeader(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "internal/ui/app.go", Additions: 45, Deletions: 12},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "renders_header.golden")
}

func TestFilesState_View_RendersFileList(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "internal/ui/app.go", Additions: 45, Deletions: 12},
		{Path: "internal/ui/model.go", Additions: 3, Deletions: 1},
		{Path: "internal/git/git.go", Additions: 120, Deletions: 0},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "renders_file_list.golden")
}

func TestFilesState_View_SelectionIndicator(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(3)

	tests := []struct {
		name       string
		cursor     int
		goldenFile string
	}{
		{"first file selected", 0, "selection_first.golden"},
		{"second file selected", 1, "selection_second.golden"},
		{"third file selected", 2, "selection_third.golden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{
				Range:         core.NewSingleCommitRange(commit),
				Files:         files,
				Cursor:        tt.cursor,
				ViewportStart: 0,
			}
			ctx := testutils.MockContext{W: 80, H: 24}

			output := s.View(ctx)

			assertFilesViewGolden(t, output.(*components.ViewBuilder), tt.goldenFile)
		})
	}
}

func TestFilesState_View_ViewportLimits(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(20)

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        10,
		ViewportStart: 5,
	}
	ctx := testutils.MockContext{W: 80, H: 10}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "viewport_limits.golden")
}

func TestFilesState_View_BinaryFiles(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "image.png", Additions: 0, Deletions: 0, IsBinary: true},
		{Path: "main.go", Additions: 10, Deletions: 5, IsBinary: false},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "binary_files.golden")
}

func TestFilesState_View_EmptyFileList(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "empty_file_list.golden")
}

func TestFilesState_View_LongFilePaths(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{
			Path:      "internal/ui/state/files/very/deeply/nested/directory/structure/with/a/very/long/filename.go",
			Additions: 10,
			Deletions: 5,
		},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}

	// Test with narrow terminal
	ctx := testutils.MockContext{W: 50, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "long_file_paths.golden")
}

func TestFilesState_View_FileStatsSummary(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "file1.go", Additions: 10, Deletions: 5},
		{Path: "file2.go", Additions: 20, Deletions: 3},
		{Path: "file3.go", Additions: 5, Deletions: 2},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "file_stats_summary.golden")
}

// Helper function test - testing internal logic
func TestFilesState_CalculateMaxStatWidth(t *testing.T) {
	tests := []struct {
		name         string
		files        []git.FileChange
		expectedAddW int
		expectedDelW int
	}{
		{
			name: "small numbers",
			files: []git.FileChange{
				{Path: "a.go", Additions: 1, Deletions: 2},
				{Path: "b.go", Additions: 9, Deletions: 8},
			},
			expectedAddW: 2, // +9 = 2 chars
			expectedDelW: 2, // -8 = 2 chars
		},
		{
			name: "large numbers",
			files: []git.FileChange{
				{Path: "a.go", Additions: 93, Deletions: 0},
				{Path: "b.go", Additions: 267, Deletions: 12},
				{Path: "c.go", Additions: 1234, Deletions: 567},
			},
			expectedAddW: 5, // +1234 = 5 chars (sign + 4 digits)
			expectedDelW: 4, // -567 = 4 chars (sign + 3 digits)
		},
		{
			name: "with binary files",
			files: []git.FileChange{
				{Path: "a.png", IsBinary: true},
				{Path: "b.go", Additions: 10, Deletions: 5},
			},
			expectedAddW: 3, // +10 = 3 chars
			expectedDelW: 2, // -5 = 2 chars
		},
		{
			name: "only zeros",
			files: []git.FileChange{
				{Path: "a.go", Additions: 0, Deletions: 0},
			},
			expectedAddW: 2, // +0 = 2 chars (minimum)
			expectedDelW: 2, // -0 = 2 chars (minimum)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addW, delW := components.CalculateMaxStatWidth(tt.files)

			if addW != tt.expectedAddW {
				t.Errorf("Addition width = %d, want %d", addW, tt.expectedAddW)
			}
			if delW != tt.expectedDelW {
				t.Errorf("Deletion width = %d, want %d", delW, tt.expectedDelW)
			}
		})
	}
}

func TestFilesState_View_StatusDisplay(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "modified.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "added.go", Status: "A", Additions: 50, Deletions: 0},
		{Path: "deleted.go", Status: "D", Additions: 0, Deletions: 30},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "status_display.golden")
}

func TestFilesState_View_DynamicAlignment(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "small.go", Status: "M", Additions: 5, Deletions: 2},
		{Path: "large.go", Status: "M", Additions: 1234, Deletions: 567},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "dynamic_alignment.golden")
}

func TestFilesState_View_WithRefs(t *testing.T) {
	commit := createTestCommit()
	// Add refs to the commit
	commit.Refs = []git.RefInfo{
		{Name: "main", Type: git.RefTypeBranch, IsHead: true},
		{Name: "origin/main", Type: git.RefTypeRemoteBranch},
	}
	files := []git.FileChange{
		{Path: "internal/ui/app.go", Additions: 45, Deletions: 12},
	}

	s := State{
		Range:         core.NewSingleCommitRange(commit),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertFilesViewGolden(t, output.(*components.ViewBuilder), "with_refs.golden")
}
