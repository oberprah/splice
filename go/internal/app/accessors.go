package app

import (
	"time"

	"github.com/oberprah/splice/internal/core"
)

// Model implements core.Context interface.
// These methods provide states with access to terminal dimensions and data fetchers.

// Width returns the terminal width
func (m *Model) Width() int {
	return m.width
}

// Height returns the terminal height
func (m *Model) Height() int {
	return m.height
}

// FetchFileChanges returns the file changes fetcher function
func (m *Model) FetchFileChanges() core.FetchFileChangesFunc {
	return m.fetchFileChanges
}

// FetchFullFileDiff returns the full file diff fetcher function
func (m *Model) FetchFullFileDiff() core.FetchFullFileDiffFunc {
	return m.fetchFullFileDiff
}

// FetchFullFileDiffForSource returns the diff fetcher for a DiffSource
func (m *Model) FetchFullFileDiffForSource() core.FetchFullFileDiffForSourceFunc {
	return m.fetchFullFileDiffForSource
}

// Now returns the current time (for testing time-dependent formatting)
func (m *Model) Now() time.Time {
	return m.nowFunc()
}
