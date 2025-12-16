package states

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// View renders the diff state
func (s *DiffState) View(ctx Context) string {
	var b strings.Builder

	// Render header
	header := s.renderHeader()
	b.WriteString(header)

	// Render separator
	separator := strings.Repeat("─", min(ctx.Width(), 80))
	b.WriteString(styles.HeaderStyle.Render(separator))
	b.WriteString("\n")

	// Calculate available height for diff content
	headerLines := strings.Count(header, "\n") + 1 // +1 for separator
	availableHeight := max(ctx.Height()-headerLines, 1)

	// Render diff content
	if len(s.Diff.Lines) == 0 {
		b.WriteString(styles.TimeStyle.Render("No changes"))
		b.WriteString("\n")
		return b.String()
	}

	// Calculate column width (each column gets half the terminal width minus separator)
	columnWidth := (ctx.Width() - 3) / 2 // -3 for " │ " separator
	if columnWidth < 20 {
		columnWidth = 20
	}

	// Calculate line number width based on max line numbers
	lineNoWidth := s.calculateLineNoWidth()

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+availableHeight, len(s.Diff.Lines))

	// Create styles for fixed-width columns
	leftColStyle := lipgloss.NewStyle().Width(columnWidth)
	rightColStyle := lipgloss.NewStyle().Width(columnWidth)
	separatorStyle := styles.HeaderStyle

	// Render visible diff lines
	for i := s.ViewportStart; i < viewportEnd; i++ {
		line := s.Diff.Lines[i]
		left, right := s.renderDiffLineParts(line, columnWidth, lineNoWidth)

		// Use Lip Gloss to join columns - it handles width properly with ANSI codes
		row := lipgloss.JoinHorizontal(
			lipgloss.Top,
			leftColStyle.Render(left),
			separatorStyle.Render(" │ "),
			rightColStyle.Render(right),
		)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader formats the diff view header
func (s *DiffState) renderHeader() string {
	// Format: abc123d · path/to/file.go · +15 -8
	var b strings.Builder

	b.WriteString(styles.HashStyle.Render(format.ToShortHash(s.Commit.Hash)))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.FilePathStyle.Render(s.File.Path))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.AdditionsStyle.Render(fmt.Sprintf("+%d", s.File.Additions)))
	b.WriteString(styles.HeaderStyle.Render(" "))
	b.WriteString(styles.DeletionsStyle.Render(fmt.Sprintf("-%d", s.File.Deletions)))
	b.WriteString("\n")

	return b.String()
}

// calculateLineNoWidth returns the width needed for line numbers
func (s *DiffState) calculateLineNoWidth() int {
	maxLineNo := 0
	for _, line := range s.Diff.Lines {
		if line.OldLineNo > maxLineNo {
			maxLineNo = line.OldLineNo
		}
		if line.NewLineNo > maxLineNo {
			maxLineNo = line.NewLineNo
		}
	}
	width := len(fmt.Sprintf("%d", maxLineNo))
	if width < 3 {
		width = 3
	}
	return width
}

// renderDiffLineParts returns the left and right column content for a diff line
func (s *DiffState) renderDiffLineParts(line diff.Line, columnWidth, lineNoWidth int) (string, string) {
	// Calculate content width (column width - lineNo - space - indicator - space)
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}

	switch line.Type {
	case diff.Context:
		// Context: show on both sides
		left := s.formatColumnContent(line.OldLineNo, " ", line.Content, lineNoWidth, contentWidth, styles.TimeStyle)
		right := s.formatColumnContent(line.NewLineNo, " ", line.Content, lineNoWidth, contentWidth, styles.TimeStyle)
		return left, right
	case diff.Remove:
		// Remove: show on left only
		left := s.formatColumnContent(line.OldLineNo, "-", line.Content, lineNoWidth, contentWidth, styles.DeletionsStyle)
		return left, ""
	case diff.Add:
		// Add: show on right only
		right := s.formatColumnContent(line.NewLineNo, "+", line.Content, lineNoWidth, contentWidth, styles.AdditionsStyle)
		return "", right
	}

	return "", ""
}

// formatColumnContent formats a single column with line number, indicator, and content
func (s *DiffState) formatColumnContent(lineNo int, indicator, content string, lineNoWidth, contentWidth int, style lipgloss.Style) string {
	// Format line number
	lineNoStr := fmt.Sprintf("%*d", lineNoWidth, lineNo)

	// Convert tabs to spaces for consistent width
	expandedContent := expandTabs(content, 4)

	// Truncate content if needed
	truncated := truncateWithEllipsis(expandedContent, contentWidth)

	// Build the column string: "123 - content"
	columnStr := lineNoStr + " " + indicator + " " + truncated

	// Apply styling
	return style.Render(columnStr)
}

// truncateWithEllipsis truncates a string and adds ellipsis if too long
func truncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	// Convert to runes to handle unicode properly
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}

	if maxWidth <= 1 {
		return "…"
	}

	return string(runes[:maxWidth-1]) + "…"
}

// expandTabs replaces tab characters with spaces
func expandTabs(s string, tabWidth int) string {
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
}
