package files

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/filetree"
)

// Update handles messages for the files state
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.DiffLoadedMsg:
		// Handle diff loading result
		if msg.Err != nil {
			// For now, just stay in files state on error
			// In the future, we could show an error message
			return s, nil
		}

		// Return command that produces PushDiffScreenMsg to navigate to DiffState
		return s, func() tea.Msg {
			return core.PushDiffScreenMsg{
				Source:    msg.Source,
				Files:     msg.Files,
				FileIndex: msg.FileIndex,
				File:      msg.File,
				Diff:      msg.Diff,
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous state using navigation pattern
			// (app.Model handles quitting when stack is empty)
			return s, func() tea.Msg {
				return core.PopScreenMsg{}
			}

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "enter":
			// Enter key: toggle folder or load diff for file
			if s.Cursor < len(s.VisibleItems) {
				cursorNode := s.VisibleItems[s.Cursor].Node
				switch node := cursorNode.(type) {
				case *filetree.FolderNode:
					// Toggle folder (collapse/expand)
					return toggleFolder(s, false, false, ctx)
				case *filetree.FileNode:
					// Load diff for file
					return s, s.loadDiff(*node.File(), ctx.FetchFullFileDiffForSource())
				}
			}
			return s, nil

		case " ":
			// Space key: toggle folder (same as Enter on folders)
			if s.Cursor < len(s.VisibleItems) {
				cursorNode := s.VisibleItems[s.Cursor].Node
				if _, ok := cursorNode.(*filetree.FolderNode); ok {
					return toggleFolder(s, false, false, ctx)
				}
			}
			return s, nil

		case "right":
			// Right arrow: expand folder only
			if s.Cursor < len(s.VisibleItems) {
				cursorNode := s.VisibleItems[s.Cursor].Node
				if _, ok := cursorNode.(*filetree.FolderNode); ok {
					return toggleFolder(s, true, false, ctx)
				}
			}
			return s, nil

		case "left":
			// Left arrow: collapse folder only
			if s.Cursor < len(s.VisibleItems) {
				cursorNode := s.VisibleItems[s.Cursor].Node
				if _, ok := cursorNode.(*filetree.FolderNode); ok {
					return toggleFolder(s, false, true, ctx)
				}
			}
			return s, nil

		case "j", "down":
			if len(s.VisibleItems) > 0 && s.Cursor < len(s.VisibleItems)-1 {
				s.Cursor++
				s.updateViewport(ctx.Height())
			}
			return s, nil

		case "k", "up":
			if s.Cursor > 0 {
				s.Cursor--
				s.updateViewport(ctx.Height())
			}
			return s, nil

		case "g":
			s.Cursor = 0
			s.ViewportStart = 0
			return s, nil

		case "G":
			if len(s.VisibleItems) > 0 {
				s.Cursor = len(s.VisibleItems) - 1
				s.updateViewport(ctx.Height())
			}
			return s, nil
		}
	}

	return s, nil
}

// updateViewport adjusts the viewport to keep the cursor visible
func (s *State) updateViewport(height int) {
	// Account for header lines when calculating available height
	headerLines := 2 // commit info + separator
	availableHeight := max(height-headerLines, 1)

	// Scroll down if cursor is below viewport
	if s.Cursor >= s.ViewportStart+availableHeight {
		s.ViewportStart = s.Cursor - availableHeight + 1
	}

	// Scroll up if cursor is above viewport
	if s.Cursor < s.ViewportStart {
		s.ViewportStart = s.Cursor
	}

	// Ensure viewport doesn't go negative
	if s.ViewportStart < 0 {
		s.ViewportStart = 0
	}
}

// loadDiff creates a command to fetch and parse the diff for a file
func (s *State) loadDiff(file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffForSourceFunc) tea.Cmd {
	return func() tea.Msg {
		// Fetch full file content and diff based on DiffSource type
		fullDiffResult, err := fetchFullFileDiff(s.Source, file)
		if err != nil {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    err,
			}
		}

		// Build the block-based diff
		fileDiff, err := diff.BuildFileDiff(
			file.Path,
			fullDiffResult.OldContent,
			fullDiffResult.NewContent,
			fullDiffResult.DiffOutput,
		)
		if err != nil {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    err,
			}
		}

		// Use the same ordering as the files view (alphabetical by path).
		orderedFiles := sortedFilesByPath(s.Files)

		// Find the file index in the ordered list
		fileIndex := 0
		for i, f := range orderedFiles {
			if f.Path == file.Path {
				fileIndex = i
				break
			}
		}

		return core.DiffLoadedMsg{
			Source:    s.Source,
			Files:     orderedFiles,
			FileIndex: fileIndex,
			File:      file,
			Diff:      fileDiff,
			Err:       nil,
		}
	}
}

func sortedFilesByPath(files []core.FileChange) []core.FileChange {
	ordered := append([]core.FileChange(nil), files...)
	sort.SliceStable(ordered, func(i, j int) bool {
		return ordered[i].Path < ordered[j].Path
	})
	return ordered
}

// toggleFolder toggles the expanded state of a folder at the cursor position.
// It creates a new state with a deep copy of the tree, modifies the folder,
// re-computes stats if needed, and re-flattens the filetree.
//
// Parameters:
// - expandOnly: if true, only expand (don't collapse already expanded folders)
// - collapseOnly: if true, only collapse (don't expand already collapsed folders)
// - ctx: core.Context for accessing screen dimensions
//
// Returns a new state with the updated tree and cursor position preserved.
func toggleFolder(s *State, expandOnly bool, collapseOnly bool, ctx core.Context) (*State, tea.Cmd) {
	if s.Cursor >= len(s.VisibleItems) {
		return s, nil
	}

	// Get the folder at cursor position
	cursorNode := s.VisibleItems[s.Cursor].Node
	folder, ok := cursorNode.(*filetree.FolderNode)
	if !ok {
		// Not a folder, nothing to toggle
		return s, nil
	}

	// Check if we should do nothing based on current state and flags
	if expandOnly && folder.IsExpanded() {
		// Already expanded, no-op
		return s, nil
	}
	if collapseOnly && !folder.IsExpanded() {
		// Already collapsed, no-op
		return s, nil
	}

	// Deep copy the tree for immutability
	newRoot := filetree.DeepCopy(s.Root)

	// Find the folder in the new tree using the cursor index.
	// Since flattening is deterministic, the same cursor index in the new tree
	// after deep copy should point to the same logical item.
	// This approach correctly handles duplicate folder names at the same depth.
	newVisibleItems := filetree.FlattenVisible(newRoot)
	if s.Cursor >= len(newVisibleItems) {
		// Shouldn't happen, but handle gracefully
		return s, nil
	}

	targetNode := newVisibleItems[s.Cursor].Node
	targetFolder, ok := targetNode.(*filetree.FolderNode)
	if !ok {
		// Shouldn't happen (we already checked it's a folder), but handle gracefully
		return s, nil
	}

	// Toggle the folder's expanded state
	toggleFolderNode(targetFolder)

	// If we collapsed a folder, re-compute stats
	if !targetFolder.IsExpanded() {
		filetree.ApplyStats(newRoot)
	}

	// Re-flatten to get new visible items (after toggle)
	newVisibleItems = filetree.FlattenVisible(newRoot)

	// Preserve cursor position or adjust if needed
	newCursor := s.Cursor
	if newCursor >= len(newVisibleItems) {
		newCursor = max(0, len(newVisibleItems)-1)
	}

	// Create new state
	newState := &State{
		Source:        s.Source,
		Files:         s.Files,
		Root:          newRoot,
		VisibleItems:  newVisibleItems,
		Cursor:        newCursor,
		ViewportStart: s.ViewportStart,
	}

	// Adjust viewport if needed
	newState.updateViewport(ctx.Height())

	return newState, nil
}

// toggleFolderNode toggles the isExpanded field of a FolderNode.
func toggleFolderNode(folder *filetree.FolderNode) {
	folder.SetExpanded(!folder.IsExpanded())
}
