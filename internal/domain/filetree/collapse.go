package filetree

import "strings"

// CollapsePaths collapses single-child folder chains into collapsed paths.
// For example, a chain like src/ → components/ → nested/ becomes a single
// FolderNode with name "src/components/nested".
//
// Rules:
// - Only collapse if the folder has exactly one child that is also a folder
// - Do NOT collapse if the folder's only child is a file
// - Do NOT collapse if the folder has multiple children
// - Continue collapsing until no more collapsing is possible in that chain
// - Preserve the isExpanded state from the first folder in the chain
// - Adjust depths after collapsing
//
// This function mutates the tree in-place and returns the root.
func CollapsePaths(root TreeNode) TreeNode {
	folder, ok := root.(*FolderNode)
	if !ok {
		return root
	}

	collapseFolder(folder)
	return root
}

// collapseFolder recursively collapses single-child folder chains.
func collapseFolder(folder *FolderNode) {
	// First, recursively process all children
	for i := range folder.children {
		if childFolder, ok := folder.children[i].(*FolderNode); ok {
			collapseFolder(childFolder)
		}
	}

	// Now collapse this folder's children if applicable
	for i := 0; i < len(folder.children); i++ {
		childFolder, ok := folder.children[i].(*FolderNode)
		if !ok {
			continue
		}

		// Collapse chain starting from this child
		collapsedName, collapsedChildren, preservedExpanded := collapseChain(childFolder)

		// Replace the child with a new collapsed folder
		folder.children[i] = &FolderNode{
			name:       collapsedName,
			depth:      childFolder.depth,
			children:   collapsedChildren,
			isExpanded: preservedExpanded,
			stats:      FolderStats{}, // Will be computed later in Step 3
		}

		// Adjust depths of all descendants
		adjustDepths(folder.children[i], childFolder.depth)
	}
}

// collapseChain collapses a chain of single-child folders starting from the given folder.
// Returns the collapsed path name, the final children, and the isExpanded state.
func collapseChain(folder *FolderNode) (name string, children []TreeNode, isExpanded bool) {
	pathComponents := []string{folder.name}
	current := folder
	isExpanded = folder.isExpanded // Preserve state from first folder

	// Keep collapsing while there's exactly one child that is a folder
	for len(current.children) == 1 {
		childFolder, ok := current.children[0].(*FolderNode)
		if !ok {
			break // Only child is a file, stop collapsing
		}

		pathComponents = append(pathComponents, childFolder.name)
		current = childFolder
	}

	// Build the collapsed name
	name = strings.Join(pathComponents, "/")
	children = current.children

	return name, children, isExpanded
}

// adjustDepths recursively adjusts the depth of all nodes in the subtree.
func adjustDepths(node TreeNode, newDepth int) {
	switch n := node.(type) {
	case *FolderNode:
		n.depth = newDepth
		for _, child := range n.children {
			adjustDepths(child, newDepth+1)
		}
	case *FileNode:
		n.depth = newDepth
	}
}
