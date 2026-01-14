package components

import (
	"fmt"
	"strings"

	"github.com/oberprah/splice/internal/domain/tree"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// TreeSection renders a tree view with file statistics and tree structure.
// Returns lines to display (blank line + stats line + all tree item lines)
//
// Parameters:
//   - items: Visible tree items (already flattened and windowed by caller)
//   - cursor: Index of the selected item in the items slice
//   - width: Panel width (currently unused but kept for consistency with FileSection)
//
// The component renders:
//  1. A blank line (separator from commit info above)
//  2. File stats line: `{N} files · +{add} -{del}`
//  3. All tree item lines with proper tree formatting
//
// Note: This component does NOT handle viewport windowing - the caller is responsible
// for passing only the visible items. This follows the same pattern as FileSection
// where viewport logic is handled in the state's View() method.
func TreeSection(items []tree.VisibleTreeItem, cursor int, width int) []string {
	lines := make([]string, 0, len(items)+2)

	// 1. Blank line separator
	lines = append(lines, "")

	// 2. File stats line
	totalAdditions, totalDeletions, fileCount := CalculateTreeStats(items)
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

	// 3. Tree item lines
	for i, item := range items {
		isSelected := i == cursor
		line := format.FormatTreeLine(item, isSelected)
		lines = append(lines, line)
	}

	return lines
}

// CalculateTreeStats calculates total additions, deletions, and file count from tree items.
// Only counts FileNode items (not folders) to avoid double-counting.
func CalculateTreeStats(items []tree.VisibleTreeItem) (additions int, deletions int, fileCount int) {
	for _, item := range items {
		// Only count file nodes, not folders
		if fileNode, ok := item.Node.(*tree.FileNode); ok {
			file := fileNode.File()
			additions += file.Additions
			deletions += file.Deletions
			fileCount++
		}
	}
	return additions, deletions, fileCount
}
