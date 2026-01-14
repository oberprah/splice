package tree

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestFlattenVisible_EmptyTree(t *testing.T) {
	root := &FolderNode{
		name:       "",
		depth:      -1,
		children:   []TreeNode{},
		isExpanded: true,
		stats:      FolderStats{},
	}

	result := FlattenVisible(root)

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty tree, got %d items", len(result))
	}
}

func TestFlattenVisible_SingleFile(t *testing.T) {
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FileNode{
				name:  "README.md",
				depth: 0,
				file:  &core.FileChange{Path: "README.md", Status: "M"},
			},
		},
	}

	result := FlattenVisible(root)

	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	item := result[0]
	fileNode, ok := item.Node.(*FileNode)
	if !ok {
		t.Fatalf("Expected FileNode, got %T", item.Node)
	}

	if fileNode.name != "README.md" {
		t.Errorf("Expected file name 'README.md', got '%s'", fileNode.name)
	}

	if !item.IsLastChild {
		t.Error("Expected IsLastChild to be true for only child")
	}

	if len(item.ParentLines) != 0 {
		t.Errorf("Expected empty ParentLines for depth 0, got %v", item.ParentLines)
	}
}

func TestFlattenVisible_AllExpanded(t *testing.T) {
	// Tree structure:
	// root
	//   ├── src/
	//   │   ├── components/
	//   │   │   └── App.tsx
	//   │   └── utils/
	//   │       └── helper.ts
	//   └── README.md
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FolderNode{
				name:       "src",
				depth:      0,
				isExpanded: true,
				children: []TreeNode{
					&FolderNode{
						name:       "components",
						depth:      1,
						isExpanded: true,
						children: []TreeNode{
							&FileNode{
								name:  "App.tsx",
								depth: 2,
								file:  &core.FileChange{Path: "src/components/App.tsx"},
							},
						},
					},
					&FolderNode{
						name:       "utils",
						depth:      1,
						isExpanded: true,
						children: []TreeNode{
							&FileNode{
								name:  "helper.ts",
								depth: 2,
								file:  &core.FileChange{Path: "src/utils/helper.ts"},
							},
						},
					},
				},
			},
			&FileNode{
				name:  "README.md",
				depth: 0,
				file:  &core.FileChange{Path: "README.md"},
			},
		},
	}

	result := FlattenVisible(root)

	// Expected order (depth-first):
	// 1. src/ (depth 0, not last child)
	// 2. components/ (depth 1, not last child of src)
	// 3. App.tsx (depth 2, last child of components)
	// 4. utils/ (depth 1, last child of src)
	// 5. helper.ts (depth 2, last child of utils)
	// 6. README.md (depth 0, last child of root)
	expectedCount := 6
	if len(result) != expectedCount {
		t.Fatalf("Expected %d items, got %d", expectedCount, len(result))
	}

	// Check order and names
	expectedNames := []string{"src", "components", "App.tsx", "utils", "helper.ts", "README.md"}
	for i, expectedName := range expectedNames {
		if result[i].Node.GetName() != expectedName {
			t.Errorf("Item %d: expected name '%s', got '%s'", i, expectedName, result[i].Node.GetName())
		}
	}

	// Check depths
	expectedDepths := []int{0, 1, 2, 1, 2, 0}
	for i, expectedDepth := range expectedDepths {
		if result[i].Node.GetDepth() != expectedDepth {
			t.Errorf("Item %d (%s): expected depth %d, got %d",
				i, result[i].Node.GetName(), expectedDepth, result[i].Node.GetDepth())
		}
	}

	// Check isLastChild flags
	expectedIsLastChild := []bool{false, false, true, true, true, true}
	for i, expected := range expectedIsLastChild {
		if result[i].IsLastChild != expected {
			t.Errorf("Item %d (%s): expected IsLastChild=%v, got %v",
				i, result[i].Node.GetName(), expected, result[i].IsLastChild)
		}
	}

	// Check parentLines for specific items
	// App.tsx (depth 2, parent src is not last, parent components is not last)
	// ParentLines should be [false, false] (two parent levels, neither is last)
	appTsxItem := result[2]
	if len(appTsxItem.ParentLines) != 2 {
		t.Errorf("App.tsx: expected 2 parent lines, got %d", len(appTsxItem.ParentLines))
	}
	if len(appTsxItem.ParentLines) == 2 {
		if appTsxItem.ParentLines[0] != false {
			t.Errorf("App.tsx: expected ParentLines[0]=false, got %v", appTsxItem.ParentLines[0])
		}
		if appTsxItem.ParentLines[1] != false {
			t.Errorf("App.tsx: expected ParentLines[1]=false, got %v", appTsxItem.ParentLines[1])
		}
	}

	// helper.ts (depth 2, parent src is not last, parent utils is last)
	// ParentLines should be [false, true]
	helperTsItem := result[4]
	if len(helperTsItem.ParentLines) != 2 {
		t.Errorf("helper.ts: expected 2 parent lines, got %d", len(helperTsItem.ParentLines))
	}
	if len(helperTsItem.ParentLines) == 2 {
		if helperTsItem.ParentLines[0] != false {
			t.Errorf("helper.ts: expected ParentLines[0]=false, got %v", helperTsItem.ParentLines[0])
		}
		if helperTsItem.ParentLines[1] != true {
			t.Errorf("helper.ts: expected ParentLines[1]=true, got %v", helperTsItem.ParentLines[1])
		}
	}
}

func TestFlattenVisible_CollapsedFolder(t *testing.T) {
	// Tree structure with collapsed folder:
	// root
	//   ├── src/ (collapsed)
	//   │   └── ... (should be hidden)
	//   └── README.md
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FolderNode{
				name:       "src",
				depth:      0,
				isExpanded: false, // Collapsed!
				children: []TreeNode{
					&FolderNode{
						name:       "components",
						depth:      1,
						isExpanded: true,
						children: []TreeNode{
							&FileNode{
								name:  "App.tsx",
								depth: 2,
								file:  &core.FileChange{Path: "src/components/App.tsx"},
							},
						},
					},
				},
			},
			&FileNode{
				name:  "README.md",
				depth: 0,
				file:  &core.FileChange{Path: "README.md"},
			},
		},
	}

	result := FlattenVisible(root)

	// Should only see: src/, README.md
	// components/ and App.tsx should be hidden
	if len(result) != 2 {
		t.Fatalf("Expected 2 items (src/ and README.md), got %d", len(result))
	}

	// Check that we got the right items
	if result[0].Node.GetName() != "src" {
		t.Errorf("Expected first item to be 'src', got '%s'", result[0].Node.GetName())
	}

	if result[1].Node.GetName() != "README.md" {
		t.Errorf("Expected second item to be 'README.md', got '%s'", result[1].Node.GetName())
	}
}

func TestFlattenVisible_MixedExpandedCollapsed(t *testing.T) {
	// Tree structure:
	// root
	//   ├── expanded/
	//   │   ├── file1.txt
	//   │   └── nested/ (collapsed)
	//   │       └── file2.txt (hidden)
	//   └── collapsed/ (collapsed)
	//       └── file3.txt (hidden)
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FolderNode{
				name:       "expanded",
				depth:      0,
				isExpanded: true,
				children: []TreeNode{
					&FileNode{
						name:  "file1.txt",
						depth: 1,
						file:  &core.FileChange{Path: "expanded/file1.txt"},
					},
					&FolderNode{
						name:       "nested",
						depth:      1,
						isExpanded: false, // Collapsed
						children: []TreeNode{
							&FileNode{
								name:  "file2.txt",
								depth: 2,
								file:  &core.FileChange{Path: "expanded/nested/file2.txt"},
							},
						},
					},
				},
			},
			&FolderNode{
				name:       "collapsed",
				depth:      0,
				isExpanded: false, // Collapsed
				children: []TreeNode{
					&FileNode{
						name:  "file3.txt",
						depth: 1,
						file:  &core.FileChange{Path: "collapsed/file3.txt"},
					},
				},
			},
		},
	}

	result := FlattenVisible(root)

	// Expected visible items:
	// 1. expanded/ (depth 0)
	// 2. file1.txt (depth 1)
	// 3. nested/ (depth 1, collapsed - children hidden)
	// 4. collapsed/ (depth 0, collapsed - children hidden)
	expectedCount := 4
	if len(result) != expectedCount {
		t.Fatalf("Expected %d items, got %d", expectedCount, len(result))
	}

	expectedNames := []string{"expanded", "file1.txt", "nested", "collapsed"}
	for i, expectedName := range expectedNames {
		if result[i].Node.GetName() != expectedName {
			t.Errorf("Item %d: expected name '%s', got '%s'", i, expectedName, result[i].Node.GetName())
		}
	}
}

func TestFlattenVisible_SkipsRootNode(t *testing.T) {
	// Verify that root node at depth -1 is not included in results
	root := &FolderNode{
		name:       "ROOT_SHOULD_NOT_APPEAR",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FileNode{
				name:  "file.txt",
				depth: 0,
				file:  &core.FileChange{Path: "file.txt"},
			},
		},
	}

	result := FlattenVisible(root)

	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	// Verify root is not in results
	for _, item := range result {
		if item.Node.GetDepth() == -1 {
			t.Error("Root node (depth -1) should not be in visible items")
		}
		if item.Node.GetName() == "ROOT_SHOULD_NOT_APPEAR" {
			t.Error("Root node should not be in visible items")
		}
	}
}

func TestFlattenVisible_ParentLinesDeepNesting(t *testing.T) {
	// Test parentLines computation for deeply nested structure
	// root
	//   └── a/
	//       └── b/
	//           └── c/
	//               └── file.txt
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FolderNode{
				name:       "a",
				depth:      0,
				isExpanded: true,
				children: []TreeNode{
					&FolderNode{
						name:       "b",
						depth:      1,
						isExpanded: true,
						children: []TreeNode{
							&FolderNode{
								name:       "c",
								depth:      2,
								isExpanded: true,
								children: []TreeNode{
									&FileNode{
										name:  "file.txt",
										depth: 3,
										file:  &core.FileChange{Path: "a/b/c/file.txt"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := FlattenVisible(root)

	// Should have: a/, b/, c/, file.txt
	if len(result) != 4 {
		t.Fatalf("Expected 4 items, got %d", len(result))
	}

	// file.txt should have 3 parent levels, all are last children
	// ParentLines should be [true, true, true]
	fileItem := result[3]
	if fileItem.Node.GetName() != "file.txt" {
		t.Fatalf("Expected last item to be 'file.txt', got '%s'", fileItem.Node.GetName())
	}

	expectedParentLines := []bool{true, true, true}
	if len(fileItem.ParentLines) != len(expectedParentLines) {
		t.Fatalf("Expected %d parent lines, got %d", len(expectedParentLines), len(fileItem.ParentLines))
	}

	for i, expected := range expectedParentLines {
		if fileItem.ParentLines[i] != expected {
			t.Errorf("ParentLines[%d]: expected %v, got %v", i, expected, fileItem.ParentLines[i])
		}
	}
}

func TestFlattenVisible_MultipleChildrenIsLastChild(t *testing.T) {
	// Test isLastChild computation with multiple children at same level
	// root
	//   ├── first/
	//   ├── second/
	//   └── third/
	root := &FolderNode{
		name:       "",
		depth:      -1,
		isExpanded: true,
		children: []TreeNode{
			&FolderNode{name: "first", depth: 0, isExpanded: true, children: []TreeNode{}},
			&FolderNode{name: "second", depth: 0, isExpanded: true, children: []TreeNode{}},
			&FolderNode{name: "third", depth: 0, isExpanded: true, children: []TreeNode{}},
		},
	}

	result := FlattenVisible(root)

	if len(result) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(result))
	}

	// First two should have IsLastChild=false, last one should have IsLastChild=true
	if result[0].IsLastChild {
		t.Error("First item should not be last child")
	}
	if result[1].IsLastChild {
		t.Error("Second item should not be last child")
	}
	if !result[2].IsLastChild {
		t.Error("Third item should be last child")
	}
}
