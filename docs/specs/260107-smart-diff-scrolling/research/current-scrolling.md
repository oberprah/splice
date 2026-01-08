# Current Scrolling and Viewport Implementation Research

## Overview

The Splice diff viewer uses a **single unified viewport** for side-by-side diff. Both panels scroll together, maintaining alignment through blank line padding. This document analyzes the current implementation.

## Scroll Position Tracking

### State Storage

**File**: `internal/ui/states/diff/state.go`

```go
type State struct {
    // Diff data
    CommitRange core.CommitRange
    File        core.FileChange
    Diff        *diff.AlignedFileDiff

    // Viewport control
    ViewportStart    int       // Single position (0-indexed)
    CurrentChangeIdx int       // Index into ChangeIndices
    ChangeIndices    []int     // Alignment indices with changes
}
```

**Key observation**: `ViewportStart` is a single integer. Both panels always render the same range of alignments.

### Viewport Initialization

```go
func New(...) *State {
    viewportStart := 0
    if d != nil && len(changeIndices) > 0 {
        viewportStart = changeIndices[0]  // Start at first change
    }
    return &State{
        ViewportStart: viewportStart,
        // ...
    }
}
```

## How Viewport Slicing Works

### Rendering Process

**File**: `internal/ui/states/diff/view.go`

```go
// Calculate available height
headerLines := strings.Count(header, "\n") + 1
availableHeight := max(ctx.Height()-headerLines, 1)

// Calculate viewport bounds
viewportEnd := min(s.ViewportStart+availableHeight, len(s.Diff.Alignments))

// Build left and right columns
leftVb := components.NewViewBuilder()
rightVb := components.NewViewBuilder()

for i := s.ViewportStart; i < viewportEnd; i++ {
    alignment := s.Diff.Alignments[i]
    left, right := s.renderAlignment(alignment, columnWidth, lineNoWidth)

    leftVb.AddLine(leftColStyle.Render(left))
    rightVb.AddLine(rightColStyle.Render(right))
}

vb.AddSplitView(leftVb, rightVb)
```

**Key observations**:
1. Single loop iterating `ViewportStart` to `viewportEnd`
2. Same alignment range used for both panels
3. Each alignment produces one line per panel

### Split View Composition

**File**: `internal/ui/components/viewbuilder.go`

```go
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder) {
    // Determine max line count
    maxLines := max(len(left.lines), len(right.lines))

    // Build separator
    separatorLines := make([]string, maxLines)
    for i := 0; i < maxLines; i++ {
        separatorLines[i] = " │ "
    }

    // Join horizontally with lipgloss
    joined := lipgloss.JoinHorizontal(lipgloss.Top, leftStr, separatorStr, rightStr)

    for _, line := range strings.Split(joined, "\n") {
        vb.AddLine(line)
    }
}
```

Lipgloss's `JoinHorizontal` with `lipgloss.Top` pads the shorter column automatically.

## Scrolling Update Logic

**File**: `internal/ui/states/diff/update.go`

```go
case "j", "down":
    maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
    if s.ViewportStart < maxViewportStart {
        s.ViewportStart++
    }

case "ctrl+d":  // Page down
    availableHeight := max(ctx.Height()-headerLines, 1)
    halfPage := availableHeight / 2
    maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
    s.ViewportStart = min(s.ViewportStart+halfPage, maxViewportStart)

case "n":  // Jump to next change
    for i, changeIdx := range s.ChangeIndices {
        if changeIdx > s.ViewportStart {
            s.CurrentChangeIdx = i
            s.ViewportStart = changeIdx
            return
        }
    }
```

**Characteristics**:
- Both panels scroll in lockstep
- No differential scrolling logic
- `ChangeIndices` enables quick navigation to changes

### Viewport Bounds

```go
func (s *State) calculateMaxViewportStart(height int) int {
    headerLines := 2
    availableHeight := max(height-headerLines, 1)

    maxStart := len(s.Diff.Alignments) - availableHeight
    if maxStart < 0 {
        maxStart = 0
    }
    return maxStart
}
```

## Summary

| Aspect | Current Implementation |
|--------|----------------------|
| Scroll position | Single `ViewportStart` integer |
| Viewport range | `[ViewportStart, ViewportStart+availableHeight)` |
| Panel coupling | Locked together (same alignment range) |
| Navigation | j/k (line), ctrl+d/u (page), n/N (change) |
| Bounds checking | Single `maxViewportStart` calculation |

## What Needs to Change

1. **Separate viewport positions**: `LeftViewportStart` and `RightViewportStart`
2. **Independent bounds**: Different max positions based on panel content length
3. **Differential scrolling logic**: Detect hunks at center, adjust scroll rates
4. **Hunk detection**: Track when hunks enter/exit center region for triggering differential behavior
