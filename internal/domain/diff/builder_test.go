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
// BuildFileDiff Tests
// ═══════════════════════════════════════════════════════════

func TestBuildFileDiff_SingleUnchangedBlock(t *testing.T) {
	// All lines unchanged - identical files
	oldContent := "line1\nline2\nline3"
	newContent := "line1\nline2\nline3"
	// Empty diff means no changes (all context lines)
	diffOutput := `@@ -1,3 +1,3 @@
 line1
 line2
 line3`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fd.Path != "test.txt" {
		t.Errorf("expected path 'test.txt', got %q", fd.Path)
	}

	// Should have one UnchangedBlock with 3 lines
	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	unchanged, ok := fd.Blocks[0].(UnchangedBlock)
	if !ok {
		t.Fatalf("expected UnchangedBlock, got %T", fd.Blocks[0])
	}
	if len(unchanged.Lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(unchanged.Lines))
	}

	// Verify line numbers are correct (1-indexed)
	for i, lp := range unchanged.Lines {
		expectedLineNo := i + 1
		if lp.LeftLineNo != expectedLineNo || lp.RightLineNo != expectedLineNo {
			t.Errorf("line %d: expected line numbers (%d, %d), got (%d, %d)",
				i, expectedLineNo, expectedLineNo, lp.LeftLineNo, lp.RightLineNo)
		}
	}
}

func TestBuildFileDiff_SingleChangeBlock_OnlyAdded(t *testing.T) {
	// File with only added lines (new file)
	oldContent := ""
	newContent := "line1\nline2"
	diffOutput := `@@ -0,0 +1,2 @@
+line1
+line2`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have one ChangeBlock with 2 AddedLine
	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[0].(ChangeBlock)
	if !ok {
		t.Fatalf("expected ChangeBlock, got %T", fd.Blocks[0])
	}
	if len(changeBlock.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(changeBlock.Lines))
	}

	// Verify both are AddedLine
	for i, cl := range changeBlock.Lines {
		if _, ok := cl.(AddedLine); !ok {
			t.Errorf("line %d: expected AddedLine, got %T", i, cl)
		}
	}
}

func TestBuildFileDiff_SingleChangeBlock_OnlyRemoved(t *testing.T) {
	// File with only removed lines (deleted file)
	oldContent := "line1\nline2"
	newContent := ""
	diffOutput := `@@ -1,2 +0,0 @@
-line1
-line2`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have one ChangeBlock with 2 RemovedLine
	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[0].(ChangeBlock)
	if !ok {
		t.Fatalf("expected ChangeBlock, got %T", fd.Blocks[0])
	}
	if len(changeBlock.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(changeBlock.Lines))
	}

	// Verify both are RemovedLine
	for i, cl := range changeBlock.Lines {
		if _, ok := cl.(RemovedLine); !ok {
			t.Errorf("line %d: expected RemovedLine, got %T", i, cl)
		}
	}
}

func TestBuildFileDiff_MixedBlocks(t *testing.T) {
	// Unchanged, then changed, then unchanged
	// Using similar text to ensure pairing
	oldContent := "same1\nold_value\nsame2"
	newContent := "same1\nnew_value\nsame2"
	diffOutput := `@@ -1,3 +1,3 @@
 same1
-old_value
+new_value
 same2`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should produce: UnchangedBlock(1), ChangeBlock(1 ModifiedLine), UnchangedBlock(1)
	if len(fd.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(fd.Blocks))
	}

	// First block: unchanged with 1 line
	unchanged1, ok := fd.Blocks[0].(UnchangedBlock)
	if !ok {
		t.Fatalf("block 0: expected UnchangedBlock, got %T", fd.Blocks[0])
	}
	if len(unchanged1.Lines) != 1 {
		t.Errorf("block 0: expected 1 line, got %d", len(unchanged1.Lines))
	}

	// Second block: change with 1 ModifiedLine
	change, ok := fd.Blocks[1].(ChangeBlock)
	if !ok {
		t.Fatalf("block 1: expected ChangeBlock, got %T", fd.Blocks[1])
	}
	if len(change.Lines) != 1 {
		t.Fatalf("block 1: expected 1 line, got %d", len(change.Lines))
	}
	modified, ok := change.Lines[0].(ModifiedLine)
	if !ok {
		t.Fatalf("block 1, line 0: expected ModifiedLine, got %T", change.Lines[0])
	}
	if modified.LeftLineNo != 2 || modified.RightLineNo != 2 {
		t.Errorf("expected line numbers (2, 2), got (%d, %d)", modified.LeftLineNo, modified.RightLineNo)
	}
	// Verify inline diff exists
	if len(modified.InlineDiff) == 0 {
		t.Error("expected inline diff, got none")
	}

	// Third block: unchanged with 1 line
	unchanged2, ok := fd.Blocks[2].(UnchangedBlock)
	if !ok {
		t.Fatalf("block 2: expected UnchangedBlock, got %T", fd.Blocks[2])
	}
	if len(unchanged2.Lines) != 1 {
		t.Errorf("block 2: expected 1 line, got %d", len(unchanged2.Lines))
	}
}

func TestBuildFileDiff_ConsecutiveChanges(t *testing.T) {
	// Multiple consecutive changes should be in one ChangeBlock
	// Lines that don't pair (too dissimilar) become separate RemovedLine/AddedLine
	oldContent := "old1\nold2\nold3"
	newContent := "new1\nnew2\nnew3"
	diffOutput := `@@ -1,3 +1,3 @@
-old1
-old2
-old3
+new1
+new2
+new3`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have one ChangeBlock (all changes are consecutive)
	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[0].(ChangeBlock)
	if !ok {
		t.Fatalf("expected ChangeBlock, got %T", fd.Blocks[0])
	}

	// Should have 6 lines total (3 removed + 3 added, since "old" and "new" are too dissimilar to pair)
	if len(changeBlock.Lines) != 6 {
		t.Errorf("expected 6 lines (3 removed + 3 added), got %d", len(changeBlock.Lines))
	}

	// Count line types
	var removed, added int
	for _, cl := range changeBlock.Lines {
		switch cl.(type) {
		case RemovedLine:
			removed++
		case AddedLine:
			added++
		}
	}

	if removed != 3 {
		t.Errorf("expected 3 RemovedLine, got %d", removed)
	}
	if added != 3 {
		t.Errorf("expected 3 AddedLine, got %d", added)
	}
}

func TestBuildFileDiff_ConsecutiveChanges_WithPairing(t *testing.T) {
	// Consecutive changes that are similar enough to pair
	oldContent := "fmt.Println(hello)\nfmt.Println(world)"
	newContent := "fmt.Println(Hello)\nfmt.Println(World)"
	diffOutput := `@@ -1,2 +1,2 @@
-fmt.Println(hello)
-fmt.Println(world)
+fmt.Println(Hello)
+fmt.Println(World)`

	fd, err := BuildFileDiff("test.go", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have one ChangeBlock
	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[0].(ChangeBlock)
	if !ok {
		t.Fatalf("expected ChangeBlock, got %T", fd.Blocks[0])
	}

	// Should have 2 ModifiedLine (paired due to high similarity)
	if len(changeBlock.Lines) != 2 {
		t.Errorf("expected 2 lines (2 modified), got %d", len(changeBlock.Lines))
	}

	for i, cl := range changeBlock.Lines {
		if _, ok := cl.(ModifiedLine); !ok {
			t.Errorf("line %d: expected ModifiedLine, got %T", i, cl)
		}
	}
}

func TestBuildFileDiff_EmptyDiff(t *testing.T) {
	// Empty files with empty diff
	fd, err := BuildFileDiff("test.txt", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty content produces one empty line due to how strings.Split works
	// strings.Split("", "\n") returns [""], which produces 1 line
	// So we expect 1 block with 1 line (the empty line appearing on both sides)
	if len(fd.Blocks) != 1 {
		t.Errorf("expected 1 block for empty diff, got %d", len(fd.Blocks))
	}

	// Verify TotalLineCount is 1 (the empty line)
	if fd.TotalLineCount() != 1 {
		t.Errorf("expected TotalLineCount 1, got %d", fd.TotalLineCount())
	}
}

func TestBuildFileDiff_MultipleChangeBlocks(t *testing.T) {
	// Changes in multiple non-consecutive locations
	oldContent := "unchanged1\nold_a\nunchanged2\nold_b\nunchanged3"
	newContent := "unchanged1\nnew_a\nunchanged2\nnew_b\nunchanged3"
	diffOutput := `@@ -1,5 +1,5 @@
 unchanged1
-old_a
+new_a
 unchanged2
-old_b
+new_b
 unchanged3`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected structure:
	// UnchangedBlock(1), ChangeBlock(?), UnchangedBlock(1), ChangeBlock(?), UnchangedBlock(1)
	// = 5 blocks total
	if len(fd.Blocks) != 5 {
		t.Fatalf("expected 5 blocks, got %d", len(fd.Blocks))
	}

	// Verify alternating pattern
	for i, block := range fd.Blocks {
		if i%2 == 0 {
			// Even indices should be UnchangedBlock
			if _, ok := block.(UnchangedBlock); !ok {
				t.Errorf("block %d: expected UnchangedBlock, got %T", i, block)
			}
		} else {
			// Odd indices should be ChangeBlock
			if _, ok := block.(ChangeBlock); !ok {
				t.Errorf("block %d: expected ChangeBlock, got %T", i, block)
			}
		}
	}
}

func TestBuildFileDiff_TotalLineCount(t *testing.T) {
	// Verify TotalLineCount works correctly across blocks
	// Using similar text to ensure pairing
	oldContent := "same1\nold_value\nsame2"
	newContent := "same1\nnew_value\nsame2"
	diffOutput := `@@ -1,3 +1,3 @@
 same1
-old_value
+new_value
 same2`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3 blocks: 1 unchanged + 1 change (1 modified line) + 1 unchanged = 3 display lines total
	expected := 3
	if fd.TotalLineCount() != expected {
		t.Errorf("expected TotalLineCount %d, got %d", expected, fd.TotalLineCount())
	}
}

func TestBuildFileDiff_PreservesTokens(t *testing.T) {
	// Verify that tokens are preserved in LinePair and ChangeLine
	oldContent := "package main"
	newContent := "package main\nimport \"fmt\""
	diffOutput := `@@ -1 +1,2 @@
 package main
+import "fmt"`

	fd, err := BuildFileDiff("test.go", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 2 blocks: UnchangedBlock and ChangeBlock
	if len(fd.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(fd.Blocks))
	}

	// First block should have tokens
	unchanged, ok := fd.Blocks[0].(UnchangedBlock)
	if !ok {
		t.Fatalf("block 0: expected UnchangedBlock, got %T", fd.Blocks[0])
	}
	if len(unchanged.Lines[0].Tokens) == 0 {
		t.Error("expected tokens in unchanged line, got none")
	}

	// Second block should have tokens
	change, ok := fd.Blocks[1].(ChangeBlock)
	if !ok {
		t.Fatalf("block 1: expected ChangeBlock, got %T", fd.Blocks[1])
	}
	added, ok := change.Lines[0].(AddedLine)
	if !ok {
		t.Fatalf("block 1, line 0: expected AddedLine, got %T", change.Lines[0])
	}
	if len(added.Tokens) == 0 {
		t.Error("expected tokens in added line, got none")
	}
}

func TestBuildFileDiff_ModifiedLineHasInlineDiff(t *testing.T) {
	// Verify inline diffs are computed for modified lines
	oldContent := "Hello"
	newContent := "Hello World"
	diffOutput := `@@ -1 +1 @@
-Hello
+Hello World`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fd.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[0].(ChangeBlock)
	if !ok {
		t.Fatalf("expected ChangeBlock, got %T", fd.Blocks[0])
	}

	if len(changeBlock.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(changeBlock.Lines))
	}

	modified, ok := changeBlock.Lines[0].(ModifiedLine)
	if !ok {
		t.Fatalf("expected ModifiedLine, got %T", changeBlock.Lines[0])
	}

	// Verify inline diff exists and makes sense
	if len(modified.InlineDiff) == 0 {
		t.Fatal("expected inline diff, got none")
	}

	// Reconstruct text from diffs to verify correctness
	var leftText, rightText string
	for _, d := range modified.InlineDiff {
		switch d.Type {
		case diffmatchpatch.DiffEqual:
			leftText += d.Text
			rightText += d.Text
		case diffmatchpatch.DiffDelete:
			leftText += d.Text
		case diffmatchpatch.DiffInsert:
			rightText += d.Text
		}
	}

	if leftText != "Hello" {
		t.Errorf("expected left text 'Hello', got %q", leftText)
	}
	if rightText != "Hello World" {
		t.Errorf("expected right text 'Hello World', got %q", rightText)
	}
}

func TestBuildFileDiff_MixedChangeTypes(t *testing.T) {
	// A change block with all types: modified, removed, and added
	oldContent := "same\nmodified_old\nremoved\nsame"
	newContent := "same\nmodified_new\nadded\nsame"
	diffOutput := `@@ -1,4 +1,4 @@
 same
-modified_old
-removed
+modified_new
+added
 same`

	fd, err := BuildFileDiff("test.txt", oldContent, newContent, diffOutput)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: UnchangedBlock, ChangeBlock (with mixed types), UnchangedBlock
	if len(fd.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(fd.Blocks))
	}

	changeBlock, ok := fd.Blocks[1].(ChangeBlock)
	if !ok {
		t.Fatalf("block 1: expected ChangeBlock, got %T", fd.Blocks[1])
	}

	// Count types in change block
	var modified, removed, added int
	for _, cl := range changeBlock.Lines {
		switch cl.(type) {
		case ModifiedLine:
			modified++
		case RemovedLine:
			removed++
		case AddedLine:
			added++
		}
	}

	// "modified_old" and "modified_new" should pair (similar), but "removed" and "added" should not
	if modified != 1 {
		t.Errorf("expected 1 ModifiedLine, got %d", modified)
	}
	if removed != 1 {
		t.Errorf("expected 1 RemovedLine, got %d", removed)
	}
	if added != 1 {
		t.Errorf("expected 1 AddedLine, got %d", added)
	}
}
