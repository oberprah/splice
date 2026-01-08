# Data Structure Options for Smart Diff Scrolling

## Problem Summary

The current implementation uses a single `ViewportStart` for both panels and relies on blank lines (via `RemovedAlignment`/`AddedAlignment`) to keep panels aligned. Smart differential scrolling requires:

1. Independent scroll positions for left and right panels
2. No blank line padding
3. Ability to detect when hunks are at viewport center
4. Different scroll rates when processing hunks

## Option 1: Keep Current Structure + Add Virtual Scroll Offsets

### Approach
Keep the existing `Alignments` array with its one-to-one mapping, but add "virtual" scroll offsets that create the illusion of independent scrolling.

```go
type State struct {
    Diff        *diff.AlignedFileDiff
    ViewportStart    int  // Base position in Alignments
    LeftScrollOffset  int  // Additional offset for left panel (can be negative)
    RightScrollOffset int  // Additional offset for right panel (can be negative)
}
```

### Rendering Logic
- Left panel renders from `ViewportStart + LeftScrollOffset`
- Right panel renders from `ViewportStart + RightScrollOffset`
- During differential scrolling in a hunk, increment one offset while keeping the other fixed

### Pros
- Minimal changes to existing code
- `Alignments` structure unchanged
- Easy to reason about "base" position

### Cons
- Offsets don't map cleanly to actual line numbers (alignments include blanks)
- Complex edge cases when offsets cause one panel to overflow
- Doesn't actually eliminate blank lines - just hides them differently
- Hard to track "which real line is at row N" for each panel

### Verdict: **Not recommended** - Doesn't solve the core problem. Still operates on alignment indices that include blank padding.

---

## Option 2: Separate Line Arrays + Hunk Boundaries

### Approach
Fundamentally restructure the data model. Store left and right lines independently without any padding alignment, then track hunk boundaries separately for synchronization.

```go
type AlignedFileDiff struct {
    Left       FileContent       // Old version lines (no padding)
    Right      FileContent       // New version lines (no padding)
    Hunks      []Hunk            // Explicit hunk boundaries
}

type Hunk struct {
    // Left side boundaries (indices into Left.Lines)
    LeftStart  int  // First line of hunk on left
    LeftEnd    int  // One past last line of hunk on left

    // Right side boundaries (indices into Right.Lines)
    RightStart int  // First line of hunk on right
    RightEnd   int  // One past last line of hunk on right

    // Precomputed for scroll rate calculation
    LeftCount  int  // LeftEnd - LeftStart
    RightCount int  // RightEnd - RightStart
}

type State struct {
    Diff             *diff.AlignedFileDiff
    LeftViewportStart  int  // Line index in Left.Lines
    RightViewportStart int  // Line index in Right.Lines
}
```

### Rendering Logic
1. Left panel renders `Left.Lines[LeftViewportStart : LeftViewportStart+viewportHeight]`
2. Right panel renders `Right.Lines[RightViewportStart : RightViewportStart+viewportHeight]`
3. No blank lines - just real content on each side

### Scrolling Logic
1. On scroll down:
   - Check if any hunk overlaps the viewport center region
   - If no hunk at center: increment both `LeftViewportStart` and `RightViewportStart`
   - If hunk at center: apply differential scrolling based on hunk sizes
2. Hunk boundaries tell us when left/right should be aligned again

### How to Determine Change Type per Line
Without alignments, we need another way to know if a line is added/removed/modified:

```go
type FileContent struct {
    Path  string
    Lines []AlignedLine
    LineTypes []LineDisplayType  // Parallel array: Unchanged, Added, Removed, Modified
}

type LineDisplayType int
const (
    Unchanged LineDisplayType = iota
    Added
    Removed
    Modified
)
```

### Pros
- Clean separation: content is content, hunks are hunks
- Independent scroll positions map directly to line numbers
- No wasted blank lines
- Explicit hunk tracking makes center detection straightforward
- Easy to reason about: "left is showing lines 50-80, right is showing lines 52-82"

### Cons
- Significant refactor of data structures
- Loses inline diff information (but can be preserved with additional structure)
- Need to handle rendering when panels have different content types at same viewport row

### Verdict: **Strong candidate** - Clean model, but loses some information from current alignments.

---

## Option 3: Hybrid - Keep Alignments for Content Type, Add Hunk Tracking

### Approach
Keep the current `Alignment` sum type for determining how to render lines, but:
1. Build a separate index for hunk boundaries
2. Track separate viewport positions for left and right
3. Map viewport positions to alignment indices for rendering

```go
type AlignedFileDiff struct {
    Left       FileContent
    Right      FileContent
    Alignments []Alignment  // Kept for content type info
    Hunks      []HunkSpan   // NEW: explicit hunk tracking
}

type HunkSpan struct {
    AlignmentStart int  // First alignment index in this hunk
    AlignmentEnd   int  // One past last alignment index

    LeftLineStart  int  // First line index on left (excluding blanks)
    LeftLineEnd    int  // One past last line on left
    RightLineStart int  // First line index on right
    RightLineEnd   int  // One past last line on right
}

type State struct {
    Diff               *diff.AlignedFileDiff
    LeftViewportStart  int  // Index into Left.Lines (not Alignments)
    RightViewportStart int  // Index into Right.Lines (not Alignments)
}
```

### Rendering Logic
This is where it gets complex. We need to map line indices to display rows:

```go
func (s *State) renderViewport() (leftLines, rightLines []string) {
    // For each row in viewport:
    for row := 0; row < viewportHeight; row++ {
        leftLineIdx := s.LeftViewportStart + row
        rightLineIdx := s.RightViewportStart + row

        // Find what to display at this row for each side
        leftContent := s.renderLeftLine(leftLineIdx)   // May be real content or empty
        rightContent := s.renderRightLine(rightLineIdx) // May be real content or empty

        leftLines = append(leftLines, leftContent)
        rightLines = append(rightLines, rightContent)
    }
}
```

### Pros
- Preserves all existing information (inline diffs, line pairing)
- Explicit hunk tracking for differential scrolling
- Independent line-based viewport positions

### Cons
- Complex mapping between line indices and alignments
- Rendering logic becomes significantly more complex
- Dual indexing (by alignment vs by line) is error-prone

### Verdict: **Possible but complex** - Trying to preserve too much may overcomplicate things.

---

## Option 4: Segment-Based Model (Recommended)

### Approach
Model the diff as a sequence of **segments**, where each segment is either:
- An **unchanged region**: same lines on both sides, scroll together
- A **hunk**: different content on left and right, differential scrolling

```go
type Segment interface {
    segment()
}

type UnchangedSegment struct {
    StartLineNo int      // Starting line number (same for both sides)
    Lines       []AlignedLine  // The actual lines (shared, since identical)
    Count       int      // Number of lines
}

type HunkSegment struct {
    LeftLines  []HunkLine   // Lines on left side
    RightLines []HunkLine   // Lines on right side
    Pairs      []LinePair   // Optional: pairs for modified lines (for inline diff)
}

type HunkLine struct {
    LineNo int
    Line   AlignedLine
    Type   LineType  // Removed, Added, or Modified
}

type LinePair struct {
    LeftIdx    int  // Index into LeftLines
    RightIdx   int  // Index into RightLines
    InlineDiff []diffmatchpatch.Diff
}

type AlignedFileDiff struct {
    Left     FileContent
    Right    FileContent
    Segments []Segment
}
```

### State Tracking
```go
type ScrollPosition struct {
    SegmentIndex int     // Which segment we're in
    LeftOffset   int     // Line offset within segment for left panel
    RightOffset  int     // Line offset within segment for right panel
}

type State struct {
    Diff     *diff.AlignedFileDiff
    Position ScrollPosition
}
```

### Scrolling Logic

**In UnchangedSegment:**
- Both offsets advance together
- When reaching end of segment, move to next segment

**In HunkSegment:**
- When hunk enters center of viewport, begin differential scrolling
- Larger side advances faster until both sides have scrolled through their entire hunk
- Scroll rates proportional to side sizes:
  - If left has 10 lines and right has 5, left scrolls 2× faster
  - Or equivalently: right "freezes" every other scroll step

### Rendering
```go
func (s *State) renderViewport(viewportHeight int) {
    leftLines := []string{}
    rightLines := []string{}

    // Start from current position and collect viewport lines
    segIdx := s.Position.SegmentIndex
    leftOff := s.Position.LeftOffset
    rightOff := s.Position.RightOffset

    for len(leftLines) < viewportHeight && segIdx < len(s.Diff.Segments) {
        seg := s.Diff.Segments[segIdx]

        switch s := seg.(type) {
        case UnchangedSegment:
            // Add lines from both sides equally
            for i := leftOff; i < s.Count && len(leftLines) < viewportHeight; i++ {
                leftLines = append(leftLines, render(s.Lines[i]))
                rightLines = append(rightLines, render(s.Lines[i]))
            }
            // Move to next segment
            segIdx++
            leftOff = 0
            rightOff = 0

        case HunkSegment:
            // Add lines independently, handling mismatched counts
            // This is where differential rendering happens
            // ...
        }
    }
}
```

### Pros
- Cleanest conceptual model: the diff IS a sequence of unchanged/changed regions
- Natural mapping to scrolling behavior
- No blank lines
- Explicit structure makes differential scrolling logic clear
- Inline diffs preserved via `LinePair`

### Cons
- Most significant refactor
- Needs careful handling of hunk rendering when sides have different line counts
- Position tracking is more complex than a simple integer

### Verdict: **Recommended** - Best balance of clarity and capability. The data structure directly models the scrolling behavior we want.

---

## Comparison Matrix

| Aspect | Option 1 (Offsets) | Option 2 (Separate) | Option 3 (Hybrid) | Option 4 (Segments) |
|--------|-------------------|---------------------|-------------------|---------------------|
| Eliminates blanks | No | Yes | Yes | Yes |
| Preserves inline diffs | Yes | No | Yes | Yes |
| Scroll logic clarity | Poor | Good | Medium | Excellent |
| Hunk detection | Complex | Simple | Simple | Built-in |
| Code complexity | Low | Medium | High | Medium |
| Conceptual clarity | Low | High | Low | High |

## Recommendation

**Option 4 (Segment-Based Model)** is recommended because:

1. **Structure matches behavior**: Segments directly represent the scrolling behavior we want - unchanged regions scroll together, hunks scroll differentially.

2. **No wasted space**: No blank lines anywhere in the data structure.

3. **Explicit hunk tracking**: No need to detect hunks at runtime - they're first-class citizens in the data model.

4. **Preserves essential info**: Inline diffs can be preserved via `LinePair` within hunks.

5. **Clear rendering logic**: Rendering walks segments sequentially, which is intuitive.

The main cost is a significant refactor of the diff building pipeline, but this is a good opportunity to simplify and clarify the codebase. The current alignment-based model was designed for blank-line padding; the new model is designed for differential scrolling.

### Simplification Opportunity

As the user mentioned, we can simplify the diff rendering for now and enhance colors later. For the initial implementation:
- Don't worry about character-level inline diffs (can add back later)
- Focus on segment structure and scrolling behavior
- Use simple added/removed styling (green/red) without word-level highlighting

This reduces the complexity of `LinePair` and lets us focus on getting the scrolling behavior right first.
