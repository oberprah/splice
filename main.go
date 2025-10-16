package main

import (
	"fmt"
	"os"

	"github.com/oberprah/splice/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create the initial model
	initialModel := ui.NewModel()

	// Start the Bubbletea program with alternate screen (fullscreen mode)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
