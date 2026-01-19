package diff

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
)

// State represents the state when viewing a file diff
type State struct {
	// File context for navigation
	Source    core.DiffSource
	Files     []core.FileChange // All files in current diff source
	FileIndex int               // Position of current file in Files

	// Current file's diff
	File core.FileChange
	Diff *diff.FileDiff

	// Viewport control
	ViewportStart   int
	CurrentBlockIdx int // Index of current change block for navigation
}

// New creates a new State with viewport positioned at the first change block.
func New(source core.DiffSource, files []core.FileChange, fileIndex int, file core.FileChange, d *diff.FileDiff) *State {
	viewportStart := 0
	currentBlockIdx := -1 // -1 means "not in a change block"

	// Find and position at the first change block
	if d != nil {
		lineOffset := 0
		for i, block := range d.Blocks {
			if _, isChange := block.(diff.ChangeBlock); isChange {
				viewportStart = lineOffset
				currentBlockIdx = i
				break
			}
			lineOffset += block.LineCount()
		}
	}

	return &State{
		Source:          source,
		Files:           files,
		FileIndex:       fileIndex,
		File:            file,
		Diff:            d,
		ViewportStart:   viewportStart,
		CurrentBlockIdx: currentBlockIdx,
	}
}
