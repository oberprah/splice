# ViewBuilder and Column Rendering Patterns

This document describes the ViewBuilder architecture and patterns used for rendering columns in Splice.

## ViewBuilder Architecture Overview

### Core Concept
ViewBuilder (`internal/ui/states/viewbuilder.go`) is a utility for building terminal UI views with automatic newline handling. It ensures views never end with trailing newlines by storing lines separately and joining them only when `String()` is called.

### Key Features
- **Line-based building**: Uses `AddLine(line string)` to add content line-by-line
- **Automatic newline joining**: Lines are joined with `"\n"` only when `String()` is called
- **Newline escaping**: Embedded newlines in content are automatically escaped to `\n` to maintain single-line structure
- **Split view support**: `AddSplitView(left, right *ViewBuilder)` enables side-by-side layout with vertical separator

### API Methods

```go
// Create new builder
func NewViewBuilder() *ViewBuilder

// Add a line to the view
func (vb *ViewBuilder) AddLine(line string)

// Get final output without trailing newline
func (vb *ViewBuilder) String() string

// Join two ViewBuilders horizontally with separator
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder)
```

## Column System and Width Management

### Split View Architecture

The split view feature (added in `260104-viewbuilder-split-view`) provides a clean abstraction for side-by-side layouts:

**Key Insight**: The separator `" │ "` is treated as a multi-line column, not a special concept. When all columns are multi-line with matching line counts, `lipgloss.JoinHorizontal` joins them line-by-line automatically.

**Implementation Pattern** (3 steps):
1. **Build left content** - Create a ViewBuilder and add left column lines
2. **Build right content** - Create another ViewBuilder and add right column lines
3. **Compose** - Call `AddSplitView(left, right)` on the parent ViewBuilder

### How AddSplitView Works

From `internal/ui/states/viewbuilder.go:40-71`:

1. Converts each ViewBuilder's lines to multi-line strings
2. Determines the maximum line count between columns
3. Builds a separator string with that many lines (each line: `" │ "`)
4. Passes all three multi-line strings to `lipgloss.JoinHorizontal(lipgloss.Top, left, separator, right)`
5. Adds the joined result's lines to the parent ViewBuilder

**Automatic Alignment**: Lipgloss automatically pads shorter columns to align properly, even without explicit width styles.

### Width Calculation Patterns

Views calculate column widths based on terminal width:

**Log View** (`log_view.go:48-61`):
```go
const (
    splitPanelWidth = 80   // Fixed width for details panel
    splitThreshold  = 160  // Minimum terminal width for split view
    separatorWidth  = 3    // Width of " │ " separator
)

// Calculate widths
logWidth := ctx.Width() - splitPanelWidth - separatorWidth
detailsWidth := splitPanelWidth
```

**Diff View** (`diff_view.go:40-44`):
```go
// Each column gets half the terminal width minus separator
columnWidth := (ctx.Width() - 3) / 2  // -3 for " │ " separator
if columnWidth < 20 {
    columnWidth = 20  // Minimum fallback
}
```

## Relevant APIs for Measuring/Constraining Content

### Width Application with Lipgloss

**Fixed-Width Column Styling**: Apply width constraints using `lipgloss.NewStyle().Width(width)`:

```go
// Log view example (log_view.go:68)
colStyle := lipgloss.NewStyle().Width(width)
vb.AddLine(colStyle.Render(line))
```

**Width Rendering**: The style can also be applied after content is assembled:

```go
// Diff view example (diff_view.go:217)
return bgStyle.Width(columnWidth).Render(columnStr)
```

### ANSI-Aware Text Measurement

The codebase has access to ANSI-aware text operations via dependencies:
- `github.com/charmbracelet/x/ansi` - Provides `StringWidth()` for ANSI-aware width calculation
- `github.com/mattn/go-runewidth` - Handles wide characters (CJK, emojis)
- `github.com/muesli/ansi` - Low-level ANSI sequence handling

**Key Functions** (from `docs/specs/251224-syntax-highlighting/research/ansi-text-operations.md`):
- `ansi.StringWidth(s)` - Calculate visible width ignoring escape sequences
- `ansi.Truncate(s, maxWidth, tail)` - Truncate to N visible chars, preserving ANSI codes
- `lipgloss.Width()` / `lipgloss.Height()` - Measure text accounting for ANSI codes

**Important**: Use `ansi.StringWidth()` instead of `len(string)` for styled text, as ANSI escape codes inflate byte length.

### Current Width Handling in formatCommitLine

The log view currently uses `len()` for width calculations (`log_view.go:135-220`):

```go
// Calculate required space for fixed elements
fixedWidth := len(selectionIndicator) + len(graphSymbols) + len(hash) + 1 +
              len(refsStr) + len(separator) + len(timePrefix) + len(time)

// Calculate remaining space for message and author
remainingWidth := max(availableWidth-fixedWidth, 10)

// Truncate message and author to fit
if len(message) > messageMaxWidth && messageMaxWidth > 3 {
    message = message[:messageMaxWidth-3] + "..."
}
```

**Note**: This pattern uses simple `len()` and string slicing, which works for unstyled text but would break with ANSI codes.

## Examples from Existing Code

### Example 1: Log View Split Columns

From `internal/ui/states/log_view.go:63-109`:

**Independent Column Builders**:
```go
func (s LogState) buildCommitListColumn(width int, ctx Context) *ViewBuilder {
    vb := NewViewBuilder()
    colStyle := lipgloss.NewStyle().Width(width)

    viewportEnd := min(s.ViewportStart+ctx.Height(), len(s.Commits))

    // Build column with viewport height
    for i := 0; i < ctx.Height(); i++ {
        var line string
        logIdx := s.ViewportStart + i
        if logIdx < viewportEnd && logIdx < len(s.Commits) {
            commit := s.Commits[logIdx]
            line = s.formatCommitLine(commit, logIdx, logIdx == s.Cursor, width, ctx)
        }
        vb.AddLine(colStyle.Render(line))  // Apply fixed width
    }

    return vb
}
```

**Key Pattern**:
- Create style with target width
- Build all lines for the column
- Apply width styling to each line before adding to ViewBuilder
- Return independent ViewBuilder for composition

### Example 2: Diff View Content Width Calculation

From `internal/ui/states/diff_view.go:140-218`:

```go
func (s *DiffState) formatColumnContent(lineNo int, indicator string, tokens []highlight.Token,
                                        lineNoWidth, contentWidth, columnWidth int,
                                        bgStyle lipgloss.Style, inlineDiff []diffmatchpatch.Diff) string {
    // Format line number
    lineNoStr := fmt.Sprintf("%*d", lineNoWidth, lineNo)

    // Render tokens with syntax highlighting
    renderedContent := s.renderTokens(tokens, contentWidth, bgStyle)

    // Build column: "123 - content"
    styledLineNo := bgStyle.Render(lineNoStr)
    styledIndicator := bgStyle.Render(" " + indicator + " ")
    columnStr := styledLineNo + styledIndicator + renderedContent

    // Apply full width styling to pad the column
    return bgStyle.Width(columnWidth).Render(columnStr)
}
```

**Key Pattern**:
- Calculate `contentWidth = columnWidth - lineNoWidth - 4` (accounting for fixed elements)
- Render content constrained to contentWidth
- Apply full columnWidth styling at the end for padding

### Example 3: Token Rendering with Width Constraint

From `internal/ui/states/diff_view.go:220-255`:

```go
func (s *DiffState) renderTokens(tokens []highlight.Token, maxWidth int, bgStyle lipgloss.Style) string {
    var result strings.Builder
    visibleWidth := 0

    for _, token := range tokens {
        expandedValue := expandTabs(token.Value, 4)

        for _, r := range expandedValue {
            if visibleWidth >= maxWidth {
                if visibleWidth == maxWidth {
                    result.WriteString("…")  // Append ellipsis
                }
                return result.String()
            }

            // Apply syntax highlighting + background
            syntaxStyle := highlight.StyleForToken(token.Type)
            combinedStyle := syntaxStyle.Inherit(bgStyle)
            result.WriteString(combinedStyle.Render(string(r)))
            visibleWidth++
        }
    }

    return result.String()
}
```

**Key Pattern**:
- Track `visibleWidth` as characters are rendered
- Iterate rune-by-rune (handles multi-byte characters correctly)
- Apply styling per-character for precise control
- Truncate with ellipsis when reaching maxWidth

## Context Interface for Width Access

From `internal/ui/states/state.go:16-23`:

```go
type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}
```

All views receive a `Context` with `Width()` and `Height()` methods, providing terminal dimensions.

## Key Patterns for Log Line Truncation

Based on the existing patterns, implementing log line truncation should:

1. **Pre-calculate all component widths** before assembling the line
2. **Use fixed-width styling** via lipgloss when exact width is required
3. **Measure styled text** with `ansi.StringWidth()` if ANSI codes are present
4. **Truncate progressively** based on priority (following requirements)
5. **Track visible width** during assembly, not byte length
6. **Handle refs specially** with balanced parentheses and graceful degradation
7. **Apply styling after truncation** to avoid measuring styled text unnecessarily

### Pattern: Calculate → Build → Style → Compose

This is the established pattern seen in both log and diff views:
1. **Calculate** widths for all components
2. **Build** content strings (with truncation if needed)
3. **Style** the assembled content
4. **Compose** using ViewBuilder
