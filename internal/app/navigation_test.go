package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/oberprah/splice/internal/git"
)

// TestNavigationStack tests that the navigation stack works correctly
func TestNavigationStack(t *testing.T) {
	// Register mock factories for each screen type
	RegisterStateFactory(LogScreen, func(data any) State {
		return mockState{}
	})
	RegisterStateFactory(FilesScreen, func(data any) State {
		return mockState{}
	})
	RegisterStateFactory(DiffScreen, func(data any) State {
		return mockState{}
	})

	// Create a model with mocks
	fixedTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewModel(
		WithInitialState(mockState{}),
		WithFetchCommits(func(int) ([]git.GitCommit, error) {
			return []git.GitCommit{{Hash: "abc123", Message: "Test", Author: "Test", Date: fixedTime}}, nil
		}),
		WithFetchFileChanges(func(string) ([]git.FileChange, error) {
			return []git.FileChange{{Path: "test.go", Status: "M"}}, nil
		}),
		WithFetchFullFileDiff(func(string, git.FileChange) (*git.FullFileDiffResult, error) {
			return &git.FullFileDiffResult{DiffOutput: "test"}, nil
		}),
	)

	// Set window size
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = m2.(Model)

	// Initial stack should be empty
	if len(m.stack) != 0 {
		t.Errorf("Initial stack should be empty, got %d", len(m.stack))
	}

	// Push LogScreen
	m2, cmd := m.Update(PushScreenMsg{
		Screen: LogScreen,
		Data: LogScreenData{
			Commits:     []git.GitCommit{{Hash: "abc123"}},
			GraphLayout: nil,
		},
	})
	m = m2.(Model)

	if len(m.stack) != 0 {
		t.Errorf("After first push (from loading), stack should still be empty, got %d", len(m.stack))
	}
	if cmd != nil {
		// Execute the command if any
		msg := cmd()
		m2, _ = m.Update(msg)
		m = m2.(Model)
	}

	// Push FilesScreen
	t.Logf("Before pushing FilesScreen: stack len = %d", len(m.stack))
	m2, _ = m.Update(PushScreenMsg{
		Screen: FilesScreen,
		Data: FilesScreenData{
			Commit: git.GitCommit{Hash: "abc123"},
			Files:  []git.FileChange{{Path: "test.go"}},
		},
	})
	m = m2.(Model)
	t.Logf("After pushing FilesScreen: stack len = %d", len(m.stack))

	if len(m.stack) != 1 {
		t.Errorf("After pushing FilesScreen, stack should have 1 item, got %d", len(m.stack))
	}

	// Push DiffScreen
	m2, _ = m.Update(PushScreenMsg{
		Screen: DiffScreen,
		Data: DiffScreenData{
			Commit: git.GitCommit{Hash: "abc123"},
			File:   git.FileChange{Path: "test.go"},
		},
	})
	m = m2.(Model)

	if len(m.stack) != 2 {
		t.Errorf("After pushing DiffScreen, stack should have 2 items, got %d", len(m.stack))
	}

	// Pop back to FilesScreen
	m2, _ = m.Update(PopScreenMsg{})
	m = m2.(Model)

	if len(m.stack) != 1 {
		t.Errorf("After popping once, stack should have 1 item, got %d", len(m.stack))
	}

	// Pop back to LogScreen
	m2, _ = m.Update(PopScreenMsg{})
	m = m2.(Model)

	if len(m.stack) != 0 {
		t.Errorf("After popping twice, stack should be empty, got %d", len(m.stack))
	}
}
