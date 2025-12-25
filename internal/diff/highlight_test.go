// ABOUTME: Tests for syntax highlighting integration in diff merging
// ABOUTME: Tests ApplySyntaxHighlighting function with various file types and edge cases

package diff

import (
	"testing"
)

func TestApplySyntaxHighlighting_GoFile(t *testing.T) {
	oldContent := "package main\n\nfunc old() {\n}\n"
	newContent := "package main\n\nfunc new() {\n}\n"

	// Create a simple diff with one change
	parsedDiff := &FileDiff{
		OldPath: "main.go",
		NewPath: "main.go",
		Lines: []Line{
			{Type: Context, Content: "package main", OldLineNo: 1, NewLineNo: 1},
			{Type: Context, Content: "", OldLineNo: 2, NewLineNo: 2},
			{Type: Remove, Content: "func old() {", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "func new() {", OldLineNo: 0, NewLineNo: 3},
			{Type: Context, Content: "}", OldLineNo: 4, NewLineNo: 4},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "main.go")

	// Verify all lines have tokens populated
	for i, line := range result.Lines {
		if line.LeftLineNo > 0 {
			// Left side should have tokens
			if line.LeftTokens == nil {
				t.Errorf("Line[%d].LeftTokens is nil, expected tokens", i)
			}
		} else {
			// Added lines should have empty left tokens
			if len(line.LeftTokens) != 0 {
				t.Errorf("Line[%d].LeftTokens = %v, want empty slice for added line", i, line.LeftTokens)
			}
		}

		if line.RightLineNo > 0 {
			// Right side should have tokens
			if line.RightTokens == nil {
				t.Errorf("Line[%d].RightTokens is nil, expected tokens", i)
			}
		} else {
			// Removed lines should have empty right tokens
			if len(line.RightTokens) != 0 {
				t.Errorf("Line[%d].RightTokens = %v, want empty slice for removed line", i, line.RightTokens)
			}
		}
	}

	// Verify the first line has tokens (package main should be tokenized)
	if len(result.Lines) > 0 {
		firstLine := result.Lines[0]
		if len(firstLine.LeftTokens) == 0 {
			t.Error("First line LeftTokens is empty, expected tokens for 'package main'")
		}
		if len(firstLine.RightTokens) == 0 {
			t.Error("First line RightTokens is empty, expected tokens for 'package main'")
		}
	}
}

func TestApplySyntaxHighlighting_UnsupportedFile(t *testing.T) {
	oldContent := "some text\nanother line\n"
	newContent := "some text\nmodified line\n"

	parsedDiff := &FileDiff{
		OldPath: "file.txt",
		NewPath: "file.txt",
		Lines: []Line{
			{Type: Context, Content: "some text", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "another line", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "modified line", OldLineNo: 0, NewLineNo: 2},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "file.unknown")

	// Even unsupported files should have tokens (Text tokens)
	// Tokens should always be populated (even if empty slice for empty lines)
	for i, line := range result.Lines {
		if line.LeftLineNo > 0 && line.LeftTokens == nil {
			t.Errorf("Line[%d].LeftTokens is nil, should be populated", i)
		}
		if line.RightLineNo > 0 && line.RightTokens == nil {
			t.Errorf("Line[%d].RightTokens is nil, should be populated", i)
		}
	}
}

func TestApplySyntaxHighlighting_EmptyFile(t *testing.T) {
	oldContent := ""
	newContent := ""

	parsedDiff := &FileDiff{
		OldPath: "empty.go",
		NewPath: "empty.go",
		Lines:   []Line{},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "empty.go")

	// Should not crash with empty file
	if len(result.Lines) != 0 {
		t.Errorf("Expected 0 lines for empty file, got %d", len(result.Lines))
	}
}

func TestApplySyntaxHighlighting_NewFile(t *testing.T) {
	oldContent := ""
	newContent := "package main\n\nfunc main() {\n}\n"

	parsedDiff := &FileDiff{
		OldPath: "new.go",
		NewPath: "new.go",
		Lines: []Line{
			{Type: Add, Content: "package main", OldLineNo: 0, NewLineNo: 1},
			{Type: Add, Content: "", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "func main() {", OldLineNo: 0, NewLineNo: 3},
			{Type: Add, Content: "}", OldLineNo: 0, NewLineNo: 4},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "new.go")

	// All lines should be Added with empty LeftTokens and populated RightTokens
	for i, line := range result.Lines {
		if line.Change != Added {
			t.Errorf("Line[%d].Change = %v, want Added", i, line.Change)
		}
		if len(line.LeftTokens) != 0 {
			t.Errorf("Line[%d].LeftTokens should be empty for new file", i)
		}
		if line.RightLineNo > 0 && line.RightTokens == nil {
			t.Errorf("Line[%d].RightTokens is nil, should be populated", i)
		}
	}
}

func TestApplySyntaxHighlighting_DeletedFile(t *testing.T) {
	oldContent := "package main\n\nfunc main() {\n}\n"
	newContent := ""

	parsedDiff := &FileDiff{
		OldPath: "deleted.go",
		NewPath: "deleted.go",
		Lines: []Line{
			{Type: Remove, Content: "package main", OldLineNo: 1, NewLineNo: 0},
			{Type: Remove, Content: "", OldLineNo: 2, NewLineNo: 0},
			{Type: Remove, Content: "func main() {", OldLineNo: 3, NewLineNo: 0},
			{Type: Remove, Content: "}", OldLineNo: 4, NewLineNo: 0},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "deleted.go")

	// All lines should be Removed with populated LeftTokens and empty RightTokens
	for i, line := range result.Lines {
		if line.Change != Removed {
			t.Errorf("Line[%d].Change = %v, want Removed", i, line.Change)
		}
		if line.LeftLineNo > 0 && line.LeftTokens == nil {
			t.Errorf("Line[%d].LeftTokens is nil, should be populated", i)
		}
		if len(line.RightTokens) != 0 {
			t.Errorf("Line[%d].RightTokens should be empty for deleted file", i)
		}
	}
}

func TestApplySyntaxHighlighting_TokensMatchContent(t *testing.T) {
	oldContent := "package main\n"
	newContent := "package main\n"

	parsedDiff := &FileDiff{
		OldPath: "test.go",
		NewPath: "test.go",
		Lines: []Line{
			{Type: Context, Content: "package main", OldLineNo: 1, NewLineNo: 1},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)
	ApplySyntaxHighlighting(result, oldContent, newContent, "test.go")

	// Verify that concatenating tokens gives us back the original content
	if len(result.Lines) > 0 {
		line := result.Lines[0]
		expectedContent := "package main"

		// Reconstruct content from left tokens
		var leftReconstructed string
		for _, token := range line.LeftTokens {
			leftReconstructed += token.Value
		}
		if leftReconstructed != expectedContent {
			t.Errorf("Left tokens don't match content: got %q, want %q", leftReconstructed, expectedContent)
		}

		// Reconstruct content from right tokens
		var rightReconstructed string
		for _, token := range line.RightTokens {
			rightReconstructed += token.Value
		}
		if rightReconstructed != expectedContent {
			t.Errorf("Right tokens don't match content: got %q, want %q", rightReconstructed, expectedContent)
		}
	}
}

func TestApplySyntaxHighlighting_LineNumberOutOfRange(t *testing.T) {
	// Test edge case where line numbers might be out of range
	oldContent := "line1\n"
	newContent := "line1\nline2\n"

	// Malformed diff that references line numbers beyond content
	parsedDiff := &FileDiff{
		OldPath: "test.go",
		NewPath: "test.go",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Add, Content: "line2", OldLineNo: 0, NewLineNo: 2},
		},
	}

	result := MergeFullFile(oldContent, newContent, parsedDiff)

	// Manually create a line with out-of-range line numbers
	result.Lines = append(result.Lines, FullFileLine{
		LeftLineNo:  999, // Out of range
		RightLineNo: 999, // Out of range
		Change:      Unchanged,
	})

	// Should not panic
	ApplySyntaxHighlighting(result, oldContent, newContent, "test.go")

	// Out of range lines should have empty tokens
	outOfRangeLine := result.Lines[len(result.Lines)-1]
	if len(outOfRangeLine.LeftTokens) != 0 {
		t.Errorf("Out of range line should have empty LeftTokens, got %v", outOfRangeLine.LeftTokens)
	}
	if len(outOfRangeLine.RightTokens) != 0 {
		t.Errorf("Out of range line should have empty RightTokens, got %v", outOfRangeLine.RightTokens)
	}
}

// Helper test to verify token type is correct
func TestApplySyntaxHighlighting_TokenTypes(t *testing.T) {
	content := "package main\n"

	parsedDiff := &FileDiff{
		OldPath: "test.go",
		NewPath: "test.go",
		Lines: []Line{
			{Type: Context, Content: "package main", OldLineNo: 1, NewLineNo: 1},
		},
	}

	result := MergeFullFile(content, content, parsedDiff)
	ApplySyntaxHighlighting(result, content, content, "test.go")

	if len(result.Lines) > 0 {
		line := result.Lines[0]

		// Should have at least 2 tokens: "package" (keyword) and "main" (name)
		if len(line.LeftTokens) < 2 {
			t.Errorf("Expected at least 2 tokens for 'package main', got %d", len(line.LeftTokens))
		}

		// Verify we found the "package" token
		hasPackage := false
		for _, token := range line.LeftTokens {
			if token.Value == "package" {
				hasPackage = true
				break
			}
		}

		if !hasPackage {
			t.Error("Expected to find 'package' keyword token")
		}
	}
}
