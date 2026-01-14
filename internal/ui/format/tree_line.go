package format

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/domain/filetree"
	"github.com/oberprah/splice/internal/ui/styles"
)

// FormatTreeLine formats a single tree line with box-drawing characters.
// This is a pure function that renders tree structure with proper indentation,
// branch characters, and styling based on node type and selection state.
//
// Format examples:
//   - Selected expanded folder:   → ├── src/
//   - Unselected collapsed folder:  └── old/ +50 -25 (3 files)
//   - Nested file:                  │   ├── M +17 -13  App.tsx
//
// The function handles:
//   - Selector rendering (→ for selected, spaces for unselected)
//   - Indentation using parentLines (│ or spaces based on ancestor structure)
//   - Branch characters (├── for non-last children, └── for last children)
//   - Node-specific content (folder names vs file stats)
//   - Style application (colors for status, selection highlighting)
func FormatTreeLine(item filetree.VisibleTreeItem, isSelected bool) string {
	var line strings.Builder

	// 1. Render selector (→ for selected, spaces for unselected)
	if isSelected {
		line.WriteString("→")
	} else {
		line.WriteString(" ")
	}

	// 2. Render indentation based on parentLines
	// Each parent level adds either "│   " (if parent has more siblings) or "    " (if parent is last)
	for _, hasMoreSiblings := range item.ParentLines {
		if hasMoreSiblings {
			line.WriteString("│   ")
		} else {
			line.WriteString("    ")
		}
	}

	// 3. Render branch character based on isLastChild
	if item.IsLastChild {
		line.WriteString("└── ")
	} else {
		line.WriteString("├── ")
	}

	// 4. Render node content based on type (folder vs file)
	switch node := item.Node.(type) {
	case *filetree.FolderNode:
		line.WriteString(formatFolderNode(node, isSelected))
	case *filetree.FileNode:
		line.WriteString(formatFileNode(node, isSelected))
	}

	return line.String()
}

// formatFolderNode formats a folder node.
// - Expanded: shows folder name only (e.g., "src/")
// - Collapsed: shows folder name + stats (e.g., "src/ +234 -67 (5 files)")
func formatFolderNode(node *filetree.FolderNode, isSelected bool) string {
	var content strings.Builder

	// Folder name styling
	folderStyle := styles.AuthorStyle // Cyan for folders
	if isSelected {
		folderStyle = styles.SelectedAuthorStyle
	}

	content.WriteString(folderStyle.Render(node.GetName()))

	// If collapsed, append stats
	if !node.IsExpanded() {
		stats := node.Stats()
		// Format: " +N -M (X files)"
		fileWord := "file"
		if stats.FileCount != 1 {
			fileWord = "files"
		}

		var statsStr strings.Builder
		statsStr.WriteString(" ")

		// Additions
		addStyle := styles.AdditionsStyle
		if isSelected {
			addStyle = styles.SelectedAdditionsStyle
		}
		statsStr.WriteString(addStyle.Render(fmt.Sprintf("+%d", stats.Additions)))

		statsStr.WriteString(" ")

		// Deletions
		delStyle := styles.DeletionsStyle
		if isSelected {
			delStyle = styles.SelectedDeletionsStyle
		}
		statsStr.WriteString(delStyle.Render(fmt.Sprintf("-%d", stats.Deletions)))

		statsStr.WriteString(" ")

		// File count
		countStyle := styles.TimeStyle
		if isSelected {
			countStyle = styles.SelectedTimeStyle
		}
		statsStr.WriteString(countStyle.Render(fmt.Sprintf("(%d %s)", stats.FileCount, fileWord)))

		content.WriteString(statsStr.String())
	}

	return content.String()
}

// formatFileNode formats a file node with status, stats, and filename.
// Format: "M +17 -13  App.tsx"
// Matches the existing file formatting from file_section.go
func formatFileNode(node *filetree.FileNode, isSelected bool) string {
	var content strings.Builder
	file := node.File()

	// Status letter
	status := file.Status
	if status == "" {
		status = "M" // Default to modified
	}

	// Style status based on type
	var statusStyle, addStyle, delStyle, pathStyle, binaryStyle = chooseFileStyles(status, isSelected)

	content.WriteString(statusStyle.Render(status))

	if file.IsBinary {
		content.WriteString(binaryStyle.Render(" (binary)"))
	} else {
		// Stats: " +N -M"
		// Note: We don't right-align stats here like in file_section.go because
		// tree indentation varies by depth, making alignment complex.
		// Simple left-aligned format is clearer in tree view.
		content.WriteString(addStyle.Render(fmt.Sprintf(" +%d", file.Additions)))
		content.WriteString(delStyle.Render(fmt.Sprintf(" -%d", file.Deletions)))
	}

	// Filename (just the name, not the full path)
	content.WriteString(pathStyle.Render("  " + node.GetName()))

	return content.String()
}

// chooseFileStyles selects appropriate styles based on status and selection state.
// Returns: statusStyle, addStyle, delStyle, pathStyle, binaryStyle
func chooseFileStyles(status string, isSelected bool) (statusStyle, addStyle, delStyle, pathStyle, binaryStyle lipgloss.Style) {
	if isSelected {
		return styles.SelectedHashStyle,
			styles.SelectedAdditionsStyle,
			styles.SelectedDeletionsStyle,
			styles.SelectedFilePathStyle,
			styles.SelectedTimeStyle
	}

	// Non-selected: status gets color based on type
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

	return statusStyle,
		styles.AdditionsStyle,
		styles.DeletionsStyle,
		styles.FilePathStyle,
		styles.TimeStyle
}
