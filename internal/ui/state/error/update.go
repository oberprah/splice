package error

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/ui/state"
)

// Update handles messages in error state
func (s State) Update(msg tea.Msg, ctx state.Context) (state.State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit
		}
	}

	return s, nil
}
