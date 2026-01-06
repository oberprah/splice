package states

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/graph"
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
func (s LogState) View(ctx Context) *ViewBuilder {
	// Check if terminal is wide enough for split view
	if ctx.Width() >= splitThreshold {
		return s.renderSplitView(ctx)
	}
	return s.renderSimpleView(ctx)
}

// renderSimpleView renders the traditional single-column log view
func (s LogState) renderSimpleView(ctx Context) *ViewBuilder {
	vb := NewViewBuilder()

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+ctx.Height(), len(s.Commits))

	// Render only visible commits
	for i := s.ViewportStart; i < viewportEnd; i++ {
		commit := s.Commits[i]
		isSelected := i == s.Cursor

		// Prepare all components (impure operations happen here)
		components := s.buildCommitLineComponents(commit, i, isSelected, ctx)

		// Call pure function with all components
		line := formatCommitLine(components, ctx.Width())
		vb.AddLine(line)
	}

	return vb
}

// renderSplitView renders the log list on the left and details panel on the right
func (s LogState) renderSplitView(ctx Context) *ViewBuilder {
	// Calculate widths
	logWidth := ctx.Width() - splitPanelWidth - separatorWidth
	detailsWidth := splitPanelWidth

	// Build columns independently
	leftVb := s.buildCommitListColumn(logWidth, ctx)
	rightVb := s.buildDetailsColumn(detailsWidth, ctx)

	// Compose the split view
	vb := NewViewBuilder()
	vb.AddSplitView(leftVb, rightVb)
	return vb
}

// buildCommitListColumn builds the left column (commit list) independently
func (s LogState) buildCommitListColumn(width int, ctx Context) *ViewBuilder {
	vb := NewViewBuilder()

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
			isSelected := logIdx == s.Cursor

			// Prepare all components (impure operations happen here)
			components := s.buildCommitLineComponents(commit, logIdx, isSelected, ctx)

			// Call pure function with all components
			line = formatCommitLine(components, width)
		}
		// Apply fixed-width styling to each line
		vb.AddLine(colStyle.Render(line))
	}

	return vb
}

// buildDetailsColumn builds the right column (details panel) independently
func (s LogState) buildDetailsColumn(width int, ctx Context) *ViewBuilder {
	vb := NewViewBuilder()

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
func (s LogState) buildCommitLineComponents(commit git.GitCommit, commitIndex int, isSelected bool, ctx Context) CommitLineComponents {
	return CommitLineComponents{
		IsSelected: isSelected,
		Graph:      s.buildGraphForCommit(commitIndex),
		Hash:       format.ToShortHash(commit.Hash),
		Refs:       commit.Refs,
		Message:    commit.Message,
		Author:     commit.Author,
		Time:       format.ToRelativeTimeFrom(commit.Date, ctx.Now()),
	}
}

// buildGraphForCommit returns the graph symbols for a commit at the given index.
// Returns empty string if no graph layout is available.
func (s LogState) buildGraphForCommit(commitIndex int) string {
	if s.GraphLayout != nil && commitIndex >= 0 && commitIndex < len(s.GraphLayout.Rows) {
		row := s.GraphLayout.Rows[commitIndex]
		return graph.RenderRow(row)
	}
	return ""
}

// renderDetailsPanel renders the details panel content for the currently selected commit
// Returns a slice of lines to display in the panel
func (s LogState) renderDetailsPanel(width, height int, ctx Context) []string {
	var lines []string

	// If no commits or cursor out of bounds, return empty panel
	if len(s.Commits) == 0 || s.Cursor < 0 || s.Cursor >= len(s.Commits) {
		return lines
	}

	commit := s.Commits[s.Cursor]

	// Always show commit info immediately (all data available in memory)
	commitInfoLines := CommitInfo(commit, width, commitBodyMaxLines, ctx)
	lines = append(lines, commitInfoLines...)

	// Render file section based on Preview state
	fileLines := s.renderFileList(width, height-len(lines))
	lines = append(lines, fileLines...)

	return lines
}

// renderFileList renders the file list based on Preview state
func (s LogState) renderFileList(width, maxLines int) []string {
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
		commit := s.Commits[s.Cursor]
		if preview.ForHash != commit.Hash {
			// Stale data, show loading
			lines = append(lines, "")
			lines = append(lines, styles.TimeStyle.Render("Loading files..."))
		} else {
			// Use FileSection component to render files
			// Calculate how many files we can show
			fileSectionLines := FileSection(preview.Files, width, nil)

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
