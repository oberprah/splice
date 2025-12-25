// ABOUTME: Tests for the diff merge algorithm that combines full file content with diff information.
// ABOUTME: Tests cover simple modifications, multiple hunks, new/deleted files, and edge cases.

package diff

import (
	"reflect"
	"testing"
)

func TestMergeFullFile_SimpleModification(t *testing.T) {
	oldContent := "line1\nline2\nline3\n"
	newContent := "line1\nmodified\nline3\n"

	// Simulate a diff that changes line2
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "modified", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	// Should have 4 lines
	if len(result.Lines) != 4 {
		t.Errorf("MergeFullFile() lines count = %d, want 4", len(result.Lines))
	}

	// Verify line types
	expectedTypes := []ChangeType{Unchanged, Removed, Added, Unchanged}
	for i, want := range expectedTypes {
		if i >= len(result.Lines) {
			break
		}
		if result.Lines[i].Change != want {
			t.Errorf("Line[%d].Change = %v, want %v", i, result.Lines[i].Change, want)
		}
	}

	// Verify ChangeIndices contains the changed lines (indices 1 and 2)
	if len(result.ChangeIndices) != 2 {
		t.Errorf("ChangeIndices count = %d, want 2", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_MultipleHunks(t *testing.T) {
	oldContent := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n"
	newContent := "line1\nmodified2\nline3\nline4\nline5\nline6\nline7\nmodified8\nline9\nline10\n"

	// Diff has changes at line 2 and line 8
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "modified2", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
			// ... gap in diff ...
			{Type: Context, Content: "line7", OldLineNo: 7, NewLineNo: 7},
			{Type: Remove, Content: "line8", OldLineNo: 8, NewLineNo: 0},
			{Type: Add, Content: "modified8", OldLineNo: 0, NewLineNo: 8},
			{Type: Context, Content: "line9", OldLineNo: 9, NewLineNo: 9},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	// Should have 12 lines (10 content + 2 removed that pair with 2 added)
	// Actually: 10 unchanged/context lines from new file + modifications
	// line1 (unchanged), line2 (removed), modified2 (added), line3-7 (unchanged),
	// line8 (removed), modified8 (added), line9-10 (unchanged)
	// = 1 + 1 + 1 + 5 + 1 + 1 + 2 = 12 lines
	if len(result.Lines) != 12 {
		t.Errorf("MergeFullFile() lines count = %d, want 12", len(result.Lines))
	}

	// Should have 4 change indices (2 removed + 2 added)
	if len(result.ChangeIndices) != 4 {
		t.Errorf("ChangeIndices count = %d, want 4", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_NewFile(t *testing.T) {
	oldContent := ""
	newContent := "line1\nline2\nline3\n"

	// All lines are additions for a new file
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Add, Content: "line1", OldLineNo: 0, NewLineNo: 1},
			{Type: Add, Content: "line2", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "line3", OldLineNo: 0, NewLineNo: 3},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	// All 3 lines should be Added
	if len(result.Lines) != 3 {
		t.Errorf("MergeFullFile() lines count = %d, want 3", len(result.Lines))
	}

	for i, line := range result.Lines {
		if line.Change != Added {
			t.Errorf("Line[%d].Change = %v, want Added", i, line.Change)
		}
		if line.LeftLineNo != 0 {
			t.Errorf("Line[%d].LeftLineNo = %d, want 0 (no left side for new file)", i, line.LeftLineNo)
		}
	}

	// All lines are changes
	if len(result.ChangeIndices) != 3 {
		t.Errorf("ChangeIndices count = %d, want 3", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_DeletedFile(t *testing.T) {
	oldContent := "line1\nline2\nline3\n"
	newContent := ""

	// All lines are deletions for a deleted file
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Remove, Content: "line1", OldLineNo: 1, NewLineNo: 0},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
			{Type: Remove, Content: "line3", OldLineNo: 3, NewLineNo: 0},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	// All 3 lines should be Removed
	if len(result.Lines) != 3 {
		t.Errorf("MergeFullFile() lines count = %d, want 3", len(result.Lines))
	}

	for i, line := range result.Lines {
		if line.Change != Removed {
			t.Errorf("Line[%d].Change = %v, want Removed", i, line.Change)
		}
		if line.RightLineNo != 0 {
			t.Errorf("Line[%d].RightLineNo = %d, want 0 (no right side for deleted file)", i, line.RightLineNo)
		}
	}

	// All lines are changes
	if len(result.ChangeIndices) != 3 {
		t.Errorf("ChangeIndices count = %d, want 3", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_UnchangedFile(t *testing.T) {
	content := "line1\nline2\nline3\n"

	// No changes in diff (empty diff)
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines:   []Line{},
	}

	result := MergeFullFile(content, content, parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	// All 3 lines should be Unchanged
	if len(result.Lines) != 3 {
		t.Errorf("MergeFullFile() lines count = %d, want 3", len(result.Lines))
	}

	for i, line := range result.Lines {
		if line.Change != Unchanged {
			t.Errorf("Line[%d].Change = %v, want Unchanged", i, line.Change)
		}
	}

	// No change indices for unchanged file
	if len(result.ChangeIndices) != 0 {
		t.Errorf("ChangeIndices count = %d, want 0", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_EmptyFiles(t *testing.T) {
	// Both files empty
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines:   []Line{},
	}

	result := MergeFullFile("", "", parsedDiff)

	if result == nil {
		t.Fatal("MergeFullFile() returned nil")
	}

	if len(result.Lines) != 0 {
		t.Errorf("MergeFullFile() lines count = %d, want 0", len(result.Lines))
	}

	if len(result.ChangeIndices) != 0 {
		t.Errorf("ChangeIndices count = %d, want 0", len(result.ChangeIndices))
	}
}

func TestMergeFullFile_LineNumbers(t *testing.T) {
	oldContent := "a\nb\nc\n"
	newContent := "a\nX\nc\n"

	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Context, Content: "a", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "b", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "X", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "c", OldLineNo: 3, NewLineNo: 3},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	// Verify line count
	if len(result.Lines) != 4 {
		t.Fatalf("Lines count = %d, want 4", len(result.Lines))
	}

	// Verify each line's structure (line numbers and change type)
	tests := []struct {
		idx         int
		leftLineNo  int
		rightLineNo int
		change      ChangeType
	}{
		{0, 1, 1, Unchanged},
		{1, 2, 0, Removed},
		{2, 0, 2, Added},
		{3, 3, 3, Unchanged},
	}

	for _, tt := range tests {
		got := result.Lines[tt.idx]
		if got.LeftLineNo != tt.leftLineNo || got.RightLineNo != tt.rightLineNo || got.Change != tt.change {
			t.Errorf("Line[%d] = {LeftLineNo: %d, RightLineNo: %d, Change: %v}, want {LeftLineNo: %d, RightLineNo: %d, Change: %v}",
				tt.idx, got.LeftLineNo, got.RightLineNo, got.Change,
				tt.leftLineNo, tt.rightLineNo, tt.change)
		}
	}
}

func TestMergeFullFile_ChangeIndicesCorrect(t *testing.T) {
	oldContent := "line1\nline2\nline3\nline4\nline5\n"
	newContent := "line1\nmodified\nline3\nline4\nadded\nline5\n"

	// Change at line2, addition after line4
	parsedDiff := &FileDiff{
		OldPath: "file.go",
		NewPath: "file.go",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "modified", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
			{Type: Context, Content: "line4", OldLineNo: 4, NewLineNo: 4},
			{Type: Add, Content: "added", OldLineNo: 0, NewLineNo: 5},
			{Type: Context, Content: "line5", OldLineNo: 5, NewLineNo: 6},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	// Verify ChangeIndices contains indices of changed lines
	// Lines: line1(0), line2-removed(1), modified-added(2), line3(3), line4(4), added(5), line5(6)
	// Changes at indices: 1, 2, 5
	expectedIndices := []int{1, 2, 5}
	if !reflect.DeepEqual(result.ChangeIndices, expectedIndices) {
		t.Errorf("ChangeIndices = %v, want %v", result.ChangeIndices, expectedIndices)
	}
}
