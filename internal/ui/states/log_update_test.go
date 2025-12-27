package states

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/messages"
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
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "j" to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
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
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "k" to move up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
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
				Preview:       PreviewNone{},
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
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
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
		Preview:       PreviewNone{},
	}
	ctx := mockContext{width: 80, height: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
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
		Preview:       PreviewNone{},
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
		Preview:       PreviewNone{},
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

func TestLogState_Update_NavigationTriggersPreviewLoading(t *testing.T) {
	commits := createTestCommits(5)

	tests := []struct {
		name           string
		initialCursor  int
		key            string
		expectedCursor int
		shouldLoadCmd  bool
	}{
		{"j triggers load", 0, "j", 1, true},
		{"k triggers load", 2, "k", 1, true},
		{"down triggers load", 0, "down", 1, true},
		{"up triggers load", 2, "up", 1, true},
		{"g triggers load", 3, "g", 0, true},
		{"G triggers load", 0, "G", 4, true},
		{"j at end doesn't trigger load", 4, "j", 4, false},
		{"k at start doesn't trigger load", 0, "k", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := LogState{
				Commits:       commits,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
				Preview:       PreviewNone{},
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

			newState, cmd := s.Update(msg, ctx)
			listState := newState.(LogState)

			// Check cursor moved to expected position
			if listState.Cursor != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, listState.Cursor)
			}

			// Check if command was returned
			if tt.shouldLoadCmd && cmd == nil {
				t.Error("Expected loadPreview command, got nil")
			} else if !tt.shouldLoadCmd && cmd != nil {
				t.Error("Expected nil command, got a command")
			}

			// If cursor moved, check Preview state is PreviewLoading
			if tt.shouldLoadCmd {
				previewLoading, ok := listState.Preview.(PreviewLoading)
				if !ok {
					t.Errorf("Expected Preview to be PreviewLoading, got %T", listState.Preview)
				} else if previewLoading.ForHash != commits[tt.expectedCursor].Hash {
					t.Errorf("Expected ForHash %s, got %s", commits[tt.expectedCursor].Hash, previewLoading.ForHash)
				}
			}
		})
	}
}

func TestLogState_Update_FilesPreviewLoadedMsg_Success(t *testing.T) {
	commits := createTestCommits(3)
	fileChanges := []git.FileChange{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.go", Status: "A"},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        1,
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[1].Hash},
	}
	ctx := mockContext{width: 80, height: 24}

	msg := messages.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   fileChanges,
		Err:     nil,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)

	// Check Preview is now PreviewLoaded
	previewLoaded, ok := listState.Preview.(PreviewLoaded)
	if !ok {
		t.Fatalf("Expected Preview to be PreviewLoaded, got %T", listState.Preview)
	}

	if previewLoaded.ForHash != commits[1].Hash {
		t.Errorf("Expected ForHash %s, got %s", commits[1].Hash, previewLoaded.ForHash)
	}

	if len(previewLoaded.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(previewLoaded.Files))
	}
}

func TestLogState_Update_FilesPreviewLoadedMsg_Error(t *testing.T) {
	commits := createTestCommits(3)
	testErr := errors.New("failed to load files")

	s := LogState{
		Commits:       commits,
		Cursor:        1,
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[1].Hash},
	}
	ctx := mockContext{width: 80, height: 24}

	msg := messages.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   nil,
		Err:     testErr,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)

	// Check Preview is now PreviewError
	previewError, ok := listState.Preview.(PreviewError)
	if !ok {
		t.Fatalf("Expected Preview to be PreviewError, got %T", listState.Preview)
	}

	if previewError.ForHash != commits[1].Hash {
		t.Errorf("Expected ForHash %s, got %s", commits[1].Hash, previewError.ForHash)
	}

	if previewError.Err != testErr {
		t.Errorf("Expected error %v, got %v", testErr, previewError.Err)
	}
}

func TestLogState_Update_FilesPreviewLoadedMsg_StaleResponse(t *testing.T) {
	commits := createTestCommits(3)
	fileChanges := []git.FileChange{
		{Path: "file1.go", Status: "M"},
	}

	s := LogState{
		Commits:       commits,
		Cursor:        2, // User navigated to commit 2
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[2].Hash},
	}
	ctx := mockContext{width: 80, height: 24}

	// Stale response for commit 1 arrives
	msg := messages.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   fileChanges,
		Err:     nil,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(LogState)

	// Preview should remain as PreviewLoading for commit 2
	previewLoading, ok := listState.Preview.(PreviewLoading)
	if !ok {
		t.Fatalf("Expected Preview to remain PreviewLoading, got %T", listState.Preview)
	}

	if previewLoading.ForHash != commits[2].Hash {
		t.Errorf("Expected ForHash %s, got %s", commits[2].Hash, previewLoading.ForHash)
	}
}
