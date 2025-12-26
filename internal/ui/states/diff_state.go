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
	Diff   *diff.AlignedFileDiff

	// Viewport control
	ViewportStart    int
	CurrentChangeIdx int // Index into ChangeIndices for navigation
	ChangeIndices    []int // Indices of alignments that have changes (for navigation)

	// Preserved FilesState data for back navigation
	FilesCommit            git.GitCommit
	FilesFiles             []git.FileChange
	FilesCursor            int
	FilesViewportStart     int
	FilesListCommits       []git.GitCommit
	FilesListCursor        int
	FilesListViewportStart int
}
