package highlight

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
)

func TestTokenizeFile_GoCode(t *testing.T) {
	content := `package main

func main() {
	println("Hello, World!")
}
`
	tokens := TokenizeFile(content, "main.go")

	// Should have 6 lines (including trailing empty line after final newline)
	if len(tokens) != 6 {
		t.Errorf("Expected 6 lines, got %d", len(tokens))
	}

	// First line should have "package" keyword token
	if len(tokens[0]) == 0 {
		t.Fatal("First line has no tokens")
	}

	foundPackage := false
	for _, token := range tokens[0] {
		if token.Value == "package" {
			foundPackage = true
			// Should be a keyword or namespace token (Chroma may use different types)
			// We just verify it's not Text
			if token.Type == chroma.Text {
				t.Errorf("Expected 'package' to be a keyword token, got Text")
			}
		}
	}
	if !foundPackage {
		t.Error("Did not find 'package' token on first line")
	}

	// Line 3 (index 2) should have "func" keyword
	foundFunc := false
	for _, token := range tokens[2] {
		if token.Value == "func" {
			foundFunc = true
			if token.Type == chroma.Text {
				t.Errorf("Expected 'func' to be a keyword token, got Text")
			}
		}
	}
	if !foundFunc {
		t.Error("Did not find 'func' token on line 3")
	}

	// Line 4 should contain "Hello, World!" string
	foundString := false
	for _, token := range tokens[3] {
		if token.Value == `"Hello, World!"` {
			foundString = true
			// Should be some kind of string token
			if token.Type == chroma.Text {
				t.Errorf("Expected string literal to be a string token, got Text")
			}
		}
	}
	if !foundString {
		t.Error("Did not find string literal on line 4")
	}
}

func TestTokenizeFile_UnsupportedFileType(t *testing.T) {
	content := `This is some text
On multiple lines
With no syntax highlighting
`
	tokens := TokenizeFile(content, "unknown.xyz")

	// Should have 4 lines (including empty line after final newline)
	if len(tokens) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(tokens))
	}

	// Each non-empty line should have exactly one Text token
	for i, line := range tokens[:3] { // Check first 3 lines
		if len(line) != 1 {
			t.Errorf("Line %d: expected 1 token, got %d", i, len(line))
			continue
		}
		if line[0].Type != chroma.Text {
			t.Errorf("Line %d: expected Text token, got %v", i, line[0].Type)
		}
	}

	// Last line (empty) should have no tokens
	if len(tokens[3]) != 0 {
		t.Errorf("Last line: expected 0 tokens, got %d", len(tokens[3]))
	}

	// Verify content is preserved
	if tokens[0][0].Value != "This is some text" {
		t.Errorf("Line 0 content mismatch: got %q", tokens[0][0].Value)
	}
	if tokens[1][0].Value != "On multiple lines" {
		t.Errorf("Line 1 content mismatch: got %q", tokens[1][0].Value)
	}
}

func TestTokenizeFile_EmptyContent(t *testing.T) {
	tokens := TokenizeFile("", "main.go")

	// Empty string split by "\n" gives [""], so one empty line
	if len(tokens) != 1 {
		t.Errorf("Expected 1 line, got %d", len(tokens))
	}

	if len(tokens[0]) != 0 {
		t.Errorf("Expected empty line to have 0 tokens, got %d", len(tokens[0]))
	}
}

func TestTokenizeFile_SingleLine(t *testing.T) {
	content := "package main"
	tokens := TokenizeFile(content, "main.go")

	// Single line without trailing newline
	if len(tokens) != 1 {
		t.Errorf("Expected 1 line, got %d", len(tokens))
	}

	if len(tokens[0]) == 0 {
		t.Fatal("First line has no tokens")
	}

	// Should have "package" token
	foundPackage := false
	for _, token := range tokens[0] {
		if token.Value == "package" {
			foundPackage = true
		}
	}
	if !foundPackage {
		t.Error("Did not find 'package' token")
	}
}

func TestTokenizeFile_MultilineString(t *testing.T) {
	// Go backtick string that spans multiple lines
	content := "package main\n\nvar s = `line 1\nline 2\nline 3`\n"
	tokens := TokenizeFile(content, "main.go")

	// Should have multiple lines
	if len(tokens) < 5 {
		t.Errorf("Expected at least 5 lines, got %d", len(tokens))
	}

	// The string should be split across lines 3-5 (indices 2-4)
	// Line 3 should start with backtick and have "line 1"
	foundStringStart := false
	for _, token := range tokens[2] {
		if token.Value == "`line 1" || token.Value == "line 1" {
			foundStringStart = true
			// Should be a string token
			if token.Type == chroma.Text {
				t.Errorf("Expected multiline string to be a string token, got Text")
			}
		}
	}
	if !foundStringStart {
		t.Error("Did not find multiline string start on line 3")
	}
}

func TestTokenizeFile_JavaScriptCode(t *testing.T) {
	content := `function hello() {
  console.log("Hello");
}
`
	tokens := TokenizeFile(content, "script.js")

	// Should have 4 lines
	if len(tokens) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(tokens))
	}

	// First line should have "function" keyword
	foundFunction := false
	for _, token := range tokens[0] {
		if token.Value == "function" {
			foundFunction = true
			if token.Type == chroma.Text {
				t.Errorf("Expected 'function' to be a keyword token, got Text")
			}
		}
	}
	if !foundFunction {
		t.Error("Did not find 'function' token on first line")
	}
}

func TestTokenizeFile_PythonCode(t *testing.T) {
	content := `def hello():
    print("Hello")
`
	tokens := TokenizeFile(content, "script.py")

	// Should have 3 lines
	if len(tokens) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(tokens))
	}

	// First line should have "def" keyword
	foundDef := false
	for _, token := range tokens[0] {
		if token.Value == "def" {
			foundDef = true
			if token.Type == chroma.Text {
				t.Errorf("Expected 'def' to be a keyword token, got Text")
			}
		}
	}
	if !foundDef {
		t.Error("Did not find 'def' token on first line")
	}
}
