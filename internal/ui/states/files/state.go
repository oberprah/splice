package files

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/tree"
)

// FilesState represents the state when displaying files changed in a commit
type State struct {
	Source        core.DiffSource
	Files         []core.FileChange      // Original flat list
	Root          tree.TreeNode          // Tree root (FolderNode at depth -1)
	VisibleItems  []tree.VisibleTreeItem // Flattened for navigation
	Cursor        int
	ViewportStart int
}

// New creates a new FilesState with cursor at the first file.
// Builds the tree structure from files, collapses paths, applies stats, and flattens to visible items.
// All folders start expanded.
func New(source core.DiffSource, files []core.FileChange) *State {
	// Build the tree structure
	root := tree.BuildTree(files)

	// Collapse single-child folder paths
	root = tree.CollapsePaths(root)

	// Compute and apply stats to all folders
	tree.ApplyStats(root)

	// Flatten to visible items for navigation
	visibleItems := tree.FlattenVisible(root)

	return &State{
		Source:        source,
		Files:         files,
		Root:          root,
		VisibleItems:  visibleItems,
		Cursor:        0,
		ViewportStart: 0,
	}
}
