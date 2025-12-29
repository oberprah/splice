# Design: Git Graph in Logs View

## 0 - Executive Summary

This design adds a visual git graph to Splice's logs view by extending the existing commit data model with parent hash information and rendering a column-based graph alongside commit lines. The graph uses Unicode box-drawing characters to visualize branch topology.

The implementation extends the current `GitCommit` struct with parent hashes and ref decorations, adds a graph layout computation layer, and modifies the commit line renderer to prepend graph characters.

Key decisions: (1) Fetch topology data by extending the existing `git log` command rather than making separate calls; (2) Compute graph layout synchronously during commit load rather than asynchronously; (3) Use the "active lanes" column assignment algorithm that keeps commits on the same branch in the same column.

## 1 - Context & Problem Statement

When viewing git history in Splice's logs view, users see a flat list of commits without any visual indication of branch relationships or merge history. This makes it difficult to understand how branches were merged, see parallel development history, or identify which branch a commit belongs to.

This design covers adding a visual git graph to the logs view. It does not cover graph toggling, ASCII fallback for non-Unicode terminals, or interactive graph manipulation.

## 2 - Current State

### 2.1 - Commit Display

Each commit line currently shows:
```
[Selector 2ch] [Hash 7ch] [Message var] - [Author var] ([Time var])
```

Example:
```
> a4c3a8a Fix memory leak - John Doe (4 min ago)
```

### 2.2 - Data Model

```go
type GitCommit struct {
    Hash    string    // Full 40-character hash
    Message string    // Subject line
    Body    string    // Commit body
    Author  string    // Author name
    Date    time.Time // Commit timestamp
}
```

### 2.3 - Git Command

```bash
git log --pretty=format:%H%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

### 2.4 - Layout

- Simple view (< 160 chars): Full terminal width for commit list
- Split view (≥ 160 chars): 77 chars for log panel, 80 chars for details panel

## 3 - Solution

### 3.1 - Data Model Changes

> **Decision:** Extend `GitCommit` with topology input data (parent hashes, refs) rather than maintaining a separate mapping structure. The `GraphLayout` is a computed rendering output derived from this data, not a duplication of commit information.
>
> Alternative considered: Keep `GitCommit` unchanged and maintain a separate `CommitTopology map[string]TopologyInfo` that maps hash → parents/refs. Rejected because it requires keeping two parallel structures in sync and complicates data passing.

```go
type GitCommit struct {
    Hash         string      // Existing
    Message      string      // Existing
    Body         string      // Existing
    Author       string      // Existing
    Date         time.Time   // Existing
    ParentHashes []string    // NEW: Parent commit hashes
    Refs         []RefInfo   // NEW: Branch/tag decorations
}

type RefInfo struct {
    Name   string  // e.g., "main", "v1.0"
    Type   RefType // Branch, RemoteBranch, Tag
    IsHead bool    // true if current HEAD
}

type RefType int
const (
    RefTypeBranch RefType = iota
    RefTypeRemoteBranch
    RefTypeTag
)
```

### 3.2 - Git Command Changes

> **Decision:** Extend the existing `git log` format string rather than making separate git calls. This avoids additional process spawning and keeps commit data coherent.

Extended format:
```bash
git log --pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e --date=iso-strict -n <limit>
```

New placeholders:
- `%P` - Parent hashes (space-separated)
- `%d` - Ref decorations (e.g., " (HEAD -> main, tag: v1.0)")

### 3.3 - Graph Layout Computation

> **Decision:** Calculate graph layout ourselves from parent data rather than using git's `--graph` output.
>
> Git provides `git log --graph` but it outputs ASCII art mixed with commit data. This doesn't work for us because:
> - We need Unicode characters, not ASCII
> - We need to integrate with our existing commit line format
> - Parsing git's mixed output would be fragile
>
> Computing from parent hashes gives us full control and is the standard approach used by lazygit, tig, and gitk.

The algorithm is the "active lanes" approach that processes commits in display order and maintains a list of active branch columns. This design is based on research into tig's graph-v2 implementation and industry-standard algorithms.

```
Graph Layout Data Flow:

[]GitCommit (with parents)
        ↓
    ComputeLayout()
        ↓
    GraphLayout {
        Rows: []Row  // One per commit
    }
        ↓
    Used during View() rendering
```

#### 3.3.1 - Algorithm Overview

Based on tig's approach and standard git graph algorithms, the computation follows these steps:

1. **Initialize active lanes**: `activeLanes []string` tracks which commit hash occupies each column
2. **Process commits in display order** (most recent first)
3. **For each commit:**
   - **Locate position**: Search for commit hash in activeLanes
     - If found: use that column (branch continuation)
     - If not found: find first empty/nil slot or append new column
   - **Update lanes**: Replace commit's position with its first parent hash
   - **Handle merge parents**: Insert additional parent hashes into activeLanes (or append if no space)
   - **Compute symbols**: Determine Unicode characters based on column relationships
   - **Collapse width**: Remove trailing empty columns to prevent unbounded growth

#### 3.3.2 - Key Insights

**First Parent Convention**: The first parent represents the branch lineage (branch being merged INTO), subsequent parents are branches being merged IN. This keeps commits on the same branch in the same column.

**Column Reuse**: Rather than always appending new columns, search for empty/nil slots first. Set lanes to `nil` when no longer active rather than removing them (which would shift positions and break vertical alignment).

**Lookahead for Symbol Selection**: Tig uses a three-row sliding window (prev, current, next) to determine proper connection characters. By examining adjacent rows, the algorithm can detect:
- Vertical continuity (│): same hash in adjacent rows
- Branch divergence (╮, ╯): parent appears in different column than child
- Merge convergence (├─): multiple parents converging to one commit
- Cross-overs (┼): lines passing through without connecting

**Forbidden Columns**: When placing merge commits, track columns where placement would cause edge overlap. This prevents visual ambiguity where multiple edges cross inappropriately.

#### 3.3.3 - Symbol Generation

The algorithm internally tracks directional flags (up, down, left, right, commit) during computation to determine which symbol to output. These flags are an implementation detail and not exposed in the public API.

The output uses a `GraphSymbol` enum. Each symbol renders as exactly 2 characters (symbol + connector or space):

| Symbol | Renders | Description |
|--------|---------|-------------|
| `SymbolEmpty` | `"  "` | Empty cell |
| `SymbolBranchPass` | `"│ "` | Branch line passes through |
| `SymbolBranchCross` | `"│─"` | Branch line crossed by merge line |
| `SymbolCommit` | `"├ "` | Commit node |
| `SymbolMergeCommit` | `"├─"` | Commit starting a merge line |
| `SymbolBranchTop` | `"╮ "` | Top of a feature branch (merge point) |
| `SymbolBranchBottom` | `"╯ "` | Bottom of a feature branch (common ancestor) |
| `SymbolMergeJoin` | `"┤ "` | Merge line joins into existing branch |
| `SymbolOctopus` | `"┬─"` | Octopus merge - new branch goes down |
| `SymbolDiverge` | `"┴─"` | Branches diverge - branch goes up |
| `SymbolMergeCross` | `"┼─"` | Merge line crosses a branch |

### 3.4 - Graph State

> **Decision:** Store computed graph layout in `LogState` as a separate `GraphLayout` structure, indexed parallel to `Commits`.
>
> **Alternative considered:** Store graph rendering data directly in `GitCommit`:
> ```go
> type GitCommit struct {
>     // ... existing fields ...
>     GraphColumn int    // Which column this commit renders in
>     GraphCells  []Cell // Pre-rendered cells for this row
> }
> ```
>
> **Inside GitCommit:**
> - Pros: Single structure, simpler to pass around; each commit "knows" how to render
> - Cons: Mixes git data with UI concerns; column assignment depends on visible commits; can't recompute without mutating; harder to test in isolation
>
> **Separate GraphLayout:**
> - Pros: Clean separation of concerns; testable independently; layout recomputable without touching commits; GitCommit reusable across views
> - Cons: Two parallel structures to coordinate; slightly more complex rendering code
>
> Chose separate `GraphLayout` because it follows existing Splice patterns (state structs own computed data) and keeps git data pure.

```go
type LogState struct {
    Commits       []git.GitCommit  // Existing
    Cursor        int              // Existing
    ViewportStart int              // Existing
    Preview       PreviewState     // Existing
    GraphLayout   *graph.Layout    // NEW
}

// In graph package
type Layout struct {
    Rows []Row
}

type Row struct {
    Symbols []GraphSymbol
}
```

### 3.5 - Rendering Changes

**Modified commit line format:**
```
[Selector 2ch] [Graph var] [Hash 7ch] [Refs var] [Message var] - [Author var] ([Time var])
```

**Example (feature branch merge):**
```
> ├─╮ e123456 (HEAD -> main) Merge feature - Alice (1 min ago)
  │ ├ d234567 Feature complete - Bob (2 min ago)
  │ ├ c345678 Feature work - Bob (3 min ago)
  ├ │ b456789 Main work - Alice (4 min ago)
  ├─╯ a567890 Initial commit - Alice (1 day ago)
```

Every commit has a `├` marker showing which column/branch it belongs to:
- Column 0 (left): main branch - commits e, b, a
- Column 1 (right): feature branch - commits d, c
- The `─╮` on the merge commit (e) shows where the feature branch merges in
- The `─╯` on commit a shows where the branches converge (their common ancestor)

Note: When reading history top-to-bottom (newest first), merges appear before their branch points. The `╯` indicates where parallel branches originated from the same parent.

See [research/test-cases.md](research/test-cases.md) for comprehensive test scenarios including linear history, octopus merges, multiple roots, and ref decorations.

**Note:** The test cases show expected visual output, but the exact graph rendering may differ during implementation as we refine the algorithm. The test cases serve as a guide for the types of scenarios to handle, not necessarily pixel-perfect expected output.

**Width allocation:**

| Element | Width |
|---------|-------|
| Selector | 2ch |
| Graph | Variable (2ch per column) |
| Hash | 7ch |
| Refs | Variable, truncated as needed |
| Message/Author/Time | Remaining space |

**Ref decorations** use git's `%d` output directly:
- `(HEAD -> main)` - HEAD pointing to branch
- `(main)` - local branch
- `(origin/main)` - remote branch
- `(tag: v1.0)` - tag

### 3.6 - Unicode Characters

> **Decision:** Use `├` (box-drawing) for commit markers.
>
> Options considered:
> - `├` (box-drawing) - Integrates seamlessly with graph lines, clearly shows branch point
> - `*` (asterisk) - Universal compatibility, used by git itself, but visually disconnected from graph lines
> - `●` (U+25CF Black Circle) - Prettier but may not render in all terminal fonts
>
> Chose `├` because it visually connects the commit to its branch line, making the graph flow more coherent.

All characters are from the Unicode Box Drawing block (U+2500–U+257F). See section 3.3.3 for the complete symbol reference.

### 3.7 - Component Interaction

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                   git command                                   │
│  git log --pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e -n <lim>  │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              internal/git/git.go                                │
│            FetchCommits() - parse commits with parents + refs                   │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │ []GitCommit
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            internal/graph/layout.go                             │
│               ComputeLayout() - column assignment algorithm                     │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │ *Layout
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          internal/ui/states/log_state.go                        │
│                         LogState{Commits, GraphLayout}                          │
└────────────────────────────────────────┬────────────────────────────────────────┘
                                         │
                                         ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          internal/ui/states/log_view.go                         │
│                  formatCommitLine() - render graph + commit info                │
└─────────────────────────────────────────────────────────────────────────────────┘
```

### 3.8 - Deferred for Simplicity

The following features from the requirements are intentionally deferred to keep the initial implementation simple:

- **Maximum graph width handling**: The column collapse mechanism should naturally keep width bounded for most repositories. If width becomes problematic in practice, we can add truncation or scrolling later.

- **Color coding for branches**: The initial implementation will render the graph in a single color. Color assignment with load-balancing (as described in the tig research) can be added as a follow-up enhancement.

## 4 - Open Questions

None. The design is ready for implementation.

## 5 - References

- [Test Cases](research/test-cases.md) - Comprehensive graph rendering test scenarios
- [Tig Algorithm Analysis](research/tig-algorithm-analysis.md) - Research into tig's graph-v2 implementation
- [Algorithm Walkthrough](research/algorithm-walkthrough.md) - Step-by-step example of simple feature branch merge
- [Algorithm Walkthrough (Complex)](research/algorithm-walkthrough-complex.md) - Step-by-step example of complex multi-branch scenario
