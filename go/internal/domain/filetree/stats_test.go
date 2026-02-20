package filetree

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestComputeStats_FileNode(t *testing.T) {
	// Create a file node
	fileChange := &core.FileChange{
		Path:      "test.txt",
		Status:    "M",
		Additions: 10,
		Deletions: 5,
		IsBinary:  false,
	}
	fileNode := &FileNode{
		name:  "test.txt",
		depth: 1,
		file:  fileChange,
	}

	stats := ComputeStats(fileNode)

	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount to be 1, got %d", stats.FileCount)
	}

	if stats.Additions != 10 {
		t.Errorf("Expected Additions to be 10, got %d", stats.Additions)
	}

	if stats.Deletions != 5 {
		t.Errorf("Expected Deletions to be 5, got %d", stats.Deletions)
	}
}

func TestComputeStats_FileNode_BinaryFile(t *testing.T) {
	// Binary files typically have 0 additions/deletions
	fileChange := &core.FileChange{
		Path:      "image.png",
		Status:    "A",
		Additions: 0,
		Deletions: 0,
		IsBinary:  true,
	}
	fileNode := &FileNode{
		name:  "image.png",
		depth: 1,
		file:  fileChange,
	}

	stats := ComputeStats(fileNode)

	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount to be 1, got %d", stats.FileCount)
	}

	if stats.Additions != 0 {
		t.Errorf("Expected Additions to be 0, got %d", stats.Additions)
	}

	if stats.Deletions != 0 {
		t.Errorf("Expected Deletions to be 0, got %d", stats.Deletions)
	}
}

func TestComputeStats_FolderNode_SingleFile(t *testing.T) {
	// Create a folder with a single file
	fileChange := &core.FileChange{
		Path:      "src/test.txt",
		Status:    "M",
		Additions: 10,
		Deletions: 5,
	}
	fileNode := &FileNode{
		name:  "test.txt",
		depth: 1,
		file:  fileChange,
	}
	folder := &FolderNode{
		name:       "src",
		depth:      0,
		children:   []TreeNode{fileNode},
		isExpanded: true,
	}

	stats := ComputeStats(folder)

	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount to be 1, got %d", stats.FileCount)
	}

	if stats.Additions != 10 {
		t.Errorf("Expected Additions to be 10, got %d", stats.Additions)
	}

	if stats.Deletions != 5 {
		t.Errorf("Expected Deletions to be 5, got %d", stats.Deletions)
	}
}

func TestComputeStats_FolderNode_MultipleFiles(t *testing.T) {
	// Create a folder with multiple files
	file1 := &FileNode{
		name:  "a.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "src/a.txt",
			Status:    "M",
			Additions: 10,
			Deletions: 5,
		},
	}
	file2 := &FileNode{
		name:  "b.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "src/b.txt",
			Status:    "A",
			Additions: 20,
			Deletions: 0,
		},
	}
	file3 := &FileNode{
		name:  "c.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "src/c.txt",
			Status:    "D",
			Additions: 0,
			Deletions: 15,
		},
	}
	folder := &FolderNode{
		name:       "src",
		depth:      0,
		children:   []TreeNode{file1, file2, file3},
		isExpanded: true,
	}

	stats := ComputeStats(folder)

	if stats.FileCount != 3 {
		t.Errorf("Expected FileCount to be 3, got %d", stats.FileCount)
	}

	if stats.Additions != 30 {
		t.Errorf("Expected Additions to be 30 (10+20+0), got %d", stats.Additions)
	}

	if stats.Deletions != 20 {
		t.Errorf("Expected Deletions to be 20 (5+0+15), got %d", stats.Deletions)
	}
}

func TestComputeStats_FolderNode_NestedFolders(t *testing.T) {
	// Create nested folder structure:
	// src/
	//   components/
	//     App.tsx (17 additions, 13 deletions)
	//   utils/
	//     helper.ts (42 additions, 0 deletions)
	appFile := &FileNode{
		name:  "App.tsx",
		depth: 2,
		file: &core.FileChange{
			Path:      "src/components/App.tsx",
			Status:    "M",
			Additions: 17,
			Deletions: 13,
		},
	}
	componentsFolder := &FolderNode{
		name:       "components",
		depth:      1,
		children:   []TreeNode{appFile},
		isExpanded: true,
	}

	helperFile := &FileNode{
		name:  "helper.ts",
		depth: 2,
		file: &core.FileChange{
			Path:      "src/utils/helper.ts",
			Status:    "A",
			Additions: 42,
			Deletions: 0,
		},
	}
	utilsFolder := &FolderNode{
		name:       "utils",
		depth:      1,
		children:   []TreeNode{helperFile},
		isExpanded: true,
	}

	srcFolder := &FolderNode{
		name:       "src",
		depth:      0,
		children:   []TreeNode{componentsFolder, utilsFolder},
		isExpanded: true,
	}

	stats := ComputeStats(srcFolder)

	if stats.FileCount != 2 {
		t.Errorf("Expected FileCount to be 2, got %d", stats.FileCount)
	}

	if stats.Additions != 59 {
		t.Errorf("Expected Additions to be 59 (17+42), got %d", stats.Additions)
	}

	if stats.Deletions != 13 {
		t.Errorf("Expected Deletions to be 13, got %d", stats.Deletions)
	}

	// Also verify stats for nested folders
	componentsStats := ComputeStats(componentsFolder)
	if componentsStats.FileCount != 1 {
		t.Errorf("Expected components FileCount to be 1, got %d", componentsStats.FileCount)
	}

	if componentsStats.Additions != 17 {
		t.Errorf("Expected components Additions to be 17, got %d", componentsStats.Additions)
	}

	utilsStats := ComputeStats(utilsFolder)
	if utilsStats.FileCount != 1 {
		t.Errorf("Expected utils FileCount to be 1, got %d", utilsStats.FileCount)
	}

	if utilsStats.Additions != 42 {
		t.Errorf("Expected utils Additions to be 42, got %d", utilsStats.Additions)
	}
}

func TestComputeStats_FolderNode_DeepNesting(t *testing.T) {
	// Create deep nested structure: a/b/c/d/deep.txt
	deepFile := &FileNode{
		name:  "deep.txt",
		depth: 4,
		file: &core.FileChange{
			Path:      "a/b/c/d/deep.txt",
			Status:    "M",
			Additions: 100,
			Deletions: 50,
		},
	}

	dFolder := &FolderNode{
		name:       "d",
		depth:      3,
		children:   []TreeNode{deepFile},
		isExpanded: true,
	}

	cFolder := &FolderNode{
		name:       "c",
		depth:      2,
		children:   []TreeNode{dFolder},
		isExpanded: true,
	}

	bFolder := &FolderNode{
		name:       "b",
		depth:      1,
		children:   []TreeNode{cFolder},
		isExpanded: true,
	}

	aFolder := &FolderNode{
		name:       "a",
		depth:      0,
		children:   []TreeNode{bFolder},
		isExpanded: true,
	}

	// Stats should bubble up correctly
	stats := ComputeStats(aFolder)

	if stats.FileCount != 1 {
		t.Errorf("Expected FileCount to be 1, got %d", stats.FileCount)
	}

	if stats.Additions != 100 {
		t.Errorf("Expected Additions to be 100, got %d", stats.Additions)
	}

	if stats.Deletions != 50 {
		t.Errorf("Expected Deletions to be 50, got %d", stats.Deletions)
	}
}

func TestComputeStats_FolderNode_EmptyFolder(t *testing.T) {
	// Empty folder should have zero stats
	folder := &FolderNode{
		name:       "empty",
		depth:      0,
		children:   []TreeNode{},
		isExpanded: true,
	}

	stats := ComputeStats(folder)

	if stats.FileCount != 0 {
		t.Errorf("Expected FileCount to be 0, got %d", stats.FileCount)
	}

	if stats.Additions != 0 {
		t.Errorf("Expected Additions to be 0, got %d", stats.Additions)
	}

	if stats.Deletions != 0 {
		t.Errorf("Expected Deletions to be 0, got %d", stats.Deletions)
	}
}

func TestComputeStats_MixedFoldersAndFiles(t *testing.T) {
	// Create a folder with both files and subfolders:
	// parent/
	//   file1.txt (5 additions, 2 deletions)
	//   subfolder/
	//     file2.txt (10 additions, 3 deletions)
	//   file3.txt (8 additions, 4 deletions)

	file1 := &FileNode{
		name:  "file1.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "parent/file1.txt",
			Status:    "M",
			Additions: 5,
			Deletions: 2,
		},
	}

	file2 := &FileNode{
		name:  "file2.txt",
		depth: 2,
		file: &core.FileChange{
			Path:      "parent/subfolder/file2.txt",
			Status:    "A",
			Additions: 10,
			Deletions: 3,
		},
	}

	subfolder := &FolderNode{
		name:       "subfolder",
		depth:      1,
		children:   []TreeNode{file2},
		isExpanded: true,
	}

	file3 := &FileNode{
		name:  "file3.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "parent/file3.txt",
			Status:    "M",
			Additions: 8,
			Deletions: 4,
		},
	}

	parent := &FolderNode{
		name:       "parent",
		depth:      0,
		children:   []TreeNode{subfolder, file1, file3},
		isExpanded: true,
	}

	stats := ComputeStats(parent)

	if stats.FileCount != 3 {
		t.Errorf("Expected FileCount to be 3, got %d", stats.FileCount)
	}

	if stats.Additions != 23 {
		t.Errorf("Expected Additions to be 23 (5+10+8), got %d", stats.Additions)
	}

	if stats.Deletions != 9 {
		t.Errorf("Expected Deletions to be 9 (2+3+4), got %d", stats.Deletions)
	}
}

func TestApplyStats_UpdatesFolderStatsInPlace(t *testing.T) {
	// Create a tree structure
	file1 := &FileNode{
		name:  "a.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "src/a.txt",
			Status:    "M",
			Additions: 10,
			Deletions: 5,
		},
	}
	file2 := &FileNode{
		name:  "b.txt",
		depth: 1,
		file: &core.FileChange{
			Path:      "src/b.txt",
			Status:    "A",
			Additions: 20,
			Deletions: 0,
		},
	}
	srcFolder := &FolderNode{
		name:       "src",
		depth:      0,
		children:   []TreeNode{file1, file2},
		isExpanded: true,
		stats:      FolderStats{}, // Initially empty
	}
	root := &FolderNode{
		name:       "",
		depth:      -1,
		children:   []TreeNode{srcFolder},
		isExpanded: true,
		stats:      FolderStats{}, // Initially empty
	}

	// Apply stats to the tree
	ApplyStats(root)

	// Verify srcFolder stats were updated
	if srcFolder.stats.FileCount != 2 {
		t.Errorf("Expected srcFolder.stats.FileCount to be 2, got %d", srcFolder.stats.FileCount)
	}

	if srcFolder.stats.Additions != 30 {
		t.Errorf("Expected srcFolder.stats.Additions to be 30, got %d", srcFolder.stats.Additions)
	}

	if srcFolder.stats.Deletions != 5 {
		t.Errorf("Expected srcFolder.stats.Deletions to be 5, got %d", srcFolder.stats.Deletions)
	}

	// Verify root stats were updated
	if root.stats.FileCount != 2 {
		t.Errorf("Expected root.stats.FileCount to be 2, got %d", root.stats.FileCount)
	}

	if root.stats.Additions != 30 {
		t.Errorf("Expected root.stats.Additions to be 30, got %d", root.stats.Additions)
	}

	if root.stats.Deletions != 5 {
		t.Errorf("Expected root.stats.Deletions to be 5, got %d", root.stats.Deletions)
	}
}

func TestApplyStats_NestedFolders(t *testing.T) {
	// Create nested structure:
	// root/
	//   parent/
	//     child/
	//       file.txt (100 additions, 50 deletions)

	file := &FileNode{
		name:  "file.txt",
		depth: 2,
		file: &core.FileChange{
			Path:      "parent/child/file.txt",
			Status:    "M",
			Additions: 100,
			Deletions: 50,
		},
	}

	childFolder := &FolderNode{
		name:       "child",
		depth:      1,
		children:   []TreeNode{file},
		isExpanded: true,
		stats:      FolderStats{},
	}

	parentFolder := &FolderNode{
		name:       "parent",
		depth:      0,
		children:   []TreeNode{childFolder},
		isExpanded: true,
		stats:      FolderStats{},
	}

	root := &FolderNode{
		name:       "",
		depth:      -1,
		children:   []TreeNode{parentFolder},
		isExpanded: true,
		stats:      FolderStats{},
	}

	ApplyStats(root)

	// All folders should have the same stats (1 file, 100 additions, 50 deletions)
	if childFolder.stats.FileCount != 1 {
		t.Errorf("Expected childFolder.stats.FileCount to be 1, got %d", childFolder.stats.FileCount)
	}
	if childFolder.stats.Additions != 100 {
		t.Errorf("Expected childFolder.stats.Additions to be 100, got %d", childFolder.stats.Additions)
	}

	if parentFolder.stats.FileCount != 1 {
		t.Errorf("Expected parentFolder.stats.FileCount to be 1, got %d", parentFolder.stats.FileCount)
	}
	if parentFolder.stats.Additions != 100 {
		t.Errorf("Expected parentFolder.stats.Additions to be 100, got %d", parentFolder.stats.Additions)
	}

	if root.stats.FileCount != 1 {
		t.Errorf("Expected root.stats.FileCount to be 1, got %d", root.stats.FileCount)
	}
	if root.stats.Additions != 100 {
		t.Errorf("Expected root.stats.Additions to be 100, got %d", root.stats.Additions)
	}
}
