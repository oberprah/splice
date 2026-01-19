package diff

import (
	"fmt"
	"os"
	"testing"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// createTestDiffSource creates a CommitRangeDiffSource for testing
func createTestDiffSource(commit core.GitCommit) core.DiffSource {
	return core.NewSingleCommitRange(commit).ToDiffSource()
}

func createTestDiffState(numLines int) *State {
	// Create unchanged line pairs for all lines
	linePairs := make([]diff.LinePair, numLines)
	for i := 0; i < numLines; i++ {
		linePairs[i] = diff.LinePair{
			LeftLineNo:  i + 1,
			RightLineNo: i + 1,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	var blocks []diff.Block
	if numLines > 0 {
		blocks = []diff.Block{
			diff.UnchangedBlock{Lines: linePairs},
		}
	}

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source:          createTestDiffSource(commit),
		File:            core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff:            &diff.FileDiff{Path: "file.go", Blocks: blocks},
		ViewportStart:   0,
		CurrentBlockIdx: -1,
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

	// Press "p" to jump to previous change (changed from "N")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
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

// Smart Navigation Tests

func TestNavigateToNextChange_ScrollsThroughMultiScreenChange(t *testing.T) {
	// Create a diff with a change block that spans multiple screens (30 lines)
	// Viewport height is 20, available height is 18 (minus 2 header lines)
	s := createTestDiffStateWithMultiScreenChange(30)
	s.ViewportStart = 5 // Start at beginning of change block (after 5 unchanged lines)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - should scroll down half page, not jump
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Available height = 20 - 2 = 18, half page = 9
	// Should scroll from 5 to 14
	if diffState.ViewportStart != 14 {
		t.Errorf("Expected ViewportStart to be 14 (scrolled half page), got %d", diffState.ViewportStart)
	}
}

func TestNavigateToNextChange_JumpsWhenChangeFullyVisible(t *testing.T) {
	// Create a diff with two small change blocks (1 line each)
	// They should fit entirely in the viewport
	s := createTestDiffStateWithChanges(50, []int{5, 25})
	s.ViewportStart = 5 // At first change
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - should jump to next change since current change is fully visible
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to second change at index 25 (but adjusted for block positions)
	// Block structure: 5 unchanged, 1 change, 19 unchanged, 1 change, 24 unchanged
	// Second change starts at line 5 + 1 + 19 = 25
	if diffState.ViewportStart != 25 {
		t.Errorf("Expected ViewportStart to be 25 (second change), got %d", diffState.ViewportStart)
	}
}

func TestNavigateToPrevChange_ScrollsThroughMultiScreenChange(t *testing.T) {
	// Create a diff with a change block that spans multiple screens (30 lines)
	s := createTestDiffStateWithMultiScreenChange(30)
	// Start in the middle of the change block
	s.ViewportStart = 20 // Deep into the change block
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - should scroll up half page since we're not at the start
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Half page = 9, should scroll from 20 to 11
	if diffState.ViewportStart != 11 {
		t.Errorf("Expected ViewportStart to be 11 (scrolled up half page), got %d", diffState.ViewportStart)
	}
}

func TestNavigateToPrevChange_JumpsWhenAtChangeStart(t *testing.T) {
	// Create a diff with two small change blocks
	s := createTestDiffStateWithChanges(50, []int{5, 25})
	s.ViewportStart = 25 // At second change start
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - should jump to previous change since we're at the start
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to first change at index 5
	if diffState.ViewportStart != 5 {
		t.Errorf("Expected ViewportStart to be 5 (first change), got %d", diffState.ViewportStart)
	}
}

func TestNavigateToPrevChange_JumpsToStartWhenWithinChange(t *testing.T) {
	// Create a diff with a change block
	s := createTestDiffStateWithChanges(50, []int{5, 15, 30})
	s.ViewportStart = 16 // One line past the start of second change (at position 15)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - should jump to start of current change since start is visible
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to start of second change at position 15
	if diffState.ViewportStart != 15 {
		t.Errorf("Expected ViewportStart to be 15 (start of current change), got %d", diffState.ViewportStart)
	}
}

func TestNavigateToNextChange_StaysAtLastChange(t *testing.T) {
	// Create a diff with changes, position at the last one
	s := createTestDiffStateWithChanges(50, []int{5, 45})
	s.ViewportStart = 45 // At last change
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - should stay in place (no more changes)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay at current position
	if diffState.ViewportStart != 45 {
		t.Errorf("Expected ViewportStart to stay at 45, got %d", diffState.ViewportStart)
	}
}

func TestNavigateToPrevChange_StaysAtFirstChange(t *testing.T) {
	// Create a diff with changes, position at the first one
	s := createTestDiffStateWithChanges(50, []int{5, 45})
	s.ViewportStart = 5 // At first change
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - should stay in place (no previous changes)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay at current position
	if diffState.ViewportStart != 5 {
		t.Errorf("Expected ViewportStart to stay at 5, got %d", diffState.ViewportStart)
	}
}

func TestNavigateToNextChange_FromUnchangedContent(t *testing.T) {
	// Create a diff with changes, starting from unchanged content
	s := createTestDiffStateWithChanges(50, []int{10, 30})
	s.ViewportStart = 0 // At the beginning (unchanged content)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - should jump to first change
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should jump to first change at position 10
	if diffState.ViewportStart != 10 {
		t.Errorf("Expected ViewportStart to be 10 (first change), got %d", diffState.ViewportStart)
	}
}

// Helper function to create a test diff state with a multi-screen change block
func createTestDiffStateWithMultiScreenChange(changeBlockSize int) *State {
	var blocks []diff.Block

	// 5 unchanged lines first
	unchangedLines := make([]diff.LinePair, 5)
	for i := 0; i < 5; i++ {
		unchangedLines[i] = diff.LinePair{
			LeftLineNo:  i + 1,
			RightLineNo: i + 1,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}},
		}
	}
	blocks = append(blocks, diff.UnchangedBlock{Lines: unchangedLines})

	// Large change block
	changeLines := make([]diff.ChangeLine, changeBlockSize)
	for i := 0; i < changeBlockSize; i++ {
		changeLines[i] = diff.AddedLine{
			RightLineNo: 6 + i,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "added line"}},
		}
	}
	blocks = append(blocks, diff.ChangeBlock{Lines: changeLines})

	// 5 more unchanged lines at the end
	unchangedLinesEnd := make([]diff.LinePair, 5)
	for i := 0; i < 5; i++ {
		unchangedLinesEnd[i] = diff.LinePair{
			LeftLineNo:  6 + changeBlockSize + i,
			RightLineNo: 6 + changeBlockSize + i,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}},
		}
	}
	blocks = append(blocks, diff.UnchangedBlock{Lines: unchangedLinesEnd})

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source:          createTestDiffSource(commit),
		File:            core.FileChange{Path: "file.go", Additions: changeBlockSize, Deletions: 0},
		Diff:            &diff.FileDiff{Path: "file.go", Blocks: blocks},
		ViewportStart:   0,
		CurrentBlockIdx: -1,
	}
}

// Helper function to create a test diff state with changes at specific positions
// The structure uses blocks: unchanged lines before each change, then the change block
func createTestDiffStateWithChanges(numLines int, changeIndices []int) *State {
	if len(changeIndices) == 0 {
		return createTestDiffState(numLines)
	}

	// Sort change indices and build blocks
	var blocks []diff.Block
	lastIdx := 0

	for _, changeIdx := range changeIndices {
		// Add unchanged block for lines before this change
		if changeIdx > lastIdx {
			linePairs := make([]diff.LinePair, changeIdx-lastIdx)
			for i := lastIdx; i < changeIdx; i++ {
				linePairs[i-lastIdx] = diff.LinePair{
					LeftLineNo:  i + 1,
					RightLineNo: i + 1,
					Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
				}
			}
			blocks = append(blocks, diff.UnchangedBlock{Lines: linePairs})
		}

		// Add change block for this change (single added line)
		blocks = append(blocks, diff.ChangeBlock{Lines: []diff.ChangeLine{
			diff.AddedLine{
				RightLineNo: changeIdx + 1,
				Tokens:      []highlight.Token{{Type: chroma.Text, Value: "added line"}},
			},
		}})
		lastIdx = changeIdx + 1
	}

	// Add remaining unchanged lines after last change
	if lastIdx < numLines {
		linePairs := make([]diff.LinePair, numLines-lastIdx)
		for i := lastIdx; i < numLines; i++ {
			linePairs[i-lastIdx] = diff.LinePair{
				LeftLineNo:  i + 1,
				RightLineNo: i + 1,
				Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
			}
		}
		blocks = append(blocks, diff.UnchangedBlock{Lines: linePairs})
	}

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source:          createTestDiffSource(commit),
		File:            core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff:            &diff.FileDiff{Path: "file.go", Blocks: blocks},
		ViewportStart:   0,
		CurrentBlockIdx: -1,
	}
}

func TestGetCurrentFileLineNumber_UnchangedBlock(t *testing.T) {
	s := createTestDiffState(10)
	s.ViewportStart = 5

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Line at index 5 (0-based) has RightLineNo=6
	expected := 6
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_ModifiedLine(t *testing.T) {
	// Create a diff with a modified line at line 4
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 1"}}},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 2"}}},
					{LeftLineNo: 3, RightLineNo: 3, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 3"}}},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.ModifiedLine{
						LeftLineNo:  4,
						RightLineNo: 4,
						LeftTokens:  []highlight.Token{{Type: chroma.Text, Value: "old line 4"}},
						RightTokens: []highlight.Token{{Type: chroma.Text, Value: "new line 4"}},
					},
				}},
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 5, RightLineNo: 5, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 5"}}},
				}},
			},
		},
		ViewportStart: 3, // Line 4 (0-indexed position 3)
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// ModifiedLine at viewport position 3 has RightLineNo=4
	expected := 4
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_AddedLine(t *testing.T) {
	// Create a diff with an added line
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 1"}}},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 2"}}},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.AddedLine{RightLineNo: 3, Tokens: []highlight.Token{{Type: chroma.Text, Value: "new line"}}},
				}},
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 3, RightLineNo: 4, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 3"}}},
					{LeftLineNo: 4, RightLineNo: 5, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 4"}}},
				}},
			},
		},
		ViewportStart: 2, // Position 2 (0-indexed) is the added line
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// AddedLine at position 2 has RightLineNo=3
	expected := 3
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_RemovedLine(t *testing.T) {
	// Create a diff with a removed line followed by unchanged lines
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 1"}}},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 2"}}},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.RemovedLine{LeftLineNo: 3, Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed line"}}},
				}},
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 4, RightLineNo: 3, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 4"}}},
					{LeftLineNo: 5, RightLineNo: 4, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 5"}}},
				}},
			},
		},
		ViewportStart: 2, // Position 2 (0-indexed) is the removed line
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// RemovedLine at position 2 - should search forward and find next line with RightLineNo
	// Next line at position 3 has RightLineNo=3
	expected := 3
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_RemovedLine_NoFollowingRightLineNo(t *testing.T) {
	// Create a diff where all remaining lines are removed
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 1"}}},
					{LeftLineNo: 2, RightLineNo: 2, Tokens: []highlight.Token{{Type: chroma.Text, Value: "line 2"}}},
				}},
				diff.ChangeBlock{Lines: []diff.ChangeLine{
					diff.RemovedLine{LeftLineNo: 3, Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed 1"}}},
					diff.RemovedLine{LeftLineNo: 4, Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed 2"}}},
					diff.RemovedLine{LeftLineNo: 5, Tokens: []highlight.Token{{Type: chroma.Text, Value: "removed 3"}}},
				}},
			},
		},
		ViewportStart: 2, // Position 2 (0-indexed) is the first removed line
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// No following lines with RightLineNo, should fall back to line 1
	expected := 1
	if lineNo != expected {
		t.Errorf("Expected line number %d (fallback), got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_NilDiff(t *testing.T) {
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source:        createTestDiffSource(commit),
		File:          core.FileChange{Path: "file.go"},
		Diff:          nil,
		ViewportStart: 0,
	}

	_, err := s.getCurrentFileLineNumber()
	if err == nil {
		t.Fatal("Expected error for nil diff, got nil")
	}

	expectedMsg := "no diff available"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestGetCurrentFileLineNumber_EmptyBlocks(t *testing.T) {
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go"},
		Diff: &diff.FileDiff{
			Path:   "file.go",
			Blocks: []diff.Block{},
		},
		ViewportStart: 0,
	}

	_, err := s.getCurrentFileLineNumber()
	if err == nil {
		t.Fatal("Expected error for empty blocks, got nil")
	}

	expectedMsg := "diff has no blocks"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestGetCurrentFileLineNumber_ViewportOutOfRange(t *testing.T) {
	s := createTestDiffState(5)
	s.ViewportStart = 10 // Out of range

	_, err := s.getCurrentFileLineNumber()
	if err == nil {
		t.Fatal("Expected error for viewport out of range, got nil")
	}

	expectedMsg := "viewport position out of range"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// Tests for getEditor()

func TestGetEditor_BothUnset(t *testing.T) {
	// Save and clear environment variables
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
		_ = os.Setenv("VISUAL", oldVisual)
	}()

	_ = os.Setenv("EDITOR", "")
	_ = os.Setenv("VISUAL", "")

	_, err := getEditor()
	if err == nil {
		t.Fatal("Expected error when both EDITOR and VISUAL are unset, got nil")
	}

	expectedMsg := "no editor configured (set $EDITOR or $VISUAL)"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestGetEditor_OnlyEditorSet(t *testing.T) {
	// Save and restore environment variables
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
		_ = os.Setenv("VISUAL", oldVisual)
	}()

	_ = os.Setenv("EDITOR", "vim")
	_ = os.Setenv("VISUAL", "")

	editor, err := getEditor()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if editor != "vim" {
		t.Errorf("Expected editor to be 'vim', got %q", editor)
	}
}

func TestGetEditor_OnlyVisualSet(t *testing.T) {
	// Save and restore environment variables
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
		_ = os.Setenv("VISUAL", oldVisual)
	}()

	_ = os.Setenv("EDITOR", "")
	_ = os.Setenv("VISUAL", "nano")

	editor, err := getEditor()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if editor != "nano" {
		t.Errorf("Expected editor to be 'nano', got %q", editor)
	}
}

func TestGetEditor_BothSet_EditorTakesPrecedence(t *testing.T) {
	// Save and restore environment variables
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
		_ = os.Setenv("VISUAL", oldVisual)
	}()

	_ = os.Setenv("EDITOR", "vim")
	_ = os.Setenv("VISUAL", "nano")

	editor, err := getEditor()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if editor != "vim" {
		t.Errorf("Expected EDITOR to take precedence, got %q instead of 'vim'", editor)
	}
}

// Tests for validation logic in openFileInEditor

func TestOpenFileInEditor_BinaryFile(t *testing.T) {
	s := createTestDiffState(10)
	s.File.IsBinary = true

	// Save and restore EDITOR
	oldEditor := os.Getenv("EDITOR")
	defer func() { _ = os.Setenv("EDITOR", oldEditor) }()
	_ = os.Setenv("EDITOR", "vim")

	cmd := s.openFileInEditor()
	if cmd == nil {
		t.Fatal("Expected command, got nil")
	}

	// Execute the command to get the message
	msg := cmd()
	editorMsg, ok := msg.(EditorFinishedMsg)
	if !ok {
		t.Fatalf("Expected EditorFinishedMsg, got %T", msg)
	}

	if editorMsg.err == nil {
		t.Fatal("Expected error for binary file, got nil")
	}

	expectedMsg := "cannot open binary file in editor"
	if editorMsg.err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, editorMsg.err.Error())
	}
}

func TestOpenFileInEditor_DeletedFile(t *testing.T) {
	s := createTestDiffState(10)
	s.File.Status = "D"

	// Save and restore EDITOR
	oldEditor := os.Getenv("EDITOR")
	defer func() { _ = os.Setenv("EDITOR", oldEditor) }()
	_ = os.Setenv("EDITOR", "vim")

	cmd := s.openFileInEditor()
	if cmd == nil {
		t.Fatal("Expected command, got nil")
	}

	// Execute the command to get the message
	msg := cmd()
	editorMsg, ok := msg.(EditorFinishedMsg)
	if !ok {
		t.Fatalf("Expected EditorFinishedMsg, got %T", msg)
	}

	if editorMsg.err == nil {
		t.Fatal("Expected error for deleted file, got nil")
	}

	expectedMsg := "cannot open: file has been deleted"
	if editorMsg.err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, editorMsg.err.Error())
	}
}

func TestOpenFileInEditor_NoEditor(t *testing.T) {
	s := createTestDiffState(10)

	// Save and clear environment variables
	oldEditor := os.Getenv("EDITOR")
	oldVisual := os.Getenv("VISUAL")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
		_ = os.Setenv("VISUAL", oldVisual)
	}()

	_ = os.Setenv("EDITOR", "")
	_ = os.Setenv("VISUAL", "")

	cmd := s.openFileInEditor()
	if cmd == nil {
		t.Fatal("Expected command, got nil")
	}

	// Execute the command to get the message
	msg := cmd()
	editorMsg, ok := msg.(EditorFinishedMsg)
	if !ok {
		t.Fatalf("Expected EditorFinishedMsg, got %T", msg)
	}

	if editorMsg.err == nil {
		t.Fatal("Expected error when no editor configured, got nil")
	}

	expectedMsg := "no editor configured (set $EDITOR or $VISUAL)"
	if editorMsg.err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, editorMsg.err.Error())
	}
}

// Tests for EditorFinishedMsg handling in Update

func TestDiffState_Update_EditorFinishedMsg_WithError(t *testing.T) {
	s := createTestDiffState(10)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Create an EditorFinishedMsg with an error
	msg := EditorFinishedMsg{
		err: fmt.Errorf("editor failed"),
	}

	newState, cmd := s.Update(msg, ctx)

	// State should remain unchanged
	diffState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain DiffState, got %T", newState)
	}
	if diffState != s {
		t.Error("Expected state to remain unchanged")
	}

	// Should return a command that produces PushErrorScreenMsg
	if cmd == nil {
		t.Fatal("Expected command for error case")
	}

	result := cmd()
	if result == nil {
		t.Fatal("Expected command to return a message")
	}

	errorMsg, ok := result.(core.PushErrorScreenMsg)
	if !ok {
		t.Fatalf("Expected PushErrorScreenMsg, got %T", result)
	}

	if errorMsg.Err == nil {
		t.Fatal("Expected error in PushErrorScreenMsg, got nil")
	}

	expectedErrorMsg := "editor failed"
	if errorMsg.Err.Error() != expectedErrorMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrorMsg, errorMsg.Err.Error())
	}
}

func TestDiffState_Update_EditorFinishedMsg_NoError(t *testing.T) {
	s := createTestDiffState(10)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Create an EditorFinishedMsg with no error (success case)
	msg := EditorFinishedMsg{
		err: nil,
	}

	newState, cmd := s.Update(msg, ctx)

	// State should remain unchanged
	diffState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain DiffState, got %T", newState)
	}
	if diffState != s {
		t.Error("Expected state to remain unchanged")
	}

	// Should not return a command (success case, just resume)
	if cmd != nil {
		t.Error("Expected nil command for success case")
	}
}

func TestDiffState_Update_OpenInEditor(t *testing.T) {
	s := createTestDiffState(10)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Save and restore EDITOR
	oldEditor := os.Getenv("EDITOR")
	defer func() {
		_ = os.Setenv("EDITOR", oldEditor)
	}()
	_ = os.Setenv("EDITOR", "vim")

	// Press "o" to open in editor
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	newState, cmd := s.Update(msg, ctx)

	// State should remain unchanged
	diffState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain DiffState, got %T", newState)
	}
	if diffState != s {
		t.Error("Expected state to remain unchanged")
	}

	// Should return a command (the editor launch command)
	if cmd == nil {
		t.Fatal("Expected non-nil command to launch editor")
	}
}

// File Navigation Tests

func createTestDiffStateWithFiles(numLines int, numFiles int, currentFileIndex int) *State {
	files := make([]core.FileChange, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = core.FileChange{
			Path:      fmt.Sprintf("file%d.go", i),
			Status:    "M",
			Additions: 5,
			Deletions: 3,
		}
	}

	// Create unchanged line pairs for all lines
	linePairs := make([]diff.LinePair, numLines)
	for i := 0; i < numLines; i++ {
		linePairs[i] = diff.LinePair{
			LeftLineNo:  i + 1,
			RightLineNo: i + 1,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	var blocks []diff.Block
	if numLines > 0 {
		blocks = []diff.Block{
			diff.UnchangedBlock{Lines: linePairs},
		}
	}

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source:          createTestDiffSource(commit),
		Files:           files,
		FileIndex:       currentFileIndex,
		File:            files[currentFileIndex],
		Diff:            &diff.FileDiff{Path: files[currentFileIndex].Path, Blocks: blocks},
		ViewportStart:   0,
		CurrentBlockIdx: -1,
	}
}

func TestNavigateToNextFile_AtFirstFile(t *testing.T) {
	// FileIndex=0, Files has 3 files
	s := createTestDiffStateWithFiles(10, 3, 0)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "]" to navigate to next file
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should remain the same (unchanged until async load completes)
	if diffState != s {
		t.Error("Expected state to remain unchanged until async load completes")
	}

	// Should return a command to load the next file's diff
	if cmd == nil {
		t.Fatal("Expected non-nil command to load next file's diff")
	}

	// Execute the command - should return a DiffLoadedMsg
	result := cmd()
	loadedMsg, ok := result.(core.DiffLoadedMsg)
	if !ok {
		t.Fatalf("Expected DiffLoadedMsg, got %T", result)
	}

	// Should be loading file index 1
	if loadedMsg.FileIndex != 1 {
		t.Errorf("Expected FileIndex to be 1, got %d", loadedMsg.FileIndex)
	}
}

func TestNavigateToNextFile_AtLastFile(t *testing.T) {
	// FileIndex=2, Files has 3 files
	s := createTestDiffStateWithFiles(10, 3, 2)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "]" to navigate to next file (should stay in place)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay in place (same state)
	if diffState != s {
		t.Error("Expected state to remain unchanged at last file")
	}

	// Should return nil command (no navigation possible)
	if cmd != nil {
		t.Error("Expected nil command when at last file")
	}
}

func TestNavigateToPrevFile_AtFirstFile(t *testing.T) {
	// FileIndex=0
	s := createTestDiffStateWithFiles(10, 3, 0)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "[" to navigate to previous file (should stay in place)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay in place (same state)
	if diffState != s {
		t.Error("Expected state to remain unchanged at first file")
	}

	// Should return nil command (no navigation possible)
	if cmd != nil {
		t.Error("Expected nil command when at first file")
	}
}

func TestNavigateToPrevFile_AtLastFile(t *testing.T) {
	// FileIndex=2, Files has 3 files
	s := createTestDiffStateWithFiles(10, 3, 2)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "[" to navigate to previous file
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should remain the same (unchanged until async load completes)
	if diffState != s {
		t.Error("Expected state to remain unchanged until async load completes")
	}

	// Should return a command to load the previous file's diff
	if cmd == nil {
		t.Fatal("Expected non-nil command to load previous file's diff")
	}

	// Execute the command - should return a DiffLoadedMsg
	result := cmd()
	loadedMsg, ok := result.(core.DiffLoadedMsg)
	if !ok {
		t.Fatalf("Expected DiffLoadedMsg, got %T", result)
	}

	// Should be loading file index 1
	if loadedMsg.FileIndex != 1 {
		t.Errorf("Expected FileIndex to be 1, got %d", loadedMsg.FileIndex)
	}
}

func createTestDiffStateWithChangesAndFiles(numLines int, changeIndices []int, numFiles int, currentFileIndex int) *State {
	files := make([]core.FileChange, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = core.FileChange{
			Path:      fmt.Sprintf("file%d.go", i),
			Status:    "M",
			Additions: 5,
			Deletions: 3,
		}
	}

	if len(changeIndices) == 0 {
		s := createTestDiffStateWithFiles(numLines, numFiles, currentFileIndex)
		return s
	}

	// Build blocks with changes
	var blocks []diff.Block
	lastIdx := 0

	for _, changeIdx := range changeIndices {
		// Add unchanged block for lines before this change
		if changeIdx > lastIdx {
			linePairs := make([]diff.LinePair, changeIdx-lastIdx)
			for i := lastIdx; i < changeIdx; i++ {
				linePairs[i-lastIdx] = diff.LinePair{
					LeftLineNo:  i + 1,
					RightLineNo: i + 1,
					Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
				}
			}
			blocks = append(blocks, diff.UnchangedBlock{Lines: linePairs})
		}

		// Add change block for this change (single added line)
		blocks = append(blocks, diff.ChangeBlock{Lines: []diff.ChangeLine{
			diff.AddedLine{
				RightLineNo: changeIdx + 1,
				Tokens:      []highlight.Token{{Type: chroma.Text, Value: "added line"}},
			},
		}})
		lastIdx = changeIdx + 1
	}

	// Add remaining unchanged lines after last change
	if lastIdx < numLines {
		linePairs := make([]diff.LinePair, numLines-lastIdx)
		for i := lastIdx; i < numLines; i++ {
			linePairs[i-lastIdx] = diff.LinePair{
				LeftLineNo:  i + 1,
				RightLineNo: i + 1,
				Tokens:      []highlight.Token{{Type: chroma.Text, Value: "test line"}},
			}
		}
		blocks = append(blocks, diff.UnchangedBlock{Lines: linePairs})
	}

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source:          createTestDiffSource(commit),
		Files:           files,
		FileIndex:       currentFileIndex,
		File:            files[currentFileIndex],
		Diff:            &diff.FileDiff{Path: files[currentFileIndex].Path, Blocks: blocks},
		ViewportStart:   0,
		CurrentBlockIdx: -1,
	}
}

func TestNavigateToNextChange_CrossesFiles(t *testing.T) {
	// At last change block, pressing 'n' should trigger file navigation
	// Create a state with a single change at position 5, positioned at that change
	s := createTestDiffStateWithChangesAndFiles(20, []int{5}, 3, 0)
	s.ViewportStart = 5 // Position at the change block
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - since we're at the last (and only) change, should navigate to next file
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should remain unchanged until async load completes
	if diffState != s {
		t.Error("Expected state to remain unchanged until async load completes")
	}

	// Should return a command to load next file
	if cmd == nil {
		t.Fatal("Expected non-nil command to load next file's diff")
	}

	// Execute the command - should return a DiffLoadedMsg for the next file
	result := cmd()
	loadedMsg, ok := result.(core.DiffLoadedMsg)
	if !ok {
		t.Fatalf("Expected DiffLoadedMsg, got %T", result)
	}

	// Should be loading file index 1 (next file)
	if loadedMsg.FileIndex != 1 {
		t.Errorf("Expected FileIndex to be 1, got %d", loadedMsg.FileIndex)
	}
}

func TestNavigateToNextChange_ClampedJumpAdvancesToNextFile(t *testing.T) {
	// Change block start beyond max viewport should still advance to the next file
	// Layout: 3 unchanged, 1 change, 1 unchanged, 1 change, 1 unchanged (7 total lines)
	// Height 6 => availableHeight 4, maxViewportStart = 3, next change starts at 5
	s := createTestDiffStateWithChangesAndFiles(7, []int{3, 5}, 2, 0)
	s.ViewportStart = 3

	mockDiffResult := &core.FullFileDiffResult{
		OldContent: "old\n",
		NewContent: "new\n",
		DiffOutput: "@@ -1 +1 @@\n-old\n+new\n",
		OldPath:    "file1.go",
		NewPath:    "file1.go",
	}

	ctx := testutils.MockContext{
		W: 80,
		H: 6,
		MockFetchFullFileDiff: func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
			return mockDiffResult, nil
		},
	}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}

	// First "n" attempts to jump to the last change block (clamped)
	_, cmd := s.Update(msg, ctx)
	if cmd != nil {
		t.Fatal("Expected no file navigation on first n")
	}

	// Second "n" should move to the next file
	_, cmd = s.Update(msg, ctx)
	if cmd == nil {
		t.Fatal("Expected non-nil command to load next file's diff")
	}

	result := cmd()
	loadedMsg, ok := result.(core.DiffLoadedMsg)
	if !ok {
		t.Fatalf("Expected DiffLoadedMsg, got %T", result)
	}
	if loadedMsg.FileIndex != 1 {
		t.Errorf("Expected FileIndex to be 1, got %d", loadedMsg.FileIndex)
	}
}

func TestNavigateToPrevChange_CrossesFiles(t *testing.T) {
	// At first change block, pressing 'p' should trigger file navigation
	// Create a state with a single change at position 5, positioned at that change
	s := createTestDiffStateWithChangesAndFiles(20, []int{5}, 3, 1) // Start at file 1 (middle)
	s.ViewportStart = 5                                             // Position at the change block
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - since we're at the first (and only) change, should navigate to previous file
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should remain unchanged until async load completes
	if diffState != s {
		t.Error("Expected state to remain unchanged until async load completes")
	}

	// Should return a command to load previous file
	if cmd == nil {
		t.Fatal("Expected non-nil command to load previous file's diff")
	}

	// Execute the command - should return a DiffLoadedMsg for the previous file
	result := cmd()
	loadedMsg, ok := result.(core.DiffLoadedMsg)
	if !ok {
		t.Fatalf("Expected DiffLoadedMsg, got %T", result)
	}

	// Should be loading file index 0 (previous file)
	if loadedMsg.FileIndex != 0 {
		t.Errorf("Expected FileIndex to be 0, got %d", loadedMsg.FileIndex)
	}
}

func TestNavigateToNextChange_StaysAtLastFile(t *testing.T) {
	// At last change in last file, should stay in place
	s := createTestDiffStateWithChangesAndFiles(20, []int{5}, 3, 2) // Last file (index 2)
	s.ViewportStart = 5                                             // Position at the change block
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "n" - at last file and last change, should stay in place
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay in place
	if diffState != s {
		t.Error("Expected state to remain unchanged at last file with no more changes")
	}

	// No command should be returned
	if cmd != nil {
		t.Error("Expected nil command when at last file with no more changes")
	}
}

func TestNavigateToPrevChange_StaysAtFirstFile(t *testing.T) {
	// At first change in first file, should stay in place
	s := createTestDiffStateWithChangesAndFiles(20, []int{5}, 3, 0) // First file (index 0)
	s.ViewportStart = 5                                             // Position at the change block
	ctx := testutils.MockContext{W: 80, H: 20}

	// Press "p" - at first file and first change, should stay in place
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// Should stay in place
	if diffState != s {
		t.Error("Expected state to remain unchanged at first file with no previous changes")
	}

	// No command should be returned
	if cmd != nil {
		t.Error("Expected nil command when at first file with no previous changes")
	}
}

func TestDiffLoadedMsg_UpdatesState(t *testing.T) {
	// Create initial state
	s := createTestDiffStateWithFiles(10, 3, 0)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Create a new diff for the next file
	newDiff := &diff.FileDiff{
		Path: "file1.go",
		Blocks: []diff.Block{
			diff.ChangeBlock{Lines: []diff.ChangeLine{
				diff.AddedLine{
					RightLineNo: 1,
					Tokens:      []highlight.Token{{Type: chroma.Text, Value: "new added line"}},
				},
			}},
		},
	}

	// Create a DiffLoadedMsg for file index 1
	msg := core.DiffLoadedMsg{
		Source:    s.Source,
		Files:     s.Files,
		FileIndex: 1,
		File:      s.Files[1],
		Diff:      newDiff,
	}

	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should be updated
	if diffState.FileIndex != 1 {
		t.Errorf("Expected FileIndex to be 1, got %d", diffState.FileIndex)
	}

	if diffState.File.Path != "file1.go" {
		t.Errorf("Expected File.Path to be 'file1.go', got %s", diffState.File.Path)
	}

	if diffState.Diff != newDiff {
		t.Error("Expected Diff to be updated to newDiff")
	}

	// ViewportStart should be positioned at first change (position 0 since change is first block)
	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0 (first change), got %d", diffState.ViewportStart)
	}

	// CurrentBlockIdx should be 0 (first change block)
	if diffState.CurrentBlockIdx != 0 {
		t.Errorf("Expected CurrentBlockIdx to be 0, got %d", diffState.CurrentBlockIdx)
	}

	// No command should be returned
	if cmd != nil {
		t.Error("Expected nil command after handling DiffLoadedMsg")
	}
}

func TestDiffLoadedMsg_WithError_StaysInPlace(t *testing.T) {
	// Create initial state
	s := createTestDiffStateWithFiles(10, 3, 0)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Create a DiffLoadedMsg with an error
	msg := core.DiffLoadedMsg{
		Source:    s.Source,
		Files:     s.Files,
		FileIndex: 1,
		File:      s.Files[1],
		Err:       fmt.Errorf("failed to load diff"),
	}

	newState, cmd := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// State should remain unchanged on error
	if diffState != s {
		t.Error("Expected state to remain unchanged on error")
	}

	// No command should be returned
	if cmd != nil {
		t.Error("Expected nil command when handling error")
	}
}

func TestPositionAtFirstChange(t *testing.T) {
	// Create a state with changes at specific positions
	s := createTestDiffStateWithChanges(50, []int{10, 30})
	s.ViewportStart = 0
	s.CurrentBlockIdx = -1

	// Call positionAtFirstChange
	s.positionAtFirstChange()

	// ViewportStart should be at first change (position 10)
	if s.ViewportStart != 10 {
		t.Errorf("Expected ViewportStart to be 10 (first change), got %d", s.ViewportStart)
	}

	// CurrentBlockIdx should be the index of the first change block (1)
	// Block structure: unchanged (0-9), change (10), unchanged (11-29), change (30), unchanged (31-49)
	if s.CurrentBlockIdx != 1 {
		t.Errorf("Expected CurrentBlockIdx to be 1 (first change block), got %d", s.CurrentBlockIdx)
	}
}

func TestPositionAtFirstChange_NoChanges(t *testing.T) {
	// Create a state with no changes (all unchanged)
	s := createTestDiffState(20)
	s.ViewportStart = 10
	s.CurrentBlockIdx = 0

	// Call positionAtFirstChange
	s.positionAtFirstChange()

	// ViewportStart should be reset to 0
	if s.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0 when no changes, got %d", s.ViewportStart)
	}

	// CurrentBlockIdx should be -1 (no change block found)
	if s.CurrentBlockIdx != -1 {
		t.Errorf("Expected CurrentBlockIdx to be -1 when no changes, got %d", s.CurrentBlockIdx)
	}
}

func TestPositionAtFirstChange_NilDiff(t *testing.T) {
	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source:          createTestDiffSource(commit),
		File:            core.FileChange{Path: "file.go"},
		Diff:            nil,
		ViewportStart:   10,
		CurrentBlockIdx: 5,
	}

	// Call positionAtFirstChange
	s.positionAtFirstChange()

	// ViewportStart should be reset to 0
	if s.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0 for nil diff, got %d", s.ViewportStart)
	}

	// CurrentBlockIdx should be -1
	if s.CurrentBlockIdx != -1 {
		t.Errorf("Expected CurrentBlockIdx to be -1 for nil diff, got %d", s.CurrentBlockIdx)
	}
}

// TestNavigateToNextChange_SmallContent tests that navigating to the next change
// works correctly when all content fits within the viewport.
// This is a regression test for the bug where viewport position was incorrectly
// clamped to 0 when calculateMaxViewportStart returns 0 (content < viewport height).
func TestNavigateToNextChange_SmallContent(t *testing.T) {
	// Create a diff with two change blocks separated by unchanged lines.
	// Structure: 5 unchanged, 1 change, 7 unchanged, 1 change, 5 unchanged
	// Total: 19 lines - all fit within a viewport of height 22 (available height = 20)
	//
	// Block structure:
	// - Block 0: Unchanged (lines 0-4), 5 lines
	// - Block 1: Change (line 5), 1 line - first change at position 5
	// - Block 2: Unchanged (lines 6-12), 7 lines
	// - Block 3: Change (line 13), 1 line - second change at position 13
	// - Block 4: Unchanged (lines 14-18), 5 lines

	var blocks []diff.Block

	// Block 0: 5 unchanged lines (positions 0-4)
	unchanged1 := make([]diff.LinePair, 5)
	for i := 0; i < 5; i++ {
		unchanged1[i] = diff.LinePair{
			LeftLineNo:  i + 1,
			RightLineNo: i + 1,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}},
		}
	}
	blocks = append(blocks, diff.UnchangedBlock{Lines: unchanged1})

	// Block 1: First change block at position 5
	blocks = append(blocks, diff.ChangeBlock{Lines: []diff.ChangeLine{
		diff.AddedLine{RightLineNo: 6, Tokens: []highlight.Token{{Type: chroma.Text, Value: "first added line"}}},
	}})

	// Block 2: 7 unchanged lines (positions 6-12)
	unchanged2 := make([]diff.LinePair, 7)
	for i := 0; i < 7; i++ {
		unchanged2[i] = diff.LinePair{
			LeftLineNo:  7 + i,
			RightLineNo: 7 + i,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}},
		}
	}
	blocks = append(blocks, diff.UnchangedBlock{Lines: unchanged2})

	// Block 3: Second change block at position 13
	blocks = append(blocks, diff.ChangeBlock{Lines: []diff.ChangeLine{
		diff.AddedLine{RightLineNo: 14, Tokens: []highlight.Token{{Type: chroma.Text, Value: "second added line"}}},
	}})

	// Block 4: 5 unchanged lines (positions 14-18)
	unchanged3 := make([]diff.LinePair, 5)
	for i := 0; i < 5; i++ {
		unchanged3[i] = diff.LinePair{
			LeftLineNo:  15 + i,
			RightLineNo: 15 + i,
			Tokens:      []highlight.Token{{Type: chroma.Text, Value: "unchanged line"}},
		}
	}
	blocks = append(blocks, diff.UnchangedBlock{Lines: unchanged3})

	commit := core.GitCommit{Hash: "abc123"}
	s := &State{
		Source:          createTestDiffSource(commit),
		File:            core.FileChange{Path: "file.go", Additions: 2, Deletions: 0},
		Diff:            &diff.FileDiff{Path: "file.go", Blocks: blocks},
		ViewportStart:   5, // Start at first change block
		CurrentBlockIdx: 1, // Index of first change block
	}

	// Use a viewport height that can fit all 19 lines of content
	// Height 22 - 2 header lines = 20 available lines, which fits all 19 lines
	// This means calculateMaxViewportStart will return 0 (since 19 - 20 < 0)
	ctx := testutils.MockContext{W: 80, H: 22}

	// Press "n" to navigate to next change
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	// BUG: ViewportStart should move to 13 (second change block position)
	// but due to the clamping bug, it gets clamped to 0 because
	// calculateMaxViewportStart returns 0 when content fits in viewport.
	//
	// Expected: ViewportStart = 13 (position of second change block)
	// Actual (bug): ViewportStart = 0 (incorrectly clamped)
	if diffState.ViewportStart != 13 {
		t.Errorf("Expected ViewportStart to be 13 (second change block), got %d", diffState.ViewportStart)
	}

	// Also verify CurrentBlockIdx was updated to the second change block (index 3)
	if diffState.CurrentBlockIdx != 3 {
		t.Errorf("Expected CurrentBlockIdx to be 3 (second change block), got %d", diffState.CurrentBlockIdx)
	}
}
