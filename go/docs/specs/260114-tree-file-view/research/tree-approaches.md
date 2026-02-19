# Tree File View Approaches Research

**Created**: 2026-01-14

This document analyzes three different approaches for representing and navigating a hierarchical file tree structure in the Splice file view. The analysis considers Splice's architectural patterns (immutability, pure functions, simplicity) and requirements for expand/collapse, path collapsing, and efficient navigation.

## Table of Contents

1. [Requirements Summary](#requirements-summary)
2. [Approach 1: Tree Structure with Flattening](#approach-1-tree-structure-with-flattening)
3. [Approach 2: Virtual Tree (Path-Based)](#approach-2-virtual-tree-path-based)
4. [Approach 3: Hybrid (Tree + Cache)](#approach-3-hybrid-tree--cache)
5. [Comparison Matrix](#comparison-matrix)
6. [Recommendation](#recommendation)

## Requirements Summary

From `/workspace/docs/specs/260114-tree-file-view/01_requirements_tree-file-view.md`:

1. **Input**: Flat `[]core.FileChange` list from git
2. **Tree Structure**: Hierarchical folders with files
3. **Expand/Collapse**: Toggle folder visibility with state tracking
4. **Path Collapsing**: Single-child folders collapsed into one line (e.g., `src/components/nested/`)
5. **Navigation**: Cursor moves through visible items (up/down)
6. **Rendering**: Tree with box-drawing characters
7. **Default State**: All folders expanded by default

**Key Operations**:
- Build tree from flat file list
- Toggle expand/collapse on folder
- Navigate cursor up/down through visible items
- Render visible items with tree structure

## Approach 1: Tree Structure with Flattening

### Overview

Build an explicit tree structure from file paths, maintain expand/collapse state in tree nodes, and flatten to a visible list for rendering and navigation.

### Data Structures

```go
// Tree node types
type TreeNodeType int

const (
    NodeTypeFolder TreeNodeType = iota
    NodeTypeFile
)

// Tree node - represents a folder or file
type TreeNode struct {
    Name     string              // Just the name (not full path)
    Type     TreeNodeType        // Folder or file
    Depth    int                 // Nesting level (0 = root)
    Children []*TreeNode         // Child nodes (folders first, then files)
    File     *core.FileChange    // Non-nil for files only

    // Folder-specific state
    IsExpanded bool               // Only relevant for folders
    Stats      FolderStats        // Aggregate stats for collapsed folders
}

// Aggregate stats for folders
type FolderStats struct {
    FileCount  int
    Additions  int
    Deletions  int
}

// State structure
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange  // Original flat list
    RootNode      *TreeNode          // Tree root
    VisibleNodes  []*TreeNode        // Flattened visible nodes (cached)
    Cursor        int                // Index into VisibleNodes
    ViewportStart int
}
```

### Algorithm: Build Tree

```go
func BuildTree(files []core.FileChange) *TreeNode {
    root := &TreeNode{
        Name:       "",
        Type:       NodeTypeFolder,
        IsExpanded: true,
        Children:   []*TreeNode{},
        Depth:      -1, // Root is -1, top-level items are 0
    }

    // Build tree by inserting each file path
    for _, file := range files {
        insertFile(root, file)
    }

    // Sort children (folders first, then alphabetically)
    sortTree(root)

    // Collapse single-child paths
    collapsePaths(root)

    return root
}

func insertFile(parent *TreeNode, file core.FileChange) {
    parts := strings.Split(file.Path, "/")
    current := parent

    // Navigate/create folders for all path parts except the last
    for i := 0; i < len(parts)-1; i++ {
        folderName := parts[i]
        child := findChild(current, folderName, NodeTypeFolder)
        if child == nil {
            child = &TreeNode{
                Name:       folderName,
                Type:       NodeTypeFolder,
                IsExpanded: true,  // Default expanded
                Children:   []*TreeNode{},
                Depth:      current.Depth + 1,
            }
            current.Children = append(current.Children, child)
        }
        current = child
    }

    // Add file node
    fileName := parts[len(parts)-1]
    fileNode := &TreeNode{
        Name:  fileName,
        Type:  NodeTypeFile,
        File:  &file,
        Depth: current.Depth + 1,
    }
    current.Children = append(current.Children, fileNode)
}

func collapsePaths(node *TreeNode) {
    if node.Type != NodeTypeFolder {
        return
    }

    // Recursively collapse children first
    for _, child := range node.Children {
        collapsePaths(child)
    }

    // Collapse this node if it has exactly one child that is a folder
    if len(node.Children) == 1 && node.Children[0].Type == NodeTypeFolder {
        child := node.Children[0]

        // Merge names: "src" + "components" = "src/components"
        node.Name = node.Name + "/" + child.Name
        node.Children = child.Children

        // Update depth of all descendants
        updateDepths(node, node.Depth)

        // Continue collapsing (recursive collapse)
        collapsePaths(node)
    }
}

func sortTree(node *TreeNode) {
    // Sort children: folders first, then alphabetically
    sort.Slice(node.Children, func(i, j int) bool {
        a, b := node.Children[i], node.Children[j]

        // Folders before files
        if a.Type != b.Type {
            return a.Type == NodeTypeFolder
        }

        // Alphabetically within type
        return a.Name < b.Name
    })

    // Recursively sort children
    for _, child := range node.Children {
        sortTree(child)
    }
}
```

### Algorithm: Flatten to Visible Nodes

```go
func FlattenVisible(root *TreeNode) []*TreeNode {
    var result []*TreeNode
    flattenVisibleRecursive(root, &result)
    return result
}

func flattenVisibleRecursive(node *TreeNode, result *[]*TreeNode) {
    // Don't add root itself
    if node.Depth >= 0 {
        *result = append(*result, node)
    }

    // If folder is expanded, add children
    if node.Type == NodeTypeFolder && node.IsExpanded {
        for _, child := range node.Children {
            flattenVisibleRecursive(child, result)
        }
    }
}
```

### Algorithm: Toggle Expand/Collapse

```go
func (s *State) toggleFolder() *State {
    if s.Cursor < 0 || s.Cursor >= len(s.VisibleNodes) {
        return s
    }

    node := s.VisibleNodes[s.Cursor]
    if node.Type != NodeTypeFolder {
        return s  // Can't toggle files
    }

    // IMPORTANT: We need to modify the tree, so we must deep copy it first
    // to maintain immutability at the state level
    newRoot := deepCopyTree(s.RootNode)

    // Find the corresponding node in the new tree
    targetNode := findNodeByPath(newRoot, getNodePath(node))
    if targetNode == nil {
        return s
    }

    // Toggle expansion
    targetNode.IsExpanded = !targetNode.IsExpanded

    // If collapsing, compute aggregate stats
    if !targetNode.IsExpanded {
        targetNode.Stats = computeStats(targetNode)
    }

    // Rebuild visible list
    newVisibleNodes := FlattenVisible(newRoot)

    // Adjust cursor if needed (keep on same node if possible)
    newCursor := findNodeIndex(newVisibleNodes, getNodePath(node))
    if newCursor == -1 {
        newCursor = min(s.Cursor, len(newVisibleNodes)-1)
    }

    return &State{
        Source:        s.Source,
        Files:         s.Files,
        RootNode:      newRoot,
        VisibleNodes:  newVisibleNodes,
        Cursor:        newCursor,
        ViewportStart: s.ViewportStart,
    }
}
```

### Algorithm: Navigation

```go
func (s *State) moveCursor(direction int) *State {
    newCursor := s.Cursor + direction

    // Bounds checking
    if newCursor < 0 {
        newCursor = 0
    }
    if newCursor >= len(s.VisibleNodes) {
        newCursor = len(s.VisibleNodes) - 1
    }

    // No need to rebuild tree or visible list - just update cursor
    return &State{
        Source:        s.Source,
        Files:         s.Files,
        RootNode:      s.RootNode,
        VisibleNodes:  s.VisibleNodes,
        Cursor:        newCursor,
        ViewportStart: s.ViewportStart,
    }
}
```

### Algorithm: Rendering

```go
func (s *State) renderTree(ctx core.Context) {
    vb := components.NewViewBuilder()

    // Calculate viewport bounds
    availableHeight := calculateAvailableHeight(ctx.Height())
    viewportEnd := min(s.ViewportStart + availableHeight, len(s.VisibleNodes))

    // Render only visible nodes
    for i := s.ViewportStart; i < viewportEnd; i++ {
        node := s.VisibleNodes[i]
        isSelected := (i == s.Cursor)

        line := formatTreeLine(node, isSelected, ctx.Width())
        vb.AddLine(line)
    }

    return vb
}

func formatTreeLine(node *TreeNode, isSelected bool, width int) string {
    var line strings.Builder

    // 1. Selection indicator
    if isSelected {
        line.WriteString("→ ")
    } else {
        line.WriteString("  ")
    }

    // 2. Indentation (2 spaces per depth level)
    indent := strings.Repeat("  ", node.Depth)
    line.WriteString(indent)

    // 3. Tree structure symbols
    // (Simplified - would need context of siblings to show ├─ vs └─)
    if node.Type == NodeTypeFolder {
        line.WriteString("├─ ")
    } else {
        line.WriteString("└─ ")
    }

    // 4. Content
    if node.Type == NodeTypeFolder {
        line.WriteString(node.Name)
        line.WriteString("/")

        if !node.IsExpanded {
            // Show aggregate stats for collapsed folder
            line.WriteString(fmt.Sprintf(" +%d -%d (%d files)",
                node.Stats.Additions,
                node.Stats.Deletions,
                node.Stats.FileCount))
        }
    } else {
        // File with stats
        file := node.File
        line.WriteString(fmt.Sprintf("%s +%d -%d  %s",
            file.Status,
            file.Additions,
            file.Deletions,
            node.Name))
    }

    return line.String()
}
```

### Analysis

**State Complexity**: Medium
- Explicit tree structure is straightforward to understand
- Deep copy needed for immutability adds complexity
- Cached visible list adds one more data structure to manage

**Navigation Efficiency**:
- Up/Down: O(1) - just index into VisibleNodes array
- Toggle: O(n) - need to deep copy tree and rebuild visible list
- Initial build: O(n log n) due to sorting

**Rendering Performance**:
- Excellent - only iterate visible nodes in viewport
- O(viewport_size) per render

**Path Collapsing**:
- Natural - collapsed paths are represented as single nodes with "/" in name
- Easy to implement during tree construction
- Expand behavior: when expanding "src/components/nested/", just expand that one node (already represents the full path)

**Testing Complexity**:
- Medium - need to test tree building, flattening, and synchronization
- Golden files work well for rendering tests
- Need deterministic tree construction tests

**Pros**:
- **Clear separation of concerns**: Tree building, flattening, and rendering are separate steps
- **Efficient navigation**: Cursor moves through flat array
- **Natural path collapsing**: Collapsed paths are just nodes with "/" in name
- **Viewport rendering**: Only render visible slice of flattened list
- **Follows Splice patterns**: Similar to graph layout (compute once, render many)

**Cons**:
- **Deep copy overhead**: Every toggle requires full tree copy for immutability
- **Two representations**: Tree structure + flattened list must stay in sync
- **Memory overhead**: Stores both tree and flattened list
- **Complexity**: More data structures to manage

## Approach 2: Virtual Tree (Path-Based)

### Overview

Keep the flat file list and compute tree structure on-demand during rendering. Track expanded folders as a set of path strings. Cursor navigates through computed visible items.

### Data Structures

```go
// Visible item types (computed on-demand)
type VisibleItemType int

const (
    ItemTypeFolder VisibleItemType = iota
    ItemTypeFile
)

// Computed visible item
type VisibleItem struct {
    Type      VisibleItemType
    Path      string           // Full path for folders, full path for files
    Name      string           // Display name (may be collapsed path like "src/components/")
    Depth     int              // Nesting depth
    FileIndex int              // Index into Files array (for files only, -1 for folders)
}

// State structure
type State struct {
    Source          core.DiffSource
    Files           []core.FileChange  // Original flat list (source of truth)
    ExpandedFolders map[string]bool    // Set of expanded folder paths
    Cursor          int                // Index into computed visible items
    ViewportStart   int
}
```

### Algorithm: Compute Visible Items

```go
func ComputeVisibleItems(files []core.FileChange, expandedFolders map[string]bool) []VisibleItem {
    // 1. Extract unique folder paths
    folderPaths := extractFolderPaths(files)

    // 2. Sort paths (lexicographic order gives us parent-before-child)
    sort.Strings(folderPaths)

    // 3. Collapse single-child paths
    collapsedFolders := collapsePaths(folderPaths)

    // 4. Build visible items list
    var items []VisibleItem

    for _, folder := range collapsedFolders {
        depth := strings.Count(folder.DisplayPath, "/")

        // Add folder item
        items = append(items, VisibleItem{
            Type:      ItemTypeFolder,
            Path:      folder.FullPath,
            Name:      folder.DisplayName,
            Depth:     depth,
            FileIndex: -1,
        })

        // If folder is expanded, add child files/folders
        if expandedFolders[folder.FullPath] {
            // Add files that belong to this folder
            for i, file := range files {
                if strings.HasPrefix(file.Path, folder.FullPath+"/") {
                    // Check if file is direct child (not in subfolder)
                    relativePath := strings.TrimPrefix(file.Path, folder.FullPath+"/")
                    if !strings.Contains(relativePath, "/") {
                        items = append(items, VisibleItem{
                            Type:      ItemTypeFile,
                            Path:      file.Path,
                            Name:      relativePath,
                            Depth:     depth + 1,
                            FileIndex: i,
                        })
                    }
                }
            }
        }
    }

    return items
}

func extractFolderPaths(files []core.FileChange) []string {
    folderSet := make(map[string]bool)

    for _, file := range files {
        parts := strings.Split(file.Path, "/")

        // Extract all folder paths
        currentPath := ""
        for i := 0; i < len(parts)-1; i++ {
            if currentPath != "" {
                currentPath += "/"
            }
            currentPath += parts[i]
            folderSet[currentPath] = true
        }
    }

    // Convert set to sorted slice
    folders := make([]string, 0, len(folderSet))
    for folder := range folderSet {
        folders = append(folders, folder)
    }

    return folders
}

type CollapsedFolder struct {
    FullPath    string  // Actual path in filesystem
    DisplayName string  // What to show (may be collapsed like "src/components/")
    DisplayPath string  // For depth calculation
}

func collapsePaths(sortedFolders []string) []CollapsedFolder {
    var result []CollapsedFolder

    for _, folder := range sortedFolders {
        // Check if this is a single-child folder chain
        collapsedPath := folder

        // Keep extending path while folder has exactly one child folder
        for {
            hasOnlyOneFolderChild := false
            var childFolder string

            for _, candidate := range sortedFolders {
                // Check if candidate is direct child
                if strings.HasPrefix(candidate, folder+"/") {
                    relativePath := strings.TrimPrefix(candidate, folder+"/")
                    if !strings.Contains(relativePath, "/") {
                        // Direct child folder
                        if !hasOnlyOneFolderChild {
                            hasOnlyOneFolderChild = true
                            childFolder = candidate
                        } else {
                            // Multiple children, stop
                            hasOnlyOneFolderChild = false
                            break
                        }
                    }
                }
            }

            if hasOnlyOneFolderChild {
                collapsedPath = childFolder
            } else {
                break
            }
        }

        // Extract display name
        displayName := filepath.Base(collapsedPath)
        if collapsedPath != folder {
            // Collapsed path - show full collapsed portion
            displayName = strings.TrimPrefix(collapsedPath, folder+"/")
        }

        result = append(result, CollapsedFolder{
            FullPath:    collapsedPath,
            DisplayName: displayName,
            DisplayPath: collapsedPath,
        })
    }

    return result
}
```

### Algorithm: Toggle Expand/Collapse

```go
func (s *State) toggleFolder(visibleItems []VisibleItem) *State {
    if s.Cursor < 0 || s.Cursor >= len(visibleItems) {
        return s
    }

    item := visibleItems[s.Cursor]
    if item.Type != ItemTypeFolder {
        return s  // Can't toggle files
    }

    // Create new expanded folders set
    newExpanded := make(map[string]bool, len(s.ExpandedFolders))
    for k, v := range s.ExpandedFolders {
        newExpanded[k] = v
    }

    // Toggle folder
    if newExpanded[item.Path] {
        delete(newExpanded, item.Path)
    } else {
        newExpanded[item.Path] = true
    }

    return &State{
        Source:          s.Source,
        Files:           s.Files,
        ExpandedFolders: newExpanded,
        Cursor:          s.Cursor,
        ViewportStart:   s.ViewportStart,
    }
}
```

### Algorithm: Navigation

```go
func (s *State) moveCursor(direction int, visibleItems []VisibleItem) *State {
    newCursor := s.Cursor + direction

    // Bounds checking
    if newCursor < 0 {
        newCursor = 0
    }
    if newCursor >= len(visibleItems) {
        newCursor = len(visibleItems) - 1
    }

    return &State{
        Source:          s.Source,
        Files:           s.Files,
        ExpandedFolders: s.ExpandedFolders,
        Cursor:          newCursor,
        ViewportStart:   s.ViewportStart,
    }
}
```

### Algorithm: Rendering

```go
func (s *State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // Compute visible items (computed on every render)
    visibleItems := ComputeVisibleItems(s.Files, s.ExpandedFolders)

    // Calculate viewport bounds
    availableHeight := calculateAvailableHeight(ctx.Height())
    viewportEnd := min(s.ViewportStart + availableHeight, len(visibleItems))

    // Render only visible nodes in viewport
    for i := s.ViewportStart; i < viewportEnd; i++ {
        item := visibleItems[i]
        isSelected := (i == s.Cursor)

        line := formatTreeLine(item, s.Files, isSelected, ctx.Width())
        vb.AddLine(line)
    }

    return vb
}
```

### Analysis

**State Complexity**: Low
- Very simple state: just files + expanded set + cursor position
- No tree structure to manage
- No caching or synchronization

**Navigation Efficiency**:
- Up/Down: O(1) - just update cursor
- Toggle: O(1) - just update expanded set
- BUT: Each operation requires recomputing visible items for rendering
- Initial build: O(n log n) due to sorting

**Rendering Performance**:
- Poor - must recompute visible items on every render
- O(n) to compute visible items (where n = number of files)
- Then O(viewport_size) to render
- Total: O(n) per render

**Path Collapsing**:
- Complex - must detect collapsed paths during visible item computation
- Need to track both full path and display name
- Expand behavior is tricky: expanding "src/components/nested/" means adding intermediate folders

**Testing Complexity**:
- Low - state is simple
- Computation is stateless (pure function)
- Golden files work well

**Pros**:
- **Simple state**: Only store expanded set, no tree structure
- **True immutability**: All computation is pure, no deep copies needed
- **No synchronization**: Can't have inconsistent state
- **Easy to test**: Pure functions with simple inputs/outputs

**Cons**:
- **Performance**: Recompute tree structure on every render
- **Inefficient**: O(n) work per render even for small viewports
- **Complex path collapsing**: Must recompute collapsed paths each time
- **Navigation requires computation**: Can't navigate without computing visible items first

## Approach 3: Hybrid (Tree + Cache)

### Overview

Build a tree structure but cache the flattened visible list. Invalidate cache only when tree structure changes (expand/collapse). Cursor navigates cached list.

### Data Structures

```go
// Same tree structure as Approach 1
type TreeNode struct {
    Name       string
    Type       TreeNodeType
    Depth      int
    Children   []*TreeNode
    File       *core.FileChange
    IsExpanded bool
    Stats      FolderStats
}

type State struct {
    Source        core.DiffSource
    Files         []core.FileChange
    RootNode      *TreeNode

    // Cached flattened list
    VisibleNodes  []*TreeNode
    CacheValid    bool  // Flag to know when to recompute

    Cursor        int
    ViewportStart int
}
```

### Algorithm: Toggle with Cache Invalidation

```go
func (s *State) toggleFolder() *State {
    if s.Cursor < 0 || s.Cursor >= len(s.VisibleNodes) {
        return s
    }

    node := s.VisibleNodes[s.Cursor]
    if node.Type != NodeTypeFolder {
        return s
    }

    // Deep copy tree for immutability
    newRoot := deepCopyTree(s.RootNode)

    // Find and toggle node
    targetNode := findNodeByPath(newRoot, getNodePath(node))
    if targetNode == nil {
        return s
    }

    targetNode.IsExpanded = !targetNode.IsExpanded
    if !targetNode.IsExpanded {
        targetNode.Stats = computeStats(targetNode)
    }

    // Invalidate cache (will be recomputed on next View)
    return &State{
        Source:        s.Source,
        Files:         s.Files,
        RootNode:      newRoot,
        VisibleNodes:  nil,  // Invalidate
        CacheValid:    false,
        Cursor:        s.Cursor,
        ViewportStart: s.ViewportStart,
    }
}
```

### Algorithm: Lazy Cache Rebuild

```go
func (s *State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // Rebuild cache if invalid
    visibleNodes := s.VisibleNodes
    if !s.CacheValid {
        visibleNodes = FlattenVisible(s.RootNode)
        // Note: Can't actually update s here (View is read-only)
        // Would need to rebuild in Update instead
    }

    // Render viewport
    availableHeight := calculateAvailableHeight(ctx.Height())
    viewportEnd := min(s.ViewportStart + availableHeight, len(visibleNodes))

    for i := s.ViewportStart; i < viewportEnd; i++ {
        node := visibleNodes[i]
        isSelected := (i == s.Cursor)
        line := formatTreeLine(node, isSelected, ctx.Width())
        vb.AddLine(line)
    }

    return vb
}
```

### Analysis

**State Complexity**: Medium-High
- Same tree structure as Approach 1
- Additional cache validity flag
- Need to manage cache invalidation logic

**Navigation Efficiency**:
- Up/Down: O(1) - index into cached list
- Toggle: O(n) - deep copy tree + flatten (but happens once, cache reused after)
- Initial build: O(n log n)

**Rendering Performance**:
- Excellent when cache valid: O(viewport_size)
- Rebuilds cache on first render after toggle: O(n)
- Amortized: very good (most renders use cache)

**Path Collapsing**:
- Same as Approach 1 - natural representation

**Testing Complexity**:
- Medium-High - need to test cache invalidation logic
- More edge cases than Approach 1
- Risk of bugs if cache isn't invalidated correctly

**Pros**:
- **Better performance than Approach 2**: Cache reused across renders
- **Better than Approach 1**: No need to rebuild cache on navigation
- **Clear performance profile**: O(1) for navigation, O(n) only on toggle

**Cons**:
- **Cache invalidation complexity**: Classic hard problem
- **Lazy rebuild complication**: View method can't update state, need to rebuild in Update
- **More state to manage**: Cache + validity flag
- **Debugging difficulty**: Hard to track when cache is invalid
- **Not idiomatic for Splice**: Adds mutability concerns (cache staleness)

## Comparison Matrix

| Criterion | Approach 1: Tree + Flatten | Approach 2: Virtual Tree | Approach 3: Tree + Cache |
|-----------|---------------------------|-------------------------|-------------------------|
| **State Complexity** | Medium (tree + flat list) | Low (just expanded set) | Medium-High (tree + cache flag) |
| **Navigation (up/down)** | O(1) | O(1) | O(1) |
| **Toggle expand/collapse** | O(n) - copy tree + flatten | O(1) - update set | O(n) - copy tree + invalidate |
| **Rendering** | O(viewport) | O(n) - compute + viewport | O(viewport) if cached, O(n) if invalid |
| **Memory Usage** | High (tree + flat list) | Low (no tree) | High (tree + cache) |
| **Path Collapsing** | Natural | Complex | Natural |
| **Testing Complexity** | Medium | Low | Medium-High |
| **Immutability** | Good (deep copy on change) | Excellent (fully immutable) | Tricky (cache invalidation) |
| **Code Clarity** | Good | Good | Poor (cache management) |
| **Splice Patterns** | Matches graph pattern | Pure functional | Adds statefulness |
| **Performance Profile** | Consistent O(n) on toggle | Worst case every render | Variable (amortized good) |

## Recommendation

**Recommended Approach: Approach 1 (Tree Structure with Flattening)**

### Rationale

1. **Matches Splice's Architecture**
   - Similar to git graph layout pattern (`internal/domain/graph/`):
     - Build layout structure once (`ComputeLayout` → `BuildTree`)
     - Store in state (`*graph.Layout` → `*TreeNode`)
     - Render by traversing structure (`RenderRow` → `formatTreeLine`)
   - Immutability through deep copy (acceptable tradeoff for clarity)
   - Clear separation: build tree (domain) → flatten (transform) → render (UI)

2. **Predictable Performance**
   - Navigation is O(1) - just cursor movement through array
   - Toggle is O(n) but predictable and only happens on user action
   - Rendering is O(viewport_size) - optimal
   - No hidden recomputation on every render (unlike Approach 2)

3. **Path Collapsing is Natural**
   - Collapsed paths are just nodes with "/" in name
   - Easy to understand: `node.Name = "src/components/nested"`
   - Expansion is simple: node already represents full collapsed path
   - Matches user mental model

4. **Testing is Straightforward**
   - Tree building: unit test with deterministic input
   - Flattening: unit test visible node list
   - Rendering: golden files (already used throughout Splice)
   - No cache invalidation edge cases to worry about

5. **Code Clarity**
   - Each operation is explicit:
     - Toggle → copy tree → rebuild visible list
     - Navigate → update cursor
     - Render → iterate visible slice
   - No hidden state (like cache validity flags)
   - Easy to debug: can inspect tree and visible list independently

6. **Avoids Approach 2's Performance Issue**
   - Approach 2 recomputes tree structure on every render
   - For large commits (hundreds of files), this is wasteful
   - Splice already shows this pattern is wrong: graph layout is computed once, not per render

7. **Avoids Approach 3's Complexity**
   - Cache invalidation is notoriously bug-prone
   - Adds conceptual complexity without clear benefit over Approach 1
   - The performance difference is negligible:
     - Approach 1: Always rebuild visible list on toggle (predictable)
     - Approach 3: Rebuild on first render after toggle (unpredictable)
   - Not worth the added complexity

### Implementation Notes

When implementing Approach 1:

1. **Deep Copy Optimization**
   - Only copy the tree on toggle (not on navigation)
   - Could use path-copying (copy only path from root to changed node)
   - For typical tree sizes (<1000 files), full deep copy is fine

2. **Immutability Pattern**
   ```go
   // Toggle returns new state with new tree
   func (s *State) toggleFolder() *State {
       newRoot := deepCopyTree(s.RootNode)
       // ... modify newRoot ...
       newVisible := FlattenVisible(newRoot)
       return &State{
           RootNode:     newRoot,
           VisibleNodes: newVisible,
           // ... other fields ...
       }
   }
   ```

3. **Tree Building**
   - Build tree in constructor (`New()`)
   - Store in state
   - Rebuild visible list only on expand/collapse

4. **Rendering**
   - Similar to Files state's current pattern
   - Iterate `s.VisibleNodes[s.ViewportStart:viewportEnd]`
   - Pass node to formatting function

5. **Testing Strategy**
   - Unit tests for tree building (various path structures)
   - Unit tests for path collapsing
   - Unit tests for flattening (expanded vs collapsed)
   - Unit tests for toggle (before/after visible lists)
   - Golden file tests for rendering

### Alternative if Performance Becomes an Issue

If profiling reveals that deep copying on toggle is too slow (unlikely for typical use):

1. **Option A**: Path-copying (only copy nodes on path to changed node)
2. **Option B**: Persistent data structures (immutable tree with structural sharing)
3. **Option C**: Accept mutability for tree structure only (document carefully)

But start with full deep copy - it's simple and likely fast enough.

## Conclusion

Approach 1 (Tree Structure with Flattening) is recommended because it:
- Matches Splice's existing patterns (graph layout)
- Provides predictable, acceptable performance
- Keeps path collapsing simple and natural
- Maintains code clarity and testability
- Avoids premature optimization (Approach 3)
- Avoids performance pitfalls (Approach 2)

The tree structure is the natural representation of hierarchical file paths, and the flattening step cleanly separates the data model from the navigation/rendering concerns. This follows Splice's principle of deep functions and clear separation of concerns.
