package diff

import (
	"testing"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/ui/testutils"
)

func createTestDiffState(numLines int) *State {
	// Create left and right file content with identical lines
	leftLines := make([]diff.AlignedLine, numLines)
	rightLines := make([]diff.AlignedLine, numLines)
	for i := 0; i < numLines; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	// Create alignments - all unchanged
	alignments := make([]diff.Alignment, numLines)
	for i := 0; i < numLines; i++ {
		alignments[i] = diff.UnchangedAlignment{
			LeftIdx:  i,
			RightIdx: i,
		}
	}

	return &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: leftLines,
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: rightLines,
			},
			Alignments: alignments,
		},
		ViewportStart:    0,
		CurrentChangeIdx: 0,
		ChangeIndices:    []int{},
	}
}

func TestDiffState_Update_ScrollDown(t *testing.T) {
	s := createTestDiffState(50)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "j" to scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 1 {
		t.Errorf("Expected ViewportStart to be 1, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_ScrollUp(t *testing.T) {
	s := createTestDiffState(50)
	s.ViewportStart = 10
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "k" to scroll up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 9 {
		t.Errorf("Expected ViewportStart to be 9, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_JumpToTop(t *testing.T) {
	s := createTestDiffState(50)
	s.ViewportStart = 25
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_JumpToBottom(t *testing.T) {
	s := createTestDiffState(50)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Available height = 20 - 2 (header) = 18
	// Max viewport = 50 - 18 = 32
	if diffState.ViewportStart != 32 {
		t.Errorf("Expected ViewportStart to be 32, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_ScrollBoundaryTop(t *testing.T) {
	s := createTestDiffState(50)
	s.ViewportStart = 0
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "k" at top boundary
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay at 0
	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to stay at 0, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_ScrollBoundaryBottom(t *testing.T) {
	s := createTestDiffState(50)
	s.ViewportStart = 32 // Already at max
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "j" at bottom boundary
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay at max
	if diffState.ViewportStart != 32 {
		t.Errorf("Expected ViewportStart to stay at 32, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_BackNavigation(t *testing.T) {
	s := createTestDiffState(50)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "q" to go back
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newState, cmd := s.Update(msg, ctx)

	// Should stay in DiffState (it returns a command that produces PopScreenMsg)
	diffState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected to stay in DiffState, got %T", newState)
	}

	// Should return a command that produces PopScreenMsg
	if cmd == nil {
		t.Fatal("Expected command for back navigation")
	}

	// Execute the command to get the message
	result := cmd()
	if result == nil {
		t.Fatal("Expected command to return a message")
	}

	// Verify it's a PopScreenMsg
	if _, ok := result.(core.PopScreenMsg); !ok {
		t.Fatalf("Expected PopScreenMsg, got %T", result)
	}

	// Verify state hasn't changed (stays DiffState until Model handles PopScreenMsg)
	if diffState != s {
		t.Error("Expected state to remain unchanged")
	}
}

func TestDiffState_Update_QuitKeys(t *testing.T) {
	tests := []struct {
		name    string
		keyType tea.KeyType
		runes   []rune
	}{
		{"ctrl+c", tea.KeyCtrlC, nil},
		{"Q", tea.KeyRunes, []rune{'Q'}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createTestDiffState(10)
			ctx := testutils.MockContext{W: 80, H: 20}

			msg := tea.KeyMsg{Type: tt.keyType, Runes: tt.runes}
			_, cmd := s.Update(msg, ctx)

			// Should return quit command
			if cmd == nil {
				t.Error("Expected quit command")
			}
		})
	}
}

func TestDiffState_Update_ArrowKeys(t *testing.T) {
	tests := []struct {
		name       string
		keyType    tea.KeyType
		initialVP  int
		expectedVP int
	}{
		{"down arrow scrolls down", tea.KeyDown, 0, 1},
		{"up arrow scrolls up", tea.KeyUp, 10, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createTestDiffState(50)
			s.ViewportStart = tt.initialVP
			ctx := testutils.MockContext{W: 80, H: 20}

			msg := tea.KeyMsg{Type: tt.keyType}
			newState, _ := s.Update(msg, ctx)

			diffState, ok := newState.(*State)
			if !ok {
				t.Fatal("Expected state to remain DiffState")
			}

			if diffState.ViewportStart != tt.expectedVP {
				t.Errorf("Expected ViewportStart to be %d, got %d", tt.expectedVP, diffState.ViewportStart)
			}
		})
	}
}

func TestDiffState_Update_EmptyDiff(t *testing.T) {
	s := createTestDiffState(0) // Empty diff
	ctx := testutils.MockContext{W: 80, H: 20}

	// Try to scroll in empty diff
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to stay at 0 for empty diff, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_JumpToNextChange(t *testing.T) {
	// Create a diff with changes at specific positions
	s := createTestDiffStateWithChanges(50, []int{5, 15, 30, 45})
	s.ViewportStart = 0
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" to jump to next change
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to first change at index 5
	if diffState.ViewportStart != 5 {
		t.Errorf("Expected ViewportStart to be 5 (first change), got %d", diffState.ViewportStart)
	}

	// Press "n" again to jump to next change
	newState, _ = diffState.Update(msg, ctx)
	diffState = newState.(*State)

	// Should jump to second change at index 15
	if diffState.ViewportStart != 15 {
		t.Errorf("Expected ViewportStart to be 15 (second change), got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_JumpToPreviousChange(t *testing.T) {
	// Create a diff with changes at specific positions
	s := createTestDiffStateWithChanges(50, []int{5, 15, 30, 45})
	s.ViewportStart = 30
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "N" to jump to previous change
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to previous change at index 15
	if diffState.ViewportStart != 15 {
		t.Errorf("Expected ViewportStart to be 15 (previous change), got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_NoChanges(t *testing.T) {
	// Create a diff with no changes
	s := createTestDiffState(50)
	s.ViewportStart = 10
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - should stay at current position
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay at current position
	if diffState.ViewportStart != 10 {
		t.Errorf("Expected ViewportStart to stay at 10 (no changes), got %d", diffState.ViewportStart)
	}
}

// Helper function to create a test diff state with changes at specific positions
func createTestDiffStateWithChanges(numLines int, changeIndices []int) *State {
	changeSet := make(map[int]bool)
	for _, idx := range changeIndices {
		changeSet[idx] = true
	}

	// Create left and right file content
	leftLines := make([]diff.AlignedLine, numLines)
	rightLines := make([]diff.AlignedLine, numLines)
	for i := 0; i < numLines; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	// Create alignments - unchanged for most lines, added for change indices
	alignments := make([]diff.Alignment, numLines)
	for i := 0; i < numLines; i++ {
		if changeSet[i] {
			// Represent changes as added lines
			alignments[i] = diff.AddedAlignment{
				RightIdx: i,
			}
		} else {
			alignments[i] = diff.UnchangedAlignment{
				LeftIdx:  i,
				RightIdx: i,
			}
		}
	}

	return &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: leftLines,
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: rightLines,
			},
			Alignments: alignments,
		},
		ViewportStart:    0,
		CurrentChangeIdx: 0,
		ChangeIndices:    changeIndices,
	}
}

// createTestStateWithSegments creates a test State with segment-based diff data
func createTestStateWithSegments(segments []diff.Segment, leftLines, rightLines []diff.AlignedLine) *State {
	return &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path:  "file.go",
				Lines: leftLines,
			},
			Right: diff.FileContent{
				Path:  "file.go",
				Lines: rightLines,
			},
			Segments: segments,
		},
		SegmentIndex:      0,
		LeftOffset:        0,
		RightOffset:       0,
		ScrollAccumulator: 0,
	}
}

// createTestLines creates a slice of test AlignedLines
func createTestLines(count int) []diff.AlignedLine {
	lines := make([]diff.AlignedLine, count)
	for i := 0; i < count; i++ {
		lines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}
	return lines
}

func TestIsAtStart(t *testing.T) {
	state := &State{
		SegmentIndex: 0,
		LeftOffset:   0,
		RightOffset:  0,
	}

	if !state.isAtStart() {
		t.Error("Expected isAtStart() to return true when at position 0,0,0")
	}

	state.LeftOffset = 1
	if state.isAtStart() {
		t.Error("Expected isAtStart() to return false when LeftOffset > 0")
	}

	state.LeftOffset = 0
	state.SegmentIndex = 1
	if state.isAtStart() {
		t.Error("Expected isAtStart() to return false when SegmentIndex > 0")
	}
}

func TestIsAtEnd(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))

	// At start, not at end
	if state.isAtEnd(5) {
		t.Error("Expected isAtEnd() to return false when at start")
	}

	// Not at end when offsets are less than line count
	state.LeftOffset = 5
	state.RightOffset = 5
	if state.isAtEnd(5) {
		t.Error("Expected isAtEnd() to return false when offsets < line count")
	}

	// At end when offsets equal line count
	state.LeftOffset = 10
	state.RightOffset = 10
	if !state.isAtEnd(5) {
		t.Error("Expected isAtEnd() to return true when offsets >= line count")
	}
}

func TestScrollDownSegment_UnchangedSegment(t *testing.T) {
	// Create unchanged segment with 10 lines
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))

	viewportHeight := 5

	// Scroll down once
	state.scrollDownSegment(viewportHeight)

	if state.LeftOffset != 1 || state.RightOffset != 1 {
		t.Errorf("Expected offsets to be (1,1), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}

	// Both sides should advance together for unchanged segment
	if state.LeftOffset != state.RightOffset {
		t.Error("Expected left and right offsets to be equal for unchanged segment")
	}
}

func TestScrollDownSegment_HunkWithDifferentialScrolling(t *testing.T) {
	// Create hunk with 6 left lines and 2 right lines (ratio 3:1)
	leftHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineRemoved},
		{SourceIdx: 1, Type: diff.HunkLineRemoved},
		{SourceIdx: 2, Type: diff.HunkLineRemoved},
		{SourceIdx: 3, Type: diff.HunkLineRemoved},
		{SourceIdx: 4, Type: diff.HunkLineRemoved},
		{SourceIdx: 5, Type: diff.HunkLineRemoved},
	}
	rightHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineAdded},
		{SourceIdx: 1, Type: diff.HunkLineAdded},
	}

	segments := []diff.Segment{
		diff.HunkSegment{LeftLines: leftHunkLines, RightLines: rightHunkLines},
	}
	state := createTestStateWithSegments(segments, createTestLines(6), createTestLines(2))

	// Use a viewport where hunk is centered (hunk at row 0, viewport 10 means center is 3-7)
	viewportHeight := 10

	// Step 1: Left +1, Right stays (accumulator 1)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 1 || state.RightOffset != 0 {
		t.Errorf("Step 1: Expected (1,0), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}

	// Step 2: Left +1, Right stays (accumulator 2)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 2 || state.RightOffset != 0 {
		t.Errorf("Step 2: Expected (2,0), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}

	// Step 3: Left +1, Right +1 (accumulator resets)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 3 || state.RightOffset != 1 {
		t.Errorf("Step 3: Expected (3,1), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestScrollDownSegment_HunkRightLarger(t *testing.T) {
	// Create hunk with 2 left lines and 6 right lines (ratio 3:1)
	leftHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineRemoved},
		{SourceIdx: 1, Type: diff.HunkLineRemoved},
	}
	rightHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineAdded},
		{SourceIdx: 1, Type: diff.HunkLineAdded},
		{SourceIdx: 2, Type: diff.HunkLineAdded},
		{SourceIdx: 3, Type: diff.HunkLineAdded},
		{SourceIdx: 4, Type: diff.HunkLineAdded},
		{SourceIdx: 5, Type: diff.HunkLineAdded},
	}

	segments := []diff.Segment{
		diff.HunkSegment{LeftLines: leftHunkLines, RightLines: rightHunkLines},
	}
	state := createTestStateWithSegments(segments, createTestLines(2), createTestLines(6))

	viewportHeight := 10

	// Step 1: Right +1, Left stays (accumulator 1)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 0 || state.RightOffset != 1 {
		t.Errorf("Step 1: Expected (0,1), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}

	// Step 2: Right +1, Left stays (accumulator 2)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 0 || state.RightOffset != 2 {
		t.Errorf("Step 2: Expected (0,2), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}

	// Step 3: Right +1, Left +1 (accumulator resets)
	state.scrollDownSegment(viewportHeight)
	if state.LeftOffset != 1 || state.RightOffset != 3 {
		t.Errorf("Step 3: Expected (1,3), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestScrollDownSegment_TransitionToNextSegment(t *testing.T) {
	// Create two unchanged segments
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 3},
		diff.UnchangedSegment{LeftStart: 3, RightStart: 3, Count: 3},
	}
	state := createTestStateWithSegments(segments, createTestLines(6), createTestLines(6))

	viewportHeight := 10

	// Scroll to end of first segment
	state.LeftOffset = 2
	state.RightOffset = 2

	// Next scroll should transition to second segment
	state.scrollDownSegment(viewportHeight)

	if state.SegmentIndex != 1 {
		t.Errorf("Expected SegmentIndex to be 1, got %d", state.SegmentIndex)
	}
	if state.LeftOffset != 0 || state.RightOffset != 0 {
		t.Errorf("Expected offsets to be reset to (0,0), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestScrollUpSegment_UnchangedSegment(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))
	state.LeftOffset = 5
	state.RightOffset = 5

	viewportHeight := 5

	// Scroll up once
	state.scrollUpSegment(viewportHeight)

	if state.LeftOffset != 4 || state.RightOffset != 4 {
		t.Errorf("Expected offsets to be (4,4), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestScrollUpSegment_AtStart(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))

	viewportHeight := 5

	// Try to scroll up at start
	state.scrollUpSegment(viewportHeight)

	// Should stay at start
	if state.SegmentIndex != 0 || state.LeftOffset != 0 || state.RightOffset != 0 {
		t.Errorf("Expected to stay at start, got segment=%d, offsets=(%d,%d)",
			state.SegmentIndex, state.LeftOffset, state.RightOffset)
	}
}

func TestScrollUpSegment_TransitionToPreviousSegment(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 3},
		diff.UnchangedSegment{LeftStart: 3, RightStart: 3, Count: 3},
	}
	state := createTestStateWithSegments(segments, createTestLines(6), createTestLines(6))
	state.SegmentIndex = 1
	state.LeftOffset = 0
	state.RightOffset = 0

	viewportHeight := 10

	// Scroll up should transition to previous segment
	state.scrollUpSegment(viewportHeight)

	if state.SegmentIndex != 0 {
		t.Errorf("Expected SegmentIndex to be 0, got %d", state.SegmentIndex)
	}
	// Should be at offset 2 (one less than end of first segment)
	if state.LeftOffset != 2 || state.RightOffset != 2 {
		t.Errorf("Expected offsets to be (2,2), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestIsHunkCentered(t *testing.T) {
	leftHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineRemoved},
		{SourceIdx: 1, Type: diff.HunkLineRemoved},
		{SourceIdx: 2, Type: diff.HunkLineRemoved},
	}
	rightHunkLines := []diff.HunkLine{
		{SourceIdx: 0, Type: diff.HunkLineAdded},
	}

	segments := []diff.Segment{
		diff.HunkSegment{LeftLines: leftHunkLines, RightLines: rightHunkLines},
	}
	state := createTestStateWithSegments(segments, createTestLines(3), createTestLines(1))

	// With viewport height 10, center zone is rows 3-7 (30%-70%)
	// Hunk at row 0 with 3 lines (0-2) overlaps center zone? No.
	// But actually, the hunk extends from row 0 to row 2, and center is 3-7
	// So hunk does NOT overlap center zone in this case

	// With viewport height 10, center starts at row 3
	// A hunk at position 0 with 3 lines ends at row 2, which is < 3 (centerStart)
	viewportHeight := 10
	if state.isHunkCentered(viewportHeight) {
		t.Error("Expected isHunkCentered() to return false when hunk doesn't overlap center")
	}

	// Create a larger hunk that does overlap center
	largeHunkLines := make([]diff.HunkLine, 8)
	for i := 0; i < 8; i++ {
		largeHunkLines[i] = diff.HunkLine{SourceIdx: i, Type: diff.HunkLineRemoved}
	}
	state.Diff.Segments = []diff.Segment{
		diff.HunkSegment{LeftLines: largeHunkLines, RightLines: rightHunkLines},
	}

	// Hunk with 8 left lines, 1 right line. Visible = 8 lines.
	// Hunk at row 0, extends to row 7. Center is 3-7.
	// Row 0-7 overlaps 3-7. So should be centered.
	if !state.isHunkCentered(viewportHeight) {
		t.Error("Expected isHunkCentered() to return true when hunk overlaps center")
	}
}

func TestIsHunkCentered_NotAHunk(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))

	if state.isHunkCentered(10) {
		t.Error("Expected isHunkCentered() to return false for UnchangedSegment")
	}
}

func TestResetToStart(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
	}
	state := createTestStateWithSegments(segments, createTestLines(10), createTestLines(10))
	state.SegmentIndex = 2
	state.LeftOffset = 5
	state.RightOffset = 3
	state.ScrollAccumulator = 2

	state.resetToStart()

	if state.SegmentIndex != 0 {
		t.Errorf("Expected SegmentIndex to be 0, got %d", state.SegmentIndex)
	}
	if state.LeftOffset != 0 || state.RightOffset != 0 {
		t.Errorf("Expected offsets to be (0,0), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
	if state.ScrollAccumulator != 0 {
		t.Errorf("Expected ScrollAccumulator to be 0, got %d", state.ScrollAccumulator)
	}
}

func TestScrollToEnd(t *testing.T) {
	segments := []diff.Segment{
		diff.UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 20},
	}
	state := createTestStateWithSegments(segments, createTestLines(20), createTestLines(20))

	viewportHeight := 5

	state.scrollToEnd(viewportHeight)

	// With 20 lines and viewport of 5, should start at offset 15
	if state.LeftOffset != 15 || state.RightOffset != 15 {
		t.Errorf("Expected offsets to be (15,15), got (%d,%d)", state.LeftOffset, state.RightOffset)
	}
}

func TestCalculateViewportHeight(t *testing.T) {
	state := &State{}

	height := state.calculateViewportHeight(24)

	// 24 - 2 (header lines) = 22
	if height != 22 {
		t.Errorf("Expected viewport height to be 22, got %d", height)
	}

	// Test with small height
	height = state.calculateViewportHeight(2)
	if height != 1 {
		t.Errorf("Expected minimum viewport height of 1, got %d", height)
	}
}
