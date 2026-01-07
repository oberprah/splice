package diff

import (
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

// ═══════════════════════════════════════════════════════════
// BuildFileContent Tests
// ═══════════════════════════════════════════════════════════

func TestBuildFileContent_SimpleGoFile(t *testing.T) {
	content := "package main\n\nfunc hello() {\n\tprintln(\"hi\")\n}"
	path := "main.go"

	fc, err := BuildFileContent(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fc.Path != path {
		t.Errorf("expected path %q, got %q", path, fc.Path)
	}

	// Should have 5 lines (split by \n produces 5 elements)
	expectedLineCount := 5
	if len(fc.Lines) != expectedLineCount {
		t.Errorf("expected %d lines, got %d", expectedLineCount, len(fc.Lines))
	}

	// Verify first line contains "package main"
	if len(fc.Lines) > 0 {
		firstLineText := fc.Lines[0].Text()
		if firstLineText != "package main" {
			t.Errorf("expected first line to be 'package main', got %q", firstLineText)
		}
	}
}

func TestBuildFileContent_EmptyFile(t *testing.T) {
	content := ""
	path := "empty.txt"

	fc, err := BuildFileContent(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fc.Path != path {
		t.Errorf("expected path %q, got %q", path, fc.Path)
	}

	// Empty content results in 1 line (empty line) due to how strings.Split works
	// strings.Split("", "\n") returns [""], not []
	if len(fc.Lines) != 1 {
		t.Errorf("expected 1 line for empty file, got %d", len(fc.Lines))
	}

	// Verify the line is empty
	if len(fc.Lines) > 0 {
		text := fc.Lines[0].Text()
		if text != "" {
			t.Errorf("expected empty line text, got %q", text)
		}
	}
}

func TestBuildFileContent_SingleLineNoNewline(t *testing.T) {
	content := "hello world"
	path := "hello.txt"

	fc, err := BuildFileContent(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Single line without newline should result in 1 line
	if len(fc.Lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(fc.Lines))
	}

	if len(fc.Lines) > 0 {
		text := fc.Lines[0].Text()
		if text != "hello world" {
			t.Errorf("expected 'hello world', got %q", text)
		}
	}
}

func TestBuildFileContent_PreservesTokens(t *testing.T) {
	// Use a simple text file to verify tokens are preserved
	content := "line1\nline2\nline3"
	path := "test.txt"

	fc, err := BuildFileContent(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify each line has tokens
	if len(fc.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(fc.Lines))
	}

	for i, expectedText := range []string{"line1", "line2", "line3"} {
		text := fc.Lines[i].Text()
		if text != expectedText {
			t.Errorf("line %d: expected %q, got %q", i, expectedText, text)
		}

		// Verify tokens exist (should have at least one token per non-empty line)
		if len(fc.Lines[i].Tokens) == 0 {
			t.Errorf("line %d: expected tokens, got none", i)
		}
	}
}

func TestBuildFileContent_SyntaxHighlighting(t *testing.T) {
	// Test that syntax highlighting produces different token types for Go code
	content := "package main"
	path := "test.go"

	fc, err := BuildFileContent(path, content)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fc.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(fc.Lines))
	}

	// The Go lexer should produce at least one token
	tokens := fc.Lines[0].Tokens
	if len(tokens) == 0 {
		t.Fatal("expected tokens from Go syntax highlighting")
	}

	// Verify text reconstructs correctly
	text := fc.Lines[0].Text()
	if text != "package main" {
		t.Errorf("expected 'package main', got %q", text)
	}
}

// ═══════════════════════════════════════════════════════════
// BuildAlignments Tests - Simple Cases
// ═══════════════════════════════════════════════════════════

func TestBuildAlignments_AllUnchanged(t *testing.T) {
	// Create identical file content for both sides
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
		},
	}

	// Create a diff with all context lines
	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Context, Content: "line2", OldLineNo: 2, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce 3 UnchangedAlignment entries
	if len(alignments) != 3 {
		t.Fatalf("expected 3 alignments, got %d", len(alignments))
	}

	for i, align := range alignments {
		ua, ok := align.(UnchangedAlignment)
		if !ok {
			t.Errorf("alignment %d: expected UnchangedAlignment, got %T", i, align)
			continue
		}
		if ua.LeftIdx != i || ua.RightIdx != i {
			t.Errorf("alignment %d: expected indices (%d, %d), got (%d, %d)",
				i, i, i, ua.LeftIdx, ua.RightIdx)
		}
	}
}

func TestBuildAlignments_SimpleModification(t *testing.T) {
	// One line changed in the middle
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("old line"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("new line"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "old line", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "new line", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce: Unchanged(0,0), Modified(1,1), Unchanged(2,2)
	if len(alignments) != 3 {
		t.Fatalf("expected 3 alignments, got %d", len(alignments))
	}

	// First alignment: unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Errorf("alignment 0: expected UnchangedAlignment, got %T", alignments[0])
	}

	// Second alignment: modified
	ma, ok := alignments[1].(ModifiedAlignment)
	if !ok {
		t.Fatalf("alignment 1: expected ModifiedAlignment, got %T", alignments[1])
	}
	if ma.LeftIdx != 1 || ma.RightIdx != 1 {
		t.Errorf("alignment 1: expected indices (1, 1), got (%d, %d)", ma.LeftIdx, ma.RightIdx)
	}
	// Verify inline diff exists
	if len(ma.InlineDiff) == 0 {
		t.Error("alignment 1: expected inline diff, got none")
	}

	// Third alignment: unchanged
	if _, ok := alignments[2].(UnchangedAlignment); !ok {
		t.Errorf("alignment 2: expected UnchangedAlignment, got %T", alignments[2])
	}
}

func TestBuildAlignments_PureRemoval(t *testing.T) {
	// One line removed
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("removed"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "removed", OldLineNo: 2, NewLineNo: 0},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 2},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce: Unchanged(0,0), Removed(1), Unchanged(2,1)
	if len(alignments) != 3 {
		t.Fatalf("expected 3 alignments, got %d", len(alignments))
	}

	// First: unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Errorf("alignment 0: expected UnchangedAlignment, got %T", alignments[0])
	}

	// Second: removed
	ra, ok := alignments[1].(RemovedAlignment)
	if !ok {
		t.Fatalf("alignment 1: expected RemovedAlignment, got %T", alignments[1])
	}
	if ra.LeftIdx != 1 {
		t.Errorf("alignment 1: expected LeftIdx=1, got %d", ra.LeftIdx)
	}

	// Third: unchanged
	ua, ok := alignments[2].(UnchangedAlignment)
	if !ok {
		t.Fatalf("alignment 2: expected UnchangedAlignment, got %T", alignments[2])
	}
	if ua.LeftIdx != 2 || ua.RightIdx != 1 {
		t.Errorf("alignment 2: expected indices (2, 1), got (%d, %d)", ua.LeftIdx, ua.RightIdx)
	}
}

func TestBuildAlignments_PureAddition(t *testing.T) {
	// One line added
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("added"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Add, Content: "added", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 2, NewLineNo: 3},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce: Unchanged(0,0), Added(1), Unchanged(1,2)
	if len(alignments) != 3 {
		t.Fatalf("expected 3 alignments, got %d", len(alignments))
	}

	// First: unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Errorf("alignment 0: expected UnchangedAlignment, got %T", alignments[0])
	}

	// Second: added
	aa, ok := alignments[1].(AddedAlignment)
	if !ok {
		t.Fatalf("alignment 1: expected AddedAlignment, got %T", alignments[1])
	}
	if aa.RightIdx != 1 {
		t.Errorf("alignment 1: expected RightIdx=1, got %d", aa.RightIdx)
	}

	// Third: unchanged
	ua, ok := alignments[2].(UnchangedAlignment)
	if !ok {
		t.Fatalf("alignment 2: expected UnchangedAlignment, got %T", alignments[2])
	}
	if ua.LeftIdx != 1 || ua.RightIdx != 2 {
		t.Errorf("alignment 2: expected indices (1, 2), got (%d, %d)", ua.LeftIdx, ua.RightIdx)
	}
}

// ═══════════════════════════════════════════════════════════
// BuildAlignments Tests - Complex Cases
// ═══════════════════════════════════════════════════════════

func TestBuildAlignments_MultipleChangesInHunk(t *testing.T) {
	// Multiple removed and added lines in one hunk
	// Some will pair (similar), some won't
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("fmt.Println(hello)"), // similar to added line
			makeAlignedLine("completely different"),
			makeAlignedLine("unchanged"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("fmt.Println(world)"), // similar to removed line
			makeAlignedLine("brand new line"),
			makeAlignedLine("unchanged"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "unchanged", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "fmt.Println(hello)", OldLineNo: 2, NewLineNo: 0},
			{Type: Remove, Content: "completely different", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "fmt.Println(world)", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "brand new line", OldLineNo: 0, NewLineNo: 3},
			{Type: Context, Content: "unchanged", OldLineNo: 4, NewLineNo: 4},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Expected: Unchanged(0,0), Modified(1,1), Removed(2), Added(2), Unchanged(3,3)
	// The two Println lines should pair; the others should not
	if len(alignments) < 4 {
		t.Fatalf("expected at least 4 alignments, got %d", len(alignments))
	}

	// First: unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Errorf("alignment 0: expected UnchangedAlignment, got %T", alignments[0])
	}

	// Should have one ModifiedAlignment for the paired Println lines
	hasModified := false
	for _, align := range alignments {
		if _, ok := align.(ModifiedAlignment); ok {
			hasModified = true
			break
		}
	}
	if !hasModified {
		t.Error("expected at least one ModifiedAlignment for paired lines")
	}

	// Last: unchanged
	lastIdx := len(alignments) - 1
	if _, ok := alignments[lastIdx].(UnchangedAlignment); !ok {
		t.Errorf("alignment %d: expected UnchangedAlignment, got %T", lastIdx, alignments[lastIdx])
	}
}

func TestBuildAlignments_NoPairing_BelowThreshold(t *testing.T) {
	// Two lines that are too dissimilar to pair
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("apple"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("zebra"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "apple", OldLineNo: 1, NewLineNo: 0},
			{Type: Add, Content: "zebra", OldLineNo: 0, NewLineNo: 1},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce: Removed(0), Added(0)
	// No pairing because similarity is too low
	if len(alignments) != 2 {
		t.Fatalf("expected 2 alignments, got %d", len(alignments))
	}

	// Should be Removed and Added, not Modified
	hasModified := false
	for _, align := range alignments {
		if _, ok := align.(ModifiedAlignment); ok {
			hasModified = true
		}
	}
	if hasModified {
		t.Error("expected no ModifiedAlignment for dissimilar lines")
	}
}

func TestBuildAlignments_EmptyFiles(t *testing.T) {
	left := FileContent{Path: "test.txt", Lines: []AlignedLine{}}
	right := FileContent{Path: "test.txt", Lines: []AlignedLine{}}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines:   []Line{},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Empty files should produce no alignments
	if len(alignments) != 0 {
		t.Errorf("expected 0 alignments for empty files, got %d", len(alignments))
	}
}

func TestBuildAlignments_NilParsedDiff(t *testing.T) {
	left := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{makeAlignedLine("line1")},
	}
	right := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{makeAlignedLine("line1")},
	}

	alignments := BuildAlignments(left, right, nil)

	// Nil diff should produce no alignments
	if alignments != nil {
		t.Errorf("expected nil alignments for nil parsedDiff, got %v", alignments)
	}
}

// ═══════════════════════════════════════════════════════════
// BuildAlignments Tests - Inline Diff Verification
// ═══════════════════════════════════════════════════════════

func TestBuildAlignments_InlineDiff_SimpleChange(t *testing.T) {
	// Test that inline diff is computed correctly
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("Hello"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("Hello World"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "Hello", OldLineNo: 1, NewLineNo: 0},
			{Type: Add, Content: "Hello World", OldLineNo: 0, NewLineNo: 1},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce one ModifiedAlignment with inline diff
	if len(alignments) != 1 {
		t.Fatalf("expected 1 alignment, got %d", len(alignments))
	}

	ma, ok := alignments[0].(ModifiedAlignment)
	if !ok {
		t.Fatalf("expected ModifiedAlignment, got %T", alignments[0])
	}

	// Verify inline diff structure
	if len(ma.InlineDiff) == 0 {
		t.Fatal("expected inline diff, got none")
	}

	// Expected diff: Equal("Hello"), Insert(" World")
	// Verify by reconstructing text
	var leftText, rightText string
	for _, diff := range ma.InlineDiff {
		switch diff.Type {
		case diffmatchpatch.DiffEqual:
			leftText += diff.Text
			rightText += diff.Text
		case diffmatchpatch.DiffDelete:
			leftText += diff.Text
		case diffmatchpatch.DiffInsert:
			rightText += diff.Text
		}
	}

	if leftText != "Hello" {
		t.Errorf("expected left text 'Hello', got %q", leftText)
	}
	if rightText != "Hello World" {
		t.Errorf("expected right text 'Hello World', got %q", rightText)
	}
}

func TestBuildAlignments_InlineDiff_CompleteReplacement(t *testing.T) {
	// Lines with no common text
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("abc"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("xyz"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "abc", OldLineNo: 1, NewLineNo: 0},
			{Type: Add, Content: "xyz", OldLineNo: 0, NewLineNo: 1},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// These lines are too dissimilar, should produce Removed + Added, not Modified
	if len(alignments) != 2 {
		t.Fatalf("expected 2 alignments, got %d", len(alignments))
	}

	// Verify we got RemovedAlignment and AddedAlignment
	hasRemoved := false
	hasAdded := false
	for _, align := range alignments {
		switch align.(type) {
		case RemovedAlignment:
			hasRemoved = true
		case AddedAlignment:
			hasAdded = true
		}
	}

	if !hasRemoved || !hasAdded {
		t.Error("expected RemovedAlignment and AddedAlignment for completely different lines")
	}
}

// ═══════════════════════════════════════════════════════════
// Integration Tests
// ═══════════════════════════════════════════════════════════

func TestIntegration_EndToEnd_SimpleModification(t *testing.T) {
	// Test the full pipeline: BuildFileContent -> BuildAlignments
	oldContent := "package main\n\nfunc hello() {\n\tfmt.Println(\"old\")\n}"
	newContent := "package main\n\nfunc hello() {\n\tfmt.Println(\"new\")\n}"

	left, err := BuildFileContent("main.go", oldContent)
	if err != nil {
		t.Fatalf("BuildFileContent (left) failed: %v", err)
	}

	right, err := BuildFileContent("main.go", newContent)
	if err != nil {
		t.Fatalf("BuildFileContent (right) failed: %v", err)
	}

	parsedDiff := &FileDiff{
		OldPath: "main.go",
		NewPath: "main.go",
		Lines: []Line{
			{Type: Context, Content: "package main", OldLineNo: 1, NewLineNo: 1},
			{Type: Context, Content: "", OldLineNo: 2, NewLineNo: 2},
			{Type: Context, Content: "func hello() {", OldLineNo: 3, NewLineNo: 3},
			{Type: Remove, Content: "\tfmt.Println(\"old\")", OldLineNo: 4, NewLineNo: 0},
			{Type: Add, Content: "\tfmt.Println(\"new\")", OldLineNo: 0, NewLineNo: 4},
			{Type: Context, Content: "}", OldLineNo: 5, NewLineNo: 5},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Expected: 3 unchanged, 1 modified, total 5 alignments
	// (but empty line might be handled differently)
	if len(alignments) < 4 {
		t.Fatalf("expected at least 4 alignments, got %d", len(alignments))
	}

	// Should have exactly one ModifiedAlignment
	modifiedCount := 0
	var modifiedAlign ModifiedAlignment
	for _, align := range alignments {
		if ma, ok := align.(ModifiedAlignment); ok {
			modifiedCount++
			modifiedAlign = ma
		}
	}

	if modifiedCount != 1 {
		t.Errorf("expected 1 ModifiedAlignment, got %d", modifiedCount)
	}

	// Verify the modified alignment has inline diff
	if len(modifiedAlign.InlineDiff) == 0 {
		t.Error("expected inline diff in ModifiedAlignment")
	}

	// Verify tokens are preserved
	if len(left.Lines[modifiedAlign.LeftIdx].Tokens) == 0 {
		t.Error("expected tokens in left line")
	}
	if len(right.Lines[modifiedAlign.RightIdx].Tokens) == 0 {
		t.Error("expected tokens in right line")
	}
}

func TestIntegration_EndToEnd_MultipleHunks(t *testing.T) {
	// Multiple separate hunks in the file
	oldContent := "line1\nold2\nline3\nold4\nline5"
	newContent := "line1\nnew2\nline3\nnew4\nline5"

	left, _ := BuildFileContent("test.txt", oldContent)
	right, _ := BuildFileContent("test.txt", newContent)

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "old2", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "new2", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
			{Type: Remove, Content: "old4", OldLineNo: 4, NewLineNo: 0},
			{Type: Add, Content: "new4", OldLineNo: 0, NewLineNo: 4},
			{Type: Context, Content: "line5", OldLineNo: 5, NewLineNo: 5},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Note: "old2" and "new2" have low similarity (only "2" in common), so they won't pair
	// Similarly, "old4" and "new4" have low similarity
	// Expected: Unchanged, Removed, Added, Unchanged, Removed, Added, Unchanged
	// Total: 7 alignments
	if len(alignments) != 7 {
		t.Fatalf("expected 7 alignments, got %d", len(alignments))
	}

	// Verify we have the expected types at key positions
	// First: unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Errorf("alignment 0: expected UnchangedAlignment, got %T", alignments[0])
	}

	// Last: unchanged
	lastIdx := len(alignments) - 1
	if _, ok := alignments[lastIdx].(UnchangedAlignment); !ok {
		t.Errorf("alignment %d: expected UnchangedAlignment, got %T", lastIdx, alignments[lastIdx])
	}

	// Should have Removed and Added (not Modified) since similarity is too low
	hasRemoved := false
	hasAdded := false
	for _, align := range alignments {
		switch align.(type) {
		case RemovedAlignment:
			hasRemoved = true
		case AddedAlignment:
			hasAdded = true
		}
	}

	if !hasRemoved {
		t.Error("expected at least one RemovedAlignment")
	}
	if !hasAdded {
		t.Error("expected at least one AddedAlignment")
	}
}

func TestIntegration_EndToEnd_MixedOperations(t *testing.T) {
	// File with additions, removals, modifications, and unchanged lines
	oldContent := "unchanged1\nremoved\nmodified_old\nunchanged2"
	newContent := "unchanged1\nmodified_new\nadded\nunchanged2"

	left, _ := BuildFileContent("test.txt", oldContent)
	right, _ := BuildFileContent("test.txt", newContent)

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "unchanged1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "removed", OldLineNo: 2, NewLineNo: 0},
			{Type: Remove, Content: "modified_old", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "modified_new", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "added", OldLineNo: 0, NewLineNo: 3},
			{Type: Context, Content: "unchanged2", OldLineNo: 4, NewLineNo: 4},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Verify we have a mix of alignment types
	hasUnchanged := false
	hasModified := false
	hasRemoved := false
	hasAdded := false

	for _, align := range alignments {
		switch align.(type) {
		case UnchangedAlignment:
			hasUnchanged = true
		case ModifiedAlignment:
			hasModified = true
		case RemovedAlignment:
			hasRemoved = true
		case AddedAlignment:
			hasAdded = true
		}
	}

	if !hasUnchanged {
		t.Error("expected at least one UnchangedAlignment")
	}
	if !hasModified {
		t.Error("expected at least one ModifiedAlignment")
	}
	if !hasRemoved {
		t.Error("expected at least one RemovedAlignment")
	}
	if !hasAdded {
		t.Error("expected at least one AddedAlignment")
	}

	// Verify first and last are unchanged
	if _, ok := alignments[0].(UnchangedAlignment); !ok {
		t.Error("expected first alignment to be UnchangedAlignment")
	}
	lastIdx := len(alignments) - 1
	if _, ok := alignments[lastIdx].(UnchangedAlignment); !ok {
		t.Error("expected last alignment to be UnchangedAlignment")
	}
}

// ═══════════════════════════════════════════════════════════
// Helper Test - Verify FileContent.LineNo
// ═══════════════════════════════════════════════════════════

func TestFileContent_LineNo(t *testing.T) {
	fc := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
		},
	}

	// LineNo should return 1-indexed line numbers
	tests := []struct {
		idx      int
		expected int
	}{
		{0, 1},
		{1, 2},
		{2, 3},
	}

	for _, tt := range tests {
		actual := fc.LineNo(tt.idx)
		if actual != tt.expected {
			t.Errorf("LineNo(%d): expected %d, got %d", tt.idx, tt.expected, actual)
		}
	}
}

// ═══════════════════════════════════════════════════════════
// Edge Cases
// ═══════════════════════════════════════════════════════════

func TestBuildAlignments_OnlyRemovals(t *testing.T) {
	// File with only removed lines (file deleted)
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
		},
	}
	right := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "line1", OldLineNo: 1, NewLineNo: 0},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce 2 RemovedAlignment entries
	if len(alignments) != 2 {
		t.Fatalf("expected 2 alignments, got %d", len(alignments))
	}

	for i, align := range alignments {
		if _, ok := align.(RemovedAlignment); !ok {
			t.Errorf("alignment %d: expected RemovedAlignment, got %T", i, align)
		}
	}
}

func TestBuildAlignments_OnlyAdditions(t *testing.T) {
	// File with only added lines (new file)
	left := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Add, Content: "line1", OldLineNo: 0, NewLineNo: 1},
			{Type: Add, Content: "line2", OldLineNo: 0, NewLineNo: 2},
		},
	}

	alignments := BuildAlignments(left, right, parsedDiff)

	// Should produce 2 AddedAlignment entries
	if len(alignments) != 2 {
		t.Fatalf("expected 2 alignments, got %d", len(alignments))
	}

	for i, align := range alignments {
		if _, ok := align.(AddedAlignment); !ok {
			t.Errorf("alignment %d: expected AddedAlignment, got %T", i, align)
		}
	}
}
