package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/states/diff"
	stateserror "github.com/oberprah/splice/internal/ui/states/error"
	"github.com/oberprah/splice/internal/ui/states/files"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/states/log"
)

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]git.GitCommit, error)

// Model represents the application model using the state pattern.
// It implements tea.Model for Bubbletea and core.Context for states.
type Model struct {
	stack             []core.State // Navigation stack - current state is always stack[len-1]
	width             int
	height            int
	fetchCommits      FetchCommitsFunc
	fetchFileChanges  core.FetchFileChangesFunc
	fetchFullFileDiff core.FetchFullFileDiffFunc
	nowFunc           func() time.Time
}

// current returns the current state (top of stack).
// Returns nil if stack is empty (during initial loading).
func (m *Model) current() core.State {
	if len(m.stack) == 0 {
		return nil
	}
	return m.stack[len(m.stack)-1]
}

// NewModel creates a new Model with default configuration.
// Use functional options (WithFetchCommits, WithInitialState, etc.) to customize.
func NewModel(opts ...ModelOption) Model {
	m := Model{
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

// Update handles messages by delegating to the current state.
// Navigation messages (Push*ScreenMsg, PopScreenMsg) are handled at this level.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case core.PushLogScreenMsg:
		m.pushState(log.New(msg.Commits, msg.GraphLayout))
		return m, msg.InitCmd

	case core.PushFilesScreenMsg:
		m.pushState(files.New(msg.Range, msg.Files))
		return m, nil

	case core.PushDiffScreenMsg:
		m.pushState(diff.New(msg.Range, msg.File, msg.Diff, msg.ChangeIndices))
		return m, nil

	case core.PushErrorScreenMsg:
		m.pushState(stateserror.New(msg.Err))
		return m, nil

	case core.PopScreenMsg:
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
			return m, nil
		}
		return m, tea.Quit
	}

	// Delegate to current state
	newState, cmd := m.current().Update(msg, &m)
	m.stack[len(m.stack)-1] = newState
	return m, cmd
}

// pushState adds a new state to the navigation stack.
// LoadingState is transient - it gets replaced instead of stacked.
func (m *Model) pushState(newState core.State) {
	if _, isLoading := m.current().(loading.State); isLoading {
		m.stack[len(m.stack)-1] = newState
	} else {
		m.stack = append(m.stack, newState)
	}
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.current().View(&m).String()
}
