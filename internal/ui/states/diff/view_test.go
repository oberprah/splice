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

	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Refactor authentication module",
	}
	commitRange := core.NewSingleCommitRange(commit)

	// Tokens for various lines
	packageAuthTokens := []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}
	emptyTokens := []highlight.Token{}
	oldHelperTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "oldHelper"}, {Type: chroma.Punctuation, Value: "() {"}}
	closeBraceTokens := []highlight.Token{{Type: chroma.Punctuation, Value: "}"}}
	funcLoginTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Login"}, {Type: chroma.Punctuation, Value: "() {"}}
	validateUserTokens := []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "validateUser"}, {Type: chroma.Punctuation, Value: "()"}}
	oldFunctionTokens := []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "oldFunction"}, {Type: chroma.Punctuation, Value: "()"}}
	newFunctionTokens := []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "newFunction"}, {Type: chroma.Punctuation, Value: "()"}}
	checkPermissionsTokens := []highlight.Token{{Type: chroma.Text, Value: "    "}, {Type: chroma.NameFunction, Value: "checkPermissions"}, {Type: chroma.Punctuation, Value: "()"}}

	state := &State{
		Source: commitRange.ToDiffSource(),
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 2,
			Deletions: 2,
		},
		Diff: &diff.FileDiff{
			Path: "internal/auth/handler.go",
			Blocks: []diff.Block{
				// First unchanged block: package auth, empty line
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: packageAuthTokens},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: emptyTokens},
				}},
				// Change block: removed oldHelper function
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.RemovedLine{LeftLineNo: 3, Tokens: oldHelperTokens},
					diff.RemovedLine{LeftLineNo: 4, Tokens: closeBraceTokens},
					diff.RemovedLine{LeftLineNo: 5, Tokens: emptyTokens},
				}},
				// Unchanged block: func Login, validateUser
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 6, RightLineNo: 3, Tokens: funcLoginTokens},
					{LeftLineNo: 7, RightLineNo: 4, Tokens: validateUserTokens},
				}},
				// Change block: modified oldFunction->newFunction, added checkPermissions
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  8,
						RightLineNo: 5,
						LeftTokens:  oldFunctionTokens,
						RightTokens: newFunctionTokens,
						InlineDiff: dmp.DiffMain(
							"    oldFunction()",
							"    newFunction()",
							false,
						),
					},
					diff.AddedLine{RightLineNo: 6, Tokens: checkPermissionsTokens},
				}},
				// Final unchanged block: closing brace
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 9, RightLineNo: 7, Tokens: closeBraceTokens},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "all_line_types.golden")
}

func TestDiffState_View_TokenRendering(t *testing.T) {
	commit := core.GitCommit{Hash: "abc123"}

	// Shared tokens for unchanged lines
	funcMainTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.NameFunction, Value: "main"},
		{Type: chroma.Punctuation, Value: "()"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Punctuation, Value: "{"},
	}
	fmtPrintlnTokens := []highlight.Token{
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.NameFunction, Value: "fmt"},
		{Type: chroma.Punctuation, Value: "."},
		{Type: chroma.NameFunction, Value: "Println"},
		{Type: chroma.Punctuation, Value: "("},
		{Type: chroma.LiteralString, Value: "\"hello\tworld\""},
		{Type: chroma.Punctuation, Value: ")"},
	}
	longCommentTokens := []highlight.Token{
		{Type: chroma.Text, Value: "\t"},
		{Type: chroma.Comment, Value: "// This is a very long comment that should be truncated when terminal is narrow"},
	}
	closeBraceTokens := []highlight.Token{
		{Type: chroma.Punctuation, Value: "}"},
	}

	state := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "test.go"},
		Diff: &diff.FileDiff{
			Path: "test.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: funcMainTokens},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: fmtPrintlnTokens},
					{LeftLineNo: 3, RightLineNo: 3, Tokens: longCommentTokens},
					{LeftLineNo: 4, RightLineNo: 4, Tokens: closeBraceTokens},
				}},
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

	commit := core.GitCommit{Hash: "abc123"}

	// Tokens for left side
	leftMainTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.NameFunction, Value: "main"},
		{Type: chroma.Punctuation, Value: "() {"},
	}
	leftIfXTokens := []highlight.Token{
		{Type: chroma.Text, Value: "    "},
		{Type: chroma.Keyword, Value: "if"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Name, Value: "x"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Operator, Value: "=="},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Number, Value: "1"},
	}
	leftOldVarTokens := []highlight.Token{
		{Type: chroma.Text, Value: "    "},
		{Type: chroma.Name, Value: "oldVar"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Operator, Value: ":="},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Number, Value: "42"},
	}

	// Tokens for right side
	rightMainTokens := []highlight.Token{
		{Type: chroma.Keyword, Value: "func"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.NameFunction, Value: "Main"},
		{Type: chroma.Punctuation, Value: "() {"},
	}
	rightIfYTokens := []highlight.Token{
		{Type: chroma.Text, Value: "    "},
		{Type: chroma.Keyword, Value: "if"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Name, Value: "y"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Operator, Value: "=="},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Number, Value: "1"},
	}
	rightNewVarTokens := []highlight.Token{
		{Type: chroma.Text, Value: "    "},
		{Type: chroma.Name, Value: "newVar"},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Operator, Value: ":="},
		{Type: chroma.Text, Value: " "},
		{Type: chroma.Number, Value: "42"},
	}

	state := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "test.go"},
		Diff: &diff.FileDiff{
			Path: "test.go",
			Blocks: []diff.Block{
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  1,
						RightLineNo: 1,
						LeftTokens:  leftMainTokens,
						RightTokens: rightMainTokens,
						InlineDiff: dmp.DiffMain(
							"func main() {",
							"func Main() {",
							false,
						),
					},
					diff.ModifiedLine{
						LeftLineNo:  2,
						RightLineNo: 2,
						LeftTokens:  leftIfXTokens,
						RightTokens: rightIfYTokens,
						InlineDiff: dmp.DiffMain(
							"    if x == 1",
							"    if y == 1",
							false,
						),
					},
					diff.ModifiedLine{
						LeftLineNo:  3,
						RightLineNo: 3,
						LeftTokens:  leftOldVarTokens,
						RightTokens: rightNewVarTokens,
						InlineDiff: dmp.DiffMain(
							"    oldVar := 42",
							"    newVar := 42",
							false,
						),
					},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "inline_diff_rendering.golden")
}

func TestDiffState_View_EmptyDiff(t *testing.T) {
	commit := core.GitCommit{Hash: "abc123"}
	state := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path:   "file.go",
			Blocks: []diff.Block{},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "empty_diff.golden")
}

func TestDiffState_View_Viewport(t *testing.T) {
	// Create a diff with many lines in a single unchanged block
	linePairs := make([]diff.LinePair, 100)
	for i := 0; i < 100; i++ {
		linePairs[i] = diff.LinePair{
			LeftLineNo:  i + 1,
			RightLineNo: i + 1,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "line content"}},
		}
	}

	commit := core.GitCommit{Hash: "abc123"}
	state := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: linePairs},
			},
		},
		ViewportStart: 50, // Start at line 50
	}

	ctx := testutils.MockContext{W: 80, H: 10}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "viewport.golden")
}

func TestDiffState_View_RangeHeader(t *testing.T) {
	// Test that the header displays range format for multi-commit ranges
	startCommit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Start commit",
	}
	endCommit := core.GitCommit{
		Hash:    "def456abc123456789012345678901234567890abc",
		Message: "End commit",
	}
	commitRange := core.NewCommitRange(startCommit, endCommit, 4)

	packageAuthTokens := []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}
	oldFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Old"}, {Type: chroma.Punctuation, Value: "() {"}}
	newFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "New"}, {Type: chroma.Punctuation, Value: "() {"}}

	state := &State{
		Source: commitRange.ToDiffSource(),
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 25,
			Deletions: 13,
		},
		Diff: &diff.FileDiff{
			Path: "internal/auth/handler.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: packageAuthTokens},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  2,
						RightLineNo: 2,
						LeftTokens:  oldFuncTokens,
						RightTokens: newFuncTokens,
						InlineDiff: []diffmatchpatch.Diff{
							{Type: diffmatchpatch.DiffEqual, Text: "func "},
							{Type: diffmatchpatch.DiffDelete, Text: "Old"},
							{Type: diffmatchpatch.DiffInsert, Text: "New"},
							{Type: diffmatchpatch.DiffEqual, Text: "() {"},
						},
					},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "range_header.golden")
}

func TestDiffState_View_UnstagedChangesHeader(t *testing.T) {
	// Test that the header displays "unstaged" for unstaged uncommitted changes
	packageAuthTokens := []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}
	oldFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Old"}, {Type: chroma.Punctuation, Value: "() {"}}
	newFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "New"}, {Type: chroma.Punctuation, Value: "() {"}}

	state := &State{
		Source: core.UncommittedChangesDiffSource{Type: core.UncommittedTypeUnstaged},
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 8,
			Deletions: 3,
		},
		Diff: &diff.FileDiff{
			Path: "internal/auth/handler.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: packageAuthTokens},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  2,
						RightLineNo: 2,
						LeftTokens:  oldFuncTokens,
						RightTokens: newFuncTokens,
						InlineDiff: []diffmatchpatch.Diff{
							{Type: diffmatchpatch.DiffEqual, Text: "func "},
							{Type: diffmatchpatch.DiffDelete, Text: "Old"},
							{Type: diffmatchpatch.DiffInsert, Text: "New"},
							{Type: diffmatchpatch.DiffEqual, Text: "() {"},
						},
					},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "unstaged_header.golden")
}

func TestDiffState_View_StagedChangesHeader(t *testing.T) {
	// Test that the header displays "staged" for staged uncommitted changes
	packageAuthTokens := []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}
	oldFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Old"}, {Type: chroma.Punctuation, Value: "() {"}}
	newFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "New"}, {Type: chroma.Punctuation, Value: "() {"}}

	state := &State{
		Source: core.UncommittedChangesDiffSource{Type: core.UncommittedTypeStaged},
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 5,
			Deletions: 2,
		},
		Diff: &diff.FileDiff{
			Path: "internal/auth/handler.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: packageAuthTokens},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  2,
						RightLineNo: 2,
						LeftTokens:  oldFuncTokens,
						RightTokens: newFuncTokens,
						InlineDiff: []diffmatchpatch.Diff{
							{Type: diffmatchpatch.DiffEqual, Text: "func "},
							{Type: diffmatchpatch.DiffDelete, Text: "Old"},
							{Type: diffmatchpatch.DiffInsert, Text: "New"},
							{Type: diffmatchpatch.DiffEqual, Text: "() {"},
						},
					},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "staged_header.golden")
}

func TestDiffState_View_AllUncommittedChangesHeader(t *testing.T) {
	// Test that the header displays "uncommitted" for all uncommitted changes
	packageAuthTokens := []highlight.Token{{Type: chroma.Keyword, Value: "package"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameNamespace, Value: "auth"}}
	oldFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "Old"}, {Type: chroma.Punctuation, Value: "() {"}}
	newFuncTokens := []highlight.Token{{Type: chroma.Keyword, Value: "func"}, {Type: chroma.Text, Value: " "}, {Type: chroma.NameFunction, Value: "New"}, {Type: chroma.Punctuation, Value: "() {"}}

	state := &State{
		Source: core.UncommittedChangesDiffSource{Type: core.UncommittedTypeAll},
		File: core.FileChange{
			Path:      "internal/auth/handler.go",
			Additions: 12,
			Deletions: 6,
		},
		Diff: &diff.FileDiff{
			Path: "internal/auth/handler.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: packageAuthTokens},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  2,
						RightLineNo: 2,
						LeftTokens:  oldFuncTokens,
						RightTokens: newFuncTokens,
						InlineDiff: []diffmatchpatch.Diff{
							{Type: diffmatchpatch.DiffEqual, Text: "func "},
							{Type: diffmatchpatch.DiffDelete, Text: "Old"},
							{Type: diffmatchpatch.DiffInsert, Text: "New"},
							{Type: diffmatchpatch.DiffEqual, Text: "() {"},
						},
					},
				}},
			},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	output := state.View(ctx)

	assertDiffViewGolden(t, output.(*components.ViewBuilder), "all_uncommitted_header.golden")
}
