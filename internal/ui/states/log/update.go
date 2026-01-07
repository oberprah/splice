package log

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/core"
)

// Update handles messages in list view state
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.FilesLoadedMsg:
		// Handle file loading result
		if msg.Err != nil {
			// For now, just stay in list state on error
			// In the future, we could show an error message
			return s, nil
		}

		// Transition to files state using navigation pattern
		return s, func() tea.Msg {
			return core.PushFilesScreenMsg{
				Range: msg.Range,
				Files: msg.Files,
			}
		}

	case core.FilesPreviewLoadedMsg:
		// Handle preview loading result
		// Check if the response is for the current selection (stale response detection)
		currentRangeHash := getRangeHash(s.GetSelectedRange())
		if len(s.Commits) == 0 || currentRangeHash != msg.ForHash {
			// Response is stale (user navigated away), discard it
			return s, nil
		}

		// Update preview state based on whether there was an error
		if msg.Err != nil {
			s.Preview = PreviewError{
				ForHash: msg.ForHash,
				Err:     msg.Err,
			}
		} else {
			s.Preview = PreviewLoaded{
				ForHash: msg.ForHash,
				Files:   msg.Files,
			}
		}
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "v":
			// Toggle visual mode
			switch cursor := s.Cursor.(type) {
			case core.CursorNormal:
				s.Cursor = core.CursorVisual{Pos: cursor.Pos, Anchor: cursor.Pos}
			case core.CursorVisual:
				s.Cursor = core.CursorNormal{Pos: cursor.Pos}
			}
			return s, nil

		case "esc":
			// Exit visual mode if active
			if visual, ok := s.Cursor.(core.CursorVisual); ok {
				s.Cursor = core.CursorNormal{Pos: visual.Pos}
			}
			return s, nil

		case "enter":
			// Load files for the selected commit or range
			if len(s.Commits) > 0 {
				commitRange := s.GetSelectedRange()
				fetchFileChanges := ctx.FetchFileChanges()
				return s, func() tea.Msg {
					// For range diff, we need parent of start..end
					var fromHash string
					if commitRange.IsSingleCommit() {
						fromHash = commitRange.End.Hash + "^"
					} else {
						fromHash = commitRange.Start.Hash + "^"
					}

					fileChanges, err := fetchFileChanges(fromHash, commitRange.End.Hash)
					return core.FilesLoadedMsg{
						Range: commitRange,
						Files: fileChanges,
						Err:   err,
					}
				}
			}
			return s, nil

		case "j", "down":
			pos := s.CursorPosition()
			if pos < len(s.Commits)-1 {
				newPos := pos + 1
				switch cursor := s.Cursor.(type) {
				case core.CursorNormal:
					s.Cursor = core.CursorNormal{Pos: newPos}
				case core.CursorVisual:
					s.Cursor = core.CursorVisual{Pos: newPos, Anchor: cursor.Anchor}
				}
				s.updateViewport(ctx.Height())
				// Trigger preview loading for the new selection
				commitRange := s.GetSelectedRange()
				rangeHash := getRangeHash(commitRange)
				s.Preview = PreviewLoading{ForHash: rangeHash}
				return s, LoadPreview(commitRange, ctx.FetchFileChanges())
			}
			return s, nil

		case "k", "up":
			pos := s.CursorPosition()
			if pos > 0 {
				newPos := pos - 1
				switch cursor := s.Cursor.(type) {
				case core.CursorNormal:
					s.Cursor = core.CursorNormal{Pos: newPos}
				case core.CursorVisual:
					s.Cursor = core.CursorVisual{Pos: newPos, Anchor: cursor.Anchor}
				}
				s.updateViewport(ctx.Height())
				// Trigger preview loading for the new selection
				commitRange := s.GetSelectedRange()
				rangeHash := getRangeHash(commitRange)
				s.Preview = PreviewLoading{ForHash: rangeHash}
				return s, LoadPreview(commitRange, ctx.FetchFileChanges())
			}
			return s, nil

		case "g":
			newPos := 0
			switch cursor := s.Cursor.(type) {
			case core.CursorNormal:
				s.Cursor = core.CursorNormal{Pos: newPos}
			case core.CursorVisual:
				s.Cursor = core.CursorVisual{Pos: newPos, Anchor: cursor.Anchor}
			}
			s.ViewportStart = 0
			// Trigger preview loading for the new selection
			if len(s.Commits) > 0 {
				commitRange := s.GetSelectedRange()
				rangeHash := getRangeHash(commitRange)
				s.Preview = PreviewLoading{ForHash: rangeHash}
				return s, LoadPreview(commitRange, ctx.FetchFileChanges())
			}
			return s, nil

		case "G":
			newPos := len(s.Commits) - 1
			switch cursor := s.Cursor.(type) {
			case core.CursorNormal:
				s.Cursor = core.CursorNormal{Pos: newPos}
			case core.CursorVisual:
				s.Cursor = core.CursorVisual{Pos: newPos, Anchor: cursor.Anchor}
			}
			s.updateViewport(ctx.Height())
			// Trigger preview loading for the new selection
			if len(s.Commits) > 0 {
				commitRange := s.GetSelectedRange()
				rangeHash := getRangeHash(commitRange)
				s.Preview = PreviewLoading{ForHash: rangeHash}
				return s, LoadPreview(commitRange, ctx.FetchFileChanges())
			}
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *State) updateViewport(height int) {
	pos := s.CursorPosition()

	// Scroll down if cursor is below viewport
	if pos >= s.ViewportStart+height {
		s.ViewportStart = pos - height + 1
	}

	// Scroll up if cursor is above viewport
	if pos < s.ViewportStart {
		s.ViewportStart = pos
	}

	// Ensure viewport doesn't go negative
	if s.ViewportStart < 0 {
		s.ViewportStart = 0
	}
}

// LoadPreview creates a command to load file changes for a commit or range (for preview in log view)
func LoadPreview(commitRange core.CommitRange, fetchFileChanges core.FetchFileChangesFunc) tea.Cmd {
	return func() tea.Msg {
		// Calculate from/to hashes using the same logic as Enter key handler
		var fromHash string
		if commitRange.IsSingleCommit() {
			fromHash = commitRange.End.Hash + "^"
		} else {
			fromHash = commitRange.Start.Hash + "^"
		}
		toHash := commitRange.End.Hash

		files, err := fetchFileChanges(fromHash, toHash)
		return core.FilesPreviewLoadedMsg{
			ForHash: getRangeHash(commitRange),
			Files:   files,
			Err:     err,
		}
	}
}

// getRangeHash returns a string representation of the commit range for tracking preview state.
// For single commits, returns just the hash. For ranges, returns "start..end".
func getRangeHash(commitRange core.CommitRange) string {
	if commitRange.IsSingleCommit() {
		return commitRange.End.Hash
	}
	return commitRange.Start.Hash + ".." + commitRange.End.Hash
}
