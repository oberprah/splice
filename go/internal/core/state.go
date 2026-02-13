package core

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// FetchFileChangesFunc is a function type for fetching file changes for a commit range
type FetchFileChangesFunc func(commitRange CommitRange) ([]FileChange, error)

// FetchFileChangesForSourceFunc is a function type for fetching file changes for a DiffSource
type FetchFileChangesForSourceFunc func(source DiffSource) ([]FileChange, error)

// FetchFullFileDiffFunc is a function type for fetching full file diff content
type FetchFullFileDiffFunc func(commitRange CommitRange, change FileChange) (*FullFileDiffResult, error)

// FetchFullFileDiffForSourceFunc fetches full file diff content for a DiffSource
type FetchFullFileDiffForSourceFunc func(source DiffSource, change FileChange) (*FullFileDiffResult, error)

// FullFileDiffResult contains the full file content before and after a change
type FullFileDiffResult struct {
	OldContent string // Content of the file before the change (empty for new files)
	NewContent string // Content of the file after the change (empty for deleted files)
	DiffOutput string // Raw unified diff output
	OldPath    string // Path of the file before the change (for renames)
	NewPath    string // Path of the file after the change
}

// Context is the interface that states use to access what they need from the model
type Context interface {
	Width() int
	Height() int
	FetchFileChanges() FetchFileChangesFunc
	FetchFullFileDiff() FetchFullFileDiffFunc
	FetchFullFileDiffForSource() FetchFullFileDiffForSourceFunc
	Now() time.Time
}

// ViewRenderer is an interface that can render to a string
// This interface is satisfied by the ViewBuilder type in ui/components
type ViewRenderer interface {
	String() string
}

// State represents the current state of the application.
// Each state implementation handles its own update logic and rendering.
type State interface {
	// View renders the state with access to the context
	View(ctx Context) ViewRenderer

	// Update handles messages and returns the next state
	Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
