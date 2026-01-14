package components

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/tree"
	"github.com/oberprah/splice/internal/ui/testutils"
)

func assertTreeSectionGolden(t *testing.T, output string, filename string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", filename)
	testutils.AssertGolden(t, output, goldenPath, *update)
}

// TestTreeSection_EmptyTree tests rendering an empty tree
func TestTreeSection_EmptyTree(t *testing.T) {
	testutils.SetupColorProfile()

	items := []tree.VisibleTreeItem{}
	files := []core.FileChange{}
	cursor := 0

	lines := TreeSection(items, files, cursor, 80)

	// Should have: blank line + stats line = 2 total (no tree items)
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines for empty tree, got %d", len(lines))
	}

	statsLine := ansi.Strip(lines[1])
	if !strings.Contains(statsLine, "0 files") {
		t.Errorf("Expected '0 files' in stats line, got: %q", statsLine)
	}
}

// TestTreeSection_SingleFile tests rendering a tree with a single file at root
func TestTreeSection_SingleFile(t *testing.T) {
	testutils.SetupColorProfile()

	file := &core.FileChange{
		Path:      "README.md",
		Status:    "M",
		Additions: 10,
		Deletions: 5,
		IsBinary:  false,
	}
	fileNode := tree.NewFileNode("README.md", 0, file)

	items := []tree.VisibleTreeItem{
		{
			Node:        fileNode,
			IsLastChild: true,
			ParentLines: []bool{},
		},
	}
	files := []core.FileChange{*file}
	cursor := 0

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_single_file.golden")
}

// TestTreeSection_FolderAndFiles tests rendering a tree with folders and files
func TestTreeSection_FolderAndFiles(t *testing.T) {
	testutils.SetupColorProfile()

	// Create test data: src/ (expanded) with two files
	file1 := &core.FileChange{Path: "src/App.tsx", Status: "M", Additions: 17, Deletions: 13, IsBinary: false}
	file2 := &core.FileChange{Path: "src/index.ts", Status: "A", Additions: 42, Deletions: 0, IsBinary: false}

	srcFolder := tree.NewFolderNode("src/", 0, true, tree.FolderStats{FileCount: 2, Additions: 59, Deletions: 13})
	fileNode1 := tree.NewFileNode("App.tsx", 1, file1)
	fileNode2 := tree.NewFileNode("index.ts", 1, file2)

	items := []tree.VisibleTreeItem{
		{
			Node:        srcFolder,
			IsLastChild: true,
			ParentLines: []bool{},
		},
		{
			Node:        fileNode1,
			IsLastChild: false,
			ParentLines: []bool{false}, // Parent is not last child
		},
		{
			Node:        fileNode2,
			IsLastChild: true,
			ParentLines: []bool{false},
		},
	}
	files := []core.FileChange{*file1, *file2}
	cursor := 0

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_folder_and_files.golden")
}

// TestTreeSection_CollapsedFolder tests rendering a collapsed folder with stats
func TestTreeSection_CollapsedFolder(t *testing.T) {
	testutils.SetupColorProfile()

	// Collapsed folder shows stats inline
	// Create dummy files to match the folder stats
	files := []core.FileChange{
		{Path: "old/file1.go", Status: "M", Additions: 20, Deletions: 10, IsBinary: false},
		{Path: "old/file2.go", Status: "M", Additions: 15, Deletions: 8, IsBinary: false},
		{Path: "old/file3.go", Status: "M", Additions: 15, Deletions: 7, IsBinary: false},
	}
	oldFolder := tree.NewFolderNode("old/", 0, false, tree.FolderStats{FileCount: 3, Additions: 50, Deletions: 25})

	items := []tree.VisibleTreeItem{
		{
			Node:        oldFolder,
			IsLastChild: true,
			ParentLines: []bool{},
		},
	}
	cursor := 0

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_collapsed_folder.golden")
}

// TestTreeSection_NestedFolders tests rendering a deeper tree structure
func TestTreeSection_NestedFolders(t *testing.T) {
	testutils.SetupColorProfile()

	// Tree structure:
	// src/
	//   ├── components/
	//   │   └── App.tsx
	//   └── utils/
	//       └── helper.ts
	file1 := &core.FileChange{Path: "src/components/App.tsx", Status: "M", Additions: 17, Deletions: 13, IsBinary: false}
	file2 := &core.FileChange{Path: "src/utils/helper.ts", Status: "A", Additions: 42, Deletions: 0, IsBinary: false}

	srcFolder := tree.NewFolderNode("src/", 0, true, tree.FolderStats{FileCount: 2, Additions: 59, Deletions: 13})
	componentsFolder := tree.NewFolderNode("components/", 1, true, tree.FolderStats{FileCount: 1, Additions: 17, Deletions: 13})
	utilsFolder := tree.NewFolderNode("utils/", 1, true, tree.FolderStats{FileCount: 1, Additions: 42, Deletions: 0})
	fileNode1 := tree.NewFileNode("App.tsx", 2, file1)
	fileNode2 := tree.NewFileNode("helper.ts", 2, file2)

	items := []tree.VisibleTreeItem{
		{
			Node:        srcFolder,
			IsLastChild: true,
			ParentLines: []bool{},
		},
		{
			Node:        componentsFolder,
			IsLastChild: false,
			ParentLines: []bool{false}, // src has more children
		},
		{
			Node:        fileNode1,
			IsLastChild: true,
			ParentLines: []bool{false, false}, // Both parents have more siblings
		},
		{
			Node:        utilsFolder,
			IsLastChild: true,
			ParentLines: []bool{false},
		},
		{
			Node:        fileNode2,
			IsLastChild: true,
			ParentLines: []bool{false, true}, // utils is last child of src
		},
	}
	files := []core.FileChange{*file1, *file2}
	cursor := 0

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_nested_folders.golden")
}

// TestTreeSection_CursorOnFolder tests cursor highlighting on a folder
func TestTreeSection_CursorOnFolder(t *testing.T) {
	testutils.SetupColorProfile()

	srcFolder := tree.NewFolderNode("src/", 0, true, tree.FolderStats{FileCount: 1, Additions: 10, Deletions: 5})
	file1 := &core.FileChange{Path: "src/App.tsx", Status: "M", Additions: 10, Deletions: 5, IsBinary: false}
	fileNode1 := tree.NewFileNode("App.tsx", 1, file1)

	items := []tree.VisibleTreeItem{
		{
			Node:        srcFolder,
			IsLastChild: true,
			ParentLines: []bool{},
		},
		{
			Node:        fileNode1,
			IsLastChild: true,
			ParentLines: []bool{false},
		},
	}
	files := []core.FileChange{*file1}
	cursor := 0 // Cursor on folder

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_cursor_on_folder.golden")
}

// TestTreeSection_CursorOnFile tests cursor highlighting on a file
func TestTreeSection_CursorOnFile(t *testing.T) {
	testutils.SetupColorProfile()

	srcFolder := tree.NewFolderNode("src/", 0, true, tree.FolderStats{FileCount: 1, Additions: 10, Deletions: 5})
	file1 := &core.FileChange{Path: "src/App.tsx", Status: "M", Additions: 10, Deletions: 5, IsBinary: false}
	fileNode1 := tree.NewFileNode("App.tsx", 1, file1)

	items := []tree.VisibleTreeItem{
		{
			Node:        srcFolder,
			IsLastChild: true,
			ParentLines: []bool{},
		},
		{
			Node:        fileNode1,
			IsLastChild: true,
			ParentLines: []bool{false},
		},
	}
	files := []core.FileChange{*file1}
	cursor := 1 // Cursor on file

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_cursor_on_file.golden")
}

// TestTreeSection_MixedStatuses tests files with different statuses
func TestTreeSection_MixedStatuses(t *testing.T) {
	testutils.SetupColorProfile()

	file1 := &core.FileChange{Path: "added.go", Status: "A", Additions: 100, Deletions: 0, IsBinary: false}
	file2 := &core.FileChange{Path: "modified.go", Status: "M", Additions: 20, Deletions: 10, IsBinary: false}
	file3 := &core.FileChange{Path: "deleted.go", Status: "D", Additions: 0, Deletions: 50, IsBinary: false}
	file4 := &core.FileChange{Path: "renamed.go", Status: "R", Additions: 0, Deletions: 0, IsBinary: false}

	items := []tree.VisibleTreeItem{
		{Node: tree.NewFileNode("added.go", 0, file1), IsLastChild: false, ParentLines: []bool{}},
		{Node: tree.NewFileNode("modified.go", 0, file2), IsLastChild: false, ParentLines: []bool{}},
		{Node: tree.NewFileNode("deleted.go", 0, file3), IsLastChild: false, ParentLines: []bool{}},
		{Node: tree.NewFileNode("renamed.go", 0, file4), IsLastChild: true, ParentLines: []bool{}},
	}
	files := []core.FileChange{*file1, *file2, *file3, *file4}
	cursor := 1 // Select modified file

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	assertTreeSectionGolden(t, output, "tree_section_mixed_statuses.golden")
}

// TestTreeSection_BinaryFile tests rendering a binary file
func TestTreeSection_BinaryFile(t *testing.T) {
	testutils.SetupColorProfile()

	binaryFile := &core.FileChange{Path: "image.png", Status: "A", Additions: 0, Deletions: 0, IsBinary: true}
	items := []tree.VisibleTreeItem{
		{Node: tree.NewFileNode("image.png", 0, binaryFile), IsLastChild: true, ParentLines: []bool{}},
	}
	files := []core.FileChange{*binaryFile}
	cursor := 0

	output := strings.Join(TreeSection(items, files, cursor, 80), "\n")
	plainOutput := ansi.Strip(output)

	if !strings.Contains(plainOutput, "(binary)") {
		t.Errorf("Expected '(binary)' marker in output, got: %q", plainOutput)
	}
}

// TestTreeSection_StatsCalculation tests that total stats are calculated correctly
func TestTreeSection_StatsCalculation(t *testing.T) {
	testutils.SetupColorProfile()

	file1 := &core.FileChange{Path: "file1.go", Status: "A", Additions: 100, Deletions: 0, IsBinary: false}
	file2 := &core.FileChange{Path: "file2.go", Status: "M", Additions: 20, Deletions: 30, IsBinary: false}
	file3 := &core.FileChange{Path: "file3.go", Status: "D", Additions: 0, Deletions: 50, IsBinary: false}

	items := []tree.VisibleTreeItem{
		{Node: tree.NewFileNode("file1.go", 0, file1), IsLastChild: false, ParentLines: []bool{}},
		{Node: tree.NewFileNode("file2.go", 0, file2), IsLastChild: false, ParentLines: []bool{}},
		{Node: tree.NewFileNode("file3.go", 0, file3), IsLastChild: true, ParentLines: []bool{}},
	}
	files := []core.FileChange{*file1, *file2, *file3}
	cursor := 0

	lines := TreeSection(items, files, cursor, 80)
	statsLine := ansi.Strip(lines[1])

	// Total: +120 -80
	if !strings.Contains(statsLine, "+120") {
		t.Errorf("Expected '+120' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "-80") {
		t.Errorf("Expected '-80' in stats line, got: %q", statsLine)
	}
	if !strings.Contains(statsLine, "3 files") {
		t.Errorf("Expected '3 files' in stats line, got: %q", statsLine)
	}
}
