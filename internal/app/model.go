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

// FetchCommitsFunc is a function type for fetching git commits
type FetchCommitsFunc func(limit int) ([]git.GitCommit, error)

// Model represents the application model using the state pattern.
// It implements tea.Model for Bubbletea and core.Context for states.
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

// NewModel creates a new Model with default configuration.
// Use functional options (WithFetchCommits, WithInitialState, etc.) to customize.
func NewModel(opts ...ModelOption) Model {
	m := Model{
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

// Update handles messages by delegating to the current state.
// Navigation messages (Push*ScreenMsg, PopScreenMsg) are handled at this level.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case core.PushLogScreenMsg:
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
			m.firstPush = false
			m.currentState = newState
			return m, nil
		}
		m.stack = append(m.stack, m.currentState)
		m.currentState = newState
		return m, nil

	case core.PopScreenMsg:
		if len(m.stack) > 0 {
			m.currentState = m.stack[len(m.stack)-1]
			m.stack = m.stack[:len(m.stack)-1]
			return m, nil
		}
		return m, tea.Quit
	}

	// Delegate to current state
	newState, cmd := m.currentState.Update(msg, &m)
	m.currentState = newState
	return m, cmd
}

// View renders the UI by delegating to the current state
func (m Model) View() string {
	return m.currentState.View(&m).String()
}
