package core

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/git"
)

// Navigation messages - each screen has its own typed message for compile-time safety

// PushLogScreenMsg signals navigation to the log screen
type PushLogScreenMsg struct {
	Commits     []git.GitCommit
	GraphLayout *graph.Layout
	InitCmd     tea.Cmd // Optional initialization command to run after creating the state
}

// PushFilesScreenMsg signals navigation to the files screen
type PushFilesScreenMsg struct {
	Range CommitRange
	Files []git.FileChange
}

// PushDiffScreenMsg signals navigation to the diff screen
type PushDiffScreenMsg struct {
	Range         CommitRange
	File          git.FileChange
	Diff          *diff.AlignedFileDiff
	ChangeIndices []int
}

// PushErrorScreenMsg signals navigation to the error screen
type PushErrorScreenMsg struct {
	Err error
}

// PopScreenMsg signals that the current screen should be popped from the navigation stack
type PopScreenMsg struct{}
