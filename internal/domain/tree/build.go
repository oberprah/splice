package tree

import (
	"sort"
	"strings"

	"github.com/oberprah/splice/internal/core"
)

// BuildTree constructs a hierarchical tree structure from a flat list of file changes.
// The root is a FolderNode at depth -1 that is always expanded.
// All folders are created as expanded by default.
// Children are sorted: folders first, then alphabetically.
func BuildTree(files []core.FileChange) TreeNode {
	root := &FolderNode{
		name:       "",
		depth:      -1,
		children:   []TreeNode{},
		isExpanded: true,
		stats:      FolderStats{},
	}

	for i := range files {
		insertFile(root, &files[i])
	}

	sortChildren(root)
	return root
}

// insertFile inserts a file into the tree, creating folder nodes as needed.
func insertFile(root *FolderNode, file *core.FileChange) {
	parts := strings.Split(file.Path, "/")
	current := root

	// Navigate/create folder nodes for all path components except the last (filename)
	for i := 0; i < len(parts)-1; i++ {
		folderName := parts[i]
		current = getOrCreateFolder(current, folderName)
	}

	// Add the file node
	filename := parts[len(parts)-1]
	fileNode := &FileNode{
		name:  filename,
		depth: current.depth + 1,
		file:  file,
	}
	current.children = append(current.children, fileNode)
}

// getOrCreateFolder finds an existing folder child or creates a new one.
func getOrCreateFolder(parent *FolderNode, name string) *FolderNode {
	// Look for existing folder with this name
	for _, child := range parent.children {
		if folder, ok := child.(*FolderNode); ok && folder.name == name {
			return folder
		}
	}

	// Create new folder
	folder := &FolderNode{
		name:       name,
		depth:      parent.depth + 1,
		children:   []TreeNode{},
		isExpanded: true,
		stats:      FolderStats{},
	}
	parent.children = append(parent.children, folder)
	return folder
}

// sortChildren recursively sorts all children in the tree.
// Folders come before files, then alphabetically by name.
func sortChildren(node TreeNode) {
	folder, ok := node.(*FolderNode)
	if !ok {
		return
	}

	// Sort children: folders first, then alphabetically
	sort.SliceStable(folder.children, func(i, j int) bool {
		iIsFolder := isFolder(folder.children[i])
		jIsFolder := isFolder(folder.children[j])

		// Folders before files
		if iIsFolder != jIsFolder {
			return iIsFolder
		}

		// Alphabetically by name
		return folder.children[i].GetName() < folder.children[j].GetName()
	})

	// Recursively sort children
	for _, child := range folder.children {
		sortChildren(child)
	}
}

// isFolder checks if a node is a FolderNode.
func isFolder(node TreeNode) bool {
	_, ok := node.(*FolderNode)
	return ok
}
