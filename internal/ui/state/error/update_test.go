package error

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestErrorState_Update_QuitKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"q key", "q"},
		{"ctrl+c", "ctrl+c"},
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

			// Should return tea.Quit command
			if cmd == nil {
				t.Error("Expected tea.Quit command, got nil")
				return
			}

			// Verify it's actually tea.Quit by executing it
			resultMsg := cmd()
			if _, ok := resultMsg.(tea.QuitMsg); !ok {
				t.Error("Expected tea.QuitMsg from command")
			}

			// State should be unchanged
			if errorState, ok := newState.(State); !ok {
				t.Error("Expected state to remain as error.State")
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
				t.Error("Expected state to remain as error.State")
			} else if errorState.Err.Error() != "test error" {
				t.Error("Error message should be unchanged")
			}
		})
	}
}
