package ui

import (
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/state/loading"

	tea "github.com/charmbracelet/bubbletea"
)

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]git.GitCommit, error)

// ModelOption is a functional option for configuring a Model
type ModelOption func(*Model)

// WithFetchCommits allows injecting a custom commit fetcher for testing
func WithFetchCommits(fn FetchCommitsFunc) ModelOption {
	return func(m *Model) {
		m.fetchCommits = fn
	}
}

// NewModel creates a new Model with initial loading state
func NewModel(opts ...ModelOption) Model {
	m := Model{
		currentState: loading.State{},
		fetchCommits: git.FetchCommits, // Default to real git command
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init initializes the model and starts loading commits
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		commits, err := m.fetchCommits(500)
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
