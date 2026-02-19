# Design: Files Preview Tree Structure

## Executive Summary

This design replaces the flat file list in the log view's preview panel with a tree structure, matching the dedicated files view. The implementation leverages the existing `filetree` package and `TreeSection` component, while also fixing an inconsistency in the codebase.

**Key insights:**
- The `filetree` package already creates fully-expanded trees by default, so no special handling is needed for the "all folders expanded" requirement
- `FileSection` is only used in the log preview - we can delete it entirely after switching to `TreeSection`
- `TreeSection` currently uses magic number cursor (-1 for no selection) while `FileSection` properly uses nullable cursor (`*int`) - we should fix this inconsistency

**Changes:**
1. Refactor `TreeSection` to use `cursor *int` instead of `cursor int` (removes magic number)
2. Update files state to work with nullable cursor
3. Use `TreeSection` with `nil` cursor in log preview
4. Delete `FileSection` component and tests entirely

**Net impact:** Code reduction (delete entire component) + improved design (remove magic number) + feature delivery (tree in preview).

## Context & Problem Statement

The log view shows a preview panel (in wide terminals) displaying files changed in the selected commit. Currently this preview uses `FileSection` to show a flat list:

```
M +17 -13  src/components/App.tsx
A +42 -0   src/utils/helper.ts
```

The dedicated files view (reached by pressing Enter) uses `TreeSection` to show a tree structure:

```
└── src/
    ├── components/
    │   └── M +17 -13  App.tsx
    └── utils/
        └── A +42 -0   helper.ts
```

This inconsistency creates cognitive overhead when switching between views. Users must mentally map between two different representations of the same data.

**Scope:** This design covers the log view's preview panel (switching to tree structure) and refactors the `TreeSection` component (removing magic number). The files view behavior remains the same, just uses the refactored signature.

## Current State

### Preview Panel Architecture

The log state (`internal/ui/states/log/`) manages the preview panel:

1. **Data Loading** - When cursor moves, log state asynchronously loads file changes for the selected commit/range
2. **Preview State** - Tracks loading status as sum type: `PreviewNone | PreviewLoading | PreviewError | PreviewLoaded`
3. **Rendering** - `renderFileList()` method checks preview state and renders appropriate content

**Current rendering code** (`view.go:222`):
```go
case PreviewLoaded:
    // Use FileSection component to render files
    fileSectionLines := components.FileSection(preview.Files, width, nil)
    // ... truncation logic ...
```

### TreeSection Component

The `TreeSection` component (`internal/ui/components/tree_section.go`) currently takes:
- `items []filetree.VisibleTreeItem` - Tree structure to render
- `files []core.FileChange` - Original flat list (for header stats)
- `cursor int` - Index of selected item (**magic number -1 for no selection**)
- `width int` - Panel width

Returns formatted lines with tree symbols, indentation, and styling.

**Design inconsistency:** `FileSection` uses `cursor *int` (nil for no selection) while `TreeSection` uses `cursor int` (-1 for no selection). We should fix `TreeSection` to match the better pattern.

### Filetree Package

The `filetree` package (`internal/domain/filetree/`) provides:

**Tree Building Pipeline:**
```
[]core.FileChange → BuildTree() → FolderNode (root)
                 → CollapsePaths() → optimized tree
                 → ApplyStats() → tree with statistics
                 → FlattenVisible() → []VisibleTreeItem
```

**Default Behavior:** All folders start expanded (`isExpanded: true`). The files state manages collapse/expand interactions, but the default state already matches our requirement.

## Solution

### High-Level Approach

This change has two parts:

**Part 1: Refactor TreeSection cursor handling (fix magic number)**
1. Change `TreeSection` signature from `cursor int` to `cursor *int`
2. Update implementation to check `cursor != nil && *cursor == i`
3. Update files state to pass `&s.Cursor` instead of `s.Cursor`

**Part 2: Use TreeSection in log preview (deliver feature)**
1. Build tree structure from flat file list in log preview
2. Call `TreeSection` with `nil` cursor (no selection)
3. Delete `FileSection` component and tests (no longer needed)

**Net impact:** Better design (no magic numbers) + code reduction (delete unused component) + consistent UX (tree in both views).

### Data Flow

**Before:**
```
preview.Files ([]FileChange)
    ↓
FileSection(files, width, nil)
    ↓
lines (flat list)
```

**After:**
```
preview.Files ([]FileChange)
    ↓
buildTreeForPreview(files) → []VisibleTreeItem
    ↓
TreeSection(items, files, nil, width)
    ↓
lines (tree structure)
```

### Component Interaction

**Before (current state):**
```
LogState.renderFileList()
    │
    └─→ FileSection(files, width, nil)
            └─→ Renders flat list

FilesState.View()
    │
    └─→ TreeSection(items, files, cursor, width)  ← cursor: int (magic -1)
            └─→ Renders tree
```

**After (this change):**
```
LogState.renderFileList()
    │
    ├─→ buildTreeForPreview(files)
    │       └─→ filetree.BuildTree()
    │       └─→ filetree.CollapsePaths()
    │       └─→ filetree.ApplyStats()
    │       └─→ filetree.FlattenVisible()
    │
    └─→ TreeSection(items, files, nil, width)  ← cursor: *int (nil)
            └─→ Renders tree

FilesState.View()
    │
    └─→ TreeSection(items, files, &cursor, width)  ← cursor: *int (&s.Cursor)
            └─→ Renders tree

[FileSection deleted - no longer needed]
```

### Tree Building for Preview

**New helper function** (location: `internal/ui/states/log/view.go`):

```go
// buildTreeForPreview builds a fully-expanded tree from flat file list.
// Returns visible items ready for TreeSection rendering.
func buildTreeForPreview(files []core.FileChange) []filetree.VisibleTreeItem {
    // Build tree - all folders start expanded
    root := filetree.BuildTree(files)

    // Collapse single-child folder chains (e.g., "src/components/nested")
    root = filetree.CollapsePaths(root)

    // Apply aggregate statistics to folders
    filetree.ApplyStats(root)

    // Flatten to visible items (all folders expanded, so all items visible)
    return filetree.FlattenVisible(root)
}
```

**Why this works:**
- `BuildTree()` creates all folders with `isExpanded: true` by default
- `FlattenVisible()` includes all descendants of expanded folders
- Result: complete tree structure with all files visible
- No state management needed since we rebuild on every render

### Rendering Changes

**Modified `renderFileList()` method** (location: `internal/ui/states/log/view.go:191`):

```go
case PreviewLoaded:
    // Check preview is current
    currentRangeHash := getRangeHash(s.GetSelectedRange())
    if preview.ForHash != currentRangeHash {
        // Stale data, show loading
        lines = append(lines, "")
        lines = append(lines, styles.TimeStyle.Render("Loading files..."))
    } else {
        // Build tree structure from files
        visibleItems := buildTreeForPreview(preview.Files)

        // Render using TreeSection (nil cursor means no selection)
        treeSectionLines := components.TreeSection(visibleItems, preview.Files, nil, width)

        // Apply truncation logic (same as before)
        if len(treeSectionLines) > maxLines {
            // ... existing truncation logic ...
        } else {
            lines = append(lines, treeSectionLines...)
        }
    }
```

**Key changes:**
1. Build tree before rendering: `buildTreeForPreview(preview.Files)`
2. Call `TreeSection` instead of `FileSection`
3. Pass `nil` as cursor (no selection in preview)
4. Keep same truncation logic for overflow handling

### Refactoring TreeSection Cursor Handling

> **Decision:** Change `TreeSection` to use `cursor *int` instead of `cursor int`, matching `FileSection`'s pattern.

**Current state (magic number):**
```go
func TreeSection(items []filetree.VisibleTreeItem, files []core.FileChange, cursor int, width int) []string {
    // ...
    for i, item := range items {
        isSelected := i == cursor  // Magic: -1 means no selection
        line := format.FormatTreeLine(item, isSelected)
        // ...
    }
}
```

**After refactoring (nullable):**
```go
func TreeSection(items []filetree.VisibleTreeItem, files []core.FileChange, cursor *int, width int) []string {
    // ...
    for i, item := range items {
        isSelected := cursor != nil && *cursor == i  // Explicit: nil means no selection
        line := format.FormatTreeLine(item, isSelected)
        // ...
    }
}
```

**Rationale:**
- Removes magic number (-1)
- Makes "no selection" explicit (nil vs present value)
- Matches `FileSection`'s existing pattern
- Type system enforces correct usage
- Follows Go idioms for optional values

**Impact on callers:**
- Files state: Change from `TreeSection(items, files, s.Cursor, width)` to `TreeSection(items, files, &s.Cursor, width)`
- Log preview: Use `TreeSection(items, files, nil, width)` for no selection

### Truncation Handling

Tree rendering produces more lines than flat list (folders add lines). The preview panel has limited height, so we need truncation.

**Current truncation** (lines 227-244 in `view.go`):
1. Keep blank line and stats line (first 2 lines)
2. Show as many file lines as fit
3. Add overflow indicator if truncated: "... and N more files"

**Updated truncation:**
1. Keep blank line and stats line (first 2 lines)
2. Show as many tree lines as fit
3. Update overflow indicator: "... and N more items" (not "files" since we may truncate folders too)

The existing logic structure works, just need to update the wording of the overflow message.

### Performance Considerations

**Tree building cost:**
- `BuildTree()`: O(n×d) where n=files, d=average depth (~5)
- `CollapsePaths()`: O(nodes) single traversal
- `ApplyStats()`: O(nodes) single traversal
- `FlattenVisible()`: O(nodes) single traversal

**Typical commits:** 5-50 files, 2-5 levels deep
- Total nodes: ~10-100 (files + folders)
- Total time: <1ms on modern hardware

**Rebuilding on every render:**
- Preview renders only when cursor moves or preview loads
- Not in hot path (happens ~1-10 times per second max)
- Cost is negligible compared to git operations

> **Decision:** Rebuild tree on every render rather than caching in state.

**Rationale:**
- Simple: no state management, no cache invalidation
- Performance is acceptable for typical use
- Matches preview's stateless nature
- Avoids complexity of tracking when to rebuild

**Alternative considered:** Cache tree in PreviewLoaded state
- **Rejected:** Adds state complexity, negligible benefit for typical usage

### Deleting FileSection Component

> **Decision:** Delete `FileSection` component and tests after switching log preview to use `TreeSection`.

**Rationale:**
- `FileSection` is only used in one place: `log/view.go:222`
- After this change, it will have zero usages
- Keeping dead code increases maintenance burden
- TreeSection provides all the same functionality plus tree structure

**Impact:**
- Delete `internal/ui/components/file_section.go` (~189 lines)
- Delete `internal/ui/components/file_section_test.go` (~200 lines)
- Net code reduction: ~389 lines removed

**Alternative considered:** Keep FileSection for potential future use
- **Rejected:** YAGNI (You Aren't Gonna Need It) - we can always restore from git if needed

## Open Questions

None. The design is straightforward and uses existing, well-tested infrastructure.

## Implementation Checklist

For reference during implementation phase:

**Part 1: Refactor TreeSection cursor**
- [ ] Change `TreeSection` signature to use `cursor *int`
- [ ] Update TreeSection implementation to handle nil cursor
- [ ] Update files state to pass `&s.Cursor` to TreeSection
- [ ] Update TreeSection tests to cover nil cursor
- [ ] Update files state tests (golden files may change due to signature change)

**Part 2: Use TreeSection in log preview**
- [ ] Add `buildTreeForPreview()` helper function in `log/view.go`
- [ ] Update `renderFileList()` to use TreeSection with nil cursor
- [ ] Update truncation overflow message from "files" to "items"
- [ ] Update golden file tests for log state preview rendering

**Part 3: Delete FileSection**
- [ ] Delete `internal/ui/components/file_section.go`
- [ ] Delete `internal/ui/components/file_section_test.go`

**Verification**
- [ ] Verify visual consistency between preview and files view
- [ ] Test with various file counts and directory depths
- [ ] Verify files state still works correctly with refactored TreeSection
