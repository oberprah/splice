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
		Range: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:  core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
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
		Range: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:  core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
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
