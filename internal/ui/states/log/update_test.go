package log

import (
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/testutils"
)

func createTestCommits(count int) []core.GitCommit {
	commits := make([]core.GitCommit, count)
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := range count {
		// Create linear parent chain (each commit's parent is the next one)
		var parents []string
		if i < count-1 {
			parents = []string{string(rune('a' + i + 1))}
		} else {
			parents = []string{} // Last commit is root
		}

		commits[i] = core.GitCommit{
			Hash:         string(rune('a' + i)),
			ParentHashes: parents,
			Message:      "Commit " + string(rune('0'+i)),
			Body:         "",
			Author:       "Author",
			Date:         baseTime.Add(time.Duration(-i) * time.Hour),
		}
	}
	return commits
}

func TestLogState_Update_NavigationDown(t *testing.T) {
	commits := createTestCommits(10)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "j" to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
	}

	listState := newState.(State)
	if listState.CursorPosition() != 1 {
		t.Errorf("Expected cursor at 1, got %d", listState.CursorPosition())
	}
}

func TestLogState_Update_NavigationUp(t *testing.T) {
	commits := createTestCommits(10)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 5},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "k" to move up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
	}

	listState := newState.(State)
	if listState.CursorPosition() != 4 {
		t.Errorf("Expected cursor at 4, got %d", listState.CursorPosition())
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
			s := State{
				Commits:       commits,
				Cursor:        core.CursorNormal{Pos: tt.initialCursor},
				ViewportStart: 0,
				Preview:       PreviewNone{},
			}
			ctx := testutils.MockContext{W: 80, H: 24}

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
			listState := newState.(State)

			if listState.CursorPosition() != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, listState.CursorPosition())
			}
		})
	}
}

func TestLogState_Update_JumpToTop(t *testing.T) {
	commits := createTestCommits(10)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 5},
		ViewportStart: 3,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
	}

	listState := newState.(State)
	if listState.CursorPosition() != 0 {
		t.Errorf("Expected cursor at 0, got %d", listState.CursorPosition())
	}
	if listState.ViewportStart != 0 {
		t.Errorf("Expected viewportStart at 0, got %d", listState.ViewportStart)
	}
}

func TestLogState_Update_JumpToBottom(t *testing.T) {
	commits := createTestCommits(10)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Error("Expected loadPreview command")
	}

	listState := newState.(State)
	if listState.CursorPosition() != 9 {
		t.Errorf("Expected cursor at 9, got %d", listState.CursorPosition())
	}
}

func TestLogState_Update_ViewportScrolling(t *testing.T) {
	commits := createTestCommits(30)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 10}

	// Move cursor down beyond viewport
	for range 15 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
		newState, _ := s.Update(msg, ctx)
		s = newState.(State)
	}

	// Cursor should be at 15
	if s.CursorPosition() != 15 {
		t.Errorf("Expected cursor at 15, got %d", s.CursorPosition())
	}

	// Viewport should have scrolled to keep cursor visible
	// With height=10, viewport should start around cursor - height + 1
	if s.ViewportStart < 6 {
		t.Errorf("Expected viewport to have scrolled, viewportStart=%d", s.ViewportStart)
	}
}

func TestLogState_Update_QuitKeys(t *testing.T) {
	commits := createTestCommits(5)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

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
			s := State{
				Commits:       commits,
				Cursor:        core.CursorNormal{Pos: tt.initialCursor},
				ViewportStart: 0,
				Preview:       PreviewNone{},
			}
			ctx := testutils.MockContext{W: 80, H: 24}

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
			listState := newState.(State)

			// Check cursor moved to expected position
			if listState.CursorPosition() != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, listState.CursorPosition())
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
	fileChanges := []core.FileChange{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.go", Status: "A"},
	}

	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 1},
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[1].Hash},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   fileChanges,
		Err:     nil,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(State)

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

	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 1},
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[1].Hash},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   nil,
		Err:     testErr,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(State)

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
	fileChanges := []core.FileChange{
		{Path: "file1.go", Status: "M"},
	}

	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 2}, // User navigated to commit 2
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[2].Hash},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Stale response for commit 1 arrives
	msg := core.FilesPreviewLoadedMsg{
		ForHash: commits[1].Hash,
		Files:   fileChanges,
		Err:     nil,
	}

	newState, cmd := s.Update(msg, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	listState := newState.(State)

	// Preview should remain as PreviewLoading for commit 2
	previewLoading, ok := listState.Preview.(PreviewLoading)
	if !ok {
		t.Fatalf("Expected Preview to remain PreviewLoading, got %T", listState.Preview)
	}

	if previewLoading.ForHash != commits[2].Hash {
		t.Errorf("Expected ForHash %s, got %s", commits[2].Hash, previewLoading.ForHash)
	}
}

func TestLogState_Update_VisualMode_NavigationLoadsRangePreview(t *testing.T) {
	commits := createTestCommits(5)

	// Start in visual mode with anchor at position 1
	s := State{
		Commits:       commits,
		Cursor:        core.CursorVisual{Pos: 1, Anchor: 1},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}

	// Mock FetchFileChanges to track what was requested
	var rangeCalled core.CommitRange
	mockFetchFileChanges := func(commitRange core.CommitRange) ([]core.FileChange, error) {
		rangeCalled = commitRange
		return []core.FileChange{
			{Path: "file1.go", Status: "M"},
			{Path: "file2.go", Status: "A"},
			{Path: "file3.go", Status: "M"},
		}, nil
	}

	ctx := testutils.MockContext{
		W:                    80,
		H:                    24,
		MockFetchFileChanges: mockFetchFileChanges,
	}

	// Press "j" to move cursor down to position 2 (selecting commits 1-2)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	newState, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Fatal("Expected loadPreview command, got nil")
	}

	listState := newState.(State)

	// Verify cursor moved to position 2
	if listState.CursorPosition() != 2 {
		t.Errorf("Expected cursor at 2, got %d", listState.CursorPosition())
	}

	// Verify Preview state is PreviewLoading for the range
	previewLoading, ok := listState.Preview.(PreviewLoading)
	if !ok {
		t.Fatalf("Expected Preview to be PreviewLoading, got %T", listState.Preview)
	}

	// In visual mode selecting commits 1-2 (where 1 is older, 2 is newer in log order):
	// - The range should be from commit[2]^ to commit[1]
	// - ForHash should represent this range (e.g., "c..b" for commits c and b)
	expectedRangeHash := commits[2].Hash + ".." + commits[1].Hash
	if previewLoading.ForHash != expectedRangeHash {
		t.Errorf("Expected ForHash %s, got %s", expectedRangeHash, previewLoading.ForHash)
	}

	// Execute the command to trigger the fetch
	resultMsg := cmd()

	// Verify the fetch was called with correct range
	if rangeCalled.Start.Hash != commits[2].Hash {
		t.Errorf("Expected range Start hash %s, got %s", commits[2].Hash, rangeCalled.Start.Hash)
	}
	if rangeCalled.End.Hash != commits[1].Hash {
		t.Errorf("Expected range End hash %s, got %s", commits[1].Hash, rangeCalled.End.Hash)
	}

	// Verify the message contains the combined files
	previewMsg, ok := resultMsg.(core.FilesPreviewLoadedMsg)
	if !ok {
		t.Fatalf("Expected FilesPreviewLoadedMsg, got %T", resultMsg)
	}

	if len(previewMsg.Files) != 3 {
		t.Errorf("Expected 3 combined files, got %d", len(previewMsg.Files))
	}

	if previewMsg.ForHash != expectedRangeHash {
		t.Errorf("Expected message ForHash %s, got %s", expectedRangeHash, previewMsg.ForHash)
	}
}
