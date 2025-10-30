package states

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

func TestFilesState_Update_NavigationDown(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "j" to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 1 {
		t.Errorf("Expected cursor to move to 1, got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_NavigationUp(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        5,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "k" to move up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 4 {
		t.Errorf("Expected cursor to move to 4, got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_NavigationJumpToTop(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        5,
		ViewportStart: 3,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 0 {
		t.Errorf("Expected cursor to jump to 0, got %d", filesState.Cursor)
	}

	if filesState.ViewportStart != 0 {
		t.Errorf("Expected viewport to reset to 0, got %d", filesState.ViewportStart)
	}
}

func TestFilesState_Update_NavigationJumpToBottom(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 9 {
		t.Errorf("Expected cursor to jump to last file (9), got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_NavigationBoundaries(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)

	tests := []struct {
		name           string
		initialCursor  int
		key            rune
		expectedCursor int
	}{
		{"up at top stays at top", 0, 'k', 0},
		{"down at bottom stays at bottom", 4, 'j', 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FilesState{
				Commit:        commit,
				Files:         files,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
			}
			ctx := mockContext{width: 80, height: 24}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			newState, _ := s.Update(msg, ctx)

			filesState, ok := newState.(*FilesState)
			if !ok {
				t.Fatal("Expected state to remain FilesState")
			}

			if filesState.Cursor != tt.expectedCursor {
				t.Errorf("Expected cursor to be %d, got %d", tt.expectedCursor, filesState.Cursor)
			}
		})
	}
}

func TestFilesState_Update_ArrowKeyNavigation(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)

	tests := []struct {
		name           string
		initialCursor  int
		keyType        tea.KeyType
		expectedCursor int
	}{
		{"down arrow moves down", 0, tea.KeyDown, 1},
		{"up arrow moves up", 5, tea.KeyUp, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := FilesState{
				Commit:        commit,
				Files:         files,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
			}
			ctx := mockContext{width: 80, height: 24}

			msg := tea.KeyMsg{Type: tt.keyType}
			newState, _ := s.Update(msg, ctx)

			filesState, ok := newState.(*FilesState)
			if !ok {
				t.Fatal("Expected state to remain FilesState")
			}

			if filesState.Cursor != tt.expectedCursor {
				t.Errorf("Expected cursor to be %d, got %d", tt.expectedCursor, filesState.Cursor)
			}
		})
	}
}

func TestFilesState_Update_ViewportScrolling(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(20)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        5,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 10}

	// Move cursor down multiple times to trigger viewport scrolling
	state := &s
	for range 10 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newState, _ := state.Update(msg, ctx)
		state = newState.(*FilesState)
	}

	// After moving down 10 times from position 5, cursor should be at 15
	if state.Cursor != 15 {
		t.Errorf("Expected cursor to be at 15, got %d", state.Cursor)
	}

	// Viewport should have scrolled to keep cursor visible
	if state.ViewportStart == 0 {
		t.Error("Expected viewport to have scrolled down")
	}
}

func TestFilesState_Update_BackNavigation(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	listCommits := []git.GitCommit{commit}
	s := FilesState{
		Commit:            commit,
		Files:             files,
		Cursor:            2,
		ViewportStart:     0,
		ListCommits:       listCommits,
		ListCursor:        3,
		ListViewportStart: 1,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "q" to go back
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newState, cmd := s.Update(msg, ctx)

	// Should transition to ListState
	listState, ok := newState.(*ListState)
	if !ok {
		t.Fatalf("Expected to transition to ListState, got %T", newState)
	}

	// Should not return a command (direct state transition)
	if cmd != nil {
		t.Error("Expected no command for direct state transition")
	}

	// Verify list state was restored correctly
	if listState.Cursor != 3 {
		t.Errorf("Expected list cursor to be restored to 3, got %d", listState.Cursor)
	}
	if listState.ViewportStart != 1 {
		t.Errorf("Expected list viewport to be restored to 1, got %d", listState.ViewportStart)
	}
	if len(listState.Commits) != len(listCommits) {
		t.Errorf("Expected commits to be restored, got %d commits", len(listState.Commits))
	}
}

func TestFilesState_Update_SingleFileList(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(1)
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Try to move down (should stay at 0)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 with single file, got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_EmptyFileList(t *testing.T) {
	commit := createTestCommit()
	files := []git.FileChange{}
	s := FilesState{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Try to move down (should stay at 0)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*FilesState)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 with empty files, got %d", filesState.Cursor)
	}
}
