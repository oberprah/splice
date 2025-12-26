package states

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

func createTestCommits(count int) []git.GitCommit {
	commits := make([]git.GitCommit, count)
	for i := range count {
		commits[i] = git.GitCommit{
			Hash:    string(rune('a' + i)),
			Message: "Commit " + string(rune('0'+i)),
			Body:    "",
			Author:  "Author",
			Date:    time.Now(),
		}
	}
	return commits
}

func TestLogState_Update_NavigationDown(t *testing.T) {
	commits := createTestCommits(10)
	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "j" to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)
	if listState.Cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", listState.Cursor)
	}
}

func TestLogState_Update_NavigationUp(t *testing.T) {
	commits := createTestCommits(10)
	s := LogState{
		Commits:       commits,
		Cursor:        5,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "k" to move up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)
	if listState.Cursor != 4 {
		t.Errorf("Expected cursor at 4, got %d", listState.Cursor)
	}
}

func TestLogState_Update_CursorBoundaries(t *testing.T) {
	commits := createTestCommits(5)

	tests := []struct {
		name           string
		initialCursor  int
		key            string
		expectedCursor int
	}{
		{"can't go below 0", 0, "k", 0},
		{"can't go above last", 4, "j", 4},
		{"down arrow at start", 0, "down", 1},
		{"up arrow at end", 4, "up", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := LogState{
				Commits:       commits,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
			}
			ctx := mockContext{width: 80, height: 24}

			var msg tea.Msg
			switch tt.key {
			case "down":
				msg = tea.KeyMsg{Type: tea.KeyDown}
			case "up":
				msg = tea.KeyMsg{Type: tea.KeyUp}
			default:
				msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			newState, _ := s.Update(msg, ctx)
			listState := newState.(LogState)

			if listState.Cursor != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, listState.Cursor)
			}
		})
	}
}

func TestLogState_Update_JumpToTop(t *testing.T) {
	commits := createTestCommits(10)
	s := LogState{
		Commits:       commits,
		Cursor:        5,
		ViewportStart: 3,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)
	if listState.Cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", listState.Cursor)
	}
	if listState.ViewportStart != 0 {
		t.Errorf("Expected viewportStart at 0, got %d", listState.ViewportStart)
	}
}

func TestLogState_Update_JumpToBottom(t *testing.T) {
	commits := createTestCommits(10)
	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)
	if listState.Cursor != 9 {
		t.Errorf("Expected cursor at 9, got %d", listState.Cursor)
	}
}

func TestLogState_Update_ViewportScrolling(t *testing.T) {
	commits := createTestCommits(30)
	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 10}

	// Move cursor down beyond viewport
	for range 15 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
		newState, _ := s.Update(msg, ctx)
		s = newState.(LogState)
	}

	// Cursor should be at 15
	if s.Cursor != 15 {
		t.Errorf("Expected cursor at 15, got %d", s.Cursor)
	}

	// Viewport should have scrolled to keep cursor visible
	// With height=10, viewport should start around cursor - height + 1
	if s.ViewportStart < 6 {
		t.Errorf("Expected viewport to have scrolled, viewportStart=%d", s.ViewportStart)
	}
}

func TestLogState_Update_QuitKeys(t *testing.T) {
	commits := createTestCommits(5)
	s := LogState{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := mockContext{width: 80, height: 24}

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"q key", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}},
		{"ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := s.Update(tt.msg, ctx)

			if cmd == nil {
				t.Error("Expected tea.Quit command, got nil")
			}
		})
	}
}
