# Design: Tree File View

**Status**: Draft (awaiting approval)
**Created**: 2026-01-14

## Executive Summary

This design replaces the flat file list with a hierarchical tree view. We'll build an explicit tree structure from file paths, maintain expand/collapse state in tree nodes, and flatten to a visible list for cursor navigation and rendering. This approach matches Splice's existing git graph pattern (build structure once, store in state, render by traversal) and provides predictable O(1) navigation with O(viewport) rendering performance.

The core decision is **Approach 1: Tree Structure with Flattening** over virtual tree computation (too slow - O(n) per render) or caching (too complex - cache invalidation bugs). We'll represent collapsed paths like `src/components/nested/` as single nodes, use box-drawing characters for tree visualization, and reuse existing cursor/viewport patterns from the files state.

Implementation impacts: new tree domain types, modified files state structure, new tree formatting component, extended update handlers for expand/collapse keys. Testing follows existing patterns with golden files for rendering.

## Context & Problem Statement

The current file view (`internal/ui/states/files/`) displays files as a flat list with full paths:

```
→ M +17 -13  src/components/App.tsx
  A  +42  -0  src/utils/helper.ts
  D   +0 -25  old/deprecated.ts
```

This creates three problems for commits touching many files in deep hierarchies:
1. **Hard to scan** - Long paths and many files make finding specific files difficult
2. **No structure visibility** - Can't see which files are in which folders
3. **High cognitive load** - Repeated path prefixes reduce readability

**This design covers**: Tree data structure, expand/collapse state management, cursor navigation through hierarchy, tree rendering with box-drawing characters.

**This design does not cover**: Search/filtering, remembering state across commits, bulk operations, visual folder change indicators.

## Current State

### Data Structure

Files are stored as a flat array:

```go
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange  // Flat list
    Cursor        int                // Index into Files
    ViewportStart int
}

type FileChange struct {
    Path      string  // Full path: "src/components/App.tsx"
    Status    string  // M, A, D, R
    Additions int
    Deletions int
    IsBinary  bool
}
```

### Rendering

`FileSection` component formats each file with full path:
- Selector indicator (`→` or space)
- Status letter (colored: A=green, M=yellow, D=red)
- Stats (`+17 -13`, right-aligned)
- Full file path

Viewport scrolling shows a slice of the flat list with fixed header accounting.

### Navigation

- Up/Down: Move cursor through flat list
- Enter: Open diff for selected file
- Viewport adjusts to keep cursor visible

## Solution

### Overview

We'll build a tree structure from the flat file list, maintain expand/collapse state in tree nodes, and flatten to a visible list for navigation and rendering. This matches Splice's git graph pattern: compute layout once, store in state, render by traversal.

```
BuildTree(files) → TreeNode (stored in state)
                     ↓
              FlattenVisible(tree) → []*TreeNode (visible items)
                     ↓
              Render(visible[viewport])
```

### Architecture Decision

> **Decision:** Use Approach 1 (Tree Structure with Flattening) over Approach 2 (Virtual Tree) and Approach 3 (Tree + Cache).
>
> **Rationale:**
> - **Matches existing patterns**: Similar to git graph layout (build once, store, render)
> - **Predictable performance**: O(1) navigation, O(viewport) rendering, O(n) toggle (acceptable - user action only)
> - **Avoids Approach 2's flaw**: No O(n) recomputation every render (wasteful for large commits)
> - **Avoids Approach 3's complexity**: Cache invalidation is bug-prone with negligible benefit
> - **Testing simplicity**: No cache edge cases, uses existing golden file pattern
>
> **Tradeoff**: Uses more memory (tree + flattened list) and requires deep copy on toggle. This is acceptable because:
> - Typical commits have <100 files (tree is small)
> - Toggle happens infrequently (user action only)
> - Memory cost is negligible compared to git data already in memory

### Data Model

#### Tree Node

```go
// Tree node types
type TreeNodeType int

const (
    NodeTypeFolder TreeNodeType = iota
    NodeTypeFile
)

// TreeNode represents a folder or file in the hierarchy
type TreeNode struct {
    Name     string           // Display name (may include path for collapsed folders)
    Type     TreeNodeType
    Depth    int              // Nesting level (0 = top level)
    Children []*TreeNode      // Sorted: folders first, then files (alphabetically)
    File     *core.FileChange // Non-nil for files only

    // Folder-specific state
    IsExpanded bool         // Only relevant for folders
    Stats      FolderStats  // Aggregate stats (used when collapsed)
}

// FolderStats holds aggregate statistics for a folder
type FolderStats struct {
    FileCount  int
    Additions  int
    Deletions  int
}
```

**Key design choices:**

1. **Collapsed paths in Name field**: A node representing `src/components/nested/` has `Name = "src/components/nested"`. This simplifies rendering - no special case logic.

2. **Children sorted at build time**: Folders first, then files, each group alphabetically. This matches requirements and simplifies rendering (no runtime sorting).

3. **Stats computed on collapse**: When a folder is collapsed, we traverse children to compute aggregate stats. Cached in the node.

#### Files State

```go
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange  // Original flat list (unchanged from git)
    RootNode      *TreeNode          // Tree root (virtual node at depth -1)
    VisibleNodes  []*TreeNode        // Flattened visible items (for cursor navigation)
    Cursor        int                // Index into VisibleNodes
    ViewportStart int
}
```

**Changes from current state:**
- Added `RootNode` - tree structure
- Added `VisibleNodes` - flattened view for navigation
- `Cursor` now indexes into `VisibleNodes` instead of `Files`

### Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Initialization (files.New)                              │
│    []FileChange → BuildTree() → *TreeNode                  │
│                                    ↓                        │
│                          FlattenVisible() → []*TreeNode    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. User toggles folder (Enter/Space/Left/Right on folder)  │
│    Deep copy tree → Toggle node → FlattenVisible() again   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Rendering (View)                                         │
│    VisibleNodes[ViewportStart:ViewportEnd]                 │
│         ↓                                                   │
│    FormatTreeLine(node) → Tree characters + file format    │
└─────────────────────────────────────────────────────────────┘
```

### Core Algorithms

#### 1. Build Tree

```
BuildTree(files []FileChange) *TreeNode:
    1. Create root node (virtual, depth -1)
    2. For each file path:
        - Split into parts
        - Navigate/create folder nodes for all parts except last
        - Create file node for last part
    3. Sort all children (folders first, alphabetical)
    4. Collapse single-child folder paths
    5. Compute stats for all folders (for future collapse operations)
    6. Return root
```

**Path collapsing logic:**
```
CollapsePaths(node):
    1. Recursively process children first
    2. If node has exactly 1 child AND that child is a folder:
        - Merge: node.Name += "/" + child.Name
        - Replace node.Children with grandchildren
        - Recursively try to collapse again
    3. Otherwise: done
```

Example: `src/` → `components/` → `nested/` → `App.tsx` becomes:
- Node: `Name="src/components/nested"`, `Children=[App.tsx node]`

#### 2. Flatten Visible

```
FlattenVisible(root) []*TreeNode:
    result = []
    Walk(node, depth):
        if depth >= 0:  // Skip virtual root
            result.append(node)
        if node.Type == Folder AND node.IsExpanded:
            for child in node.Children:
                Walk(child, depth + 1)

    Walk(root, -1)
    return result
```

This produces the list of visible items in display order for cursor navigation.

#### 3. Toggle Folder

```
Toggle(state, cursor) State:
    1. Get node at cursor from VisibleNodes
    2. If not a folder: return unchanged state
    3. Deep copy entire tree
    4. Find corresponding node in new tree
    5. Toggle IsExpanded
    6. If collapsing: compute and cache Stats
    7. Flatten new tree to get new VisibleNodes
    8. Adjust cursor if needed (keep on same folder)
    9. Adjust viewport to keep cursor visible
    10. Return new state
```

**Immutability**: We deep copy the entire tree on toggle to maintain immutability. This is acceptable because:
- Toggles are infrequent (user action only)
- Trees are small (typical commit has <100 files)
- Deep copy is simple and predictable

### Tree Rendering

#### Box-Drawing Characters

We'll use Unicode box-drawing characters (same as git graph):

```
├── Item      (branch to item)
└── Last      (last child in folder)
│            (vertical continuation line)
```

**Rendering pattern for depth:**
```
Depth 0:  "→ src/"
Depth 1:  "  ├── components/"
Depth 2:  "  │   └── M +17 -13  App.tsx"
```

Each depth level adds indentation. For non-last items use `├──`, for last items use `└──`. Parent lines use `│` for continuation.

#### Tree Line Format

```
FormatTreeLine(node, isSelected, width) string:
    line = ""

    // 1. Selector prefix (cursor indicator)
    if isSelected:
        line += "→ "
    else:
        line += "  "

    // 2. Tree structure indentation
    for i in 0..<node.Depth:
        if i needs continuation line:
            line += "│   "
        else:
            line += "    "

    // 3. Branch character
    if node.Depth > 0:
        if node is last child:
            line += "└── "
        else:
            line += "├── "

    // 4. Content
    if node.Type == Folder:
        if node.IsExpanded:
            line += style(node.Name + "/")
        else:
            line += style(node.Name + "/ ") + stats(node.Stats)
    else:
        line += style(node.File.Status) + " " +
                stats(node.File) + "  " +
                style(node.Name)

    return line
```

**Key insight**: Determining "is last child" and "needs continuation line" requires parent context. We'll need to either:
- Pass parent context to formatting function, OR
- Store this metadata in TreeNode during flattening

> **Decision:** Store rendering metadata during flattening. Add `IsLastChild` bool to TreeNode (or create separate `VisibleTreeItem` struct). This keeps formatting function pure and simple.

Revised VisibleNodes type:

```go
type VisibleTreeItem struct {
    Node          *TreeNode
    IsLastChild   bool
    ParentLines   []bool  // Which depth levels need │ continuation
}
```

### Navigation Updates

The files state's `Update` method needs new key handlers:

**Existing keys (no change):**
- Up/Down: Move cursor through VisibleNodes
- Enter on file: Push diff screen
- Backspace: Pop to log screen

**New key behavior:**
- **Enter on folder**: Toggle expand/collapse
- **Space on folder**: Toggle expand/collapse
- **Right arrow on folder**: Expand (if collapsed)
- **Left arrow on folder**: Collapse (if expanded)

**Implementation:**
```go
case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Space):
    if cursorNode.Type == NodeTypeFolder:
        return toggleFolder(s, ctx)
    else:
        return pushDiffScreen(s, ctx)

case key.Matches(msg, keys.Right):
    if cursorNode.Type == NodeTypeFolder && !cursorNode.IsExpanded:
        return toggleFolder(s, ctx)
    return s, nil

case key.Matches(msg, keys.Left):
    if cursorNode.Type == NodeTypeFolder && cursorNode.IsExpanded:
        return toggleFolder(s, ctx)
    return s, nil
```

### Component Structure

Following Splice's patterns, we'll organize as:

**New domain logic** (`internal/domain/tree/`):
- `tree.go` - TreeNode type, BuildTree function
- `collapse.go` - Path collapsing algorithm
- `flatten.go` - FlattenVisible function
- `stats.go` - ComputeStats for folders

**Modified files state** (`internal/ui/states/files/`):
- `state.go` - Add RootNode, VisibleNodes fields
- `update.go` - Add toggle handlers
- `view.go` - Use new TreeSection component

**New UI component** (`internal/ui/components/`):
- `tree_section.go` - TreeSection(items, cursor) renders tree view
- `tree_line_format.go` - FormatTreeLine(item, selected, width)

**Rationale for domain/tree package:**
- Tree building is pure domain logic (no UI concerns)
- Reusable across other features (if needed)
- Matches graph layout pattern (domain/graph/)
- Easy to unit test in isolation

### Visual Design

**Expanded tree:**
```
→ src/
  ├── components/
  │   └── M +17 -13  App.tsx
  ├── utils/
  │   └── A +42  -0  helper.ts
  └── M +5 -2  index.ts
  old/
  └── D +0 -15  deprecated.ts
```

**With collapsed path and collapsed folder:**
```
→ src/components/nested/deep/
  └── M +17 -13  App.tsx
  old/ +50 -25 (3 files)
```

**Styling:**
- Folders: Same color as current file paths
- Selected folder: Bold + brighter (match selected file)
- Stats in collapsed folders: Same as file stats (green +, red -)
- Tree characters: Dim gray (non-intrusive)
- File lines: Unchanged from current format

### Performance Characteristics

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Initial tree build | O(n log n) | n = file count; sorting dominates |
| Flatten visible | O(n) | Traverse all visible nodes |
| Cursor movement | O(1) | Index into array |
| Toggle folder | O(n) | Deep copy + flatten; only on user action |
| Render viewport | O(v) | v = viewport size; typically 20-50 lines |

For typical commit sizes (<100 files), all operations are imperceptible. Even for large commits (1000 files), only toggle incurs noticeable cost (~few ms), which is acceptable for user-initiated action.

## Open Questions

None - design is complete and ready for implementation.

## References

- [Requirements](01_requirements_tree-file-view.md) - Feature requirements
- [Current Implementation](research/current-implementation.md) - Flat list architecture
- [State Patterns](research/state-patterns.md) - Cursor, viewport, immutability patterns
- [Rendering Patterns](research/rendering-patterns.md) - ViewBuilder, styling, tree symbols
- [Tree Approaches](research/tree-approaches.md) - Detailed comparison of 3 approaches
