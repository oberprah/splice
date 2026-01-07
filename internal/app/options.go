package app

import (
	"time"

	"github.com/oberprah/splice/internal/core"
)

// ModelOption is a functional option for configuring a Model
type ModelOption func(*Model)

// WithFetchCommits allows injecting a custom commit fetcher for testing
func WithFetchCommits(fn FetchCommitsFunc) ModelOption {
	return func(m *Model) {
		m.fetchCommits = fn
	}
}

// WithFetchFileChanges allows injecting a custom file changes fetcher for testing
func WithFetchFileChanges(fn core.FetchFileChangesFunc) ModelOption {
	return func(m *Model) {
		m.fetchFileChanges = fn
	}
}

// WithFetchFullFileDiff allows injecting a custom full file diff fetcher for testing
func WithFetchFullFileDiff(fn core.FetchFullFileDiffFunc) ModelOption {
	return func(m *Model) {
		m.fetchFullFileDiff = fn
	}
}

// WithNow allows injecting a custom time function for deterministic testing
func WithNow(fn func() time.Time) ModelOption {
	return func(m *Model) {
		m.nowFunc = fn
	}
}

// WithInitialState allows setting a custom initial state for testing
func WithInitialState(state core.State) ModelOption {
	return func(m *Model) {
		m.stack = []core.State{state}
	}
}
