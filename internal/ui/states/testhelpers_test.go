package states

import "github.com/oberprah/splice/internal/git"

// mockContext is a test helper that implements the Context interface
type mockContext struct {
	width  int
	height int
}

func (m mockContext) Width() int {
	return m.width
}

func (m mockContext) Height() int {
	return m.height
}

func (m mockContext) FetchFileChanges() FetchFileChangesFunc {
	// Return a mock function that returns empty file changes
	return func(commitHash string) ([]git.FileChange, error) {
		return []git.FileChange{}, nil
	}
}
