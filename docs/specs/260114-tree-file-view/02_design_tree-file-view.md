# Design: Tree File View

**Status**: Draft (awaiting approval)
**Created**: 2026-01-14

## Executive Summary

Replace the flat file list with a hierarchical tree view. Build an explicit tree structure from file paths, maintain expand/collapse state in tree nodes, and flatten to a visible list for cursor navigation and rendering. This matches Splice's git graph pattern (build once, store, render by traversal) with O(1) navigation and O(viewport) rendering. Use proper sum types (FolderNode/FileNode) to make illegal states unrepresentable.

## Context & Problem Statement

Current file view displays a flat list with full paths, creating problems for commits with many files in deep hierarchies: hard to scan, no structure visibility, high cognitive load from repeated path prefixes.

**Scope**: Tree data structure, expand/collapse, cursor navigation, box-drawing character rendering.
**Out of scope**: Search/filtering, state persistence across commits, bulk operations.

## Current State

Files stored as flat `[]core.FileChange` array with cursor indexing directly into it. Each file renders with full path (`src/components/App.tsx`). Navigation via up/down through flat list.

## Solution

### Architecture Decision

> **Decision:** Tree Structure with Flattening (over virtual tree computation or caching).
>
> **Rationale:** Matches git graph pattern (build layout once, store, render). Predictable performance: O(1) navigation, O(viewport) rendering, O(n) toggle (acceptable - user action only). Avoids virtual tree's O(n) per-render cost and cache invalidation complexity.
>
> **Tradeoff:** More memory (tree + flattened list), deep copy on toggle. Acceptable for typical commit sizes (<100 files).

### Data Model

**Sum type for tree nodes** (makes illegal states unrepresentable):

```go
// Tree node interface
type TreeNode interface {
    isTreeNode()
    GetName() string
    GetDepth() int
}

// Folder node
type FolderNode struct {
    name       string        // May be collapsed path: "src/components/nested"
    depth      int
    children   []TreeNode    // Sorted: folders first, then files (alphabetical)
    isExpanded bool
    stats      FolderStats   // Aggregate stats for collapsed display
}

// File node
type FileNode struct {
    name  string              // Just filename: "App.tsx"
    depth int
    file  *core.FileChange    // Pointer to original file data
}

type FolderStats struct {
    fileCount  int
    additions  int
    deletions  int
}
```

**Files state:**

```go
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange    // Original flat list
    Root          TreeNode             // Tree root (FolderNode at depth -1)
    VisibleItems  []VisibleTreeItem    // Flattened for navigation
    Cursor        int
    ViewportStart int
}

// Rendering metadata computed during flattening
type VisibleTreeItem struct {
    node        TreeNode
    isLastChild bool
    parentLines []bool  // Which depth levels need │ continuation
}
```

### Core Algorithms

**Build tree:**
1. Create root FolderNode (depth -1, expanded)
2. For each file: split path, create/navigate FolderNodes, add FileNode
3. Sort all children (folders first, alphabetical)
4. Collapse single-child folder paths (`src/` → `components/` → `nested/` becomes `FolderNode{name: "src/components/nested"}`)
5. Compute stats for all folders

**Flatten visible:**
```
Walk(node):
    if depth >= 0: append to result with rendering metadata
    if node is FolderNode and expanded:
        for each child: Walk(child)
```

**Toggle folder:**
1. Deep copy tree (immutability)
2. Find folder node, toggle `isExpanded`
3. If collapsing: compute and cache stats
4. Re-flatten to get new visible items
5. Return new state

### Tree Rendering

Use box-drawing characters like git graph: `├──`, `└──`, `│`

**Example output:**
```
→ src/
  ├── components/
  │   └── M +17 -13  App.tsx
  └── utils/
      └── A +42  -0  helper.ts
  old/ +50 -25 (3 files)
```

**Format:** Selector (`→`) + indentation (`│   ` or `    `) + branch (`├──` or `└──`) + content (folder name or file stats).

Rendering metadata (`isLastChild`, `parentLines`) computed during flattening to keep formatting function pure.

### Navigation

**New key handlers:**
- Enter/Space on folder: Toggle expand/collapse
- Enter on file: Open diff (existing behavior)
- Right arrow on folder: Expand
- Left arrow on folder: Collapse

Type safety via sum types:
```go
switch node := cursorNode.(type) {
case *FolderNode:
    return toggleFolder(s, ctx)
case *FileNode:
    return pushDiffScreen(s, node.file, ctx)
}
```

### Component Structure

**New domain package** (`internal/domain/tree/`):
- `tree.go` - TreeNode interface, FolderNode, FileNode types
- `build.go` - BuildTree function
- `collapse.go` - Path collapsing algorithm
- `flatten.go` - FlattenVisible function

**Modified files state** (`internal/ui/states/files/`):
- `state.go` - Add Root, VisibleItems fields
- `update.go` - Add toggle handlers
- `view.go` - Use TreeSection component

**New UI component** (`internal/ui/components/`):
- `tree_section.go` - TreeSection() renders tree with header
- `tree_line_format.go` - FormatTreeLine() formats individual lines

### Performance

- Build: O(n log n) - sorting dominates
- Flatten: O(visible nodes)
- Navigate: O(1) - index into array
- Toggle: O(n) - deep copy + flatten (infrequent user action)
- Render: O(viewport) - optimal

Imperceptible for typical commits (<100 files). For large commits (1000 files), toggle takes ~few ms (acceptable).

## References

- [Requirements](01_requirements_tree-file-view.md)
- [Tree Approaches Comparison](research/tree-approaches.md) - Detailed analysis of alternatives
- [State Patterns](research/state-patterns.md) - Splice cursor/viewport patterns
- [Rendering Patterns](research/rendering-patterns.md) - ViewBuilder and tree symbols
