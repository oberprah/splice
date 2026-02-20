# Current Files View Implementation

## Commit Header Format

Location: `internal/ui/states/files_view.go:44-91`

The files view displays a multi-line header:
```
abc123d · John Doe committed 2 hours ago · 3 files · +45 -12

Subject line

Body paragraph 1...
Body paragraph 2...
```

Header components:
- Short hash (7 chars) styled with `HashStyle`
- Author name styled with `AuthorStyle`
- Relative time styled with `TimeStyle`
- File count with proper pluralization
- Total additions/deletions across all files

## File Entry Rendering

Location: `internal/ui/states/files_view.go:127-179`

File line format:
```
> M +17 -13  src/components/App.tsx
  A +130 -0  src/components/FileList.tsx
  D +0  -45  src/old-file.ts
```

Components per file:
1. Selection indicator (1 char): `>` or space
2. Status letter (1 char): A/M/D/R
3. Additions (right-aligned): `+XX`
4. Deletions (right-aligned): `-XX`
5. File path (remaining space)

Dynamic alignment:
- `calculateMaxStatWidth()` determines column widths
- Stats right-align to match widest value
- Binary files show "(binary)" instead of line counts

## File Change Data Structure

Location: `internal/git/git.go:20-27`

```go
type FileChange struct {
    Path      string  // File path relative to repository root
    Status    string  // Git status: M, A, D, R, etc.
    Additions int     // Lines added
    Deletions int     // Lines deleted
    IsBinary  bool    // True if binary file
}
```

## Data Loading

Git commands used (`internal/git/git.go:169-239`):
1. `git diff-tree --no-commit-id --name-status -r <hash>` - Gets file statuses
2. `git diff-tree --no-commit-id --numstat -r <hash>` - Gets line counts

The function `FetchFileChanges()` merges both outputs into `[]FileChange`.

## Styling

Status letters (A/M/D/R) use default style. Color is applied to stats and paths:

Non-selected lines:
- Additions: `AdditionsStyle` (green)
- Deletions: `DeletionsStyle` (red)
- File path: `FilePathStyle`

Selected lines (bold + brighter):
- Additions: `SelectedAdditionsStyle`
- Deletions: `SelectedDeletionsStyle`
- File path: `SelectedFilePathStyle`

## Reusable Patterns for Split Panel

1. **Header rendering**: Use `renderHeader()` approach but abbreviated
2. **File formatting**: Adapt `formatFileLine()` for narrow columns
3. **Data loading**: Reuse `FetchFileChanges()` command pattern
4. **Styling**: Same style constants for consistency
5. **Stat alignment**: `calculateMaxStatWidth()` ensures visual alignment
