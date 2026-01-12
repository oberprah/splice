# Research: Styling Patterns

**Date:** 2026-01-12

## Executive Summary

Splice uses Lip Gloss minimally and conservatively. The codebase focuses on **color styling only** via `Foreground()`, `Background()`, and `Bold()` methods. There are **no existing modals, boxes, borders, or padding** in the UI. All visual structure is created through plain text characters and color.

## How Lip Gloss is Used

**Current Usage:**
- `lipgloss.NewStyle()` - Create styles
- `.Foreground()` - Set text color with `lipgloss.AdaptiveColor`
- `.Background()` - Set background color
- `.Bold()` - Make text bold
- `.Render(text)` - Apply style to text
- `.Width(n)` - Set column width for split view alignment
- `lipgloss.JoinHorizontal()` - Join columns side-by-side

**NOT Used:**
- `.Border()` / `.BorderStyle()` - No box borders
- `.Padding()` / `.Margin()` - No padding/margins
- `.Height()` / `.MaxHeight()` / `.MaxWidth()` - No size constraints
- `.Align()` - No text alignment
- Any gradients, shadows, or advanced styling

**File:** `/home/user/splice/internal/ui/styles/styles.go`

## Main Style Definitions

All styles defined as package-level variables:

**Commit metadata:**
- `HashStyle` - Commit hash (amber/yellow)
- `MessageStyle` - Commit message (gray/white)
- `AuthorStyle` - Author name (cyan)
- `TimeStyle` - Relative timestamp (muted gray)
- `HeaderStyle` - Metadata separators (subtle gray)

**Selection states:**
- `SelectedHashStyle`, `SelectedMessageStyle`, etc. - Bold variants

**File diff statistics:**
- `AdditionsStyle` - Green (+N additions)
- `DeletionsStyle` - Red (-N deletions)
- `FilePathStyle` - File names
- `SelectedFilePathStyle` - Bold file names

**Diff line backgrounds:**
- `DiffAdditionsStyle` - Subtle green for added lines
- `DiffDeletionsStyle` - Subtle red for deleted lines
- Bright variants for inline highlights

## Color Scheme Strategy

**AdaptiveColor** for light/dark terminals:
```go
Foreground: lipgloss.AdaptiveColor{
    Light: "172",  // Darker for light terminals
    Dark:  "214",  // Brighter for dark terminals
}
```

Uses xterm 256-color palette (0-255). No hex colors.

## How Components Get Styles

**Pattern 1: Global Access**
```go
import "github.com/oberprah/splice/internal/ui/styles"
line.WriteString(styles.HashStyle.Render(shortHash))
```

**Pattern 2: Passed as Parameters**
```go
func formatColumnContent(
    lineNo int,
    indicator string,
    tokens []highlight.Token,
    bgStyle lipgloss.Style,  // Style passed in
    ...
) string
```

Styles constructed once at package init and reused.

## Existing Modal/Overlay Styles

**None exist.** Current transient screens (LoadingState, ErrorState) are minimal text with no decoration.

## Border and Box Rendering

**Horizontal separator:**
```go
separator := strings.Repeat("─", min(ctx.Width(), 80))
vb.AddLine(styles.HeaderStyle.Render(separator))
```

**Vertical separator:**
```go
separatorLines[i] = " │ "  // Box-drawing character
```

**Box-drawing characters available:**
- `─` `│` - Lines
- `┌` `┐` `└` `┘` - Corners
- `├` `┤` `┬` `┴` `┼` - Junctions

**NOT currently used** but available for modal design.

## Manual Layout Implementation

The codebase manually:
1. Calculates available space
2. Builds strings line-by-line with ViewBuilder
3. Applies width styling per-line
4. Uses `JoinHorizontal()` for side-by-side composition

## Pattern Recommendations for Help Overlay

**Styling Approach:**
1. Use existing color palette - Add `HelpStyle`, `HelpBorderStyle`, `HelpKeyStyle` to styles.go
2. Text-based box using box-drawing characters
3. Manual centering: `(width - contentWidth) / 2`
4. Cap width at ~60 chars for readability

**Visual Options:**

Simple box:
```
┌──────────────────────────────────────┐
│  COMMANDS                            │
│                                      │
│  j/k - scroll up/down                │
│  g/G - first/last commit             │
│  ?   - toggle help                   │
└──────────────────────────────────────┘
```

Minimal (no box):
```
 COMMANDS

 j/k - scroll up/down
 g/G - first/last commit
 ?   - toggle help
```

## ViewBuilder Pattern

All views use `ViewBuilder`:
- Adds lines: `AddLine(string)`
- Joins with `\n`
- Supports split views: `AddSplitView(left, right)`

Help overlay should use ViewBuilder for consistency.

## Recommended Colors for Help

| Element | Light | Dark | Usage |
|---------|-------|------|-------|
| Title/Key | `220` | `220` | Key bindings (like HashStyle) |
| Description | `252` | `252` | Command descriptions |
| Border | `243` | `248` | Box border (like HeaderStyle) |

## Summary

Splice takes a **minimalist, text-based approach** to UI rendering. The help overlay should follow the same philosophy: use existing color palette, build with text characters for structure, implement sizing/centering manually.
