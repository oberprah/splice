# Research: Git Graph Data Extraction

## Executive Summary

Git provides several commands to extract commit topology and graph data. The recommended approach is using `git rev-list --parents` combined with `git show-ref` for ref decorations. This provides all necessary data for rendering a commit graph without relying on git's built-in ASCII graph rendering.

## Key Git Commands

### 1. Primary: `git rev-list --parents`

Provides commit hash followed by parent hash(es):

```bash
git rev-list --parents HEAD -n 10
```

Output:
```
4d2232e b988c86           # Regular commit (1 parent)
b988c86 0479343 aa79aa1   # Merge commit (2 parents)
0479343 9dd98c1           # Regular commit
```

**Advantages:**
- Single efficient command for topology data
- Handles all commit types: linear, merge (2 parents), octopus (3+ parents), root commits
- No parsing of ASCII art required

### 2. Ref Decorations: `git show-ref --head`

Maps refs (branches, tags, HEAD) to commit hashes:

```bash
git show-ref --head
```

Output:
```
4d2232e HEAD
4d2232e refs/heads/feature/git-graph-spec
b988c86 refs/heads/main
```

### 3. Alternative: Extended `git log` Format

Combine commits with decorations in a single command:

```bash
git log --pretty=format:"%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e" --date=iso-strict
```

Format placeholders:
- `%H` - Full commit hash
- `%P` - Parent hashes (space-separated)
- `%d` - Ref decorations (e.g., " (HEAD -> main, origin/main)")
- `%an` - Author name
- `%ad` - Author date
- `%s` - Subject line
- `%b` - Body

### 4. Reference: `git log --graph`

Git's built-in ASCII art renderer - useful for understanding expected output but not for direct use:

```bash
git log --graph --oneline -n 5
```

## Data Structures

### Extended GitCommit

```go
type GitCommit struct {
    Hash         string       // Full 40-character hash
    Message      string       // First line (subject) of commit message
    Body         string       // Everything after first line
    Author       string       // Author name only (no email)
    Date         time.Time    // Commit timestamp
    ParentHashes []string     // NEW: Parent commit hashes
    Refs         []RefInfo    // NEW: Branch/tag decorations
}

type RefInfo struct {
    Name   string   // e.g., "main", "feature-x", "v1.0"
    Type   RefType  // HEAD, Branch, RemoteBranch, Tag
    IsHead bool     // true if this is the current HEAD
}

type RefType int

const (
    RefTypeBranch RefType = iota
    RefTypeRemoteBranch
    RefTypeTag
)
```

### Graph Layout Data

```go
type GraphLayout struct {
    CommitColumns map[string]int    // commitHash -> column index
    MaxColumns    int               // Maximum width of graph
    RowData       []GraphRow        // One per commit, in display order
}

type GraphRow struct {
    CommitHash   string
    Column       int               // Which column has the commit marker
    ActiveLanes  []LaneInfo        // All lanes active on this row
    Characters   []rune            // Pre-rendered characters for each column
    Colors       []int             // Color index for each column
}

type LaneInfo struct {
    Column      int
    BranchID    int               // For color assignment
    IsCommit    bool              // true if this lane has a commit marker
    IsMergeIn   bool              // true if merging into this row
    IsMergeOut  bool              // true if branching out from this row
}
```

## Edge Cases

### Merge Commits (2 Parents)
```
4d2232e b988c86 aa79aa1
        ↑       ↑
        first   second parent
```
First parent is the branch being merged into; second is the branch being merged.

### Octopus Merges (3+ Parents)
```
abc1234 parent1 parent2 parent3 parent4
```
Rare but valid - the Linux kernel has commits with up to 66 parents.

### Root Commits (No Parents)
```
initial123
```
No parent hashes after the commit hash. Multiple root commits possible (orphan branches).

### Detached HEAD
```bash
git show-ref --head
abc1234 HEAD                    # HEAD not pointing to a branch
abc1234 refs/heads/main         # Commit still has branch refs
```
HEAD appears without `refs/heads/` prefix.

## Performance Considerations

- **Data Size**: ~50 bytes per commit for topology data (hash + parent hashes)
- **Command Execution**: `git rev-list` is very fast (< 100ms for 1000 commits)
- **Memory**: Graph layout data ~200-300 bytes per commit
- **Recommendation**: Fetch topology data with initial commit load, compute graph layout once

## Integration with Current Splice Architecture

Current `FetchCommits()` command:
```bash
git log --pretty=format:%H%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

Extended command with parent data:
```bash
git log --pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

Changes needed:
1. Add `%P` (parent hashes) to format string
2. Add `%d` (decorations) to format string
3. Parse new fields in `parseCommits()`
4. Add new fields to `GitCommit` struct
