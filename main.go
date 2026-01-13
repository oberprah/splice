package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/cli"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/states/directdiff"
	"github.com/oberprah/splice/internal/ui/states/loading"
)

func main() {
	// Parse command line arguments
	cmd, args := cli.ParseCommand(os.Args)

	var initialState core.State

	if cmd == "diff" {
		// Parse diff arguments (pure parsing, no git calls)
		diffArgs, err := cli.ParseDiffArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Validate that the diff has changes (requires git)
		if err := git.ValidateDiffHasChanges(diffArgs.RawSpec, diffArgs.UncommittedType); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Create DiffSource from parsed spec
		var diffSource core.DiffSource
		if diffArgs.IsCommitRange() {
			// Commit range - resolve refs to commits (requires git)
			commitRange, err := git.ResolveCommitRange(diffArgs.RawSpec)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			diffSource = commitRange
		} else {
			// Uncommitted changes
			diffSource = core.UncommittedChangesDiffSource{Type: *diffArgs.UncommittedType}
		}

		// Create DirectDiffLoadingState
		initialState = directdiff.New(diffSource)
	} else {
		// Default to log view
		initialState = loading.State{}
	}

	// Create the initial model with the appropriate state
	initialModel := app.NewModel(
		app.WithInitialState(initialState),
	)

	// Start the Bubbletea program with alternate screen (fullscreen mode)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
