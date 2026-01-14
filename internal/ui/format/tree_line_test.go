package format

import (
	"flag"
	"strings"
	"testing"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/tree"
	"github.com/oberprah/splice/internal/ui/testutils"
)

var updateGolden = flag.Bool("update", false, "update golden files")

// TestFormatTreeLine_FileNode tests formatting of file nodes
func TestFormatTreeLine_FileNode(t *testing.T) {
	testutils.SetupColorProfile()

	tests := []struct {
		name       string
		item       tree.VisibleTreeItem
		isSelected bool
		contains   []string
	}{
		{
			name: "file at root, selected",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("README.md", 0, &core.FileChange{
					Path:      "README.md",
					Status:    "M",
					Additions: 10,
					Deletions: 5,
					IsBinary:  false,
				}),
				IsLastChild: true,
				ParentLines: []bool{},
			},
			isSelected: true,
			contains:   []string{"→", "└──", "M", "+10", "-5", "README.md"},
		},
		{
			name: "file at root, not selected",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("README.md", 0, &core.FileChange{
					Path:      "README.md",
					Status:    "A",
					Additions: 20,
					Deletions: 0,
					IsBinary:  false,
				}),
				IsLastChild: true,
				ParentLines: []bool{},
			},
			isSelected: false,
			contains:   []string{"└──", "A", "+20", "-0", "README.md"},
		},
		{
			name: "file nested, not last child",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("App.tsx", 2, &core.FileChange{
					Path:      "src/components/App.tsx",
					Status:    "M",
					Additions: 17,
					Deletions: 13,
					IsBinary:  false,
				}),
				IsLastChild: false,
				ParentLines: []bool{false, true},
			},
			isSelected: false,
			contains:   []string{"    ", "│   ", "├──", "M", "+17", "-13", "App.tsx"},
		},
		{
			name: "file nested, last child",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("helper.ts", 2, &core.FileChange{
					Path:      "src/utils/helper.ts",
					Status:    "A",
					Additions: 42,
					Deletions: 0,
					IsBinary:  false,
				}),
				IsLastChild: true,
				ParentLines: []bool{false, true},
			},
			isSelected: false,
			contains:   []string{"    ", "│   ", "└──", "A", "+42", "-0", "helper.ts"},
		},
		{
			name: "binary file",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("image.png", 1, &core.FileChange{
					Path:      "assets/image.png",
					Status:    "A",
					Additions: 0,
					Deletions: 0,
					IsBinary:  true,
				}),
				IsLastChild: true,
				ParentLines: []bool{false},
			},
			isSelected: false,
			contains:   []string{"    ", "└──", "A", "(binary)", "image.png"},
		},
		{
			name: "deleted file",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("old.js", 1, &core.FileChange{
					Path:      "src/old.js",
					Status:    "D",
					Additions: 0,
					Deletions: 100,
					IsBinary:  false,
				}),
				IsLastChild: false,
				ParentLines: []bool{false},
			},
			isSelected: false,
			contains:   []string{"    ", "├──", "D", "+0", "-100", "old.js"},
		},
		{
			name: "renamed file",
			item: tree.VisibleTreeItem{
				Node: tree.NewFileNode("new_name.js", 0, &core.FileChange{
					Path:      "new_name.js",
					Status:    "R",
					Additions: 5,
					Deletions: 5,
					IsBinary:  false,
				}),
				IsLastChild: true,
				ParentLines: []bool{},
			},
			isSelected: false,
			contains:   []string{"└──", "R", "+5", "-5", "new_name.js"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTreeLine(tt.item, tt.isSelected)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected result to contain %q, got:\n%s", substr, result)
				}
			}
		})
	}
}

// TestFormatTreeLine_FolderNode tests formatting of folder nodes
func TestFormatTreeLine_FolderNode(t *testing.T) {
	testutils.SetupColorProfile()

	tests := []struct {
		name       string
		item       tree.VisibleTreeItem
		isSelected bool
		contains   []string
		notContain []string
	}{
		{
			name: "expanded folder at root",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("src/", 0, true, tree.FolderStats{
					FileCount: 5,
					Additions: 234,
					Deletions: 67,
				}),
				IsLastChild: false,
				ParentLines: []bool{},
			},
			isSelected: false,
			contains:   []string{"├──", "src/"},
			notContain: []string{"+234", "-67", "(5 files)"},
		},
		{
			name: "expanded folder selected",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("components/", 1, true, tree.FolderStats{
					FileCount: 3,
					Additions: 100,
					Deletions: 50,
				}),
				IsLastChild: false,
				ParentLines: []bool{false},
			},
			isSelected: true,
			contains:   []string{"→", "    ", "├──", "components/"},
			notContain: []string{"+100", "-50", "(3 files)"},
		},
		{
			name: "collapsed folder shows stats",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("old/", 0, false, tree.FolderStats{
					FileCount: 3,
					Additions: 50,
					Deletions: 25,
				}),
				IsLastChild: true,
				ParentLines: []bool{},
			},
			isSelected: false,
			contains:   []string{"└──", "old/", "+50", "-25", "(3 files)"},
		},
		{
			name: "collapsed folder selected shows stats",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("legacy/", 1, false, tree.FolderStats{
					FileCount: 10,
					Additions: 500,
					Deletions: 300,
				}),
				IsLastChild: true,
				ParentLines: []bool{false},
			},
			isSelected: true,
			contains:   []string{"→", "    ", "└──", "legacy/", "+500", "-300", "(10 files)"},
		},
		{
			name: "collapsed path (merged folders)",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("src/components/nested/", 0, false, tree.FolderStats{
					FileCount: 2,
					Additions: 30,
					Deletions: 10,
				}),
				IsLastChild: true,
				ParentLines: []bool{},
			},
			isSelected: false,
			contains:   []string{"└──", "src/components/nested/", "+30", "-10", "(2 files)"},
		},
		{
			name: "folder with deep nesting",
			item: tree.VisibleTreeItem{
				Node: tree.NewFolderNode("deeply/", 3, true, tree.FolderStats{
					FileCount: 1,
					Additions: 5,
					Deletions: 2,
				}),
				IsLastChild: true,
				ParentLines: []bool{false, true, false},
			},
			isSelected: false,
			contains:   []string{"    ", "│   ", "    ", "└──", "deeply/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTreeLine(tt.item, tt.isSelected)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected result to contain %q, got:\n%s", substr, result)
				}
			}

			for _, substr := range tt.notContain {
				if strings.Contains(result, substr) {
					t.Errorf("Expected result NOT to contain %q, got:\n%s", substr, result)
				}
			}
		})
	}
}

// TestFormatTreeLine_TreeCharacters tests the box-drawing character logic
func TestFormatTreeLine_TreeCharacters(t *testing.T) {
	testutils.SetupColorProfile()

	tests := []struct {
		name        string
		depth       int
		isLastChild bool
		parentLines []bool
		wantBranch  string   // The branch character we expect (├── or └──)
		wantPrefix  []string // The prefix parts before the branch
	}{
		{
			name:        "root level, last child",
			depth:       0,
			isLastChild: true,
			parentLines: []bool{},
			wantBranch:  "└──",
			wantPrefix:  []string{},
		},
		{
			name:        "root level, not last child",
			depth:       0,
			isLastChild: false,
			parentLines: []bool{},
			wantBranch:  "├──",
			wantPrefix:  []string{},
		},
		{
			name:        "depth 1, last child, parent has more siblings",
			depth:       1,
			isLastChild: true,
			parentLines: []bool{true}, // Parent is not last child
			wantBranch:  "└──",
			wantPrefix:  []string{"│   "},
		},
		{
			name:        "depth 1, not last child, parent is last child",
			depth:       1,
			isLastChild: false,
			parentLines: []bool{false}, // Parent is last child
			wantBranch:  "├──",
			wantPrefix:  []string{"    "},
		},
		{
			name:        "depth 2, complex parentLines",
			depth:       2,
			isLastChild: false,
			parentLines: []bool{false, true}, // Grandparent is last, parent is not
			wantBranch:  "├──",
			wantPrefix:  []string{"    ", "│   "},
		},
		{
			name:        "depth 3, all parents not last",
			depth:       3,
			isLastChild: true,
			parentLines: []bool{true, true, true},
			wantBranch:  "└──",
			wantPrefix:  []string{"│   ", "│   ", "│   "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple file node for testing
			item := tree.VisibleTreeItem{
				Node: tree.NewFileNode("test.txt", tt.depth, &core.FileChange{
					Path:      "test.txt",
					Status:    "M",
					Additions: 1,
					Deletions: 1,
				}),
				IsLastChild: tt.isLastChild,
				ParentLines: tt.parentLines,
			}

			result := FormatTreeLine(item, false)

			// Check for branch character
			if !strings.Contains(result, tt.wantBranch) {
				t.Errorf("Expected branch character %q, got:\n%s", tt.wantBranch, result)
			}

			// Check for prefix parts
			for _, prefix := range tt.wantPrefix {
				if !strings.Contains(result, prefix) {
					t.Errorf("Expected prefix to contain %q, got:\n%s", prefix, result)
				}
			}
		})
	}
}

// TestFormatTreeLine_SelectionHighlighting tests that selection changes styling
func TestFormatTreeLine_SelectionHighlighting(t *testing.T) {
	testutils.SetupColorProfile()

	item := tree.VisibleTreeItem{
		Node: tree.NewFileNode("test.js", 0, &core.FileChange{
			Path:      "test.js",
			Status:    "M",
			Additions: 10,
			Deletions: 5,
		}),
		IsLastChild: true,
		ParentLines: []bool{},
	}

	selectedResult := FormatTreeLine(item, true)
	unselectedResult := FormatTreeLine(item, false)

	// Selected should have the arrow selector
	if !strings.HasPrefix(selectedResult, "→") {
		t.Errorf("Expected selected line to start with '→', got: %s", selectedResult)
	}

	// Unselected should have spaces
	if !strings.HasPrefix(unselectedResult, " ") {
		t.Errorf("Expected unselected line to start with space, got: %s", unselectedResult)
	}

	// Both should contain the same structural elements
	if !strings.Contains(selectedResult, "test.js") {
		t.Error("Selected result should contain filename")
	}
	if !strings.Contains(unselectedResult, "test.js") {
		t.Error("Unselected result should contain filename")
	}
}

// TestFormatTreeLine_Golden is a golden file test with a sample tree structure
func TestFormatTreeLine_Golden(t *testing.T) {
	testutils.SetupColorProfile()

	// Create a sample tree structure matching the design doc example:
	// → src/
	//   ├── components/
	//   │   └── M +17 -13  App.tsx
	//   └── utils/
	//       └── A +42  -0  helper.ts
	//   old/ +50 -25 (3 files)

	items := []tree.VisibleTreeItem{
		{
			Node: tree.NewFolderNode("src/", 0, true, tree.FolderStats{
				FileCount: 2,
				Additions: 59,
				Deletions: 13,
			}),
			IsLastChild: false,
			ParentLines: []bool{},
		},
		{
			Node: tree.NewFolderNode("components/", 1, true, tree.FolderStats{
				FileCount: 1,
				Additions: 17,
				Deletions: 13,
			}),
			IsLastChild: false,
			ParentLines: []bool{false},
		},
		{
			Node: tree.NewFileNode("App.tsx", 2, &core.FileChange{
				Path:      "src/components/App.tsx",
				Status:    "M",
				Additions: 17,
				Deletions: 13,
			}),
			IsLastChild: true,
			ParentLines: []bool{false, false},
		},
		{
			Node: tree.NewFolderNode("utils/", 1, true, tree.FolderStats{
				FileCount: 1,
				Additions: 42,
				Deletions: 0,
			}),
			IsLastChild: true,
			ParentLines: []bool{false},
		},
		{
			Node: tree.NewFileNode("helper.ts", 2, &core.FileChange{
				Path:      "src/utils/helper.ts",
				Status:    "A",
				Additions: 42,
				Deletions: 0,
			}),
			IsLastChild: true,
			ParentLines: []bool{false, true},
		},
		{
			Node: tree.NewFolderNode("old/", 0, false, tree.FolderStats{
				FileCount: 3,
				Additions: 50,
				Deletions: 25,
			}),
			IsLastChild: true,
			ParentLines: []bool{},
		},
	}

	var output strings.Builder
	for i, item := range items {
		isSelected := i == 0 // First item selected
		line := FormatTreeLine(item, isSelected)
		output.WriteString(line)
		output.WriteString("\n")
	}

	testutils.AssertGolden(t, output.String(), "tree_line_sample.golden", *updateGolden)
}

// TestFormatTreeLine_FolderNameFormatting tests folder name rendering
func TestFormatTreeLine_FolderNameFormatting(t *testing.T) {
	testutils.SetupColorProfile()

	tests := []struct {
		name       string
		folderName string
		contains   string
	}{
		{
			name:       "simple folder name",
			folderName: "src/",
			contains:   "src/",
		},
		{
			name:       "collapsed path",
			folderName: "src/components/nested/",
			contains:   "src/components/nested/",
		},
		{
			name:       "folder with special chars",
			folderName: "my-folder_v2/",
			contains:   "my-folder_v2/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tree.VisibleTreeItem{
				Node:        tree.NewFolderNode(tt.folderName, 0, true, tree.FolderStats{}),
				IsLastChild: true,
				ParentLines: []bool{},
			}

			result := FormatTreeLine(item, false)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain folder name %q, got:\n%s", tt.contains, result)
			}
		})
	}
}

// TestFormatTreeLine_StatsFormatting tests the stats display for collapsed folders
func TestFormatTreeLine_StatsFormatting(t *testing.T) {
	testutils.SetupColorProfile()

	tests := []struct {
		name     string
		stats    tree.FolderStats
		contains []string
	}{
		{
			name: "single file",
			stats: tree.FolderStats{
				FileCount: 1,
				Additions: 10,
				Deletions: 5,
			},
			contains: []string{"+10", "-5", "(1 file)"},
		},
		{
			name: "multiple files",
			stats: tree.FolderStats{
				FileCount: 5,
				Additions: 234,
				Deletions: 67,
			},
			contains: []string{"+234", "-67", "(5 files)"},
		},
		{
			name: "zero stats",
			stats: tree.FolderStats{
				FileCount: 2,
				Additions: 0,
				Deletions: 0,
			},
			contains: []string{"+0", "-0", "(2 files)"},
		},
		{
			name: "only additions",
			stats: tree.FolderStats{
				FileCount: 1,
				Additions: 100,
				Deletions: 0,
			},
			contains: []string{"+100", "-0", "(1 file)"},
		},
		{
			name: "only deletions",
			stats: tree.FolderStats{
				FileCount: 3,
				Additions: 0,
				Deletions: 200,
			},
			contains: []string{"+0", "-200", "(3 files)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tree.VisibleTreeItem{
				Node:        tree.NewFolderNode("folder/", 0, false, tt.stats),
				IsLastChild: true,
				ParentLines: []bool{},
			}

			result := FormatTreeLine(item, false)

			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("Expected result to contain %q, got:\n%s", substr, result)
				}
			}
		})
	}
}
