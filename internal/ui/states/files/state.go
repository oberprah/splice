package files

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git"
)

// FilesState represents the state when displaying files changed in a commit
type State struct {
	Range         core.CommitRange
	Files         []git.FileChange
	Cursor        int
	ViewportStart int
}

// New creates a new FilesState with cursor at the first file.
func New(commitRange core.CommitRange, files []git.FileChange) *State {
	return &State{
		Range:         commitRange,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
}
