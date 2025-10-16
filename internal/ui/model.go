package ui

import (
	"github.com/oberprah/splice/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// Model represents the application state using the state pattern
type Model struct {
	state  State // Current state (LoadingState, ListState, or ErrorState)
	width  int   // Terminal width
	height int   // Terminal height
}

// NewModel creates a new Model with initial loading state
func NewModel() Model {
	return Model{
		state: LoadingState{},
	}
}

// commitsLoadedMsg is sent when commits have been loaded
type commitsLoadedMsg struct {
	commits []git.GitCommit
	err     error
}

// loadCommitsCmd loads commits from git in the background
func loadCommitsCmd() tea.Msg {
	commits, err := git.FetchCommits(500)
	return commitsLoadedMsg{commits: commits, err: err}
}

// Init initializes the model and starts loading commits
func (m Model) Init() tea.Cmd {
	return loadCommitsCmd
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
	newState, cmd := m.state.Update(msg, &m)
	m.state = newState
	return m, cmd
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.state.View(&m)
}
