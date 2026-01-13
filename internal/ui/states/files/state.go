package files

import (
	"github.com/oberprah/splice/internal/core"
)

// FilesState represents the state when displaying files changed in a commit
type State struct {
	Source        core.DiffSource
	Files         []core.FileChange
	Cursor        int
	ViewportStart int
}

// New creates a new FilesState with cursor at the first file.
func New(source core.DiffSource, files []core.FileChange) *State {
	return &State{
		Source:        source,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
}
