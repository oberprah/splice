# Implementation Plan: Git Graph in Logs View

## Overview

This implementation follows a TDD approach, starting with the core algorithm and rendering before integrating with the rest of the application. The test cases from `research/test-cases.md` serve as our acceptance criteria.

## Strategy

**Core-first approach:** Implement the graph algorithm as a standalone, pure function that takes commits with parent hashes and outputs graph symbols. This allows thorough testing before any integration work.

**TDD with golden tests:** Use the 9 test cases from the design doc as table-driven tests. Each case validates the algorithm produces the expected graph output.

**Incremental complexity:** Start with the simplest case (linear history) and progressively add support for merges, octopus merges, and complex scenarios.

## Steps

- [ ] Step 1: Implement core graph layout algorithm with TDD
  - [x] 1a: Data structures and symbol rendering
  - [x] 1b: Column assignment
  - [ ] 1c: Lane updates
  - [ ] 1d: Symbol generation
  - [ ] 1e: Full algorithm integration
- [ ] Step 2: Extend GitCommit with parent hashes and update git parsing
- [ ] Step 3: Integrate graph rendering into LogState view
- [ ] Step 4: Add ref decorations (branch names, tags)
- [ ] Validation: Verify with running application

## Step Details

### Step 1: Implement core graph layout algorithm with TDD

Create `internal/graph/` package with the layout algorithm. Break into sub-steps by algorithm component for incremental validation.

**Input:** Slice of commits with parent hashes (use simple `Commit` struct, not full `GitCommit`)
**Output:** `Layout` struct with rows of symbols

#### Step 1a: Data structures and symbol rendering

Define the foundational types and symbol-to-string rendering.

**Scope:**
- `internal/graph/types.go`: `Commit`, `Layout`, `Row`, `GraphSymbol` enum
- `internal/graph/symbols.go`: `RenderRow()` function, symbol constants
- `internal/graph/symbols_test.go`: Test symbol rendering

**GraphSymbol enum** (from design doc 3.3.3):
```go
const (
    SymbolEmpty       // "  "
    SymbolBranchPass  // "│ "
    SymbolBranchCross // "│─"
    SymbolCommit      // "├ "
    SymbolMergeCommit // "├─"
    SymbolBranchTop   // "╮ "
    SymbolBranchBottom// "╯ "
    SymbolMergeJoin   // "┤ "
    SymbolOctopus     // "┬─"
    SymbolDiverge     // "┴─"
    SymbolMergeCross  // "┼─"
)
```

**Deliverable:** Types compile, symbols render to correct 2-char strings.

#### Step 1b: Column assignment

Implement finding where a commit belongs in the active lanes.

**Scope:**
- `findInLanes(hash, activeLanes)` - find existing position for hash
- `findEmptyLane(activeLanes)` - find first nil/empty slot
- `assignColumn(hash, activeLanes)` - orchestrate: find existing or allocate new

**Tests:**
- Hash found at specific index → returns that index
- Hash not found, empty slot exists → returns empty slot index
- Hash not found, no empty slots → appends new column

**Deliverable:** Column assignment works correctly for various lane states.

#### Step 1c: Lane updates

Implement updating active lanes after processing a commit.

**Scope:**
- `updateLanes(col, parents, activeLanes)` - replace with first parent, add merge parents
- `collapseTrailing(activeLanes)` - remove trailing empty/nil lanes

**Key behaviors:**
- First parent replaces commit's position (branch continuation)
- Additional parents (merges) find empty slots or append
- Trailing nil lanes removed to prevent unbounded width

**Tests:**
- Single parent: replaces position
- Merge commit: first parent in place, second parent added
- Octopus merge: multiple parents added
- Collapse removes trailing empty lanes only

**Deliverable:** Lane state correctly evolves through commit sequence.

#### Step 1d: Symbol generation

Implement determining which symbol to render for each cell.

**Scope:**
- Internal connectivity flags (up, down, left, right, isCommit, isMerge)
- `generateRowSymbols(commitCol, prevLanes, currentLanes, nextLanes)` - produce symbols for a row
- Map connectivity patterns to GraphSymbol values

**Key patterns:**
- Commit marker: `├` or `├─` (if merging right)
- Vertical pass: `│` (lane continues through)
- Branch top: `╮` (merge point from right)
- Branch bottom: `╯` (branch joins from above)
- Cross: `┼─` (merge line crosses vertical lane)

**Tests:**
- Linear commit → `├`
- Merge commit → `├─╮`
- Vertical continuation → `│`
- Branch convergence → `├─╯`

**Deliverable:** Correct symbols generated for various connectivity scenarios.

#### Step 1e: Full algorithm integration

Wire all components into `ComputeLayout()` and validate with test cases.

**Scope:**
- `internal/graph/layout.go`: `ComputeLayout(commits []Commit) *Layout`
- `internal/graph/layout_test.go`: All 9 test cases from `research/test-cases.md`

**Algorithm flow:**
1. Initialize empty `activeLanes`
2. For each commit in display order:
   a. Assign column (Step 1b)
   b. Generate symbols for row (Step 1d)
   c. Update lanes (Step 1c)
   d. Append row to layout
3. Return completed layout

**Test cases:**
1. Linear History
2. Simple Feature Branch Merge
3. Two Parallel Feature Branches (Octopus)
4. Sequential Merges
5. Sequential Merges with commits on main
6. Root Commit
7. Multiple Roots
8. With Tags and Remote Refs (graph only, no refs)
9. Complex Multi-Branch

**Deliverable:** All 9 test cases pass, algorithm handles edge cases correctly.

### Step 2: Extend GitCommit with parent hashes and update git parsing

Modify the git package to fetch and parse parent hash data.

**Scope:**
- Add `ParentHashes []string` to `GitCommit` struct
- Update `FetchCommits()` to include `%P` in format string
- Update `ParseGitLogOutput()` to parse parent hashes
- Update existing tests to include parent hashes in test data

**Git command change:**
```bash
# Current
git log --pretty=format:%H%x00%an%x00%ad%x00%s%x00%b%x1e

# New (add %P for parents)
git log --pretty=format:%H%x00%P%x00%an%x00%ad%x00%s%x00%b%x1e
```

**Note:** Ref decorations (`%d`) are deferred per design doc section 3.8 - we'll add refs in a follow-up enhancement.

### Step 3: Integrate graph rendering into LogState view

Wire up the graph layout to the logs view rendering.

**Scope:**
- Add `GraphLayout *graph.Layout` field to `LogState`
- Compute layout when commits are loaded (in `LoadingState` → `LogState` transition)
- Modify `formatCommitLine()` to prepend graph symbols
- Update golden files for log view tests

**View changes:**
- Graph symbols appear after selection indicator, before hash
- Each symbol is 2 characters wide (symbol + connector/space)
- Graph width is variable based on number of active branches

**Format change:**
```
# Before
> a4c3a8a Fix memory leak - John Doe (4 min ago)

# After
> ├─╮ a4c3a8a Merge feature - John Doe (4 min ago)
```

### Step 4: Add ref decorations (branch names, tags)

Add branch names, tags, and HEAD indicator to the commit display.

**Scope:**
- Add `Refs []RefInfo` to `GitCommit` struct (as per design doc section 3.1)
- Update git command to include `%d` for ref decorations
- Parse ref decorations into structured `RefInfo` data
- Render refs inline with commit information in `formatCommitLine()`
- Add test case 8 (Tags and Remote Refs) validation

**RefInfo struct:**
```go
type RefInfo struct {
    Name   string  // e.g., "main", "v1.0"
    Type   RefType // Branch, RemoteBranch, Tag
    IsHead bool    // true if current HEAD
}
```

**Git command change:**
```bash
# From Step 2
git log --pretty=format:%H%x00%P%x00%an%x00%ad%x00%s%x00%b%x1e

# Add %d for decorations
git log --pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e
```

**Display format:**
```
> ├ a4c3a8a (HEAD -> main, origin/main) Latest commit - Alice (1 min ago)
  ├ b5d4b9b (tag: v1.0) Release - Alice (1 week ago)
```

### Validation: Verify with running application

Manual testing against a real git repository with various branch topologies.

**Test scenarios:**
1. Linear history (no branches)
2. Simple feature branch and merge
3. Multiple parallel branches
4. Repository with octopus merges (if available)
5. Verify split view still works correctly

## Progress

### Step 1: Implement core graph layout algorithm with TDD
Status: 🔄 In progress

#### 1a: Data structures and symbol rendering
Status: ✅ Complete

**Files created:**
- `internal/graph/types.go`: Commit, Layout, Row, GraphSymbol types
- `internal/graph/symbols.go`: RenderRow() function, symbol string mapping
- `internal/graph/symbols_test.go`: Tests for symbol rendering

**Notes:**
- All 11 GraphSymbol values defined with 2-character string representations
- Tests verify each symbol renders correctly and is exactly 2 runes wide
- Unknown/invalid symbols default to empty ("  ")

#### 1b: Column assignment
Status: ✅ Complete

**Files created:**
- `internal/graph/lanes.go`: findInLanes, findEmptyLane, assignColumn, collapseTrailingEmpty
- `internal/graph/lanes_test.go`: Tests for all lane functions

**Notes:**
- `assignColumn` prioritizes: existing lane → empty slot → new column
- `collapseTrailingEmpty` prevents unbounded width growth
- Empty strings represent cleared lanes (branch completed)

#### 1c: Lane updates
Status: Pending

#### 1d: Symbol generation
Status: Pending

#### 1e: Full algorithm integration
Status: Pending

### Step 2: Extend GitCommit with parent hashes
Status: Pending

### Step 3: Integrate graph rendering into LogState view
Status: Pending

### Step 4: Add ref decorations (branch names, tags)
Status: Pending

### Validation
Status: Pending

## Discoveries

(To be filled during implementation)

## Verification

- [ ] All tests pass (`go test ./...`)
- [ ] All 9 test cases from design doc produce correct output
- [ ] Log view renders graphs in simple view mode
- [ ] Log view renders graphs in split view mode
- [ ] No performance regression for large repositories
