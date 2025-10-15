package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// model represents the application state
type model struct {
	message string
}

// Init is called when the program starts
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard input
		switch msg.String() {
		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the UI
func (m model) View() string {
	return fmt.Sprintf("%s\n\nPress 'q' to quit.\n", m.message)
}

func main() {
	// Create the initial model
	initialModel := model{
		message: "Hello World from Splice!",
	}

	// Start the Bubbletea program
	p := tea.NewProgram(initialModel)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
