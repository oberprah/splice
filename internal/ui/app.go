package ui

import (
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/state/loading"

	tea "github.com/charmbracelet/bubbletea"
)

// NewModel creates a new Model with initial loading state
func NewModel() Model {
	return Model{
		currentState: loading.State{},
	}
}

// Init initializes the model and starts loading commits
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		commits, err := git.FetchCommits(500)
		return loading.CommitsLoadedMsg{Commits: commits, Err: err}
	}
}

// Update handles messages by delegating to the current state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize at model level (shared across all states)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to current state's update logic
	newState, cmd := m.currentState.Update(msg, &m)
	m.currentState = newState
	return m, cmd
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.currentState.View(&m)
}
