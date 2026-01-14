package tree

import "github.com/oberprah/splice/internal/core"

// TreeNode is the interface for all tree nodes (folders and files).
// It uses the sum type pattern to make illegal states unrepresentable.
type TreeNode interface {
	isTreeNode()
	GetName() string
	GetDepth() int
}

// FolderNode represents a directory in the tree.
type FolderNode struct {
	name       string      // May be collapsed path in future: "src/components/nested"
	depth      int         // Distance from root (-1 for root itself)
	children   []TreeNode  // Sorted: folders first, then files (alphabetical)
	isExpanded bool        // Whether the folder is expanded in the UI
	stats      FolderStats // Aggregate stats for collapsed display (computed later)
}

// FileNode represents a file in the tree.
type FileNode struct {
	name  string           // Just the filename: "App.tsx"
	depth int              // Distance from root
	file  *core.FileChange // Pointer to original file data
}

// FolderStats contains aggregate statistics for a folder.
// Used when displaying collapsed folders.
type FolderStats struct {
	FileCount int // Total number of files in this folder (recursively)
	Additions int // Total additions in this folder (recursively)
	Deletions int // Total deletions in this folder (recursively)
}

// VisibleTreeItem contains a tree node along with rendering metadata
// computed during the flattening process.
type VisibleTreeItem struct {
	Node        TreeNode // The tree node to render
	IsLastChild bool     // Whether this is the last child of its parent
	ParentLines []bool   // Which depth levels need │ continuation character
}

// Implement TreeNode interface for FolderNode
func (f *FolderNode) isTreeNode()     {}
func (f *FolderNode) GetName() string { return f.name }
func (f *FolderNode) GetDepth() int   { return f.depth }

// Getters for FolderNode fields (needed for UI formatting)
func (f *FolderNode) IsExpanded() bool     { return f.isExpanded }
func (f *FolderNode) Stats() FolderStats   { return f.stats }
func (f *FolderNode) Children() []TreeNode { return f.children }

// Implement TreeNode interface for FileNode
func (f *FileNode) isTreeNode()     {}
func (f *FileNode) GetName() string { return f.name }
func (f *FileNode) GetDepth() int   { return f.depth }

// Getter for FileNode file data (needed for UI formatting)
func (f *FileNode) File() *core.FileChange { return f.file }

// NewFolderNode creates a new FolderNode (for testing).
func NewFolderNode(name string, depth int, isExpanded bool, stats FolderStats) *FolderNode {
	return &FolderNode{
		name:       name,
		depth:      depth,
		children:   []TreeNode{},
		isExpanded: isExpanded,
		stats:      stats,
	}
}

// NewFileNode creates a new FileNode (for testing).
func NewFileNode(name string, depth int, file *core.FileChange) *FileNode {
	return &FileNode{
		name:  name,
		depth: depth,
		file:  file,
	}
}
