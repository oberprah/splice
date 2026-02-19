# Data Structure Options for Navigation

## Problem Statement

Current `ChangeIndices []int` stores individual line indices. This makes it hard to:
1. Determine when we've scrolled through an entire change (for smart navigation)
2. Group related changes (consecutive Modified/Added/Removed lines form one logical change)

## Option A: Keep Flat Alignments, Add Hunk Metadata

Keep the existing `[]Alignment` structure but add a parallel structure for navigation:

```go
type ChangeRegion struct {
    StartIdx int  // First alignment index in this region
    EndIdx   int  // Last alignment index in this region (inclusive)
}
```

**State becomes:**
```go
type State struct {
    // ... existing fields ...
    ChangeRegions []ChangeRegion  // Groups of consecutive changes
}
```

**Pros:**
- Minimal change to existing code
- View rendering unchanged
- Easy to compute from existing alignments

**Cons:**
- Two parallel data structures that must stay in sync
- "Regions" are computed, not first-class concepts

## Option B: Two-Level Structure (User's Suggestion)

Replace flat `[]Alignment` with a list of "blocks" where each block is either unchanged or a change:

```go
type Block interface {
    block()
}

type UnchangedBlock struct {
    Alignments []UnchangedAlignment  // Consecutive unchanged lines
}

type ChangeBlock struct {
    Alignments []Alignment  // Mix of Modified, Added, Removed
}
```

**State becomes:**
```go
type AlignedFileDiff struct {
    Left   FileContent
    Right  FileContent
    Blocks []Block  // Replaces Alignments
}
```

**Pros:**
- Cleaner conceptual model - matches how diffs are structured
- Navigation naturally operates on blocks
- Easy to answer "what block am I in?" and "where does this block end?"

**Cons:**
- Breaking change to view rendering (needs to iterate blocks then alignments)
- More complex data structure
- Need to rewrite `BuildAlignments` function

## Option C: Flat Alignments + Region Index

Keep `[]Alignment` flat but add computed region boundaries:

```go
type AlignedFileDiff struct {
    Left       FileContent
    Right      FileContent
    Alignments []Alignment
    Regions    []Region  // Computed during build
}

type Region struct {
    Start    int
    End      int
    IsChange bool  // true for change regions, false for unchanged
}
```

**Pros:**
- View rendering stays simple (iterate flat alignments)
- Navigation uses regions for smart behavior
- Computed once during build, not on every navigation

**Cons:**
- Still two parallel structures
- Less elegant than Option B

## Decision Criteria

1. **Conceptual clarity**: Does the structure match how users think about diffs?
2. **Navigation simplicity**: Is it easy to implement smart navigation?
3. **View compatibility**: How much does the view need to change?
4. **Change scope**: How much existing code needs modification?

## Recommendation

**Option B (Two-Level Structure)** is the cleanest conceptual model and aligns with the user's suggestion. It makes navigation logic straightforward:
- Next/previous change = next/previous `ChangeBlock`
- Scrolling through a change = scrolling within current block's alignments
- End of change = end of current block

The view changes are mechanical (one more level of iteration) but result in cleaner code overall.
