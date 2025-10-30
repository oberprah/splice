package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages in error state
func (s ErrorState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}

	return s, nil
}
