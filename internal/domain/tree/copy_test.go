package tree

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestDeepCopy_FileNode(t *testing.T) {
	file := &core.FileChange{Path: "test.go", Status: "M"}
	original := &FileNode{
		name:  "test.go",
		depth: 1,
		file:  file,
	}

	copied := DeepCopy(original)

	// Verify it's a FileNode
	copiedFile, ok := copied.(*FileNode)
	if !ok {
		t.Fatalf("Expected *FileNode, got %T", copied)
	}

	// Verify values match
	if copiedFile.name != original.name {
		t.Errorf("Expected name %s, got %s", original.name, copiedFile.name)
	}
	if copiedFile.depth != original.depth {
		t.Errorf("Expected depth %d, got %d", original.depth, copiedFile.depth)
	}
	if copiedFile.file != original.file {
		t.Error("Expected file pointer to be the same (shared immutable data)")
	}

	// Verify it's a different instance
	if copiedFile == original {
		t.Error("Expected different instance, got same pointer")
	}
}

func TestDeepCopy_FolderNode_Empty(t *testing.T) {
	original := &FolderNode{
		name:       "src",
		depth:      0,
		children:   []TreeNode{},
		isExpanded: true,
		stats:      FolderStats{FileCount: 0},
	}

	copied := DeepCopy(original)

	// Verify it's a FolderNode
	copiedFolder, ok := copied.(*FolderNode)
	if !ok {
		t.Fatalf("Expected *FolderNode, got %T", copied)
	}

	// Verify values match
	if copiedFolder.name != original.name {
		t.Errorf("Expected name %s, got %s", original.name, copiedFolder.name)
	}
	if copiedFolder.depth != original.depth {
		t.Errorf("Expected depth %d, got %d", original.depth, copiedFolder.depth)
	}
	if copiedFolder.isExpanded != original.isExpanded {
		t.Errorf("Expected isExpanded %v, got %v", original.isExpanded, copiedFolder.isExpanded)
	}

	// Verify it's a different instance
	if copiedFolder == original {
		t.Error("Expected different instance, got same pointer")
	}
}

func TestDeepCopy_FolderNode_WithChildren(t *testing.T) {
	file1 := &core.FileChange{Path: "src/app.go", Status: "M"}
	file2 := &core.FileChange{Path: "src/test.go", Status: "A"}

	original := &FolderNode{
		name:  "src",
		depth: 0,
		children: []TreeNode{
			&FileNode{name: "app.go", depth: 1, file: file1},
			&FileNode{name: "test.go", depth: 1, file: file2},
		},
		isExpanded: true,
		stats:      FolderStats{FileCount: 2, Additions: 30, Deletions: 5},
	}

	copied := DeepCopy(original)

	// Verify it's a FolderNode
	copiedFolder, ok := copied.(*FolderNode)
	if !ok {
		t.Fatalf("Expected *FolderNode, got %T", copied)
	}

	// Verify children count
	if len(copiedFolder.children) != len(original.children) {
		t.Fatalf("Expected %d children, got %d", len(original.children), len(copiedFolder.children))
	}

	// Verify children are different instances
	for i := range original.children {
		if copiedFolder.children[i] == original.children[i] {
			t.Errorf("Expected child %d to be a different instance", i)
		}
	}

	// Verify children values match
	originalChild1 := original.children[0].(*FileNode)
	copiedChild1 := copiedFolder.children[0].(*FileNode)
	if copiedChild1.name != originalChild1.name {
		t.Errorf("Expected child name %s, got %s", originalChild1.name, copiedChild1.name)
	}

	// Verify it's a different instance
	if copiedFolder == original {
		t.Error("Expected different instance, got same pointer")
	}
}

func TestDeepCopy_NestedFolders(t *testing.T) {
	file := &core.FileChange{Path: "src/utils/helper.go", Status: "M"}

	original := &FolderNode{
		name:  "src",
		depth: 0,
		children: []TreeNode{
			&FolderNode{
				name:  "utils",
				depth: 1,
				children: []TreeNode{
					&FileNode{name: "helper.go", depth: 2, file: file},
				},
				isExpanded: true,
				stats:      FolderStats{FileCount: 1},
			},
		},
		isExpanded: true,
		stats:      FolderStats{FileCount: 1},
	}

	copied := DeepCopy(original)

	// Verify it's a FolderNode
	copiedFolder, ok := copied.(*FolderNode)
	if !ok {
		t.Fatalf("Expected *FolderNode, got %T", copied)
	}

	// Verify nested structure
	if len(copiedFolder.children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(copiedFolder.children))
	}

	nestedFolder, ok := copiedFolder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected nested folder to be *FolderNode, got %T", copiedFolder.children[0])
	}

	if nestedFolder.name != "utils" {
		t.Errorf("Expected nested folder name 'utils', got %s", nestedFolder.name)
	}

	// Verify nested child
	if len(nestedFolder.children) != 1 {
		t.Fatalf("Expected nested folder to have 1 child, got %d", len(nestedFolder.children))
	}

	// Verify all instances are different
	originalNested := original.children[0].(*FolderNode)
	if nestedFolder == originalNested {
		t.Error("Expected nested folder to be a different instance")
	}
}

func TestDeepCopy_Independence(t *testing.T) {
	file := &core.FileChange{Path: "test.go", Status: "M"}

	original := &FolderNode{
		name:  "src",
		depth: 0,
		children: []TreeNode{
			&FileNode{name: "test.go", depth: 1, file: file},
		},
		isExpanded: true,
		stats:      FolderStats{FileCount: 1},
	}

	copied := DeepCopy(original)
	copiedFolder := copied.(*FolderNode)

	// Modify the copy
	copiedFolder.isExpanded = false

	// Verify original is unchanged
	if !original.isExpanded {
		t.Error("Modifying copy should not affect original")
	}
}
