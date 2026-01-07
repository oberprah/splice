package files

import (
	"github.com/oberprah/splice/internal/git"
)

// FilesState represents the state when displaying files changed in a commit
type State struct {
	Commit        git.GitCommit
	Files         []git.FileChange
	Cursor        int
	ViewportStart int
}

// New creates a new FilesState with cursor at the first file.
func New(commit git.GitCommit, files []git.FileChange) *State {
	return &State{
		Commit:        commit,
		Files:         files,
		Cursor:        0,
		ViewportStart: 0,
	}
}
