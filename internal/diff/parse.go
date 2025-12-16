package diff

import (
	"regexp"
	"strconv"
	"strings"
)

// LineType represents the type of a diff line
type LineType int

const (
	Context LineType = iota // Unchanged line (space prefix)
	Add                     // Added line (+ prefix)
	Remove                  // Removed line (- prefix)
)

// Line represents a single line in the diff
type Line struct {
	Type      LineType
	Content   string // Line content without +/- prefix
	OldLineNo int    // Line number in old file (0 if N/A)
	NewLineNo int    // Line number in new file (0 if N/A)
}

// FileDiff represents a parsed file diff
type FileDiff struct {
	OldPath string
	NewPath string
	Lines   []Line
}

// hunkHeaderRegex matches hunk headers like "@@ -14,8 +14,8 @@" or "@@ -1 +1,2 @@"
var hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)

// ParseUnifiedDiff parses a unified diff string into a FileDiff structure
func ParseUnifiedDiff(raw string) (FileDiff, error) {
	var result FileDiff
	lines := strings.Split(raw, "\n")

	// Remove trailing empty string from split (artifact of splitting)
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var oldLineNo, newLineNo int
	inHunk := false

	for _, line := range lines {
		// Parse file headers
		if strings.HasPrefix(line, "--- a/") {
			result.OldPath = strings.TrimPrefix(line, "--- a/")
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			result.NewPath = strings.TrimPrefix(line, "+++ b/")
			continue
		}

		// Parse hunk headers
		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			oldLineNo, _ = strconv.Atoi(matches[1])
			newLineNo, _ = strconv.Atoi(matches[3])
			inHunk = true
			continue
		}

		// Skip non-hunk content
		if !inHunk {
			continue
		}

		// Parse diff lines
		if len(line) == 0 {
			// Empty line in git output - skip
			// Actual blank lines in source code have a prefix (space/+/-)
			// e.g., " " for context blank, "+" for added blank, "-" for removed blank
			continue
		}

		prefix := line[0]
		content := ""
		if len(line) > 1 {
			content = line[1:]
		}

		switch prefix {
		case ' ':
			// Context line (unchanged)
			result.Lines = append(result.Lines, Line{
				Type:      Context,
				Content:   content,
				OldLineNo: oldLineNo,
				NewLineNo: newLineNo,
			})
			oldLineNo++
			newLineNo++
		case '-':
			// Removed line
			result.Lines = append(result.Lines, Line{
				Type:      Remove,
				Content:   content,
				OldLineNo: oldLineNo,
				NewLineNo: 0,
			})
			oldLineNo++
		case '+':
			// Added line
			result.Lines = append(result.Lines, Line{
				Type:      Add,
				Content:   content,
				OldLineNo: 0,
				NewLineNo: newLineNo,
			})
			newLineNo++
		case '\\':
			// "\ No newline at end of file" - skip
			continue
		default:
			// Unknown prefix, likely end of hunk or other metadata
			inHunk = false
		}
	}

	return result, nil
}
