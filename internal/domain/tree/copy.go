package tree

// DeepCopy creates a deep copy of a tree node and all its descendants.
// This is needed for immutability when toggling folders.
func DeepCopy(node TreeNode) TreeNode {
	switch n := node.(type) {
	case *FolderNode:
		// Deep copy the folder and all its children
		childrenCopy := make([]TreeNode, len(n.children))
		for i, child := range n.children {
			childrenCopy[i] = DeepCopy(child)
		}

		return &FolderNode{
			name:       n.name,
			depth:      n.depth,
			children:   childrenCopy,
			isExpanded: n.isExpanded,
			stats:      n.stats, // FolderStats is a value type, safe to copy
		}

	case *FileNode:
		// FileNode contains a pointer to FileChange, which we don't copy
		// (it's immutable data from the commit)
		return &FileNode{
			name:  n.name,
			depth: n.depth,
			file:  n.file, // Pointer to immutable data
		}

	default:
		// Should never happen with our sum types
		return nil
	}
}
