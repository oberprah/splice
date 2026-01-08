# Current Diff Structure Research

## Overview

The Splice diff viewer uses an alignment-based data model for side-by-side diff rendering. This document analyzes the current implementation to understand what needs to change for smart differential scrolling.

## Core Data Structures

### AlignedFileDiff

**File**: `internal/domain/diff/alignment.go`

```go
type AlignedFileDiff struct {
    Left       FileContent  // Old version of the file
    Right      FileContent  // New version of the file
    Alignments []Alignment  // One entry per display row
}

type FileContent struct {
    Header string   // File path/mode header
    Lines  []string // Actual file lines (no blanks)
}
```

**Key insight**: The `Alignments` slice has one entry **per display row** in the side-by-side view. When one side has more lines than the other, the shorter side gets blank lines.

### Alignment Types (Sum Type Pattern)

```go
type Alignment interface {
    alignment() // sealed interface
}

type UnchangedAlignment struct {
    LeftIdx  int   // Index into left FileContent.Lines
    RightIdx int   // Index into right FileContent.Lines
}

type ModifiedAlignment struct {
    LeftIdx    int                   // Index into left FileContent.Lines
    RightIdx   int                   // Index into right FileContent.Lines
    InlineDiff []diffmatchpatch.Diff // Character-level diff
}

type RemovedAlignment struct {
    LeftIdx int  // Index into left FileContent.Lines
    // Right side is implicitly blank
}

type AddedAlignment struct {
    RightIdx int  // Index into right FileContent.Lines
    // Left side is implicitly blank
}
```

## How Blank Lines Work

Blank lines are **not stored** in the data structure. They're **generated at render time**:

**File**: `internal/ui/states/diff/view.go`

```go
case diff.RemovedAlignment:
    leftLine := s.Diff.Left.Lines[a.LeftIdx]
    left := s.formatColumnContent(...)
    return left, ""  // Empty string for right = blank line

case diff.AddedAlignment:
    rightLine := s.Diff.Right.Lines[a.RightIdx]
    right := s.formatColumnContent(...)
    return "", right  // Empty string for left = blank line
```

The empty strings are then styled with the column width, resulting in visual blank lines that maintain alignment.

## How Alignments Are Built

**File**: `internal/domain/diff/builder.go`

The `BuildAlignments` function walks through both files:

1. **UnchangedAlignment**: Both sides have identical content at corresponding positions
2. **ModifiedAlignment**: Paired removed/added lines (using Dice similarity threshold of 0.5)
3. **RemovedAlignment**: Unpaired removed lines (no corresponding addition)
4. **AddedAlignment**: Unpaired added lines (no corresponding removal)

Within hunks, lines are paired using a greedy matching algorithm. Unpaired lines become RemovedAlignment or AddedAlignment, which render as blank on the opposite side.

## How Side-by-Side View Is Rendered

**File**: `internal/ui/states/diff/view.go`

```go
// Calculate viewport bounds
viewportEnd := min(s.ViewportStart+availableHeight, len(s.Diff.Alignments))

// Build left and right columns independently
leftVb := components.NewViewBuilder()
rightVb := components.NewViewBuilder()

// Render visible alignments
for i := s.ViewportStart; i < viewportEnd; i++ {
    alignment := s.Diff.Alignments[i]
    left, right := s.renderAlignment(alignment, columnWidth, lineNoWidth)

    leftVb.AddLine(leftColStyle.Render(left))
    rightVb.AddLine(rightColStyle.Render(right))
}

// Compose split view
vb.AddSplitView(leftVb, rightVb)
```

**Key observations**:
1. Loop through contiguous range of `Alignments` based on single `ViewportStart`
2. Each alignment renders to both left and right strings
3. Strings added to independent ViewBuilders
4. ViewBuilders composed horizontally with separator

## Pain Points for Smart Diff Scrolling

### 1. Single Viewport Position
- `ViewportStart` is shared by both panels
- No way to scroll one side independently

### 2. One-to-One Mapping
- Each alignment = exactly one display row
- Can't have different row counts on each side

### 3. No Explicit Hunk Tracking
- Hunks exist logically during parsing
- Not preserved in final `Alignments` structure
- Differential scrolling needs to know hunk boundaries

### 4. Rendering Assumption
- `AddSplitView` expects equal line counts from both ViewBuilders
- Lipgloss pads shorter side, but design assumes equal counts

## Summary

| Aspect | Current State |
|--------|---------------|
| Viewport tracking | Single `ViewportStart` for both panels |
| Line alignment | Blank lines via empty strings from RemovedAlignment/AddedAlignment |
| Hunk boundaries | Not explicitly tracked after parsing |
| Panel heights | Always equal (enforced by one-to-one alignment mapping) |

## Implications

To implement smart differential scrolling, we need to either:
1. Track separate viewport positions and render different ranges per panel
2. Fundamentally change the alignment data structure to support independent panel content
3. Add hunk boundary tracking for detecting when differential scrolling should activate
