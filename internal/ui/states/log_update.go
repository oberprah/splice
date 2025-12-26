package states

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/messages"
)

// Update handles messages in list view state
func (s LogState) Update(msg tea.Msg, ctx Context) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.FilesLoadedMsg:
		// Handle file loading result
		if msg.Err != nil {
			// For now, just stay in list state on error
			// In the future, we could show an error message
			return s, nil
		}

		// Transition to files state, passing data to restore list state
		return &FilesState{
			Commit:            msg.Commit,
			Files:             msg.Files,
			Cursor:            0,
			ViewportStart:     0,
			ListCommits:       msg.ListCommits,
			ListCursor:        msg.ListCursor,
			ListViewportStart: msg.ListViewportStart,
		}, nil

	case messages.FilesPreviewLoadedMsg:
		// Handle preview loading result
		// Check if the response is for the current cursor commit (stale response detection)
		if len(s.Commits) == 0 || s.Commits[s.Cursor].Hash != msg.ForHash {
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

		case "enter":
			// Load files for the selected commit
			if len(s.Commits) > 0 {
				selectedCommit := s.Commits[s.Cursor]
				return s, func() tea.Msg {
					fileChanges, err := git.FetchFileChanges(selectedCommit.Hash)
					return messages.FilesLoadedMsg{
						Commit:            selectedCommit,
						Files:             fileChanges,
						Err:               err,
						ListCommits:       s.Commits,
						ListCursor:        s.Cursor,
						ListViewportStart: s.ViewportStart,
					}
				}
			}
			return s, nil

		case "j", "down":
			if s.Cursor < len(s.Commits)-1 {
				s.Cursor++
				s.updateViewport(ctx.Height())
				// Trigger preview loading for the new cursor position
				commitHash := s.Commits[s.Cursor].Hash
				s.Preview = PreviewLoading{ForHash: commitHash}
				return s, loadPreview(commitHash)
			}
			return s, nil

		case "k", "up":
			if s.Cursor > 0 {
				s.Cursor--
				s.updateViewport(ctx.Height())
				// Trigger preview loading for the new cursor position
				commitHash := s.Commits[s.Cursor].Hash
				s.Preview = PreviewLoading{ForHash: commitHash}
				return s, loadPreview(commitHash)
			}
			return s, nil

		case "g":
			s.Cursor = 0
			s.ViewportStart = 0
			// Trigger preview loading for the top commit
			if len(s.Commits) > 0 {
				commitHash := s.Commits[s.Cursor].Hash
				s.Preview = PreviewLoading{ForHash: commitHash}
				return s, loadPreview(commitHash)
			}
			return s, nil

		case "G":
			s.Cursor = len(s.Commits) - 1
			s.updateViewport(ctx.Height())
			// Trigger preview loading for the bottom commit
			if len(s.Commits) > 0 {
				commitHash := s.Commits[s.Cursor].Hash
				s.Preview = PreviewLoading{ForHash: commitHash}
				return s, loadPreview(commitHash)
			}
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *LogState) updateViewport(height int) {
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

// loadPreview returns a command that loads file changes for the preview panel
func loadPreview(commitHash string) tea.Cmd {
	return func() tea.Msg {
		files, err := git.FetchFileChanges(commitHash)
		return messages.FilesPreviewLoadedMsg{
			ForHash: commitHash,
			Files:   files,
			Err:     err,
		}
	}
}
