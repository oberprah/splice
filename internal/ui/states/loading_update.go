package states

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/graph"
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

		// Load preview for the first commit (at cursor position 0)
		firstCommitHash := msg.Commits[0].Hash
		return &LogState{
			Commits:       msg.Commits,
			Cursor:        0,
			ViewportStart: 0,
			Preview:       PreviewLoading{ForHash: firstCommitHash},
			GraphLayout:   layout,
		}, loadPreview(firstCommitHash, ctx.FetchFileChanges())
	}

	return s, nil
}
