// ABOUTME: Merges full file content with diff information to produce a unified view.
// ABOUTME: Enables showing entire files with changes highlighted rather than just diff hunks.

package diff

import (
	"strings"
)

// ChangeType represents the type of change for a line in the full file diff
type ChangeType int

const (
	Unchanged ChangeType = iota // Line exists in both old and new versions
	Added                       // Line only exists in new version
	Removed                     // Line only exists in old version
)

// FullFileLine represents a single line in the merged full file view
type FullFileLine struct {
	LeftLineNo   int        // Line number in old file (0 if not present)
	RightLineNo  int        // Line number in new file (0 if not present)
	LeftContent  string     // Content from old file (empty if added)
	RightContent string     // Content from new file (empty if removed)
	Change       ChangeType // Type of change for this line
}

// FullFileDiff represents a complete file with all lines and change information
type FullFileDiff struct {
	OldPath       string         // Path in the old version
	NewPath       string         // Path in the new version
	Lines         []FullFileLine // All lines in the merged view
	ChangeIndices []int          // Indices of lines that have changes (for navigation)
}

// MergeFullFile merges the old and new file content with the parsed diff
// to produce a full file view with change highlighting.
func MergeFullFile(oldContent, newContent string, parsedDiff *FileDiff) *FullFileDiff {
	result := &FullFileDiff{
		OldPath:       parsedDiff.OldPath,
		NewPath:       parsedDiff.NewPath,
		Lines:         make([]FullFileLine, 0),
		ChangeIndices: make([]int, 0),
	}

	// Split content into lines
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	// Build maps of changed lines from the diff
	// Key: line number, Value: line content and type
	removedLines := make(map[int]string) // old line numbers that are removed
	addedLines := make(map[int]string)   // new line numbers that are added

	for _, line := range parsedDiff.Lines {
		switch line.Type {
		case Remove:
			removedLines[line.OldLineNo] = line.Content
		case Add:
			addedLines[line.NewLineNo] = line.Content
		}
	}

	// Two-pointer walk through both files
	oldIdx := 0 // current position in old file (0-indexed)
	newIdx := 0 // current position in new file (0-indexed)

	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		oldLineNo := oldIdx + 1 // 1-indexed line number
		newLineNo := newIdx + 1 // 1-indexed line number

		// Check if current old line is removed
		if oldIdx < len(oldLines) {
			if _, isRemoved := removedLines[oldLineNo]; isRemoved {
				result.Lines = append(result.Lines, FullFileLine{
					LeftLineNo:   oldLineNo,
					RightLineNo:  0,
					LeftContent:  oldLines[oldIdx],
					RightContent: "",
					Change:       Removed,
				})
				result.ChangeIndices = append(result.ChangeIndices, len(result.Lines)-1)
				oldIdx++
				continue
			}
		}

		// Check if current new line is added
		if newIdx < len(newLines) {
			if _, isAdded := addedLines[newLineNo]; isAdded {
				result.Lines = append(result.Lines, FullFileLine{
					LeftLineNo:   0,
					RightLineNo:  newLineNo,
					LeftContent:  "",
					RightContent: newLines[newIdx],
					Change:       Added,
				})
				result.ChangeIndices = append(result.ChangeIndices, len(result.Lines)-1)
				newIdx++
				continue
			}
		}

		// Otherwise it's an unchanged line - advance both pointers
		if oldIdx < len(oldLines) && newIdx < len(newLines) {
			result.Lines = append(result.Lines, FullFileLine{
				LeftLineNo:   oldLineNo,
				RightLineNo:  newLineNo,
				LeftContent:  oldLines[oldIdx],
				RightContent: newLines[newIdx],
				Change:       Unchanged,
			})
			oldIdx++
			newIdx++
		} else if oldIdx < len(oldLines) {
			// Only old lines left - these should be removed
			result.Lines = append(result.Lines, FullFileLine{
				LeftLineNo:   oldLineNo,
				RightLineNo:  0,
				LeftContent:  oldLines[oldIdx],
				RightContent: "",
				Change:       Removed,
			})
			result.ChangeIndices = append(result.ChangeIndices, len(result.Lines)-1)
			oldIdx++
		} else if newIdx < len(newLines) {
			// Only new lines left - these should be added
			result.Lines = append(result.Lines, FullFileLine{
				LeftLineNo:   0,
				RightLineNo:  newLineNo,
				LeftContent:  "",
				RightContent: newLines[newIdx],
				Change:       Added,
			})
			result.ChangeIndices = append(result.ChangeIndices, len(result.Lines)-1)
			newIdx++
		}
	}

	return result
}

// splitLines splits content into lines, handling trailing newlines properly
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	// Remove trailing newline to avoid empty last element
	content = strings.TrimSuffix(content, "\n")
	if content == "" {
		return []string{}
	}

	return strings.Split(content, "\n")
}
