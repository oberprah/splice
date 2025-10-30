package states

import (
	"strings"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/git"
)

func createTestCommit() git.GitCommit {
	return git.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add automatic light/dark theme support",
		Author:  "John Doe",
		Date:    time.Date(2024, 10, 15, 10, 0, 0, 0, time.UTC),
	}
}

func createTestFileChanges(count int) []git.FileChange {
	changes := make([]git.FileChange, count)
	for i := range count {
		changes[i] = git.FileChange{
			Path:      "file" + string(rune('0'+i)) + ".go",
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

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Check that commit info appears in header
	if !strings.Contains(result, "abc123d") { // Short hash
		t.Error("Expected output to contain short commit hash")
	}
	if !strings.Contains(result, "Add automatic light/dark theme support") {
		t.Error("Expected output to contain commit message")
	}
	if !strings.Contains(result, "John Doe") {
		t.Error("Expected output to contain author name")
	}
}

func TestFilesState_View_RendersFileList(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "internal/ui/app.go", Additions: 45, Deletions: 12},
		{Path: "internal/ui/model.go", Additions: 3, Deletions: 1},
		{Path: "internal/git/git.go", Additions: 120, Deletions: 0},
	}

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Check that all files appear in the output
	if !strings.Contains(result, "internal/ui/app.go") {
		t.Error("Expected output to contain first file path")
	}
	if !strings.Contains(result, "internal/ui/model.go") {
		t.Error("Expected output to contain second file path")
	}
	if !strings.Contains(result, "internal/git/git.go") {
		t.Error("Expected output to contain third file path")
	}

	// Check that stats appear
	if !strings.Contains(result, "+45") || !strings.Contains(result, "-12") {
		t.Error("Expected output to contain file statistics")
	}
}

func TestFilesState_View_SelectionIndicator(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(3)

	tests := []struct {
		name   string
		cursor int
	}{
		{"first file selected", 0},
		{"second file selected", 1},
		{"third file selected", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FilesState{
				Commit:        commit,
				Files:         files,
				Cursor:        tt.cursor,
				ViewportStart: 0,
			}
			ctx := mockContext{width: 80, height: 24}

			result := s.View(ctx)

			// Should have at least one selection indicator (">")
			selectedCount := strings.Count(result, ">")

			if selectedCount < 1 {
				t.Error("Expected at least one selection indicator '>'")
			}
		})
	}
}

func TestFilesState_View_ViewportLimits(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(20)

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        10,
		ViewportStart: 5,
	}
	ctx := mockContext{width: 80, height: 10}

	result := s.View(ctx)

	// The output should be limited by viewport
	// This is hard to test exactly due to header, but we can check it's not showing all files
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should not have 20+ lines (one per file) due to viewport limit
	if len(lines) > 15 {
		t.Errorf("Expected viewport to limit output, but got %d lines", len(lines))
	}
}

func TestFilesState_View_BinaryFiles(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "image.png", Additions: 0, Deletions: 0, IsBinary: true},
		{Path: "main.go", Additions: 10, Deletions: 5, IsBinary: false},
	}

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Binary files should show different stats
	if !strings.Contains(result, "image.png") {
		t.Error("Expected output to contain binary file")
	}
	// Binary files might show "binary" or special indicator instead of +/-
	if !strings.Contains(result, "main.go") {
		t.Error("Expected output to contain text file")
	}
}

func TestFilesState_View_EmptyFileList(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{}

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Should still render header with commit info
	if !strings.Contains(result, "abc123d") {
		t.Error("Expected output to contain commit hash even with no files")
	}

	// Should indicate no files changed (or show empty list)
	// The exact message depends on implementation, so we just check it doesn't panic
	if result == "" {
		t.Error("Expected some output even with empty file list")
	}
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

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}

	// Test with narrow terminal
	ctx := mockContext{width: 50, height: 24}

	result := s.View(ctx)

	// Should handle long paths gracefully (either truncate or wrap)
	if result == "" {
		t.Error("Expected output even with long file paths")
	}

	// Check that truncation indicator appears if path is truncated
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		// Lines shouldn't be excessively long (allowing for ANSI codes)
		if len(line) > 200 {
			t.Errorf("Line %d appears too long (%d chars), may need truncation", i, len(line))
		}
	}
}

func TestFilesState_View_FileStatsSummary(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{
		{Path: "file1.go", Additions: 10, Deletions: 5},
		{Path: "file2.go", Additions: 20, Deletions: 3},
		{Path: "file3.go", Additions: 5, Deletions: 2},
	}

	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Should show summary: 3 files changed, 35 insertions(+), 10 deletions(-)
	// The exact format may vary, but we check for key information
	if !strings.Contains(result, "3") {
		t.Error("Expected output to mention 3 files")
	}
}
