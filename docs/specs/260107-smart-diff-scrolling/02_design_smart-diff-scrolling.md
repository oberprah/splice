# Design: Smart Diff Scrolling

## Executive Summary

Smart diff scrolling eliminates blank line padding from the side-by-side diff view by allowing each panel to scroll independently while staying synchronized at change boundaries. When a hunk (region of changes) is centered in the viewport, the side with more lines scrolls faster so both sides exit the hunk aligned.

The solution introduces a **segment-based data model** where the diff is represented as a sequence of `UnchangedSegment` and `HunkSegment` entries. Unchanged segments scroll both panels together; hunk segments trigger differential scrolling based on the line count difference between sides. This structure directly models the desired behavior and eliminates the need for blank line padding entirely.

The initial implementation will simplify rendering by omitting inline (word-level) diff highlighting, focusing on correct scrolling behavior. Inline highlighting can be added as a future enhancement.

## Context & Problem Statement

The current side-by-side diff view uses blank lines to keep panels vertically aligned. When one side has additions without corresponding deletions, the other side shows blank lines. This wastes screen space and creates visual noise, especially in diffs where changes are heavily asymmetric.

```
Current behavior (with blanks):          Desired behavior (no blanks):
┌─────────────┬─────────────┐           ┌─────────────┬─────────────┐
│ - old line  │             │           │ - old line  │ + new line A│
│             │ + new line A│    →      │   context   │ + new line B│
│             │ + new line B│           │             │   context   │
│   context   │   context   │           └─────────────┴─────────────┘
└─────────────┴─────────────┘           (panels scroll independently)
```

IDE diff viewers (VSCode, IntelliJ) solve this with smart scrolling—both sides scroll independently but resynchronize at change boundaries.

**Scope:**
- This design covers the side-by-side diff view only
- Unified diff view is unaffected
- Color/styling changes (blue for modified, word-level highlighting) are out of scope—can be added later

## Current State

### Data Model

The current `AlignedFileDiff` uses a flat `Alignments` array where each entry represents one display row:

```go
type AlignedFileDiff struct {
    Left       FileContent
    Right      FileContent
    Alignments []Alignment  // One per display row
}

type Alignment interface { alignment() }  // Sum type

type UnchangedAlignment struct { LeftIdx, RightIdx int }
type ModifiedAlignment struct { LeftIdx, RightIdx int; InlineDiff ... }
type RemovedAlignment struct { LeftIdx int }  // Right side is blank
type AddedAlignment struct { RightIdx int }   // Left side is blank
```

Blank lines are generated implicitly: `RemovedAlignment` renders left content + blank right, `AddedAlignment` renders blank left + right content.

### Viewport

A single `ViewportStart` integer controls both panels:

```go
type State struct {
    Diff           *diff.AlignedFileDiff
    ViewportStart  int  // Index into Alignments
}
```

Both panels always show `Alignments[ViewportStart : ViewportStart+viewportHeight]`.

### Limitations for Smart Scrolling

1. **Single viewport**: Can't scroll panels independently
2. **One-to-one mapping**: Each alignment = one row, enforces equal panel heights
3. **Implicit blanks**: Blank lines exist only at render time, not modeled
4. **No hunk boundaries**: Hunks aren't tracked after alignment building

## Solution

### Segment-Based Data Model

> **Decision:** Replace the flat `Alignments` array with a `Segments` slice where each segment is either an unchanged region or a hunk. This directly models scrolling behavior—unchanged regions scroll together, hunks scroll differentially.

The new structure:

```
┌─────────────────────────────────────────────────────┐
│                  AlignedFileDiff                    │
├─────────────────────────────────────────────────────┤
│  Left: FileContent                                  │
│  Right: FileContent                                 │
│  Segments: []Segment                                │
│    ├── UnchangedSegment (lines 1-10)               │
│    ├── HunkSegment (left: 3 lines, right: 5 lines) │
│    ├── UnchangedSegment (lines 14-50)              │
│    ├── HunkSegment (left: 8 lines, right: 2 lines) │
│    └── UnchangedSegment (lines 53-100)             │
└─────────────────────────────────────────────────────┘
```

### Data Types

```go
// Segment represents a contiguous region of the diff
type Segment interface {
    segment()  // Sealed marker
}

// UnchangedSegment: identical content on both sides, scroll together
type UnchangedSegment struct {
    LeftStart  int  // Start index into Left.Lines
    RightStart int  // Start index into Right.Lines
    Count      int  // Number of lines (same for both sides)
}

// HunkSegment: changed content, each side has its own lines
type HunkSegment struct {
    LeftLines  []HunkLine  // Lines on left (removals + modified-old)
    RightLines []HunkLine  // Lines on right (additions + modified-new)
}

// HunkLine: a single line within a hunk
type HunkLine struct {
    SourceIdx int       // Index into Left.Lines or Right.Lines
    Type      LineType  // Added, Removed, or Modified
}

type LineType int
const (
    LineTypeAdded LineType = iota
    LineTypeRemoved
    LineTypeModified
)
```

### State and Scroll Position

> **Decision:** Track scroll position as a segment index plus left/right offsets within that segment. This enables independent scrolling while maintaining segment-level synchronization.

```go
type State struct {
    Diff *diff.AlignedFileDiff

    // Scroll position
    SegmentIndex  int  // Which segment the viewport starts in
    LeftOffset    int  // Line offset for left panel within segment
    RightOffset   int  // Line offset for right panel within segment
}
```

For an `UnchangedSegment`, `LeftOffset == RightOffset` always (they scroll together).
For a `HunkSegment`, offsets can differ during differential scrolling.

### Scrolling Behavior

```
                    ┌────────────────────────────────────┐
                    │         Scroll Down Logic          │
                    └────────────────────────────────────┘
                                    │
                                    ▼
                    ┌────────────────────────────────────┐
                    │  Is current segment a hunk AND     │
                    │  is hunk centered in viewport?     │
                    └────────────────────────────────────┘
                           │                    │
                          YES                  NO
                           │                    │
                           ▼                    ▼
              ┌─────────────────────┐  ┌─────────────────────┐
              │  Differential scroll │  │  Normal scroll:      │
              │  - Advance larger    │  │  - Increment both    │
              │    side faster       │  │    LeftOffset and    │
              │  - Smaller side      │  │    RightOffset       │
              │    advances slower   │  └─────────────────────┘
              │    or freezes        │
              └─────────────────────┘
                           │
                           ▼
              ┌─────────────────────────────────────┐
              │  When both sides have scrolled      │
              │  through entire hunk → move to next │
              │  segment with offsets = 0           │
              └─────────────────────────────────────┘
```

#### Differential Scroll Rate Calculation

When a hunk is at viewport center with `leftCount` and `rightCount` lines:

```
ratio = max(leftCount, rightCount) / min(leftCount, rightCount)

If left is larger:
  - Left advances every scroll step
  - Right advances every `ratio` scroll steps

If right is larger:
  - Right advances every scroll step
  - Left advances every `ratio` scroll steps
```

> **Decision:** Use simple integer ratio-based scrolling rather than fractional positions. This keeps the implementation straightforward and the scrolling predictable.

Example: Left has 6 lines, right has 2 lines, ratio = 3.
- Scroll step 1: Left +1, Right +0
- Scroll step 2: Left +1, Right +0
- Scroll step 3: Left +1, Right +1
- Scroll step 4: Left +1, Right +0
- Scroll step 5: Left +1, Right +0
- Scroll step 6: Left +1, Right +1 → both done, move to next segment

### Viewport Center Detection

The "center" is the middle rows of the viewport. A hunk is "centered" when any part of it overlaps this region.

```go
const centerZoneRatio = 0.4  // Middle 40% of viewport

func isHunkCentered(viewportHeight int, hunkTopRow, hunkBottomRow int) bool {
    centerStart := viewportHeight * 0.3  // 30% from top
    centerEnd := viewportHeight * 0.7    // 70% from top

    // Hunk overlaps center zone
    return hunkTopRow < centerEnd && hunkBottomRow > centerStart
}
```

> **Decision:** Use a 40% center zone (from 30%-70% of viewport). This provides a reasonable buffer before differential scrolling kicks in, preventing jarring transitions at viewport edges.

### Rendering

Rendering collects lines from both panels independently until the viewport is filled:

```
function renderViewport(viewportHeight):
    leftLines = []
    rightLines = []

    seg = segments[segmentIndex]
    leftOff = leftOffset
    rightOff = rightOffset

    while len(leftLines) < viewportHeight:
        if seg is UnchangedSegment:
            // Add matching lines to both panels
            for i = leftOff; i < seg.Count && len(leftLines) < viewportHeight; i++:
                leftLines.add(seg.line(i))
                rightLines.add(seg.line(i))
            // Move to next segment

        if seg is HunkSegment:
            // Add lines independently
            // Left panel shows LeftLines[leftOff...]
            // Right panel shows RightLines[rightOff...]
            // May result in different content at same row

    return leftLines, rightLines
```

The key difference from current rendering: panels can show different content at the same viewport row during hunk display.

### Navigation (n/N Jump to Change)

Change navigation jumps between hunk segments:

```go
func jumpToNextHunk() {
    for i := s.SegmentIndex + 1; i < len(s.Diff.Segments); i++ {
        if _, isHunk := s.Diff.Segments[i].(HunkSegment); isHunk {
            s.SegmentIndex = i
            s.LeftOffset = 0
            s.RightOffset = 0
            return
        }
    }
}
```

### Building Segments from Parsed Diff

The segment building algorithm processes the existing parsed diff output:

```
1. Walk through both files using diff line information
2. Accumulate consecutive unchanged lines into UnchangedSegment
3. When encountering a change:
   a. Flush any accumulated unchanged lines as UnchangedSegment
   b. Collect all consecutive changed lines (removed/added/modified)
   c. Build HunkSegment with LeftLines and RightLines
4. Continue until end of files
```

This is similar to the current `BuildAlignments` logic but outputs segments instead of individual alignments.

### Simplifications for Initial Implementation

> **Decision:** Omit inline (word-level) diff highlighting for the initial implementation. Focus on correct scrolling behavior first; inline highlighting can be added as a future enhancement.

Rationale: The current inline diff logic is complex (character-by-character rendering with highlight maps). Getting scrolling right is the priority. Without inline diffs:
- `ModifiedAlignment` effectively becomes `RemovedAlignment` + `AddedAlignment` (show old line as removed, new line as added)
- No need for `LinePair` or `InlineDiff` in the initial data model
- Simpler rendering: just apply line-level styling (red for removed, green for added)

This can be enhanced later by adding pairing information to `HunkSegment` for lines that are modifications rather than pure add/remove.

## Component Interactions

```
┌─────────────────────────────────────────────────────────────────┐
│                           User Input                            │
│                      (j/k, Ctrl+D/U, n/N)                       │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        DiffState.Update                         │
│  - Determines scroll type (normal vs differential)              │
│  - Updates SegmentIndex, LeftOffset, RightOffset                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         DiffState.View                          │
│  - Walks segments from current position                         │
│  - Builds left and right column content independently           │
│  - Composes split view                                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         ViewBuilder                             │
│  - AddSplitView joins left and right columns                    │
│  - Handles panels with potentially different content per row    │
└─────────────────────────────────────────────────────────────────┘
```

## Visual Walkthrough

Using the example from the requirements (two lines added, nothing deleted):

```
Initial state:
  Segments: [UnchangedSeg(1-5), HunkSeg(left:0, right:2), UnchangedSeg(6-10)]
  Position: Seg=0, LeftOff=0, RightOff=0

Display (hunk not yet at center):
┌─────────────────────┬─────────────────────┐
│   1  one            │   1  one            │
│   2  two            │   2  two            │
│   3  three          │   3  three          │  ← center zone
│   4  four           │   4  four           │  ← center zone
│   5  five           │   5  five           │
│   6  six            │+  6  NEW-A          │  ← hunk starts
└─────────────────────┴─────────────────────┘

After scrolling down (hunk enters center):
┌─────────────────────┬─────────────────────┐
│   3  three          │   4  four           │
│   4  four           │   5  five           │  ← center zone
│   5  five           │+  6  NEW-A          │  ← center zone, differential!
│   6  six            │+  7  NEW-B          │
│   7  seven          │   8  six            │
│   8  eight          │   9  seven          │
└─────────────────────┴─────────────────────┘
  (Left at seg[2] offset 0, Right still processing hunk)

After differential scrolling completes:
┌─────────────────────┬─────────────────────┐
│   4  four           │   5  five           │
│   5  five           │+  6  NEW-A          │
│   6  six            │+  7  NEW-B          │  ← center zone
│   7  seven          │   8  six            │  ← center zone, aligned!
│   8  eight          │   9  seven          │
│   9  nine           │  10  eight          │
└─────────────────────┴─────────────────────┘
  (Both now in unchanged segment, synchronized)
```

## Open Questions

No open questions - the design is ready for implementation review.

The main trade-off (omitting inline diff highlighting initially) was made explicitly to reduce complexity and focus on the core scrolling feature. This can be revisited as a follow-up enhancement.
