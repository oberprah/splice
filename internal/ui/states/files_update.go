package states

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/messages"
)

// Update handles messages for the files state
func (s *FilesState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.DiffLoadedMsg:
		// Handle diff loading result
		if msg.Err != nil {
			// For now, just stay in files state on error
			// In the future, we could show an error message
			return s, nil
		}

		// Calculate initial viewport position - scroll to first change
		viewportStart := 0
		if msg.Diff != nil && len(msg.Diff.ChangeIndices) > 0 {
			viewportStart = msg.Diff.ChangeIndices[0]
		}

		// Transition to diff state
		return &DiffState{
			Commit:                 msg.Commit,
			File:                   msg.File,
			Diff:                   msg.Diff,
			ViewportStart:          viewportStart,
			CurrentChangeIdx:       0,
			FilesCommit:            msg.FilesCommit,
			FilesFiles:             msg.FilesFiles,
			FilesCursor:            msg.FilesCursor,
			FilesViewportStart:     msg.FilesViewportStart,
			FilesListCommits:       msg.FilesListCommits,
			FilesListCursor:        msg.FilesListCursor,
			FilesListViewportStart: msg.FilesListViewportStart,
		}, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous list state
			return &LogState{
				Commits:       s.ListCommits,
				Cursor:        s.ListCursor,
				ViewportStart: s.ListViewportStart,
			}, nil

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "enter":
			// Load diff for selected file
			if len(s.Files) > 0 && s.Cursor < len(s.Files) {
				file := s.Files[s.Cursor]
				return s, s.loadDiff(file)
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
func (s *FilesState) updateViewport(height int) {
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
func (s *FilesState) loadDiff(file git.FileChange) tea.Cmd {
	commit := s.Commit
	files := s.Files
	cursor := s.Cursor
	viewportStart := s.ViewportStart
	listCommits := s.ListCommits
	listCursor := s.ListCursor
	listViewportStart := s.ListViewportStart

	return func() tea.Msg {
		// Fetch full file content and diff
		fullDiffResult, err := git.FetchFullFileDiff(commit.Hash, file)
		if err != nil {
			return messages.DiffLoadedMsg{
				Commit:                 commit,
				File:                   file,
				Err:                    err,
				FilesCommit:            commit,
				FilesFiles:             files,
				FilesCursor:            cursor,
				FilesViewportStart:     viewportStart,
				FilesListCommits:       listCommits,
				FilesListCursor:        listCursor,
				FilesListViewportStart: listViewportStart,
			}
		}

		// Parse the diff
		parsedDiff, err := diff.ParseUnifiedDiff(fullDiffResult.DiffOutput)
		if err != nil {
			return messages.DiffLoadedMsg{
				Commit:                 commit,
				File:                   file,
				Err:                    err,
				FilesCommit:            commit,
				FilesFiles:             files,
				FilesCursor:            cursor,
				FilesViewportStart:     viewportStart,
				FilesListCommits:       listCommits,
				FilesListCursor:        listCursor,
				FilesListViewportStart: listViewportStart,
			}
		}

		// Merge full file content with diff information
		fullFileDiff := diff.MergeFullFile(fullDiffResult.OldContent, fullDiffResult.NewContent, &parsedDiff)

		// Apply syntax highlighting to the diff
		diff.ApplySyntaxHighlighting(fullFileDiff, fullDiffResult.OldContent, fullDiffResult.NewContent, file.Path)

		return messages.DiffLoadedMsg{
			Commit:                 commit,
			File:                   file,
			Diff:                   fullFileDiff,
			Err:                    nil,
			FilesCommit:            commit,
			FilesFiles:             files,
			FilesCursor:            cursor,
			FilesViewportStart:     viewportStart,
			FilesListCommits:       listCommits,
			FilesListCursor:        listCursor,
			FilesListViewportStart: listViewportStart,
		}
	}
}

