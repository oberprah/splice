package directdiff

import (
	"github.com/oberprah/splice/internal/core"
)

// State represents the loading state when fetching file changes for a DiffSource.
// This state is transient and immediately fetches files, then transitions to FilesState.
type State struct {
	Source core.DiffSource
}

// New creates a new DirectDiffLoadingState for the given diff source.
func New(source core.DiffSource) State {
	return State{Source: source}
}
