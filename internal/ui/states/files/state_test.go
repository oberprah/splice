package files

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/filetree"
)

func TestNew_BuildsTreeStructure(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)

	s := New(source, files)

	// Verify root exists
	if s.Root == nil {
		t.Fatal("Expected Root to be initialized")
	}

	// Verify root is a FolderNode
	rootFolder, ok := s.Root.(*filetree.FolderNode)
	if !ok {
		t.Fatalf("Expected Root to be *filetree.FolderNode, got %T", s.Root)
	}

	// Verify root is at depth -1 and expanded
	if rootFolder.GetDepth() != -1 {
		t.Errorf("Expected root depth -1, got %d", rootFolder.GetDepth())
	}
	if !rootFolder.IsExpanded() {
		t.Error("Expected root to be expanded")
	}

	// Verify root has children (src/ and README.md)
	if len(rootFolder.Children()) == 0 {
		t.Error("Expected root to have children")
	}
}

func TestNew_PopulatesVisibleItems(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)

	s := New(source, files)

	// Since all folders start expanded, VisibleItems should contain all nodes
	if len(s.VisibleItems) == 0 {
		t.Fatal("Expected VisibleItems to be populated")
	}

	// Should have folders and files visible
	// Expected structure (with collapsed paths):
	// - README.md
	// - src/
	//   - app.go
	//   - utils/
	//     - helper.go
	// Minimum: 1 README.md + 1 src + 1 app.go + 1 utils + 1 helper.go = 5 items
	// But collapsed paths might reduce this to: README.md, src/, app.go, utils/, helper.go
	if len(s.VisibleItems) < 3 {
		t.Errorf("Expected at least 3 visible items, got %d", len(s.VisibleItems))
	}
}

func TestNew_AllFoldersStartExpanded(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/components/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "src/utils/helper.go", Status: "A", Additions: 20, Deletions: 0},
	}
	source := createTestDiffSource(commit)

	s := New(source, files)

	// Walk through all visible items and check that all folders are expanded
	for _, item := range s.VisibleItems {
		if folder, ok := item.Node.(*filetree.FolderNode); ok {
			if !folder.IsExpanded() {
				t.Errorf("Expected folder %s to be expanded", folder.GetName())
			}
		}
	}
}

func TestNew_PreservesOriginalFiles(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
		{Path: "README.md", Status: "M", Additions: 5, Deletions: 2},
	}
	source := createTestDiffSource(commit)

	s := New(source, files)

	// Original Files array should be preserved
	if len(s.Files) != 2 {
		t.Errorf("Expected 2 files in Files array, got %d", len(s.Files))
	}
	if s.Files[0].Path != "src/app.go" {
		t.Errorf("Expected first file to be src/app.go, got %s", s.Files[0].Path)
	}
}

func TestNew_InitializesCursorAndViewport(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{
		{Path: "src/app.go", Status: "M", Additions: 10, Deletions: 5},
	}
	source := createTestDiffSource(commit)

	s := New(source, files)

	if s.Cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", s.Cursor)
	}
	if s.ViewportStart != 0 {
		t.Errorf("Expected ViewportStart to be 0, got %d", s.ViewportStart)
	}
}

func TestNew_EmptyFileList(t *testing.T) {
	commit := createTestCommit()
	files := []core.FileChange{}
	source := createTestDiffSource(commit)

	s := New(source, files)

	// Should still create a tree with just the root
	if s.Root == nil {
		t.Fatal("Expected Root to be initialized even with empty files")
	}

	// VisibleItems should be empty (root is at depth -1 and not included)
	if len(s.VisibleItems) != 0 {
		t.Errorf("Expected 0 visible items for empty files, got %d", len(s.VisibleItems))
	}
}
