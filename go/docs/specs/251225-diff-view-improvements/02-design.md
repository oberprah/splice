# Design: Diff View Improvements

## Executive Summary

This design addresses two usability issues in Splice's diff view: misaligned modified lines and lack of inline change highlighting. The solution involves two complementary features:

1. **Line Pairing**: Match removed and added lines that are modifications of each other, displaying them on the same row for direct visual comparison. Uses token-based Dice coefficient similarity with greedy matching - simple, fast (O(m×n)), and good enough for typical code diffs.

2. **Inline Highlighting**: For paired lines, highlight the specific characters that changed using character-level Myers diff via the `sergi/go-diff` library. Changed portions get a brighter background; unchanged portions keep the standard diff background.

The approach introduces a cleaner data model that separates file content from alignment logic, using Go's sum type pattern (interface + concrete types) to make illegal states unrepresentable. One external dependency (`sergi/go-diff`), ~300-400 lines of new code.

## Context & Problem Statement

Splice's diff view currently displays removed and added lines on separate rows, even when they represent a modification to the same line. This forces users to mentally match corresponding lines and scan for differences.

```
Current behavior (misaligned):
┌─────────────────────────────────────┬─────────────────────────────────────┐
│ 2 - fmt.Println("Hello")            │                                     │
│                                     │ 2 + fmt.Println("Hello World")      │
└─────────────────────────────────────┴─────────────────────────────────────┘

Desired behavior (aligned with inline highlighting):
┌─────────────────────────────────────┬─────────────────────────────────────┐
│ 2 - fmt.Println("Hello")            │ 2 + fmt.Println("Hello World")      │
│                 ───────             │                 ─────────────       │
│                 subtle              │                 BRIGHT green        │
└─────────────────────────────────────┴─────────────────────────────────────┘
```

**Scope**: This design covers line pairing and inline highlighting. It does not cover performance optimization for extremely large files or configurable highlighting colors.

## Current State

The diff view uses a two-pointer walk algorithm (`MergeFullFile` in `internal/diff/merge.go`) to produce `FullFileLine` structs:

```go
type FullFileLine struct {
    LeftLineNo   int              // 0 if not present on left
    RightLineNo  int              // 0 if not present on right
    LeftTokens   []highlight.Token
    RightTokens  []highlight.Token
    Change       ChangeType       // Unchanged, Added, or Removed
}
```

**Current rendering behavior**:
- Unchanged lines: Both sides populated, shown on same row
- Removed lines: Only left side populated, right side empty
- Added lines: Only right side populated, left side empty

**Key limitation**: The algorithm has no concept of "pairing" - it doesn't know which removed line corresponds to which added line. Each changed line is processed independently.

## Solution

### Overview

The solution builds file content (with syntax tokens) for each side, then produces an alignment sequence that pairs lines and computes inline diffs. Processing operates at the "hunk" level - contiguous regions of changes bounded by unchanged lines. Within each hunk:

1. **Collect** removed and added line indices
2. **Pair** them using similarity-based matching
3. **Compute inline diffs** for each pair
4. **Emit alignments** for rendering

```
Example: 2 removed lines, 2 added lines (one pairs, one doesn't)

Left file (old)              Right file (new)           Alignments
────────────                 ───────────────            ──────────
Line 1: func hello()         Line 1: func hello()   →   UnchangedAlignment{0, 0}
Line 2: print("Hi")          Line 2: print("Hello") →   ModifiedAlignment{1, 1, diffs}
Line 3: print("World")                              →   RemovedAlignment{2}
                             Line 3: return nil     →   AddedAlignment{2}
Line 4: }                    Line 4: }              →   UnchangedAlignment{3, 3}
```

### Component 1: Line Pairing Algorithm

**Goal**: Given removed lines R and added lines A, find pairs (r, a) where r and a are likely modifications of each other.

> **Decision**: Use token-based Dice coefficient similarity with greedy matching.
>
> Alternatives considered:
> - *Levenshtein distance*: More accurate but O(m×n) per pair comparison is expensive for long lines. Dice is O(m+n) per comparison.
> - *Sequential pairing*: Trivial but produces poor matches when lines are inserted/deleted.
> - *Hungarian algorithm*: Optimal global matching but O(n³) and complex - overkill for typical diff sizes.
>
> Dice coefficient is simple, fast, and works well for code where identifiers/keywords provide strong signals.

**Similarity function**:

```
Dice(A, B) = 2 × |tokens(A) ∩ tokens(B)| / (|tokens(A)| + |tokens(B)|)
```

- Tokenize by splitting on non-alphanumeric characters
- `"fmt.Println(name)"` → `["fmt", "Println", "name"]`
- Score ranges from 0 (no overlap) to 1 (identical)

> **Decision**: Use 0.5 as the similarity threshold.
>
> Lines below this threshold are left unpaired. Research suggests 0.4-0.6 works well for code; we'll start at 0.5 and tune based on real-world testing. The threshold can be adjusted without API changes.

**Pairing algorithm (greedy)**:

1. Compute similarity scores for all (removed, added) pairs
2. Filter pairs below threshold
3. Sort remaining pairs by score (descending)
4. Greedily match: take highest-scoring pair, mark both lines as used, repeat

This produces good results in practice. If user feedback reveals poor pairings, we can upgrade to Hungarian algorithm later.

### Component 2: Inline Diff Highlighting

**Goal**: For each paired (removed, added) line, identify which characters changed.

> **Decision**: Use character-level Myers diff via `sergi/go-diff`.
>
> Alternatives considered:
> - *Word-level diff*: Requires tokenization, loses precision for operators (`==` vs `===`).
> - *Custom LCS implementation*: Possible (~200 lines) but unnecessary given quality of sergi/go-diff.
> - *pmezard/go-difflib*: Abandoned by maintainer; uses non-minimal algorithm.
>
> Character-level is essential for code diffs where single characters matter. sergi/go-diff is battle-tested (based on Google's diff-match-patch), well-maintained, and has a clean API.

**Output format**:

```go
type Diff struct {
    Type Operation  // DiffEqual, DiffInsert, or DiffDelete
    Text string
}

// Example: "Hello" → "Hello World"
[]Diff{
    {Type: DiffEqual, Text: "Hello"},
    {Type: DiffInsert, Text: " World"},
}
```

For rendering:
- On the **old line**: Show `DiffEqual` + `DiffDelete` segments (ignore `DiffInsert`)
- On the **new line**: Show `DiffEqual` + `DiffInsert` segments (ignore `DiffDelete`)
- `DiffEqual` portions get the standard diff background
- `DiffDelete`/`DiffInsert` portions get a brighter version of the background

### Component 3: Data Structure Changes

Replace the existing `FullFileLine` model with a cleaner separation of concerns:

```go
// ═══════════════════════════════════════════════════════════
// CONTENT: File lines with syntax highlighting (no layout concerns)
// ═══════════════════════════════════════════════════════════

type Line struct {
    Tokens []highlight.Token
}

// Text returns the raw text content (used for similarity matching)
func (l *Line) Text() string {
    var b strings.Builder
    for _, t := range l.Tokens {
        b.WriteString(t.Text)
    }
    return b.String()
}

type FileContent struct {
    Path  string
    Lines []Line
}

// LineNo returns the 1-indexed line number for display
func (fc *FileContent) LineNo(idx int) int {
    return idx + 1
}

// ═══════════════════════════════════════════════════════════
// ALIGNMENT: Sum type representing how lines relate
// ═══════════════════════════════════════════════════════════

// Alignment is a sum type - exactly one of the concrete types below
type Alignment interface {
    alignment()  // marker method (unexported = sealed)
}

type UnchangedAlignment struct {
    LeftIdx  int
    RightIdx int
}

type ModifiedAlignment struct {
    LeftIdx    int
    RightIdx   int
    InlineDiff []diffmatchpatch.Diff
}

type RemovedAlignment struct {
    LeftIdx int
    // Right side is implicitly a filler - no field needed
}

type AddedAlignment struct {
    RightIdx int
    // Left side is implicitly a filler - no field needed
}

func (UnchangedAlignment) alignment() {}
func (ModifiedAlignment) alignment()  {}
func (RemovedAlignment) alignment()   {}
func (AddedAlignment) alignment()     {}

// ═══════════════════════════════════════════════════════════
// TOP-LEVEL: Combines content + alignment
// ═══════════════════════════════════════════════════════════

type FileDiff struct {
    Left       FileContent
    Right      FileContent
    Alignments []Alignment  // one entry per display row
}
```

> **Decision**: Use a sum type (interface + concrete types) for alignments rather than a single struct with optional fields.
>
> Benefits:
> - `InlineDiff` only exists on `ModifiedAlignment` - impossible to set it elsewhere
> - Filler indices are implicit - `RemovedAlignment` has no `RightIdx` field
> - Type switch in rendering explicitly handles all cases (compiler-checked)
> - Clear separation: `FileContent` knows nothing about alignment; `Alignment` knows nothing about tokens

### Component 4: Rendering Updates

**Main render loop using type switch**:

```go
for _, align := range diff.Alignments {
    switch a := align.(type) {
    case UnchangedAlignment: // render both sides with neutral background
    case ModifiedAlignment:  // render both sides with inline diff highlighting
    case RemovedAlignment:   // render left side + filler on right
    case AddedAlignment:     // render filler on left + right side
    }
}
```

**For `ModifiedAlignment` rows (inline highlighting)**:

1. Render left column with removed line + inline highlighting
2. Render right column with added line + inline highlighting
3. For inline highlighting:
   - Iterate through `InlineDiff`
   - On left side: render `DiffEqual` with subtle background, `DiffDelete` with bright background
   - On right side: render `DiffEqual` with subtle background, `DiffInsert` with bright background

**For `RemovedAlignment` / `AddedAlignment` rows (filler)**:

- Render the populated side normally
- Render the empty side with appropriate filler background color to maintain visual alignment

**Style integration with syntax highlighting**:

The existing `renderTokens` function applies syntax foreground colors combined with diff background colors via `lipgloss.Inherit()`. For inline highlighting:

1. Break syntax tokens at inline diff boundaries
2. Apply three-layer styling: syntax foreground + diff background + inline highlight background
3. Use brighter versions of existing `DiffAdditionsStyle` / `DiffDeletionsStyle` for changed portions

> **Decision**: Use 1.3× brighter backgrounds for changed portions within lines.
>
> VS Code uses 1.4× for dark themes. We'll start with 1.3× as a conservative choice that maintains readability while providing clear differentiation. The exact values can be tuned.

### Processing Pipeline

```
Git diff output
    ↓
ParseUnifiedDiff()              → raw hunks with line ranges
    ↓
BuildFileContent(oldFile)       → Left FileContent (with syntax tokens)
BuildFileContent(newFile)       → Right FileContent (with syntax tokens)
    ↓
BuildAlignments(Left, Right, hunks)
    ├── Walk through file lines
    ├── Unchanged regions → UnchangedAlignment
    ├── Changed regions (hunks):
    │   ├── Collect removed/added line indices
    │   ├── Compute similarity matrix (Dice)
    │   ├── Greedy pair matching
    │   ├── Compute inline diffs for pairs (Myers)
    │   └── Emit ModifiedAlignment / RemovedAlignment / AddedAlignment
    └── Return []Alignment
    ↓
FileDiff{Left, Right, Alignments}
    ↓
DiffState.View()                → type switch over Alignments, render rows
```

**Key functions**:

- `BuildFileContent(path string, content string) FileContent`: Tokenizes file content using Chroma, produces `[]Line` with syntax tokens
- `BuildAlignments(left, right FileContent, hunks []Hunk) []Alignment`: Main algorithm that produces the alignment sequence with pairing and inline diffs

### State Diagram

```
                    ┌──────────────────────────────────────┐
                    │    BuildAlignments: Scanning Lines   │
                    └──────────────────────────────────────┘
                                      │
              ┌───────────────────────┼───────────────────────┐
              │                       │                       │
              ▼                       ▼                       ▼
     ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
     │  Unchanged  │         │   Removed   │         │    Added    │
     │   (emit     │         │  (collect   │         │  (collect   │
     │  Unchanged- │         │   in hunk)  │         │   in hunk)  │
     │  Alignment) │         └─────────────┘         └─────────────┘
     └─────────────┘                 │                       │
              │                      └───────────┬───────────┘
              │                                  │
              │                                  ▼
              │                   ┌──────────────────────────┐
              │                   │    Hunk Complete?        │
              │                   │ (hit unchanged or EOF)   │
              │                   └──────────────────────────┘
              │                                  │
              │                                  ▼
              │                   ┌──────────────────────────┐
              │                   │   Pair & Align Hunk      │
              │                   │   1. Compute similarity  │
              │                   │   2. Greedy match        │
              │                   │   3. Compute inline diff │
              │                   │   4. Emit alignments     │
              │                   └──────────────────────────┘
              │                                  │
              └──────────────────────────────────┘
                                     │
                                     ▼
                          ┌──────────────────┐
                          │  Output:         │
                          │  []Alignment     │
                          │  ┌────────────┐  │
                          │  │ Unchanged  │  │
                          │  │ Modified   │  │
                          │  │ Removed    │  │
                          │  │ Added      │  │
                          │  └────────────┘  │
                          └──────────────────┘
```

### Edge Cases

1. **Empty lines**: Empty lines can create false similarity matches. Handle by giving empty lines a similarity score of 0 with each other unless they're identical.

2. **Many-to-one changes**: If 2 lines are removed and 1 is added (or vice versa), some lines will be unpaired. Display unpaired lines as before (alone with filler on opposite side).

3. **No good matches**: If all similarity scores are below threshold, treat as pure removals + pure additions (current behavior).

4. **Identical lines in different positions**: Rare in practice. Greedy algorithm naturally handles by pairing the first high-scoring match found.

5. **Very long lines**: Dice coefficient is O(m+n) so handles long lines well. For inline diff, character-level Myers is O(n×d) where d is number of differences - fast for typical lines.

## Migration Notes

This design replaces the existing `FullFileLine` / `FullFileDiff` model with a new `FileDiff` structure. Key changes:

| Current | New |
|---------|-----|
| `FullFileLine` with mixed left/right data | Separate `FileContent` per side |
| `ChangeType` enum (Unchanged/Added/Removed) | Sum type with 4 alignment types |
| `MergeFullFile()` produces interleaved rows | `BuildAlignments()` produces alignment sequence |
| Implicit fillers (`LineNo == 0`) | Explicit via `RemovedAlignment` / `AddedAlignment` |

## Open Questions

None. All design decisions have been made based on research findings. Implementation can proceed without further input.

The similarity threshold (0.5) and brightness multiplier (1.3×) are tunable parameters that can be adjusted based on user feedback without changing the design.
