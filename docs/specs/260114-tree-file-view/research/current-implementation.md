# Current File View Implementation

## Overview
The file view displays a flat list of changed files in a commit or uncommitted changes.

## Data Structure

**FileChange** (`internal/core/git_types.go:31-38`)
```go
type FileChange struct {
    Path      string // File path relative to repository root
    Status    string // M (modified), A (added), D (deleted), R (renamed)
    Additions int    // Number of lines added
    Deletions int    // Number of lines deleted
    IsBinary  bool   // True if the file is binary
}
```

**FilesState** (`internal/ui/states/files/state.go:8-13`)
```go
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange  // Flat list of files
    Cursor        int                // Current selection
    ViewportStart int                // Scroll position
}
```

## Current Display Format

Files are displayed as a flat list with each file showing:
```
→ M +17 -13  src/components/App.tsx
  A  +42  -0  src/utils/helper.ts
  D   +0 -25  old/deprecated.ts
```

Format breakdown:
- `→` = cursor indicator (selected file)
- `M`/`A`/`D`/`R` = status (colored by type)
- `+17 -13` = additions/deletions (right-aligned, colored)
- Full file path

## Rendering Logic

**FileSection component** (`internal/ui/components/file_section.go`)
- Renders: blank line + stats summary + file lines
- Stats summary: `5 files · +234 -45`
- Each file formatted via `FormatFileLine()`

**Key characteristics:**
- Files stored as flat `[]core.FileChange` array
- Paths are full strings (e.g., `internal/ui/states/files/view.go`)
- No folder grouping or hierarchy
- Scrolling viewport shows subset of files
- Cursor navigation moves through flat list

## Navigation
- Up/Down: Move cursor through flat list
- Enter: Open diff view for selected file
- Viewport scrolls to keep cursor visible

## Limitations for Tree View
1. Files are stored as flat list - no hierarchy
2. Full paths displayed for every file
3. No folder grouping or collapsing
4. Inefficient for commits with many files in deep paths
