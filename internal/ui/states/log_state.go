package states

import (
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/graph"
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
type LogState struct {
	Commits       []git.GitCommit
	Cursor        int
	ViewportStart int
	Preview       PreviewState
	GraphLayout   *graph.Layout // Computed graph layout for commits
}
