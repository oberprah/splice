package states

import "strings"

// ViewBuilder builds view output with automatic newline handling.
// It ensures views never end with a trailing newline by storing
// lines separately and joining them only when String() is called.
type ViewBuilder struct {
	lines []string
}

// NewViewBuilder creates a new ViewBuilder.
func NewViewBuilder() *ViewBuilder {
	return &ViewBuilder{}
}

// AddLine adds a line to the view. Newlines are automatically
// added between lines (via strings.Join) but not after the last line.
// If the input contains newline characters, they are escaped (shown
// as literal \n) to prevent breaking the line structure while
// preserving visibility of the original content.
func (vb *ViewBuilder) AddLine(line string) {
	// Escape any embedded newlines to maintain single-line structure
	if strings.Contains(line, "\n") {
		line = strings.ReplaceAll(line, "\n", `\n`)
	}
	vb.lines = append(vb.lines, line)
}

// String returns the final view output without trailing newline.
// Lines are joined with "\n" separator.
func (vb *ViewBuilder) String() string {
	return strings.Join(vb.lines, "\n")
}
