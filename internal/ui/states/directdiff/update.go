package directdiff

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

// Update handles messages during direct diff loading
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.FilesLoadedMsg:
		// Handle errors
		if msg.Err != nil {
			return s, func() tea.Msg {
				return core.PushErrorScreenMsg{Err: msg.Err}
			}
		}

		// Treat empty file lists as an error
		if len(msg.Files) == 0 {
			return s, func() tea.Msg {
				return core.PushErrorScreenMsg{
					Err: fmt.Errorf("no changes found"),
				}
			}
		}

		// Successfully loaded files - transition to files view
		return s, func() tea.Msg {
			return core.PushFilesScreenMsg{
				Source: msg.Source,
				Files:  msg.Files,
			}
		}
	}

	return s, nil
}

// FetchFileChangesForSource creates a command to fetch file changes based on the DiffSource type.
// This is exported so it can be used when pushing DirectDiffLoadingState from main.go.
func FetchFileChangesForSource(source core.DiffSource) tea.Cmd {
	return func() tea.Msg {
		var files []core.FileChange
		var err error

		// Type switch on source to determine which git function to call
		switch src := source.(type) {
		case core.CommitRangeDiffSource:
			// For commit ranges, use existing FetchFileChanges with CommitRange
			files, err = git.FetchFileChanges(src.ToCommitRange())

		case core.UncommittedChangesDiffSource:
			// For uncommitted changes, type switch on Type field
			switch src.Type {
			case core.UncommittedTypeUnstaged:
				files, err = git.FetchUnstagedFileChanges()
			case core.UncommittedTypeStaged:
				files, err = git.FetchStagedFileChanges()
			case core.UncommittedTypeAll:
				files, err = git.FetchAllUncommittedFileChanges()
			default:
				err = fmt.Errorf("unknown uncommitted type: %v", src.Type)
			}

		default:
			err = fmt.Errorf("unknown diff source type: %T", source)
		}

		return core.FilesLoadedMsg{
			Source: source,
			Files:  files,
			Err:    err,
		}
	}
}
