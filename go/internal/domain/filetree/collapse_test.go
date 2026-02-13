package filetree

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

// TestCollapsePaths_SingleChildFolderChain verifies that a chain of single-child
// folders is collapsed into a single FolderNode with a combined path.
func TestCollapsePaths_SingleChildFolderChain(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/components/nested/App.tsx"},
	}

	root := BuildTree(files)
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)
	if len(rootFolder.children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(rootFolder.children))
	}

	// Should have collapsed src/components/nested into one folder
	folder := rootFolder.children[0].(*FolderNode)
	expectedName := "src/components/nested"
	if folder.name != expectedName {
		t.Errorf("expected folder name %q, got %q", expectedName, folder.name)
	}

	// Should have depth 0 (first level below root)
	if folder.depth != 0 {
		t.Errorf("expected depth 0, got %d", folder.depth)
	}

	// Should have 1 child (the file)
	if len(folder.children) != 1 {
		t.Fatalf("expected 1 child (file), got %d", len(folder.children))
	}

	// Verify the file is still there
	file := folder.children[0].(*FileNode)
	if file.name != "App.tsx" {
		t.Errorf("expected file name %q, got %q", "App.tsx", file.name)
	}

	// File depth should be 1 (below the collapsed folder)
	if file.depth != 1 {
		t.Errorf("expected file depth 1, got %d", file.depth)
	}
}

// TestCollapsePaths_FolderWithMultipleChildren verifies that folders with
// multiple children are NOT collapsed.
func TestCollapsePaths_FolderWithMultipleChildren(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/components/App.tsx"},
		{Path: "src/utils/helper.ts"},
	}

	root := BuildTree(files)
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)
	if len(rootFolder.children) != 1 {
		t.Fatalf("expected 1 child (src), got %d", len(rootFolder.children))
	}

	// src should NOT be collapsed because it has two children
	srcFolder := rootFolder.children[0].(*FolderNode)
	if srcFolder.name != "src" {
		t.Errorf("expected folder name %q, got %q", "src", srcFolder.name)
	}

	// Should have 2 children (components and utils)
	if len(srcFolder.children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(srcFolder.children))
	}

	// Verify both children exist
	componentsFolder := srcFolder.children[0].(*FolderNode)
	if componentsFolder.name != "components" {
		t.Errorf("expected first child to be %q, got %q", "components", componentsFolder.name)
	}

	utilsFolder := srcFolder.children[1].(*FolderNode)
	if utilsFolder.name != "utils" {
		t.Errorf("expected second child to be %q, got %q", "utils", utilsFolder.name)
	}
}

// TestCollapsePaths_FolderWithOnlyFile verifies that folders whose only child
// is a file are NOT collapsed.
func TestCollapsePaths_FolderWithOnlyFile(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/App.tsx"},
	}

	root := BuildTree(files)
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)
	if len(rootFolder.children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(rootFolder.children))
	}

	// src should NOT be collapsed because its only child is a file
	srcFolder := rootFolder.children[0].(*FolderNode)
	if srcFolder.name != "src" {
		t.Errorf("expected folder name %q, got %q", "src", srcFolder.name)
	}

	// Should have 1 child (the file)
	if len(srcFolder.children) != 1 {
		t.Fatalf("expected 1 child (file), got %d", len(srcFolder.children))
	}

	file := srcFolder.children[0].(*FileNode)
	if file.name != "App.tsx" {
		t.Errorf("expected file name %q, got %q", "App.tsx", file.name)
	}
}

// TestCollapsePaths_MixedScenario tests a complex tree with both collapsible
// and non-collapsible paths.
func TestCollapsePaths_MixedScenario(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/components/nested/deep/App.tsx"}, // Should collapse src/components/nested/deep
		{Path: "src/utils/helper.ts"},                // src has multiple children, don't collapse
		{Path: "test/e2e/basic.test.ts"},             // Should collapse test/e2e
		{Path: "README.md"},                          // Root level file
	}

	root := BuildTree(files)
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)
	if len(rootFolder.children) != 3 {
		t.Fatalf("expected 3 children at root, got %d", len(rootFolder.children))
	}

	// Children are sorted: folders first, then files
	// First child: src (has multiple children, should NOT collapse)
	srcFolder := rootFolder.children[0].(*FolderNode)
	if srcFolder.name != "src" {
		t.Errorf("expected first child to be src, got %q", srcFolder.name)
	}
	if len(srcFolder.children) != 2 {
		t.Fatalf("expected src to have 2 children, got %d", len(srcFolder.children))
	}

	// src/components/nested/deep should be collapsed
	componentsFolder := srcFolder.children[0].(*FolderNode)
	expectedName := "components/nested/deep"
	if componentsFolder.name != expectedName {
		t.Errorf("expected collapsed folder name %q, got %q", expectedName, componentsFolder.name)
	}
	if len(componentsFolder.children) != 1 {
		t.Fatalf("expected components folder to have 1 child, got %d", len(componentsFolder.children))
	}

	// src/utils should NOT be collapsed (has only file)
	utilsFolder := srcFolder.children[1].(*FolderNode)
	if utilsFolder.name != "utils" {
		t.Errorf("expected utils folder, got %q", utilsFolder.name)
	}

	// Second child: test/e2e should be collapsed
	testFolder := rootFolder.children[1].(*FolderNode)
	expectedTestName := "test/e2e"
	if testFolder.name != expectedTestName {
		t.Errorf("expected collapsed folder name %q, got %q", expectedTestName, testFolder.name)
	}
	if len(testFolder.children) != 1 {
		t.Fatalf("expected test folder to have 1 child, got %d", len(testFolder.children))
	}

	// Third child: README.md (file at root)
	readmeFile := rootFolder.children[2].(*FileNode)
	if readmeFile.name != "README.md" {
		t.Errorf("expected third child to be README.md, got %q", readmeFile.name)
	}
}

// TestCollapsePaths_EmptyTree verifies that collapsing an empty tree is safe.
func TestCollapsePaths_EmptyTree(t *testing.T) {
	root := BuildTree([]core.FileChange{})
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)
	if len(rootFolder.children) != 0 {
		t.Errorf("expected empty tree, got %d children", len(rootFolder.children))
	}
}

// TestCollapsePaths_PreservesExpandedState verifies that the isExpanded
// state is preserved during collapsing.
func TestCollapsePaths_PreservesExpandedState(t *testing.T) {
	files := []core.FileChange{
		{Path: "src/components/App.tsx"},
	}

	root := BuildTree(files)

	// Manually collapse a folder before calling CollapsePaths
	rootFolder := root.(*FolderNode)
	srcFolder := rootFolder.children[0].(*FolderNode)
	srcFolder.isExpanded = false

	collapsed := CollapsePaths(root)

	// Verify the collapsed folder preserves the expanded state
	collapsedRoot := collapsed.(*FolderNode)
	collapsedSrc := collapsedRoot.children[0].(*FolderNode)

	// Should have collapsed src/components into one folder
	expectedName := "src/components"
	if collapsedSrc.name != expectedName {
		t.Errorf("expected collapsed folder name %q, got %q", expectedName, collapsedSrc.name)
	}

	// Should preserve the isExpanded state from the first folder in the chain
	if collapsedSrc.isExpanded != false {
		t.Errorf("expected isExpanded to be false, got true")
	}
}

// TestCollapsePaths_DepthAdjustment verifies that depths are correctly
// adjusted after collapsing.
func TestCollapsePaths_DepthAdjustment(t *testing.T) {
	files := []core.FileChange{
		{Path: "a/b/c/d/file.txt"},
	}

	root := BuildTree(files)
	collapsed := CollapsePaths(root)

	rootFolder := collapsed.(*FolderNode)

	// Should have one collapsed folder "a/b/c/d"
	folder := rootFolder.children[0].(*FolderNode)
	if folder.name != "a/b/c/d" {
		t.Errorf("expected collapsed folder name %q, got %q", "a/b/c/d", folder.name)
	}

	// Collapsed folder should be at depth 0 (first level below root)
	if folder.depth != 0 {
		t.Errorf("expected depth 0, got %d", folder.depth)
	}

	// File should be at depth 1
	file := folder.children[0].(*FileNode)
	if file.depth != 1 {
		t.Errorf("expected file depth 1, got %d", file.depth)
	}
}
