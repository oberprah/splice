package states

import (
	"testing"
)

// Per-file helper that adds subdirectory prefix
func assertViewBuilderGolden(t *testing.T, output *ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), "viewbuilder/"+filename, *update)
}

func TestAddSplitView_EqualLineCounts(t *testing.T) {
	left := NewViewBuilder()
	left.AddLine("Left Line 1")
	left.AddLine("Left Line 2")
	left.AddLine("Left Line 3")

	right := NewViewBuilder()
	right.AddLine("Right Line 1")
	right.AddLine("Right Line 2")
	right.AddLine("Right Line 3")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "equal_line_counts.golden")
}

func TestAddSplitView_LeftTaller(t *testing.T) {
	left := NewViewBuilder()
	left.AddLine("Left Line 1")
	left.AddLine("Left Line 2")
	left.AddLine("Left Line 3")
	left.AddLine("Left Line 4")
	left.AddLine("Left Line 5")

	right := NewViewBuilder()
	right.AddLine("Right Line 1")
	right.AddLine("Right Line 2")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "left_taller.golden")
}

func TestAddSplitView_RightTaller(t *testing.T) {
	left := NewViewBuilder()
	left.AddLine("Left Line 1")
	left.AddLine("Left Line 2")

	right := NewViewBuilder()
	right.AddLine("Right Line 1")
	right.AddLine("Right Line 2")
	right.AddLine("Right Line 3")
	right.AddLine("Right Line 4")
	right.AddLine("Right Line 5")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "right_taller.golden")
}

func TestAddSplitView_LeftEmpty(t *testing.T) {
	left := NewViewBuilder()

	right := NewViewBuilder()
	right.AddLine("Right Line 1")
	right.AddLine("Right Line 2")
	right.AddLine("Right Line 3")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "left_empty.golden")
}

func TestAddSplitView_RightEmpty(t *testing.T) {
	left := NewViewBuilder()
	left.AddLine("Left Line 1")
	left.AddLine("Left Line 2")
	left.AddLine("Left Line 3")

	right := NewViewBuilder()

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "right_empty.golden")
}

func TestAddSplitView_BothEmpty(t *testing.T) {
	left := NewViewBuilder()
	right := NewViewBuilder()

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "both_empty.golden")
}

func TestAddSplitView_SingleLine(t *testing.T) {
	left := NewViewBuilder()
	left.AddLine("Single left")

	right := NewViewBuilder()
	right.AddLine("Single right")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "single_line.golden")
}

func TestAddSplitView_MultipleCallsComposable(t *testing.T) {
	// Test that AddSplitView can be called multiple times
	// to build complex layouts (e.g., split view followed by full-width content)
	left := NewViewBuilder()
	left.AddLine("Split Left 1")
	left.AddLine("Split Left 2")

	right := NewViewBuilder()
	right.AddLine("Split Right 1")
	right.AddLine("Split Right 2")

	vb := NewViewBuilder()
	vb.AddLine("Header Line")
	vb.AddSplitView(left, right)
	vb.AddLine("Footer Line")

	assertViewBuilderGolden(t, vb, "multiple_calls.golden")
}

func TestAddSplitView_DifferentWidthContent(t *testing.T) {
	// Test with different width content to ensure lipgloss handles padding
	left := NewViewBuilder()
	left.AddLine("Short")
	left.AddLine("Medium length")
	left.AddLine("This is a very long line of text")

	right := NewViewBuilder()
	right.AddLine("A")
	right.AddLine("AB")
	right.AddLine("ABC")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	assertViewBuilderGolden(t, vb, "different_widths.golden")
}

func TestAddSplitView_NoTrailingNewline(t *testing.T) {
	// Verify that the result doesn't have a trailing newline
	// (following ViewBuilder's convention)
	left := NewViewBuilder()
	left.AddLine("Left")

	right := NewViewBuilder()
	right.AddLine("Right")

	vb := NewViewBuilder()
	vb.AddSplitView(left, right)

	output := vb.String()
	if len(output) > 0 && output[len(output)-1] == '\n' {
		t.Errorf("Output has trailing newline, but ViewBuilder should never end with one")
	}
}
