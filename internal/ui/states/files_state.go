package states

import (
	"github.com/oberprah/splice/internal/git"
)

// FilesState represents the state when displaying files changed in a commit
type FilesState struct {
	Commit        git.GitCommit
	Files         []git.FileChange
	Cursor        int
	ViewportStart int
	// Data to restore list state when going back
	ListCommits       []git.GitCommit
	ListCursor        int
	ListViewportStart int
}
