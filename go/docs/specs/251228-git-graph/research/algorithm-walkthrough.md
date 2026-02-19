# Git Graph Algorithm Walkthrough

A concrete example showing how the tig-inspired algorithm processes commits step-by-step to generate graph layout.

## Example: Feature Branch Merge

### Git History

```
* e123456 (HEAD -> main) Merge feature branch
|\
| * d234567 Feature complete
| * c345678 Feature work
* | b456789 Main work
|/
* a567890 Initial commit
```

**Commit relationships:**
```
e123456: parents = [b456789, d234567]  (merge commit - 2 parents)
d234567: parents = [c345678]
c345678: parents = [a567890]
b456789: parents = [a567890]
a567890: parents = []                   (root commit)
```

### Algorithm Execution

Processing commits in display order (newest to oldest): **e → d → c → b → a**

---

## Step 1: Process commit `e123456` (Merge)

**Input:**
- Hash: `e123456`
- Parents: `[b456789, d234567]` (first parent = b456789)
- activeLanes: `[]` (empty, first commit)

**Column Assignment:**
1. Search for `e123456` in activeLanes: **not found**
2. No existing columns, so append new column
3. Assign to column **0**

**Update Lanes:**
1. Place first parent `b456789` in column 0
2. Second parent `d234567` needs column → append to column 1
3. activeLanes: `[b456789, d234567]`

**Symbol Generation:**

| Column | Hash    | Flags | Char | Explanation |
|--------|---------|-------|------|-------------|
| 0      | current | Commit=true, ContinuedDown=true (b456789 below), ContinuedRight=true (merge from col 1) | `├─` | Commit marker with merge line to right |
| 1      | -       | ContinuedDown=true (d234567 below), ContinuedLeft=true (merging left) | `╮` | Branch merging into column 0 |

**Row state after step 1:**
```
prevRow:    nil
currentRow: [├─, ╮]  (e123456 at col 0)
nextRow:    computing...
activeLanes: [b456789, d234567]
```

**Visual:**
```
├─╮ e123456 (HEAD -> main) Merge feature
```

---

## Step 2: Process commit `d234567`

**Input:**
- Hash: `d234567`
- Parents: `[c345678]`
- activeLanes: `[b456789, d234567]`

**Column Assignment:**
1. Search for `d234567` in activeLanes: **found at column 1**
2. Assign to column **1**

**Update Lanes:**
1. Replace `d234567` with its first parent `c345678` at column 1
2. activeLanes: `[b456789, c345678]`

**Symbol Generation (with lookahead):**

Looking at three rows:
- prev: e123456 at col 0 (has merge edge to col 1)
- current: d234567 at col 1
- next: will have c345678 at col 1

| Column | Hash | Flags | Char | Explanation |
|--------|------|-------|------|-------------|
| 0      | -    | ContinuedUp=true (e at col 0), ContinuedDown=true (b below), ParentDown=true | `│` | Vertical continuation for main branch |
| 1      | current | Commit=true, ContinuedUp=true (merge edge from e), ContinuedDown=true (c below) | `├` | Commit marker on feature branch |

**Row state after step 2:**
```
prevRow:    [├─, ╮]    (e123456)
currentRow: [│, ├]     (d234567 at col 1)
nextRow:    computing...
activeLanes: [b456789, c345678]
```

**Visual so far:**
```
├─╮ e123456 (HEAD -> main) Merge feature
│ ├ d234567 Feature complete
```

---

## Step 3: Process commit `c345678`

**Input:**
- Hash: `c345678`
- Parents: `[a567890]`
- activeLanes: `[b456789, c345678]`

**Column Assignment:**
1. Search for `c345678` in activeLanes: **found at column 1**
2. Assign to column **1**

**Update Lanes:**
1. Replace `c345678` with its parent `a567890` at column 1
2. But `a567890` is also the parent of `b456789` (in column 0)
3. Both branches will converge at `a567890`
4. activeLanes: `[b456789, a567890]`

**Symbol Generation:**

Looking at three rows:
- prev: d234567 at col 1
- current: c345678 at col 1
- next: will need to show convergence (b at col 0, a at col 1)

| Column | Hash | Flags | Char | Explanation |
|--------|------|-------|------|-------------|
| 0      | -    | ContinuedUp=true, ContinuedDown=true | `│` | Main branch continues |
| 1      | current | Commit=true, ContinuedUp=true (d above), ContinuedDown=true (a below) | `├` | Commit marker |

**Row state after step 3:**
```
prevRow:    [│, ├]     (d234567)
currentRow: [│, ├]     (c345678 at col 1)
nextRow:    computing...
activeLanes: [b456789, a567890]
```

**Visual so far:**
```
├─╮ e123456 (HEAD -> main) Merge feature
│ ├ d234567 Feature complete
│ ├ c345678 Feature work
```

---

## Step 4: Process commit `b456789`

**Input:**
- Hash: `b456789`
- Parents: `[a567890]`
- activeLanes: `[b456789, a567890]`

**Column Assignment:**
1. Search for `b456789` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. Replace `b456789` with its parent `a567890` at column 0
2. Now we have `a567890` in BOTH columns 0 and 1
3. This is the convergence point - both branches share same parent
4. activeLanes: `[a567890, a567890]`

**Symbol Generation:**

Looking at three rows:
- prev: c345678 at col 1
- current: b456789 at col 0
- next: will have a567890 (both branches converge)

| Column | Hash | Flags | Char | Explanation |
|--------|------|-------|------|-------------|
| 0      | current | Commit=true, ContinuedUp=true (e's first parent), ContinuedDown=true (a below) | `├` | Commit on main branch |
| 1      | -    | ContinuedUp=true (c above), ContinuedDown=true (a below), ContinuedLeft=true (converging) | `│` | Feature branch continues down |

**Row state after step 4:**
```
prevRow:    [│, ├]     (c345678)
currentRow: [├, │]     (b456789 at col 0)
nextRow:    computing...
activeLanes: [a567890, a567890]
```

**Visual so far:**
```
├─╮ e123456 (HEAD -> main) Merge feature
│ ├ d234567 Feature complete
│ ├ c345678 Feature work
├ │ b456789 Main work
```

---

## Step 5: Process commit `a567890` (Root)

**Input:**
- Hash: `a567890`
- Parents: `[]` (root commit, no parents)
- activeLanes: `[a567890, a567890]`

**Column Assignment:**
1. Search for `a567890` in activeLanes: **found at column 0** (first occurrence)
2. Assign to column **0**

**Update Lanes:**
1. Root commit has no parents
2. After this commit, both branches end
3. Set both lanes to nil/empty
4. activeLanes: `[]` (or `[nil, nil]`)

**Column Collapse:**
- Detect trailing empty columns
- Remove column 1 since it's no longer needed
- activeLanes: `[]`

**Symbol Generation:**

Looking at three rows:
- prev: b456789 at col 0, vertical line at col 1
- current: a567890 at col 0 (convergence point)
- next: nil (end of history)

| Column | Hash | Flags | Char | Explanation |
|--------|------|-------|------|-------------|
| 0      | current | Commit=true, ContinuedUp=true (b above), ContinuedRight=true (branch from col 1 joins) | `├─` | Commit where branches converge |
| 1      | -    | ContinuedUp=true (feature branch), ContinuedLeft=true (joining to col 0) | `╯` | Corner showing branch end/join |

**Row state after step 5:**
```
prevRow:    [├, │]     (b456789)
currentRow: [├─, ╯]    (a567890 at col 0)
nextRow:    nil
activeLanes: []
```

**Final Visual:**
```
├─╮ e123456 (HEAD -> main) Merge feature
│ ├ d234567 Feature complete
│ ├ c345678 Feature work
├ │ b456789 Main work
├─╯ a567890 Initial commit
```

---

## Complete Algorithm State Table

| Step | Commit | Column | activeLanes before | activeLanes after | Row Output |
|------|--------|--------|--------------------|-------------------|------------|
| 1 | e123456 | 0 | `[]` | `[b456789, d234567]` | `[├─, ╮]` |
| 2 | d234567 | 1 | `[b456789, d234567]` | `[b456789, c345678]` | `[│, ├]` |
| 3 | c345678 | 1 | `[b456789, c345678]` | `[b456789, a567890]` | `[│, ├]` |
| 4 | b456789 | 0 | `[b456789, a567890]` | `[a567890, a567890]` | `[├, │]` |
| 5 | a567890 | 0 | `[a567890, a567890]` | `[]` | `[├─, ╯]` |

---

## Key Observations

### 1. First Parent Convention

**Commit e123456 (merge):**
- First parent: `b456789` (main branch) → stays in column 0
- Second parent: `d234567` (feature branch) → placed in column 1

This keeps the main branch in a consistent column throughout history.

### 2. Column Reuse

After step 5, both lanes become empty, but we don't remove column 1 until after rendering. This prevents mid-history shifts.

### 3. Three-Row Lookahead

At step 5 (commit a567890):
- **prevRow** shows `[├, │]` - b at col 0, vertical at col 1
- **currentRow** is a567890 at col 0
- By looking at prev, we know col 1 has a line coming down that needs to connect

This lookahead enables us to generate the `╯` character correctly.

### 4. Symbol Flag Examples

**For cell [0, 0] in step 2 (vertical line on main branch):**
```go
CellFlags{
    Commit:         false,
    ContinuedUp:    true,  // e123456 above
    ContinuedDown:  true,  // b456789 below
    ContinuedLeft:  false,
    ContinuedRight: false,
    ParentDown:     true,  // b456789 is parent of e123456
    Merge:          false,
}
// Flags → Character: │
```

**For cell [4, 1] in step 5 (corner joining left):**
```go
CellFlags{
    Commit:         false,
    ContinuedUp:    true,  // Feature branch from above
    ContinuedDown:  false, // Ends here
    ContinuedLeft:  true,  // Joins to column 0
    ContinuedRight: false,
    ParentDown:     false,
    Merge:          false,
}
// Flags → Character: ╯
```

### 5. Forbidden Columns

In this example, when processing c345678 (step 3):
- Column 0 is **forbidden** because `b456789` occupies it
- Placing c345678 in column 0 would cause overlap
- Algorithm correctly keeps it in column 1

---

## Implementation Pseudocode

Based on this walkthrough:

```go
func ComputeLayout(commits []GitCommit) *Layout {
    var activeLanes []string
    var rows []Row
    var prevRow, currentRow *Row

    for _, commit := range commits {
        // 1. Find or assign column
        col := findInLanes(commit.Hash, activeLanes)
        if col == -1 {
            col = findEmptyLane(activeLanes)
            if col == -1 {
                col = len(activeLanes)
                activeLanes = append(activeLanes, "")
            }
        }

        // 2. Build next row (lookahead)
        nextRow := buildNextRow(activeLanes)

        // 3. Generate symbols using three-row window
        currentRow = generateSymbols(prevRow, currentRow, nextRow, activeLanes, col)

        // 4. Update active lanes
        activeLanes[col] = ""
        if len(commit.ParentHashes) > 0 {
            activeLanes[col] = commit.ParentHashes[0] // First parent

            // Handle additional parents (merges)
            for i := 1; i < len(commit.ParentHashes); i++ {
                emptyCol := findEmptyLane(activeLanes)
                if emptyCol == -1 {
                    activeLanes = append(activeLanes, commit.ParentHashes[i])
                } else {
                    activeLanes[emptyCol] = commit.ParentHashes[i]
                }
            }
        }

        // 5. Collapse trailing empty lanes
        activeLanes = collapseTrailing(activeLanes)

        // 6. Save row
        rows = append(rows, *currentRow)

        // 7. Rotate window
        prevRow = currentRow
    }

    return &Layout{Rows: rows}
}
```

---

## Character Selection Logic

```go
func selectCharacter(flags CellFlags) rune {
    switch {
    case flags.Commit && flags.ContinuedUp && !flags.ContinuedRight:
        return '├'
    case flags.Commit && flags.ContinuedUp && flags.ContinuedRight:
        return '├' // Will be rendered as ├─ (with horizontal line)
    case flags.ContinuedUp && flags.ContinuedDown && !flags.ContinuedLeft && !flags.ContinuedRight:
        return '│'
    case flags.ContinuedDown && flags.ContinuedRight && !flags.ContinuedUp:
        return '╮'
    case flags.ContinuedUp && flags.ContinuedLeft && !flags.ContinuedDown:
        return '╯'
    case flags.ContinuedUp && flags.ContinuedDown && flags.ContinuedRight:
        return '├'
    case flags.ContinuedUp && flags.ContinuedDown && flags.ContinuedLeft:
        return '┤'
    case flags.ContinuedUp && flags.ContinuedDown && flags.ContinuedLeft && flags.ContinuedRight:
        return '┼'
    default:
        return ' '
    }
}
```

---

## Testing This Example

```go
func TestFeatureBranchMerge(t *testing.T) {
    commits := []GitCommit{
        {Hash: "e123456", ParentHashes: []string{"b456789", "d234567"}},
        {Hash: "d234567", ParentHashes: []string{"c345678"}},
        {Hash: "c345678", ParentHashes: []string{"a567890"}},
        {Hash: "b456789", ParentHashes: []string{"a567890"}},
        {Hash: "a567890", ParentHashes: []string{}},
    }

    layout := ComputeLayout(commits)

    expected := []string{
        "├─╮",
        "│ ├",
        "│ ├",
        "├ │",
        "├─╯",
    }

    for i, row := range layout.Rows {
        got := renderRow(row)
        if got != expected[i] {
            t.Errorf("Row %d: got %q, want %q", i, got, expected[i])
        }
    }
}
```

This walkthrough demonstrates how the algorithm maintains state, assigns columns, generates symbols, and produces the final visual output.
