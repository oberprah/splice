package tree

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestBuildTree_EmptyInput(t *testing.T) {
	files := []core.FileChange{}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	if folder.GetDepth() != -1 {
		t.Errorf("Expected root depth to be -1, got %d", folder.GetDepth())
	}

	if !folder.isExpanded {
		t.Error("Expected root to be expanded")
	}

	if len(folder.children) != 0 {
		t.Errorf("Expected no children, got %d", len(folder.children))
	}
}

func TestBuildTree_SingleFileAtRoot(t *testing.T) {
	files := []core.FileChange{
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	if len(folder.children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(folder.children))
	}

	file, ok := folder.children[0].(*FileNode)
	if !ok {
		t.Fatalf("Expected child to be *FileNode, got %T", folder.children[0])
	}

	if file.name != "README.md" {
		t.Errorf("Expected file name 'README.md', got '%s'", file.name)
	}

	if file.GetDepth() != 0 {
		t.Errorf("Expected file depth 0, got %d", file.GetDepth())
	}

	if file.file.Path != "README.md" {
		t.Errorf("Expected file path 'README.md', got '%s'", file.file.Path)
	}
}

func TestBuildTree_MultipleFilesInNestedFolders(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/components/App.tsx", Status: "M", Additions: 17, Deletions: 13},
		{Path: "src/utils/helper.ts", Status: "A", Additions: 42, Deletions: 0},
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	if len(folder.children) != 2 {
		t.Fatalf("Expected 2 children (README.md and src/), got %d", len(folder.children))
	}

	// First child should be src/ folder (folders first)
	srcFolder, ok := folder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected first child to be *FolderNode, got %T", folder.children[0])
	}

	if srcFolder.name != "src" {
		t.Errorf("Expected folder name 'src', got '%s'", srcFolder.name)
	}

	if srcFolder.GetDepth() != 0 {
		t.Errorf("Expected src folder depth 0, got %d", srcFolder.GetDepth())
	}

	if !srcFolder.isExpanded {
		t.Error("Expected src folder to be expanded")
	}

	if len(srcFolder.children) != 2 {
		t.Fatalf("Expected src to have 2 children (components/ and utils/), got %d", len(srcFolder.children))
	}

	// Check components folder
	componentsFolder, ok := srcFolder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected components to be *FolderNode, got %T", srcFolder.children[0])
	}

	if componentsFolder.name != "components" {
		t.Errorf("Expected folder name 'components', got '%s'", componentsFolder.name)
	}

	if componentsFolder.GetDepth() != 1 {
		t.Errorf("Expected components folder depth 1, got %d", componentsFolder.GetDepth())
	}

	if len(componentsFolder.children) != 1 {
		t.Fatalf("Expected components to have 1 child (App.tsx), got %d", len(componentsFolder.children))
	}

	appFile, ok := componentsFolder.children[0].(*FileNode)
	if !ok {
		t.Fatalf("Expected App.tsx to be *FileNode, got %T", componentsFolder.children[0])
	}

	if appFile.name != "App.tsx" {
		t.Errorf("Expected file name 'App.tsx', got '%s'", appFile.name)
	}

	if appFile.GetDepth() != 2 {
		t.Errorf("Expected App.tsx depth 2, got %d", appFile.GetDepth())
	}

	// Second child should be README.md file
	readmeFile, ok := folder.children[1].(*FileNode)
	if !ok {
		t.Fatalf("Expected second child to be *FileNode, got %T", folder.children[1])
	}

	if readmeFile.name != "README.md" {
		t.Errorf("Expected file name 'README.md', got '%s'", readmeFile.name)
	}
}

func TestBuildTree_FolderAndFileSorting(t *testing.T) {
	files := []core.FileChange{
		{Path: "zebra.txt", Status: "M"},
		{Path: "beta/file.txt", Status: "M"},
		{Path: "alpha.txt", Status: "A"},
		{Path: "delta/file.txt", Status: "M"},
		{Path: "charlie.txt", Status: "D"},
	}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	if len(folder.children) != 5 {
		t.Fatalf("Expected 5 children, got %d", len(folder.children))
	}

	// First two should be folders (alphabetically: beta, delta)
	expectedFolders := []string{"beta", "delta"}
	for i, expectedName := range expectedFolders {
		node, ok := folder.children[i].(*FolderNode)
		if !ok {
			t.Fatalf("Expected child %d to be *FolderNode, got %T", i, folder.children[i])
		}
		if node.name != expectedName {
			t.Errorf("Expected folder %d to be '%s', got '%s'", i, expectedName, node.name)
		}
	}

	// Last three should be files (alphabetically: alpha.txt, charlie.txt, zebra.txt)
	expectedFiles := []string{"alpha.txt", "charlie.txt", "zebra.txt"}
	for i, expectedName := range expectedFiles {
		node, ok := folder.children[i+2].(*FileNode)
		if !ok {
			t.Fatalf("Expected child %d to be *FileNode, got %T", i+2, folder.children[i+2])
		}
		if node.name != expectedName {
			t.Errorf("Expected file %d to be '%s', got '%s'", i, expectedName, node.name)
		}
	}
}

func TestBuildTree_FilesAtDifferentDepths(t *testing.T) {
	files := []core.FileChange{
		{Path: "a/b/c/d/deep.txt", Status: "M"},
		{Path: "a/shallow.txt", Status: "A"},
		{Path: "root.txt", Status: "D"},
	}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	if len(folder.children) != 2 {
		t.Fatalf("Expected 2 children (a/ and root.txt), got %d", len(folder.children))
	}

	// Check 'a' folder
	aFolder, ok := folder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected 'a' to be *FolderNode, got %T", folder.children[0])
	}

	if aFolder.GetDepth() != 0 {
		t.Errorf("Expected 'a' folder depth 0, got %d", aFolder.GetDepth())
	}

	if len(aFolder.children) != 2 {
		t.Fatalf("Expected 'a' to have 2 children (b/ and shallow.txt), got %d", len(aFolder.children))
	}

	// Check 'b' folder (should be first - folders before files)
	bFolder, ok := aFolder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected 'b' to be *FolderNode, got %T", aFolder.children[0])
	}

	if bFolder.GetDepth() != 1 {
		t.Errorf("Expected 'b' folder depth 1, got %d", bFolder.GetDepth())
	}

	// Navigate down to deep.txt
	cFolder, ok := bFolder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected 'c' to be *FolderNode, got %T", bFolder.children[0])
	}

	if cFolder.GetDepth() != 2 {
		t.Errorf("Expected 'c' folder depth 2, got %d", cFolder.GetDepth())
	}

	dFolder, ok := cFolder.children[0].(*FolderNode)
	if !ok {
		t.Fatalf("Expected 'd' to be *FolderNode, got %T", cFolder.children[0])
	}

	if dFolder.GetDepth() != 3 {
		t.Errorf("Expected 'd' folder depth 3, got %d", dFolder.GetDepth())
	}

	deepFile, ok := dFolder.children[0].(*FileNode)
	if !ok {
		t.Fatalf("Expected deep.txt to be *FileNode, got %T", dFolder.children[0])
	}

	if deepFile.GetDepth() != 4 {
		t.Errorf("Expected deep.txt depth 4, got %d", deepFile.GetDepth())
	}

	if deepFile.name != "deep.txt" {
		t.Errorf("Expected file name 'deep.txt', got '%s'", deepFile.name)
	}
}

func TestBuildTree_PreservesFileChangeData(t *testing.T) {
	files := []core.FileChange{
		{Path: "test.txt", Status: "M", Additions: 10, Deletions: 5, IsBinary: false},
		{Path: "image.png", Status: "A", Additions: 0, Deletions: 0, IsBinary: true},
	}
	root := BuildTree(files)

	folder, ok := root.(*FolderNode)
	if !ok {
		t.Fatalf("Expected root to be *FolderNode, got %T", root)
	}

	// Check image.png (should be first - alphabetically before test.txt)
	imageFile, ok := folder.children[0].(*FileNode)
	if !ok {
		t.Fatalf("Expected first child to be *FileNode, got %T", folder.children[0])
	}

	if imageFile.file.Status != "A" {
		t.Errorf("Expected status 'A', got '%s'", imageFile.file.Status)
	}

	if imageFile.file.IsBinary != true {
		t.Errorf("Expected IsBinary to be true, got %v", imageFile.file.IsBinary)
	}

	// Check test.txt
	testFile, ok := folder.children[1].(*FileNode)
	if !ok {
		t.Fatalf("Expected second child to be *FileNode, got %T", folder.children[1])
	}

	if testFile.file.Additions != 10 {
		t.Errorf("Expected 10 additions, got %d", testFile.file.Additions)
	}

	if testFile.file.Deletions != 5 {
		t.Errorf("Expected 5 deletions, got %d", testFile.file.Deletions)
	}
}
