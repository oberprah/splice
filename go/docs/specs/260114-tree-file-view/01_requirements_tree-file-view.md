# Requirements: Tree File View

**Status**: Draft (awaiting approval)
**Created**: 2026-01-14

## Problem Statement

The current file view displays files as a flat list with full paths, creating three key problems:

1. **Hard to scan/navigate** - Long file paths and many files make it difficult to quickly find specific files
2. **Hard to understand structure** - No visual representation of which files are in which folders
3. **Cognitive load** - Repeated path prefixes (e.g., `src/components/...`, `src/components/...`) make the list harder to read

These issues compound with commits that touch many files in deep directory structures.

## Goals

1. **Visual hierarchy** - Display files in a tree structure that clearly shows folder relationships
2. **Reduce clutter** - Collapse boring paths (single-child folders) and show aggregate stats for collapsed folders
3. **Maintain speed** - Keep keyboard-driven navigation fast and intuitive
4. **Backward compatible UX** - Preserve core behaviors (cursor navigation, entering diffs) while enhancing the view

## Non-Goals

The following are explicitly out of scope for this feature:

- Filtering or searching within the tree
- Remembering expansion state across different commits
- Bulk expand/collapse operations (expand all, collapse all)
- Visual indication of folders with changes vs empty nested folders
- Configuration options for tree display style

Focus is purely on tree display and basic expand/collapse navigation.

## User Impact

**Who benefits**: All Splice users, especially those working on:
- Projects with deep folder hierarchies
- Commits touching many files
- Codebases with long path names

**Expected improvements**:
- Faster file location and comprehension
- Reduced eye strain from reading repeated paths
- Better understanding of change scope at a glance

## Functional Requirements

### FR1: Tree Structure Display

Files must be displayed in a hierarchical tree structure using tree characters (like Unix `tree` command):

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

**Rules**:
- Use tree drawing characters: `├──`, `└──`, `│`
- Each file shows: status indicator, stats (+N -N), and filename only (not full path)
- File format matches current display: `M +17 -13  App.tsx`

### FR2: Item Ordering

Within each folder level:
1. **Folders first** - All subfolders grouped at top
2. **Then files** - All files below folders
3. **Alphabetically sorted** - Each group sorted alphabetically

Example:
```
src/
├── components/     (folder - alphabetically first)
├── utils/          (folder - alphabetically second)
├── M +5 -2  App.tsx    (file - alphabetically first)
└── A +10 -0 index.ts   (file - alphabetically second)
```

### FR3: Path Collapsing

**Single-child folders** (folders containing only one subfolder, no files) must be collapsed into a single line:

```
src/components/nested/deep/    (collapsed path)
└── M +17 -13  App.tsx
```

**Single-file folders** (folders containing only one file, no subfolders) display normally:

```
utils/
└── M +10 -5  helper.ts
```

### FR4: Default Expansion State

All folders must be expanded by default when viewing a commit's files.

### FR5: Folder Collapse/Expand

**Collapsed state**:
- Shows folder path with aggregate stats and file count
- Format: `foldername/ +N -N (X files)`
- Example: `src/ +234 -67 (5 files)`

**Expanded state**:
- Shows folder path without stats
- Children visible below
- Example: `src/`

**Collapsed path expansion**:
- When expanding a collapsed path like `src/components/nested/deep/`, expand the entire path at once, revealing all intermediate folders

### FR6: Navigation Controls

**Folder toggle** (expand/collapse):
- Right arrow - Expand folder
- Left arrow - Collapse folder
- Enter - Toggle folder
- Space - Toggle folder

**Cursor movement**:
- Up/Down arrows - Move cursor through all visible items (folders and files)
- Cursor can land on any folder or file

**File selection**:
- Enter on file - Open diff view (existing behavior)

### FR7: Visual Selection

Cursor indicator (`→`) shows selected item, works for both folders and files (consistent with current behavior).

## Non-Functional Requirements

### NFR1: Performance

Tree construction and rendering must not introduce noticeable lag for typical commit sizes (< 1000 files).

### NFR2: Consistency

- Tree drawing characters must render correctly in all supported terminals
- Folder/file styling must follow existing color scheme
- Selection highlighting must match current files view behavior

## Open Questions

None - ready for design phase.

## References

- [Current Implementation Research](research/current-implementation.md) - Details of existing flat file list
