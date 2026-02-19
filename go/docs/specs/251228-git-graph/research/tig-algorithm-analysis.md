# Tig Git Graph Algorithm Analysis

## Overview

Research into how tig (text-mode interface for git) implements its git graph visualization, conducted to inform Splice's graph implementation design.

**Date**: 2025-12-30
**Primary Source**: [tig repository - graph.c](https://github.com/jonas/tig/blob/master/src/graph.c)

## Tig's Implementation Architecture

Tig maintains **two separate graph algorithm implementations** (graph-v1.c and graph-v2.c), indicating iterative refinement. The graph-v2 implementation is the current standard and is the focus of this analysis.

### Core Design Pattern

- **Callback-based architecture** with function pointers
- Different graph implementations pluggable via `init_graph()`
- Allows switching between ASCII, Unicode, and ncurses rendering modes

## Data Structures

### Graph Columns and Rows

```c
struct graph_column {
    struct graph_symbol symbol;
    const char *id;  // Parent SHA1 ID
};

struct graph_row {
    size_t size;
    struct graph_column *columns;
};
```

**Multi-row state management:**
- `prev_row` - Previous commit
- `row` - Current commit being processed
- `next_row` - Upcoming commit
- `parents` - Parent commits

This three-row lookahead window allows the algorithm to determine connection patterns before rendering.

### Graph Symbol Structure

```c
struct graph_symbol {
    unsigned int color:8;
    unsigned int commit:1;
    unsigned int boundary:1;
    unsigned int initial:1;
    unsigned int merge:1;
    unsigned int continued_down:1;
    unsigned int continued_up:1;
    unsigned int continued_right:1;
    unsigned int continued_left:1;
    unsigned int continued_up_left:1;
    unsigned int parent_down:1;
    unsigned int parent_right:1;
    unsigned int below_commit:1;
    unsigned int flanked:1;
    unsigned int next_right:1;
    unsigned int matches_commit:1;
    unsigned int shift_left:1;
    unsigned int continue_shift:1;
    unsigned int below_shift:1;
    unsigned int new_column:1;
    unsigned int empty:1;
};
```

These 40+ bit flags track properties that determine visual representation:
- **Directional continuity**: up, down, left, right
- **Merge status**: merge, boundary, initial
- **Positioning**: shift_left, new_column, empty
- **Connection patterns**: parent_down, parent_right, flanked

## Algorithm Flow (Graph-v2)

### 1. Initialization

```c
init_graph_v2()
```

- Allocates graph context
- Registers rendering function pointers
- Initializes color hash table for tracking color assignments

### 2. Per-Commit Processing

```
1. Intern ID: Store commit SHA using intern_string() for memory efficiency
2. Locate Position: Find column via graph_find_column_by_id()
3. Register Parents: Insert parent IDs into the parents row
4. Expand Capacity: Grow prev_row, row, next_row if needed
5. Generate Symbols: Populate symbol flags based on continuity patterns
6. Commit Row: Rotate row buffers (next_row → row → prev_row)
7. Collapse Width: Remove trailing empty columns
8. Assign Colors: Map IDs to palette indices with load-balancing
```

### 3. Symbol Generation

The `graph_generate_symbols()` function evaluates each column position:

```c
for (pos = 0; pos < row->size; pos++) {
    symbol->commit = (pos == graph->position);
    symbol->continued_down = continued_down(row, next_row, pos);
    symbol->continued_up = continued_down(prev_row, row, pos);
    symbol->parent_down = parent_down(parents, next_row, pos);
    symbol->shift_left = shift_left(row, prev_row, pos);
    symbol->color = get_color(graph, id);
}
```

Each symbol's properties derive from comparing column IDs across previous, current, and next rows.

### 4. Merge and Branch Handling

**Merge Detection:**
- Identified through `graph_is_merge()`
- Checks for multiple parent symbols

**Branch Processing:**
- When `shift_left` conditions trigger, columns collapse rightward
- Maintains visual continuity while reducing width

**Column Management:**
- `graph_expand()` - Adds columns when parents exceed capacity
- `graph_collapse()` - Removes trailing empty columns
- `graph_remove_collapsed_columns()` - Eliminates redundant columns when commits share identity

### 5. Color Assignment

Uses hash-based storage with load balancing:

```c
struct colors {
    htab_t id_map;
    size_t count[GRAPH_COLORS];
};
```

**Strategy:**
- Hash table maps commit IDs to color indices
- `colors_get_free_color()` implements load-balancing
- Tracks usage counts per color
- Assigns colors with lowest frequency
- Colors persist until `colors_remove_id()` purges entries

**Benefits:**
- Avoids adjacent branches having same color
- More visually distinguishable graph
- Better color distribution across complex histories

### 6. Unicode Character Rendering

The `graph_symbol_to_utf8()` function maps flag combinations to Unicode box-drawing characters:

```c
if (graph_symbol_cross_merge(symbol))
    return "─┼";
if (graph_symbol_vertical_merge(symbol))
    return "─┤";
```

**Characters used:**
- Commit markers: `◯` (boundary), `◎` (initial), `●` (merge), `∙` (regular)
- Lines: `─` (horizontal), `│` (vertical)
- Junctions: `┼` (cross), `┤` (tee right), `├` (tee left)
- Corners: `╭` `╮` (rounded corners), `┴` `┬` (tees)

### 7. Multiple Output Format Support

Three symbol conversion functions:
- `graph_symbol_to_utf8()` - Unicode characters (default)
- `graph_symbol_to_ascii()` - ASCII fallback for limited terminals
- `graph_symbol_to_chtype()` - ncurses terminal graphics using ACS_* constants

## Key Design Decisions

### 1. Real-time Layout

- Processes commits incrementally without forward-scanning
- Uses sliding three-row window (prev, current, next)
- No need to load entire history into memory
- Efficient for large repositories

### 2. Column Reuse

- Rather than creating new columns constantly, searches for existing commit IDs
- Claims first empty column via `graph_find_column_by_id()`
- Keeps graph width manageable

### 3. Symbol Pattern Matching

- Complex merges handled through intelligent symbol pattern matching
- Avoids expensive graph traversal algorithms
- Flag-based approach enables fast character lookup

### 4. Memory Efficiency

- String interning via `intern_string()` for commit SHA storage
- Reuses same string pointer for identical SHAs
- Significant memory savings in large repositories

## Comparison: QGit's Alternative Approach

QGit implements a different "lanes" strategy focused on state tracking.

### Core Data Structure

```cpp
class Lanes {
    QVector<int> typeVec;           // Visual state of each lane
    QVector<QByteArray> nextShaVec; // Expected commit SHA in each lane
    int activeLane;                 // Current commit's lane
    bool boundary;                  // Boundary commit flag
};
```

### Lane Types

- `NODE` - Regular commit point
- `NODE_L`, `NODE_R` - Left/right-side nodes
- `BRANCH` - Branch marker
- `TAIL`, `TAIL_L`, `TAIL_R` - Fork points and corners
- `JOIN` - Merge points
- `CROSS` - Lines passing through without connecting
- `EMPTY` - Available lane slot

### Key Differences from Tig

**Lane Allocation:**
```
"first check empty lanes starting from pos"
```

QGit's approach:
1. Searches for EMPTY lanes and repopulates them
2. Only appends new lanes when all existing ones are occupied
3. More aggressive about keeping graph compact

**State Management:**
- Maintains persistent lanes across commits
- Visual representation changes based on commit relationships
- `findNextSha()` enables efficient position lookups

**Branch/Merge Handling:**
- `setFork()` - Marks fork ranges with TAIL indicators
- `setMerge()` - Identifies parent commits across lanes
- `afterMerge()` - Converts CROSS lanes to NOT_ACTIVE
- `changeActiveLane()` - Creates branches for new commit SHAs

## General Algorithm Principles

Based on research from multiple sources (pvigier's blog, DoltHub implementation), the standard approach involves:

### 1. Topological Ordering

**Temporal topological sorting** determines vertical positions:
- Process commits from newest to oldest (HEAD to initial)
- Use depth-first search with post-order traversal
- Ensures parents always appear below children
- Time complexity: O(n log(n) + m)

### 2. The "Active Lanes" Algorithm

Maintains a list tracking occupied columns:

```
procedure straight_branches(C)
    Initialize empty active branches list B
    for each commit c (lowest to highest row):
        compute forbidden columns J(c)
        if valid branch child exists (not in J(c)):
            select and replace it with c
        else:
            insert c in first nil slot or append
        set c's column = index in B
```

**Techniques:**
- Set entries to `nil` rather than removing (preserves positions)
- Compute forbidden coordinates to prevent overlap
- Prefer available `nil` slots before appending

### 3. Branch Children vs. Merge Children

Critical distinction for proper layout:

**Branch Children:**
- Definition: First parent of child equals current commit
- Edges follow same column (no crossing)
- One branch child can "replace" parent in active branches
- Continues branch lineage

**Merge Children:**
- Definition: First parent of child differs from current commit
- Edges cross columns (require careful placement)
- Multiple merge children possible per commit
- Ends a branch through integration

### 4. Processing Direction

Modern implementations process **from HEAD to initial** (reverse topological order):
- Children positioned before parents
- Parent's column depends on already-determined child positions
- Enables cascading calculation

**Three cases for column assignment:**

1. **No children**: Head commits → new column
2. **Branch children present**: Parent adopts leftmost branch child's column
3. **Only merge children**: Search from leftmost child's column forward for available slot

### 5. First Parent Convention

Git's first parent has special meaning:
- Represents the branch lineage (branch being merged INTO)
- Subsequent parents are branches being merged IN
- Keeps commits on same logical branch in same column
- Critical for visual branch continuity

## Implementation Recommendations for Splice

Based on tig's proven approach:

### 1. Symbol Generation with Flags

Use bit flags similar to tig's `graph_symbol` structure:

```go
type CellFlags struct {
    Commit         bool
    ContinuedUp    bool
    ContinuedDown  bool
    ContinuedLeft  bool
    ContinuedRight bool
    ParentDown     bool
    Merge          bool
}
```

**Benefits:**
- Clearer logic for character selection
- Easier to test flag combinations
- Separates layout from rendering

### 2. Three-Row Lookahead

Implement sliding window (prev, current, next):

```go
type GraphComputer struct {
    prevRow    *Row
    currentRow *Row
    nextRow    *Row
    activeLanes []string
}
```

**Enables:**
- Proper connection character determination
- Detection of vertical continuity
- Branch divergence/convergence patterns
- Cross-over handling

### 3. Column Collapse

Implement active column removal:

```go
func (g *GraphComputer) collapseTrailingColumns() {
    // Remove trailing nil entries from activeLanes
    for i := len(g.activeLanes) - 1; i >= 0; i-- {
        if g.activeLanes[i] != "" {
            g.activeLanes = g.activeLanes[:i+1]
            break
        }
    }
}
```

Prevents unbounded graph width growth.

### 4. Color Load-Balancing

Rather than sequential assignment:

```go
type ColorTracker struct {
    assignments map[string]int // hash -> color
    counts      []int          // usage per color
}

func (c *ColorTracker) assignColor(hash string) int {
    // Find color with lowest usage count
    minCount, minColor := c.counts[0], 0
    for i, count := range c.counts {
        if count < minCount {
            minCount = count
            minColor = i
        }
    }
    c.assignments[hash] = minColor
    c.counts[minColor]++
    return minColor
}
```

**Result:** More visually distinct adjacent branches.

### 5. Forbidden Columns

Track columns where placement would cause overlap:

```go
func (g *GraphComputer) computeForbiddenColumns(commit *GitCommit) []bool {
    forbidden := make([]bool, len(g.activeLanes))

    // Mark columns where edge overlap would occur
    for i, lane := range g.activeLanes {
        if lane != "" && lane != commit.Hash {
            // Check if placing commit here would cross existing edge
            if wouldCauseOverlap(commit, i, g.activeLanes) {
                forbidden[i] = true
            }
        }
    }

    return forbidden
}
```

Prevents visual ambiguity in complex merge scenarios.

### 6. Testing Strategy

Comprehensive test cases for:
- **Linear history**: Simple chain, no branches
- **Single branch**: Feature branch merged to main
- **Multiple branches**: Parallel development
- **Octopus merge**: 3+ parents converging
- **Criss-cross merge**: Branches merge multiple times
- **Multiple roots**: Disconnected histories
- **Branch deletion**: Abandoned branches
- **Complex topology**: Real-world scenarios

Use golden file testing for visual output verification.

## Key Takeaways

1. **Flag-based approach** is superior to pattern matching for symbol generation
2. **Three-row lookahead** is essential for proper connection characters
3. **Column reuse** (with nil slots) prevents unbounded width
4. **Color load-balancing** significantly improves readability
5. **Forbidden columns** prevents visual ambiguity
6. **First parent convention** is critical for maintaining branch continuity

Tig's implementation represents 15+ years of refinement and real-world usage. Following its core principles while adapting to Go idioms will result in a robust, maintainable implementation.

## References

- [Tig Repository - graph.c](https://github.com/jonas/tig/blob/master/src/graph.c)
- [Tig Repository - graph-v1.c](https://github.com/jonas/tig/blob/master/src/graph-v1.c)
- [QGit lanes.cpp](https://github.com/tibirna/qgit/blob/master/src/lanes.cpp)
- [Commit Graph Drawing Algorithms - pvigier's blog](https://pvigier.github.io/2019/05/06/commit-graph-drawing-algorithms.html)
- [Drawing a Commit Graph - DoltHub Blog](https://www.dolthub.com/blog/2024-08-07-drawing-a-commit-graph/)
- [Git Graph Drawing Collection](https://github.com/indigane/git-graph-drawing)
