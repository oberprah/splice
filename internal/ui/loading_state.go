package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadingState represents the state when commits are being loaded
type LoadingState struct{}

// View renders the loading message
func (s LoadingState) View(m *Model) string {
	return "  Loading commits...\n"
}

// Update handles messages during loading
func (s LoadingState) Update(msg tea.Msg, m *Model) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case commitsLoadedMsg:
		// Handle errors
		if msg.err != nil {
			return ErrorState{err: msg.err}, nil
		}

		// Treat empty repositories as an error
		if len(msg.commits) == 0 {
			return ErrorState{err: fmt.Errorf("no commits found in repository")}, nil
		}

		// Successfully loaded commits - transition to list view
		return ListState{
			commits:       msg.commits,
			cursor:        0,
			viewportStart: 0,
		}, nil
	}

	return s, nil
}
