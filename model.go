package main

import (
	"github.com/oberprah/splice/git"

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
