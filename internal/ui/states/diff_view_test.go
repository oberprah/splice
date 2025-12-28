package states

import (
	"strings"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/highlight"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Per-file helper that adds subdirectory prefix
func assertDiffViewGolden(t *testing.T, output, filename string) {
	t.Helper()
	assertGolden(t, output, "diff_view/"+filename, *update)
}

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
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "internal/ui/states/diff_view.go",
				Lines: []diff.AlignedLine{},
			},
			Right: diff.FileContent{
				Path:  "internal/ui/states/diff_view.go",
				Lines: []diff.AlignedLine{},
			},
			Alignments: []diff.Alignment{},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "header.golden")
}

func TestDiffState_View_EmptyDiff(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: []diff.AlignedLine{},
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: []diff.AlignedLine{},
			},
			Alignments: []diff.Alignment{},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "empty_diff.golden")
}

func TestDiffState_View_UnchangedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}}},
				},
			},
			Right: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{
					LeftIdx:  0,
					RightIdx: 0,
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "unchanged_lines.golden")
}

func TestDiffState_View_RemovedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
				},
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: []diff.AlignedLine{},
			},
			Alignments: []diff.Alignment{
				diff.RemovedAlignment{
					LeftIdx: 4, // 5th line (0-indexed)
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "removed_lines.golden")
}

func TestDiffState_View_AddedLines(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: []diff.AlignedLine{},
			},
			Right: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "added line"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.AddedAlignment{
					RightIdx: 6, // 7th line (0-indexed)
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "added_lines.golden")
}

func TestDiffState_View_SideBySideSeparator(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "line"}}},
				},
			},
			Right: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "line"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{
					LeftIdx:  0,
					RightIdx: 0,
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "side_by_side_separator.golden")
}

func TestDiffState_View_Viewport(t *testing.T) {
	// Create a diff with many lines
	leftLines := make([]diff.AlignedLine, 100)
	rightLines := make([]diff.AlignedLine, 100)
	alignments := make([]diff.Alignment, 100)
	for i := 0; i < 100; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "line content"}},
		}
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "line content"}},
		}
		alignments[i] = diff.UnchangedAlignment{
			LeftIdx:  i,
			RightIdx: i,
		}
	}

	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: leftLines,
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: rightLines,
			},
			Alignments: alignments,
		},
		ViewportStart: 50, // Start at line 50
	}

	ctx := &mockContext{width: 80, height: 10}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "viewport.golden")
}

func TestDiffState_View_SyntaxHighlightedTokens(t *testing.T) {
	state := &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{
						Tokens: []highlight.Token{
							{Type: chroma.Keyword, Value: "package"},
							{Type: chroma.Text, Value: " "},
							{Type: chroma.NameNamespace, Value: "main"},
						},
					},
				},
			},
			Right: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{
						Tokens: []highlight.Token{
							{Type: chroma.Keyword, Value: "package"},
							{Type: chroma.Text, Value: " "},
							{Type: chroma.NameNamespace, Value: "main"},
						},
					},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{
					LeftIdx:  0,
					RightIdx: 0,
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output, "syntax_highlighted_tokens.golden")
}

// Helper function test - testing internal logic
func TestRenderTokens_MultipleTokens(t *testing.T) {
	state := &DiffState{}
	tokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.NameFunction, Value: "main"},
		{Type: chroma.Punctuation, Value: "()"},
	}

	bgStyle := lipgloss.NewStyle() // Empty style for testing
	result := state.renderTokens(tokens, 100, bgStyle)

	// Should contain all token values
	if !strings.Contains(result, "func") {
		t.Error("Result should contain 'func'")
	}
	if !strings.Contains(result, "main") {
		t.Error("Result should contain 'main'")
	}
	if !strings.Contains(result, "()") {
		t.Error("Result should contain '()'")
	}
}

// Helper function test - testing internal logic
func TestRenderTokens_Truncation(t *testing.T) {
	state := &DiffState{}
	tokens := []highlight.Token{
		{Type: chroma.Text, Value: "this is a very long line that should be truncated"},
	}

	bgStyle := lipgloss.NewStyle() // Empty style for testing
	result := state.renderTokens(tokens, 10, bgStyle)

	// Should be truncated to approximately 10 characters plus ellipsis
	// We check that it's shorter than the original and contains ellipsis
	if len([]rune(result)) > 50 {
		t.Error("Result should be truncated")
	}
	if !strings.Contains(result, "…") {
		t.Error("Truncated result should contain ellipsis")
	}
}

// Helper function test - testing internal logic
func TestRenderTokens_EmptyTokens(t *testing.T) {
	state := &DiffState{}
	tokens := []highlight.Token{}

	bgStyle := lipgloss.NewStyle() // Empty style for testing
	result := state.renderTokens(tokens, 100, bgStyle)

	if result != "" {
		t.Error("Empty tokens should produce empty result")
	}
}

// Helper function test - testing internal logic
func TestRenderTokens_TabExpansion(t *testing.T) {
	state := &DiffState{}
	tokens := []highlight.Token{
		{Type: chroma.Text, Value: "hello\tworld"},
	}

	bgStyle := lipgloss.NewStyle() // Empty style for testing
	result := state.renderTokens(tokens, 100, bgStyle)

	// Should not contain tab character after expansion
	if strings.Contains(result, "\t") {
		t.Error("Result should not contain tab characters")
	}
	// Should contain the text
	if !strings.Contains(result, "hello") || !strings.Contains(result, "world") {
		t.Error("Result should contain expanded tab content")
	}
}

// Helper function test - testing internal logic
func TestRenderTokensWithInlineDiff_TabsInModifiedLines(t *testing.T) {
	// Import diffmatchpatch for creating inline diffs
	dmp := diffmatchpatch.New()

	// Test case: A line with tabs where the change happens after a tab
	// Old: "func\tmain() {"
	// New: "func\tMain() {"
	// The 'm' changes to 'M' after the tab

	state := &DiffState{}

	// Left side (old): "func\tmain() {"
	leftTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: "\t"}, // Tab character
		{Type: chroma.NameFunction, Value: "main"},
		{Type: chroma.Punctuation, Value: "() {"},
	}

	// Right side (new): "func\tMain() {"
	rightTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: "\t"},           // Tab character
		{Type: chroma.NameFunction, Value: "Main"}, // Changed to uppercase M
		{Type: chroma.Punctuation, Value: "() {"},
	}

	// Create inline diff between the two lines
	leftText := "func\tmain() {"
	rightText := "func\tMain() {"
	inlineDiff := dmp.DiffMain(leftText, rightText, false)

	bgStyle := lipgloss.NewStyle()

	// Render left side (should highlight 'm' in 'main')
	leftResult := state.renderTokensWithInlineDiff(leftTokens, 100, bgStyle, inlineDiff, false)

	// Render right side (should highlight 'M' in 'Main')
	rightResult := state.renderTokensWithInlineDiff(rightTokens, 100, bgStyle, inlineDiff, true)

	// Basic sanity checks
	if strings.Contains(leftResult, "\t") {
		t.Error("Left result should not contain tab characters (should be expanded)")
	}
	if strings.Contains(rightResult, "\t") {
		t.Error("Right result should not contain tab characters (should be expanded)")
	}

	// Should contain the text content
	if !strings.Contains(leftResult, "func") || !strings.Contains(leftResult, "main") {
		t.Error("Left result should contain 'func' and 'main'")
	}
	if !strings.Contains(rightResult, "func") || !strings.Contains(rightResult, "Main") {
		t.Error("Right result should contain 'func' and 'Main'")
	}

	// The test passes if no panic occurs and content is rendered
	// The exact styling is hard to test due to ANSI codes, but we've verified:
	// 1. Tabs are expanded
	// 2. Content is present
	// 3. No crashes due to position misalignment
}

// Helper function test - testing internal logic
func TestRenderTokensWithInlineDiff_MultipleTabsWithChanges(t *testing.T) {
	dmp := diffmatchpatch.New()

	// Test with multiple tabs to ensure alignment is correct
	// Old: "if\t\tx\t==\t1"
	// New: "if\t\ty\t==\t1"

	state := &DiffState{}

	leftTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "if"},
		{Type: chroma.Text, Value: "\t\t"}, // Two tabs
		{Type: chroma.Name, Value: "x"},
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.Operator, Value: "=="},
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.Number, Value: "1"},
	}

	rightTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "if"},
		{Type: chroma.Text, Value: "\t\t"}, // Two tabs
		{Type: chroma.Name, Value: "y"},    // Changed from x to y
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.Operator, Value: "=="},
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.Number, Value: "1"},
	}

	leftText := "if\t\tx\t==\t1"
	rightText := "if\t\ty\t==\t1"
	inlineDiff := dmp.DiffMain(leftText, rightText, false)

	bgStyle := lipgloss.NewStyle()

	// Render both sides
	leftResult := state.renderTokensWithInlineDiff(leftTokens, 100, bgStyle, inlineDiff, false)
	rightResult := state.renderTokensWithInlineDiff(rightTokens, 100, bgStyle, inlineDiff, true)

	// Verify no tabs remain
	if strings.Contains(leftResult, "\t") {
		t.Error("Left result should not contain tab characters")
	}
	if strings.Contains(rightResult, "\t") {
		t.Error("Right result should not contain tab characters")
	}

	// Verify content is present
	if !strings.Contains(leftResult, "if") || !strings.Contains(leftResult, "x") {
		t.Error("Left result should contain 'if' and 'x'")
	}
	if !strings.Contains(rightResult, "if") || !strings.Contains(rightResult, "y") {
		t.Error("Right result should contain 'if' and 'y'")
	}
}
