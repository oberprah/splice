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
