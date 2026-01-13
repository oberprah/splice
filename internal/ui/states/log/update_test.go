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

func TestLogState_Update_QExitsVisualMode(t *testing.T) {
	commits := createTestCommits(5)
	// Start in visual mode with selection from pos 0 to anchor 2
	s := State{
		Commits:       commits,
		Cursor:        core.CursorVisual{Pos: 2, Anchor: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "q" to exit visual mode (not quit)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	newState, cmd := s.Update(msg, ctx)

	logState, ok := newState.(State)
	if !ok {
		t.Fatalf("Expected State, got %T", newState)
	}

	// Should be in normal mode at position 2
	normal, ok := logState.Cursor.(core.CursorNormal)
	if !ok {
		t.Fatalf("Expected CursorNormal after q in visual mode, got %T", logState.Cursor)
	}
	if normal.Pos != 2 {
		t.Errorf("Expected cursor at position 2, got %d", normal.Pos)
	}

	// Should return a command to load preview (not tea.Quit)
	if cmd == nil {
		t.Error("Expected preview load command, got nil")
	}
}

func TestLogState_Update_QQuitsInNormalMode(t *testing.T) {
	commits := createTestCommits(5)
	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "q" in normal mode to quit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	_, cmd := s.Update(msg, ctx)

	// Should return tea.Quit command
	if cmd == nil {
		t.Fatal("Expected tea.Quit command, got nil")
	}

	// Execute the command and check if it's a quit message
	resultMsg := cmd()
	if resultMsg != tea.Quit() {
		t.Errorf("Expected tea.Quit message, got %T", resultMsg)
	}
}

func TestLogState_Update_EnterCreatesCommitRangeDiffSource(t *testing.T) {
	commits := createTestCommits(5)

	tests := []struct {
		name          string
		cursor        core.CursorState
		expectedStart string
		expectedEnd   string
		expectedCount int
		description   string
	}{
		{
			name:          "single commit in normal mode",
			cursor:        core.CursorNormal{Pos: 1},
			expectedStart: commits[1].Hash,
			expectedEnd:   commits[1].Hash,
			expectedCount: 1,
			description:   "Single commit should have Start == End with Count = 1",
		},
		{
			name:          "range in visual mode (anchor < pos)",
			cursor:        core.CursorVisual{Pos: 2, Anchor: 1},
			expectedStart: commits[2].Hash,
			expectedEnd:   commits[1].Hash,
			expectedCount: 2,
			description:   "Range with anchor < pos should have Start = newer (pos), End = older (anchor)",
		},
		{
			name:          "range in visual mode (pos < anchor)",
			cursor:        core.CursorVisual{Pos: 1, Anchor: 3},
			expectedStart: commits[3].Hash,
			expectedEnd:   commits[1].Hash,
			expectedCount: 3,
			description:   "Range with pos < anchor should have Start = newer (pos), End = older (anchor)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{
				Commits:       commits,
				Cursor:        tt.cursor,
				ViewportStart: 0,
				Preview:       PreviewNone{},
			}

			mockFileChanges := []core.FileChange{
				{Path: "file1.go", Status: "M"},
			}
			mockFetchFileChanges := func(commitRange core.CommitRange) ([]core.FileChange, error) {
				return mockFileChanges, nil
			}

			ctx := testutils.MockContext{
				W:                    80,
				H:                    24,
				MockFetchFileChanges: mockFetchFileChanges,
			}

			// Press Enter
			msg := tea.KeyMsg{Type: tea.KeyEnter}
			_, cmd := s.Update(msg, ctx)

			if cmd == nil {
				t.Fatal("Expected command, got nil")
			}

			// Execute the command to get FilesLoadedMsg
			resultMsg := cmd()
			filesLoadedMsg, ok := resultMsg.(core.FilesLoadedMsg)
			if !ok {
				t.Fatalf("Expected FilesLoadedMsg, got %T", resultMsg)
			}

			// Verify the DiffSource is a CommitRangeDiffSource
			commitRangeSource, ok := filesLoadedMsg.Source.(core.CommitRangeDiffSource)
			if !ok {
				t.Fatalf("Expected CommitRangeDiffSource, got %T", filesLoadedMsg.Source)
			}

			// Verify the CommitRangeDiffSource has correct values
			if commitRangeSource.Start.Hash != tt.expectedStart {
				t.Errorf("Expected Start.Hash %s, got %s (%s)", tt.expectedStart, commitRangeSource.Start.Hash, tt.description)
			}

			if commitRangeSource.End.Hash != tt.expectedEnd {
				t.Errorf("Expected End.Hash %s, got %s (%s)", tt.expectedEnd, commitRangeSource.End.Hash, tt.description)
			}

			if commitRangeSource.Count != tt.expectedCount {
				t.Errorf("Expected Count %d, got %d (%s)", tt.expectedCount, commitRangeSource.Count, tt.description)
			}

			// Verify files are included
			if len(filesLoadedMsg.Files) != len(mockFileChanges) {
				t.Errorf("Expected %d files, got %d", len(mockFileChanges), len(filesLoadedMsg.Files))
			}

			// Verify no error
			if filesLoadedMsg.Err != nil {
				t.Errorf("Expected no error, got %v", filesLoadedMsg.Err)
			}
		})
	}
}

func TestLogState_Update_FilesLoadedMsgCreatesPushFilesScreenMsg(t *testing.T) {
	commits := createTestCommits(3)
	fileChanges := []core.FileChange{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.go", Status: "A"},
	}

	// Create a CommitRangeDiffSource
	commitRange := core.NewSingleCommitRange(commits[1])
	diffSource := commitRange.ToDiffSource()

	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 1},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesLoadedMsg{
		Source: diffSource,
		Files:  fileChanges,
		Err:    nil,
	}

	_, cmd := s.Update(msg, ctx)

	if cmd == nil {
		t.Fatal("Expected command, got nil")
	}

	// Execute the command to get PushFilesScreenMsg
	resultMsg := cmd()
	pushFilesMsg, ok := resultMsg.(core.PushFilesScreenMsg)
	if !ok {
		t.Fatalf("Expected PushFilesScreenMsg, got %T", resultMsg)
	}

	// Verify Source is preserved
	commitRangeSource, ok := pushFilesMsg.Source.(core.CommitRangeDiffSource)
	if !ok {
		t.Fatalf("Expected CommitRangeDiffSource in PushFilesScreenMsg, got %T", pushFilesMsg.Source)
	}

	if commitRangeSource.Start.Hash != commits[1].Hash {
		t.Errorf("Expected Start.Hash %s, got %s", commits[1].Hash, commitRangeSource.Start.Hash)
	}

	if commitRangeSource.End.Hash != commits[1].Hash {
		t.Errorf("Expected End.Hash %s, got %s", commits[1].Hash, commitRangeSource.End.Hash)
	}

	// Verify Files are preserved
	if len(pushFilesMsg.Files) != len(fileChanges) {
		t.Errorf("Expected %d files, got %d", len(fileChanges), len(pushFilesMsg.Files))
	}
}

func TestLogState_Update_FilesLoadedMsgWithError(t *testing.T) {
	commits := createTestCommits(3)
	testErr := errors.New("failed to load files")

	diffSource := core.NewSingleCommitRange(commits[1]).ToDiffSource()

	s := State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 1},
		ViewportStart: 0,
		Preview:       PreviewNone{},
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesLoadedMsg{
		Source: diffSource,
		Files:  nil,
		Err:    testErr,
	}

	newState, cmd := s.Update(msg, ctx)

	// Should return the same state and no command on error
	if cmd != nil {
		t.Error("Expected nil command on error, got command")
	}

	logState := newState.(State)
	if logState.CursorPosition() != 1 {
		t.Errorf("Expected state to remain unchanged, cursor at %d", logState.CursorPosition())
	}
}
