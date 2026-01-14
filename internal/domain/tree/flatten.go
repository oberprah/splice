package tree

// FlattenVisible walks the tree depth-first and returns a flat list of visible items
// with rendering metadata (isLastChild, parentLines) for box-drawing characters.
//
// The root node (depth -1) is skipped, but its children are processed.
// For expanded FolderNodes: includes the folder and recurses into children.
// For collapsed FolderNodes: includes the folder but skips children.
// For FileNodes: always includes them.
func FlattenVisible(root TreeNode) []VisibleTreeItem {
	var result []VisibleTreeItem

	// Start walking from root, but skip the root itself (depth -1)
	folder, ok := root.(*FolderNode)
	if !ok {
		// Root should always be a FolderNode
		return result
	}

	// Process root's children
	for i, child := range folder.children {
		isLastChild := i == len(folder.children)-1
		parentLines := []bool{} // Root is at depth -1, so no parent lines yet
		walk(child, isLastChild, parentLines, &result)
	}

	return result
}

// walk recursively processes a node and its children, building the flattened list.
// It computes rendering metadata (isLastChild, parentLines) as it goes.
func walk(node TreeNode, isLastChild bool, parentLines []bool, result *[]VisibleTreeItem) {
	// Add this node to the result (it's visible)
	item := VisibleTreeItem{
		Node:        node,
		IsLastChild: isLastChild,
		ParentLines: parentLines,
	}
	*result = append(*result, item)

	// If it's a folder and expanded, recurse into children
	if folder, ok := node.(*FolderNode); ok && folder.isExpanded {
		// Build parent lines for children
		// Each child needs to know about all ancestor continuation lines
		// ParentLines[i] = true means "draw │ at level i" (ancestor has more siblings)
		// If current node is NOT last child, its children need a continuation line
		childParentLines := make([]bool, len(parentLines)+1)
		copy(childParentLines, parentLines)
		childParentLines[len(parentLines)] = !isLastChild

		// Process each child
		for i, child := range folder.children {
			childIsLastChild := i == len(folder.children)-1
			walk(child, childIsLastChild, childParentLines, result)
		}
	}
	// If folder is collapsed, or it's a file, we don't recurse
}
