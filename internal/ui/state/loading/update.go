package loading

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/state"
	errorstate "github.com/oberprah/splice/internal/ui/state/error"
	"github.com/oberprah/splice/internal/ui/state/list"
)

// CommitsLoadedMsg is sent when commits have been loaded
type CommitsLoadedMsg struct {
	Commits []git.GitCommit
	Err     error
}

// Update handles messages during loading
func (s State) Update(msg tea.Msg, ctx state.Context) (state.State, tea.Cmd) {
	switch msg := msg.(type) {
	case CommitsLoadedMsg:
		// Handle errors
		if msg.Err != nil {
			return errorstate.State{Err: msg.Err}, nil
		}

		// Treat empty repositories as an error
		if len(msg.Commits) == 0 {
			return errorstate.State{Err: fmt.Errorf("no commits found in repository")}, nil
		}

		// Successfully loaded commits - transition to list view
		return list.State{
			Commits:       msg.Commits,
			Cursor:        0,
			ViewportStart: 0,
		}, nil
	}

	return s, nil
}
