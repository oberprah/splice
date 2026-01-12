package diff

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
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

// getCurrentFileLineNumber maps the current viewport position to a file line number.
// It handles all alignment types and returns a 1-indexed line number suitable for
// opening in an editor. For RemovedAlignment (deleted lines), it searches forward
// to find the next alignment with a RightIdx, falling back to line 1 if none found.
func (s *State) getCurrentFileLineNumber() (int, error) {
	if s.Diff == nil {
		return 0, fmt.Errorf("no diff available")
	}

	if len(s.Diff.Alignments) == 0 {
		return 0, fmt.Errorf("diff has no alignments")
	}

	if s.ViewportStart >= len(s.Diff.Alignments) {
		return 0, fmt.Errorf("viewport position out of range")
	}

	alignment := s.Diff.Alignments[s.ViewportStart]

	switch a := alignment.(type) {
	case diff.UnchangedAlignment:
		return a.RightIdx + 1, nil
	case diff.ModifiedAlignment:
		return a.RightIdx + 1, nil
	case diff.AddedAlignment:
		return a.RightIdx + 1, nil
	case diff.RemovedAlignment:
		// RemovedAlignment has no RightIdx (deleted line doesn't exist in new file)
		// Search forward for the next alignment with a RightIdx
		for i := s.ViewportStart + 1; i < len(s.Diff.Alignments); i++ {
			switch next := s.Diff.Alignments[i].(type) {
			case diff.UnchangedAlignment:
				return next.RightIdx + 1, nil
			case diff.ModifiedAlignment:
				return next.RightIdx + 1, nil
			case diff.AddedAlignment:
				return next.RightIdx + 1, nil
			}
		}
		// No alignment with RightIdx found after the removed line - fall back to line 1
		return 1, nil
	default:
		return 0, fmt.Errorf("unknown alignment type")
	}
}
