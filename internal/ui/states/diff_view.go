package states

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/highlight"
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

	// Handle nil or empty diff
	if s.Diff == nil || len(s.Diff.Lines) == 0 {
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
		left, right := s.renderFullFileLine(line, columnWidth, lineNoWidth)

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
	if s.Diff == nil {
		return 3
	}

	maxLineNo := 0
	for _, line := range s.Diff.Lines {
		if line.LeftLineNo > maxLineNo {
			maxLineNo = line.LeftLineNo
		}
		if line.RightLineNo > maxLineNo {
			maxLineNo = line.RightLineNo
		}
	}
	width := len(fmt.Sprintf("%d", maxLineNo))
	if width < 3 {
		width = 3
	}
	return width
}

// renderFullFileLine returns the left and right column content for a full file diff line
func (s *DiffState) renderFullFileLine(line diff.FullFileLine, columnWidth, lineNoWidth int) (string, string) {
	// Calculate content width (column width - lineNo - space - indicator - space)
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}

	switch line.Change {
	case diff.Unchanged:
		// Unchanged: show on both sides with normal styling
		left := s.formatColumnContent(line.LeftLineNo, " ", line.LeftTokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle)
		right := s.formatColumnContent(line.RightLineNo, " ", line.RightTokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle)
		return left, right
	case diff.Removed:
		// Removed: show on left only with deletion style
		left := s.formatColumnContent(line.LeftLineNo, "-", line.LeftTokens, lineNoWidth, contentWidth, columnWidth, styles.DiffDeletionsStyle)
		return left, ""
	case diff.Added:
		// Added: show on right only with addition style
		right := s.formatColumnContent(line.RightLineNo, "+", line.RightTokens, lineNoWidth, contentWidth, columnWidth, styles.DiffAdditionsStyle)
		return "", right
	}

	return "", ""
}

// formatColumnContent formats a single column with line number, indicator, and tokens
// Tokens are rendered with syntax highlighting (foreground colors) and then wrapped
// with the background style for diff changes.
func (s *DiffState) formatColumnContent(lineNo int, indicator string, tokens []highlight.Token, lineNoWidth, contentWidth, columnWidth int, bgStyle lipgloss.Style) string {
	// Format line number (blank if 0)
	var lineNoStr string
	if lineNo == 0 {
		lineNoStr = strings.Repeat(" ", lineNoWidth)
	} else {
		lineNoStr = fmt.Sprintf("%*d", lineNoWidth, lineNo)
	}

	// Render tokens with syntax highlighting (foreground) and diff background
	// The background is applied to each character during rendering
	renderedContent := s.renderTokens(tokens, contentWidth, bgStyle)

	// Build the column string: "123 - content"
	// Line number and indicator need background too
	styledLineNo := bgStyle.Render(lineNoStr)
	styledIndicator := bgStyle.Render(" " + indicator + " ")
	columnStr := styledLineNo + styledIndicator + renderedContent

	// Apply full width styling to pad the column
	return bgStyle.Width(columnWidth).Render(columnStr)
}

// renderTokens renders tokens with syntax highlighting and truncates to maxWidth if needed.
// Applies both foreground (syntax) and background (diff) colors to each character.
// Returns the concatenated, styled tokens.
func (s *DiffState) renderTokens(tokens []highlight.Token, maxWidth int, bgStyle lipgloss.Style) string {
	if len(tokens) == 0 {
		return ""
	}

	var result strings.Builder
	visibleWidth := 0

	for _, token := range tokens {
		// Expand tabs to spaces before processing
		expandedValue := expandTabs(token.Value, 4)

		// Convert to runes for proper width calculation
		runes := []rune(expandedValue)

		for _, r := range runes {
			if visibleWidth >= maxWidth {
				// We've reached the max width, append ellipsis and stop
				if visibleWidth == maxWidth {
					result.WriteString("…")
				}
				return result.String()
			}

			// Apply syntax highlighting style (foreground) with diff background
			syntaxStyle := highlight.StyleForToken(token.Type)
			// Combine: copy syntax foreground, then apply diff background
			combinedStyle := syntaxStyle.Copy().Inherit(bgStyle)
			result.WriteString(combinedStyle.Render(string(r)))
			visibleWidth++
		}
	}

	return result.String()
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
