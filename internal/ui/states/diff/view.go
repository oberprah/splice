package diff

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// View renders the diff state
func (s *State) View(ctx app.Context) app.ViewRenderer {
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
	if s.Diff == nil || len(s.Diff.Alignments) == 0 {
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

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+availableHeight, len(s.Diff.Alignments))

	// Create styles for fixed-width columns
	leftColStyle := lipgloss.NewStyle().Width(columnWidth)
	rightColStyle := lipgloss.NewStyle().Width(columnWidth)

	// Build left and right columns independently
	leftVb := components.NewViewBuilder()
	rightVb := components.NewViewBuilder()

	// Render visible alignments into separate ViewBuilders
	for i := s.ViewportStart; i < viewportEnd; i++ {
		alignment := s.Diff.Alignments[i]
		left, right := s.renderAlignment(alignment, columnWidth, lineNoWidth)

		// Apply fixed width styling to each line before adding to ViewBuilders
		leftVb.AddLine(leftColStyle.Render(left))
		rightVb.AddLine(rightColStyle.Render(right))
	}

	// Compose the split view
	vb.AddSplitView(leftVb, rightVb)

	return vb
}

// renderHeader formats the diff view header
func (s *State) renderHeader() string {
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
func (s *State) calculateLineNoWidth() int {
	if s.Diff == nil {
		return 3
	}

	maxLineNo := 0
	for _, alignment := range s.Diff.Alignments {
		switch a := alignment.(type) {
		case diff.UnchangedAlignment:
			leftNo := s.Diff.Left.LineNo(a.LeftIdx)
			rightNo := s.Diff.Right.LineNo(a.RightIdx)
			if leftNo > maxLineNo {
				maxLineNo = leftNo
			}
			if rightNo > maxLineNo {
				maxLineNo = rightNo
			}
		case diff.ModifiedAlignment:
			leftNo := s.Diff.Left.LineNo(a.LeftIdx)
			rightNo := s.Diff.Right.LineNo(a.RightIdx)
			if leftNo > maxLineNo {
				maxLineNo = leftNo
			}
			if rightNo > maxLineNo {
				maxLineNo = rightNo
			}
		case diff.RemovedAlignment:
			leftNo := s.Diff.Left.LineNo(a.LeftIdx)
			if leftNo > maxLineNo {
				maxLineNo = leftNo
			}
		case diff.AddedAlignment:
			rightNo := s.Diff.Right.LineNo(a.RightIdx)
			if rightNo > maxLineNo {
				maxLineNo = rightNo
			}
		}
	}
	width := len(fmt.Sprintf("%d", maxLineNo))
	if width < 3 {
		width = 3
	}
	return width
}

// renderAlignment returns the left and right column content for an alignment using type switch
func (s *State) renderAlignment(alignment diff.Alignment, columnWidth, lineNoWidth int) (string, string) {
	// Calculate content width (column width - lineNo - space - indicator - space)
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}

	switch a := alignment.(type) {
	case diff.UnchangedAlignment:
		// Unchanged: show on both sides with normal styling
		leftLine := s.Diff.Left.Lines[a.LeftIdx]
		rightLine := s.Diff.Right.Lines[a.RightIdx]
		leftNo := s.Diff.Left.LineNo(a.LeftIdx)
		rightNo := s.Diff.Right.LineNo(a.RightIdx)
		left := s.formatColumnContent(leftNo, " ", leftLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle, nil)
		right := s.formatColumnContent(rightNo, " ", rightLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle, nil)
		return left, right

	case diff.ModifiedAlignment:
		// Modified: show on both sides with inline diff highlighting
		leftLine := s.Diff.Left.Lines[a.LeftIdx]
		rightLine := s.Diff.Right.Lines[a.RightIdx]
		leftNo := s.Diff.Left.LineNo(a.LeftIdx)
		rightNo := s.Diff.Right.LineNo(a.RightIdx)
		left := s.formatColumnContent(leftNo, "-", leftLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffDeletionsStyle, a.InlineDiff)
		right := s.formatColumnContent(rightNo, "+", rightLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffAdditionsStyle, a.InlineDiff)
		return left, right

	case diff.RemovedAlignment:
		// Removed: show on left only with deletion style
		leftLine := s.Diff.Left.Lines[a.LeftIdx]
		leftNo := s.Diff.Left.LineNo(a.LeftIdx)
		left := s.formatColumnContent(leftNo, "-", leftLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffDeletionsStyle, nil)
		return left, ""

	case diff.AddedAlignment:
		// Added: show on right only with addition style
		rightLine := s.Diff.Right.Lines[a.RightIdx]
		rightNo := s.Diff.Right.LineNo(a.RightIdx)
		right := s.formatColumnContent(rightNo, "+", rightLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.DiffAdditionsStyle, nil)
		return "", right
	}

	return "", ""
}

// formatColumnContent formats a single column with line number, indicator, and tokens
// Tokens are rendered with syntax highlighting (foreground colors) and then wrapped
// with the background style for diff changes.
// If inlineDiff is provided, it applies inline highlighting for modified lines.
func (s *State) formatColumnContent(lineNo int, indicator string, tokens []highlight.Token, lineNoWidth, contentWidth, columnWidth int, bgStyle lipgloss.Style, inlineDiff []diffmatchpatch.Diff) string {
	// Format line number (blank if 0)
	var lineNoStr string
	if lineNo == 0 {
		lineNoStr = strings.Repeat(" ", lineNoWidth)
	} else {
		lineNoStr = fmt.Sprintf("%*d", lineNoWidth, lineNo)
	}

	// Render tokens with syntax highlighting (foreground) and diff background
	// The background is applied to each character during rendering
	var renderedContent string
	if inlineDiff != nil {
		// Modified line: apply inline highlighting
		renderedContent = s.renderTokensWithInlineDiff(tokens, contentWidth, bgStyle, inlineDiff, indicator == "+")
	} else {
		// Unchanged, pure added, or pure removed: normal rendering
		renderedContent = s.renderTokens(tokens, contentWidth, bgStyle)
	}

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

// renderTokensWithInlineDiff renders tokens with inline diff highlighting for modified lines.
// It breaks syntax tokens at inline diff boundaries and applies brighter backgrounds for changed portions.
// isRightSide determines whether we're rendering the left (removed) or right (added) side.
func (s *State) renderTokensWithInlineDiff(tokens []highlight.Token, maxWidth int, bgStyle lipgloss.Style, inlineDiff []diffmatchpatch.Diff, isRightSide bool) string {
	if len(tokens) == 0 {
		return ""
	}

	// Get the brighter style for changed portions
	var brightStyle lipgloss.Style
	if isRightSide {
		brightStyle = styles.DiffAdditionsBrightStyle
	} else {
		brightStyle = styles.DiffDeletionsBrightStyle
	}

	// Build highlight map from inline diff
	// We need to account for tab expansion when building the map, since tabs
	// in the inline diff text will be expanded to 4 spaces during rendering.
	highlightMap := make(map[int]bool) // position -> isHighlighted
	currentPos := 0
	for _, diff := range inlineDiff {
		if isRightSide {
			// Right side: highlight DiffInsert, show DiffEqual normally, skip DiffDelete
			switch diff.Type {
			case diffmatchpatch.DiffEqual:
				// Iterate through runes, expanding tabs
				for _, r := range diff.Text {
					if r == '\t' {
						// Tab expands to 4 spaces
						for i := 0; i < 4; i++ {
							highlightMap[currentPos] = false
							currentPos++
						}
					} else {
						highlightMap[currentPos] = false
						currentPos++
					}
				}
			case diffmatchpatch.DiffInsert:
				// Iterate through runes, expanding tabs
				for _, r := range diff.Text {
					if r == '\t' {
						// Tab expands to 4 spaces
						for i := 0; i < 4; i++ {
							highlightMap[currentPos] = true
							currentPos++
						}
					} else {
						highlightMap[currentPos] = true
						currentPos++
					}
				}
			case diffmatchpatch.DiffDelete:
				// Skip deleted text on right side
			}
		} else {
			// Left side: highlight DiffDelete, show DiffEqual normally, skip DiffInsert
			switch diff.Type {
			case diffmatchpatch.DiffEqual:
				// Iterate through runes, expanding tabs
				for _, r := range diff.Text {
					if r == '\t' {
						// Tab expands to 4 spaces
						for i := 0; i < 4; i++ {
							highlightMap[currentPos] = false
							currentPos++
						}
					} else {
						highlightMap[currentPos] = false
						currentPos++
					}
				}
			case diffmatchpatch.DiffDelete:
				// Iterate through runes, expanding tabs
				for _, r := range diff.Text {
					if r == '\t' {
						// Tab expands to 4 spaces
						for i := 0; i < 4; i++ {
							highlightMap[currentPos] = true
							currentPos++
						}
					} else {
						highlightMap[currentPos] = true
						currentPos++
					}
				}
			case diffmatchpatch.DiffInsert:
				// Skip inserted text on left side
			}
		}
	}

	// Now render tokens with inline highlighting
	var result strings.Builder
	visibleWidth := 0
	charPos := 0

	for _, token := range tokens {
		// Expand tabs to spaces before processing
		expandedValue := expandTabs(token.Value, 4)

		for _, r := range expandedValue {
			if visibleWidth >= maxWidth {
				// We've reached the max width, append ellipsis and stop
				if visibleWidth == maxWidth {
					result.WriteString("…")
				}
				return result.String()
			}

			// Determine which background to use
			var charBgStyle lipgloss.Style
			if highlighted, exists := highlightMap[charPos]; exists && highlighted {
				charBgStyle = brightStyle
			} else {
				charBgStyle = bgStyle
			}

			// Apply syntax highlighting style (foreground) with appropriate background
			syntaxStyle := highlight.StyleForToken(token.Type)
			combinedStyle := syntaxStyle.Inherit(charBgStyle)
			result.WriteString(combinedStyle.Render(string(r)))
			visibleWidth++
			charPos++
		}
	}

	return result.String()
}

// expandTabs replaces tab characters with spaces
func expandTabs(s string, tabWidth int) string {
	return strings.ReplaceAll(s, "\t", strings.Repeat(" ", tabWidth))
}
