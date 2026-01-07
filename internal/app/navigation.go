package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/git"
)

// Screen represents different screens in the application
type Screen int

const (
	LogScreen Screen = iota
	FilesScreen
	DiffScreen
	ErrorScreen
)

// PushScreenMsg signals that a new screen should be pushed onto the navigation stack
type PushScreenMsg struct {
	Screen  Screen
	Data    any     // Screen-specific data
	InitCmd tea.Cmd // Optional initialization command to run after creating the state
}

// PopScreenMsg signals that the current screen should be popped from the navigation stack
type PopScreenMsg struct{}

// LogScreenData contains data needed to create a LogState
type LogScreenData struct {
	Commits     []git.GitCommit
	GraphLayout *graph.Layout
}

// FilesScreenData contains data needed to create a FilesState
type FilesScreenData struct {
	Commit git.GitCommit
	Files  []git.FileChange
}

// DiffScreenData contains data needed to create a DiffState
type DiffScreenData struct {
	Commit        git.GitCommit
	File          git.FileChange
	Diff          *diff.AlignedFileDiff
	ChangeIndices []int
}

// ErrorScreenData contains data needed to create an ErrorState
type ErrorScreenData struct {
	Err error
}
