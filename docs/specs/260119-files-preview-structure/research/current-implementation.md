# Current Implementation: Files Preview vs Files View

## Overview

There are two places where files are displayed in Splice:

1. **Files Preview** - Right panel in the log view (split view mode)
2. **Files View** - Dedicated full-screen state when you press Enter on a commit

Currently, these two views display files differently.

## Files Preview (Log State - Right Panel)

**Location**: `internal/ui/states/log/view.go:222`

**Component Used**: `components.FileSection`

**Display Format**: Flat list of files
```
M +17 -13  src/components/App.tsx
A +42 -0   src/utils/helper.ts
D +0  -18  old/deprecated.js
```

**Characteristics**:
- No folder hierarchy
- No tree structure
- Simple linear list of all files
- Shows status, additions/deletions, and full path

**Code**:
```go
fileSectionLines := components.FileSection(preview.Files, width, nil)
```

## Files View (Dedicated State)

**Location**: `internal/ui/states/files/view.go:39`

**Component Used**: `components.TreeSection`

**Display Format**: Tree structure with collapsible folders
```
📁 src/
  📁 components/
    → M +17 -13  App.tsx
  📁 utils/
    A +42 -0   helper.ts
📁 old/
  D +0  -18  deprecated.js
```

**Characteristics**:
- Hierarchical tree structure
- Folders can be collapsed/expanded
- Shows folder icons and indentation
- Uses `filetree.VisibleTreeItem` data structure
- Interactive navigation (up/down, collapse/expand with h/l)

**Code**:
```go
treeSectionLines := components.TreeSection(s.VisibleItems, s.Files, s.Cursor, ctx.Width())
```

## Key Differences

| Aspect | Files Preview | Files View |
|--------|---------------|------------|
| Component | `FileSection` | `TreeSection` |
| Structure | Flat list | Tree hierarchy |
| Data Type | `[]core.FileChange` | `[]filetree.VisibleTreeItem` |
| Folders | No folders shown | Collapsible folders |
| Selection | None (display only) | Cursor navigation |
| State Management | None needed | Tracks collapsed/expanded state |

## Implementation Details

### FileSection Component
- File: `internal/ui/components/file_section.go`
- Takes flat `[]core.FileChange` array
- Renders each file as a single line
- No tree logic or state

### TreeSection Component
- File: `internal/ui/components/tree_section.go`
- Takes `[]filetree.VisibleTreeItem` (pre-processed tree structure)
- Renders with tree symbols and indentation
- Relies on `filetree` package for tree building and collapse logic

### FiletTree Package
- Package: `internal/domain/filetree`
- Builds tree structure from flat file list
- Manages collapse/expand state
- Generates visible items based on state
