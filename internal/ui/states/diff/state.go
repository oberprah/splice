package diff

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
)

// DiffState represents the state when viewing a file diff
type State struct {
	// Current diff data
	CommitRange core.CommitRange
	File        core.FileChange
	Diff        *diff.AlignedFileDiff

	// Viewport control
	ViewportStart    int
	CurrentChangeIdx int   // Index into ChangeIndices for navigation
	ChangeIndices    []int // Indices of alignments that have changes (for navigation)
}

// New creates a new DiffState with viewport positioned at the first change.
func New(commitRange core.CommitRange, file core.FileChange, d *diff.AlignedFileDiff, changeIndices []int) *State {
	viewportStart := 0
	if d != nil && len(changeIndices) > 0 {
		viewportStart = changeIndices[0]
	}
	return &State{
		CommitRange:      commitRange,
		File:             file,
		Diff:             d,
		ChangeIndices:    changeIndices,
		ViewportStart:    viewportStart,
		CurrentChangeIdx: 0,
	}
}
