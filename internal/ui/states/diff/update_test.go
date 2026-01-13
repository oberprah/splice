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

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
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

	commit := core.GitCommit{Hash: "abc123"}
	return &State{
		Source: createTestDiffSource(commit),
		File:   core.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
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

func TestGetCurrentFileLineNumber_UnchangedAlignment(t *testing.T) {
	s := createTestDiffState(10)
	s.ViewportStart = 5

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// UnchangedAlignment at index 5 has RightIdx=5, so lineNo should be 6 (1-indexed)
	expected := 6
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_ModifiedAlignment(t *testing.T) {
	// Create a diff with a modified line at index 3
	leftLines := make([]diff.AlignedLine, 5)
	rightLines := make([]diff.AlignedLine, 5)
	for i := 0; i < 5; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	alignments := []diff.Alignment{
		diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
		diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1},
		diff.UnchangedAlignment{LeftIdx: 2, RightIdx: 2},
		diff.ModifiedAlignment{LeftIdx: 3, RightIdx: 3},
		diff.UnchangedAlignment{LeftIdx: 4, RightIdx: 4},
	}

	s := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left:       diff.FileContent{Path: "file.go", Lines: leftLines},
			Right:      diff.FileContent{Path: "file.go", Lines: rightLines},
			Alignments: alignments,
		},
		ViewportStart: 3,
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// ModifiedAlignment at index 3 has RightIdx=3, so lineNo should be 4 (1-indexed)
	expected := 4
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_AddedAlignment(t *testing.T) {
	// Create a diff with an added line at index 2
	leftLines := make([]diff.AlignedLine, 4)
	rightLines := make([]diff.AlignedLine, 5)
	for i := 0; i < 4; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}
	for i := 0; i < 5; i++ {
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	alignments := []diff.Alignment{
		diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
		diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1},
		diff.AddedAlignment{RightIdx: 2},
		diff.UnchangedAlignment{LeftIdx: 2, RightIdx: 3},
		diff.UnchangedAlignment{LeftIdx: 3, RightIdx: 4},
	}

	s := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left:       diff.FileContent{Path: "file.go", Lines: leftLines},
			Right:      diff.FileContent{Path: "file.go", Lines: rightLines},
			Alignments: alignments,
		},
		ViewportStart: 2,
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// AddedAlignment at index 2 has RightIdx=2, so lineNo should be 3 (1-indexed)
	expected := 3
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_RemovedAlignment(t *testing.T) {
	// Create a diff with a removed line at index 2, followed by unchanged lines
	leftLines := make([]diff.AlignedLine, 5)
	rightLines := make([]diff.AlignedLine, 4)
	for i := 0; i < 5; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}
	for i := 0; i < 4; i++ {
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	alignments := []diff.Alignment{
		diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
		diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1},
		diff.RemovedAlignment{LeftIdx: 2}, // Removed line - no RightIdx
		diff.UnchangedAlignment{LeftIdx: 3, RightIdx: 2},
		diff.UnchangedAlignment{LeftIdx: 4, RightIdx: 3},
	}

	s := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left:       diff.FileContent{Path: "file.go", Lines: leftLines},
			Right:      diff.FileContent{Path: "file.go", Lines: rightLines},
			Alignments: alignments,
		},
		ViewportStart: 2,
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// RemovedAlignment at index 2 - should search forward and find next alignment with RightIdx
	// Next alignment is at index 3 with RightIdx=2, so lineNo should be 3 (1-indexed)
	expected := 3
	if lineNo != expected {
		t.Errorf("Expected line number %d, got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_RemovedAlignment_NoFollowingRightIdx(t *testing.T) {
	// Create a diff where all remaining alignments after viewport are removed
	leftLines := make([]diff.AlignedLine, 5)
	rightLines := make([]diff.AlignedLine, 2)
	for i := 0; i < 5; i++ {
		leftLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}
	for i := 0; i < 2; i++ {
		rightLines[i] = diff.AlignedLine{
			Tokens: []highlight.Token{{Type: chroma.Text, Value: "test line"}},
		}
	}

	alignments := []diff.Alignment{
		diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
		diff.UnchangedAlignment{LeftIdx: 1, RightIdx: 1},
		diff.RemovedAlignment{LeftIdx: 2},
		diff.RemovedAlignment{LeftIdx: 3},
		diff.RemovedAlignment{LeftIdx: 4},
	}

	s := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left:       diff.FileContent{Path: "file.go", Lines: leftLines},
			Right:      diff.FileContent{Path: "file.go", Lines: rightLines},
			Alignments: alignments,
		},
		ViewportStart: 2,
	}

	lineNo, err := s.getCurrentFileLineNumber()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// No following alignments with RightIdx, should fall back to line 1
	expected := 1
	if lineNo != expected {
		t.Errorf("Expected line number %d (fallback), got %d", expected, lineNo)
	}
}

func TestGetCurrentFileLineNumber_NilDiff(t *testing.T) {
	s := &State{
		CommitRange:   core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
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

func TestGetCurrentFileLineNumber_EmptyAlignments(t *testing.T) {
	s := &State{
		CommitRange: core.NewSingleCommitRange(core.GitCommit{Hash: "abc123"}),
		File:        core.FileChange{Path: "file.go"},
		Diff: &diff.AlignedFileDiff{
			Left:       diff.FileContent{Path: "file.go", Lines: []diff.AlignedLine{}},
			Right:      diff.FileContent{Path: "file.go", Lines: []diff.AlignedLine{}},
			Alignments: []diff.Alignment{},
		},
		ViewportStart: 0,
	}

	_, err := s.getCurrentFileLineNumber()
	if err == nil {
		t.Fatal("Expected error for empty alignments, got nil")
	}

	expectedMsg := "diff has no alignments"
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
