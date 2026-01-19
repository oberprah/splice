package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// View renders the diff state
func (s *State) View(ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Render header
	header := s.renderHeader()
	// Split header into lines and add each line (trimming trailing newline first)
	for _, line := range strings.Split(strings.TrimSuffix(header, "\n"), "\n") {
		vb.AddLine(line)
	}

	// Render separator
	separator := strings.Repeat("─", min(ctx.Width(), 80))
	vb.AddLine(styles.HeaderStyle.Render(separator))

	// Calculate available height for diff content
	headerLines := strings.Count(header, "\n") + 1 // +1 for separator
	availableHeight := max(ctx.Height()-headerLines, 1)

	// Handle nil or empty diff
	if s.Diff == nil || len(s.Diff.Blocks) == 0 {
		vb.AddLine(styles.TimeStyle.Render("No changes"))
		return vb
	}

	// Calculate column width (each column gets half the terminal width minus separator)
	columnWidth := (ctx.Width() - 3) / 2 // -3 for " │ " separator
	if columnWidth < 20 {
		columnWidth = 20
	}

	// Calculate line number width based on max line numbers
	lineNoWidth := s.calculateLineNoWidth()

	// Create styles for fixed-width columns
	leftColStyle := lipgloss.NewStyle().Width(columnWidth)
	rightColStyle := lipgloss.NewStyle().Width(columnWidth)

	// Build left and right columns independently
	leftVb := components.NewViewBuilder()
	rightVb := components.NewViewBuilder()

	// Render visible lines from blocks
	linesRendered := 0
	currentLine := 0 // Global line counter

	for _, block := range s.Diff.Blocks {
		switch b := block.(type) {
		case diff.UnchangedBlock:
			for _, linePair := range b.Lines {
				if currentLine >= s.ViewportStart && linesRendered < availableHeight {
					left, right := s.renderLinePair(linePair, columnWidth, lineNoWidth)
					leftVb.AddLine(leftColStyle.Render(left))
					rightVb.AddLine(rightColStyle.Render(right))
					linesRendered++
				}
				currentLine++
				if linesRendered >= availableHeight {
					break
				}
			}
		case diff.ChangeBlock:
			for _, changeLine := range b.Lines {
				if currentLine >= s.ViewportStart && linesRendered < availableHeight {
					left, right := s.renderChangeLine(changeLine, columnWidth, lineNoWidth)
					leftVb.AddLine(leftColStyle.Render(left))
					rightVb.AddLine(rightColStyle.Render(right))
					linesRendered++
				}
				currentLine++
				if linesRendered >= availableHeight {
					break
				}
			}
		}
		if linesRendered >= availableHeight {
			break
		}
	}

	// Compose the split view
	vb.AddSplitView(leftVb, rightVb)

	return vb
}

// renderHeader formats the diff view header
func (s *State) renderHeader() string {
	// Format for single commit: abc123d · path/to/file.go · +15 -8
	// Format for range: abc123d..def456e · path/to/file.go · +15 -8
	var b strings.Builder

	// Display commit hash or range based on DiffSource type
	switch src := s.Source.(type) {
	case core.CommitRangeDiffSource:
		if src.Count == 1 {
			// Single commit
			b.WriteString(styles.HashStyle.Render(format.ToShortHash(src.End.Hash)))
		} else {
			// Commit range
			startHash := format.ToShortHash(src.Start.Hash)
			endHash := format.ToShortHash(src.End.Hash)
			b.WriteString(styles.HashStyle.Render(fmt.Sprintf("%s..%s", startHash, endHash)))
		}
	case core.UncommittedChangesDiffSource:
		// Display uncommitted changes label based on type
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			b.WriteString(styles.HashStyle.Render("unstaged"))
		case core.UncommittedTypeStaged:
			b.WriteString(styles.HashStyle.Render("staged"))
		case core.UncommittedTypeAll:
			b.WriteString(styles.HashStyle.Render("uncommitted"))
		}
	}

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
func (s *State) calculateLineNoWidth() int {
	if s.Diff == nil {
		return 3
	}

	maxLineNo := 0
	for _, block := range s.Diff.Blocks {
		switch b := block.(type) {
		case diff.UnchangedBlock:
			for _, lp := range b.Lines {
				if lp.LeftLineNo > maxLineNo {
					maxLineNo = lp.LeftLineNo
				}
				if lp.RightLineNo > maxLineNo {
					maxLineNo = lp.RightLineNo
				}
			}
		case diff.ChangeBlock:
			for _, cl := range b.Lines {
				switch line := cl.(type) {
				case diff.RemovedLine:
					if line.LeftLineNo > maxLineNo {
						maxLineNo = line.LeftLineNo
					}
				case diff.AddedLine:
					if line.RightLineNo > maxLineNo {
						maxLineNo = line.RightLineNo
					}
				}
			}
		}
	}

	width := len(fmt.Sprintf("%d", maxLineNo))
	if width < 3 {
		width = 3
	}
	return width
}

// renderLinePair renders an unchanged line pair
func (s *State) renderLinePair(lp diff.LinePair, columnWidth, lineNoWidth int) (string, string) {
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}
	left := s.formatColumnContent(lp.LeftLineNo, " ", lp.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle)
	right := s.formatColumnContent(lp.RightLineNo, " ", lp.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle)
	return left, right
}

// renderChangeLine renders a change line (modified, removed, or added)
func (s *State) renderChangeLine(cl diff.ChangeLine, columnWidth, lineNoWidth int) (string, string) {
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}

	switch line := cl.(type) {
	case diff.RemovedLine:
		left := s.formatColumnContent(line.LeftLineNo, "-", line.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffDeletionsStyle)
		return left, ""
	case diff.AddedLine:
		right := s.formatColumnContent(line.RightLineNo, "+", line.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffAdditionsStyle)
		return "", right
	}
	return "", ""
}

// formatColumnContent formats a single column with line number, indicator, and tokens
// Tokens are rendered with syntax highlighting (foreground colors) and then wrapped
// with the background style for diff changes.
func (s *State) formatColumnContent(lineNo int, indicator string, tokens []highlight.Token, lineNoWidth, contentWidth, columnWidth int, bgStyle lipgloss.Style) string {
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
func (s *State) renderTokens(tokens []highlight.Token, maxWidth int, bgStyle lipgloss.Style) string {
	if len(tokens) == 0 {
		return ""
	}

	var result strings.Builder
	visibleWidth := 0

	for _, token := range tokens {
		// Expand tabs to spaces before processing
		expandedValue := expandTabs(token.Value, 4)

		// Range over string directly for proper rune iteration
		for _, r := range expandedValue {
			if visibleWidth >= maxWidth {
				// We've reached the max width, append ellipsis and stop
				if visibleWidth == maxWidth {
					result.WriteString("…")
				}
				return result.String()
			}

			// Apply syntax highlighting style (foreground) with diff background
			syntaxStyle := highlight.StyleForToken(token.Type)
			// Combine syntax foreground with diff background (assignment is sufficient, no need for deprecated Copy)
			combinedStyle := syntaxStyle.Inherit(bgStyle)
			result.WriteString(combinedStyle.Render(string(r)))
			visibleWidth++
		}
	}

	return result.String()
}

// expandTabs replaces tab characters with spaces
func expandTabs(s string, tabWidth int) string {
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
}
