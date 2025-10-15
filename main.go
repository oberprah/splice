package main

import (
	"fmt"
	"os"

	"github.com/oberprah/splice/git"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create the initial model
	initialModel := Model{
		state:          LoadingView,
		loading:        true,
		cursor:         0,
		viewportStart:  0,
		viewportHeight: 0,
		commits:        []git.GitCommit{},
	}

	// Start the Bubbletea program with alternate screen (fullscreen mode)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
