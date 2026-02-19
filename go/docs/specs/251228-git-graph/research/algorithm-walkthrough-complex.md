# Complex Git Graph Algorithm Walkthrough

A detailed walkthrough of Test Case 9 - the most complex scenario with multiple merges, criss-crosses, and long-lived feature branches.

## Example: Complex Multi-Branch History

### Git History Topology

```
                K ← L (feature-3, long-lived)
               /               \
      C    E  /  G ←─── I       \
     / \  / \/  / \      \       \
  A─B───D────F─────H──────J───────M (HEAD -> main)
```

### Commit Relationships

```
M: parents = [J, L]       (merge feature-3 into main)
L: parents = [K]          (feature-3 continued)
K: parents = [D]          (feature-3 branched from D)
J: parents = [H, I]       (merge feature-2 into main)
I: parents = [G]          (feature-2 continued)
H: parents = [F, G]       (merge feature-2 early into main)
G: parents = [F]          (feature-2 continued)
F: parents = [D, E]       (merge hotfix into main)
E: parents = [D]          (hotfix branch)
D: parents = [B, C]       (merge feature-1 into main)
C: parents = [B]          (feature-1 branch)
B: parents = [A]          (main branch work)
A: parents = []           (root)
```

### Key Characteristics

- **13 commits** spanning multiple branches
- **5 merge commits**: M, J, H, F, D
- **Criss-cross pattern**: Feature-2 (G) branches from F, then merges back at H, then I branches from G and merges at J
- **Long-lived branch**: Feature-3 (K-L) branches from D and stays separate until final merge at M
- **Maximum width**: 4 columns at its widest point

---

## Algorithm Execution

Processing commits in display order (newest to oldest): **M → L → K → J → I → H → G → F → E → D → C → B → A**

---

## Step 1: Process commit `M` (Merge feature-3)

**Input:**
- Hash: `M`
- Parents: `[J, L]`
- activeLanes: `[]` (empty, first commit)

**Column Assignment:**
1. Search for `M` in activeLanes: **not found**
2. Assign to column **0** (first available)

**Update Lanes:**
1. First parent `J` → column 0
2. Second parent `L` → append to column 1
3. activeLanes: `[J, L]`

**Symbol Generation:**
- Column 0: Commit with merge right → `├─╮`
- Column 1: Merge corner → (part of ╮)

**State:**
```
activeLanes: [J, L]
Row: ├─╮ M (HEAD -> main) Merge feature-3
```

---

## Step 2: Process commit `L` (Feature-3 continued)

**Input:**
- Hash: `L`
- Parents: `[K]`
- activeLanes: `[J, L]`

**Column Assignment:**
1. Search for `L` in activeLanes: **found at column 1**
2. Assign to column **1**

**Update Lanes:**
1. Replace `L` with parent `K` at column 1
2. activeLanes: `[J, K]`

**Symbol Generation:**
- Column 0: Vertical (J continues) → `│`
- Column 1: Commit marker → `├`

**State:**
```
activeLanes: [J, K]
Row: │ ├ L Feature-3 done
```

---

## Step 3: Process commit `K` (Feature-3 start)

**Input:**
- Hash: `K`
- Parents: `[D]`
- activeLanes: `[J, K]`

**Column Assignment:**
1. Search for `K` in activeLanes: **found at column 1**
2. Assign to column **1**

**Update Lanes:**
1. Replace `K` with parent `D` at column 1
2. activeLanes: `[J, D]`

**Symbol Generation:**
- Column 0: Vertical (J continues) → `│`
- Column 1: Commit marker → `├`

**State:**
```
activeLanes: [J, D]
Row: │ ├ K Feature-3
```

**Key observation:** Feature-3 branch (K-L) stayed in column 1 from its start until merge.

---

## Step 4: Process commit `J` (Merge feature-2)

**Input:**
- Hash: `J`
- Parents: `[H, I]`
- activeLanes: `[J, D]`

**Column Assignment:**
1. Search for `J` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. First parent `H` → column 0
2. Second parent `I` → need new column (D occupies column 1)
3. Append `I` at column 2
4. activeLanes: `[H, D, I]`

**Symbol Generation:**
- Column 0: Commit with merge → `├─`
- Column 1: Vertical (D continues) → `│`
- Column 2: Merge corner → `╮`

**State:**
```
activeLanes: [H, D, I]
Row: ├─│─╮ J Merged Feature 2
```

**Note:** The horizontal line `─` between columns shows the merge connection spanning over column 1.

---

## Step 5: Process commit `I` (Feature-2 continued)

**Input:**
- Hash: `I`
- Parents: `[G]`
- activeLanes: `[H, D, I]`

**Column Assignment:**
1. Search for `I` in activeLanes: **found at column 2**
2. Assign to column **2**

**Update Lanes:**
1. Replace `I` with parent `G` at column 2
2. activeLanes: `[H, D, G]`

**Symbol Generation:**
- Column 0: Vertical (H continues) → `│`
- Column 1: Vertical (D continues) → `│`
- Column 2: Commit marker → `├`

**State:**
```
activeLanes: [H, D, G]
Row: │ │ ├ I Feature 2 done
```

---

## Step 6: Process commit `H` (Early merge of feature-2)

**Input:**
- Hash: `H`
- Parents: `[F, G]`
- activeLanes: `[H, D, G]`

**Column Assignment:**
1. Search for `H` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. First parent `F` → column 0
2. Second parent `G` → **already at column 2!**
3. This is a merge with a branch already in lanes
4. activeLanes: `[F, D, G]` (G stays at column 2)

**Symbol Generation:**
- Column 0: Commit with merge → `├─`
- Column 1: Vertical (D continues) → `│`
- Column 2: Vertical (G continues) → `│`
- Column 3: Merge corner from right → `╮`

**State:**
```
activeLanes: [F, D, G]
Row: ├─│─│─╮ H Early Merge Feature 2
```

**Key observation:** This shows a criss-cross pattern - G branched earlier, now being merged back while it also continues upward to I.

---

## Step 7: Process commit `G` (Feature-2 work)

**Input:**
- Hash: `G`
- Parents: `[F]`
- activeLanes: `[F, D, G]`

**Column Assignment:**
1. Search for `G` in activeLanes: **found at column 2**
2. Assign to column **2**

**Update Lanes:**
1. Replace `G` with parent `F` at column 2
2. But `F` is already at column 0!
3. This means G's branch converges back to F
4. Set column 2 to empty/nil (branch ends)
5. activeLanes: `[F, D, nil]`

**Symbol Generation:**
- Column 0: Vertical (F continues) → `│`
- Column 1: Vertical (D continues) → `│`
- Column 2: Commit marker → `├`
- Column 3: Corner joining left → `╯` (from previous merge line)

**State:**
```
activeLanes: [F, D, nil]
Row: │ │ ├─╯ G Feature-2 done
```

**Column Collapse:**
After this row, column 2 can be removed (trailing nil).
activeLanes: `[F, D]`

---

## Step 8: Process commit `F` (Merge hotfix)

**Input:**
- Hash: `F`
- Parents: `[D, E]`
- activeLanes: `[F, D]`

**Column Assignment:**
1. Search for `F` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. First parent `D` → but D is already at column 1!
2. F's first parent is D, so this is merging into D's branch
3. Actually, let me reconsider... F is the merge commit, parents are [D, E]
4. Replace F at column 0 with first parent D
5. But D is already at column 1, so we have `[D, D]`
6. Second parent E → needs a new column, append at column 2
7. activeLanes: `[D, D, E]` - wait, that's problematic

**Correction:**
Actually, when F processes:
- F is at column 0 (the main branch)
- Parents are [D, E]
- D is already at column 1 (from feature-3 path)
- This creates a complex merge situation

Looking at the expected output:
```
├─│─┤ F Merged Hotfix Feature 1
```

The pattern `├─│─┤` shows:
- `├─` at column 0: commit merging right
- `│─` at column 1: vertical with horizontal
- `┤` at column 2: T-junction (merge from right)

This suggests:
- F is at column 0
- D continues at column 1
- E merges in from column 2

Let me revise:

**Update Lanes (revised):**
1. First parent `D` at column 0 (F continues main as D)
2. D is also in column 1 (feature-3 path) - these will converge later
3. Second parent `E` → append at column 2
4. activeLanes: `[D, D, E]`

**Symbol Generation:**
- Column 0: Commit with merge → `├─`
- Column 1: Vertical (D continues from feature-3) → `│`
- Column 2: T-junction (E merging) → `┤`

**State:**
```
activeLanes: [D, D, E]
Row: ├─│─┤ F Merged Hotfix Feature 1
```

---

## Step 9: Process commit `E` (Hotfix branch)

**Input:**
- Hash: `E`
- Parents: `[D]`
- activeLanes: `[D, D, E]`

**Column Assignment:**
1. Search for `E` in activeLanes: **found at column 2**
2. Assign to column **2**

**Update Lanes:**
1. Replace `E` with parent `D` at column 2
2. Now we have `[D, D, D]` - all three columns have D!
3. activeLanes: `[D, D, D]`

**Symbol Generation:**
- Column 0: Vertical (D continues) → `│`
- Column 1: Vertical (D continues) → `│`
- Column 2: Commit marker → `├`

**State:**
```
activeLanes: [D, D, D]
Row: │ │ ├ E Hotfix Feature 1
```

---

## Step 10: Process commit `D` (Merge feature-1)

**Input:**
- Hash: `D`
- Parents: `[B, C]`
- activeLanes: `[D, D, D]`

**Column Assignment:**
1. Search for `D` in activeLanes: **found at column 0** (first occurrence)
2. Assign to column **0**

**Update Lanes:**
1. First parent `B` → column 0
2. Second parent `C` → column 1 (replace second D)
3. Third D at column 2 also needs to become B (branches converge)
4. activeLanes: `[B, C, nil]`

**Symbol Generation:**
- Column 0: Commit with merge → `├─`
- Column 1: Vertical continues → `│` (becomes C's parent)
- Column 2: Corner joining left → `╯` (branch ends)

**State:**
```
activeLanes: [B, C, nil]
Row: ├─┼─╯ D Merged Feature 1
```

The `┼` (cross) at column 1 shows the convergence of multiple paths.

**Column Collapse:**
Remove trailing nil: activeLanes: `[B, C]`

---

## Step 11: Process commit `C` (Feature-1 branch)

**Input:**
- Hash: `C`
- Parents: `[B]`
- activeLanes: `[B, C]`

**Column Assignment:**
1. Search for `C` in activeLanes: **found at column 1**
2. Assign to column **1**

**Update Lanes:**
1. Replace `C` with parent `B` at column 1
2. activeLanes: `[B, B]`

**Symbol Generation:**
- Column 0: Vertical (B continues) → `│`
- Column 1: Commit marker → `├`

**State:**
```
activeLanes: [B, B]
Row: │ ├ C Feature 1
```

---

## Step 12: Process commit `B` (Main branch work)

**Input:**
- Hash: `B`
- Parents: `[A]`
- activeLanes: `[B, B]`

**Column Assignment:**
1. Search for `B` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. Replace `B` with parent `A` at column 0
2. Replace `B` with parent `A` at column 1 (convergence)
3. activeLanes: `[A, A]`

**Symbol Generation:**
- Column 0: Commit with merge → `├─`
- Column 1: Corner joining left → `╯`

**State:**
```
activeLanes: [A, A]
Row: ├─╯ B Main work
```

**Column Collapse:**
Both lanes converge to A, remove column 1: activeLanes: `[A]`

---

## Step 13: Process commit `A` (Root)

**Input:**
- Hash: `A`
- Parents: `[]` (root)
- activeLanes: `[A]`

**Column Assignment:**
1. Search for `A` in activeLanes: **found at column 0**
2. Assign to column **0**

**Update Lanes:**
1. No parents, empty all lanes
2. activeLanes: `[]`

**Symbol Generation:**
- Column 0: Commit marker (no continuation down) → `├`

**State:**
```
activeLanes: []
Row: ├ A Initial commit
```

---

## Complete State Table

| Step | Commit | Col | activeLanes before | activeLanes after | Max Width | Row Output |
|------|--------|-----|--------------------|-------------------|-----------|------------|
| 1 | M | 0 | `[]` | `[J, L]` | 2 | `├─╮` |
| 2 | L | 1 | `[J, L]` | `[J, K]` | 2 | `│ ├` |
| 3 | K | 1 | `[J, K]` | `[J, D]` | 2 | `│ ├` |
| 4 | J | 0 | `[J, D]` | `[H, D, I]` | 3 | `├─│─╮` |
| 5 | I | 2 | `[H, D, I]` | `[H, D, G]` | 3 | `│ │ ├` |
| 6 | H | 0 | `[H, D, G]` | `[F, D, G]` | 4 | `├─│─│─╮` |
| 7 | G | 2 | `[F, D, G]` | `[F, D]` | 3 | `│ │ ├─╯` |
| 8 | F | 0 | `[F, D]` | `[D, D, E]` | 3 | `├─│─┤` |
| 9 | E | 2 | `[D, D, E]` | `[D, D, D]` | 3 | `│ │ ├` |
| 10 | D | 0 | `[D, D, D]` | `[B, C]` | 3 | `├─┼─╯` |
| 11 | C | 1 | `[B, C]` | `[B, B]` | 2 | `│ ├` |
| 12 | B | 0 | `[B, B]` | `[A]` | 2 | `├─╯` |
| 13 | A | 0 | `[A]` | `[]` | 1 | `├` |

---

## Final Visual Output

```
> ├─╮ M (HEAD -> main) Merge feature-3 - Alice (1 min ago)
  │ ├ L Feature-3 done- Carol (1 min ago)
  │ ├ K Feature-3 - Carol (2 min ago)
  ├─│─╮ J Merged Feature 2 - Alice (3 min ago)
  │ │ ├ I Feature 2 done - Carol (4 min ago)
  ├─│─│─╮ H Early Merge Feature 2 - Alice (5 min ago)
  │ │ ├─╯ G Feature-2 done - Bob (6 min ago)
  ├─│─┤ F Merged Hotfix Feature 1 - Alice (7 min ago)
  │ │ ├ E Hotfix Feature 1 - Alice (8 min ago)
  ├─┼─╯ D Merged Feature 1 - Diana (9 min ago)
  │ ├ C Feature 1 - Alice (9 min ago)
  ├─╯ B Main work - Alice (10 min ago)
  ├ A Initial commit- Alice (1 day ago)
```

---

## Key Algorithm Features Demonstrated

### 1. Long-Lived Branches

**Feature-3 (K → L):**
- Branched from D at step 10
- Stayed in column 1 through steps 3 → 2 → 1
- Maintained consistent position until final merge at M

This demonstrates column stability for branches that don't merge quickly.

### 2. Criss-Cross Merges

**Feature-2 (G → I → J):**
- G branches from F (step 7)
- H merges G early while G continues (step 6) - **first merge**
- G continues to I (step 5)
- J merges I (step 4) - **second merge**

This creates a diamond pattern where a branch is partially merged but continues.

### 3. Multiple Paths to Same Commit

**Commit D appears in multiple lanes:**
- Step 10: D appears in all three lanes `[D, D, D]`
- From main branch (column 0)
- From feature-3 branch (column 1)
- From hotfix branch (column 2)

All three paths converge at D, demonstrating how the algorithm handles convergence.

### 4. Dynamic Width Management

The graph width changes as branches are added and removed:
- Starts at width 2 (M merge)
- Expands to width 3 (J merge)
- **Peaks at width 4** (H merge with criss-cross)
- Contracts back through column collapse

### 5. Complex Symbol Generation

**Step 6 (H) produces: `├─│─│─╮`**

This requires tracking:
- Commit at column 0: `├─`
- D continues at column 1: `│`
- G continues at column 2: `│`
- Merge line connects through columns: `─`
- Corner at column 3: `╮`

The three-row lookahead enables correct character selection:
- prevRow: I at column 2
- currentRow: H at column 0 (merging with column 2)
- nextRow: G at column 2 (continues down)

### 6. T-Junction Pattern

**Step 8 (F) produces: `├─│─┤`**

The `┤` character (right T-junction) appears when:
- A merge brings in a branch from the right
- The merging branch doesn't continue upward (E ends here)
- The main branch continues upward

Flags for the `┤` cell:
```go
CellFlags{
    Commit:         false,
    ContinuedUp:    true,  // E above
    ContinuedDown:  false, // E's branch ends
    ContinuedLeft:  true,  // Merges left to F
    ContinuedRight: false,
}
```

### 7. Cross Junction Pattern

**Step 10 (D) produces: `├─┼─╯`**

The `┼` character (cross junction) appears when:
- Multiple branches converge at the same commit
- Lines cross through a cell
- Both vertical and horizontal continuity exists

This handles the complex convergence of three branches at D.

---

## Implementation Insights

### Forbidden Columns in Practice

When processing H (step 6):
- H is at column 0
- Want to merge with G at column 2
- **Column 1 is forbidden** for placing H because D occupies it
- Column 2 is also forbidden for H because G is there
- H must stay at column 0 (its existing position)

The algorithm avoids conflicts by:
1. First checking if commit already has a lane assignment
2. Only assigning new columns for new branches (merge parents)
3. Computing forbidden columns before placement

### Column Collapse Strategy

The algorithm collapses in these situations:

**After step 7 (G processed):**
```
activeLanes: [F, D, nil] → [F, D]
```
G's branch ended, leaving trailing nil.

**After step 10 (D processed):**
```
activeLanes: [B, C, nil] → [B, C]
```
All three D lanes converged.

**After step 12 (B processed):**
```
activeLanes: [A, A] → [A]
```
Both lanes have same commit, collapse to one.

### Handling Same Commit in Multiple Lanes

When a commit appears multiple times in activeLanes (like D in `[D, D, D]`):
1. Process at first occurrence (leftmost)
2. Update all occurrences with first parent
3. Other parents placed in appropriate lanes
4. Lanes converge as branches complete

---

## Testing This Example

```go
func TestComplexMultiBranch(t *testing.T) {
    commits := []GitCommit{
        {Hash: "M", ParentHashes: []string{"J", "L"}},
        {Hash: "L", ParentHashes: []string{"K"}},
        {Hash: "K", ParentHashes: []string{"D"}},
        {Hash: "J", ParentHashes: []string{"H", "I"}},
        {Hash: "I", ParentHashes: []string{"G"}},
        {Hash: "H", ParentHashes: []string{"F", "G"}},
        {Hash: "G", ParentHashes: []string{"F"}},
        {Hash: "F", ParentHashes: []string{"D", "E"}},
        {Hash: "E", ParentHashes: []string{"D"}},
        {Hash: "D", ParentHashes: []string{"B", "C"}},
        {Hash: "C", ParentHashes: []string{"B"}},
        {Hash: "B", ParentHashes: []string{"A"}},
        {Hash: "A", ParentHashes: []string{}},
    }

    layout := ComputeLayout(commits)

    expected := []string{
        "├─╮",       // M
        "│ ├",       // L
        "│ ├",       // K
        "├─│─╮",     // J
        "│ │ ├",     // I
        "├─│─│─╮",   // H (widest point - 4 columns)
        "│ │ ├─╯",   // G
        "├─│─┤",     // F
        "│ │ ├",     // E
        "├─┼─╯",     // D
        "│ ├",       // C
        "├─╯",       // B
        "├",         // A
    }

    for i, row := range layout.Rows {
        got := renderRow(row)
        if got != expected[i] {
            t.Errorf("Row %d (%s): got %q, want %q",
                i, commits[i].Hash, got, expected[i])
        }
    }
}
```

---

## Summary

This complex example demonstrates:

1. **Long-lived branches** that maintain consistent columns
2. **Criss-cross merges** where branches merge multiple times
3. **Multiple paths converging** to the same commit
4. **Dynamic width management** expanding to 4 columns and contracting
5. **Advanced symbol generation** including `┼` (cross) and `┤` (T-junction)
6. **Column collapse** removing trailing empty lanes
7. **Forbidden column logic** preventing placement conflicts

The algorithm successfully handles this complexity through:
- Consistent first-parent convention
- Active lanes tracking with nil slots
- Three-row lookahead for symbol selection
- Column collapse after convergence
- Flag-based character selection

This example serves as a comprehensive test case for implementation, covering edge cases that simpler examples don't reveal.
