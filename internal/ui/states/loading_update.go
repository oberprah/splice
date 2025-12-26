package states

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

// CommitsLoadedMsg is sent when commits have been loaded
type CommitsLoadedMsg struct {
	Commits []git.GitCommit
	Err     error
}

// Update handles messages during loading
func (s LoadingState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case CommitsLoadedMsg:
		// Handle errors
		if msg.Err != nil {
			return &ErrorState{Err: msg.Err}, nil
		}

		// Treat empty repositories as an error
		if len(msg.Commits) == 0 {
			return &ErrorState{Err: fmt.Errorf("no commits found in repository")}, nil
		}

		// Successfully loaded commits - transition to list view
		return &LogState{
			Commits:       msg.Commits,
			Cursor:        0,
			ViewportStart: 0,
			Preview:       PreviewNone{},
		}, nil
	}

	return s, nil
}
