package states

import (
	"github.com/oberprah/splice/internal/diff"
	"github.com/oberprah/splice/internal/git"
)

// DiffState represents the state when viewing a file diff
type DiffState struct {
	// Current diff data
	Commit git.GitCommit
	File   git.FileChange
	Diff   diff.FileDiff

	// Viewport control
	ViewportStart int

	// Preserved FilesState data for back navigation
	FilesCommit            git.GitCommit
	FilesFiles             []git.FileChange
	FilesCursor            int
	FilesViewportStart     int
	FilesListCommits       []git.GitCommit
	FilesListCursor        int
	FilesListViewportStart int
}
