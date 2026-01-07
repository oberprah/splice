package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/app"
)

// Compile-time check that ViewBuilder implements app.ViewRenderer
var _ app.ViewRenderer = (*ViewBuilder)(nil)

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

// AddSplitView joins two ViewBuilders horizontally with a vertical separator.
// The left and right ViewBuilders are rendered side-by-side with " │ " between them.
// If the two sides have different line counts, lipgloss automatically pads the shorter
// side to align the columns properly.
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder) {
	// Convert each ViewBuilder's lines to a multi-line string
	leftStr := left.String()
	rightStr := right.String()

	// Determine the maximum line count between the two columns
	leftLineCount := len(left.lines)
	rightLineCount := len(right.lines)
	maxLines := leftLineCount
	if rightLineCount > maxLines {
		maxLines = rightLineCount
	}

	// Build a separator string with that many lines (each line being " │ ")
	separatorLines := make([]string, maxLines)
	for i := 0; i < maxLines; i++ {
		separatorLines[i] = " │ "
	}
	separatorStr := strings.Join(separatorLines, "\n")

	// Use lipgloss.JoinHorizontal to join left, separator, and right
	joined := lipgloss.JoinHorizontal(lipgloss.Top, leftStr, separatorStr, rightStr)

	// Add the joined result's lines to the parent ViewBuilder
	for _, line := range strings.Split(joined, "\n") {
		vb.AddLine(line)
	}
}
