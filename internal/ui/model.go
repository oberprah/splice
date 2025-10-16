package ui

import (
	"github.com/oberprah/splice/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// ViewState represents the current state of the application
type ViewState int

const (
	LoadingView ViewState = iota
	ListView
	ErrorView
)

// Model represents the application state
type Model struct {
	// Data
	commits []git.GitCommit

	// Navigation
	cursor         int
	viewportStart  int
	viewportHeight int

	// View
	state  ViewState
	width  int
	height int

	// Status
	loading bool
	err     error
}

// NewModel creates a new Model with initial state
func NewModel() Model {
	return Model{
		state:          LoadingView,
		loading:        true,
		cursor:         0,
		viewportStart:  0,
		viewportHeight: 0,
		commits:        []git.GitCommit{},
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
