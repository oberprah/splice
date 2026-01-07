package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]git.GitCommit, error)

// Model represents the application model using the state pattern
type Model struct {
	stack             []State // Navigation stack - previous states preserved exactly
	currentState      State
	firstPush         bool // True until first push occurs (for LoadingState special case)
	width             int
	height            int
	fetchCommits      FetchCommitsFunc
	fetchFileChanges  FetchFileChangesFunc
	fetchFullFileDiff FetchFullFileDiffFunc
	nowFunc           func() time.Time
}

// Width returns the terminal width
func (m *Model) Width() int {
	return m.width
}

// Height returns the terminal height
func (m *Model) Height() int {
	return m.height
}

// FetchFileChanges returns the file changes fetcher function
func (m *Model) FetchFileChanges() FetchFileChangesFunc {
	return m.fetchFileChanges
}

// FetchFullFileDiff returns the full file diff fetcher function
func (m *Model) FetchFullFileDiff() FetchFullFileDiffFunc {
	return m.fetchFullFileDiff
}

// Now returns the current time (for testing time-dependent formatting)
func (m *Model) Now() time.Time {
	return m.nowFunc()
}

// ModelOption is a functional option for configuring a Model
type ModelOption func(*Model)

// WithFetchCommits allows injecting a custom commit fetcher for testing
func WithFetchCommits(fn FetchCommitsFunc) ModelOption {
	return func(m *Model) {
		m.fetchCommits = fn
	}
}

// WithFetchFileChanges allows injecting a custom file changes fetcher for testing
func WithFetchFileChanges(fn FetchFileChangesFunc) ModelOption {
	return func(m *Model) {
		m.fetchFileChanges = fn
	}
}

// WithFetchFullFileDiff allows injecting a custom full file diff fetcher for testing
func WithFetchFullFileDiff(fn FetchFullFileDiffFunc) ModelOption {
	return func(m *Model) {
		m.fetchFullFileDiff = fn
	}
}

// WithNow allows injecting a custom time function for deterministic testing
func WithNow(fn func() time.Time) ModelOption {
	return func(m *Model) {
		m.nowFunc = fn
	}
}

// WithInitialState allows setting a custom initial state for testing
func WithInitialState(state State) ModelOption {
	return func(m *Model) {
		m.currentState = state
	}
}

// NewModel creates a new Model with initial loading state
// Note: This function requires states to be registered via RegisterStateFactory
// The loading state must be imported in main.go to trigger its init() registration
func NewModel(opts ...ModelOption) Model {
	m := Model{
		// currentState will be set by WithInitialState option or remain nil
		// main.go should provide the initial loading state
		firstPush:         true,                  // First push will not add to stack (LoadingState)
		fetchCommits:      git.FetchCommits,      // Default to real git command
		fetchFileChanges:  git.FetchFileChanges,  // Default to real git command
		fetchFullFileDiff: git.FetchFullFileDiff, // Default to real git command
		nowFunc:           time.Now,              // Default to real time
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
		return CommitsLoadedMsg{Commits: commits, Err: err}
	}
}

// Update handles messages by delegating to the current state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages at model level before delegating to state
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case PushScreenMsg:
		// Special case: LoadingState is the initial state and should not be pushed
		// When LoadingState transitions to LogState, replace currentState directly
		if m.firstPush {
			// First transition from LoadingState - replace, don't push
			m.firstPush = false
			m.currentState = CreateState(msg.Screen, msg.Data)
			return m, msg.InitCmd
		}
		// Normal push: save current state and create new one
		m.stack = append(m.stack, m.currentState)
		m.currentState = CreateState(msg.Screen, msg.Data)
		return m, msg.InitCmd

	case PopScreenMsg:
		// Pop the previous state from the stack
		if len(m.stack) > 0 {
			m.currentState = m.stack[len(m.stack)-1]
			m.stack = m.stack[:len(m.stack)-1]
			return m, nil
		}
		// If stack is empty, quit the application
		// This happens when popping from LoadingState or an error that occurred during loading
		return m, tea.Quit
	}

	// Delegate to current state's update logic
	newState, cmd := m.currentState.Update(msg, &m)
	m.currentState = newState
	return m, cmd
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.currentState.View(&m).String()
}
