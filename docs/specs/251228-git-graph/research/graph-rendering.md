# Research: Graph Rendering Algorithms

## Executive Summary

Git commit graph rendering involves three main challenges: column assignment (determining which column each branch occupies), edge layout (drawing connecting lines), and color assignment (distinguishing branches visually). This document synthesizes approaches from Git, lazygit, tig, and academic sources.

## Column Assignment Algorithm

### Core Concept: Active Branches List

The fundamental approach maintains an "active branches" list where each index represents a column:

```
Algorithm: AssignColumns(commits in topological order)
    active_lanes = []

    for each commit:
        # Find if this commit continues an existing lane
        lane = findLaneWithCommit(commit, active_lanes)

        if lane found:
            # Continue in same column
            commit.column = lane.column
            lane.commit = commit.firstParent
        else:
            # Allocate new column
            commit.column = findAvailableColumn(active_lanes)
            active_lanes.add(newLane(commit.column, commit.firstParent))

        # Handle merge parents (secondary parents)
        for each parent in commit.parents[1:]:
            parentLane = findLaneWithCommit(parent, active_lanes)
            if not found:
                # Merge source needs a lane
                allocateLane(parent, active_lanes)
```

### Key Insight: First Parent Rule

- **First parent**: The branch being merged INTO - continues vertically in same column
- **Other parents**: Branches being merged - get diagonal/horizontal edges to their columns

This keeps commits on the same branch aligned vertically, making the graph readable.

## Unicode Box-Drawing Characters

### Essential Character Set

```
Basic Lines:
  │  U+2502  Vertical
  ─  U+2500  Horizontal

Junctions:
  ├  U+251C  Vertical and Right (commit with merge)
  ┤  U+2524  Vertical and Left
  ┬  U+252C  Horizontal and Down
  ┴  U+2534  Horizontal and Up

Merge Curves:
  ╮  U+256E  Arc Down and Left (merge coming from right)
  ╯  U+256F  Arc Up and Left (merge ending)
  ╭  U+256D  Arc Down and Right (branch starting)
  ╰  U+2570  Arc Up and Right (merge coming from left)

Commit Marker:
  *  Standard asterisk for commit point
  ●  U+25CF  Black circle (alternative)
```

### Character Selection Rules

| Scenario | Character |
|----------|-----------|
| Vertical continuation | │ |
| Commit on line | * or ● |
| Merge from right | ╮ then ─ |
| Merge from left | ╭ then ─ |
| Branch ends | ╯ or ╰ |
| Commit with merge | ├ |

## Graph Rendering Examples

### Linear History
```
* abc1234 Add feature
│
* def5678 Update code
│
* 9ab0123 Initial commit
```

### Simple Merge
```
*   abc1234 Merge feature
├─╮
│ * def5678 Feature work
│ │
* │ 9ab0123 Main work
├─╯
* 456cdef Common ancestor
```

### Multiple Branches
```
*     abc1234 Merge all
├─╮─╮
│ │ * def5678 Feature B
│ │ │
│ * │ 9ab0123 Feature A
│ │ │
* │ │ 456cdef Main work
├─╯ │
├───╯
* 789abcd Common ancestor
```

## Color Assignment Strategy

### Branch Color Model

Each unique branch lineage gets a distinct color:

```go
var branchColors = []lipgloss.Color{
    lipgloss.Color("196"), // Bright Red
    lipgloss.Color("46"),  // Bright Green
    lipgloss.Color("226"), // Bright Yellow
    lipgloss.Color("21"),  // Bright Blue
    lipgloss.Color("201"), // Bright Magenta
    lipgloss.Color("51"),  // Bright Cyan
}

func colorForBranch(branchID int) lipgloss.Color {
    return branchColors[branchID % len(branchColors)]
}
```

### Color Propagation

1. Assign color when branch first appears
2. Propagate color down through first-parent chain
3. Merge edges use the color of their source branch

## Width Management

### Estimating Graph Width

```
graphWidth = maxActiveColumns * columnWidth + padding
columnWidth = 2  // "│ " or "├─"
padding = 1      // space before commit info

Example:
  3 active branches → 3 * 2 + 1 = 7 characters
```

### Width Constraints

| Terminal Width | Max Graph Width | Rationale |
|----------------|-----------------|-----------|
| < 80 | 10 chars (5 cols) | Preserve commit info space |
| 80-120 | 16 chars (8 cols) | Balance graph and info |
| 120-160 | 20 chars (10 cols) | More space for graph |
| > 160 (split) | 20 chars (10 cols) | Fixed for split view |

### Overflow Handling

When graph exceeds max width:
1. Truncate rightmost columns
2. Show indicator (e.g., `…`) for truncated branches
3. Continue tracking truncated branches internally

## Handling Complex Scenarios

### Octopus Merges (3+ Parents)

```
*     Merge commit
├─╮─╮─╮
│ │ │ │
│ │ │ * Parent 4
│ │ * │ Parent 3
│ * │ │ Parent 2
* │ │ │ Parent 1 (first)
├─╯ │ │
├───╯ │
├─────╯
```

Render same as multiple sequential merges but with multiple incoming edges.

### Branch Crossings

Avoid by:
1. Ordering columns by commit date (older branches on left)
2. Reassigning columns when branches end
3. Preferring adjacent column allocation

## Implementation Approach

### Data Structures

```go
type GraphRenderer struct {
    Layout    *GraphLayout
    MaxWidth  int
    Colors    []lipgloss.Style
}

type GraphLayout struct {
    Rows []GraphRow
}

type GraphRow struct {
    Columns []GraphCell
}

type GraphCell struct {
    Char    rune
    ColorID int
    IsEmpty bool
}
```

### Rendering Pipeline

```
1. Compute Layout
   - Input: []GitCommit with parent hashes
   - Output: GraphLayout with column assignments

2. Generate Characters
   - For each row, determine character at each column
   - Apply merge/branch rules

3. Apply Colors
   - Assign color ID to each branch
   - Color each cell based on its branch

4. Render to String
   - Convert cells to styled strings
   - Join with commit info
```

## Performance

- **Layout computation**: O(n) where n = number of commits
- **Character generation**: O(n * m) where m = max columns
- **Memory**: ~10 bytes per commit for layout data

For 500 commits with 10 max columns: < 5KB memory, < 5ms computation.

## References

- Git's graph.c implementation
- [pvigier's blog on commit graph algorithms](https://pvigier.github.io/2019/05/06/commit-graph-drawing-algorithms.html)
- lazygit's Go-based graph rendering
- tig's ncurses graph implementation
