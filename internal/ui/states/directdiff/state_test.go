package directdiff

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/testutils"
)

// createTestCommit creates a test commit for use in tests
func createTestCommit() core.GitCommit {
	return core.GitCommit{
		Hash:    "abc123",
		Message: "Test commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestDirectDiffLoadingState_Update_FilesLoadedWithCommitRange(t *testing.T) {
	commit := createTestCommit()
	commitRange := core.NewSingleCommitRange(commit)
	source := commitRange.ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	files := []core.FileChange{
		{Path: "file1.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "file2.go", Status: "A", Additions: 20, Deletions: 0},
	}

	msg := core.FilesLoadedMsg{
		Source: source,
		Files:  files,
		Err:    nil,
	}

	newState, cmd := s.Update(msg, ctx)

	// Should remain in DirectDiffLoadingState
	if _, ok := newState.(State); !ok {
		t.Errorf("Expected state to remain DirectDiffLoadingState, got %T", newState)
	}

	// Should return a command that produces PushFilesScreenMsg
	if cmd == nil {
		t.Fatal("Expected a command to be returned")
	}

	// Execute the command
	result := cmd()
	pushMsg, ok := result.(core.PushFilesScreenMsg)
	if !ok {
		t.Fatalf("Expected PushFilesScreenMsg, got %T", result)
	}

	// Verify the message contents
	if len(pushMsg.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(pushMsg.Files))
	}

	// Verify ExitOnPop is true for direct diff flow
	if !pushMsg.ExitOnPop {
		t.Error("Expected ExitOnPop to be true for direct diff flow")
	}

	// Verify Source type matches
	if _, ok := pushMsg.Source.(core.CommitRangeDiffSource); !ok {
		t.Errorf("Expected Source to be CommitRangeDiffSource, got %T", pushMsg.Source)
	}
}

func TestDirectDiffLoadingState_Update_FilesLoadedWithUncommittedChanges(t *testing.T) {
	tests := []struct {
		name string
		typ  core.UncommittedType
	}{
		{"unstaged", core.UncommittedTypeUnstaged},
		{"staged", core.UncommittedTypeStaged},
		{"all", core.UncommittedTypeAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := core.UncommittedChangesDiffSource{Type: tt.typ}

			s := State{Source: source}
			ctx := testutils.MockContext{W: 80, H: 24}

			files := []core.FileChange{
				{Path: "modified.go", Status: "M", Additions: 5, Deletions: 3},
			}

			msg := core.FilesLoadedMsg{
				Source: source,
				Files:  files,
				Err:    nil,
			}

			newState, cmd := s.Update(msg, ctx)

			// Should remain in DirectDiffLoadingState
			if _, ok := newState.(State); !ok {
				t.Errorf("Expected state to remain DirectDiffLoadingState, got %T", newState)
			}

			// Should return a command that produces PushFilesScreenMsg
			if cmd == nil {
				t.Fatal("Expected a command to be returned")
			}

			// Execute the command
			result := cmd()
			pushMsg, ok := result.(core.PushFilesScreenMsg)
			if !ok {
				t.Fatalf("Expected PushFilesScreenMsg, got %T", result)
			}

			// Verify ExitOnPop is true
			if !pushMsg.ExitOnPop {
				t.Error("Expected ExitOnPop to be true for direct diff flow")
			}

			// Verify Source type matches
			uncommittedSrc, ok := pushMsg.Source.(core.UncommittedChangesDiffSource)
			if !ok {
				t.Errorf("Expected Source to be UncommittedChangesDiffSource, got %T", pushMsg.Source)
			} else if uncommittedSrc.Type != tt.typ {
				t.Errorf("Expected Type to be %v, got %v", tt.typ, uncommittedSrc.Type)
			}
		})
	}
}

func TestDirectDiffLoadingState_Update_FilesLoadedError(t *testing.T) {
	commit := createTestCommit()
	commitRange := core.NewSingleCommitRange(commit)
	source := commitRange.ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesLoadedMsg{
		Source: source,
		Files:  nil,
		Err:    fmt.Errorf("not a git repository"),
	}

	newState, cmd := s.Update(msg, ctx)

	// Should remain in DirectDiffLoadingState
	if _, ok := newState.(State); !ok {
		t.Errorf("Expected state to remain DirectDiffLoadingState, got %T", newState)
	}

	// Should return a command that produces PushErrorScreenMsg
	if cmd == nil {
		t.Fatal("Expected a command to be returned")
	}

	// Execute the command
	result := cmd()
	errorMsg, ok := result.(core.PushErrorScreenMsg)
	if !ok {
		t.Fatalf("Expected PushErrorScreenMsg, got %T", result)
	}

	// Verify the error
	if errorMsg.Err == nil {
		t.Error("Expected error to be set")
	}

	if errorMsg.Err.Error() != "not a git repository" {
		t.Errorf("Expected error message 'not a git repository', got %q", errorMsg.Err.Error())
	}
}

func TestDirectDiffLoadingState_Update_EmptyFiles(t *testing.T) {
	commit := createTestCommit()
	commitRange := core.NewSingleCommitRange(commit)
	source := commitRange.ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	msg := core.FilesLoadedMsg{
		Source: source,
		Files:  []core.FileChange{},
		Err:    nil,
	}

	newState, cmd := s.Update(msg, ctx)

	// Should remain in DirectDiffLoadingState
	if _, ok := newState.(State); !ok {
		t.Errorf("Expected state to remain DirectDiffLoadingState, got %T", newState)
	}

	// Should return a command that produces PushErrorScreenMsg
	if cmd == nil {
		t.Fatal("Expected a command to be returned")
	}

	// Execute the command
	result := cmd()
	errorMsg, ok := result.(core.PushErrorScreenMsg)
	if !ok {
		t.Fatalf("Expected PushErrorScreenMsg, got %T", result)
	}

	// Verify the error
	if errorMsg.Err == nil {
		t.Error("Expected error to be set for empty files")
	}
}

func TestDirectDiffLoadingState_Update_OtherMessages(t *testing.T) {
	commit := createTestCommit()
	source := core.NewSingleCommitRange(commit).ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	// Test that other message types don't change the state
	newState, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, ctx)

	if cmd != nil {
		t.Error("Expected nil command for unhandled message")
	}

	// Should return the same state
	if _, ok := newState.(State); !ok {
		t.Error("Expected state to remain as DirectDiffLoadingState")
	}
}
