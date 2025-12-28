package states

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

// FetchFileChangesFunc is a function type for fetching file changes for a commit
type FetchFileChangesFunc func(commitHash string) ([]git.FileChange, error)

// FetchFullFileDiffFunc is a function type for fetching full file diff content
type FetchFullFileDiffFunc func(commitHash string, change git.FileChange) (*git.FullFileDiffResult, error)

// Context is the interface that states use to access what they need from the model
type Context interface {
	Width() int
	Height() int
	FetchFileChanges() FetchFileChangesFunc
	FetchFullFileDiff() FetchFullFileDiffFunc
}

// State represents the current state of the application.
// Each state implementation handles its own update logic and rendering.
type State interface {
	// View renders the state with access to the context
	View(ctx Context) string

	// Update handles messages and returns the next state
	Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
