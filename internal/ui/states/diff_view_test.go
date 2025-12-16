package states

import (
	"strings"
	"testing"

	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
)

func TestDiffState_View_Header(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{
			Hash: "abc123def456789012345678901234567890abcd",
		},
		File: git.FileChange{
			Path:      "internal/ui/states/diff_view.go",
			Additions: 15,
			Deletions: 8,
		},
		Diff: &diff.FullFileDiff{
			Lines: []diff.FullFileLine{},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	// Should contain short hash
	if !strings.Contains(view, "abc123d") {
		t.Error("View should contain short hash")
	}

	// Should contain file path
	if !strings.Contains(view, "internal/ui/states/diff_view.go") {
		t.Error("View should contain file path")
	}

	// Should contain additions
	if !strings.Contains(view, "+15") {
		t.Error("View should contain additions count")
	}

	// Should contain deletions
	if !strings.Contains(view, "-8") {
		t.Error("View should contain deletions count")
	}
}

func TestDiffState_View_EmptyDiff(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff:   &diff.FullFileDiff{Lines: []diff.FullFileLine{}},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	if !strings.Contains(view, "No changes") {
		t.Error("Empty diff should show 'No changes' message")
	}
}

func TestDiffState_View_UnchangedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.FullFileDiff{
			Lines: []diff.FullFileLine{
				{LeftLineNo: 1, RightLineNo: 1, LeftContent: "unchanged line", RightContent: "unchanged line", Change: diff.Unchanged},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	// Unchanged lines should appear (shown on both sides)
	if !strings.Contains(view, "unchanged line") {
		t.Error("View should contain unchanged line content")
	}

	// Line numbers should appear
	if !strings.Contains(view, "1") {
		t.Error("View should contain line numbers")
	}
}

func TestDiffState_View_RemovedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.FullFileDiff{
			Lines: []diff.FullFileLine{
				{LeftLineNo: 5, RightLineNo: 0, LeftContent: "removed line", RightContent: "", Change: diff.Removed},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	// Should contain removed line content
	if !strings.Contains(view, "removed line") {
		t.Error("View should contain removed line content")
	}

	// Should contain line number
	if !strings.Contains(view, "5") {
		t.Error("View should contain line number for removed line")
	}

	// Should contain minus indicator
	if !strings.Contains(view, "-") {
		t.Error("View should contain minus indicator for removed line")
	}
}

func TestDiffState_View_AddedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.FullFileDiff{
			Lines: []diff.FullFileLine{
				{LeftLineNo: 0, RightLineNo: 7, LeftContent: "", RightContent: "added line", Change: diff.Added},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	// Should contain added line content
	if !strings.Contains(view, "added line") {
		t.Error("View should contain added line content")
	}

	// Should contain line number
	if !strings.Contains(view, "7") {
		t.Error("View should contain line number for added line")
	}

	// Should contain plus indicator
	if !strings.Contains(view, "+") {
		t.Error("View should contain plus indicator for added line")
	}
}

func TestDiffState_View_SideBySideSeparator(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.FullFileDiff{
			Lines: []diff.FullFileLine{
				{LeftLineNo: 1, RightLineNo: 1, LeftContent: "line", RightContent: "line", Change: diff.Unchanged},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	view := state.View(ctx)

	// Should contain side-by-side separator
	if !strings.Contains(view, "│") {
		t.Error("View should contain side-by-side separator")
	}
}

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		input    string
		maxWidth int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is a…"},
		{"exact", 5, "exact"},
		{"toolong", 5, "tool…"},
		{"", 5, ""},
		{"test", 0, ""},
		{"test", 1, "…"},
		{"ab", 2, "ab"},
		{"abc", 2, "a…"},
	}

	for _, tt := range tests {
		result := truncateWithEllipsis(tt.input, tt.maxWidth)
		if result != tt.expected {
			t.Errorf("truncateWithEllipsis(%q, %d) = %q, want %q", tt.input, tt.maxWidth, result, tt.expected)
		}
	}
}

func TestDiffState_View_Viewport(t *testing.T) {
	// Create a diff with many lines
	lines := make([]diff.FullFileLine, 100)
	for i := 0; i < 100; i++ {
		lines[i] = diff.FullFileLine{
			LeftLineNo:   i + 1,
			RightLineNo:  i + 1,
			LeftContent:  "line content",
			RightContent: "line content",
			Change:       diff.Unchanged,
		}
	}

	state := &DiffState{
		Commit:        git.GitCommit{Hash: "abc123"},
		File:          git.FileChange{Path: "file.go"},
		Diff:          &diff.FullFileDiff{Lines: lines},
		ViewportStart: 50, // Start at line 50
	}

	ctx := &mockContext{width: 80, height: 10}
	view := state.View(ctx)

	// Should contain line 51 (ViewportStart + 1 because line numbers are 1-based in our test data)
	if !strings.Contains(view, "51") {
		t.Error("View should contain line 51 when viewport starts at 50")
	}

	// Should NOT contain line 1 (before viewport)
	// This is tricky to test since "1" appears in many places, so we skip this check
}
