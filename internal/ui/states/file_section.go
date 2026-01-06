package states

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/styles"
)

// FileSection renders file statistics and file list
// Returns lines to display (blank line + stats line + all file lines)
//
// Parameters:
//   - files: Files to display
//   - width: Panel width
//   - cursor: Selected file index (-1 for no selection)
//   - showSelector: Whether to show `>` selection indicator
//
// The component renders:
//  1. A blank line (separator from commit info above)
//  2. File stats line: `{N} files · +{add} -{del}`
//  3. All file lines with proper formatting
//
// Note: This component does NOT handle loading/error states or truncate the file list.
// The caller is responsible for those concerns.
func FileSection(files []git.FileChange, width int, cursor int, showSelector bool) []string {
	lines := make([]string, 0, len(files)+2)

	// 1. Blank line separator
	lines = append(lines, "")

	// 2. File stats line
	totalAdditions, totalDeletions := CalculateTotalStats(files)
	fileCount := len(files)
	fileWord := "file"
	if fileCount != 1 {
		fileWord = "files"
	}

	var statsLine strings.Builder
	statsLine.WriteString(styles.HeaderStyle.Render(fmt.Sprintf("%d %s", fileCount, fileWord)))
	statsLine.WriteString(styles.HeaderStyle.Render(" · "))
	statsLine.WriteString(styles.AdditionsStyle.Render(fmt.Sprintf("+%d", totalAdditions)))
	statsLine.WriteString(styles.HeaderStyle.Render(" "))
	statsLine.WriteString(styles.DeletionsStyle.Render(fmt.Sprintf("-%d", totalDeletions)))
	lines = append(lines, statsLine.String())

	// 3. File lines
	maxAddWidth, maxDelWidth := CalculateMaxStatWidth(files)

	for i, file := range files {
		isSelected := i == cursor
		line := FormatFileLine(FormatFileLineParams{
			File:         file,
			IsSelected:   isSelected,
			Width:        width,
			MaxAddWidth:  maxAddWidth,
			MaxDelWidth:  maxDelWidth,
			ShowSelector: showSelector,
		})
		lines = append(lines, line)
	}

	return lines
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
