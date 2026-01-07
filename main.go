package main

import (
	"fmt"
	"os"

	"github.com/oberprah/splice/internal/app"
	_ "github.com/oberprah/splice/internal/ui/states" // Trigger init() for state factory registration
	"github.com/oberprah/splice/internal/ui/states/loading"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	// Create the initial model with LoadingState
	initialModel := app.NewModel(
		app.WithInitialState(loading.State{}),
	)

	// Start the Bubbletea program with alternate screen (fullscreen mode)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
