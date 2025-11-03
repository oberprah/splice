package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages for the files state
func (s *FilesState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
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

