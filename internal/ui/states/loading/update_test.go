package loading

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

func TestLoadingState_Update_CommitsLoaded(t *testing.T) {
	tests := []struct {
		name               string
		msg                core.CommitsLoadedMsg
		expectLoadingState bool
		checkCmd           func(t *testing.T, cmd tea.Cmd)
	}{
		{
			name: "successful load with commits",
			msg: core.CommitsLoadedMsg{
				Commits: []git.GitCommit{
					{Hash: "abc123", Message: "Test", Body: "", Author: "Author", Date: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)},
					{Hash: "def456", Message: "Test2", Body: "", Author: "Author2", Date: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)},
				},
				Err: nil,
			},
			expectLoadingState: true,
			checkCmd: func(t *testing.T, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatal("Expected a command to be returned")
				}

				// Execute the command - it should return PushLogScreenMsg
				msg := cmd()
				pushMsg, ok := msg.(core.PushLogScreenMsg)
				if !ok {
					t.Fatalf("Expected PushLogScreenMsg, got %T", msg)
				}

				if len(pushMsg.Commits) != 2 {
					t.Errorf("Expected 2 commits, got %d", len(pushMsg.Commits))
				}

				if pushMsg.GraphLayout == nil {
					t.Error("Expected GraphLayout to be set")
				}
			},
		},
		{
			name: "load error",
			msg: core.CommitsLoadedMsg{
				Commits: nil,
				Err:     fmt.Errorf("not a git repository"),
			},
			expectLoadingState: true,
			checkCmd: func(t *testing.T, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatal("Expected a command to be returned")
				}

				// Execute the command - it should return PushErrorScreenMsg
				msg := cmd()
				pushMsg, ok := msg.(core.PushErrorScreenMsg)
				if !ok {
					t.Fatalf("Expected PushErrorScreenMsg, got %T", msg)
				}

				if pushMsg.Err == nil {
					t.Error("Expected error to be set")
				}
				if pushMsg.Err.Error() != "not a git repository" {
					t.Errorf("Expected error message 'not a git repository', got %q", pushMsg.Err.Error())
				}
			},
		},
		{
			name: "empty repository",
			msg: core.CommitsLoadedMsg{
				Commits: []git.GitCommit{},
				Err:     nil,
			},
			expectLoadingState: true,
			checkCmd: func(t *testing.T, cmd tea.Cmd) {
				if cmd == nil {
					t.Fatal("Expected a command to be returned")
				}

				// Execute the command - it should return PushErrorScreenMsg
				msg := cmd()
				pushMsg, ok := msg.(core.PushErrorScreenMsg)
				if !ok {
					t.Fatalf("Expected PushErrorScreenMsg, got %T", msg)
				}

				if pushMsg.Err == nil {
					t.Error("Expected error to be set for empty repository")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{}
			ctx := mockContext{width: 80, height: 24}

			newState, cmd := s.Update(tt.msg, ctx)

			if tt.expectLoadingState {
				// State should remain LoadingState when returning navigation message
				if _, ok := newState.(State); !ok {
					t.Errorf("Expected LoadingState, got %T", newState)
				}
				if tt.checkCmd != nil {
					tt.checkCmd(t, cmd)
				}
			}
		})
	}
}

func TestLoadingState_Update_OtherMessages(t *testing.T) {
	s := State{}
	ctx := mockContext{width: 80, height: 24}

	// Test that other message types don't change the state
	newState, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}, ctx)

	if cmd != nil {
		t.Error("Expected nil command")
	}

	// Should return the same loading state
	if _, ok := newState.(State); !ok {
		t.Error("Expected state to remain as LoadingState")
	}
}
