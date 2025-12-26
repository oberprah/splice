# Requirements: Diff View Improvements

## Problem Statement

The current diff view has two usability issues:

1. **Misaligned modified lines**: When a line is modified (removed then added), the old and new versions appear on separate rows. The left side shows the old version with empty space on the right; the next row shows empty space on the left with the new version on the right. This makes it harder to compare what changed.

2. **No inline change highlighting**: When viewing a modified line, there's no indication of *what* changed within the line. Users must manually scan and compare to find the difference, even for small changes like a renamed variable.

### Current Behavior (Before)

```
┌─────────────────────────────────────┬─────────────────────────────────────┐
│ 1   func hello() {                  │                                     │
│ 2 - fmt.Println("Hello")            │                                     │
│                                     │ 1   func hello() {                  │
│                                     │ 2 + name := "World"                 │
│                                     │ 3 + fmt.Println("Hello " + name)    │
│ 3   return nil                      │                                     │
│ 4   }                               │                                     │
│                                     │ 4   return nil                      │
│                                     │ 5   }                               │
└─────────────────────────────────────┴─────────────────────────────────────┘
```

- Context lines appear only on one side at a time
- Removed and added lines on separate rows with empty opposite side
- Hard to visually track what corresponds to what

### Desired Behavior (After)

```
┌─────────────────────────────────────┬─────────────────────────────────────┐
│ 1   func hello() {                  │ 1   func hello() {                  │
│ 2 - fmt.Println("Hello")            │ 2 + name := "World"                 │
│   -                                 │ 3 + fmt.Println("Hello " + name)    │
│ 3   return nil                      │ 4   return nil                      │
│ 4   }                               │ 5   }                               │
└─────────────────────────────────────┴─────────────────────────────────────┘
```

- Context lines appear on both sides (aligned)
- Deleted lines on left, added lines on right, same rows
- Filler row (empty with background) when sides don't match up
- Much easier to see the relationship between old and new

### Inline Highlighting (Additional Enhancement)

For paired modified lines, highlight what changed within the line:

```
┌─────────────────────────────────────┬─────────────────────────────────────┐
│ 5 - fmt.Println(name)               │ 5 + fmt.Println(userName)           │
│     ─────────── ────                │     ─────────── ────────            │
│     subtle red  BOLD red            │     subtle green BOLD green         │
└─────────────────────────────────────┴─────────────────────────────────────┘
```

- The entire line has a subtle background color (red for removed, green for added)
- The changed portion (`name` vs `userName`) additionally gets a bolder/brighter version of that background color
- Unchanged portions within the line keep the standard subtle background

## Goals

1. **Side-by-side pairing**: When a line is modified, show the old version (left) and new version (right) on the same row, enabling direct visual comparison.

2. **Inline change highlighting**: Within paired modified lines, visually highlight the portions that changed (the specific characters or words that differ between old and new).

## Non-Goals

- Performance optimization for extremely large files (can iterate later)
- Configurable highlight colors or styles
- Complex edge case handling (keep implementation simple, iterate as needed)

## User Impact

- **Faster comprehension**: Users can immediately see what changed in a modified line without manual scanning
- **Reduced cognitive load**: Side-by-side comparison is more natural than mentally reconstructing changes across rows
- **Better code review experience**: Quickly spot the essence of a change

## Key Requirements

### Functional

1. **Aligned Side-by-Side Layout**
   - Context (unchanged) lines appear on both left and right sides, aligned on the same row
   - Removed lines appear on the left; added lines appear on the right
   - When one side has more lines than the other in a change region, use filler rows (empty with background color) to maintain alignment
   - Pairing strategy for inline highlighting: Prefer similarity matching if feasible; may fall back to simpler sequential pairing if complexity is too high

2. **Inline Change Highlighting**
   - Within paired lines, identify the changed portions
   - Granularity: Character-level or word-level (whichever is simpler to implement)
   - Changed portions use a brighter/bolder version of the existing diff background color
   - Unchanged portions within the line use the standard (subtle) diff background

### Non-Functional

- Keep implementation simple - this is an initial version that can be iterated on
- Maintain existing syntax highlighting (inline change highlighting should layer on top)
- Preserve existing keyboard navigation and viewport behavior

## Open Questions for Design Phase

1. What algorithm should be used for similarity matching? (e.g., Levenshtein distance, longest common subsequence)
2. What threshold determines if two lines are "similar enough" to pair?
3. How to handle cases where pairing is ambiguous (multiple lines with similar similarity scores)?
4. How to compute character/word-level diffs within paired lines?
5. What exact colors should "brighter/bolder" backgrounds use?

## References

- [Current implementation research](research/current-implementation.md)
