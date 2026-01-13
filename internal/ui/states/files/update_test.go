package files

import (
	"fmt"
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

func TestFilesState_Update_NavigationDown(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(10)
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "j" to move down
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
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
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        5,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "k" to move up
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
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
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        5,
		ViewportStart: 3,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "g" to jump to top
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
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
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
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
	commitRange := core.NewSingleCommitRange(commit)
	diffSource := commitRange.ToDiffSource()

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
			s := State{
				Source:        diffSource,
				Files:         files,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
			}
			ctx := testutils.MockContext{W: 80, H: 24}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}}
			newState, _ := s.Update(msg, ctx)

			filesState, ok := newState.(*State)
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
			s := State{
				Source:        createTestDiffSource(commit),
				Files:         files,
				Cursor:        tt.initialCursor,
				ViewportStart: 0,
			}
			ctx := testutils.MockContext{W: 80, H: 24}

			msg := tea.KeyMsg{Type: tt.keyType}
			newState, _ := s.Update(msg, ctx)

			filesState, ok := newState.(*State)
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
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        5,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 10}

	// Move cursor down multiple times to trigger viewport scrolling
	state := &s
	for range 10 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newState, _ := state.Update(msg, ctx)
		state = newState.(*State)
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
	s := State{
		Source:        createTestDiffSource(commit),
		Files:         files,
		Cursor:        2,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "q" to go back
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	newState, cmd := s.Update(msg, ctx)

	// Should remain in FilesState
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected to remain in FilesState, got %T", newState)
	}

	// State should be unchanged
	if filesState.Cursor != 2 {
		t.Errorf("Expected cursor to remain at 2, got %d", filesState.Cursor)
	}

	// Should return a command that produces PopScreenMsg
	if cmd == nil {
		t.Fatal("Expected command to produce PopScreenMsg")
	}

	// Execute the command to get the message
	cmdMsg := cmd()
	popMsg, ok := cmdMsg.(core.PopScreenMsg)
	if !ok {
		t.Fatalf("Expected PopScreenMsg, got %T", cmdMsg)
	}

	// Verify it's a PopScreenMsg (no fields to check)
	_ = popMsg
}

func TestFilesState_Update_SingleFileList(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(1)
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Try to move down (should stay at 0)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 with single file, got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_EmptyFileList(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{}
	commitRange := core.NewSingleCommitRange(commit)
	s := State{
		Source:        commitRange.ToDiffSource(),
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Try to move down (should stay at 0)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	if filesState.Cursor != 0 {
		t.Errorf("Expected cursor to stay at 0 with empty files, got %d", filesState.Cursor)
	}
}

func TestFilesState_Update_EnterKeyReturnsCommand(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	s := State{
		Source:        createTestDiffSource(commit),
		Files:         files,
		Cursor:        2,
		ViewportStart: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "enter" to select a file
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, cmd := s.Update(msg, ctx)

	// Should stay in FilesState while loading
	if _, ok := newState.(*State); !ok {
		t.Fatalf("Expected to stay in FilesState while loading, got %T", newState)
	}

	// Should return a command to load the diff
	if cmd == nil {
		t.Error("Expected a command to load the diff")
	}
}

func TestFilesState_Update_EnterOnEmptyFiles(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{}
	s := State{
		Source: createTestDiffSource(commit), // OLD:CommitRange: core.NewSingleCommitRange(commit),
		Files:  files,
		Cursor: 0,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "enter" with no files
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, cmd := s.Update(msg, ctx)

	// Should stay in FilesState
	if _, ok := newState.(*State); !ok {
		t.Fatalf("Expected to stay in FilesState, got %T", newState)
	}

	// Should NOT return a command
	if cmd != nil {
		t.Error("Expected no command when files list is empty")
	}
}

func TestFilesState_Update_DiffLoadedMsgSuccess(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	s := State{
		Source:        createTestDiffSource(commit),
		Files:         files,
		Cursor:        2,
		ViewportStart: 1,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Simulate DiffLoadedMsg with success
	msg := core.DiffLoadedMsg{
		Source: core.NewSingleCommitRange(commit).ToDiffSource(),
		File:   files[2],
		Diff: &diff.AlignedFileDiff{
			Left: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "test"}}},
				},
			},
			Right: diff.FileContent{
				Path: "file.go",
				Lines: []diff.AlignedLine{
					{Tokens: []highlight.Token{{Type: chroma.Text, Value: "test"}}},
				},
			},
			Alignments: []diff.Alignment{
				diff.UnchangedAlignment{LeftIdx: 0, RightIdx: 0},
			},
		},
		ChangeIndices: []int{},
		Err:           nil,
	}
	newState, cmd := s.Update(msg, ctx)

	// Should stay in FilesState (it returns a command that produces PushScreenMsg)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected to stay in FilesState, got %T", newState)
	}

	// Should return a command that produces PushScreenMsg
	if cmd == nil {
		t.Fatal("Expected command for navigation")
	}

	// Execute the command to get the message
	result := cmd()
	if result == nil {
		t.Fatal("Expected command to return a message")
	}

	// Verify it's a PushDiffScreenMsg
	pushMsg, ok := result.(core.PushDiffScreenMsg)
	if !ok {
		t.Fatalf("Expected PushDiffScreenMsg, got %T", result)
	}

	if len(pushMsg.Diff.Alignments) != 1 {
		t.Errorf("Expected 1 diff alignment, got %d", len(pushMsg.Diff.Alignments))
	}

	// Verify state hasn't changed (stays FilesState until Model handles PushScreenMsg)
	if filesState != &s {
		t.Error("Expected state to remain unchanged")
	}
}

func TestFilesState_Update_DiffLoadedMsgError(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	s := State{
		Source: createTestDiffSource(commit), // OLD:CommitRange: core.NewSingleCommitRange(commit),
		Files:  files,
		Cursor: 2,
	}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Simulate DiffLoadedMsg with error
	msg := core.DiffLoadedMsg{
		Source: core.NewSingleCommitRange(commit).ToDiffSource(),
		File:   files[2],
		Err:    fmt.Errorf("failed to load diff"),
	}
	newState, _ := s.Update(msg, ctx)

	// Should stay in FilesState on error
	if _, ok := newState.(*State); !ok {
		t.Fatalf("Expected to stay in FilesState on error, got %T", newState)
	}
}

// TestFetchFileDiffForSource_CommitRange tests that commit range sources use the injected function
func TestFetchFileDiffForSource_CommitRange(t *testing.T) {
	commit := createTestCommit()
	commitRange := core.NewSingleCommitRange(commit)
	source := commitRange.ToDiffSource()
	file := core.FileChange{Path: "test.go", Status: "M"}

	// Mock function that tracks if it was called
	called := false
	mockFetch := func(cr core.CommitRange, f core.FileChange) (*core.FullFileDiffResult, error) {
		called = true
		if cr.Start.Hash != commit.Hash || cr.End.Hash != commit.Hash {
			t.Errorf("Expected commit range with hash %s, got %s..%s", commit.Hash, cr.Start.Hash, cr.End.Hash)
		}
		if f.Path != file.Path {
			t.Errorf("Expected file path %s, got %s", file.Path, f.Path)
		}
		return &core.FullFileDiffResult{
			OldContent: "old",
			NewContent: "new",
			DiffOutput: "diff",
		}, nil
	}

	result, err := fetchFileDiffForSource(source, file, mockFetch)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !called {
		t.Error("Expected mock function to be called")
	}
	if result.OldContent != "old" {
		t.Errorf("Expected old content 'old', got %s", result.OldContent)
	}
}

// TestFetchFileDiffForSource_UnknownType tests error handling for unknown diff source types
func TestFetchFileDiffForSource_UnknownType(t *testing.T) {
	// Create an invalid source by type assertion (this is for testing error handling)
	file := core.FileChange{Path: "test.go"}
	mockFetch := func(cr core.CommitRange, f core.FileChange) (*core.FullFileDiffResult, error) {
		return nil, fmt.Errorf("should not be called")
	}

	// Test with nil source (will cause type switch default case)
	_, err := fetchFileDiffForSource(nil, file, mockFetch)
	if err == nil {
		t.Error("Expected error for unknown diff source type")
	}
	if err != nil && !contains(err.Error(), "unknown diff source type") {
		t.Errorf("Expected 'unknown diff source type' error, got: %v", err)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
