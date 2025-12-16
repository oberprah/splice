package states

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
)

func createTestDiffState(numLines int) *DiffState {
	lines := make([]diff.Line, numLines)
	for i := 0; i < numLines; i++ {
		lines[i] = diff.Line{
			Type:      diff.Context,
			Content:   "test line",
			OldLineNo: i + 1,
			NewLineNo: i + 1,
		}
	}

	return &DiffState{
		Commit: git.GitCommit{Hash: "abc123"},
		File:   git.FileChange{Path: "file.go", Additions: 5, Deletions: 3},
		Diff: diff.FileDiff{
			OldPath: "file.go",
			NewPath: "file.go",
			Lines:   lines,
		},
		ViewportStart:          0,
		FilesCommit:            git.GitCommit{Hash: "abc123"},
		FilesFiles:             []git.FileChange{{Path: "file.go"}},
		FilesCursor:            0,
		FilesViewportStart:     0,
		FilesListCommits:       []git.GitCommit{{Hash: "abc123"}},
		FilesListCursor:        0,
		FilesListViewportStart: 0,
	}
}

func TestDiffState_Update_ScrollDown(t *testing.T) {
	s := createTestDiffState(50)
	ctx := &mockContext{width: 80, height: 20}

	// Press "j" to scroll down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
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
	ctx := &mockContext{width: 80, height: 20}

	// Press "k" to scroll up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
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
	ctx := &mockContext{width: 80, height: 20}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0, got %d", diffState.ViewportStart)
	}
}

func TestDiffState_Update_JumpToBottom(t *testing.T) {
	s := createTestDiffState(50)
	ctx := &mockContext{width: 80, height: 20}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
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
	ctx := &mockContext{width: 80, height: 20}

	// Press "k" at top boundary
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
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
	ctx := &mockContext{width: 80, height: 20}

	// Press "j" at bottom boundary
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
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
	s.FilesCursor = 3
	s.FilesViewportStart = 1
	ctx := &mockContext{width: 80, height: 20}

	// Press "q" to go back
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newState, cmd := s.Update(msg, ctx)

	// Should transition to FilesState
	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatalf("Expected to transition to FilesState, got %T", newState)
	}

	// Should not return a command
	if cmd != nil {
		t.Error("Expected no command for back navigation")
	}

	// Verify restored cursor position
	if filesState.Cursor != 3 {
		t.Errorf("Expected cursor to be restored to 3, got %d", filesState.Cursor)
	}

	if filesState.ViewportStart != 1 {
		t.Errorf("Expected viewport to be restored to 1, got %d", filesState.ViewportStart)
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
			ctx := &mockContext{width: 80, height: 20}

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
		name          string
		keyType       tea.KeyType
		initialVP     int
		expectedVP    int
	}{
		{"down arrow scrolls down", tea.KeyDown, 0, 1},
		{"up arrow scrolls up", tea.KeyUp, 10, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := createTestDiffState(50)
			s.ViewportStart = tt.initialVP
			ctx := &mockContext{width: 80, height: 20}

			msg := tea.KeyMsg{Type: tt.keyType}
			newState, _ := s.Update(msg, ctx)

			diffState, ok := newState.(*DiffState)
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
	ctx := &mockContext{width: 80, height: 20}

	// Try to scroll in empty diff
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	diffState, ok := newState.(*DiffState)
	if !ok {
		t.Fatal("Expected state to remain DiffState")
	}

	if diffState.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to stay at 0 for empty diff, got %d", diffState.ViewportStart)
	}
}
