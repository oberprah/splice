package files

import (
	"fmt"
	"testing"

	"github.com/alecthomas/chroma/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/filetree"
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
	s := New(commitRange.ToDiffSource(), files)
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
	s := New(commitRange.ToDiffSource(), files)
	s.Cursor = 5
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
	s := New(commitRange.ToDiffSource(), files)
	s.Cursor = 5
	s.ViewportStart = 3
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
	s := New(commitRange.ToDiffSource(), files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Press "G" to jump to bottom
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	newState, _ := s.Update(msg, ctx)

	filesState, ok := newState.(*State)
	if !ok {
		t.Fatal("Expected state to remain FilesState")
	}

	// With tree structure, the last visible item may not be at index 9
	// Just check that we jumped to the last visible item
	expectedLast := len(filesState.VisibleItems) - 1
	if filesState.Cursor != expectedLast {
		t.Errorf("Expected cursor to jump to last visible item (%d), got %d", expectedLast, filesState.Cursor)
	}
}

func TestFilesState_Update_NavigationBoundaries(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	commitRange := core.NewSingleCommitRange(commit)
	diffSource := commitRange.ToDiffSource()

	// Create state once to get the number of visible items
	s := New(diffSource, files)
	lastIdx := len(s.VisibleItems) - 1

	for _, tt := range []struct {
		name           string
		initialCursor  int
		key            rune
		expectedCursor int
	}{
		{"up at top stays at top", 0, 'k', 0},
		{"down at bottom stays at bottom", lastIdx, 'j', lastIdx},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := New(diffSource, files)
			s.Cursor = tt.initialCursor
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
			s := New(createTestDiffSource(commit), files)
			s.Cursor = tt.initialCursor
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
	s := New(commitRange.ToDiffSource(), files)
	s.Cursor = 5
	ctx := testutils.MockContext{W: 80, H: 10}

	initialCursor := s.Cursor

	// Move cursor down multiple times to trigger viewport scrolling
	state := s
	for range 10 {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
		newState, _ := state.Update(msg, ctx)
		state = newState.(*State)
	}

	// After moving down 10 times from initial position, cursor should have moved
	expectedCursor := min(initialCursor+10, len(state.VisibleItems)-1)
	if state.Cursor != expectedCursor {
		t.Errorf("Expected cursor to be at %d, got %d", expectedCursor, state.Cursor)
	}

	// Viewport should have scrolled to keep cursor visible
	if state.ViewportStart == 0 {
		t.Error("Expected viewport to have scrolled down")
	}
}

func TestFilesState_Update_BackNavigation(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	s := New(createTestDiffSource(commit), files)
	s.Cursor = 2
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
	s := New(commitRange.ToDiffSource(), files)
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
	s := New(commitRange.ToDiffSource(), files)
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
	s := New(createTestDiffSource(commit), files)

	// Find a file node in VisibleItems
	fileIdx := -1
	for i, item := range s.VisibleItems {
		if _, ok := item.Node.(*filetree.FileNode); ok {
			fileIdx = i
			break
		}
	}

	if fileIdx == -1 {
		t.Skip("No file nodes in visible items")
	}

	s.Cursor = fileIdx
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
	s := New(createTestDiffSource(commit), files)
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
	s := New(createTestDiffSource(commit), files)
	s.Cursor = 2
	s.ViewportStart = 1
	ctx := testutils.MockContext{W: 80, H: 24}

	// Simulate DiffLoadedMsg with success using new block-based structure
	msg := core.DiffLoadedMsg{
		Source:    core.NewSingleCommitRange(commit).ToDiffSource(),
		Files:     files,
		FileIndex: 2,
		File:      files[2],
		Diff: &diff.FileDiff{
			Path: "file.go",
			Blocks: []diff.Block{
				diff.UnchangedBlock{Lines: []diff.LinePair{
					{LeftLineNo: 1, RightLineNo: 1, Tokens: []highlight.Token{{Type: chroma.Text, Value: "test"}}},
				}},
			},
		},
		Err: nil,
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

	if len(pushMsg.Diff.Blocks) != 1 {
		t.Errorf("Expected 1 diff block, got %d", len(pushMsg.Diff.Blocks))
	}

	// Verify state hasn't changed (stays FilesState until Model handles PushScreenMsg)
	if filesState != s {
		t.Error("Expected state to remain unchanged")
	}
}

func TestFilesState_Update_DiffLoadedMsgError(t *testing.T) {
	commit := createTestCommit()
	files := createTestFileChanges(5)
	s := New(createTestDiffSource(commit), files)
	s.Cursor = 2
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

// Tests for tree toggle functionality

func TestFilesState_Update_EnterOnFolder_TogglesExpanded(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find the index of the src/ folder in VisibleItems
	srcFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if folder.GetName() == "src" {
				srcFolderIdx = i
				break
			}
		}
	}

	if srcFolderIdx == -1 {
		t.Fatal("Expected to find src/ folder in visible items")
	}

	// Position cursor on the folder
	s.Cursor = srcFolderIdx

	// Initial state: folder should be expanded
	initialVisibleCount := len(s.VisibleItems)

	// Press Enter to toggle (collapse)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, _ := s.Update(msg, ctx)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState)
	}

	// After collapsing, should have fewer visible items
	if len(filesState.VisibleItems) >= initialVisibleCount {
		t.Errorf("Expected fewer visible items after collapse, got %d (was %d)", len(filesState.VisibleItems), initialVisibleCount)
	}

	// Press Enter again to toggle (expand)
	msg = tea.KeyMsg{Type: tea.KeyEnter}
	newState2, _ := filesState.Update(msg, ctx)
	filesState2, ok := newState2.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState2)
	}

	// After expanding, should be back to original count
	if len(filesState2.VisibleItems) != initialVisibleCount {
		t.Errorf("Expected %d visible items after re-expand, got %d", initialVisibleCount, len(filesState2.VisibleItems))
	}
}

func TestFilesState_Update_SpaceOnFolder_TogglesExpanded(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find the src/ folder
	srcFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if folder.GetName() == "src" {
				srcFolderIdx = i
				break
			}
		}
	}

	if srcFolderIdx == -1 {
		t.Fatal("Expected to find src/ folder in visible items")
	}

	s.Cursor = srcFolderIdx
	initialVisibleCount := len(s.VisibleItems)

	// Press Space to toggle (collapse)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	newState, _ := s.Update(msg, ctx)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState)
	}

	// Should have fewer visible items after collapse
	if len(filesState.VisibleItems) >= initialVisibleCount {
		t.Errorf("Expected fewer visible items after collapse, got %d (was %d)", len(filesState.VisibleItems), initialVisibleCount)
	}
}

func TestFilesState_Update_RightArrowOnFolder_ExpandsOnly(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find and collapse the src/ folder first
	srcFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if folder.GetName() == "src" {
				srcFolderIdx = i
				break
			}
		}
	}

	if srcFolderIdx == -1 {
		t.Fatal("Expected to find src/ folder in visible items")
	}

	s.Cursor = srcFolderIdx

	// Collapse it first with Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, _ := s.Update(msg, ctx)
	filesState := newState.(*State)

	collapsedVisibleCount := len(filesState.VisibleItems)

	// Now use Right arrow to expand
	msg = tea.KeyMsg{Type: tea.KeyRight}
	newState2, _ := filesState.Update(msg, ctx)
	filesState2, ok := newState2.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState2)
	}

	// Should have more visible items after expand
	if len(filesState2.VisibleItems) <= collapsedVisibleCount {
		t.Errorf("Expected more visible items after expand, got %d (was %d)", len(filesState2.VisibleItems), collapsedVisibleCount)
	}

	// Pressing Right again on already expanded folder should be a no-op
	expandedVisibleCount := len(filesState2.VisibleItems)
	msg = tea.KeyMsg{Type: tea.KeyRight}
	newState3, _ := filesState2.Update(msg, ctx)
	filesState3 := newState3.(*State)

	if len(filesState3.VisibleItems) != expandedVisibleCount {
		t.Errorf("Expected no change when pressing Right on expanded folder, got %d items (was %d)", len(filesState3.VisibleItems), expandedVisibleCount)
	}
}

func TestFilesState_Update_LeftArrowOnFolder_CollapsesOnly(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find the src/ folder (should be expanded by default)
	srcFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if folder.GetName() == "src" {
				srcFolderIdx = i
				break
			}
		}
	}

	if srcFolderIdx == -1 {
		t.Fatal("Expected to find src/ folder in visible items")
	}

	s.Cursor = srcFolderIdx
	initialVisibleCount := len(s.VisibleItems)

	// Press Left to collapse
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	newState, _ := s.Update(msg, ctx)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState)
	}

	// Should have fewer visible items after collapse
	collapsedVisibleCount := len(filesState.VisibleItems)
	if collapsedVisibleCount >= initialVisibleCount {
		t.Errorf("Expected fewer visible items after collapse, got %d (was %d)", collapsedVisibleCount, initialVisibleCount)
	}

	// Pressing Left again on already collapsed folder should be a no-op
	msg = tea.KeyMsg{Type: tea.KeyLeft}
	newState2, _ := filesState.Update(msg, ctx)
	filesState2 := newState2.(*State)

	if len(filesState2.VisibleItems) != collapsedVisibleCount {
		t.Errorf("Expected no change when pressing Left on collapsed folder, got %d items (was %d)", len(filesState2.VisibleItems), collapsedVisibleCount)
	}
}

func TestFilesState_Update_EnterOnFile_OpensFileDiff(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find a file node in VisibleItems
	fileIdx := -1
	for i, item := range s.VisibleItems {
		if _, ok := item.Node.(*filetree.FileNode); ok {
			fileIdx = i
			break
		}
	}

	if fileIdx == -1 {
		t.Fatal("Expected to find a file node in visible items")
	}

	s.Cursor = fileIdx

	// Press Enter on file
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, cmd := s.Update(msg, ctx)

	// Should stay in FilesState while loading
	if _, ok := newState.(*State); !ok {
		t.Fatalf("Expected to stay in FilesState while loading, got %T", newState)
	}

	// Should return a command to load the diff
	if cmd == nil {
		t.Error("Expected a command to load the diff for file")
	}
}

func TestFilesState_Update_TogglePreservesCursorPosition(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
		{Path: "docs/README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find the src/ folder
	srcFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if folder.GetName() == "src" {
				srcFolderIdx = i
				break
			}
		}
	}

	if srcFolderIdx == -1 {
		t.Fatal("Expected to find src/ folder in visible items")
	}

	s.Cursor = srcFolderIdx

	// Toggle (collapse)
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, _ := s.Update(msg, ctx)
	filesState := newState.(*State)

	// Cursor should still be on the src/ folder (or adjusted appropriately)
	// At minimum, cursor should be valid
	if filesState.Cursor < 0 || filesState.Cursor >= len(filesState.VisibleItems) {
		t.Errorf("Expected cursor to be valid after toggle, got %d (visible items: %d)", filesState.Cursor, len(filesState.VisibleItems))
	}
}

func TestFilesState_Update_ArrowKeysOnFile_NoToggle(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find the file node
	fileIdx := -1
	for i, item := range s.VisibleItems {
		if _, ok := item.Node.(*filetree.FileNode); ok {
			fileIdx = i
			break
		}
	}

	if fileIdx == -1 {
		t.Fatal("Expected to find file node in visible items")
	}

	s.Cursor = fileIdx
	initialVisibleCount := len(s.VisibleItems)

	// Press Left on file (should be no-op)
	msg := tea.KeyMsg{Type: tea.KeyLeft}
	newState, _ := s.Update(msg, ctx)
	filesState := newState.(*State)

	if len(filesState.VisibleItems) != initialVisibleCount {
		t.Error("Expected no change when pressing Left on file")
	}

	// Press Right on file (should be no-op)
	msg = tea.KeyMsg{Type: tea.KeyRight}
	newState2, _ := filesState.Update(msg, ctx)
	filesState2 := newState2.(*State)

	if len(filesState2.VisibleItems) != initialVisibleCount {
		t.Error("Expected no change when pressing Right on file")
	}
}

func TestFilesState_Update_CollapseFolder_MaintainsViewportPosition(t *testing.T) {
	// This test reproduces the bug where collapsing a folder causes unexpected viewport scrolling.
	// Expected: After collapsing a folder, the viewport should remain stable and show items above
	//           the collapsed folder.
	// Actual (buggy): The viewport scrolls down, hiding all items above the collapsed folder.

	commit := createTestCommit()
	// Create a file structure with multiple top-level folders
	// This mimics the structure shown in the bug report:
	// docs/, internal/, sandbox/, test/
	files := []core.FileChange{
		// docs/ folder
		{Path: "docs/specs/file1.md", Status: "A", Additions: 100, Deletions: 0},
		{Path: "docs/specs/file2.md", Status: "A", Additions: 200, Deletions: 0},
		{Path: "docs/specs/file3.md", Status: "A", Additions: 300, Deletions: 0},

		// internal/ folder with many files
		{Path: "internal/ui/states/files/state.go", Status: "M", Additions: 50, Deletions: 10},
		{Path: "internal/ui/states/files/update.go", Status: "M", Additions: 100, Deletions: 20},
		{Path: "internal/ui/states/files/view.go", Status: "M", Additions: 30, Deletions: 5},
		{Path: "internal/domain/tree/filetree.go", Status: "A", Additions: 200, Deletions: 0},

		// sandbox/ folder
		{Path: "sandbox/docker-compose.yml", Status: "M", Additions: 3, Deletions: 3},
		{Path: "sandbox/sandbox.sh", Status: "M", Additions: 5, Deletions: 3},

		// test/ folder with many files (this is what we'll collapse)
		{Path: "test/e2e/testdata/file1.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "test/e2e/testdata/file2.go", Status: "M", Additions: 15, Deletions: 8},
		{Path: "test/e2e/tree_navigation_test.go", Status: "A", Additions: 431, Deletions: 0},
	}

	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 20}

	// Find the test/e2e folder in visible items
	testE2EFolderIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			// Looking for collapsed path "test/e2e"
			if folder.GetName() == "test/e2e" {
				testE2EFolderIdx = i
				break
			}
		}
	}

	if testE2EFolderIdx == -1 {
		t.Fatal("Expected to find test/e2e folder in visible items")
	}

	// Position cursor on test/e2e folder
	s.Cursor = testE2EFolderIdx
	s.ViewportStart = 0 // Start with viewport at the top

	// Record the initial viewport state
	initialViewportStart := s.ViewportStart
	initialVisibleCount := len(s.VisibleItems)

	// Collapse the test/e2e folder by pressing Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, _ := s.Update(msg, ctx)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState)
	}

	// After collapse, visible items count should decrease
	if len(filesState.VisibleItems) >= initialVisibleCount {
		t.Errorf("Expected fewer visible items after collapse, got %d (was %d)",
			len(filesState.VisibleItems), initialVisibleCount)
	}

	// BUG: The viewport should not scroll down significantly just because we collapsed a folder.
	// If the cursor was visible before, the viewport should remain stable.
	// With the bug, ViewportStart gets set to a value near the cursor position,
	// hiding all the folders above it.

	// The viewport should either:
	// 1. Stay at 0 (if cursor was already visible), or
	// 2. Adjust minimally to keep cursor visible
	//
	// But it should NOT scroll down to hide the docs/, internal/, and sandbox/ folders
	// that were visible before the collapse.

	// With a height of 20 and cursor around index 3-5 (depending on exact tree structure),
	// the viewport should still start at 0 or very close to it.
	// The bug causes it to jump to a much higher value (close to cursor position).

	if filesState.ViewportStart > 2 {
		t.Errorf("BUG: Viewport scrolled unexpectedly far after collapse. "+
			"ViewportStart=%d (was %d). This hides folders that should remain visible. "+
			"Expected ViewportStart to remain at 0 or adjust minimally.",
			filesState.ViewportStart, initialViewportStart)
	}

	// Additional verification: With ViewportStart near 0 and cursor still valid,
	// we should be able to see multiple folders in the viewport
	// (not just the one collapsed folder)
	if filesState.ViewportStart == filesState.Cursor {
		t.Errorf("BUG: ViewportStart (%d) equals Cursor (%d), meaning only the cursor line "+
			"and items below it are visible. Items above the cursor are hidden.",
			filesState.ViewportStart, filesState.Cursor)
	}
}

func TestFilesState_Update_ToggleFolderWithDuplicateName_TogglesCorrectFolder(t *testing.T) {
	// This test reproduces the bug where toggling a folder with a duplicate name AND depth
	// incorrectly toggles the first folder with that name instead of the one under the cursor.
	//
	// Setup: Create a tree structure where path collapsing creates two folders with the same name+depth
	// Tree before collapsing:
	//   frontend/
	//     src/
	//       components/
	//         file1.js
	//         file2.js
	//   backend/
	//     src/
	//       components/
	//         file3.go
	//         file4.go
	//
	// After collapsing single-child chains:
	//   frontend/
	//     src/components/    <- name="src/components", depth=1
	//       file1.js
	//       file2.js
	//   backend/
	//     src/components/    <- name="src/components", depth=1 (SAME as frontend!)
	//       file3.go
	//       file4.go
	//
	// Expected: When cursor is on "backend/src/components/" and Enter is pressed,
	//           only "backend/src/components/" should be collapsed.
	// Actual (buggy): The "frontend/src/components/" gets collapsed instead because findFolderInTree
	//                 matches on name+depth only, returning the first match.

	commit := createTestCommit()
	files := []core.FileChange{
		// frontend/ has multiple children, preventing collapse
		{Path: "frontend/package.json", Status: "M", Additions: 2, Deletions: 1},
		{Path: "frontend/src/components/Button.js", Status: "M", Additions: 10, Deletions: 5},
		{Path: "frontend/src/components/Input.js", Status: "A", Additions: 20, Deletions: 0},

		// backend/ has multiple children, preventing collapse
		{Path: "backend/go.mod", Status: "M", Additions: 1, Deletions: 1},
		{Path: "backend/src/components/Handler.go", Status: "M", Additions: 15, Deletions: 3},
		{Path: "backend/src/components/Parser.go", Status: "A", Additions: 25, Deletions: 0},
	}

	source := createTestDiffSource(commit)
	s := New(source, files)
	ctx := testutils.MockContext{W: 80, H: 24}

	// Find both "src/components" folders in the visible items
	// Both should have name="src/components" and depth=1
	frontendSrcIdx := -1
	backendSrcIdx := -1
	for i, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			name := folder.GetName()
			depth := folder.GetDepth()

			if name == "src/components" && depth == 1 {
				// First src/components folder encountered should be backend (alphabetical)
				if backendSrcIdx == -1 {
					backendSrcIdx = i
				} else if frontendSrcIdx == -1 {
					frontendSrcIdx = i
				}
			}
		}
	}

	if backendSrcIdx == -1 {
		t.Fatal("Expected to find backend/src/components folder in visible items")
	}
	if frontendSrcIdx == -1 {
		t.Fatal("Expected to find frontend/src/components folder in visible items")
	}

	// Verify both folders start expanded
	backendSrcFolder := s.VisibleItems[backendSrcIdx].Node.(*filetree.FolderNode)
	frontendSrcFolder := s.VisibleItems[frontendSrcIdx].Node.(*filetree.FolderNode)

	if !backendSrcFolder.IsExpanded() {
		t.Fatal("Expected backend/src/components folder to start expanded")
	}
	if !frontendSrcFolder.IsExpanded() {
		t.Fatal("Expected frontend/src/components folder to start expanded")
	}

	// Verify they have the same name and depth (this is the bug condition)
	if backendSrcFolder.GetName() != frontendSrcFolder.GetName() {
		t.Fatalf("Expected both folders to have the same name, got %q and %q",
			backendSrcFolder.GetName(), frontendSrcFolder.GetName())
	}
	if backendSrcFolder.GetDepth() != frontendSrcFolder.GetDepth() {
		t.Fatalf("Expected both folders to have the same depth, got %d and %d",
			backendSrcFolder.GetDepth(), frontendSrcFolder.GetDepth())
	}

	// Position cursor on frontend/src/components (the second "src/components" folder)
	s.Cursor = frontendSrcIdx

	// Press Enter to toggle frontend/src/components
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newState, _ := s.Update(msg, ctx)
	filesState, ok := newState.(*State)
	if !ok {
		t.Fatalf("Expected state to remain FilesState, got %T", newState)
	}

	// Find both folders again in the new state
	newBackendSrcIdx := -1
	newFrontendSrcIdx := -1
	for i, item := range filesState.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			name := folder.GetName()
			depth := folder.GetDepth()

			if name == "src/components" && depth == 1 {
				if newBackendSrcIdx == -1 {
					newBackendSrcIdx = i
				} else if newFrontendSrcIdx == -1 {
					newFrontendSrcIdx = i
				}
			}
		}
	}

	if newBackendSrcIdx == -1 {
		t.Fatal("Expected to find backend/src/components folder in visible items after toggle")
	}
	if newFrontendSrcIdx == -1 {
		t.Fatal("Expected to find frontend/src/components folder in visible items after toggle")
	}

	newBackendSrcFolder := filesState.VisibleItems[newBackendSrcIdx].Node.(*filetree.FolderNode)
	newFrontendSrcFolder := filesState.VisibleItems[newFrontendSrcIdx].Node.(*filetree.FolderNode)

	// BUG VERIFICATION: backend/src/components should still be expanded (we didn't toggle it)
	// but with the bug, it gets collapsed instead of frontend/src/components
	if !newBackendSrcFolder.IsExpanded() {
		t.Errorf("BUG DETECTED: backend/src/components folder was incorrectly collapsed. " +
			"Expected it to remain expanded (cursor was on frontend/src/components, not backend/src/components). " +
			"This happens because findFolderInTree matches on name+depth only, " +
			"returning the first match instead of the folder under the cursor.")
	}

	// The correct behavior: frontend/src/components should be collapsed, backend/src/components should remain expanded
	if newFrontendSrcFolder.IsExpanded() {
		t.Errorf("Expected frontend/src/components to be collapsed (it was toggled), but it is still expanded")
	}

	// Additional verification: files under backend/src/components should still be visible
	foundBackendFile := false
	for _, item := range filesState.VisibleItems {
		if file, ok := item.Node.(*filetree.FileNode); ok {
			if file.File().Path == "backend/src/components/Handler.go" || file.File().Path == "backend/src/components/Parser.go" {
				foundBackendFile = true
				break
			}
		}
	}

	if !foundBackendFile {
		t.Errorf("Expected to still see files under backend/src/components (should remain expanded), but they are not visible")
	}

	// Files under frontend/src/components should NOT be visible (it's collapsed)
	foundFrontendFile := false
	for _, item := range filesState.VisibleItems {
		if file, ok := item.Node.(*filetree.FileNode); ok {
			if file.File().Path == "frontend/src/components/Button.js" || file.File().Path == "frontend/src/components/Input.js" {
				foundFrontendFile = true
				break
			}
		}
	}

	if foundFrontendFile {
		t.Errorf("Expected frontend/src/components files to be hidden (folder should be collapsed), but they are still visible")
	}
}
