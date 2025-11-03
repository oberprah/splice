package states

import "github.com/oberprah/splice/internal/git"

// LogState represents the state when displaying the commit log
type LogState struct {
	Commits       []git.GitCommit
	Cursor        int
	ViewportStart int
}
