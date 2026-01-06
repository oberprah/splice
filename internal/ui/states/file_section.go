package states

import (
	"fmt"
	"strings"

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
