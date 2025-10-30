package states

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages in list view state
func (s ListState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "j", "down":
			if s.Cursor < len(s.Commits)-1 {
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
			s.Cursor = len(s.Commits) - 1
			s.updateViewport(ctx.Height())
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *ListState) updateViewport(height int) {
	// Scroll down if cursor is below viewport
	if s.Cursor >= s.ViewportStart+height {
		s.ViewportStart = s.Cursor - height + 1
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
