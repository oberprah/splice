package components

import (
	"strings"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// TestCommitInfo tests the complete CommitInfo function
func TestCommitInfo_BasicCommit(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add new feature",
		Body:    "",
		Author:  "Alice",
		Date:    fixedTime,
		Refs:    []core.RefInfo{},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfo(commit, 80, 0, ctx)

	// Should have: metadata, blank, subject
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	// Check structure
	if lines[1] != "" {
		t.Errorf("Expected blank line at position 1, got: %q", lines[1])
	}
}

func TestCommitInfo_WithRefs(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add new feature",
		Body:    "",
		Author:  "Alice",
		Date:    fixedTime,
		Refs: []core.RefInfo{
			{Name: "main", Type: core.RefTypeBranch, IsHead: true},
			{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
			{Name: "HEAD", Type: core.RefTypeBranch, IsHead: true},
		},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfo(commit, 80, 0, ctx)

	// Should have: metadata, refs, blank, subject
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines, got %d", len(lines))
	}

	// Refs line should be comma-separated
	refsLine := stripANSI(lines[1])
	if !strings.Contains(refsLine, "main") || !strings.Contains(refsLine, "origin/main") {
		t.Errorf("Refs line missing expected refs: %q", refsLine)
	}
}

func TestCommitInfo_WithBody(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add new feature",
		Body:    "This is a detailed explanation.\n\nIt has multiple paragraphs.",
		Author:  "Alice",
		Date:    fixedTime,
		Refs:    []core.RefInfo{},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfo(commit, 80, 0, ctx)

	// Should have: metadata, blank, subject, blank, body lines (3)
	if len(lines) < 6 {
		t.Errorf("Expected at least 6 lines (with body), got %d", len(lines))
	}

	// Check blank lines at correct positions
	if lines[1] != "" {
		t.Errorf("Expected blank line after metadata, got: %q", lines[1])
	}
	if lines[3] != "" {
		t.Errorf("Expected blank line before body, got: %q", lines[3])
	}
}

func TestCommitInfo_BodyTruncation(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Subject",
		Body:    "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7\nLine 8",
		Author:  "Alice",
		Date:    fixedTime,
		Refs:    []core.RefInfo{},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfo(commit, 80, 5, ctx) // Limit to 5 body lines

	// Should have: metadata, blank, subject, blank, 5 body lines, truncation indicator
	expectedLines := 1 + 1 + 1 + 1 + 5 + 1
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}

	// Last line should be truncation indicator
	lastLine := stripANSI(lines[len(lines)-1])
	if !strings.Contains(lastLine, "more lines") {
		t.Errorf("Expected truncation indicator, got: %q", lastLine)
	}
	if !strings.Contains(lastLine, "3 more lines") {
		t.Errorf("Expected '3 more lines', got: %q", lastLine)
	}
}

func TestCommitInfo_BodyTruncationWithWrapping(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	// Create a body with one long line that will wrap to multiple lines
	longLine := strings.Repeat("word ", 50) // Creates a very long line
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Subject",
		Body:    longLine + "\nShort line",
		Author:  "Alice",
		Date:    fixedTime,
		Refs:    []core.RefInfo{},
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfo(commit, 40, 3, ctx) // Narrow width, limit to 3 lines

	// The long line should wrap to multiple lines and be truncated
	lastLine := stripANSI(lines[len(lines)-1])
	if !strings.Contains(lastLine, "more lines") {
		t.Errorf("Expected truncation indicator when wrapping causes overflow")
	}
}

// TestRenderMetadataLine tests metadata line rendering and smart truncation
func TestRenderMetadataLine_Full(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:   "abc123def456789012345678901234567890abcd",
		Author: "Alice",
		Date:   fixedTime,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	line := renderMetadataLine(commit, 80, ctx)

	// Strip ANSI codes to check plain text
	plainText := stripANSI(line)

	// Should contain all components
	if !strings.Contains(plainText, "abc123d") {
		t.Errorf("Missing hash in: %q", plainText)
	}
	if !strings.Contains(plainText, "Alice") {
		t.Errorf("Missing author in: %q", plainText)
	}
	if !strings.Contains(plainText, "committed") {
		t.Errorf("Missing 'committed' in: %q", plainText)
	}
	if !strings.Contains(plainText, "ago") {
		t.Errorf("Missing time in: %q", plainText)
	}
}

func TestRenderMetadataLine_TruncateAuthor(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:   "abc123def456789012345678901234567890abcd",
		Author: "Very Long Author Name That Needs Truncation",
		Date:   fixedTime,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	line := renderMetadataLine(commit, 50, ctx) // Narrow width forces truncation

	plainText := stripANSI(line)

	// Should contain truncated author with ellipsis
	if !strings.Contains(plainText, "…") {
		t.Errorf("Expected ellipsis for truncated author in: %q", plainText)
	}
	// Should still contain hash and time
	if !strings.Contains(plainText, "abc123d") {
		t.Errorf("Missing hash in: %q", plainText)
	}
	// Should still have "committed"
	if !strings.Contains(plainText, "committed") {
		t.Errorf("Missing 'committed' in: %q", plainText)
	}
}

func TestRenderMetadataLine_DropCommitted(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:   "abc123def456789012345678901234567890abcd",
		Author: "Very Long Author Name",
		Date:   fixedTime,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	line := renderMetadataLine(commit, 35, ctx) // Very narrow width

	plainText := stripANSI(line)

	// Should drop "committed" but keep hash, truncated author, and time
	if strings.Contains(plainText, "committed") {
		t.Errorf("Should not contain 'committed' at this width: %q", plainText)
	}
	if !strings.Contains(plainText, "abc123d") {
		t.Errorf("Missing hash in: %q", plainText)
	}
	if !strings.Contains(plainText, "…") {
		t.Errorf("Expected truncated author in: %q", plainText)
	}
}

func TestRenderMetadataLine_DropAuthor(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:   "abc123def456789012345678901234567890abcd",
		Author: "Alice",
		Date:   fixedTime,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	line := renderMetadataLine(commit, 20, ctx) // Extremely narrow width

	plainText := stripANSI(line)

	// Should only have hash and time
	if !strings.Contains(plainText, "abc123d") {
		t.Errorf("Missing hash in: %q", plainText)
	}
	if strings.Contains(plainText, "Alice") {
		t.Errorf("Should not contain author at this width: %q", plainText)
	}
	if strings.Contains(plainText, "committed") {
		t.Errorf("Should not contain 'committed' at this width: %q", plainText)
	}
}

func TestRenderMetadataLine_UTF8Author(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:   "abc123def456789012345678901234567890abcd",
		Author: "José García", // UTF-8 characters
		Date:   fixedTime,
	}

	ctx := testutils.MockContext{W: 80, H: 24}
	line := renderMetadataLine(commit, 80, ctx)

	plainText := stripANSI(line)

	// Should handle UTF-8 correctly
	if !strings.Contains(plainText, "José García") {
		t.Errorf("UTF-8 author name not preserved: %q", plainText)
	}
}

// TestTruncateWithEllipsis tests the truncation helper
func TestTruncateWithEllipsis_NoTruncation(t *testing.T) {
	result := truncateWithEllipsis("short", 10)
	if result != "short" {
		t.Errorf("Expected 'short', got %q", result)
	}
}

func TestTruncateWithEllipsis_ExactFit(t *testing.T) {
	result := truncateWithEllipsis("exact", 5)
	if result != "exact" {
		t.Errorf("Expected 'exact', got %q", result)
	}
}

func TestTruncateWithEllipsis_Truncate(t *testing.T) {
	result := truncateWithEllipsis("verylongtext", 8)
	if result != "verylon…" {
		t.Errorf("Expected 'verylon…', got %q", result)
	}
	// Verify length is correct (counting runes)
	if len([]rune(result)) != 8 {
		t.Errorf("Expected 8 runes, got %d", len([]rune(result)))
	}
}

func TestTruncateWithEllipsis_UTF8(t *testing.T) {
	result := truncateWithEllipsis("José García Martínez", 10)
	expected := "José Garc…"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
	// Verify it counts runes, not bytes
	if len([]rune(result)) != 10 {
		t.Errorf("Expected 10 runes, got %d", len([]rune(result)))
	}
}

func TestTruncateWithEllipsis_SingleChar(t *testing.T) {
	result := truncateWithEllipsis("text", 1)
	if result != "…" {
		t.Errorf("Expected '…', got %q", result)
	}
}

func TestTruncateWithEllipsis_ZeroWidth(t *testing.T) {
	result := truncateWithEllipsis("text", 0)
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

// TestRenderRefsLines tests ref rendering
func TestRenderRefsLines_SingleRef(t *testing.T) {
	refs := []core.RefInfo{
		{Name: "main", Type: core.RefTypeBranch, IsHead: true},
	}

	lines := renderRefsLines(refs, 80)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	plainText := stripANSI(lines[0])
	if plainText != "main" {
		t.Errorf("Expected 'main', got %q", plainText)
	}
}

func TestRenderRefsLines_MultipleRefs(t *testing.T) {
	refs := []core.RefInfo{
		{Name: "main", Type: core.RefTypeBranch, IsHead: true},
		{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
		{Name: "HEAD", Type: core.RefTypeBranch, IsHead: true},
	}

	lines := renderRefsLines(refs, 80)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	plainText := stripANSI(lines[0])
	expected := "main, origin/main, HEAD"
	if plainText != expected {
		t.Errorf("Expected %q, got %q", expected, plainText)
	}
}

func TestRenderRefsLines_Wrapping(t *testing.T) {
	// Create many refs that will need wrapping
	refs := []core.RefInfo{
		{Name: "main", Type: core.RefTypeBranch, IsHead: true},
		{Name: "origin/main", Type: core.RefTypeRemoteBranch, IsHead: false},
		{Name: "feature-branch-one", Type: core.RefTypeBranch, IsHead: false},
		{Name: "feature-branch-two", Type: core.RefTypeBranch, IsHead: false},
	}

	lines := renderRefsLines(refs, 30) // Narrow width forces wrapping

	if len(lines) < 2 {
		t.Errorf("Expected multiple lines due to wrapping, got %d", len(lines))
	}
}

// TestRenderBodyLines tests body rendering
func TestRenderBodyLines_SingleLine(t *testing.T) {
	body := "This is a single line body."
	lines := renderBodyLines(body, 80, 0)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	plainText := stripANSI(lines[0])
	if plainText != body {
		t.Errorf("Expected %q, got %q", body, plainText)
	}
}

func TestRenderBodyLines_MultipleLines(t *testing.T) {
	body := "Line 1\nLine 2\nLine 3"
	lines := renderBodyLines(body, 80, 0)

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if stripANSI(lines[0]) != "Line 1" {
		t.Errorf("Expected 'Line 1', got %q", lines[0])
	}
}

func TestRenderBodyLines_WithBlankLines(t *testing.T) {
	body := "Paragraph 1\n\nParagraph 2"
	lines := renderBodyLines(body, 80, 0)

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines (including blank), got %d", len(lines))
	}

	if lines[1] != "" {
		t.Errorf("Expected blank line at position 1, got %q", lines[1])
	}
}

func TestRenderBodyLines_Wrapping(t *testing.T) {
	// Long line that will wrap
	body := "This is a very long line that should wrap when the width is narrow enough to force wrapping behavior"
	lines := renderBodyLines(body, 30, 0)

	if len(lines) <= 1 {
		t.Errorf("Expected multiple lines due to wrapping, got %d", len(lines))
	}
}

func TestRenderBodyLines_Truncation(t *testing.T) {
	body := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7"
	lines := renderBodyLines(body, 80, 3)

	// Should have 3 lines + truncation indicator
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (3 + indicator), got %d", len(lines))
	}

	lastLine := stripANSI(lines[3])
	if !strings.Contains(lastLine, "4 more lines") {
		t.Errorf("Expected '4 more lines' in truncation indicator, got %q", lastLine)
	}
}

func TestRenderBodyLines_NoTruncationWhenUnderLimit(t *testing.T) {
	body := "Line 1\nLine 2"
	lines := renderBodyLines(body, 80, 5)

	// Should have exactly 2 lines, no truncation indicator
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines (no truncation), got %d", len(lines))
	}

	// Last line should not be truncation indicator
	lastLine := stripANSI(lines[1])
	if strings.Contains(lastLine, "more lines") {
		t.Errorf("Should not have truncation indicator: %q", lastLine)
	}
}

// TestWrapText tests the text wrapping function
func TestWrapText_FitsOneLine(t *testing.T) {
	text := "short text"
	lines := wrapText(text, 80)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}
	if lines[0] != text {
		t.Errorf("Expected %q, got %q", text, lines[0])
	}
}

func TestWrapText_MultipleLines(t *testing.T) {
	text := "word1 word2 word3 word4 word5"
	lines := wrapText(text, 15)

	if len(lines) < 2 {
		t.Errorf("Expected multiple lines, got %d", len(lines))
	}

	// Each line should be <= 15 characters
	for i, line := range lines {
		if len([]rune(line)) > 15 {
			t.Errorf("Line %d exceeds width: %q", i, line)
		}
	}
}

func TestWrapText_EmptyString(t *testing.T) {
	text := ""
	lines := wrapText(text, 80)

	if len(lines) != 1 {
		t.Errorf("Expected 1 line for empty string, got %d", len(lines))
	}
	if lines[0] != "" {
		t.Errorf("Expected empty string, got %q", lines[0])
	}
}

func TestWrapText_SingleLongWord(t *testing.T) {
	text := "verylongwordthatcannotbewrapped"
	lines := wrapText(text, 10)

	// Should put it on one line even if it exceeds width
	// (we don't break words)
	if len(lines) != 1 {
		t.Errorf("Expected 1 line (word cannot be broken), got %d", len(lines))
	}
}

func TestWrapText_UTF8(t *testing.T) {
	text := "José García Martínez lives in España"
	lines := wrapText(text, 20)

	// Should handle UTF-8 correctly when measuring width
	if len(lines) < 2 {
		t.Errorf("Expected wrapping with UTF-8 text, got %d lines", len(lines))
	}
}

func TestWrapText_ZeroWidth(t *testing.T) {
	text := "some text"
	lines := wrapText(text, 0)

	// Should return original text when width is 0
	if len(lines) != 1 || lines[0] != text {
		t.Errorf("Expected original text for zero width, got %v", lines)
	}
}

// TestCommitInfoFromRange tests the CommitInfoFromRange function
func TestCommitInfoFromRange_SingleCommit(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	commit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "Add new feature",
		Body:    "This is the body",
		Author:  "Alice",
		Date:    fixedTime,
		Refs:    []core.RefInfo{},
	}

	commitRange := core.NewSingleCommitRange(commit)
	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfoFromRange(commitRange, 80, 0, ctx)

	// For single commit, should delegate to CommitInfo
	// Should have: metadata, blank, subject, blank, body
	if len(lines) < 5 {
		t.Errorf("Expected at least 5 lines (delegated to CommitInfo), got %d", len(lines))
	}

	// Check that first line contains hash (metadata)
	firstLine := stripANSI(lines[0])
	if !strings.Contains(firstLine, "abc123d") {
		t.Errorf("Expected hash in first line, got: %q", firstLine)
	}
}

func TestCommitInfoFromRange_MultipleCommits(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	startCommit := core.GitCommit{
		Hash:    "abc123def456789012345678901234567890abcd",
		Message: "First commit",
		Author:  "Alice",
		Date:    fixedTime,
	}
	endCommit := core.GitCommit{
		Hash:    "def456abc123789012345678901234567890def4",
		Message: "Last commit",
		Author:  "Bob",
		Date:    fixedTime,
	}

	commitRange := core.NewCommitRange(startCommit, endCommit, 3)
	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfoFromRange(commitRange, 80, 0, ctx)

	// For range, should only show range header (1 line)
	if len(lines) != 1 {
		t.Errorf("Expected 1 line for range, got %d", len(lines))
	}

	// Check format: abc123d..def456a (3 commits)
	line := stripANSI(lines[0])
	if !strings.Contains(line, "abc123d") {
		t.Errorf("Expected start hash in range header, got: %q", line)
	}
	if !strings.Contains(line, "def456a") {
		t.Errorf("Expected end hash in range header, got: %q", line)
	}
	if !strings.Contains(line, "3 commits") {
		t.Errorf("Expected commit count in range header, got: %q", line)
	}
	if !strings.Contains(line, "..") {
		t.Errorf("Expected '..' separator in range header, got: %q", line)
	}
}

func TestCommitInfoFromRange_RangeFormat(t *testing.T) {
	fixedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	startCommit := core.GitCommit{
		Hash:   "1234567890123456789012345678901234567890",
		Author: "Alice",
		Date:   fixedTime,
	}
	endCommit := core.GitCommit{
		Hash:   "abcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Author: "Bob",
		Date:   fixedTime,
	}

	commitRange := core.NewCommitRange(startCommit, endCommit, 5)
	ctx := testutils.MockContext{W: 80, H: 24}
	lines := CommitInfoFromRange(commitRange, 80, 0, ctx)

	line := stripANSI(lines[0])
	expected := "1234567..abcdefa (5 commits)"
	if line != expected {
		t.Errorf("Expected %q, got %q", expected, line)
	}
}
