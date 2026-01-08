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
	"github.com/sergi/go-diff/diffmatchpatch"
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
	if s.Diff == nil || (len(s.Diff.Alignments) == 0 && len(s.Diff.Segments) == 0) {
		vb.AddLine(styles.TimeStyle.Render("No changes"))
		return vb
	}

	// Calculate column width (each column gets half the terminal width minus separator)
	columnWidth := (ctx.Width() - 3) / 2 // -3 for " │ " separator
	if columnWidth < 20 {
		columnWidth = 20
	}

	// Use segment-based rendering if segments are available
	if len(s.Diff.Segments) > 0 {
		s.renderWithSegments(vb, availableHeight, columnWidth)
	} else {
		// Fall back to alignment-based rendering for backward compatibility
		s.renderWithAlignments(vb, availableHeight, columnWidth)
	}

	return vb
}

// renderWithSegments renders the diff using segment-based data model.
// This eliminates blank line padding by allowing each panel to scroll independently.
func (s *State) renderWithSegments(vb *components.ViewBuilder, availableHeight, columnWidth int) {
	// Calculate line number width using segments
	lineNoWidth := s.calculateSegmentLineNoWidth()

	// Create styles for fixed-width columns
	leftColStyle := lipgloss.NewStyle().Width(columnWidth)
	rightColStyle := lipgloss.NewStyle().Width(columnWidth)

	// Collect lines for both panels
	leftLines, rightLines := s.collectViewportLines(availableHeight, columnWidth, lineNoWidth)

	// Build left and right columns independently
	leftVb := components.NewViewBuilder()
	rightVb := components.NewViewBuilder()

	// Add collected lines to ViewBuilders
	for i := 0; i < len(leftLines); i++ {
		leftVb.AddLine(leftColStyle.Render(leftLines[i].content))
	}
	for i := 0; i < len(rightLines); i++ {
		rightVb.AddLine(rightColStyle.Render(rightLines[i].content))
	}

	// Compose the split view
	vb.AddSplitView(leftVb, rightVb)
}

// renderWithAlignments renders the diff using the legacy alignment-based data model.
// This is used for backward compatibility during migration to segment-based rendering.
func (s *State) renderWithAlignments(vb *components.ViewBuilder, availableHeight, columnWidth int) {
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
}

// renderHeader formats the diff view header
func (s *State) renderHeader() string {
	// Format for single commit: abc123d · path/to/file.go · +15 -8
	// Format for range: abc123d..def456e · path/to/file.go · +15 -8
	var b strings.Builder

	// Display commit hash or range
	if s.CommitRange.IsSingleCommit() {
		b.WriteString(styles.HashStyle.Render(format.ToShortHash(s.CommitRange.End.Hash)))
	} else {
		startHash := format.ToShortHash(s.CommitRange.Start.Hash)
		endHash := format.ToShortHash(s.CommitRange.End.Hash)
		b.WriteString(styles.HashStyle.Render(fmt.Sprintf("%s..%s", startHash, endHash)))
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

// renderedLine represents a single rendered line for one panel.
// It contains the formatted string ready to be added to a ViewBuilder.
type renderedLine struct {
	content string
}

// collectViewportLines walks segments and collects lines for rendering.
// Returns left and right lines (may have different content at same row for hunks).
// This is the core of segment-based rendering that eliminates blank line padding.
func (s *State) collectViewportLines(availableHeight, columnWidth, lineNoWidth int) (leftLines, rightLines []renderedLine) {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return nil, nil
	}

	leftLines = make([]renderedLine, 0, availableHeight)
	rightLines = make([]renderedLine, 0, availableHeight)

	// Calculate content width (column width - lineNo - space - indicator - space)
	contentWidth := columnWidth - lineNoWidth - 4 // "123 - " = lineNo + space + indicator + space
	if contentWidth < 5 {
		contentWidth = 5
	}

	// Start from current segment position
	segIdx := s.SegmentIndex
	leftOff := s.LeftOffset
	rightOff := s.RightOffset

	// Collect lines until we've filled the viewport for the left panel
	// (we use left as the primary since we scroll both panels together initially)
	for len(leftLines) < availableHeight && segIdx < len(s.Diff.Segments) {
		seg := s.Diff.Segments[segIdx]

		switch segment := seg.(type) {
		case diff.UnchangedSegment:
			// Unchanged: same content goes to both panels
			for i := leftOff; i < segment.Count && len(leftLines) < availableHeight; i++ {
				leftIdx := segment.LeftStart + i
				rightIdx := segment.RightStart + i

				// Render left side
				leftLine := s.Diff.Left.Lines[leftIdx]
				leftNo := s.Diff.Left.LineNo(leftIdx)
				leftContent := s.formatColumnContent(leftNo, " ", leftLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle, nil)
				leftLines = append(leftLines, renderedLine{content: leftContent})

				// Render right side (same content for unchanged)
				rightLine := s.Diff.Right.Lines[rightIdx]
				rightNo := s.Diff.Right.LineNo(rightIdx)
				rightContent := s.formatColumnContent(rightNo, " ", rightLine.Tokens, lineNoWidth, contentWidth, columnWidth, styles.TimeStyle, nil)
				rightLines = append(rightLines, renderedLine{content: rightContent})
			}
			// Move to next segment
			segIdx++
			leftOff = 0
			rightOff = 0

		case diff.HunkSegment:
			// Hunk: collect lines independently for each panel
			leftCount := len(segment.LeftLines)
			rightCount := len(segment.RightLines)

			// Determine how many lines to collect from this hunk
			// For now, we use simple synchronized scrolling (same offset for both)
			// Differential scrolling will be implemented in Step 5
			maxHunkLines := max(leftCount-leftOff, rightCount-rightOff)

			for i := 0; i < maxHunkLines && len(leftLines) < availableHeight; i++ {
				// Left panel
				if leftOff+i < leftCount {
					hunkLine := segment.LeftLines[leftOff+i]
					leftIdx := hunkLine.SourceIdx
					leftLine := s.Diff.Left.Lines[leftIdx]
					leftNo := s.Diff.Left.LineNo(leftIdx)
					indicator, bgStyle := s.hunkLineStyle(hunkLine.Type, true)
					leftContent := s.formatColumnContent(leftNo, indicator, leftLine.Tokens, lineNoWidth, contentWidth, columnWidth, bgStyle, nil)
					leftLines = append(leftLines, renderedLine{content: leftContent})
				} else {
					// Left side exhausted - add filler row
					fillerContent := s.formatFillerLine(lineNoWidth, columnWidth, styles.TimeStyle)
					leftLines = append(leftLines, renderedLine{content: fillerContent})
				}

				// Right panel
				if rightOff+i < rightCount {
					hunkLine := segment.RightLines[rightOff+i]
					rightIdx := hunkLine.SourceIdx
					rightLine := s.Diff.Right.Lines[rightIdx]
					rightNo := s.Diff.Right.LineNo(rightIdx)
					indicator, bgStyle := s.hunkLineStyle(hunkLine.Type, false)
					rightContent := s.formatColumnContent(rightNo, indicator, rightLine.Tokens, lineNoWidth, contentWidth, columnWidth, bgStyle, nil)
					rightLines = append(rightLines, renderedLine{content: rightContent})
				} else {
					// Right side exhausted - add filler row
					fillerContent := s.formatFillerLine(lineNoWidth, columnWidth, styles.TimeStyle)
					rightLines = append(rightLines, renderedLine{content: fillerContent})
				}
			}
			// Move to next segment
			segIdx++
			leftOff = 0
			rightOff = 0
		}
	}

	return leftLines, rightLines
}

// hunkLineStyle returns the indicator and background style for a hunk line type.
// isLeft determines whether we're rendering the left (removed) or right (added) panel.
func (s *State) hunkLineStyle(lineType diff.HunkLineType, isLeft bool) (string, lipgloss.Style) {
	switch lineType {
	case diff.HunkLineRemoved:
		return "-", styles.DiffDeletionsStyle
	case diff.HunkLineAdded:
		return "+", styles.DiffAdditionsStyle
	case diff.HunkLineModified:
		// Modified lines: use removed style on left, added style on right
		if isLeft {
			return "-", styles.DiffDeletionsStyle
		}
		return "+", styles.DiffAdditionsStyle
	default:
		return " ", styles.TimeStyle
	}
}

// formatFillerLine creates an empty filler line with proper styling for alignment.
// This is used when one panel has fewer lines than the other in a hunk.
func (s *State) formatFillerLine(lineNoWidth, columnWidth int, bgStyle lipgloss.Style) string {
	// Create empty content with proper width
	return bgStyle.Width(columnWidth).Render("")
}

// calculateSegmentLineNoWidth returns the width needed for line numbers using segments.
func (s *State) calculateSegmentLineNoWidth() int {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return 3
	}

	maxLineNo := 0
	for _, seg := range s.Diff.Segments {
		switch segment := seg.(type) {
		case diff.UnchangedSegment:
			// Max line number is at the end of this segment
			leftNo := segment.LeftStart + segment.Count
			rightNo := segment.RightStart + segment.Count
			if leftNo > maxLineNo {
				maxLineNo = leftNo
			}
			if rightNo > maxLineNo {
				maxLineNo = rightNo
			}
		case diff.HunkSegment:
			// Check all lines in the hunk
			for _, hunkLine := range segment.LeftLines {
				lineNo := hunkLine.SourceIdx + 1 // 1-indexed
				if lineNo > maxLineNo {
					maxLineNo = lineNo
				}
			}
			for _, hunkLine := range segment.RightLines {
				lineNo := hunkLine.SourceIdx + 1 // 1-indexed
				if lineNo > maxLineNo {
					maxLineNo = lineNo
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
