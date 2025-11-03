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
	header := s.renderHeader()
	b.WriteString(header)

	// Render separator
	separator := strings.Repeat("─", min(ctx.Width(), 80))
	b.WriteString(styles.HeaderStyle.Render(separator))
	b.WriteString("\n")

	// Calculate available height for file list (subtract header lines)
	// Count actual header lines (including body if present)
	headerLines := strings.Count(header, "\n") + 1 // +1 for separator
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
func (s *FilesState) renderHeader() string {
	// Format:
	// abc123d · John Doe committed 2 hours ago · 3 files · +45 -12
	//
	// Subject line
	//
	// Body paragraph 1...
	// Body paragraph 2...

	var b strings.Builder

	// First line: hash · author committed time ago · X files · +Y -Z
	totalAdditions, totalDeletions := s.calculateTotalStats()

	b.WriteString(styles.HashStyle.Render(format.ToShortHash(s.Commit.Hash)))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.AuthorStyle.Render(s.Commit.Author))
	b.WriteString(styles.HeaderStyle.Render(" committed "))
	b.WriteString(styles.TimeStyle.Render(format.ToRelativeTime(s.Commit.Date)))
	b.WriteString(styles.HeaderStyle.Render(" · "))

	// File stats
	fileCount := len(s.Files)
	fileWord := "file"
	if fileCount != 1 {
		fileWord = "files"
	}
	b.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("%d %s", fileCount, fileWord)))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdditions)))
	b.WriteString(styles.HeaderStyle.Render(" "))
	b.WriteString(styles.DeletionsStyle.Render(fmt.Sprintf("-%d", totalDeletions)))
	b.WriteString("\n\n")

	// Subject line
	b.WriteString(styles.MessageStyle.Render(s.Commit.Message))
	b.WriteString("\n")

	// Body (if exists)
	if s.Commit.Body != "" {
		b.WriteString("\n")
		b.WriteString(styles.MessageStyle.Render(s.Commit.Body))
		b.WriteString("\n")
	}

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

// calculateMaxStatWidth calculates the maximum width needed for additions and deletions
func (s *FilesState) calculateMaxStatWidth() (int, int) {
	maxAddWidth := 2 // Minimum: +0
	maxDelWidth := 2 // Minimum: -0

	for _, file := range s.Files {
		if !file.IsBinary {
			// Calculate width: sign (1) + digits
			addWidth := len(fmt.Sprintf("%d", file.Additions)) + 1
			delWidth := len(fmt.Sprintf("%d", file.Deletions)) + 1

			if addWidth > maxAddWidth {
				maxAddWidth = addWidth
			}
			if delWidth > maxDelWidth {
				maxDelWidth = delWidth
			}
		}
	}

	return maxAddWidth, maxDelWidth
}

// formatFileLine formats a single file line with proper styling
func (s *FilesState)formatFileLine(file git.FileChange, isSelected bool, width int) string {
	// Format: > M +17 -13  src/components/App.tsx
	// or:       A +130 -0  src/components/FileList.tsx

	var line strings.Builder

	// Selection indicator (1 char: ">" or " ")
	if isSelected {
		line.WriteString(">")
	} else {
		line.WriteString(" ")
	}

	// Status letter (1 char, padded)
	status := file.Status
	if status == "" {
		status = "M" // Default to modified
	}
	line.WriteString(" ")
	line.WriteString(status)

	// Calculate dynamic widths based on all files
	maxAddWidth, maxDelWidth := s.calculateMaxStatWidth()

	// Apply styling based on selection
	if isSelected {
		// Color the additions and deletions separately for selected line
		if file.IsBinary {
			line.WriteString(styles.SelectedTimeStyle.Render(" (binary)"))
		} else {
			// Split the stats to color them separately with dynamic width
			addStr := fmt.Sprintf(" %+*d", maxAddWidth, file.Additions)
			delStr := fmt.Sprintf(" %+*d", maxDelWidth, -file.Deletions)
			line.WriteString(styles.SelectedAdditionsStyle.Render(addStr))
			line.WriteString(styles.SelectedDeletionsStyle.Render(delStr))
		}
		line.WriteString(styles.SelectedFilePathStyle.Render("  " + file.Path))
	} else {
		// Color the additions and deletions separately
		if file.IsBinary {
			line.WriteString(styles.TimeStyle.Render(" (binary)"))
		} else {
			// Split the stats to color them separately with dynamic width
			addStr := fmt.Sprintf(" %+*d", maxAddWidth, file.Additions)
			delStr := fmt.Sprintf(" %+*d", maxDelWidth, -file.Deletions)
			line.WriteString(styles.AdditionsStyle.Render(addStr))
			line.WriteString(styles.DeletionsStyle.Render(delStr))
		}
		line.WriteString(styles.FilePathStyle.Render("  " + file.Path))
	}

	return line.String()
}

