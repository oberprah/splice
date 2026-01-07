package loading

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/domain/graph"
)

// Update handles messages during loading
func (s State) Update(msg tea.Msg, ctx app.Context) (app.State, tea.Cmd) {
	switch msg := msg.(type) {
	case app.CommitsLoadedMsg:
		// Handle errors
		if msg.Err != nil {
			return s, func() tea.Msg {
				return app.PushScreenMsg{
					Screen: app.ErrorScreen,
					Data: app.ErrorScreenData{
						Err: msg.Err,
					},
				}
			}
		}

		// Treat empty repositories as an error
		if len(msg.Commits) == 0 {
			return s, func() tea.Msg {
				return app.PushScreenMsg{
					Screen: app.ErrorScreen,
					Data: app.ErrorScreenData{
						Err: fmt.Errorf("no commits found in repository"),
					},
				}
			}
		}

		// Successfully loaded commits - transition to list view
		// Convert GitCommits to graph.Commits for layout computation
		graphCommits := make([]graph.Commit, len(msg.Commits))
		for i, commit := range msg.Commits {
			graphCommits[i] = graph.Commit{
				Hash:    commit.Hash,
				Parents: commit.ParentHashes,
			}
		}

		// Compute graph layout
		layout := graph.ComputeLayout(graphCommits)

		// Return a command that produces PushScreenMsg to navigate to LogState
		// The log state factory in register.go will handle the initial preview loading
		return s, func() tea.Msg {
			return app.PushScreenMsg{
				Screen: app.LogScreen,
				Data: app.LogScreenData{
					Commits:     msg.Commits,
					GraphLayout: layout,
				},
			}
		}
	}

	return s, nil
}
