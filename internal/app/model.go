package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/states/diff"
	stateserror "github.com/oberprah/splice/internal/ui/states/error"
	"github.com/oberprah/splice/internal/ui/states/files"
	"github.com/oberprah/splice/internal/ui/states/log"
)

// Model represents the application model using the state pattern
type Model struct {
	stack             []core.State // Navigation stack - previous states preserved exactly
	currentState      core.State
	firstPush         bool // True until first push occurs (for LoadingState special case)
	width             int
	height            int
	fetchCommits      FetchCommitsFunc
	fetchFileChanges  core.FetchFileChangesFunc
	fetchFullFileDiff core.FetchFullFileDiffFunc
	nowFunc           func() time.Time
}

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]git.GitCommit, error)

// Width returns the terminal width
func (m *Model) Width() int {
	return m.width
}

// Height returns the terminal height
func (m *Model) Height() int {
	return m.height
}

// FetchFileChanges returns the file changes fetcher function
func (m *Model) FetchFileChanges() core.FetchFileChangesFunc {
	return m.fetchFileChanges
}

// FetchFullFileDiff returns the full file diff fetcher function
func (m *Model) FetchFullFileDiff() core.FetchFullFileDiffFunc {
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
func WithFetchFileChanges(fn core.FetchFileChangesFunc) ModelOption {
	return func(m *Model) {
		m.fetchFileChanges = fn
	}
}

// WithFetchFullFileDiff allows injecting a custom full file diff fetcher for testing
func WithFetchFullFileDiff(fn core.FetchFullFileDiffFunc) ModelOption {
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
func WithInitialState(state core.State) ModelOption {
	return func(m *Model) {
		m.currentState = state
	}
}

// NewModel creates a new Model with initial loading state
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
		return core.CommitsLoadedMsg{Commits: commits, Err: err}
	}
}

// Update handles messages by delegating to the current state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle navigation messages at model level before delegating to state
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case core.PushLogScreenMsg:
		// Create LogState with initial preview loading for the first commit
		firstCommitHash := msg.Commits[0].Hash
		newState := &log.State{
			Commits:       msg.Commits,
			Cursor:        0,
			ViewportStart: 0,
			Preview:       log.PreviewLoading{ForHash: firstCommitHash},
			GraphLayout:   msg.GraphLayout,
		}
		if m.firstPush {
			m.firstPush = false
			m.currentState = newState
			return m, msg.InitCmd
		}
		m.stack = append(m.stack, m.currentState)
		m.currentState = newState
		return m, msg.InitCmd

	case core.PushFilesScreenMsg:
		newState := &files.State{
			Commit:        msg.Commit,
			Files:         msg.Files,
			Cursor:        0,
			ViewportStart: 0,
		}
		m.stack = append(m.stack, m.currentState)
		m.currentState = newState
		return m, nil

	case core.PushDiffScreenMsg:
		// Calculate initial viewport position - scroll to first change
		viewportStart := 0
		if msg.Diff != nil && len(msg.ChangeIndices) > 0 {
			viewportStart = msg.ChangeIndices[0]
		}
		newState := &diff.State{
			Commit:           msg.Commit,
			File:             msg.File,
			Diff:             msg.Diff,
			ChangeIndices:    msg.ChangeIndices,
			ViewportStart:    viewportStart,
			CurrentChangeIdx: 0,
		}
		m.stack = append(m.stack, m.currentState)
		m.currentState = newState
		return m, nil

	case core.PushErrorScreenMsg:
		newState := stateserror.State{Err: msg.Err}
		if m.firstPush {
			// Error during loading - don't push, just replace
			m.firstPush = false
			m.currentState = newState
			return m, nil
		}
		m.stack = append(m.stack, m.currentState)
		m.currentState = newState
		return m, nil

	case core.PopScreenMsg:
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
