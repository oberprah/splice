# Current Log Line Rendering Implementation

This document describes how git log lines are currently rendered in Splice.

## Overview of Log Rendering Flow

The log line rendering in Splice follows this flow:

1. **Entry Point**: `LogState.View()` in `internal/ui/states/log_view.go`
   - Checks terminal width to determine if split view (≥160 chars) or simple view should be rendered
   - Routes to either `renderSplitView()` or `renderSimpleView()`

2. **Core Rendering**: Both views call `formatCommitLine()` for each visible commit
   - Located at line 135-220 in `log_view.go`
   - This is the central function that assembles all components into a single line

3. **Component Assembly**: `formatCommitLine()` builds the line from these components (in order):
   - Selection indicator (2 chars)
   - Graph symbols (variable width, from GraphLayout)
   - Commit hash (7 chars)
   - Space (1 char)
   - Refs/decorations (variable, optional)
   - Message (variable)
   - Separator " - " (3 chars)
   - Author (variable)
   - Time prefix " " (1 char)
   - Relative time (variable)

## Component-by-Component Breakdown

### 1. Selection Indicator (Lines 146-149)
```go
selectionIndicator := "  "
if isSelected {
    selectionIndicator = "> "
}
```
- Fixed 2 characters: `"> "` or `"  "`
- Never truncated

### 2. Graph Symbols (Lines 152-156)
```go
var graphSymbols string
if s.GraphLayout != nil && commitIndex >= 0 && commitIndex < len(s.GraphLayout.Rows) {
    row := s.GraphLayout.Rows[commitIndex]
    graphSymbols = graph.RenderRow(row)
}
```
- Variable width: Each symbol is 2 characters (defined in `internal/graph/symbols.go`)
- Examples: `"│ "`, `"├─"`, `"├─╮ "`, etc.
- Never truncated - width accepted as necessary for graph visualization

### 3. Commit Hash (Line 159)
```go
hash := format.ToShortHash(commit.Hash)  // 7 chars
```
- Fixed 7 characters (implemented in `internal/ui/format/hash.go`)
- Never truncated

### 4. Refs/Decorations (Lines 160, 113-132)
```go
refsStr := formatRefs(commit.Refs)  // Variable (includes trailing space if present)
```

The `formatRefs()` function (lines 113-132):
```go
func formatRefs(refs []git.RefInfo) string {
    if len(refs) == 0 {
        return ""
    }

    var parts []string
    for _, ref := range refs {
        var formatted string
        switch ref.Type {
        case git.RefTypeTag:
            formatted = fmt.Sprintf("tag: %s", ref.Name)
        default:
            // For branches, just use the name
            formatted = ref.Name
        }
        parts = append(parts, formatted)
    }

    return fmt.Sprintf("(%s) ", strings.Join(parts, ", "))
}
```
- Format: `"(HEAD -> main, tag: v1.0) "` or `""` if no refs
- Variable width
- **Currently has NO truncation logic** - refs are included in full or not at all
- Always includes trailing space if refs exist

### 5. Message (Lines 161, 179-181)
```go
message := commit.Message  // Variable
// ...later...
if len(message) > messageMaxWidth && messageMaxWidth > 3 {
    message = message[:messageMaxWidth-3] + "..."
}
```
- Variable width
- **Current truncation**: Simple right truncation with `"..."` when exceeds `messageMaxWidth`
- `messageMaxWidth` is calculated as 2/3 of remaining width (line 176)

### 6. Separator (Line 162)
```go
separator := " - "  // 3 chars
```
- Fixed 3 characters
- Never truncated

### 7. Author (Lines 163, 183-185)
```go
author := commit.Author  // Variable
// ...later...
if len(author) > authorMaxWidth && authorMaxWidth > 3 {
    author = author[:authorMaxWidth-3] + "..."
}
```
- Variable width
- **Current truncation**: Simple right truncation with `"..."` when exceeds `authorMaxWidth`
- `authorMaxWidth` is calculated as 1/3 of remaining width (line 177)

### 8. Time (Lines 164-165)
```go
timePrefix := " "  // 1 char
time := format.ToRelativeTimeFrom(commit.Date, ctx.Now())  // Variable
```
- Time prefix: Fixed 1 character
- Time value: Variable width (implemented in `internal/ui/format/time.go`)
- Format examples: `"just now"`, `"1 min ago"`, `"2 hours ago"`, `"Jan 2, 2006"` (for old commits)
- **Never truncated** - always shown in full

## Current Width Calculation (Lines 168-177)

```go
// Calculate required space for fixed elements (including graph symbols and refs)
fixedWidth := len(selectionIndicator) + len(graphSymbols) + len(hash) + 1 + len(refsStr) + len(separator) + len(timePrefix) + len(time)

// Calculate remaining space for message and author
remainingWidth := max(availableWidth-fixedWidth, 10)

// Truncate message and author to fit remaining space
messageMaxWidth := remainingWidth * 2 / 3 // Give 2/3 to message
authorMaxWidth := remainingWidth - messageMaxWidth
```

**Current Strategy**:
1. Treats all components except message and author as "fixed"
2. Calculates remaining width for message and author
3. Splits remaining width: 2/3 for message, 1/3 for author
4. Truncates message and author independently if they exceed their allocated width

**Problems with Current Approach**:
- Refs are treated as "fixed" but can be arbitrarily long (e.g., very long branch names)
- No graceful degradation - refs can consume huge amounts of space
- Time is never truncated/dropped even in narrow terminals
- Truncation can result in very narrow message/author fields if refs or graph are large
- Author and message use simple 2/3 and 1/3 split regardless of actual content length

## Styling Application (Lines 188-218)

After truncation, components are styled using styles from `internal/ui/styles/styles.go`:

**Selected line styles**:
- `SelectedHashStyle` (bold amber/yellow)
- `SelectedMessageStyle` (bold black/white)
- `SelectedAuthorStyle` (bold cyan)
- `SelectedTimeStyle` (bold gray)

**Unselected line styles**:
- `HashStyle` (amber/yellow)
- `MessageStyle` (gray/white)
- `AuthorStyle` (cyan)
- `TimeStyle` (dim gray)

Refs use `TimeStyle` (dim gray) for both selected and unselected states.

## Key Files and Functions

### Primary Rendering Files

1. **`internal/ui/states/log_view.go`**
   - `LogState.View()` (line 22) - Main entry point
   - `renderSimpleView()` (line 31) - Single column view
   - `renderSplitView()` (line 48) - Split view with details panel
   - `formatCommitLine()` (line 135) - **Core line formatting function**
   - `formatRefs()` (line 113) - Refs formatting

2. **`internal/ui/states/log_state.go`**
   - `LogState` struct (line 43) - State definition
   - Contains: Commits, Cursor, ViewportStart, Preview, GraphLayout

### Supporting Files

3. **`internal/ui/format/hash.go`**
   - `ToShortHash()` - Converts full hash to 7-char short hash

4. **`internal/ui/format/time.go`**
   - `ToRelativeTimeFrom()` - Converts timestamp to relative time string

5. **`internal/graph/symbols.go`**
   - `RenderRow()` - Renders graph symbols for a commit row
   - Each symbol is exactly 2 characters wide

6. **`internal/ui/styles/styles.go`**
   - All color/style definitions for each component

7. **`internal/git/git.go`**
   - `GitCommit` struct (line 28) - Commit data structure
   - `RefInfo` struct (line 21) - Branch/tag reference data

### Test Files

8. **`internal/ui/states/log_view_test.go`**
   - `TestLogState_View_LineTruncation` (line 112) - Tests truncation behavior
   - Uses golden files for snapshot testing

## Current Truncation Issues

Based on the requirements, the current implementation has these issues:

1. **No refs truncation**: Refs can be arbitrarily long and are always shown in full
   - Can result in unbalanced parentheses when line is cut off
   - No graceful degradation (e.g., showing count instead of names)

2. **No message length cap**: Messages aren't capped at 72 chars initially
   - Very long messages get proportional space rather than being capped

3. **No time dropping**: Time is always shown even in narrow terminals
   - Should be dropped in severely constrained layouts

4. **No author name shortening**: Author truncation is simple right-truncation only
   - No progressive degradation (25 chars → 5 chars → drop)

5. **Fixed ratio allocation**: Message and author always use 2/3 and 1/3 split
   - Not adaptive based on actual content or importance
