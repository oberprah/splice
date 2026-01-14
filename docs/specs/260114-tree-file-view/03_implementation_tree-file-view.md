# Implementation: Tree File View

**Requirements:** `01_requirements_tree-file-view.md`
**Design:** `02_design_tree-file-view.md`

## Steps

### Step 1: Create tree data structures and BuildTree

**Goal:** Create the domain package with tree node types and the BuildTree function that constructs the hierarchical tree from a flat file list.

**Structure:**
- New package: `internal/domain/tree/`
- Files: `tree.go` (types), `build.go` (construction)
- Types: `TreeNode` interface, `FolderNode`, `FileNode`, `FolderStats`, `VisibleTreeItem`
- Function: `BuildTree(files []core.FileChange) TreeNode`

**Verify:**
- Unit tests for BuildTree with various file structures:
  - Single file at root
  - Multiple files in nested folders
  - Files at different depths
  - Folder/file sorting (folders first, alphabetical)
- All tests pass: `go test ./internal/domain/tree/...`

**Read:**
- `docs/specs/260114-tree-file-view/02_design_tree-file-view.md` (design doc)
- `internal/core/git_types.go` (FileChange type)

**Status:** ✅ Complete

**Implementation:**
- Created package `internal/domain/tree/` with three files:
  - `tree.go`: Defines `TreeNode` interface and sum types (`FolderNode`, `FileNode`)
  - `build.go`: Implements `BuildTree()` function
  - `build_test.go`: Comprehensive test coverage (6 test cases)
- All tree node types follow sum type pattern as specified in design
- `FolderNode` and `FileNode` implement `TreeNode` interface with marker method `isTreeNode()`
- `BuildTree()` creates root at depth -1, splits file paths, builds folder hierarchy
- Children sorted with folders first, then alphabetically (via `sortChildren()`)
- Helper functions: `insertFile()`, `getOrCreateFolder()`, `sortChildren()`, `isFolder()`
- Exported fields in `FolderStats` and `VisibleTreeItem` for future use in Steps 3-4

**Tests:**
All 6 tests pass:
- `TestBuildTree_EmptyInput`: Verifies empty tree structure
- `TestBuildTree_SingleFileAtRoot`: Single file handling
- `TestBuildTree_MultipleFilesInNestedFolders`: Complex nested structure with multiple files
- `TestBuildTree_FolderAndFileSorting`: Validates folder-first alphabetical sorting
- `TestBuildTree_FilesAtDifferentDepths`: Deep nesting (4+ levels)
- `TestBuildTree_PreservesFileChangeData`: Ensures FileChange data integrity

**Commit:** 3d4125d7c0b75fe33f843529bdf1264327c8a6fd

**Implementation decisions:**
- Exported `FolderStats` and `VisibleTreeItem` fields to avoid linter warnings about unused fields (will be used in Steps 3-4)
- Used `sort.SliceStable` to maintain order when sorting equals (though not strictly needed here)
- Path splitting uses standard library `strings.Split()` with "/" separator
- No path collapsing in this step (deferred to Step 2 as per design)

---

### Step 2: Implement path collapsing

**Goal:** Add path collapsing logic to merge single-child folder chains into collapsed paths.

**Structure:**
- File: `internal/domain/tree/collapse.go`
- Function: `CollapsePaths(root TreeNode) TreeNode`
- Mutates tree in-place by collapsing folder chains like `src/ → components/ → nested/` into single `FolderNode{name: "src/components/nested"}`

**Verify:**
- Unit tests for CollapsePaths:
  - Single-child folder chain collapses
  - Folders with multiple children don't collapse
  - Folders with files don't collapse (only child is a file)
  - Mixed scenarios
- Tests run: `go test ./internal/domain/tree/...`

**Read:**
- `internal/domain/tree/tree.go` (node types from Step 1)
- `internal/domain/tree/build.go` (tree structure from Step 1)

**Status:** ✅ Complete

**Implementation:**
- Created `internal/domain/tree/collapse.go` with `CollapsePaths()` function
- Created `internal/domain/tree/collapse_test.go` with comprehensive test coverage (7 test cases)
- `CollapsePaths()` recursively walks the tree and collapses single-child folder chains
- Helper functions:
  - `collapseFolder()`: Recursively processes folders bottom-up
  - `collapseChain()`: Collapses a chain starting from a folder, returns combined name and final children
  - `adjustDepths()`: Recursively adjusts depths after collapsing
- Algorithm preserves `isExpanded` state from the first folder in the chain
- Only collapses when folder has exactly one child that is also a folder (not a file)
- Folders with multiple children are never collapsed

**Tests:**
All 7 tests pass:
- `TestCollapsePaths_SingleChildFolderChain`: Verifies basic chain collapsing (src/components/nested)
- `TestCollapsePaths_FolderWithMultipleChildren`: Ensures folders with multiple children don't collapse
- `TestCollapsePaths_FolderWithOnlyFile`: Ensures folders whose only child is a file don't collapse
- `TestCollapsePaths_MixedScenario`: Complex tree with both collapsible and non-collapsible paths
- `TestCollapsePaths_EmptyTree`: Safety check for empty input
- `TestCollapsePaths_PreservesExpandedState`: Verifies isExpanded state preservation
- `TestCollapsePaths_DepthAdjustment`: Verifies correct depth adjustment after collapsing

**Commit:** bc4a88969e95dc3ae88f7e46ab2e6dd9c3e85e7b

**Implementation decisions:**
- Used bottom-up recursive approach to ensure all children are collapsed before parent
- Loop condition lifted into `for` statement for cleaner code (per staticcheck)
- Depths adjusted after collapsing to maintain correct tree structure
- Mutates tree in-place for efficiency (matches design doc)
- Stats field left empty (will be populated in Step 3)

---

### Step 3: Implement folder stats computation

**Goal:** Calculate aggregate statistics (file count, additions, deletions) for each folder in the tree.

**Structure:**
- File: `internal/domain/tree/stats.go`
- Function: `ComputeStats(node TreeNode) FolderStats`
- Recursively computes stats for collapsed folder display

**Verify:**
- Unit tests for ComputeStats:
  - File node returns stats from file change
  - Folder node sums all descendant file stats
  - Nested folders aggregate correctly
- Tests pass: `go test ./internal/domain/tree/...`

**Read:**
- `internal/domain/tree/tree.go` (FolderStats type)
- `internal/core/types.go` (FileChange stats fields)

**Status:** Pending

---

### Step 4: Implement tree flattening

**Goal:** Flatten the tree into a visible list with rendering metadata for cursor navigation.

**Structure:**
- File: `internal/domain/tree/flatten.go`
- Function: `FlattenVisible(root TreeNode) []VisibleTreeItem`
- Walks tree depth-first, computes `isLastChild` and `parentLines` for box-drawing characters

**Verify:**
- Unit tests for FlattenVisible:
  - All expanded: returns all nodes in depth-first order
  - Collapsed folder: skips children
  - Metadata computed correctly (isLastChild, parentLines)
- Tests pass: `go test ./internal/domain/tree/...`

**Read:**
- `internal/domain/tree/tree.go` (VisibleTreeItem type)

**Status:** Pending

---

### Step 5: Implement tree line formatting

**Goal:** Create formatting functions that render tree lines with box-drawing characters.

**Structure:**
- New file: `internal/ui/format/tree_line.go`
- Functions:
  - `FormatTreeLine(item VisibleTreeItem, isSelected bool, styles *styles.Styles) string`
  - Helper for box-drawing characters based on metadata

**Verify:**
- Unit tests for FormatTreeLine:
  - Correct tree characters (`├──`, `└──`, `│`) based on isLastChild
  - Proper indentation from parentLines
  - File vs folder formatting
  - Selection highlighting
- Golden file test with sample tree structure
- Tests pass: `go test ./internal/ui/format/...`

**Read:**
- `internal/domain/tree/tree.go` (VisibleTreeItem)
- `internal/ui/format/file_format.go` (existing file formatting patterns)
- `internal/ui/styles/styles.go` (style types)

**Status:** Pending

---

### Step 6: Create TreeSection component

**Goal:** Build the TreeSection component that renders the complete tree view.

**Structure:**
- New file: `internal/ui/components/tree_section.go`
- Function: `TreeSection(items []VisibleTreeItem, cursor int, viewport ViewportInfo, styles *styles.Styles) string`
- Uses FormatTreeLine for each visible item
- Applies viewport windowing and cursor highlighting

**Verify:**
- Golden file tests:
  - Tree with folders and files
  - Collapsed folders showing stats
  - Cursor on folder vs file
  - Viewport windowing with partial tree
- Tests pass: `go test ./internal/ui/components/...`

**Read:**
- `internal/ui/format/tree_line.go` (from Step 5)
- `internal/ui/components/file_section.go` (viewport patterns)
- `internal/ui/components/types.go` (ViewportInfo)

**Status:** Pending

---

### Step 7: Integrate tree into files state

**Goal:** Modify FilesState to use tree structure and add toggle functionality.

**Structure:**
- Modify: `internal/ui/states/files/state.go`
  - Add fields: `Root TreeNode`, `VisibleItems []VisibleTreeItem`
  - Update constructor to build tree from files
- Modify: `internal/ui/states/files/update.go`
  - Add toggle handlers (Enter/Space/Left/Right on folders)
  - Add helper: `toggleFolder(s *State, expandOnly bool, collapseOnly bool) (*State, tea.Cmd)`
  - Uses deep copy, toggles isExpanded, re-flattens

**Verify:**
- Unit tests for state construction:
  - Tree built from files on state creation
  - VisibleItems populated correctly
- Unit tests for toggleFolder:
  - Enter/Space toggle
  - Right arrow expands only
  - Left arrow collapses only
  - No-op on file nodes
- Integration test: navigate and toggle, verify state changes
- Tests pass: `go test ./internal/ui/states/files/...`

**Read:**
- `internal/ui/states/files/state.go` (current state structure)
- `internal/ui/states/files/update.go` (current navigation handlers)
- `internal/domain/tree/` (all tree functions from Steps 1-4)

**Status:** Pending

---

### Step 8: Update files view to render tree

**Goal:** Replace flat file list rendering with tree rendering.

**Structure:**
- Modify: `internal/ui/states/files/view.go`
  - Use `TreeSection(s.VisibleItems, s.Cursor, ...)` instead of file list rendering
  - Keep header and help text formatting

**Verify:**
- Golden file tests for files state:
  - Tree view with expanded folders
  - Tree view with collapsed folders
  - Cursor highlighting on folders/files
  - Update existing golden files
- Tests pass: `go test ./internal/ui/states/files/...`
- Visual inspection: `go run . --help` (or test repo)

**Read:**
- `internal/ui/states/files/view.go` (current rendering)
- `internal/ui/components/tree_section.go` (from Step 6)

**Status:** Pending

---

### Step 9: E2E testing

**Goal:** Create end-to-end tests for the complete tree file view feature.

**Structure:**
- New test: `test/e2e/tree_navigation_test.go`
- Scenarios:
  - Navigate through tree (up/down)
  - Toggle folders (expand/collapse)
  - Enter file to view diff
  - Collapsed path expansion
  - Back from diff returns to tree

**Verify:**
- All E2E tests pass: `go test ./test/e2e/...`
- Manual testing with tape-runner: `./run-tape test/tapes/tree-navigation.tape`

**Read:**
- `test/e2e/` (existing E2E test patterns)
- `docs/specs/260114-tree-file-view/01_requirements_tree-file-view.md` (acceptance criteria)

**Status:** Pending

---

## Final Verification

- [ ] Full test suite passes: `go test ./...`
- [ ] All golden files reviewed and appropriate
- [ ] All requirements from FR1-FR7 verified
- [ ] Design decisions (sum types, tree structure, flattening) followed
- [ ] Manual testing with tape-runner successful

## Summary

(To be completed after implementation)
