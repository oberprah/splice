package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/states/diff"
	"github.com/oberprah/splice/internal/ui/states/directdiff"
	stateserror "github.com/oberprah/splice/internal/ui/states/error"
	"github.com/oberprah/splice/internal/ui/states/files"
	"github.com/oberprah/splice/internal/ui/states/loading"
	"github.com/oberprah/splice/internal/ui/states/log"
)

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]core.GitCommit, error)

// Model represents the application model using the state pattern.
// It implements tea.Model for Bubbletea and core.Context for states.
type Model struct {
	stack                     []core.State // Navigation stack - current state is always stack[len-1]
	width                     int
	height                    int
	fetchCommits              FetchCommitsFunc
	fetchFileChanges          core.FetchFileChangesFunc
	fetchFileChangesForSource core.FetchFileChangesForSourceFunc
	fetchFullFileDiff         core.FetchFullFileDiffFunc
	nowFunc                   func() time.Time
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
		fetchCommits: git.FetchCommits, // Default to real git command
		fetchFileChanges: func(commitRange core.CommitRange) ([]core.FileChange, error) {
			return git.FetchFileChanges(commitRange)
		},
		fetchFullFileDiff: func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
			return git.FetchFullFileDiff(commitRange, change)
		},
		nowFunc: time.Now, // Default to real time
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init initializes the model and starts the appropriate loading command.
// For LoadingState: fetches commits
// For DirectDiffLoadingState: fetches files for the diff source
func (m Model) Init() tea.Cmd {
	// Check if the initial state is DirectDiffLoadingState
	if state, ok := m.current().(directdiff.State); ok {
		// Use injected function if available (for testing), otherwise use default
		if m.fetchFileChangesForSource != nil {
			return func() tea.Msg {
				files, err := m.fetchFileChangesForSource(state.Source)
				return core.FilesLoadedMsg{
					Source: state.Source,
					Files:  files,
					Err:    err,
				}
			}
		}
		return directdiff.FetchFileChangesForSource(state.Source)
	}

	// Default: fetch commits (for LoadingState)
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
		m.pushState(files.New(msg.Source, msg.Files))
		return m, nil

	case core.PushDiffScreenMsg:
		m.pushState(diff.New(msg.Source, msg.File, msg.Diff, msg.ChangeIndices))
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
// LoadingState and DirectDiffLoadingState are transient - they get replaced instead of stacked.
func (m *Model) pushState(newState core.State) {
	current := m.current()
	_, isLoading := current.(loading.State)
	_, isDirectDiffLoading := current.(directdiff.State)

	if isLoading || isDirectDiffLoading {
		m.stack[len(m.stack)-1] = newState
	} else {
		m.stack = append(m.stack, newState)
	}
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.current().View(&m).String()
}
