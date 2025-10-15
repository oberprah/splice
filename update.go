package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle keyboard input
		switch msg.String() {
		case "q", "ctrl+c":
			// Quit the application
			return m, tea.Quit

		case "j", "down":
			// Move cursor down
			if m.state == ListView && len(m.commits) > 0 {
				if m.cursor < len(m.commits)-1 {
					m.cursor++
					m.updateViewport()
				}
			}

		case "k", "up":
			// Move cursor up
			if m.state == ListView && len(m.commits) > 0 {
				if m.cursor > 0 {
					m.cursor--
					m.updateViewport()
				}
			}

		case "g":
			// Jump to top
			if m.state == ListView && len(m.commits) > 0 {
				m.cursor = 0
				m.viewportStart = 0
			}

		case "G":
			// Jump to bottom
			if m.state == ListView && len(m.commits) > 0 {
				m.cursor = len(m.commits) - 1
				m.updateViewport()
			}
		}

	case tea.WindowSizeMsg:
		// Handle terminal resize
		m.width = msg.Width
		m.height = msg.Height
		m.viewportHeight = msg.Height
		m.updateViewport()

	case commitsLoadedMsg:
		// Handle commits loaded
		if msg.err != nil {
			m.state = ErrorView
			m.err = msg.err
			m.loading = false
		} else {
			m.commits = msg.commits
			m.state = ListView
			m.loading = false
			if len(m.commits) > 0 {
				m.cursor = 0
			}
		}
	}

	return m, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (m *Model) updateViewport() {
	// Scroll down if cursor is below viewport
	if m.cursor >= m.viewportStart+m.viewportHeight {
		m.viewportStart = m.cursor - m.viewportHeight + 1
	}

	// Scroll up if cursor is above viewport
	if m.cursor < m.viewportStart {
		m.viewportStart = m.cursor
	}

	// Ensure viewport doesn't go negative
	if m.viewportStart < 0 {
		m.viewportStart = 0
	}
}
