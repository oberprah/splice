package diff

import (
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/git"
)

// DiffState represents the state when viewing a file diff
type State struct {
	// Current diff data
	Commit git.GitCommit
	File   git.FileChange
	Diff   *diff.AlignedFileDiff

	// Viewport control
	ViewportStart    int
	CurrentChangeIdx int   // Index into ChangeIndices for navigation
	ChangeIndices    []int // Indices of alignments that have changes (for navigation)
}
