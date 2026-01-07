package log

import (
	"github.com/oberprah/splice/internal/domain/graph"
	"github.com/oberprah/splice/internal/git"
)

// PreviewState is a sum type representing the state of the preview panel.
// Use type assertion to determine which variant is present.
type PreviewState interface {
	isPreviewState()
}

// PreviewNone indicates no preview is being shown
type PreviewNone struct{}

func (PreviewNone) isPreviewState() {}

// PreviewLoading indicates a preview is being loaded for the given commit hash
type PreviewLoading struct {
	ForHash string
}

func (PreviewLoading) isPreviewState() {}

// PreviewLoaded indicates the preview has been successfully loaded
type PreviewLoaded struct {
	ForHash string
	Files   []git.FileChange
}

func (PreviewLoaded) isPreviewState() {}

// PreviewError indicates an error occurred while loading the preview
type PreviewError struct {
	ForHash string
	Err     error
}

func (PreviewError) isPreviewState() {}

// LogState represents the state when displaying the commit log
type State struct {
	Commits       []git.GitCommit
	Cursor        int
	ViewportStart int
	Preview       PreviewState
	GraphLayout   *graph.Layout // Computed graph layout for commits
}

// New creates a new LogState with the given commits and graph layout.
// Cursor starts at position 0 with preview loading for the first commit.
func New(commits []git.GitCommit, layout *graph.Layout) *State {
	return &State{
		Commits:       commits,
		Cursor:        0,
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: commits[0].Hash},
		GraphLayout:   layout,
	}
}
