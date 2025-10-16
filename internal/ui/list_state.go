package ui

import (
	"strings"

	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"

	tea "github.com/charmbracelet/bubbletea"
)

// ListState represents the state when displaying a list of commits
type ListState struct {
	commits       []git.GitCommit
	cursor        int
	viewportStart int
}

// View renders the list of commits
func (s ListState) View(m *Model) string {
	var b strings.Builder

	// Calculate the end of the viewport
	viewportEnd := min(s.viewportStart+m.height, len(s.commits))

	// Render only visible commits
	for i := s.viewportStart; i < viewportEnd; i++ {
		commit := s.commits[i]
		line := s.formatCommitLine(commit, i == s.cursor, m.width)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// Update handles messages in list view state
func (s ListState) Update(msg tea.Msg, m *Model) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return s, tea.Quit

		case "j", "down":
			if s.cursor < len(s.commits)-1 {
				s.cursor++
				s.updateViewport(m.height)
			}
			return s, nil

		case "k", "up":
			if s.cursor > 0 {
				s.cursor--
				s.updateViewport(m.height)
			}
			return s, nil

		case "g":
			s.cursor = 0
			s.viewportStart = 0
			return s, nil

		case "G":
			s.cursor = len(s.commits) - 1
			s.updateViewport(m.height)
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *ListState) updateViewport(height int) {
	// Scroll down if cursor is below viewport
	if s.cursor >= s.viewportStart+height {
		s.viewportStart = s.cursor - height + 1
	}

	// Scroll up if cursor is above viewport
	if s.cursor < s.viewportStart {
		s.viewportStart = s.cursor
	}

	// Ensure viewport doesn't go negative
	if s.viewportStart < 0 {
		s.viewportStart = 0
	}
}

// formatCommitLine formats a single commit line with proper styling
func (s ListState) formatCommitLine(commit git.GitCommit, isSelected bool, width int) string {
	// Format: hash message - author (time ago)
	// Example: a4c3a8a Fix memory leak in parser - John Doe (4 min ago)

	// Determine available width (accounting for selection indicator and spacing)
	availableWidth := width
	if availableWidth <= 0 {
		availableWidth = 80 // Default fallback
	}

	// Selection indicator (2 chars: "> " or "  ")
	selectionIndicator := "  "
	if isSelected {
		selectionIndicator = "> "
	}

	// Format the base components
	hash := format.ToShortHash(commit.Hash)    // 7 chars
	message := commit.Message                  // Variable
	separator := " - "                         // 3 chars
	author := commit.Author                    // Variable
	timePrefix := " "                          // 1 char
	time := format.ToRelativeTime(commit.Date) // Variable

	// Calculate required space for fixed elements
	fixedWidth := len(selectionIndicator) + len(hash) + 1 + len(separator) + len(timePrefix) + len(time)

	// Calculate remaining space for message and author
	remainingWidth := max(availableWidth-fixedWidth,
		// Terminal too narrow, show minimal format
		10)

	// Truncate message and author to fit remaining space
	messageMaxWidth := remainingWidth * 2 / 3 // Give 2/3 to message
	authorMaxWidth := remainingWidth - messageMaxWidth

	if len(message) > messageMaxWidth && messageMaxWidth > 3 {
		message = message[:messageMaxWidth-3] + "..."
	}

	if len(author) > authorMaxWidth && authorMaxWidth > 3 {
		author = author[:authorMaxWidth-3] + "..."
	}

	// Build the line with styling
	var line strings.Builder

	line.WriteString(selectionIndicator)

	if isSelected {
		// For selected lines, use bold styles
		line.WriteString(styles.SelectedHashStyle.Render(hash))
		line.WriteString(" ")
		line.WriteString(styles.SelectedMessageStyle.Render(message))
		line.WriteString(separator)
		line.WriteString(styles.SelectedAuthorStyle.Render(author))
		line.WriteString(timePrefix)
		line.WriteString(styles.SelectedTimeStyle.Render(time))
	} else {
		// For unselected lines, apply regular styles
		line.WriteString(styles.HashStyle.Render(hash))
		line.WriteString(" ")
		line.WriteString(styles.MessageStyle.Render(message))
		line.WriteString(separator)
		line.WriteString(styles.AuthorStyle.Render(author))
		line.WriteString(timePrefix)
		line.WriteString(styles.TimeStyle.Render(time))
	}

	return line.String()
}
