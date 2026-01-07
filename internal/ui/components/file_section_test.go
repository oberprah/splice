package components

import (
	"strings"
	"testing"

	"github.com/oberprah/splice/internal/core"

	"github.com/charmbracelet/x/ansi"
)

// TestFileSection tests the complete FileSection function
func TestFileSection_SingleFile(t *testing.T) {
	files := []core.FileChange{
		{
			Path:      "src/main.go",
			Status:    "M",
			Additions: 10,
			Deletions: 5,
			IsBinary:  false,
		},
	}

	lines := FileSection(files, 80, nil)

	// Should have: blank line, stats line, 1 file line = 3 total
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// First line should be blank
	if lines[0] != "" {
		t.Errorf("Expected blank line at position 0, got: %q", lines[0])
	}

	// Second line should be stats line
	statsLine := ansi.Strip(lines[1])
	if !strings.Contains(statsLine, "1 file") {
		t.Errorf("Expected '1 file' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "+10") {
		t.Errorf("Expected '+10' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "-5") {
		t.Errorf("Expected '-5' in stats line, got: %q", statsLine)
	}

	// Third line should be file line
	fileLine := ansi.Strip(lines[2])
	if !strings.Contains(fileLine, "src/main.go") {
		t.Errorf("Expected 'src/main.go' in file line, got: %q", fileLine)
	}
	if !strings.Contains(fileLine, "M") {
		t.Errorf("Expected 'M' status in file line, got: %q", fileLine)
	}
}

func TestFileSection_MultipleFiles(t *testing.T) {
	files := []core.FileChange{
		{Path: "file1.go", Status: "M", Additions: 10, Deletions: 5, IsBinary: false},
		{Path: "file2.go", Status: "A", Additions: 20, Deletions: 0, IsBinary: false},
		{Path: "file3.go", Status: "D", Additions: 0, Deletions: 15, IsBinary: false},
	}

	lines := FileSection(files, 80, nil)

	// Should have: blank line + stats line + 3 file lines = 5 total
	if len(lines) != 5 {
		t.Errorf("Expected 5 lines, got %d", len(lines))
	}

	// Check stats line totals
	statsLine := ansi.Strip(lines[1])
	if !strings.Contains(statsLine, "3 files") {
		t.Errorf("Expected '3 files' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "+30") { // 10 + 20 + 0 = 30
		t.Errorf("Expected '+30' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "-20") { // 5 + 0 + 15 = 20
		t.Errorf("Expected '-20' in stats line, got: %q", statsLine)
	}

	// Check file lines exist
	for i, expectedPath := range []string{"file1.go", "file2.go", "file3.go"} {
		fileLine := ansi.Strip(lines[2+i])
		if !strings.Contains(fileLine, expectedPath) {
			t.Errorf("Expected %q in file line %d, got: %q", expectedPath, i, fileLine)
		}
	}
}

func TestFileSection_WithSelection(t *testing.T) {
	files := []core.FileChange{
		{Path: "file1.go", Status: "M", Additions: 10, Deletions: 5, IsBinary: false},
		{Path: "file2.go", Status: "A", Additions: 20, Deletions: 0, IsBinary: false},
	}

	cursor := 1
	lines := FileSection(files, 80, &cursor) // Select second file

	// Should have: blank line + stats line + 2 file lines = 4 total
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	// First file line should have space prefix (not selected)
	firstFileLine := ansi.Strip(lines[2])
	if !strings.HasPrefix(firstFileLine, " ") {
		t.Errorf("Expected space prefix for unselected file, got: %q", firstFileLine)
	}

	// Second file line should have "→" prefix (selected)
	secondFileLine := ansi.Strip(lines[3])
	if !strings.HasPrefix(secondFileLine, "→") {
		t.Errorf("Expected '→' prefix for selected file, got: %q", secondFileLine)
	}
}

func TestFileSection_WithoutSelector(t *testing.T) {
	files := []core.FileChange{
		{Path: "file1.go", Status: "M", Additions: 10, Deletions: 5, IsBinary: false},
	}

	lines := FileSection(files, 80, nil) // No selector (nil cursor)

	// File line should NOT have selector prefix
	fileLine := ansi.Strip(lines[2])
	// Should start with status letter, not a space or "→"
	if strings.HasPrefix(fileLine, " ") || strings.HasPrefix(fileLine, "→") {
		t.Errorf("Expected no selector prefix, got: %q", fileLine)
	}
	if !strings.HasPrefix(fileLine, "M") {
		t.Errorf("Expected status letter 'M' to be first, got: %q", fileLine)
	}
}

func TestFileSection_BinaryFile(t *testing.T) {
	files := []core.FileChange{
		{Path: "image.png", Status: "A", Additions: 0, Deletions: 0, IsBinary: true},
	}

	lines := FileSection(files, 80, nil)

	fileLine := ansi.Strip(lines[2])
	if !strings.Contains(fileLine, "(binary)") {
		t.Errorf("Expected '(binary)' marker in file line, got: %q", fileLine)
	}
	// Should not show addition/deletion counts for binary files
	if strings.Contains(fileLine, "+0") {
		t.Errorf("Should not show addition count for binary file, got: %q", fileLine)
	}
}

func TestFileSection_EmptyFileList(t *testing.T) {
	files := []core.FileChange{}

	lines := FileSection(files, 80, nil)

	// Should have: blank line + stats line = 2 total (no file lines)
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines for empty file list, got %d", len(lines))
	}

	statsLine := ansi.Strip(lines[1])
	if !strings.Contains(statsLine, "0 files") {
		t.Errorf("Expected '0 files' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "+0") {
		t.Errorf("Expected '+0' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "-0") {
		t.Errorf("Expected '-0' in stats line, got: %q", statsLine)
	}
}

func TestFileSection_AllFileStatuses(t *testing.T) {
	files := []core.FileChange{
		{Path: "added.go", Status: "A", Additions: 10, Deletions: 0, IsBinary: false},
		{Path: "modified.go", Status: "M", Additions: 5, Deletions: 3, IsBinary: false},
		{Path: "deleted.go", Status: "D", Additions: 0, Deletions: 20, IsBinary: false},
		{Path: "renamed.go", Status: "R", Additions: 0, Deletions: 0, IsBinary: false},
	}

	lines := FileSection(files, 80, nil)

	// Should have: blank line + stats line + 4 file lines = 6 total
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines, got %d", len(lines))
	}

	// Check that all status letters appear
	allLines := strings.Join(lines, "\n")
	allLinesPlain := ansi.Strip(allLines)

	// Should contain all file paths
	for _, path := range []string{"added.go", "modified.go", "deleted.go", "renamed.go"} {
		if !strings.Contains(allLinesPlain, path) {
			t.Errorf("Expected file path %q in output", path)
		}
	}
}

// Test helper functions

func TestCalculateTotalStats_MultipleFiles(t *testing.T) {
	files := []core.FileChange{
		{Additions: 10, Deletions: 5},
		{Additions: 20, Deletions: 15},
		{Additions: 5, Deletions: 0},
	}

	additions, deletions := CalculateTotalStats(files)

	if additions != 35 {
		t.Errorf("Expected 35 additions, got %d", additions)
	}
	if deletions != 20 {
		t.Errorf("Expected 20 deletions, got %d", deletions)
	}
}

func TestCalculateTotalStats_EmptyList(t *testing.T) {
	files := []core.FileChange{}

	additions, deletions := CalculateTotalStats(files)

	if additions != 0 {
		t.Errorf("Expected 0 additions, got %d", additions)
	}
	if deletions != 0 {
		t.Errorf("Expected 0 deletions, got %d", deletions)
	}
}

func TestCalculateMaxStatWidth_SimpleCase(t *testing.T) {
	files := []core.FileChange{
		{Additions: 5, Deletions: 3, IsBinary: false},
		{Additions: 10, Deletions: 8, IsBinary: false},
	}

	addWidth, delWidth := CalculateMaxStatWidth(files)

	// Max additions: 10 (2 digits) + 1 (sign) = 3
	if addWidth != 3 {
		t.Errorf("Expected addWidth 3, got %d", addWidth)
	}
	// Max deletions: 8 (1 digit) + 1 (sign) = 2
	if delWidth != 2 {
		t.Errorf("Expected delWidth 2, got %d", delWidth)
	}
}

func TestCalculateMaxStatWidth_LargeNumbers(t *testing.T) {
	files := []core.FileChange{
		{Additions: 999, Deletions: 1234, IsBinary: false},
		{Additions: 12, Deletions: 5, IsBinary: false},
	}

	addWidth, delWidth := CalculateMaxStatWidth(files)

	// Max additions: 999 (3 digits) + 1 (sign) = 4
	if addWidth != 4 {
		t.Errorf("Expected addWidth 4, got %d", addWidth)
	}
	// Max deletions: 1234 (4 digits) + 1 (sign) = 5
	if delWidth != 5 {
		t.Errorf("Expected delWidth 5, got %d", delWidth)
	}
}

func TestCalculateMaxStatWidth_IgnoresBinaryFiles(t *testing.T) {
	files := []core.FileChange{
		{Additions: 5, Deletions: 3, IsBinary: false},
		{Additions: 99999, Deletions: 99999, IsBinary: true}, // Should be ignored
	}

	addWidth, delWidth := CalculateMaxStatWidth(files)

	// Should only consider non-binary files
	// Max additions: 5 (1 digit) + 1 (sign) = 2
	if addWidth != 2 {
		t.Errorf("Expected addWidth 2 (ignoring binary), got %d", addWidth)
	}
	// Max deletions: 3 (1 digit) + 1 (sign) = 2
	if delWidth != 2 {
		t.Errorf("Expected delWidth 2 (ignoring binary), got %d", delWidth)
	}
}

func TestCalculateMaxStatWidth_MinimumWidth(t *testing.T) {
	files := []core.FileChange{
		{Additions: 0, Deletions: 0, IsBinary: false},
	}

	addWidth, delWidth := CalculateMaxStatWidth(files)

	// Minimum width should be 2 (+0, -0)
	if addWidth != 2 {
		t.Errorf("Expected minimum addWidth 2, got %d", addWidth)
	}
	if delWidth != 2 {
		t.Errorf("Expected minimum delWidth 2, got %d", delWidth)
	}
}

func TestCalculateMaxStatWidth_EmptyList(t *testing.T) {
	files := []core.FileChange{}

	addWidth, delWidth := CalculateMaxStatWidth(files)

	// Should return minimum width even for empty list
	if addWidth != 2 {
		t.Errorf("Expected minimum addWidth 2, got %d", addWidth)
	}
	if delWidth != 2 {
		t.Errorf("Expected minimum delWidth 2, got %d", delWidth)
	}
}

// Test FormatFileLine function

func TestFormatFileLine_BasicFormatting(t *testing.T) {
	file := core.FileChange{
		Path:      "src/main.go",
		Status:    "M",
		Additions: 10,
		Deletions: 5,
		IsBinary:  false,
	}

	params := FormatFileLineParams{
		File:         file,
		IsSelected:   false,
		Width:        80,
		MaxAddWidth:  3,
		MaxDelWidth:  2,
		ShowSelector: false,
	}

	line := FormatFileLine(params)
	plainLine := ansi.Strip(line)

	// Should contain status, stats, and path
	if !strings.Contains(plainLine, "M") {
		t.Errorf("Expected status 'M', got: %q", plainLine)
	}
	if !strings.Contains(plainLine, "+10") {
		t.Errorf("Expected '+10', got: %q", plainLine)
	}
	if !strings.Contains(plainLine, "-5") {
		t.Errorf("Expected '-5', got: %q", plainLine)
	}
	if !strings.Contains(plainLine, "src/main.go") {
		t.Errorf("Expected path 'src/main.go', got: %q", plainLine)
	}
}

func TestFormatFileLine_WithSelector(t *testing.T) {
	file := core.FileChange{
		Path:      "file.go",
		Status:    "M",
		Additions: 1,
		Deletions: 1,
		IsBinary:  false,
	}

	// Selected line with selector
	selectedParams := FormatFileLineParams{
		File:         file,
		IsSelected:   true,
		Width:        80,
		MaxAddWidth:  2,
		MaxDelWidth:  2,
		ShowSelector: true,
	}

	selectedLine := FormatFileLine(selectedParams)
	plainSelected := ansi.Strip(selectedLine)

	if !strings.HasPrefix(plainSelected, "→") {
		t.Errorf("Expected '→' prefix for selected line, got: %q", plainSelected)
	}

	// Unselected line with selector
	unselectedParams := selectedParams
	unselectedParams.IsSelected = false

	unselectedLine := FormatFileLine(unselectedParams)
	plainUnselected := ansi.Strip(unselectedLine)

	if !strings.HasPrefix(plainUnselected, " ") {
		t.Errorf("Expected space prefix for unselected line, got: %q", plainUnselected)
	}
}

func TestFormatFileLine_BinaryFile(t *testing.T) {
	file := core.FileChange{
		Path:      "image.png",
		Status:    "A",
		Additions: 0,
		Deletions: 0,
		IsBinary:  true,
	}

	params := FormatFileLineParams{
		File:         file,
		IsSelected:   false,
		Width:        80,
		MaxAddWidth:  2,
		MaxDelWidth:  2,
		ShowSelector: false,
	}

	line := FormatFileLine(params)
	plainLine := ansi.Strip(line)

	if !strings.Contains(plainLine, "(binary)") {
		t.Errorf("Expected '(binary)' marker, got: %q", plainLine)
	}
	// Should NOT contain addition/deletion counts
	if strings.Contains(plainLine, "+") || strings.Contains(plainLine, "-") {
		t.Errorf("Binary file should not show +/- counts, got: %q", plainLine)
	}
}

func TestFormatFileLine_DefaultStatus(t *testing.T) {
	file := core.FileChange{
		Path:      "file.go",
		Status:    "", // Empty status
		Additions: 1,
		Deletions: 1,
		IsBinary:  false,
	}

	params := FormatFileLineParams{
		File:         file,
		IsSelected:   false,
		Width:        80,
		MaxAddWidth:  2,
		MaxDelWidth:  2,
		ShowSelector: false,
	}

	line := FormatFileLine(params)
	plainLine := ansi.Strip(line)

	// Should default to "M" when status is empty
	if !strings.Contains(plainLine, "M") {
		t.Errorf("Expected default status 'M', got: %q", plainLine)
	}
}

func TestFormatFileLine_WidthAlignment(t *testing.T) {
	files := []core.FileChange{
		{Path: "file1.go", Status: "M", Additions: 5, Deletions: 3, IsBinary: false},
		{Path: "file2.go", Status: "M", Additions: 100, Deletions: 200, IsBinary: false},
	}

	// Calculate max widths
	maxAddWidth, maxDelWidth := CalculateMaxStatWidth(files)

	// Format both lines
	line1 := FormatFileLine(FormatFileLineParams{
		File:         files[0],
		IsSelected:   false,
		Width:        80,
		MaxAddWidth:  maxAddWidth,
		MaxDelWidth:  maxDelWidth,
		ShowSelector: false,
	})

	line2 := FormatFileLine(FormatFileLineParams{
		File:         files[1],
		IsSelected:   false,
		Width:        80,
		MaxAddWidth:  maxAddWidth,
		MaxDelWidth:  maxDelWidth,
		ShowSelector: false,
	})

	plain1 := ansi.Strip(line1)
	plain2 := ansi.Strip(line2)

	// The stats should be padded to align
	// Both should have the same number of characters before the path
	// Extract the part before the path
	idx1 := strings.Index(plain1, "file1.go")
	idx2 := strings.Index(plain2, "file2.go")

	if idx1 != idx2 {
		t.Errorf("Stats columns not aligned: line1 path at %d, line2 path at %d\nLine1: %q\nLine2: %q",
			idx1, idx2, plain1, plain2)
	}
}
