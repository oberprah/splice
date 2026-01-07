package error

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/app"
)

// Update handles messages in error state
func (s State) Update(msg tea.Msg, ctx app.Context) (app.State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous state using navigation pattern
			// If there's no previous state (error from LoadingState), quit the app
			return s, func() tea.Msg {
				return app.PopScreenMsg{}
			}
		case "ctrl+c", "Q":
			return s, tea.Quit
		}
	}

	return s, nil
}
