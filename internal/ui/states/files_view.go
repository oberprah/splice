package states

import (
	"fmt"
	"strings"

	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// View renders the files state
func (s *FilesState) View(ctx Context) string {
	var b strings.Builder

	// Render header with commit info
	b.WriteString(s.renderHeader())
	b.WriteString("\n")

	// Render separator
	separator := strings.Repeat("─", min(ctx.Width(), 80))
	b.WriteString(styles.HeaderStyle.Render(separator))
	b.WriteString("\n")

	// Calculate available height for file list (subtract header lines)
	headerLines := 2 // commit info + separator
	availableHeight := max(ctx.Height()-headerLines, 1)

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+availableHeight, len(s.Files))

	// Render only visible files
	for i := s.ViewportStart; i < viewportEnd; i++ {
		file := s.Files[i]
		line := s.formatFileLine(file, i == s.Cursor, ctx.Width())
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader formats the commit information header
func (s *FilesState)renderHeader() string {
	// Format: Commit: abc123d Add feature
	// Author: John Doe
	// Date:   2 hours ago
	// Files:  3 changed, 45 insertions(+), 12 deletions(-)

	var b strings.Builder

	// Commit line
	b.WriteString(styles.HeaderStyle.Render("Commit: "))
	b.WriteString(styles.HashStyle.Render(format.ToShortHash(s.Commit.Hash)))
	b.WriteString(" ")
	b.WriteString(styles.MessageStyle.Render(s.Commit.Message))
	b.WriteString("\n")

	// Author line
	b.WriteString(styles.HeaderStyle.Render("Author: "))
	b.WriteString(styles.AuthorStyle.Render(s.Commit.Author))
	b.WriteString("\n")

	// Date line
	b.WriteString(styles.HeaderStyle.Render("Date:   "))
	b.WriteString(styles.TimeStyle.Render(format.ToRelativeTime(s.Commit.Date)))
	b.WriteString("\n")

	// Files summary line
	totalAdditions, totalDeletions := s.calculateTotalStats()
	b.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("Files:  %d changed, %d insertions(+), %d deletions(-)",
		len(s.Files), totalAdditions, totalDeletions)))

	return b.String()
}

// calculateTotalStats calculates total additions and deletions across all files
func (s *FilesState)calculateTotalStats() (int, int) {
	var totalAdditions, totalDeletions int
	for _, file := range s.Files {
		totalAdditions += file.Additions
		totalDeletions += file.Deletions
	}
	return totalAdditions, totalDeletions
}

// formatFileLine formats a single file line with proper styling
func (s *FilesState)formatFileLine(file git.FileChange, isSelected bool, width int) string {
	// Format: > path/to/file.go         (+45, -12)
	// or:       path/to/file.go         (+3, -1)

	availableWidth := width
	if availableWidth <= 0 {
		availableWidth = 80 // Default fallback
	}

	// Selection indicator (2 chars: "> " or "  ")
	selectionIndicator := "  "
	if isSelected {
		selectionIndicator = "> "
	}

	// Format the stats
	var stats string
	if file.IsBinary {
		stats = "(binary)"
	} else {
		stats = fmt.Sprintf("(+%d, -%d)", file.Additions, file.Deletions)
	}

	// Calculate spacing
	fixedWidth := len(selectionIndicator) + len(stats) + 2 // +2 for spacing
	pathMaxWidth := max(availableWidth-fixedWidth, 20)

	// Truncate path if necessary
	path := file.Path
	if len(path) > pathMaxWidth && pathMaxWidth > 3 {
		path = path[:pathMaxWidth-3] + "..."
	}

	// Build the line with styling
	var line strings.Builder

	line.WriteString(selectionIndicator)

	if isSelected {
		// Selected line styles
		line.WriteString(styles.SelectedFilePathStyle.Render(path))

		// Pad with spaces to align stats
		padding := availableWidth - len(selectionIndicator) - len(path) - len(stats) - 1
		if padding > 0 {
			line.WriteString(strings.Repeat(" ", padding))
		} else {
			line.WriteString(" ")
		}

		if file.IsBinary {
			line.WriteString(styles.SelectedTimeStyle.Render(stats))
		} else {
			// Format stats with colors
			addStr := fmt.Sprintf("+%d", file.Additions)
			delStr := fmt.Sprintf("-%d", file.Deletions)
			line.WriteString("(")
			line.WriteString(styles.SelectedAdditionsStyle.Render(addStr))
			line.WriteString(", ")
			line.WriteString(styles.SelectedDeletionsStyle.Render(delStr))
			line.WriteString(")")
		}
	} else {
		// Unselected line styles
		line.WriteString(styles.FilePathStyle.Render(path))

		// Pad with spaces to align stats
		padding := availableWidth - len(selectionIndicator) - len(path) - len(stats) - 1
		if padding > 0 {
			line.WriteString(strings.Repeat(" ", padding))
		} else {
			line.WriteString(" ")
		}

		if file.IsBinary {
			line.WriteString(styles.TimeStyle.Render(stats))
		} else {
			// Format stats with colors
			addStr := fmt.Sprintf("+%d", file.Additions)
			delStr := fmt.Sprintf("-%d", file.Deletions)
			line.WriteString("(")
			line.WriteString(styles.AdditionsStyle.Render(addStr))
			line.WriteString(", ")
			line.WriteString(styles.DeletionsStyle.Render(delStr))
			line.WriteString(")")
		}
	}

	return line.String()
}

