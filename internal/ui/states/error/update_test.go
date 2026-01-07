package error

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
)

func TestErrorState_Update_QuitKeys(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		expectQuit      bool
		expectPopScreen bool
	}{
		{"q key", "q", false, true},
		{"ctrl+c", "ctrl+c", true, false},
		{"Q key", "Q", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{Err: fmt.Errorf("test error")}
			ctx := mockContext{width: 80, height: 24}

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			newState, cmd := s.Update(msg, ctx)

			// Should return a command
			if cmd == nil {
				t.Error("Expected command, got nil")
				return
			}

			// Verify it returns the expected message type
			resultMsg := cmd()
			if tt.expectQuit {
				if _, ok := resultMsg.(tea.QuitMsg); !ok {
					t.Errorf("Expected tea.QuitMsg from command, got %T", resultMsg)
				}
			} else if tt.expectPopScreen {
				if _, ok := resultMsg.(core.PopScreenMsg); !ok {
					t.Errorf("Expected PopScreenMsg from command, got %T", resultMsg)
				}
			}

			// State should be unchanged
			if errorState, ok := newState.(State); !ok {
				t.Error("Expected state to remain as ErrorState")
			} else if errorState.Err.Error() != "test error" {
				t.Error("Error message should be unchanged")
			}
		})
	}
}

func TestErrorState_Update_OtherKeys(t *testing.T) {
	s := State{Err: fmt.Errorf("test error")}
	ctx := mockContext{width: 80, height: 24}

	// Test various other keys that should do nothing
	keys := []string{"j", "k", "enter", "esc"}

	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}

			newState, cmd := s.Update(msg, ctx)

			// Should return no command
			if cmd != nil {
				t.Error("Expected nil command for non-quit keys")
			}

			// State should be unchanged
			if errorState, ok := newState.(State); !ok {
				t.Error("Expected state to remain as ErrorState")
			} else if errorState.Err.Error() != "test error" {
				t.Error("Error message should be unchanged")
			}
		})
	}
}
