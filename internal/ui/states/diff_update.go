package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the diff state
func (s *DiffState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous files state
			return &FilesState{
				Commit:            s.FilesCommit,
				Files:             s.FilesFiles,
				Cursor:            s.FilesCursor,
				ViewportStart:     s.FilesViewportStart,
				ListCommits:       s.FilesListCommits,
				ListCursor:        s.FilesListCursor,
				ListViewportStart: s.FilesListViewportStart,
			}, nil

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "j", "down":
			// Scroll down
			if len(s.Diff.Lines) > 0 {
				maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
				if s.ViewportStart < maxViewportStart {
					s.ViewportStart++
				}
			}
			return s, nil

		case "k", "up":
			// Scroll up
			if s.ViewportStart > 0 {
				s.ViewportStart--
			}
			return s, nil

		case "ctrl+d":
			// Scroll down half page
			if len(s.Diff.Lines) > 0 {
				headerLines := 2
				availableHeight := max(ctx.Height()-headerLines, 1)
				halfPage := availableHeight / 2
				maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
				s.ViewportStart = min(s.ViewportStart+halfPage, maxViewportStart)
			}
			return s, nil

		case "ctrl+u":
			// Scroll up half page
			headerLines := 2
			availableHeight := max(ctx.Height()-headerLines, 1)
			halfPage := availableHeight / 2
			s.ViewportStart = max(s.ViewportStart-halfPage, 0)
			return s, nil

		case "g":
			// Jump to top
			s.ViewportStart = 0
			return s, nil

		case "G":
			// Jump to bottom
			if len(s.Diff.Lines) > 0 {
				s.ViewportStart = s.calculateMaxViewportStart(ctx.Height())
			}
			return s, nil
		}
	}

	return s, nil
}

// calculateMaxViewportStart returns the maximum valid viewport start position
func (s *DiffState) calculateMaxViewportStart(height int) int {
	// Account for header lines
	headerLines := 2 // header + separator
	availableHeight := max(height-headerLines, 1)

	maxStart := len(s.Diff.Lines) - availableHeight
	if maxStart < 0 {
		maxStart = 0
	}
	return maxStart
}
