package diff

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
)

// Update handles messages for the diff state
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous state using navigation pattern
			return s, func() tea.Msg {
				return core.PopScreenMsg{}
			}

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "j", "down":
			// Scroll down
			if s.Diff != nil && len(s.Diff.Alignments) > 0 {
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
			if s.Diff != nil && len(s.Diff.Alignments) > 0 {
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
			if s.Diff != nil && len(s.Diff.Alignments) > 0 {
				s.ViewportStart = s.calculateMaxViewportStart(ctx.Height())
			}
			return s, nil

		case "n":
			// Jump to next change
			s.jumpToNextChange(ctx.Height())
			return s, nil

		case "N":
			// Jump to previous change
			s.jumpToPreviousChange(ctx.Height())
			return s, nil
		}
	}

	return s, nil
}

// jumpToNextChange scrolls to the next change after the current viewport
func (s *State) jumpToNextChange(height int) {
	if s.Diff == nil || len(s.ChangeIndices) == 0 {
		return
	}

	// Find the next change index after the current viewport position
	for i, changeIdx := range s.ChangeIndices {
		if changeIdx > s.ViewportStart {
			s.CurrentChangeIdx = i
			s.ViewportStart = changeIdx
			// Clamp to max viewport
			maxViewport := s.calculateMaxViewportStart(height)
			if s.ViewportStart > maxViewport {
				s.ViewportStart = maxViewport
			}
			return
		}
	}

	// If no change found after current position, optionally wrap to first
	// For now, stay at current position (no wrap)
}

// jumpToPreviousChange scrolls to the previous change before the current viewport
func (s *State) jumpToPreviousChange(height int) {
	if s.Diff == nil || len(s.ChangeIndices) == 0 {
		return
	}

	// Find the previous change index before the current viewport position
	for i := len(s.ChangeIndices) - 1; i >= 0; i-- {
		changeIdx := s.ChangeIndices[i]
		if changeIdx < s.ViewportStart {
			s.CurrentChangeIdx = i
			s.ViewportStart = changeIdx
			return
		}
	}

	// If no change found before current position, optionally wrap to last
	// For now, stay at current position (no wrap)
}

// calculateMaxViewportStart returns the maximum valid viewport start position
func (s *State) calculateMaxViewportStart(height int) int {
	if s.Diff == nil {
		return 0
	}

	// Account for header lines
	headerLines := 2 // header + separator
	availableHeight := max(height-headerLines, 1)

	maxStart := len(s.Diff.Alignments) - availableHeight
	if maxStart < 0 {
		maxStart = 0
	}
	return maxStart
}
