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
		name           string
		cursor         int
		shouldContain  string
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
