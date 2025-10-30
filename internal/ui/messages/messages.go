package messages

import "github.com/oberprah/splice/internal/git"

// FilesLoadedMsg is sent when files for a commit have been loaded
type FilesLoadedMsg struct {
	Commit git.GitCommit
	Files  []git.FileChange
	Err    error
	// Store original list state data to return to
	ListCommits       []git.GitCommit
	ListCursor        int
	ListViewportStart int
}
