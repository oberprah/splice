package ui

import "github.com/oberprah/splice/internal/ui/states"

// Model represents the application model using the state pattern
type Model struct {
	currentState      states.State
	width             int
	height            int
	fetchCommits      FetchCommitsFunc
	fetchFileChanges  states.FetchFileChangesFunc
	fetchFullFileDiff states.FetchFullFileDiffFunc
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
func (m *Model) FetchFileChanges() states.FetchFileChangesFunc {
	return m.fetchFileChanges
}

// FetchFullFileDiff returns the full file diff fetcher function
func (m *Model) FetchFullFileDiff() states.FetchFullFileDiffFunc {
	return m.fetchFullFileDiff
}
