package files

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/git"
)

// Update handles messages for the files state
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.DiffLoadedMsg:
		// Handle diff loading result
		if msg.Err != nil {
			// For now, just stay in files state on error
			// In the future, we could show an error message
			return s, nil
		}

		// Return command that produces PushDiffScreenMsg to navigate to DiffState
		return s, func() tea.Msg {
			return core.PushDiffScreenMsg{
				Source:        msg.Source,
				File:          msg.File,
				Diff:          msg.Diff,
				ChangeIndices: msg.ChangeIndices,
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous state using navigation pattern
			// (app.Model handles quitting when stack is empty)
			return s, func() tea.Msg {
				return core.PopScreenMsg{}
			}

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "enter":
			// Load diff for selected file
			if len(s.Files) > 0 && s.Cursor < len(s.Files) {
				file := s.Files[s.Cursor]
				return s, s.loadDiff(file, ctx.FetchFullFileDiff())
			}
			return s, nil

		case "j", "down":
			if len(s.Files) > 0 && s.Cursor < len(s.Files)-1 {
				s.Cursor++
				s.updateViewport(ctx.Height())
			}
			return s, nil

		case "k", "up":
			if s.Cursor > 0 {
				s.Cursor--
				s.updateViewport(ctx.Height())
			}
			return s, nil

		case "g":
			s.Cursor = 0
			s.ViewportStart = 0
			return s, nil

		case "G":
			if len(s.Files) > 0 {
				s.Cursor = len(s.Files) - 1
				s.updateViewport(ctx.Height())
			}
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *State) updateViewport(height int) {
	// Account for header lines when calculating available height
	headerLines := 2 // commit info + separator
	availableHeight := max(height-headerLines, 1)

	// Scroll down if cursor is below viewport
	if s.Cursor >= s.ViewportStart+availableHeight {
		s.ViewportStart = s.Cursor - availableHeight + 1
	}

	// Scroll up if cursor is above viewport
	if s.Cursor < s.ViewportStart {
		s.ViewportStart = s.Cursor
	}

	// Ensure viewport doesn't go negative
	if s.ViewportStart < 0 {
		s.ViewportStart = 0
	}
}

// loadDiff creates a command to fetch and parse the diff for a file
func (s *State) loadDiff(file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffFunc) tea.Cmd {
	return func() tea.Msg {
		// Fetch full file content and diff based on DiffSource type
		fullDiffResult, err := fetchFileDiffForSource(s.Source, file, fetchFullFileDiff)
		if err != nil {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    err,
			}
		}

		// Build the complete aligned diff with change indices
		alignedDiff, changeIndices, err := diff.BuildAlignedFileDiff(
			file.Path,
			fullDiffResult.OldContent,
			fullDiffResult.NewContent,
			fullDiffResult.DiffOutput,
		)
		if err != nil {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    err,
			}
		}

		return core.DiffLoadedMsg{
			Source:        s.Source,
			File:          file,
			Diff:          alignedDiff,
			ChangeIndices: changeIndices,
			Err:           nil,
		}
	}
}

// fetchFileDiffForSource fetches the full file diff based on the DiffSource type.
// Uses type switch to call the appropriate git function for each source type.
func fetchFileDiffForSource(source core.DiffSource, file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffFunc) (*core.FullFileDiffResult, error) {
	switch src := source.(type) {
	case core.CommitRangeDiffSource:
		// For commit ranges, use the injected fetchFullFileDiff function
		commitRange := src.ToCommitRange()
		return fetchFullFileDiff(commitRange, file)

	case core.UncommittedChangesDiffSource:
		// For uncommitted changes, type switch on Type field to call appropriate git function
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			return git.FetchUnstagedFileDiff(file)
		case core.UncommittedTypeStaged:
			return git.FetchStagedFileDiff(file)
		case core.UncommittedTypeAll:
			return git.FetchAllUncommittedFileDiff(file)
		default:
			return nil, fmt.Errorf("unknown uncommitted type: %v", src.Type)
		}

	default:
		return nil, fmt.Errorf("unknown diff source type: %T", source)
	}
}
