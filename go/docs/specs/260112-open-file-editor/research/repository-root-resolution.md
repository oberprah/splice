# Research: Repository Root Resolution

## Current State

**Repository root is NOT currently stored or calculated in Splice.**

- `app.Model` does NOT have a repository root field
- `core.Context` interface does NOT provide repository root access
- `FileChange.Path` is stored as relative to repository root

## How File Paths Are Currently Handled

The codebase uses **git's `:(top)` pathspec magic** for git operations:

**Location:** `/home/user/splice/internal/git/git.go:445-467`

```go
cmd := exec.Command("git", "show", commitHash, "--format=", "--", ":(top)"+filePath)
```

This works for git operations but **does NOT help with opening files in an editor**, which requires absolute paths.

## How to Determine Repository Root

Use the standard git command:

```bash
git rev-parse --show-toplevel
```

This command:
- Returns the absolute path to the repository root
- Works regardless of current working directory
- Is well-supported across all git versions
- Returns an error if not in a git repository

## Implementation Approach

Add a new function to the git package:

```go
// Add to /home/user/splice/internal/git/git.go

func GetRepositoryRoot() (string, error) {
    cmd := exec.Command("git", "rev-parse", "--show-toplevel")

    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("failed to get repository root: %v", err)
    }

    return strings.TrimSpace(out.String()), nil
}
```

## Path Resolution Workflow

```
FileChange.Path (relative)        e.g., "internal/ui/app.go"
         ↓
git rev-parse --show-toplevel     e.g., "/home/user/splice"
         ↓
filepath.Join(root, path)         e.g., "/home/user/splice/internal/ui/app.go"
         ↓
Launch editor with absolute path
```

## Access from Diff State

The diff state can call `git.GetRepositoryRoot()` directly when handling the 'o' key:

```go
repoRoot, err := git.GetRepositoryRoot()
if err != nil {
    // show error
}
absolutePath := filepath.Join(repoRoot, s.File.Path)
```

## Imports Needed

```go
import (
    "path/filepath"  // For filepath.Join()
)
```

The `path/filepath` package is already used in test files throughout the codebase.
