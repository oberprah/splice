# Git Diff Commands Research

Research findings for git commands needed to support uncommitted changes and direct diff viewing.

## Current Implementation (Committed Ranges)

From `internal/git/git.go`:
- `git diff --name-status <hash>^..<hash>` - List files with status
- `git diff --numstat <hash>^..<hash>` - File changes with line counts
- `git show <hash>:<path>` - File content at specific commit
- `git diff <hash>^..<hash> -- <path>` - Unified diff for single file

## Uncommitted Changes Commands

### File List Commands

| Scenario | Command for name-status | Command for numstat |
|----------|------------------------|-------------------|
| **Unstaged** (working tree vs index) | `git diff --name-status` | `git diff --numstat` |
| **Staged** (index vs HEAD) | `git diff --staged --name-status` | `git diff --staged --numstat` |
| **All uncommitted** (working tree vs HEAD) | `git diff HEAD --name-status` | `git diff HEAD --numstat` |

### File Content Commands

#### Unstaged Changes (working tree vs index)
- **Old version** (from index): `git show :path`
- **New version** (from working tree): Read from filesystem
- **Unified diff**: `git diff -- path`

#### Staged Changes (index vs HEAD)
- **Old version** (from HEAD): `git show HEAD:path`
- **New version** (from index): `git show :path`
- **Unified diff**: `git diff --staged -- path`

#### All Uncommitted (working tree vs HEAD)
- **Old version** (from HEAD): `git show HEAD:path`
- **New version** (from working tree): Read from filesystem
- **Unified diff**: `git diff HEAD -- path`

## Key Technical Details

### Index File Syntax

The colon syntax `git show :path` references files in the index (staging area):
```bash
git show :path              # Get staged version of file
git show --cached -- path   # Alternative syntax (more verbose)
```

This is distinct from commit references:
```bash
git show HEAD:path          # File from HEAD commit
git show abc123:path        # File from specific commit
git show :path              # File from index/staging area
```

### Flag Synonyms

`--staged` and `--cached` are exact synonyms in git diff:
```bash
git diff --staged == git diff --cached
```

Both are equally valid; `--staged` is more intuitive for users.

### Working Tree File Access

Working tree files cannot be retrieved via git commands - must read directly from filesystem:
```go
// For working tree files
content, err := os.ReadFile(path)
```

This applies to:
- Unstaged changes (new version)
- All uncommitted changes (new version)

### Edge Cases

**Deleted Files:**
- Old content exists in git
- New content is empty/missing
- Handle file read errors gracefully

**Added Files:**
- Old content doesn't exist in git (handle "does not exist" error)
- New content from working tree

**Renamed Files:**
- Git diff shows as deletion + addition or with rename status (R)
- May need special handling depending on status parsing

**Binary Files:**
- Git diff marks them as binary
- Should skip content fetch and show "Binary file" message

## Commit Range Variants

### Two-Dot Range
```bash
git diff main..feature      # Changes in feature not in main
```
Equivalent to `git diff main feature`

### Three-Dot Range
```bash
git diff main...feature     # Changes since common ancestor
```
Useful for viewing what changed on a branch since it diverged.

## Implementation Recommendations

### New Functions Needed in `git.go`

```go
// File list fetchers
func FetchUnstagedFileChanges() ([]FileChange, error)
func FetchStagedFileChanges() ([]FileChange, error)
func FetchAllUncommittedFileChanges() ([]FileChange, error)

// File content fetchers
func FetchIndexFileContent(path string) (string, error)
func FetchWorkingTreeFileContent(path string) (string, error)

// Diff fetchers
func FetchUnstagedFileDiff(path string) (*FullFileDiffResult, error)
func FetchStagedFileDiff(path string) (*FullFileDiffResult, error)
func FetchUncommittedFileDiff(path string) (*FullFileDiffResult, error)
```

### Unified Abstraction

Consider creating a unified interface that works for both committed and uncommitted diffs:

```go
type DiffSource interface {
    FetchFileChanges() ([]FileChange, error)
    FetchFileContent(path string) (old, new string, error)
    FetchUnifiedDiff(path string) (string, error)
}
```

This allows the same code path for Files/Diff states regardless of diff source type.

## Exit Codes and Validation

Git diff exit codes:
- `0`: No differences (or command succeeded with no output)
- `1`: Differences exist
- `128+`: Error (invalid ref, bad path, etc.)

Use `--quiet` flag to check without output:
```bash
git diff --quiet <spec>
# Exit 0 = no changes
# Exit 1 = has changes
# Exit 128+ = error
```

## Path Handling

All paths should be relative to repository root. Use `:(top)` pathspec prefix for safety:
```bash
git show :./path              # Relative to current directory
git show :./(top)path         # Relative to repo root
```

Or ensure all operations happen from repo root.
