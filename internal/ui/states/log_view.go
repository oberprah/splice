package states

import (
	"fmt"
	"strings"

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

	// Render metadata line if files are loaded
	metadataLine := s.renderMetadataLine(commit, width, ctx)
	if metadataLine != "" {
		lines = append(lines, metadataLine)
		lines = append(lines, "") // Blank line after metadata
	}

	// Render commit message (subject + body)
	messageLines := s.renderCommitMessage(commit, width)
	lines = append(lines, messageLines...)

	// Add separator line
	separator := strings.Repeat("─", width)
	lines = append(lines, styles.HeaderStyle.Render(separator))

	// Render file list based on Preview state
	fileLines := s.renderFileList(width, height-len(lines))
	lines = append(lines, fileLines...)

	return lines
}

// renderMetadataLine renders the commit metadata line if files are available
// Returns empty string if files are not loaded yet
func (s LogState) renderMetadataLine(commit git.GitCommit, width int, ctx Context) string {
	// Only show metadata if we have loaded files
	if previewLoaded, ok := s.Preview.(PreviewLoaded); ok && previewLoaded.ForHash == commit.Hash {
		metadata := RenderCommitMetadata(commit, previewLoaded.Files, ctx)
		// Truncate if needed
		if len(metadata) > width {
			// Note: This is approximate due to ANSI codes, but better than nothing
			return metadata[:width-3] + "..."
		}
		return metadata
	}
	return ""
}

// renderCommitMessage renders the commit subject and body (truncated)
func (s LogState) renderCommitMessage(commit git.GitCommit, width int) []string {
	var lines []string

	// Subject line (always show)
	subject := commit.Message
	if len(subject) > width {
		subject = subject[:width-3] + "..."
	}
	lines = append(lines, styles.MessageStyle.Render(subject))

	// Body (if exists, limit to commitBodyMaxLines)
	if commit.Body != "" {
		// Add blank line between subject and body
		lines = append(lines, "")

		// Split body into lines and wrap to width
		bodyLines := strings.Split(commit.Body, "\n")
		lineCount := 0

		for _, bodyLine := range bodyLines {
			if lineCount >= commitBodyMaxLines {
				// Add truncation indicator
				lines = append(lines, styles.TimeStyle.Render("..."))
				break
			}

			// Wrap long lines
			if len(bodyLine) > width {
				wrapped := wrapText(bodyLine, width)
				for _, wrappedLine := range wrapped {
					if lineCount >= commitBodyMaxLines {
						lines = append(lines, styles.TimeStyle.Render("..."))
						break
					}
					lines = append(lines, styles.MessageStyle.Render(wrappedLine))
					lineCount++
				}
			} else {
				lines = append(lines, styles.MessageStyle.Render(bodyLine))
				lineCount++
			}

			if lineCount >= commitBodyMaxLines {
				break
			}
		}
	}

	return lines
}

// renderFileList renders the file list based on Preview state
func (s LogState) renderFileList(width, maxLines int) []string {
	var lines []string

	// Check Preview state
	switch preview := s.Preview.(type) {
	case PreviewNone:
		// No preview loaded yet
		lines = append(lines, styles.TimeStyle.Render("Loading..."))

	case PreviewLoading:
		// Loading in progress
		lines = append(lines, styles.TimeStyle.Render("Loading..."))

	case PreviewError:
		// Error occurred
		lines = append(lines, styles.DeletionsStyle.Render("Unable to load files"))

	case PreviewLoaded:
		// Check that the preview is for the current commit
		commit := s.Commits[s.Cursor]
		if preview.ForHash != commit.Hash {
			// Stale data, show loading
			lines = append(lines, styles.TimeStyle.Render("Loading..."))
		} else {
			// Render files
			lines = s.renderFiles(preview.Files, width, maxLines)
		}

	default:
		// Unknown state, show loading
		lines = append(lines, styles.TimeStyle.Render("Loading..."))
	}

	return lines
}

// renderFiles renders the file list with status indicators and stats
func (s LogState) renderFiles(files []git.FileChange, width, maxLines int) []string {
	var lines []string

	// Determine how many files we can show
	filesShown := 0
	for i, file := range files {
		if filesShown >= maxLines {
			// Add overflow indicator
			remaining := len(files) - i
			if remaining > 0 {
				indicator := fmt.Sprintf("... and %d more file", remaining)
				if remaining > 1 {
					indicator += "s"
				}
				lines = append(lines, styles.TimeStyle.Render(indicator))
			}
			break
		}

		line := s.formatFileEntry(file, width)
		lines = append(lines, line)
		filesShown++
	}

	return lines
}

// formatFileEntry formats a single file entry with status, stats, and path
// Format: "Status +add -del  path"
func (s LogState) formatFileEntry(file git.FileChange, width int) string {
	// For log view, we don't show selection indicator and use fixed widths for stats
	// since we don't know all files at format time in the preview panel
	maxAddWidth := len(fmt.Sprintf("+%d", file.Additions)) + 1
	maxDelWidth := len(fmt.Sprintf("-%d", file.Deletions)) + 1

	// Ensure minimum widths
	if maxAddWidth < 2 {
		maxAddWidth = 2
	}
	if maxDelWidth < 2 {
		maxDelWidth = 2
	}

	return FormatFileLine(FormatFileLineParams{
		File:         file,
		IsSelected:   false,
		Width:        width,
		MaxAddWidth:  maxAddWidth,
		MaxDelWidth:  maxDelWidth,
		ShowSelector: false,
	})
}

// wrapText wraps text to the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var currentLine strings.Builder
	for _, word := range words {
		// If adding this word would exceed width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
