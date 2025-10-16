package state

import tea "github.com/charmbracelet/bubbletea"

// Context is the interface that states use to access what they need from the model
type Context interface {
	Width() int
	Height() int
}

// State represents the current state of the application.
// Each state implementation handles its own update logic and rendering.
type State interface {
	// View renders the state with access to the context
	View(ctx Context) string

	// Update handles messages and returns the next state
	Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
