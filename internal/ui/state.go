package ui

import tea "github.com/charmbracelet/bubbletea"

// State represents the current state of the application.
// Each state implementation handles its own update logic and rendering.
type State interface {
	// View renders the state with access to model dimensions
	View(m *Model) string

	// Update handles messages and returns the next state
	Update(msg tea.Msg, m *Model) (State, tea.Cmd)
}
