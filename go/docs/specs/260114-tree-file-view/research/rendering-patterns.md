# Rendering Patterns in Splice

This document describes how Splice renders complex UI components, handles text formatting, and manages styling patterns. This research serves as a foundation for implementing the tree-based file view.

## Table of Contents

1. [ViewBuilder Pattern](#viewbuilder-pattern)
2. [Styling with Lipgloss](#styling-with-lipgloss)
3. [Selection Highlighting](#selection-highlighting)
4. [Text Width Calculations and Truncation](#text-width-calculations-and-truncation)
5. [Tree-Like Rendering: Git Graph](#tree-like-rendering-git-graph)
6. [Key Takeaways for Tree File View](#key-takeaways-for-tree-file-view)

## ViewBuilder Pattern

### Core Concept

`ViewBuilder` (`internal/ui/components/viewbuilder.go`) is the fundamental building block for all views in Splice. It provides a line-based abstraction for building terminal UI views with automatic newline handling.

### Key Features

```go
type ViewBuilder struct {
    lines []string
}

// Core API
func NewViewBuilder() *ViewBuilder
func (vb *ViewBuilder) AddLine(line string)
func (vb *ViewBuilder) String() string
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder)
```

**Critical Behavior**:
- Stores lines separately, joins with `"\n"` only when `String()` is called
- **Never ends with trailing newline** (enforced by `strings.Join`)
- Embedded newlines in content are **escaped** to `\n` to maintain single-line structure
- Implements `core.ViewRenderer` interface for type safety

### Split View Architecture

The `AddSplitView` method enables side-by-side layouts:

```go
func (vb *ViewBuilder) AddSplitView(left *ViewBuilder, right *ViewBuilder) {
    leftStr := left.String()
    rightStr := right.String()

    // Build separator as multi-line string
    maxLines := max(len(left.lines), len(right.lines))
    separatorLines := make([]string, maxLines)
    for i := 0; i < maxLines; i++ {
        separatorLines[i] = " │ "
    }
    separatorStr := strings.Join(separatorLines, "\n")

    // Join horizontally
    joined := lipgloss.JoinHorizontal(lipgloss.Top, leftStr, separatorStr, rightStr)

    // Add to parent
    for _, line := range strings.Split(joined, "\n") {
        vb.AddLine(line)
    }
}
```

**Key Insight**: The separator `" │ "` is treated as a multi-line column, not a special concept. When all columns are multi-line with matching line counts, `lipgloss.JoinHorizontal` joins them line-by-line automatically and pads shorter columns.

### Usage Pattern: Build → Compose

**Example from Log View** (`internal/ui/states/log/view.go:52-66`):

```go
func (s State) renderSplitView(ctx core.Context) core.ViewRenderer {
    // Calculate widths
    logWidth := ctx.Width() - splitPanelWidth - separatorWidth
    previewWidth := splitPanelWidth

    // Build columns independently
    leftVb := s.buildCommitListColumn(logWidth, ctx).(*components.ViewBuilder)
    rightVb := s.buildFilesPreviewColumn(previewWidth, ctx).(*components.ViewBuilder)

    // Compose
    vb := components.NewViewBuilder()
    vb.AddSplitView(leftVb, rightVb)
    return vb
}
```

**Pattern Steps**:
1. Calculate column widths based on terminal dimensions
2. Build each column independently in its own ViewBuilder
3. Compose using `AddSplitView`

**Important**: Each column builder must produce exactly `ctx.Height()` lines for proper alignment. Missing lines are padded by lipgloss automatically.

## Styling with Lipgloss

### Style Definition Pattern

All styles are defined centrally in `internal/ui/styles/styles.go` using `lipgloss.NewStyle()`:

```go
// Adaptive colors (different for light/dark terminals)
HashStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
    Light: "172",  // darker amber
    Dark:  "214",  // bright amber
})

MessageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
    Light: "237",  // dark gray
    Dark:  "252",  // bright white
})

// Selected variants (bold + brighter)
SelectedHashStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
    Light: "208",  // bright orange
    Dark:  "220",  // brighter yellow
}).Bold(true)
```

**Design Principles**:
- **Adaptive colors**: Define both light and dark terminal variants
- **Semantic naming**: `HashStyle`, `MessageStyle`, `AuthorStyle` (not `OrangeStyle`, `GrayStyle`)
- **Selection convention**: Selected styles are bold + brighter color variants
- **Background styles**: Used for diff additions/deletions with very subtle colors

### Style Application Patterns

#### Pattern 1: Apply Style to Component

Most common pattern - style individual components before assembly:

```go
// From log_line_format.go:293-315
line.WriteString(hashStyle.Render(hash))
line.WriteString(" ")
line.WriteString(messageStyle.Render(message))
line.WriteString(" - ")
line.WriteString(authorStyle.Render(author))
```

#### Pattern 2: Fixed-Width Styling

Apply width constraints using `lipgloss.NewStyle().Width(width)`:

```go
// From log/view.go:73
colStyle := lipgloss.NewStyle().Width(width)
vb.AddLine(colStyle.Render(line))
```

This ensures content is exactly N characters wide, padding with spaces if needed.

#### Pattern 3: Combined Style Inheritance

Combine multiple styles (used in diff view for syntax highlighting + background):

```go
// From diff view
syntaxStyle := highlight.StyleForToken(token.Type)
combinedStyle := syntaxStyle.Inherit(bgStyle)
result.WriteString(combinedStyle.Render(string(r)))
```

### Width Calculation Patterns

**Log View** (`internal/ui/states/log/view.go:14-16`):
```go
const (
    splitPanelWidth = 80   // Fixed width for files preview panel
    splitThreshold  = 160  // Minimum terminal width for split view
    separatorWidth  = 3    // Width of " │ " separator
)

logWidth := ctx.Width() - splitPanelWidth - separatorWidth
```

**Diff View**:
```go
columnWidth := (ctx.Width() - 3) / 2  // Each column gets half minus separator
if columnWidth < 20 {
    columnWidth = 20  // Minimum fallback
}
```

## Selection Highlighting

### LineDisplayState Enum

Selection state is modeled as an enum (`internal/ui/components/line_display_state.go`):

```go
type LineDisplayState int

const (
    LineStateNone         LineDisplayState = iota  // Not selected, not cursor
    LineStateCursor                                // Normal mode cursor (→)
    LineStateSelected                              // Visual mode selected (▌)
    LineStateVisualCursor                          // Visual mode cursor (█)
)

func (s LineDisplayState) SelectorString() string {
    switch s {
    case LineStateNone:        return "  "
    case LineStateCursor:      return "→ "
    case LineStateSelected:    return "▌ "
    case LineStateVisualCursor: return "█ "
    }
}
```

**Design Insight**: The enum encapsulates both the state and its visual representation. This follows the principle "make illegal states unrepresentable" - you can't have a cursor without a selector string.

### Selection Pattern in Components

**File Line Component** (`internal/ui/components/file_section.go:112-188`):

```go
func FormatFileLine(params FormatFileLineParams) string {
    var line strings.Builder

    // Selection indicator
    if params.ShowSelector {
        if params.IsSelected {
            line.WriteString("→")
        } else {
            line.WriteString(" ")
        }
    }

    // Choose styles based on selection
    if params.IsSelected {
        line.WriteString(styles.SelectedHashStyle.Render(status))
        line.WriteString(styles.SelectedAdditionsStyle.Render(addStr))
        line.WriteString(styles.SelectedFilePathStyle.Render(path))
    } else {
        line.WriteString(statusStyle.Render(status))
        line.WriteString(styles.AdditionsStyle.Render(addStr))
        line.WriteString(styles.FilePathStyle.Render(path))
    }

    return line.String()
}
```

**Pattern Steps**:
1. Determine selection state (usually via boolean parameter)
2. Add selector prefix if enabled
3. Choose appropriate style variant (normal vs selected)
4. Apply styles and assemble line

**Key Principle**: Components receive selection state as input, not current state. This makes them pure functions.

### Cursor State Management

State implementations compute display state for each line:

```go
// From log/view.go:135-156
func (s State) getLineDisplayState(index int) components.LineDisplayState {
    pos := s.CursorPosition()

    switch cursor := s.Cursor.(type) {
    case core.CursorNormal:
        if index == pos {
            return components.LineStateCursor
        }
        return components.LineStateNone
    case core.CursorVisual:
        if index == pos {
            return components.LineStateVisualCursor
        }
        if core.IsInSelection(cursor, index) {
            return components.LineStateSelected
        }
        return components.LineStateNone
    }
}
```

## Text Width Calculations and Truncation

### Unicode-Aware Width Calculation

Splice uses `unicode/utf8` for rune counting (handles multi-byte characters):

```go
// From log_line_format.go:245-265
func measureLineWidth(selector, graph, hash, refs, message, author, time string) int {
    width := utf8.RuneCountInString(selector) + utf8.RuneCountInString(graph) + utf8.RuneCountInString(hash)

    if refs != "" {
        width += 1 + utf8.RuneCountInString(refs)
    } else {
        width += 1
    }

    width += utf8.RuneCountInString(message)

    if author != "" {
        width += 3 + utf8.RuneCountInString(author)  // " - " + author
    }

    if time != "" {
        width += 1 + utf8.RuneCountInString(time)
    }

    return width
}
```

**Important**: Use `utf8.RuneCountInString()` instead of `len()` for user-visible text. `len()` returns byte count, which is wrong for multi-byte UTF-8 characters.

### Truncation Pattern: Progressive Truncation

The log line formatting implements a 10-level progressive truncation strategy (`internal/ui/components/log_line_format.go:320-397`):

```go
func FormatCommitLine(components CommitLineComponents, availableWidth int) string {
    // Build initial components
    selector := components.DisplayState.SelectorString()
    graph := components.Graph
    hash := components.Hash
    message := components.Message
    author := components.Author
    time := components.Time
    refs := buildRefs(components.Refs, RefsLevelFull)

    // Apply truncation levels sequentially until line fits
    level := 0
    for measureLineWidth(selector, graph, hash, refs, message, author, time) > availableWidth && level < 10 {
        switch level {
        case 0: message = capMessage(message, 72)
        case 1: author = truncateAuthor(author, 25)
        case 2: refs = buildRefs(components.Refs, RefsLevelShortenIndividual)
        case 3: refs = buildRefs(components.Refs, RefsLevelFirstPlusCount)
        case 4: refs = buildRefs(components.Refs, RefsLevelCountOnly)
        case 5: author = truncateAuthor(author, 5)
        case 6: time = ""
        case 7: message = capMessage(message, 40)
        case 8: author = ""
        case 9:
            refs = ""
            // Final truncation: measure what we'd have and truncate message to fit
            plainLine := selector + graph + hash + " " + message
            visualWidth := utf8.RuneCountInString(plainLine)
            if visualWidth > availableWidth {
                excess := visualWidth - availableWidth
                targetMsgLen := max(utf8.RuneCountInString(message) - excess, 5)
                message = capMessage(message, targetMsgLen)
            }
        }
        level++
    }

    // Assemble and style the line
    return assembleLine(selector, graph, hash, refs, message, author, time, components.DisplayState)
}
```

**Key Principles**:
1. **Measure first**: Calculate total width before styling
2. **Truncate plain text**: Work with unstyled strings for accurate measurement
3. **Progressive degradation**: Remove/shorten components in priority order
4. **Preserve minimums**: Ensure at least 5 chars for critical components
5. **Style last**: Apply styling after all truncation decisions

### Truncation Helper Functions

```go
// Truncate with ellipsis
func capMessage(message string, maxLen int) string {
    if utf8.RuneCountInString(message) <= maxLen {
        return message
    }
    if maxLen < 1 {
        return ""
    }
    // Convert to runes to properly truncate multi-byte characters
    runes := []rune(message)
    return string(runes[:maxLen-1]) + "…"
}
```

**Pattern**: Convert to `[]rune` slice, truncate, add ellipsis. This handles multi-byte UTF-8 correctly.

### Refs Truncation Levels

Refs have special 4-level truncation:

```go
type RefsLevel int

const (
    RefsLevelFull              RefsLevel = iota  // (HEAD -> main, tag: v1.0)
    RefsLevelShortenIndividual                   // (HEAD -> mai…, tag: v1…)
    RefsLevelFirstPlusCount                      // (main +2 more)
    RefsLevelCountOnly                           // (3 refs)
)
```

**Design**: Gracefully degrade from full display to just a count while maintaining readable context.

## Tree-Like Rendering: Git Graph

### Graph Architecture Overview

The git graph rendering (`internal/domain/graph/`) demonstrates complex tree-like structure rendering. It's organized into separate concerns:

**Files**:
- `types.go` - Data structures (Commit, Layout, Row, GraphSymbol enum)
- `layout.go` - Main algorithm (ComputeLayout)
- `lanes.go` - Lane management (assignColumn, updateLanes, collapseTrailingEmpty)
- `generate.go` - Symbol generation (generateRowSymbols, detect functions)
- `symbols.go` - Symbol rendering (String() method, RenderRow)

### Data Structures

```go
// types.go
type GraphSymbol int

const (
    SymbolEmpty        GraphSymbol = iota  // "  "
    SymbolBranchPass                       // "│ "
    SymbolBranchCross                      // "│─"
    SymbolCommit                           // "├ "
    SymbolMergeCommit                      // "├─"
    SymbolBranchTop                        // "╮ "
    SymbolBranchBottom                     // "╯ "
    SymbolMergeJoin                        // "┤ "
    SymbolOctopus                          // "┬─"
    SymbolDiverge                          // "┴─"
    SymbolMergeCross                       // "┼─"
)

type Row struct {
    Symbols []GraphSymbol
}

type Layout struct {
    Rows []Row
}
```

**Design Principles**:
- **Each symbol is exactly 2 characters** (enables predictable width)
- **GraphSymbol is an enum**, not strings (type safety)
- **Rows are symbol arrays**, allowing column-based operations
- **Layout is pure data**, no rendering logic

### Symbol Rendering

```go
// symbols.go
var symbolStrings = map[GraphSymbol]string{
    SymbolEmpty:        "  ",
    SymbolBranchPass:   "│ ",
    SymbolCommit:       "├ ",
    // ... etc
}

func (s GraphSymbol) String() string {
    if str, ok := symbolStrings[s]; ok {
        return str
    }
    return "  "  // Default to empty
}

func RenderRow(row Row) string {
    result := ""
    for _, sym := range row.Symbols {
        result += sym.String()
    }
    return result
}
```

**Pattern**: Enum values map to display strings via a lookup table. Rendering is a simple traversal.

### Layout Algorithm

The layout algorithm (`layout.go:11-89`) processes commits in display order:

```go
func ComputeLayout(commits []Commit) *Layout {
    var rows []Row
    var lanes []string  // Active lanes (column index → commit hash)

    for _, commit := range commits {
        // 1. Assign column for this commit
        col, lanes := assignColumn(commit.Hash, lanes)

        // 2. Detect converging columns (branches merging)
        convergingColumns := detectConvergingColumns(col, commit.Hash, lanes)

        // 3. Clear converging columns
        for _, convergingCol := range convergingColumns {
            lanes[convergingCol] = ""
        }

        // 4. Update lanes with parent information
        updateResult := updateLanes(col, commit.Parents, lanes)
        lanes = updateResult.Lanes

        // 5. Detect passing columns
        passingColumns := detectPassingColumns(col, lanes, updateResult.MergeColumns, convergingColumns)

        // 6. Generate symbols for this row
        row := generateRowSymbols(col, len(lanes), updateResult.MergeColumns,
                                   convergingColumns, passingColumns,
                                   updateResult.ExistingLanesMerge,
                                   updateResult.ConvergesToParent)
        rows = append(rows, row)

        // 7. Collapse trailing empty lanes
        lanes = collapseTrailingEmpty(lanes)
    }

    return &Layout{Rows: rows}
}
```

**Key Concepts**:

1. **Lanes array**: `lanes[i]` contains the commit hash expected at column `i` in the next row
2. **Column assignment**: Find existing lane or allocate new column
3. **Convergence detection**: Multiple lanes pointing to same commit hash
4. **Lane update**: Replace commit with its first parent, allocate columns for merge parents
5. **Symbol generation**: Choose box-drawing character based on column roles

### Lane Management

```go
// lanes.go
func assignColumn(hash string, lanes []string) (int, []string) {
    // Check if hash already in a lane (branch continuation)
    if idx := findInLanes(hash, lanes); idx >= 0 {
        return idx, lanes
    }

    // Look for empty slot to reuse
    if idx := findEmptyLane(lanes); idx >= 0 {
        lanes[idx] = hash
        return idx, lanes
    }

    // Append new column
    lanes = append(lanes, hash)
    return len(lanes) - 1, lanes
}

func collapseTrailingEmpty(lanes []string) []string {
    // Find last non-empty lane
    lastNonEmpty := -1
    for i := len(lanes) - 1; i >= 0; i-- {
        if lanes[i] != "" {
            lastNonEmpty = i
            break
        }
    }

    if lastNonEmpty < 0 {
        return []string{}
    }
    return lanes[:lastNonEmpty+1]
}
```

**Pattern**: The lanes array dynamically grows/shrinks as branches appear/complete. Empty slots are reused before expanding.

### Symbol Generation

The symbol generation (`generate.go:13-140`) chooses the appropriate box-drawing character for each column:

```go
func generateRowSymbols(commitCol int, numCols int, mergeColumns []int,
                       convergingColumns []int, passingColumns []int,
                       existingLanesMerge []int, convergesToParent bool) Row {
    symbols := make([]GraphSymbol, numCols)

    // Build sets for quick lookup
    mergeSet := makeSet(mergeColumns)
    convergeSet := makeSet(convergingColumns)
    passingSet := makeSet(passingColumns)

    for col := 0; col < numCols; col++ {
        if col == commitCol {
            // This is the commit column
            if convergesToParent {
                symbols[col] = SymbolMergeCommit  // ├─
            } else if hasHorizontalLine {
                symbols[col] = SymbolMergeCommit  // ├─
            } else {
                symbols[col] = SymbolCommit  // ├
            }
        } else if mergeSet[col] && convergeSet[col] {
            // Both merging AND converging
            symbols[col] = SymbolMergeCross  // ┼─
        } else if mergeSet[col] {
            // Merge parent column
            symbols[col] = SymbolBranchTop  // ╮
        } else if convergeSet[col] {
            // Converging column
            symbols[col] = SymbolBranchBottom  // ╯
        } else if passingSet[col] {
            // Passing lane
            if inMergeLine {
                symbols[col] = SymbolBranchCross  // │─
            } else {
                symbols[col] = SymbolBranchPass  // │
            }
        } else {
            symbols[col] = SymbolEmpty  // "  "
        }
    }

    return Row{Symbols: symbols}
}
```

**Pattern**: Classify each column by its role (commit, merge, converge, pass, empty) and assign appropriate symbol.

### Integration into Views

Graph is rendered as part of commit lines:

```go
// From log/view.go:158-166
func (s State) buildGraphForCommit(commitIndex int) string {
    if s.GraphLayout != nil && commitIndex >= 0 && commitIndex < len(s.GraphLayout.Rows) {
        row := s.GraphLayout.Rows[commitIndex]
        return graph.RenderRow(row)
    }
    return ""
}

// From log/view.go:123-133
lineComponents := components.CommitLineComponents{
    Graph:        s.buildGraphForCommit(commitIndex),
    Hash:         format.ToShortHash(commit.Hash),
    Message:      commit.Message,
    // ...
}
line := components.FormatCommitLine(lineComponents, ctx.Width())
```

**Integration Pattern**:
1. Compute graph layout once (during state initialization)
2. Store `*graph.Layout` in state
3. For each visible line, extract corresponding row and render it
4. Pass rendered string to line formatting component

### Box-Drawing Characters

The graph uses Unicode box-drawing characters (U+2500 block):

| Symbol | Character | Unicode | Purpose |
|--------|-----------|---------|---------|
| `│` | Vertical | U+2502 | Branch continuation |
| `─` | Horizontal | U+2500 | Merge line |
| `├` | Vertical and Right | U+251C | Commit |
| `┤` | Vertical and Left | U+2524 | Merge join |
| `┬` | Horizontal and Down | U+252C | Octopus merge |
| `┴` | Horizontal and Up | U+2534 | Divergence |
| `╮` | Arc Down and Left | U+256E | Merge from right |
| `╯` | Arc Up and Left | U+256F | Merge ending |
| `┼` | Cross | U+253C | Merge crossing branch |

**Important**: Each box-drawing character is 1 column wide in monospace terminals. Combined with a space, each symbol is exactly 2 characters.

## Key Takeaways for Tree File View

Based on the patterns observed:

### 1. Use ViewBuilder for Composition

Build tree structure line-by-line using ViewBuilder:

```go
vb := components.NewViewBuilder()
for _, node := range treeNodes {
    line := formatTreeLine(node)
    vb.AddLine(line)
}
return vb
```

### 2. Model Tree Structure with Enums

Follow the GraphSymbol pattern - define tree symbols as an enum:

```go
type TreeSymbol int

const (
    TreeEmpty     TreeSymbol = iota  // "  "
    TreeBranch                       // "├─"
    TreeLastBranch                   // "└─"
    TreePass                         // "│ "
    // ... etc
)
```

### 3. Separate Layout from Rendering

Follow graph architecture:
- **Layout computation**: Pure domain logic, returns data structure
- **Symbol mapping**: Enum to string conversion
- **Rendering**: Simple traversal and concatenation

### 4. Handle Width Constraints Early

Calculate indentation and content width before rendering:

```go
indentWidth := depth * 2  // 2 chars per level
contentWidth := totalWidth - indentWidth - prefixWidth
truncatedPath := truncateWithEllipsis(path, contentWidth)
```

### 5. Use Progressive Truncation

For long paths, follow the truncation pattern:
1. Try full path
2. Try truncating filename
3. Try showing parent + truncated filename
4. Show just filename with ellipsis

### 6. Make Selection Work Like File List

Reuse the selector pattern from `file_section.go`:

```go
type TreeLineParams struct {
    Node       TreeNode
    IsSelected bool
    Width      int
    ShowSelector bool
}

func FormatTreeLine(params TreeLineParams) string {
    if params.ShowSelector {
        if params.IsSelected {
            line.WriteString("→")
        } else {
            line.WriteString(" ")
        }
    }

    // Apply selected style if needed
    if params.IsSelected {
        line.WriteString(styles.SelectedFilePathStyle.Render(node.Name))
    } else {
        line.WriteString(styles.FilePathStyle.Render(node.Name))
    }
}
```

### 7. Use Rune Counting for Width

Always use `utf8.RuneCountInString()` for user-visible text:

```go
// Good
width := utf8.RuneCountInString(text)

// Bad (byte count, breaks on UTF-8)
width := len(text)
```

### 8. Keep Components Pure

Pass all state as parameters, not via receiver:

```go
// Good - pure function
func FormatTreeLine(node TreeNode, isSelected bool, width int) string

// Bad - impure (depends on state)
func (s *State) FormatTreeLine(node TreeNode) string
```

### 9. Handle Fixed-Width Columns

Use lipgloss width styling for precise column control:

```go
colStyle := lipgloss.NewStyle().Width(width)
vb.AddLine(colStyle.Render(line))
```

### 10. Test with Golden Files

Follow the existing test pattern - render to string and compare with golden files:

```go
func TestTreeRendering(t *testing.T) {
    tree := buildTestTree()
    output := RenderTree(tree, 80)
    testutils.AssertGolden(t, output, "tree_rendering.golden", *update)
}
```

## Relevant Code Locations

For reference when implementing the tree file view:

- **ViewBuilder**: `/workspace/internal/ui/components/viewbuilder.go`
- **Styles**: `/workspace/internal/ui/styles/styles.go`
- **File Section**: `/workspace/internal/ui/components/file_section.go`
- **Log Line Format**: `/workspace/internal/ui/components/log_line_format.go`
- **Line Display State**: `/workspace/internal/ui/components/line_display_state.go`
- **Graph Layout**: `/workspace/internal/domain/graph/layout.go`
- **Graph Symbols**: `/workspace/internal/domain/graph/symbols.go`
- **Graph Generation**: `/workspace/internal/domain/graph/generate.go`
- **Log View**: `/workspace/internal/ui/states/log/view.go`
- **Files View**: `/workspace/internal/ui/states/files/view.go`
