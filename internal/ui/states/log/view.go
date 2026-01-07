package log

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

const (
	splitPanelWidth    = 80  // Fixed width for details panel
	splitThreshold     = 160 // Minimum terminal width to show split view
	separatorWidth     = 3   // Width of " │ " separator
	commitBodyMaxLines = 5   // Maximum lines for commit body in preview
)

// View renders the list of commits
func (s State) View(ctx core.Context) core.ViewRenderer {
	// Check if terminal is wide enough for split view
	if ctx.Width() >= splitThreshold {
		return s.renderSplitView(ctx)
	}
	return s.renderSimpleView(ctx)
}

// renderSimpleView renders the traditional single-column log view
func (s State) renderSimpleView(ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+ctx.Height(), len(s.Commits))

	// Render only visible commits
	for i := s.ViewportStart; i < viewportEnd; i++ {
		commit := s.Commits[i]

		// Prepare all components (impure operations happen here)
		lineComponents := s.buildCommitLineComponents(commit, i, false, ctx)

		// Call pure function with all components
		line := components.FormatCommitLine(lineComponents, ctx.Width())
		vb.AddLine(line)
	}

	return vb
}

// renderSplitView renders the log list on the left and details panel on the right
func (s State) renderSplitView(ctx core.Context) core.ViewRenderer {
	// Calculate widths
	logWidth := ctx.Width() - splitPanelWidth - separatorWidth
	detailsWidth := splitPanelWidth

	// Build columns independently
	leftVb := s.buildCommitListColumn(logWidth, ctx).(*components.ViewBuilder)
	rightVb := s.buildDetailsColumn(detailsWidth, ctx).(*components.ViewBuilder)

	// Compose the split view
	vb := components.NewViewBuilder()
	vb.AddSplitView(leftVb, rightVb)
	return vb
}

// buildCommitListColumn builds the left column (commit list) independently
func (s State) buildCommitListColumn(width int, ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Create style for fixed-width column
	colStyle := lipgloss.NewStyle().Width(width)

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+ctx.Height(), len(s.Commits))

	// Build the column with viewport height
	for i := 0; i < ctx.Height(); i++ {
		var line string
		logIdx := s.ViewportStart + i
		if logIdx < viewportEnd && logIdx < len(s.Commits) {
			commit := s.Commits[logIdx]

			// Prepare all components (impure operations happen here)
			lineComponents := s.buildCommitLineComponents(commit, logIdx, false, ctx)

			// Call pure function with all components
			line = components.FormatCommitLine(lineComponents, width)
		}
		// Apply fixed-width styling to each line
		vb.AddLine(colStyle.Render(line))
	}

	return vb
}

// buildDetailsColumn builds the right column (details panel) independently
func (s State) buildDetailsColumn(width int, ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Create style for fixed-width column
	colStyle := lipgloss.NewStyle().Width(width)

	// Render the details panel content
	detailsLines := s.renderDetailsPanel(width, ctx.Height(), ctx)

	// Build the column with viewport height
	for i := 0; i < ctx.Height(); i++ {
		var line string
		if i < len(detailsLines) {
			line = detailsLines[i]
		}
		// Apply fixed-width styling to each line
		vb.AddLine(colStyle.Render(line))
	}

	return vb
}

// buildCommitLineComponents prepares all components for formatting a commit line.
// This is where impure operations (time formatting, graph lookup) happen.
func (s State) buildCommitLineComponents(commit git.GitCommit, commitIndex int, isSelected bool, ctx core.Context) components.CommitLineComponents {
	return components.CommitLineComponents{
		DisplayState: s.getLineDisplayState(commitIndex),
		Graph:        s.buildGraphForCommit(commitIndex),
		Hash:         format.ToShortHash(commit.Hash),
		Refs:         commit.Refs,
		Message:      commit.Message,
		Author:       commit.Author,
		Time:         format.ToRelativeTimeFrom(commit.Date, ctx.Now()),
	}
}

// getLineDisplayState computes the display state for a commit line based on cursor mode and position.
func (s State) getLineDisplayState(index int) components.LineDisplayState {
	pos := s.CursorPosition()

	switch cursor := s.Cursor.(type) {
	case core.CursorNormal:
		if index == pos {
			return components.LineStateCursor
		}
		return components.LineStateNone
	case core.CursorVisual:
		if index == pos {
			return components.LineStateVisualCursor
		}
		if core.IsInSelection(cursor, index) {
			return components.LineStateSelected
		}
		return components.LineStateNone
	}
	return components.LineStateNone
}

// buildGraphForCommit returns the graph symbols for a commit at the given index.
// Returns empty string if no graph layout is available.
func (s State) buildGraphForCommit(commitIndex int) string {
	if s.GraphLayout != nil && commitIndex >= 0 && commitIndex < len(s.GraphLayout.Rows) {
		row := s.GraphLayout.Rows[commitIndex]
		return graph.RenderRow(row)
	}
	return ""
}

// renderDetailsPanel renders the details panel content for the currently selected commit
// Returns a slice of lines to display in the panel
func (s State) renderDetailsPanel(width, height int, ctx core.Context) []string {
	var lines []string

	// If no commits or cursor out of bounds, return empty panel
	pos := s.CursorPosition()
	if len(s.Commits) == 0 || pos < 0 || pos >= len(s.Commits) {
		return lines
	}

	commit := s.Commits[pos]

	// Always show commit info immediately (all data available in memory)
	commitInfoLines := components.CommitInfo(commit, width, commitBodyMaxLines, ctx)
	lines = append(lines, commitInfoLines...)

	// Render file section based on Preview state
	fileLines := s.renderFileList(width, height-len(lines))
	lines = append(lines, fileLines...)

	return lines
}

// renderFileList renders the file list based on Preview state
func (s State) renderFileList(width, maxLines int) []string {
	var lines []string

	// Check Preview state
	switch preview := s.Preview.(type) {
	case PreviewNone:
		// No preview loaded yet - show loading state for file section
		lines = append(lines, "")
		lines = append(lines, styles.TimeStyle.Render("Loading files..."))

	case PreviewLoading:
		// Loading in progress - show loading state for file section
		lines = append(lines, "")
		lines = append(lines, styles.TimeStyle.Render("Loading files..."))

	case PreviewError:
		// Error occurred - show error state for file section
		lines = append(lines, "")
		lines = append(lines, styles.DeletionsStyle.Render("Unable to load files"))

	case PreviewLoaded:
		// Check that the preview is for the current commit
		commit := s.Commits[s.CursorPosition()]
		if preview.ForHash != commit.Hash {
			// Stale data, show loading
			lines = append(lines, "")
			lines = append(lines, styles.TimeStyle.Render("Loading files..."))
		} else {
			// Use FileSection component to render files
			// Calculate how many files we can show
			fileSectionLines := components.FileSection(preview.Files, width, nil)

			// Truncate to available space if needed
			if len(fileSectionLines) > maxLines {
				// Keep blank line and stats line, truncate file list
				lines = append(lines, fileSectionLines[0]) // blank line
				lines = append(lines, fileSectionLines[1]) // stats line

				// Add as many file lines as will fit, leaving room for overflow indicator
				filesShown := 0
				for i := 2; i < len(fileSectionLines) && i-2 < maxLines-3; i++ {
					lines = append(lines, fileSectionLines[i])
					filesShown++
				}

				// Add overflow indicator if needed
				remaining := len(preview.Files) - filesShown
				if remaining > 0 {
					indicator := fmt.Sprintf("... and %d more file", remaining)
					if remaining > 1 {
						indicator += "s"
					}
					lines = append(lines, styles.TimeStyle.Render(indicator))
				}
			} else {
				lines = append(lines, fileSectionLines...)
			}
		}

	default:
		// Unknown state, show loading
		lines = append(lines, "")
		lines = append(lines, styles.TimeStyle.Render("Loading files..."))
	}

	return lines
}
