package files

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
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
			// Exit or go back depending on ExitOnPop flag
			if s.ExitOnPop {
				return s, tea.Quit
			}
			// Go back to the previous state using navigation pattern
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
	// Extract CommitRange from Source (for now, we only support CommitRangeDiffSource)
	// This will be updated in later steps when we add support for uncommitted changes
	var commitRange core.CommitRange
	switch src := s.Source.(type) {
	case core.CommitRangeDiffSource:
		commitRange = src.ToCommitRange()
	case core.UncommittedChangesDiffSource:
		// TODO: Handle uncommitted changes in a future step
		// For now, return an error
		return func() tea.Msg {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    fmt.Errorf("uncommitted changes not yet supported"),
			}
		}
	default:
		return func() tea.Msg {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    fmt.Errorf("unknown diff source type"),
			}
		}
	}

	return func() tea.Msg {
		// Fetch full file content and diff
		fullDiffResult, err := fetchFullFileDiff(commitRange, file)
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
