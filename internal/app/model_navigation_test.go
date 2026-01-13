package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/states/loading"
)

// TestNavigationStack tests that the navigation stack works correctly.
// With the stack-only model, current state is always stack[len-1].
// LoadingState is transient - it gets replaced on first push, not stacked.
func TestNavigationStack(t *testing.T) {
	// Create a model with LoadingState as initial state
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewModel(
		WithInitialState(loading.State{}),
		WithFetchCommits(func(int) ([]core.GitCommit, error) {
			return []core.GitCommit{{Hash: "abc123", Message: "Test", Author: "Test", Date: fixedTime}}, nil
		}),
		WithFetchFileChanges(func(commitRange core.CommitRange) ([]core.FileChange, error) {
			return []core.FileChange{{Path: "test.go", Status: "M"}}, nil
		}),
		WithFetchFullFileDiff(func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
			return &core.FullFileDiffResult{DiffOutput: "test"}, nil
		}),
	)

	// Set window size
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = m2.(Model)

	// Initial stack should have LoadingState
	if len(m.stack) != 1 {
		t.Errorf("Initial stack should have 1 item (LoadingState), got %d", len(m.stack))
	}

	// Push LogScreen - LoadingState is transient, so it gets replaced (stack stays at 1)
	m2, cmd := m.Update(core.PushLogScreenMsg{
		Commits:     []core.GitCommit{{Hash: "abc123"}},
		GraphLayout: nil,
	})
	m = m2.(Model)

	if len(m.stack) != 1 {
		t.Errorf("After first push (replacing LoadingState), stack should have 1 item, got %d", len(m.stack))
	}
	if cmd != nil {
		// Execute the command if any
		msg := cmd()
		m2, _ = m.Update(msg)
		m = m2.(Model)
	}

	// Push FilesScreen - normal push, adds to stack
	testCommit := core.GitCommit{Hash: "abc123"}
	commitRange := core.NewSingleCommitRange(testCommit)
	m2, _ = m.Update(core.PushFilesScreenMsg{
		Source: commitRange.ToDiffSource(),
		Files:  []core.FileChange{{Path: "test.go"}},
	})
	m = m2.(Model)

	if len(m.stack) != 2 {
		t.Errorf("After pushing FilesScreen, stack should have 2 items, got %d", len(m.stack))
	}

	// Push DiffScreen - normal push, adds to stack
	m2, _ = m.Update(core.PushDiffScreenMsg{
		Source: commitRange.ToDiffSource(),
		File:   core.FileChange{Path: "test.go"},
	})
	m = m2.(Model)

	if len(m.stack) != 3 {
		t.Errorf("After pushing DiffScreen, stack should have 3 items, got %d", len(m.stack))
	}

	// Pop back to FilesScreen
	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)

	if len(m.stack) != 2 {
		t.Errorf("After popping once, stack should have 2 items, got %d", len(m.stack))
	}

	// Pop back to LogScreen
	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)

	if len(m.stack) != 1 {
		t.Errorf("After popping twice, stack should have 1 item (LogState), got %d", len(m.stack))
	}
}

// TestNavigationWithCommitRangeDiffSource tests navigation stack with CommitRangeDiffSource
func TestNavigationWithCommitRangeDiffSource(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewModel(
		WithInitialState(loading.State{}),
		WithFetchCommits(func(int) ([]core.GitCommit, error) {
			return []core.GitCommit{{Hash: "abc123", Message: "Test", Author: "Test", Date: fixedTime}}, nil
		}),
		WithFetchFileChanges(func(commitRange core.CommitRange) ([]core.FileChange, error) {
			return []core.FileChange{{Path: "test.go", Status: "M"}}, nil
		}),
		WithFetchFullFileDiff(func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
			return &core.FullFileDiffResult{DiffOutput: "test"}, nil
		}),
	)

	// Push LogScreen
	m2, _ := m.Update(core.PushLogScreenMsg{
		Commits:     []core.GitCommit{{Hash: "abc123"}},
		GraphLayout: nil,
	})
	m = m2.(Model)

	// Push FilesScreen with CommitRangeDiffSource
	testCommit := core.GitCommit{Hash: "abc123"}
	commitRange := core.NewSingleCommitRange(testCommit)
	diffSource := commitRange.ToDiffSource()

	m2, _ = m.Update(core.PushFilesScreenMsg{
		Source: diffSource,
		Files:  []core.FileChange{{Path: "test.go", Status: "M"}},
	})
	m = m2.(Model)

	if len(m.stack) != 2 {
		t.Errorf("After pushing FilesScreen with CommitRangeDiffSource, stack should have 2 items, got %d", len(m.stack))
	}

	// Push DiffScreen with CommitRangeDiffSource
	m2, _ = m.Update(core.PushDiffScreenMsg{
		Source: diffSource,
		File:   core.FileChange{Path: "test.go", Status: "M"},
	})
	m = m2.(Model)

	if len(m.stack) != 3 {
		t.Errorf("After pushing DiffScreen with CommitRangeDiffSource, stack should have 3 items, got %d", len(m.stack))
	}

	// Pop back through the stack
	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)
	if len(m.stack) != 2 {
		t.Errorf("After popping once, stack should have 2 items, got %d", len(m.stack))
	}

	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)
	if len(m.stack) != 1 {
		t.Errorf("After popping twice, stack should have 1 item, got %d", len(m.stack))
	}
}

// TestNavigationWithUncommittedChangesDiffSource tests navigation stack with UncommittedChangesDiffSource
func TestNavigationWithUncommittedChangesDiffSource(t *testing.T) {
	tests := []struct {
		name            string
		uncommittedType core.UncommittedType
	}{
		{"unstaged", core.UncommittedTypeUnstaged},
		{"staged", core.UncommittedTypeStaged},
		{"all", core.UncommittedTypeAll},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
			m := NewModel(
				WithInitialState(loading.State{}),
				WithFetchCommits(func(int) ([]core.GitCommit, error) {
					return []core.GitCommit{{Hash: "abc123", Message: "Test", Author: "Test", Date: fixedTime}}, nil
				}),
				WithFetchFileChanges(func(commitRange core.CommitRange) ([]core.FileChange, error) {
					return []core.FileChange{{Path: "test.go", Status: "M"}}, nil
				}),
				WithFetchFullFileDiff(func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
					return &core.FullFileDiffResult{DiffOutput: "test"}, nil
				}),
			)

			// Push LogScreen
			m2, _ := m.Update(core.PushLogScreenMsg{
				Commits:     []core.GitCommit{{Hash: "abc123"}},
				GraphLayout: nil,
			})
			m = m2.(Model)

			// Push FilesScreen with UncommittedChangesDiffSource
			diffSource := core.UncommittedChangesDiffSource{Type: tt.uncommittedType}

			m2, _ = m.Update(core.PushFilesScreenMsg{
				Source: diffSource,
				Files:  []core.FileChange{{Path: "test.go", Status: "M"}},
			})
			m = m2.(Model)

			if len(m.stack) != 2 {
				t.Errorf("After pushing FilesScreen with UncommittedChangesDiffSource, stack should have 2 items, got %d", len(m.stack))
			}

			// Push DiffScreen with UncommittedChangesDiffSource
			m2, _ = m.Update(core.PushDiffScreenMsg{
				Source: diffSource,
				File:   core.FileChange{Path: "test.go", Status: "M"},
			})
			m = m2.(Model)

			if len(m.stack) != 3 {
				t.Errorf("After pushing DiffScreen with UncommittedChangesDiffSource, stack should have 3 items, got %d", len(m.stack))
			}

			// Pop back through the stack
			m2, _ = m.Update(core.PopScreenMsg{})
			m = m2.(Model)
			if len(m.stack) != 2 {
				t.Errorf("After popping once, stack should have 2 items, got %d", len(m.stack))
			}

			m2, _ = m.Update(core.PopScreenMsg{})
			m = m2.(Model)
			if len(m.stack) != 1 {
				t.Errorf("After popping twice, stack should have 1 item, got %d", len(m.stack))
			}
		})
	}
}

// TestNavigationWithMixedDiffSources tests navigation with both diff source types
func TestNavigationWithMixedDiffSources(t *testing.T) {
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewModel(
		WithInitialState(loading.State{}),
		WithFetchCommits(func(int) ([]core.GitCommit, error) {
			return []core.GitCommit{{Hash: "abc123", Message: "Test", Author: "Test", Date: fixedTime}}, nil
		}),
		WithFetchFileChanges(func(commitRange core.CommitRange) ([]core.FileChange, error) {
			return []core.FileChange{{Path: "test.go", Status: "M"}}, nil
		}),
		WithFetchFullFileDiff(func(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
			return &core.FullFileDiffResult{DiffOutput: "test"}, nil
		}),
	)

	// Push LogScreen
	m2, _ := m.Update(core.PushLogScreenMsg{
		Commits:     []core.GitCommit{{Hash: "abc123"}},
		GraphLayout: nil,
	})
	m = m2.(Model)

	// Push FilesScreen with CommitRangeDiffSource
	testCommit := core.GitCommit{Hash: "abc123"}
	commitRange := core.NewSingleCommitRange(testCommit)
	commitDiffSource := commitRange.ToDiffSource()

	m2, _ = m.Update(core.PushFilesScreenMsg{
		Source: commitDiffSource,
		Files:  []core.FileChange{{Path: "test1.go", Status: "M"}},
	})
	m = m2.(Model)

	if len(m.stack) != 2 {
		t.Errorf("After pushing FilesScreen with CommitRangeDiffSource, stack should have 2 items, got %d", len(m.stack))
	}

	// Push FilesScreen with UncommittedChangesDiffSource
	uncommittedDiffSource := core.UncommittedChangesDiffSource{Type: core.UncommittedTypeUnstaged}

	m2, _ = m.Update(core.PushFilesScreenMsg{
		Source: uncommittedDiffSource,
		Files:  []core.FileChange{{Path: "test2.go", Status: "M"}},
	})
	m = m2.(Model)

	if len(m.stack) != 3 {
		t.Errorf("After pushing FilesScreen with UncommittedChangesDiffSource, stack should have 3 items, got %d", len(m.stack))
	}

	// Pop back through the stack
	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)
	if len(m.stack) != 2 {
		t.Errorf("After popping once, stack should have 2 items, got %d", len(m.stack))
	}

	m2, _ = m.Update(core.PopScreenMsg{})
	m = m2.(Model)
	if len(m.stack) != 1 {
		t.Errorf("After popping twice, stack should have 1 item, got %d", len(m.stack))
	}
}
