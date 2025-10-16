package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// ErrorState represents the state when an error has occurred
type ErrorState struct {
	err error
}

// View renders the error message
func (s ErrorState) View(m *Model) string {
	return fmt.Sprintf("  Error: %v\n", s.err)
}

// Update handles messages in error state
func (s ErrorState) Update(msg tea.Msg, m *Model) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}

	return s, nil
}
