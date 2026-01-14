package tree

// ComputeStats recursively computes aggregate statistics for a tree node.
// For FileNode: returns stats from the associated FileChange (1 file, additions, deletions).
// For FolderNode: recursively sums all descendant file stats.
func ComputeStats(node TreeNode) FolderStats {
	switch n := node.(type) {
	case *FileNode:
		// For a file, return stats directly from the FileChange
		return FolderStats{
			FileCount: 1,
			Additions: n.file.Additions,
			Deletions: n.file.Deletions,
		}

	case *FolderNode:
		// For a folder, recursively sum stats from all children
		var total FolderStats
		for _, child := range n.children {
			childStats := ComputeStats(child)
			total.FileCount += childStats.FileCount
			total.Additions += childStats.Additions
			total.Deletions += childStats.Deletions
		}
		return total

	default:
		// Should never happen with our sum types, but return zero stats for safety
		return FolderStats{}
	}
}

// ApplyStats recursively computes and applies statistics to all FolderNodes in the tree.
// This mutates the tree in-place, updating the stats field of each FolderNode.
// Call this after building and collapsing the tree to populate folder statistics.
func ApplyStats(node TreeNode) {
	folder, ok := node.(*FolderNode)
	if !ok {
		// FileNodes don't have stats to update
		return
	}

	// First, recursively apply stats to all children
	for _, child := range folder.children {
		ApplyStats(child)
	}

	// Then compute and store stats for this folder
	folder.stats = ComputeStats(folder)
}
