# Design: Add Split View Support to ViewBuilder

## Executive Summary

Multiple views (log, diff) need split-panel layouts with a vertical separator. Currently, each view implements this independently, leading to duplication and inconsistent patterns.

We'll add `AddSplitView(left, right *ViewBuilder)` to ViewBuilder, providing a clean, reusable abstraction for split layouts. Callers build left and right content independently using ViewBuilder, then compose them. This eliminates duplication, ensures consistent split rendering across views, and maintains clear separation of concerns (caller formats, ViewBuilder joins).

## Context & Problem Statement

**The Problem:** Log view and diff view both implement split-panel layouts. The implementation is duplicated across views, making it harder to maintain and evolve. Each view has its own row-by-row joining logic, separator handling, and alignment code.

When we need to:
- Add a new view with split layout
- Change separator appearance
- Adjust alignment behavior

We must update multiple places. This violates DRY and makes the codebase harder to maintain.

**Scope:** This design covers adding split view support to ViewBuilder as a reusable abstraction. It does not cover the content formatting within each column (that remains each view's responsibility).

## Current State

Each view implements split rendering independently:

```go
// log_view.go - Implements split layout with row-by-row joining
func (s LogState) renderSplitView(ctx Context) *ViewBuilder {
    vb := NewViewBuilder()
    for i := 0; i < ctx.Height(); i++ {
        logLine := s.formatCommitLine(...)
        detailLine := detailsLines[i]

        row := lipgloss.JoinHorizontal(
            lipgloss.Top,
            logColStyle.Render(logLine),
            separatorStyle.Render(" │ "),
            detailsColStyle.Render(detailLine),
        )

        for _, line := range strings.Split(row, "\n") {
            vb.AddLine(line)
        }
    }
    return vb
}

// diff_view.go - Similar pattern duplicated
// files_view.go - Could benefit from split view but doesn't use it
```

**Issues with current approach:**
- Logic duplicated across multiple views
- Row-by-row joining is complex and error-prone
- No consistent abstraction for split layouts
- Adding new split views requires reimplementing the pattern

## Solution

### Core Insight from Lipgloss Experiments

The separator `" │ "` is not a special concept - it's just another column. When all columns are multi-line with matching line counts, JoinHorizontal joins them line-by-line:

```go
left := "Line 1\nLine 2\nLine 3"
separator := " │ \n │ \n │ "  // Multi-line matches left/right
right := "Line 1\nLine 2\nLine 3"

lipgloss.JoinHorizontal(lipgloss.Top, left, separator, right)
// Output:
// Line 1 │ Line 1
// Line 2 │ Line 2
// Line 3 │ Line 3
```

Additionally, JoinHorizontal **automatically pads shorter lines** to align columns, even without explicit width styles.

### API Design

Add a method to ViewBuilder:

```go
// AddSplitView joins two ViewBuilders horizontally with a vertical separator
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder)
```

> **Decision:** Use `*ViewBuilder` parameters instead of `[]string`.
>
> **Rationale:** Composition of builders is more flexible and idiomatic. Callers can build left/right content using the same ViewBuilder API, then compose them. This allows complex content (wrapped text, formatted sections) to be built naturally.

> **Decision:** No width parameters in the API.
>
> **Rationale:** Width/wrapping is the caller's responsibility. The caller knows `ctx.Width()`, terminal constraints, and content requirements. ViewBuilder's job is mechanical joining, not layout decisions. Callers can apply lipgloss width styles before creating ViewBuilders if needed.

### How It Works

The key insight: treat the separator as a multi-line column that matches the tallest content column.

`AddSplitView` takes two ViewBuilders as input. Each ViewBuilder contains lines of content. The method:

1. Converts each ViewBuilder's lines to multi-line strings
2. Determines the maximum line count between the two columns
3. Builds a separator string with that many lines (each line being `" │ "`)
4. Passes all three multi-line strings (left, separator, right) to `lipgloss.JoinHorizontal`
5. Takes the joined result and adds its lines to the parent ViewBuilder

Lipgloss handles the horizontal joining and automatically aligns columns by padding shorter lines.

### Usage Pattern

Views using split layout follow a simple three-step pattern:

1. **Build left content** - Create a ViewBuilder and add left column lines
2. **Build right content** - Create another ViewBuilder and add right column lines
3. **Compose** - Call `AddSplitView(left, right)` on the parent ViewBuilder

Example from log view:
```go
vb := NewViewBuilder()

// Build columns independently
leftVb := buildCommitList(...)
rightVb := buildDetailsPanel(...)

// Compose with separator
vb.AddSplitView(leftVb, rightVb)
```

The caller retains full control over content formatting, while ViewBuilder handles the layout mechanics.

### Tradeoffs

**Pros:**
- Eliminates duplication across views
- Clean separation of concerns: caller formats content, ViewBuilder handles layout
- Lipgloss handles alignment automatically (no manual padding needed)
- Composable API enables mixing split and full-width content
- Easier to test (can test split logic independently)

**Cons:**
- Slightly more verbose: must build two ViewBuilders before composing
- However, this verbosity brings clarity - the two-step process (build content, then compose) is easier to understand and maintain than inline row-by-row loops

## Open Questions

None. The approach is validated by lipgloss experiments and provides a clean, composable abstraction for split views.
