package diff

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Per-file helper that adds subdirectory prefix
func assertDiffViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), ""+filename, *update)
}

func TestDiffState_View_AllLineTypes(t *testing.T) {
	dmp := diffmatchpatch.New()

	state := &State{
		Commit: git.GitCommit{
			Hash:    "abc123def456789012345678901234567890abcd",
			Message: "Refactor authentication module",
		},
		File: git.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 2,
			Deletions: 2,
		},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "internal/auth/handler.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}},
					{Tokens: []highlight.Token{}}, // empty line
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "oldHelper"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
					{Tokens: []highlight.Token{}}, // empty line
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Login"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "validateUser"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFunction"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "internal/auth/handler.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}},
					{Tokens: []highlight.Token{}}, // empty line
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Login"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "validateUser"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newFunction"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "checkPermissions"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0}, // package auth
				diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1}, // empty line
				diff.RemovedAlignment{LeftIdx: 2},                // func oldHelper() {
				diff.RemovedAlignment{LeftIdx: 3},                // }
				diff.RemovedAlignment{LeftIdx: 4},                // empty line
				diff.UnchangedAlignment{LeftIdx: 5, RightIdx: 2}, // func Login() {
				diff.UnchangedAlignment{LeftIdx: 6, RightIdx: 3}, // validateUser()
				diff.ModifiedAlignment{
					LeftIdx:  7,
					RightIdx: 4,
					InlineDiff: dmp.DiffMain(
						"    oldFunction()",
						"    newFunction()",
						false,
					),
				}, // oldFunction -> newFunction
				diff.AddedAlignment{RightIdx: 5},                 // added checkPermissions()
				diff.UnchangedAlignment{LeftIdx: 8, RightIdx: 6}, // }
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "all_line_types.golden")
}

func TestDiffState_View_TokenRendering(t *testing.T) {
	state := &State{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "test.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					// Multiple token types with syntax highlighting
					{Tokens: []highlight.Token{
						{Type: chroma.Keyword, Value: "func"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.NameFunction, Value: "main"},
						{Type: chroma.Punctuation, Value: "()"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Punctuation, Value: "{"},
					}},
					// Tab expansion - tabs should be converted to spaces
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "\t"},
						{Type: chroma.NameFunction, Value: "fmt"},
						{Type: chroma.Punctuation, Value: "."},
						{Type: chroma.NameFunction, Value: "Println"},
						{Type: chroma.Punctuation, Value: "("},
						{Type: chroma.LiteralString, Value: "\"hello\tworld\""},
						{Type: chroma.Punctuation, Value: ")"},
					}},
					// Long line truncation - this comment is too long for narrow terminals
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "\t"},
						{Type: chroma.Comment, Value: "// This is a very long comment that should be truncated when terminal is narrow"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Punctuation, Value: "}"},
					}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{
						{Type: chroma.Keyword, Value: "func"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.NameFunction, Value: "main"},
						{Type: chroma.Punctuation, Value: "()"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Punctuation, Value: "{"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "\t"},
						{Type: chroma.NameFunction, Value: "fmt"},
						{Type: chroma.Punctuation, Value: "."},
						{Type: chroma.NameFunction, Value: "Println"},
						{Type: chroma.Punctuation, Value: "("},
						{Type: chroma.LiteralString, Value: "\"hello\tworld\""},
						{Type: chroma.Punctuation, Value: ")"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "\t"},
						{Type: chroma.Comment, Value: "// This is a very long comment that should be truncated when terminal is narrow"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Punctuation, Value: "}"},
					}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
				diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1},
				diff.UnchangedAlignment{LeftIdx: 2, RightIdx: 2},
				diff.UnchangedAlignment{LeftIdx: 3, RightIdx: 3},
			},
		},
	}

	// Test with narrow width to trigger truncation
	ctx := &mockContext{width: 60, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "token_rendering.golden")
}

func TestDiffState_View_InlineDiffRendering(t *testing.T) {
	dmp := diffmatchpatch.New()

	state := &State{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "test.go"},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{
						{Type: chroma.Keyword, Value: "func"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.NameFunction, Value: "main"},
						{Type: chroma.Punctuation, Value: "() {"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "    "},
						{Type: chroma.Keyword, Value: "if"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Name, Value: "x"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Operator, Value: "=="},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Number, Value: "1"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "    "},
						{Type: chroma.Name, Value: "oldVar"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Operator, Value: ":="},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Number, Value: "42"},
					}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{
						{Type: chroma.Keyword, Value: "func"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.NameFunction, Value: "Main"},
						{Type: chroma.Punctuation, Value: "() {"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "    "},
						{Type: chroma.Keyword, Value: "if"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Name, Value: "y"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Operator, Value: "=="},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Number, Value: "1"},
					}},
					{Tokens: []highlight.Token{
						{Type: chroma.Text, Value: "    "},
						{Type: chroma.Name, Value: "newVar"},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Operator, Value: ":="},
						{Type: chroma.Text, Value: " "},
						{Type: chroma.Number, Value: "42"},
					}},
				},
			},
			Alignments: []diff.Alignment{
				diff.ModifiedAlignment{
					LeftIdx:  0,
					RightIdx: 0,
					InlineDiff: dmp.DiffMain(
						"func main() {",
						"func Main() {",
						false,
					),
				},
				diff.ModifiedAlignment{
					LeftIdx:  1,
					RightIdx: 1,
					InlineDiff: dmp.DiffMain(
						"    if x == 1",
						"    if y == 1",
						false,
					),
				},
				diff.ModifiedAlignment{
					LeftIdx:  2,
					RightIdx: 2,
					InlineDiff: dmp.DiffMain(
						"    oldVar := 42",
						"    newVar := 42",
						false,
					),
				},
			},
		},
	}

	ctx := &mockContext{width: 80, height: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "inline_diff_rendering.golden")
}

func TestDiffState_View_EmptyDiff(t *testing.T) {
	state := &State{
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

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "empty_diff.golden")
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

	state := &State{
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

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "viewport.golden")
}
