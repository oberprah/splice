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
	Diff   *diff.FullFileDiff

	// Viewport control
	ViewportStart    int
	CurrentChangeIdx int // Index into Diff.ChangeIndices for navigation

	// Preserved FilesState data for back navigation
	FilesCommit            git.GitCommit
	FilesFiles             []git.FileChange
	FilesCursor            int
	FilesViewportStart     int
	FilesListCommits       []git.GitCommit
	FilesListCursor        int
	FilesListViewportStart int
}
