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
		components := CommitLineComponents{
			Selector: buildSelector(isSelected),
			Graph:    s.buildGraphForCommit(i),
			Hash:     format.ToShortHash(commit.Hash),
			Refs:     commit.Refs,
			Message:  commit.Message,
			Author:   commit.Author,
			Time:     format.ToRelativeTimeFrom(commit.Date, ctx.Now()),
		}

		// Call pure function with all components
		line := formatCommitLine(components, ctx.Width(), isSelected)
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
			components := CommitLineComponents{
				Selector: buildSelector(isSelected),
				Graph:    s.buildGraphForCommit(logIdx),
				Hash:     format.ToShortHash(commit.Hash),
				Refs:     commit.Refs,
				Message:  commit.Message,
				Author:   commit.Author,
				Time:     format.ToRelativeTimeFrom(commit.Date, ctx.Now()),
			}

			// Call pure function with all components
			line = formatCommitLine(components, width, isSelected)
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

// RefsLevel represents the truncation level for refs display
type RefsLevel int

const (
	RefsLevelFull RefsLevel = iota
	RefsLevelShortenIndividual
	RefsLevelFirstPlusCount
	RefsLevelCountOnly
)

// CommitLineComponents holds all pre-computed components for a commit line.
// This struct enables formatCommitLine to be a pure function.
type CommitLineComponents struct {
	Selector string
	Graph    string
	Hash     string
	Refs     []git.RefInfo
	Message  string
	Author   string
	Time     string
}

// buildSelector returns the selection indicator string.
// Returns "> " for selected items, "  " for unselected.
func buildSelector(isSelected bool) string {
	if isSelected {
		return "> "
	}
	return "  "
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

// capMessage truncates a message to maxLen characters with "..." suffix.
// Returns the original message if it fits within maxLen.
func capMessage(message string, maxLen int) string {
	if len(message) <= maxLen {
		return message
	}
	if maxLen < 3 {
		return ""
	}
	return message[:maxLen-3] + "..."
}

// truncateAuthor truncates an author name to maxLen characters with "..." suffix.
// Returns the original author if it fits within maxLen.
// Returns empty string if maxLen < 3.
func truncateAuthor(author string, maxLen int) string {
	if len(author) <= maxLen {
		return author
	}
	if maxLen < 3 {
		return ""
	}
	return author[:maxLen-3] + "..."
}

// truncateEntireLine hard-truncates an assembled line to maxWidth characters.
// Uses "..." suffix if maxWidth >= 3, otherwise truncates to available space.
func truncateEntireLine(line string, maxWidth int) string {
	if len(line) <= maxWidth {
		return line
	}
	if maxWidth <= 0 {
		return ""
	}
	if maxWidth < 3 {
		return line[:maxWidth]
	}
	return line[:maxWidth-3] + "..."
}

// formatRefsFull formats all refs with their full names.
// Returns formatted refs like "(HEAD -> main, tag: v1.0)" with trailing space.
func formatRefsFull(refs []git.RefInfo) string {
	if len(refs) == 0 {
		return ""
	}

	var parts []string
	for _, ref := range refs {
		var formatted string
		switch ref.Type {
		case git.RefTypeTag:
			formatted = fmt.Sprintf("tag: %s", ref.Name)
		default:
			// For branches, just use the name
			formatted = ref.Name
		}
		parts = append(parts, formatted)
	}

	return fmt.Sprintf("(%s) ", strings.Join(parts, ", "))
}

// formatRefsShortenedIndividual formats refs with individual names truncated to maxLen.
// Uses "…" (single ellipsis char) for truncation to save space.
// Note: "…" is 3 bytes in UTF-8, so maxLen is treated as byte count.
func formatRefsShortenedIndividual(refs []git.RefInfo, maxLen int) string {
	if len(refs) == 0 {
		return ""
	}

	var parts []string
	for _, ref := range refs {
		var formatted string
		switch ref.Type {
		case git.RefTypeTag:
			name := ref.Name
			if len(name) > maxLen {
				if maxLen < 3 {
					name = ""
				} else {
					name = name[:maxLen-3] + "…" // "…" is 3 bytes
				}
			}
			formatted = fmt.Sprintf("tag: %s", name)
		default:
			// For branches, truncate the name
			name := ref.Name
			if len(name) > maxLen {
				if maxLen < 3 {
					name = ""
				} else {
					name = name[:maxLen-3] + "…" // "…" is 3 bytes
				}
			}
			formatted = name
		}
		parts = append(parts, formatted)
	}

	return fmt.Sprintf("(%s) ", strings.Join(parts, ", "))
}

// formatRefsFirstPlusCount formats refs showing only the first ref plus a count.
// Prefers showing the current branch (HEAD ref) if present.
// First ref is still truncated if needed with "…".
// Note: "…" is 3 bytes in UTF-8, so maxLen is treated as byte count.
func formatRefsFirstPlusCount(refs []git.RefInfo, maxLen int) string {
	if len(refs) == 0 {
		return ""
	}

	// Find the HEAD ref (current branch) if it exists
	var firstRef git.RefInfo
	foundHead := false
	for _, ref := range refs {
		if ref.IsHead {
			firstRef = ref
			foundHead = true
			break
		}
	}

	// If no HEAD ref, use the first ref
	if !foundHead {
		firstRef = refs[0]
	}

	// Format the first ref
	var formatted string
	switch firstRef.Type {
	case git.RefTypeTag:
		name := firstRef.Name
		if len(name) > maxLen {
			if maxLen < 3 {
				name = ""
			} else {
				name = name[:maxLen-3] + "…" // "…" is 3 bytes
			}
		}
		formatted = fmt.Sprintf("tag: %s", name)
	default:
		name := firstRef.Name
		if len(name) > maxLen {
			if maxLen < 3 {
				name = ""
			} else {
				name = name[:maxLen-3] + "…" // "…" is 3 bytes
			}
		}
		formatted = name
	}

	// Calculate remaining refs count
	remaining := len(refs) - 1
	if remaining > 0 {
		return fmt.Sprintf("(%s +%d more) ", formatted, remaining)
	}

	return fmt.Sprintf("(%s) ", formatted)
}

// buildRefs builds the refs string at the specified truncation level.
// Returns empty string if no refs, otherwise returns formatted string with trailing space.
func buildRefs(refs []git.RefInfo, level RefsLevel) string {
	if len(refs) == 0 {
		return ""
	}

	switch level {
	case RefsLevelFull:
		return formatRefsFull(refs)
	case RefsLevelShortenIndividual:
		return formatRefsShortenedIndividual(refs, 30)
	case RefsLevelFirstPlusCount:
		return formatRefsFirstPlusCount(refs, 30)
	case RefsLevelCountOnly:
		return fmt.Sprintf("(%d refs) ", len(refs))
	default:
		return formatRefsFull(refs)
	}
}

// measureLineWidth calculates the total width of a commit line.
// Accounts for all components and spacing: selector + graph + hash + space + refs + message + separator + author + space + time.
// Note: refs already includes trailing space if non-empty.
func measureLineWidth(selector, graph, hash, refs, message, author, time string) int {
	width := len(selector) + len(graph) + len(hash)

	if refs != "" {
		width += 1 + len(refs) // space before refs + refs (which includes trailing space)
	} else {
		width += 1 // space after hash when no refs
	}

	width += len(message)

	if author != "" {
		width += 3 + len(author) // " - " + author
	}

	if time != "" {
		width += 1 + len(time) // space + time
	}

	return width
}

// assembleLine assembles the final commit line with proper spacing, separators, and styling.
// This is a pure function that builds the styled string from plain components.
func assembleLine(selector, graph, hash, refs, message, author, time string, isSelected bool) string {
	var line strings.Builder

	// Add selector and graph (no styling)
	line.WriteString(selector)
	line.WriteString(graph)

	// Choose styles based on selection
	var hashStyle, messageStyle, authorStyle, timeStyle lipgloss.Style
	if isSelected {
		hashStyle = styles.SelectedHashStyle
		messageStyle = styles.SelectedMessageStyle
		authorStyle = styles.SelectedAuthorStyle
		timeStyle = styles.SelectedTimeStyle
	} else {
		hashStyle = styles.HashStyle
		messageStyle = styles.MessageStyle
		authorStyle = styles.AuthorStyle
		timeStyle = styles.TimeStyle
	}

	// Add hash with space
	line.WriteString(hashStyle.Render(hash))
	line.WriteString(" ")

	// Add refs (with space) if present
	if refs != "" {
		line.WriteString(timeStyle.Render(refs)) // Use time style for refs (dim)
	}

	// Add message
	line.WriteString(messageStyle.Render(message))

	// Add separator + author if both present
	if author != "" && message != "" {
		line.WriteString(" - ")
		line.WriteString(authorStyle.Render(author))
	}

	// Add time (with space prefix) if present
	if time != "" {
		line.WriteString(" ")
		line.WriteString(timeStyle.Render(time))
	}

	return line.String()
}

// formatCommitLine applies progressive truncation to fit a commit line within available width.
// Pure function - all inputs provided via CommitLineComponents struct, no side effects.
func formatCommitLine(components CommitLineComponents, availableWidth int, isSelected bool) string {
	// 1. Extract components (already computed by caller)
	selector := components.Selector
	graph := components.Graph
	hash := components.Hash
	message := components.Message
	author := components.Author
	time := components.Time

	// Build refs at full level initially
	refs := buildRefs(components.Refs, RefsLevelFull)

	// 2. Apply truncation levels sequentially until line fits
	level := 0
	for measureLineWidth(selector, graph, hash, refs, message, author, time) > availableWidth && level < 10 {
		switch level {
		case 0:
			// Level 0: Cap message at 72 chars
			message = capMessage(message, 72)
		case 1:
			// Level 1: Truncate author to 25 chars
			author = truncateAuthor(author, 25)
		case 2:
			// Level 2: Shorten refs Level 1 - Truncate individual ref names
			refs = buildRefs(components.Refs, RefsLevelShortenIndividual)
		case 3:
			// Level 3: Shorten refs Level 2 - Show first ref + count
			refs = buildRefs(components.Refs, RefsLevelFirstPlusCount)
		case 4:
			// Level 4: Shorten refs Level 3 - Show total count only
			refs = buildRefs(components.Refs, RefsLevelCountOnly)
		case 5:
			// Level 5: Truncate author to 5 chars
			author = truncateAuthor(author, 5)
		case 6:
			// Level 6: Drop time
			time = ""
		case 7:
			// Level 7: Shorten message to 40 chars
			message = capMessage(message, 40)
		case 8:
			// Level 8: Drop author
			author = ""
		case 9:
			// Level 9: Drop refs, assemble minimal line, then truncate entire line to fit
			refs = ""
			assembledLine := assembleLine(selector, graph, hash, refs, message, author, time, isSelected)
			if len(assembledLine) > availableWidth {
				return truncateEntireLine(assembledLine, availableWidth)
			}
			return assembledLine
		}
		level++
	}

	// 3. Assemble and style the line
	return assembleLine(selector, graph, hash, refs, message, author, time, isSelected)
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
