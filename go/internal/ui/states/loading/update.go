package loading

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/ui/states/log"
)

// Update handles messages during loading
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.CommitsLoadedMsg:
		// Handle errors
		if msg.Err != nil {
			return s, func() tea.Msg {
				return core.PushErrorScreenMsg{Err: msg.Err}
			}
		}

		// Treat empty repositories as an error
		if len(msg.Commits) == 0 {
			return s, func() tea.Msg {
				return core.PushErrorScreenMsg{
					Err: fmt.Errorf("no commits found in repository"),
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

		// Return a command that produces PushLogScreenMsg to navigate to LogState
		// Include InitCmd to load preview for the first commit
		firstCommit := msg.Commits[0]
		initialRange := core.NewCommitRange(firstCommit, firstCommit, 1)
		return s, func() tea.Msg {
			return core.PushLogScreenMsg{
				Commits:     msg.Commits,
				GraphLayout: layout,
				InitCmd:     log.LoadPreview(initialRange, ctx.FetchFileChanges()),
			}
		}
	}

	return s, nil
}
