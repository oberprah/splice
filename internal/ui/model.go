package ui

import "github.com/oberprah/splice/internal/ui/state"

// Model represents the application model using the state pattern
type Model struct {
	currentState state.State
	width        int
	height       int
}

// Width returns the terminal width
func (m *Model) Width() int {
	return m.width
}

// Height returns the terminal height
func (m *Model) Height() int {
	return m.height
}
