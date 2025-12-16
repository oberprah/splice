package messages

import (
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
)

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

// DiffLoadedMsg is sent when diff content for a file has been loaded
type DiffLoadedMsg struct {
	Commit git.GitCommit
	File   git.FileChange
	Diff   *diff.FullFileDiff
	Err    error
	// Store original FilesState data to return to
	FilesCommit            git.GitCommit
	FilesFiles             []git.FileChange
	FilesCursor            int
	FilesViewportStart     int
	FilesListCommits       []git.GitCommit
	FilesListCursor        int
	FilesListViewportStart int
}
