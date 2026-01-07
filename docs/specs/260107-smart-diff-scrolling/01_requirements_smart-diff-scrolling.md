# Requirements: Smart Diff Scrolling

## Problem Statement

The current side-by-side diff view uses blank lines to keep both panels vertically aligned. This wastes screen space and creates visual noise, especially in diffs where one side has significantly more changes than the other.

IDE diff viewers (e.g., VSCode) solve this by using smart scrolling - both sides scroll independently but stay synchronized at change boundaries, eliminating the need for blank lines.

## Goals

- Remove blank lines from the side-by-side diff view
- Implement differential scrolling so corresponding content stays aligned
- Keep the hunk centered in the viewport during differential scrolling
- Maintain a smooth, intuitive scrolling experience

## Non-Goals (Out of Scope)

- Color changes (blue for modified lines, word-level highlighting) - future enhancement
- Changes to unified diff view
- Visible scrollbar implementation

## User Impact

Users viewing side-by-side diffs will see:
- More actual code on screen (no wasted space from blank lines)
- Corresponding lines aligned after each hunk
- Hunks centered in viewport with context above and below

## Key Requirements

### Scrolling Behavior

1. **Normal scrolling**: Both panels scroll together when outside of hunks

2. **Differential scrolling at hunks**: When a hunk has different sizes on each side:
   - The side with fewer lines scrolls slower
   - The side with more lines continues at normal rate
   - Differential scrolling occurs when the hunk is in the center of the viewport
   - Continues until the bottom of the hunk is aligned (first unchanged line after hunk appears at the same row on both sides)

3. **Symmetric behavior**: Same logic applies when scrolling up (aligns at top of hunk)

4. **Multiple hunks**: Each hunk independently triggers differential scrolling as it enters/exits the viewport center

5. **Large hunks**: When a hunk exceeds the viewport height, the smaller side stays frozen while the larger side scrolls through the entire hunk

6. **Page up/down**: Same differential scrolling logic applies for large jumps

### Visual Example

Two lines added, nothing deleted:

```
POSITION 1 (both scroll together, hunk approaching center)
┌─────────────────────┬─────────────────────┐
│   2  two            │   2  two            │
│   3  three          │   3  three          │
│   4  four           │   4  four           │  ← center
│   5  five           │   5  five           │  ← center
│   6  six            │+  6  NEW-A          │
│   7  seven          │+  7  NEW-B          │
└─────────────────────┴─────────────────────┘

POSITION 2 (hunk at center, only RIGHT scrolls)
┌─────────────────────┬─────────────────────┐
│   2  two            │   3  three          │
│   3  three          │   4  four           │
│   4  four           │   5  five           │  ← center
│   5  five           │+  6  NEW-A          │  ← center
│   6  six            │+  7  NEW-B          │
│   7  seven          │   8  six            │
└─────────────────────┴─────────────────────┘

POSITION 3 (aligned at center)
┌─────────────────────┬─────────────────────┐
│   2  two            │   4  four           │
│   3  three          │   5  five           │
│   4  four           │+  6  NEW-A          │  ← center
│   5  five           │+  7  NEW-B          │  ← center
│   6  six            │   8  six            │  ← aligned!
│   7  seven          │   9  seven          │  ← aligned!
└─────────────────────┴─────────────────────┘

POSITION 4 (both scroll together again)
┌─────────────────────┬─────────────────────┐
│   3  three          │   5  five           │
│   4  four           │+  6  NEW-A          │
│   5  five           │+  7  NEW-B          │  ← center
│   6  six            │   8  six            │  ← center, aligned
│   7  seven          │   9  seven          │
│   8  eight          │  10  eight          │
└─────────────────────┴─────────────────────┘
```

Key observations:
- No blank lines on either side
- Right scrolled 2 extra times to "absorb" its 2 added lines
- Content alignment preserved (left line 6 "six" = right line 8 "six")
- Hunk stays centered during differential scrolling

## Open Questions for Design Phase

1. How to efficiently track scroll positions for both panels independently?
2. How to detect when a hunk enters/exits the center region?
3. Data structure changes needed to remove blank line padding?
