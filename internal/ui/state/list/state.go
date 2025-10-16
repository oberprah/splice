package list

import "github.com/oberprah/splice/internal/git"

// State represents the state when displaying a list of commits
type State struct {
	Commits       []git.GitCommit
	Cursor        int
	ViewportStart int
}
