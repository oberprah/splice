package log

import (
	"github.com/oberprah/splice/internal/core"
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
	Cursor        core.CursorState
	ViewportStart int
	Preview       PreviewState
	GraphLayout   *graph.Layout // Computed graph layout for commits
}

// New creates a new LogState with the given commits and graph layout.
// Cursor starts at position 0 with preview loading for the first commit.
func New(commits []git.GitCommit, layout *graph.Layout) *State {
	// Initial state is always a single commit at position 0
	initialHash := commits[0].Hash
	return &State{
		Commits:       commits,
		Cursor:        core.CursorNormal{Pos: 0},
		ViewportStart: 0,
		Preview:       PreviewLoading{ForHash: initialHash},
		GraphLayout:   layout,
	}
}

// CursorPosition returns the current cursor position.
func (s *State) CursorPosition() int {
	return s.Cursor.Position()
}

// IsVisualMode returns true if in visual selection mode.
func (s *State) IsVisualMode() bool {
	_, ok := s.Cursor.(core.CursorVisual)
	return ok
}

// GetSelectedRange returns a CommitRange for the current selection.
// For normal mode, returns a single-commit range.
// For visual mode, returns a range from anchor to cursor (normalized to chronological order).
func (s *State) GetSelectedRange() core.CommitRange {
	min, max := core.SelectionRange(s.Cursor)

	// In git log, index 0 is the newest commit, so:
	// - max index = older commit (Start)
	// - min index = newer commit (End)
	startCommit := s.Commits[max]
	endCommit := s.Commits[min]
	count := max - min + 1

	return core.NewCommitRange(startCommit, endCommit, count)
}
