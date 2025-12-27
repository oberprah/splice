package states

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

func TestLoadingState_Update_CommitsLoaded(t *testing.T) {
	tests := []struct {
		name              string
		msg               CommitsLoadedMsg
		expectedStateType string
		checkState        func(t *testing.T, s any)
	}{
		{
			name: "successful load with commits",
			msg: CommitsLoadedMsg{
				Commits: []git.GitCommit{
					{Hash: "abc123", Message: "Test", Body: "", Author: "Author", Date: time.Now()},
					{Hash: "def456", Message: "Test2", Body: "", Author: "Author2", Date: time.Now()},
				},
				Err: nil,
			},
			expectedStateType: "LogState",
			checkState: func(t *testing.T, s any) {
				listState, ok := s.(*LogState)
				if !ok {
					t.Fatal("Expected *LogState")
				}
				if len(listState.Commits) != 2 {
					t.Errorf("Expected 2 commits, got %d", len(listState.Commits))
				}
				if listState.Cursor != 0 {
					t.Errorf("Expected cursor at 0, got %d", listState.Cursor)
				}
				if listState.ViewportStart != 0 {
					t.Errorf("Expected viewportStart at 0, got %d", listState.ViewportStart)
				}
				// Verify preview is loading for the first commit
				previewLoading, ok := listState.Preview.(PreviewLoading)
				if !ok {
					t.Errorf("Expected Preview to be PreviewLoading, got %T", listState.Preview)
				} else if previewLoading.ForHash != "abc123" {
					t.Errorf("Expected preview loading for hash 'abc123', got %q", previewLoading.ForHash)
				}
			},
		},
		{
			name: "load error",
			msg: CommitsLoadedMsg{
				Commits: nil,
				Err:     fmt.Errorf("not a git repository"),
			},
			expectedStateType: "ErrorState",
			checkState: func(t *testing.T, s any) {
				errorState, ok := s.(*ErrorState)
				if !ok {
					t.Fatal("Expected *ErrorState")
				}
				if errorState.Err == nil {
					t.Error("Expected error to be set")
				}
				if errorState.Err.Error() != "not a git repository" {
					t.Errorf("Expected error message 'not a git repository', got %q", errorState.Err.Error())
				}
			},
		},
		{
			name: "empty repository",
			msg: CommitsLoadedMsg{
				Commits: []git.GitCommit{},
				Err:     nil,
			},
			expectedStateType: "ErrorState",
			checkState: func(t *testing.T, s any) {
				errorState, ok := s.(*ErrorState)
				if !ok {
					t.Fatal("Expected *ErrorState")
				}
				if errorState.Err == nil {
					t.Error("Expected error to be set for empty repository")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := LoadingState{}
			ctx := mockContext{width: 80, height: 24}

			newState, cmd := s.Update(tt.msg, ctx)

			// Check command based on state type
			if tt.expectedStateType == "LogState" {
				// LogState should return a command to load preview
				if cmd == nil {
					t.Error("Expected a command to load preview for LogState")
				}
			} else {
				// Error states should not return a command
				if cmd != nil {
					t.Error("Expected nil command for error states")
				}
			}

			// Check state type and properties
			tt.checkState(t, newState)
		})
	}
}

func TestLoadingState_Update_OtherMessages(t *testing.T) {
	s := LoadingState{}
	ctx := mockContext{width: 80, height: 24}

	// Test that other message types don't change the state
	newState, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	// Should return the same loading state
	if _, ok := newState.(LoadingState); !ok {
		t.Error("Expected state to remain as LoadingState")
	}
}
