package states

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// RenderCommitMetadata renders the metadata line for a commit
// Format: abc123d · John Doe committed 2 hours ago · 3 files · +45 -12
func RenderCommitMetadata(commit git.GitCommit, files []git.FileChange) string {
	var b strings.Builder

	// Calculate total stats
	totalAdditions, totalDeletions := CalculateTotalStats(files)

	// First line: hash · author committed time ago · X files · +Y -Z
	b.WriteString(styles.HashStyle.Render(format.ToShortHash(commit.Hash)))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.AuthorStyle.Render(commit.Author))
	b.WriteString(styles.HeaderStyle.Render(" committed "))
	b.WriteString(styles.TimeStyle.Render(format.ToRelativeTime(commit.Date)))
	b.WriteString(styles.HeaderStyle.Render(" · "))

	// File stats
	fileCount := len(files)
	fileWord := "file"
	if fileCount != 1 {
		fileWord = "files"
	}
	b.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("%d %s", fileCount, fileWord)))
	b.WriteString(styles.HeaderStyle.Render(" · "))
	b.WriteString(styles.AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdditions)))
	b.WriteString(styles.HeaderStyle.Render(" "))
	b.WriteString(styles.DeletionsStyle.Render(fmt.Sprintf("-%d", totalDeletions)))

	return b.String()
}

// CalculateTotalStats calculates total additions and deletions across all files
func CalculateTotalStats(files []git.FileChange) (int, int) {
	var totalAdditions, totalDeletions int
	for _, file := range files {
		totalAdditions += file.Additions
		totalDeletions += file.Deletions
	}
	return totalAdditions, totalDeletions
}

// CalculateMaxStatWidth calculates the maximum width needed for additions and deletions
func CalculateMaxStatWidth(files []git.FileChange) (int, int) {
	maxAddWidth := 2 // Minimum: +0
	maxDelWidth := 2 // Minimum: -0

	for _, file := range files {
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

// FormatFileLineParams contains parameters for formatting a file line
type FormatFileLineParams struct {
	File         git.FileChange
	IsSelected   bool
	Width        int
	MaxAddWidth  int
	MaxDelWidth  int
	ShowSelector bool // If true, shows "> " or "  " prefix
}

// FormatFileLine formats a single file line with proper styling
// Format with selector: > M +17 -13  src/components/App.tsx
// Format without selector: M +17 -13  src/components/App.tsx
func FormatFileLine(params FormatFileLineParams) string {
	var line strings.Builder

	// Selection indicator (if enabled)
	if params.ShowSelector {
		if params.IsSelected {
			line.WriteString(">")
		} else {
			line.WriteString(" ")
		}
	}

	// Status letter (1 char, padded)
	status := params.File.Status
	if status == "" {
		status = "M" // Default to modified
	}

	// Style status based on type (for non-selected lines)
	var statusStyle lipgloss.Style
	if !params.IsSelected {
		switch status {
		case "A":
			statusStyle = styles.AdditionsStyle // Green
		case "M":
			statusStyle = styles.HashStyle // Yellow/amber
		case "D":
			statusStyle = styles.DeletionsStyle // Red
		case "R":
			statusStyle = styles.AuthorStyle // Cyan/blue
		default:
			statusStyle = styles.TimeStyle // Gray for unknown
		}
	}

	if params.ShowSelector {
		line.WriteString(" ")
	}

	// Apply styling based on selection
	if params.IsSelected {
		// For selected lines, use selected styles
		line.WriteString(styles.SelectedHashStyle.Render(status))

		if params.File.IsBinary {
			line.WriteString(styles.SelectedTimeStyle.Render(" (binary)"))
		} else {
			// Split the stats to color them separately with dynamic width
			// Format: +N for additions, -N for deletions
			addStr := fmt.Sprintf(" +%*d", params.MaxAddWidth-1, params.File.Additions)
			delStr := fmt.Sprintf(" -%*d", params.MaxDelWidth-1, params.File.Deletions)
			line.WriteString(styles.SelectedAdditionsStyle.Render(addStr))
			line.WriteString(styles.SelectedDeletionsStyle.Render(delStr))
		}
		line.WriteString(styles.SelectedFilePathStyle.Render("  " + params.File.Path))
	} else {
		// For unselected lines, use standard styles
		line.WriteString(statusStyle.Render(status))

		if params.File.IsBinary {
			line.WriteString(styles.TimeStyle.Render(" (binary)"))
		} else {
			// Split the stats to color them separately with dynamic width
			// Format: +N for additions, -N for deletions
			addStr := fmt.Sprintf(" +%*d", params.MaxAddWidth-1, params.File.Additions)
			delStr := fmt.Sprintf(" -%*d", params.MaxDelWidth-1, params.File.Deletions)
			line.WriteString(styles.AdditionsStyle.Render(addStr))
			line.WriteString(styles.DeletionsStyle.Render(delStr))
		}
		line.WriteString(styles.FilePathStyle.Render("  " + params.File.Path))
	}

	return line.String()
}

// TruncatePathFromLeft truncates a file path from the left if it exceeds maxWidth
// Returns "...filename.ext" or "...dir/filename.ext" style truncation
func TruncatePathFromLeft(path string, maxWidth int) string {
	if len(path) <= maxWidth {
		return path
	}
	if maxWidth <= 3 {
		return path // Can't truncate meaningfully
	}
	return "..." + path[len(path)-(maxWidth-3):]
}
