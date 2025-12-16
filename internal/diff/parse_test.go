package diff

import (
	"reflect"
	"testing"
)

func TestParseUnifiedDiff_SimpleHunk(t *testing.T) {
	// Note: empty context lines in diff have a space prefix, not truly empty
	raw := "diff --git a/file.go b/file.go\n" +
		"index abc123..def456 100644\n" +
		"--- a/file.go\n" +
		"+++ b/file.go\n" +
		"@@ -1,4 +1,4 @@\n" +
		" package main\n" +
		" \n" + // empty context line (space prefix)
		"-func old() {}\n" +
		"+func new() {}\n" +
		" \n" // empty context line (space prefix)
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if result.OldPath != "file.go" {
		t.Errorf("OldPath = %q, want %q", result.OldPath, "file.go")
	}
	if result.NewPath != "file.go" {
		t.Errorf("NewPath = %q, want %q", result.NewPath, "file.go")
	}

	expectedLines := []Line{
		{Type: Context, Content: "package main", OldLineNo: 1, NewLineNo: 1},
		{Type: Context, Content: "", OldLineNo: 2, NewLineNo: 2},
		{Type: Remove, Content: "func old() {}", OldLineNo: 3, NewLineNo: 0},
		{Type: Add, Content: "func new() {}", OldLineNo: 0, NewLineNo: 3},
		{Type: Context, Content: "", OldLineNo: 4, NewLineNo: 4},
	}

	if len(result.Lines) != len(expectedLines) {
		t.Fatalf("Lines count = %d, want %d", len(result.Lines), len(expectedLines))
	}

	for i, want := range expectedLines {
		got := result.Lines[i]
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Line[%d] = %+v, want %+v", i, got, want)
		}
	}
}

func TestParseUnifiedDiff_MultipleHunks(t *testing.T) {
	raw := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,3 @@
 package main
-var old = 1
+var new = 1
@@ -10,3 +10,3 @@
 func test() {
-    return old
+    return new
 }
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	// First hunk starts at line 1
	if result.Lines[0].OldLineNo != 1 {
		t.Errorf("First hunk line OldLineNo = %d, want 1", result.Lines[0].OldLineNo)
	}

	// Find the start of second hunk (line 10)
	foundSecondHunk := false
	for _, line := range result.Lines {
		if line.OldLineNo == 10 || line.NewLineNo == 10 {
			foundSecondHunk = true
			break
		}
	}
	if !foundSecondHunk {
		t.Error("Second hunk starting at line 10 not found")
	}
}

func TestParseUnifiedDiff_AdditionsOnly(t *testing.T) {
	raw := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,2 +1,4 @@
 line1
+added1
+added2
 line2
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	addCount := 0
	for _, line := range result.Lines {
		if line.Type == Add {
			addCount++
		}
	}
	if addCount != 2 {
		t.Errorf("Add lines count = %d, want 2", addCount)
	}
}

func TestParseUnifiedDiff_DeletionsOnly(t *testing.T) {
	raw := `diff --git a/file.go b/file.go
index abc123..def456 100644
--- a/file.go
+++ b/file.go
@@ -1,4 +1,2 @@
 line1
-deleted1
-deleted2
 line2
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	removeCount := 0
	for _, line := range result.Lines {
		if line.Type == Remove {
			removeCount++
		}
	}
	if removeCount != 2 {
		t.Errorf("Remove lines count = %d, want 2", removeCount)
	}
}

func TestParseUnifiedDiff_LineNumbers(t *testing.T) {
	raw := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -5,4 +5,5 @@
 context at 5
-removed at 6
+added at 6
+added at 7
 context at 7->8
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	tests := []struct {
		idx       int
		wantType  LineType
		wantOld   int
		wantNew   int
	}{
		{0, Context, 5, 5},  // context at 5
		{1, Remove, 6, 0},   // removed at 6
		{2, Add, 0, 6},      // added at 6
		{3, Add, 0, 7},      // added at 7
		{4, Context, 7, 8},  // context at 7->8
	}

	for _, tt := range tests {
		if tt.idx >= len(result.Lines) {
			t.Errorf("Line[%d] not found, only %d lines", tt.idx, len(result.Lines))
			continue
		}
		got := result.Lines[tt.idx]
		if got.Type != tt.wantType {
			t.Errorf("Line[%d].Type = %v, want %v", tt.idx, got.Type, tt.wantType)
		}
		if got.OldLineNo != tt.wantOld {
			t.Errorf("Line[%d].OldLineNo = %d, want %d", tt.idx, got.OldLineNo, tt.wantOld)
		}
		if got.NewLineNo != tt.wantNew {
			t.Errorf("Line[%d].NewLineNo = %d, want %d", tt.idx, got.NewLineNo, tt.wantNew)
		}
	}
}

func TestParseUnifiedDiff_NoNewlineAtEOF(t *testing.T) {
	raw := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1,2 +1,2 @@
 line1
-old
\ No newline at end of file
+new
\ No newline at end of file
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	// Should have parsed the lines without including "\ No newline" markers
	for _, line := range result.Lines {
		if line.Content == "\\ No newline at end of file" {
			t.Error("Should not include 'No newline at end of file' as a line")
		}
	}
}

func TestParseUnifiedDiff_EmptyDiff(t *testing.T) {
	raw := ``
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(result.Lines) != 0 {
		t.Errorf("Lines count = %d, want 0", len(result.Lines))
	}
}

func TestParseUnifiedDiff_HunkHeaderWithoutCount(t *testing.T) {
	// Hunk header without count means count=1: "@@ -1 +1 @@"
	raw := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1 +1 @@
-old
+new
`
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	if len(result.Lines) != 2 {
		t.Errorf("Lines count = %d, want 2", len(result.Lines))
	}
	if result.Lines[0].OldLineNo != 1 {
		t.Errorf("First line OldLineNo = %d, want 1", result.Lines[0].OldLineNo)
	}
}

func TestParseUnifiedDiff_BlankLinesWithPrefix(t *testing.T) {
	// Blank lines in source code have a prefix in unified diff format:
	// - " " (space) for context blank lines
	// - "+" for added blank lines
	// - "-" for removed blank lines
	raw := "diff --git a/file.go b/file.go\n" +
		"--- a/file.go\n" +
		"+++ b/file.go\n" +
		"@@ -1,4 +1,5 @@\n" +
		" line1\n" +
		" \n" + // context blank line (space prefix)
		"-old\n" +
		"+new\n" +
		"+\n" + // added blank line (+ prefix, no content)
		" line4\n"
	result, err := ParseUnifiedDiff(raw)
	if err != nil {
		t.Fatalf("ParseUnifiedDiff() error = %v", err)
	}

	// Should have 6 lines: line1, context blank, old, new, added blank, line4
	if len(result.Lines) != 6 {
		t.Fatalf("Lines count = %d, want 6", len(result.Lines))
	}

	// Verify context blank line
	if result.Lines[1].Type != Context {
		t.Errorf("Line[1].Type = %v, want Context", result.Lines[1].Type)
	}
	if result.Lines[1].Content != "" {
		t.Errorf("Line[1].Content = %q, want empty", result.Lines[1].Content)
	}
	if result.Lines[1].OldLineNo != 2 || result.Lines[1].NewLineNo != 2 {
		t.Errorf("Line[1] line numbers = (%d,%d), want (2,2)", result.Lines[1].OldLineNo, result.Lines[1].NewLineNo)
	}

	// Verify added blank line (inserted after +new, so at new line 4)
	if result.Lines[4].Type != Add {
		t.Errorf("Line[4].Type = %v, want Add", result.Lines[4].Type)
	}
	if result.Lines[4].Content != "" {
		t.Errorf("Line[4].Content = %q, want empty", result.Lines[4].Content)
	}
	if result.Lines[4].OldLineNo != 0 || result.Lines[4].NewLineNo != 4 {
		t.Errorf("Line[4] line numbers = (%d,%d), want (0,4)", result.Lines[4].OldLineNo, result.Lines[4].NewLineNo)
	}
}
