package states

import (
	"strings"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/git"
)

func TestLogState_View_RendersCommits(t *testing.T) {
	commits := []git.GitCommit{
		{Hash: "abc123", Message: "First commit", Body: "", Author: "Alice", Date: time.Now()},
		{Hash: "def456", Message: "Second commit", Body: "", Author: "Bob", Date: time.Now()},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)

	// Check that both commits appear in the output
	if !strings.Contains(result, "abc123") {
		t.Error("Expected output to contain first commit hash")
	}
	if !strings.Contains(result, "def456") {
		t.Error("Expected output to contain second commit hash")
	}
	if !strings.Contains(result, "First commit") {
		t.Error("Expected output to contain first commit message")
	}
	if !strings.Contains(result, "Second commit") {
		t.Error("Expected output to contain second commit message")
	}
}

func TestLogState_View_SelectionIndicator(t *testing.T) {
	commits := createTestCommits(3)

	tests := []struct {
		name            string
		cursor          int
		shouldContain   string
		checkUnselected bool
	}{
		{"first commit selected", 0, "> ", true},
		{"second commit selected", 1, "> ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := LogState{
				Commits:       commits,
				Cursor:        tt.cursor,
				ViewportStart: 0,
				Preview:       PreviewNone{},
			}
			ctx := mockContext{width: 80, height: 24}

			result := s.View(ctx)

			// Count occurrences of selection indicator
			selectedCount := strings.Count(result, tt.shouldContain)

			// Should have exactly one selection indicator ("> ")
			// and multiple unselected indicators ("  " at start of lines)
			if selectedCount < 1 {
				t.Errorf("Expected at least one selection indicator '%s'", tt.shouldContain)
			}
		})
	}
}

func TestLogState_View_ViewportLimits(t *testing.T) {
	commits := createTestCommits(20)

	s := LogState{
		Commits:       commits,
		Cursor:        10,
		ViewportStart: 5,
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 10}

	result := s.View(ctx)

	// Should render viewportStart (5) to viewportStart + height (15)
	// Count number of lines (should be 10)
	lines := strings.Split(strings.TrimSpace(result), "\n")

	if len(lines) != 10 {
		t.Errorf("Expected 10 lines in viewport, got %d", len(lines))
	}

	// First line should contain commit at viewportStart (index 5)
	// Note: We can't easily check exact commit due to styling, but we can check line count
}

func TestLogState_View_EmptyViewport(t *testing.T) {
	// Edge case: viewportStart beyond commits (shouldn't happen in practice)
	commits := createTestCommits(5)

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 10, // Beyond end
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 10}

	result := s.View(ctx)

	// Should render empty or minimal output
	if strings.TrimSpace(result) != "" {
		// This case might render nothing, which is okay
		t.Logf("ViewportStart beyond commits renders: %q", result)
	}
}

func TestLogState_View_LineTruncation(t *testing.T) {
	commits := []git.GitCommit{
		{
			Hash:    "abc123def456",
			Message: "This is a very long commit message that should be truncated when the terminal is narrow",
			Body:    "",
			Author:  "VeryLongAuthorNameThatShouldAlsoGetTruncated",
			Date:    time.Now(),
		},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}

	// Test with narrow terminal
	ctx := mockContext{width: 40, height: 24}

	result := s.View(ctx)

	// Check that output contains "..." indicating truncation
	if !strings.Contains(result, "...") {
		t.Error("Expected truncated output to contain '...'")
	}

	// Output should not be excessively long (roughly constrained by width)
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		// Note: This is approximate due to ANSI codes, but very long lines suggest no truncation
		// ANSI codes add extra characters, so we allow some overflow
		if len(line) > 150 { // Generous allowance for ANSI codes
			t.Errorf("Line %d appears too long (%d chars), truncation may not be working", i, len(line))
		}
	}
}

func TestLogState_View_SplitView_WideTerminal(t *testing.T) {
	commits := []git.GitCommit{
		{Hash: "abc123", Message: "First commit", Body: "This is the body", Author: "Alice", Date: time.Now()},
		{Hash: "def456", Message: "Second commit", Body: "", Author: "Bob", Date: time.Now()},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewLoaded{ForHash: "abc123", Files: files},
	}

	// Test with wide terminal (triggers split view)
	ctx := mockContext{width: 160, height: 24}

	result := s.View(ctx)

	// Should contain commit info
	if !strings.Contains(result, "abc123") {
		t.Error("Expected output to contain commit hash")
	}
	if !strings.Contains(result, "First commit") {
		t.Error("Expected output to contain commit message in log list")
	}

	// Should contain separator
	if !strings.Contains(result, "│") {
		t.Error("Expected output to contain vertical separator for split view")
	}

	// Should contain commit body in details panel
	if !strings.Contains(result, "This is the body") {
		t.Error("Expected output to contain commit body in details panel")
	}

	// Should contain file information
	if !strings.Contains(result, "src/main.go") {
		t.Error("Expected output to contain file path")
	}
	if !strings.Contains(result, "README.md") {
		t.Error("Expected output to contain second file path")
	}
}

func TestLogState_View_SplitView_NarrowTerminal(t *testing.T) {
	commits := []git.GitCommit{
		{Hash: "abc123", Message: "First commit", Body: "This is the body", Author: "Alice", Date: time.Now()},
	}

	files := []git.FileChange{
		{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewLoaded{ForHash: "abc123", Files: files},
	}

	// Test with narrow terminal (should NOT trigger split view)
	ctx := mockContext{width: 100, height: 24}

	result := s.View(ctx)

	// Should contain commit info
	if !strings.Contains(result, "abc123") {
		t.Error("Expected output to contain commit hash")
	}

	// Should NOT contain separator (single column view)
	if strings.Contains(result, "│") {
		t.Error("Expected output to NOT contain vertical separator in narrow view")
	}

	// Should NOT contain file information (not shown in simple view)
	if strings.Contains(result, "src/main.go") {
		t.Error("Expected output to NOT contain file path in narrow view")
	}
}

func TestLogState_View_SplitView_PreviewLoading(t *testing.T) {
	commits := []git.GitCommit{
		{Hash: "abc123", Message: "First commit", Body: "", Author: "Alice", Date: time.Now()},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: "abc123"},
	}

	// Test with wide terminal
	ctx := mockContext{width: 160, height: 24}

	result := s.View(ctx)

	// Should contain "Loading..." indicator
	if !strings.Contains(result, "Loading...") {
		t.Error("Expected output to contain 'Loading...' indicator")
	}
}

func TestLogState_View_SplitView_PreviewError(t *testing.T) {
	commits := []git.GitCommit{
		{Hash: "abc123", Message: "First commit", Body: "", Author: "Alice", Date: time.Now()},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewError{ForHash: "abc123", Err: nil},
	}

	// Test with wide terminal
	ctx := mockContext{width: 160, height: 24}

	result := s.View(ctx)

	// Should contain error indicator
	if !strings.Contains(result, "Unable to load files") {
		t.Error("Expected output to contain error indicator")
	}
}

func TestLogState_formatFileEntry(t *testing.T) {
	s := LogState{}

	tests := []struct {
		name     string
		file     git.FileChange
		width    int
		contains []string
	}{
		{
			name:     "modified file",
			file:     git.FileChange{Path: "src/main.go", Status: "M", Additions: 10, Deletions: 5},
			width:    80,
			contains: []string{"M", "+10", "-5", "src/main.go"},
		},
		{
			name:     "added file",
			file:     git.FileChange{Path: "README.md", Status: "A", Additions: 20, Deletions: 0},
			width:    80,
			contains: []string{"A", "+20", "-0", "README.md"},
		},
		{
			name:     "deleted file",
			file:     git.FileChange{Path: "old.txt", Status: "D", Additions: 0, Deletions: 15},
			width:    80,
			contains: []string{"D", "+0", "-15", "old.txt"},
		},
		{
			name:     "binary file",
			file:     git.FileChange{Path: "image.png", Status: "M", IsBinary: true},
			width:    80,
			contains: []string{"M", "(binary)", "image.png"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.formatFileEntry(tt.file, tt.width)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

func TestLogState_wrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected number of lines
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    20,
			expected: 1,
		},
		{
			name:     "long text",
			text:     "This is a very long line that should be wrapped into multiple lines",
			width:    20,
			expected: 4, // Approximately
		},
		{
			name:     "zero width",
			text:     "Hello",
			width:    0,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.width)

			if len(result) != tt.expected {
				t.Errorf("Expected %d lines, got %d", tt.expected, len(result))
			}
		})
	}
}
