package files

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/domain/filetree"
	"github.com/oberprah/splice/internal/git"
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
				Source:        msg.Source,
				File:          msg.File,
				Diff:          msg.Diff,
				ChangeIndices: msg.ChangeIndices,
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
					return s, s.loadDiff(*node.File(), ctx.FetchFullFileDiff())
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

		case "ctrl+d":
			// Scroll down half page
			if len(s.VisibleItems) > 0 {
				headerLines := 2
				availableHeight := max(ctx.Height()-headerLines, 1)
				halfPage := availableHeight / 2
				s.Cursor = min(s.Cursor+halfPage, len(s.VisibleItems)-1)
				s.updateViewport(ctx.Height())
			}
			return s, nil

		case "ctrl+u":
			// Scroll up half page
			if len(s.VisibleItems) > 0 {
				headerLines := 2
				availableHeight := max(ctx.Height()-headerLines, 1)
				halfPage := availableHeight / 2
				s.Cursor = max(s.Cursor-halfPage, 0)
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
func (s *State) loadDiff(file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffFunc) tea.Cmd {
	return func() tea.Msg {
		// Fetch full file content and diff based on DiffSource type
		fullDiffResult, err := fetchFileDiffForSource(s.Source, file, fetchFullFileDiff)
		if err != nil {
			return core.DiffLoadedMsg{
				Source: s.Source,
				File:   file,
				Err:    err,
			}
		}

		// Build the complete aligned diff with change indices
		alignedDiff, changeIndices, err := diff.BuildAlignedFileDiff(
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

		return core.DiffLoadedMsg{
			Source:        s.Source,
			File:          file,
			Diff:          alignedDiff,
			ChangeIndices: changeIndices,
			Err:           nil,
		}
	}
}

// fetchFileDiffForSource fetches the full file diff based on the DiffSource type.
// Uses type switch to call the appropriate git function for each source type.
func fetchFileDiffForSource(source core.DiffSource, file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffFunc) (*core.FullFileDiffResult, error) {
	switch src := source.(type) {
	case core.CommitRangeDiffSource:
		// For commit ranges, use the injected fetchFullFileDiff function
		commitRange := src.ToCommitRange()
		return fetchFullFileDiff(commitRange, file)

	case core.UncommittedChangesDiffSource:
		// For uncommitted changes, type switch on Type field to call appropriate git function
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			return git.FetchUnstagedFileDiff(file)
		case core.UncommittedTypeStaged:
			return git.FetchStagedFileDiff(file)
		case core.UncommittedTypeAll:
			return git.FetchAllUncommittedFileDiff(file)
		default:
			return nil, fmt.Errorf("unknown uncommitted type: %v", src.Type)
		}

	default:
		return nil, fmt.Errorf("unknown diff source type: %T", source)
	}
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
