package diff

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/alecthomas/chroma/v2"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/testutils"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update golden files")

// Per-file helper that adds subdirectory prefix
func assertDiffViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", filename)
	testutils.AssertGolden(t, output.String(), goldenPath, *update)
}

func TestDiffState_View_AllLineTypes(t *testing.T) {
	dmp := diffmatchpatch.New()

	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{
			Hash:    "abc123def456789012345678901234567890abcd",
			Message: "Refactor authentication module",
		}),
		File: core.FileChange{
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

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "all_line_types.golden")
}

func TestDiffState_View_TokenRendering(t *testing.T) {
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "test.go"},
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
	ctx := testutils.MockContext{W: 60, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "token_rendering.golden")
}

func TestDiffState_View_InlineDiffRendering(t *testing.T) {
	dmp := diffmatchpatch.New()

	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "test.go"},
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

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "inline_diff_rendering.golden")
}

func TestDiffState_View_EmptyDiff(t *testing.T) {
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
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

	ctx := testutils.MockContext{W: 80, H: 24}
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
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
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

	ctx := testutils.MockContext{W: 80, H: 10}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "viewport.golden")
}

func TestDiffState_View_RangeHeader(t *testing.T) {
	// Test that the header displays range format for multi-commit ranges
	state := &State{
		CommitRange: core.NewCommitRange(
			core.GitCommit{
				Hash:    "abc123def456789012345678901234567890abcd",
				Message: "Start commit",
			},
			core.GitCommit{
				Hash:    "def456abc123456789012345678901234567890abc",
				Message: "End commit",
			},
			4, // 4 commits in range
		),
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 25,
			Deletions: 13,
		},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "internal/auth/handler.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Old"}, {Type: chroma.Punctuation, Value: "() {"}}},
				},
			},
			Right: diff.FileContent{
				Path: "internal/auth/handler.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "New"}, {Type: chroma.Punctuation, Value: "() {"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
				diff.ModifiedAlignment{
					LeftIdx:  1,
					RightIdx: 1,
					InlineDiff: []diffmatchpatch.Diff{
						{Type: diffmatchpatch.DiffEqual, Text: "func "},
						{Type: diffmatchpatch.DiffDelete, Text: "Old"},
						{Type: diffmatchpatch.DiffInsert, Text: "New"},
						{Type: diffmatchpatch.DiffEqual, Text: "() {"},
					},
				},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "range_header.golden")
}

// ═══════════════════════════════════════════════════════════
// SEGMENT-BASED RENDERING TESTS
// These tests verify the new segment-based rendering that eliminates blank line padding.
// ═══════════════════════════════════════════════════════════

func TestDiffState_View_SegmentPureAdditions(t *testing.T) {
	// Test a hunk with only additions (right side has more lines)
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "test.go", Additions: 3, Deletions: 0},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "main"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "main"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "fmt"}, {Type: chroma.Punctuation, Value: "."}, {Type: chroma.NameFunction, Value: "Println"}, {Type: chroma.Punctuation, Value: "("}, {Type: chroma.LiteralString, Value: "\"hello\""}, {Type: chroma.Punctuation, Value: ")"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "fmt"}, {Type: chroma.Punctuation, Value: "."}, {Type: chroma.NameFunction, Value: "Println"}, {Type: chroma.Punctuation, Value: "("}, {Type: chroma.LiteralString, Value: "\"world\""}, {Type: chroma.Punctuation, Value: ")"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "fmt"}, {Type: chroma.Punctuation, Value: "."}, {Type: chroma.NameFunction, Value: "Println"}, {Type: chroma.Punctuation, Value: "("}, {Type: chroma.LiteralString, Value: "\"!\""}, {Type: chroma.Punctuation, Value: ")"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 2}, // package, func main
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{}, // No deletions
					RightLines: []diff.HunkLine{
						{SourceIdx: 2, Type: diff.HunkLineAdded},
						{SourceIdx: 3, Type: diff.HunkLineAdded},
						{SourceIdx: 4, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 2, RightStart: 5, Count: 1}, // }
			},
		},
		SegmentIndex: 0,
		LeftOffset:   0,
		RightOffset:  0,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "segment_pure_additions.golden")
}

func TestDiffState_View_SegmentPureDeletions(t *testing.T) {
	// Test a hunk with only deletions (left side has more lines)
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "def456"}),
		File:        core.FileChange{Path: "test.go", Additions: 0, Deletions: 3},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "main"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFunc1"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFunc2"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFunc3"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "main"}, {Type: chroma.Punctuation, Value: "() {"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 2}, // package, func main
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{
						{SourceIdx: 2, Type: diff.HunkLineRemoved},
						{SourceIdx: 3, Type: diff.HunkLineRemoved},
						{SourceIdx: 4, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{}, // No additions
				},
				diff.UnchangedSegment{LeftStart: 5, RightStart: 2, Count: 1}, // }
			},
		},
		SegmentIndex: 0,
		LeftOffset:   0,
		RightOffset:  0,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "segment_pure_deletions.golden")
}

func TestDiffState_View_SegmentMixedChanges(t *testing.T) {
	// Test a hunk with both additions and deletions (different line counts)
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "789abc"}),
		File:        core.FileChange{Path: "test.go", Additions: 2, Deletions: 3},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldLine1"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldLine2"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldLine3"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newLine1"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newLine2"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 1}, // package
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{
						{SourceIdx: 1, Type: diff.HunkLineRemoved},
						{SourceIdx: 2, Type: diff.HunkLineRemoved},
						{SourceIdx: 3, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 1, Type: diff.HunkLineAdded},
						{SourceIdx: 2, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 4, RightStart: 3, Count: 1}, // }
			},
		},
		SegmentIndex: 0,
		LeftOffset:   0,
		RightOffset:  0,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "segment_mixed_changes.golden")
}

func TestDiffState_View_SegmentMultipleHunks(t *testing.T) {
	// Test rendering with multiple hunks separated by unchanged regions
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "multi12"}),
		File:        core.FileChange{Path: "test.go", Additions: 2, Deletions: 2},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFirst"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// separator"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldSecond"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newFirst"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// separator"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newSecond"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 1}, // package
				diff.HunkSegment{ // First hunk
					LeftLines: []diff.HunkLine{
						{SourceIdx: 1, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 1, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 2, RightStart: 2, Count: 1}, // separator
				diff.HunkSegment{ // Second hunk
					LeftLines: []diff.HunkLine{
						{SourceIdx: 3, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 3, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 4, RightStart: 4, Count: 1}, // }
			},
		},
		SegmentIndex: 0,
		LeftOffset:   0,
		RightOffset:  0,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "segment_multiple_hunks.golden")
}

func TestDiffState_View_SegmentStartAtHunk(t *testing.T) {
	// Test rendering when viewport starts at a hunk segment (simulating navigation to change)
	state := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "start12"}),
		File:        core.FileChange{Path: "test.go", Additions: 2, Deletions: 1},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// line 2"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// line 3"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldLine"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Right: diff.FileContent{
				Path: "test.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "main"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// line 2"}}},
					{Tokens: []highlight.Token{{Type: chroma.Comment, Value: "// line 3"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newLine1"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newLine2"}, {Type: chroma.Punctuation, Value: "()"}}},
					{Tokens: []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}},
				},
			},
			Segments: []diff.Segment{
				diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 3}, // package, line 2, line 3
				diff.HunkSegment{
					LeftLines: []diff.HunkLine{
						{SourceIdx: 3, Type: diff.HunkLineRemoved},
					},
					RightLines: []diff.HunkLine{
						{SourceIdx: 3, Type: diff.HunkLineAdded},
						{SourceIdx: 4, Type: diff.HunkLineAdded},
					},
				},
				diff.UnchangedSegment{LeftStart: 4, RightStart: 5, Count: 1}, // }
			},
		},
		SegmentIndex: 1, // Start at the hunk segment
		LeftOffset:   0,
		RightOffset:  0,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "segment_start_at_hunk.golden")
}
