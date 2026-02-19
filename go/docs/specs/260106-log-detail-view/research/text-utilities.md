# Text Wrapping and Truncation Utilities Research

Research into available text utilities in the Splice codebase for implementing detail view improvements.

## Available Text Libraries

### Direct Dependencies
- **lipgloss v1.1.0** - TUI styling and layout library
- **Go stdlib** - `strings`, `unicode/utf8` packages

### Indirect Dependencies
- **go-runewidth** (via lipgloss) - Available but not explicitly used

### Not in Use
- **glamour** - Not included in dependencies

## Existing Wrapping Functions

### wrapText() - Simple Word Wrap

**Location**: `internal/ui/states/log_view.go` (lines 341-373)

**Signature**:
```go
func wrapText(text string, width int) []string
```

**Behavior**:
- Takes plain text string and maximum width
- Splits input on whitespace (`strings.Fields()`)
- Builds lines that don't exceed width by breaking on word boundaries
- Returns slice of wrapped lines
- Returns `[]string{""}` if width <= 0 or no words found

**Usage**:
- Used in `renderCommitMessage()` to wrap long body lines to 80-char panel width
- Only wraps lines longer than panel width

**Limitations**:
- Works on plain text only (no ANSI codes/styling)
- Doesn't preserve styled text during wrapping
- Simple word-break strategy

### Body Wrapping Implementation

**Location**: `internal/ui/states/log_view.go` (lines 208-244) in `renderCommitMessage()`

**Process**:
1. Splits commit body on `\n` characters
2. For each body line longer than 80 chars, calls `wrapText()`
3. Applies styling to wrapped lines
4. Respects `commitBodyMaxLines` constant (5)
5. Shows "..." indicator when body exceeds line limit

**Current Behavior**:
- Panel width: 80 characters (constant `splitPanelWidth`)
- Max body lines: 5 (constant `commitBodyMaxLines`)
- Truncation indicator: "..." (3 characters)
- No wrapping for subject line (truncated instead)

## Existing Truncation Functions

### capMessage() - Message Truncation

**Location**: `internal/ui/states/log_line_format.go` (lines 39-51)

**Signature**:
```go
func capMessage(message string, maxLen int) string
```

**Behavior**:
- Truncates message to maxLen runes (Unicode characters)
- Uses "…" (U+2026 ellipsis, 1 character) as indicator
- Returns original if fits within maxLen
- Uses `utf8.RuneCountInString()` for rune counting

**Usage**:
- Used in progressive truncation of commit lines
- Part of 10-level truncation strategy in `formatCommitLine()`

### truncateAuthor() - Author Name Truncation

**Location**: `internal/ui/states/log_line_format.go` (lines 53-66)

**Signature**:
```go
func truncateAuthor(author string, maxLen int) string
```

**Behavior**:
- Identical to `capMessage()`
- Truncates author name to maxLen runes
- Uses "…" as indicator
- Returns empty string if maxLen < 1

### TruncatePathFromLeft() - Path Truncation

**Location**: `internal/ui/states/commit_render.go` (lines 165-175)

**Signature**:
```go
func TruncatePathFromLeft(path string, maxWidth int) string
```

**Behavior**:
- Truncates file paths from the LEFT side (preserving filename)
- Uses "..." (3 characters) for truncation indicator
- Returns original path if it fits
- Example: `/very/long/path/to/file.go` → `...to/file.go`

### Metadata Line Truncation Bug

**Location**: `internal/ui/states/log_view.go` (lines 183-195) in `renderMetadataLine()`

**Current Code**:
```go
metadata := RenderCommitMetadata(commit, previewLoaded.Files, ctx)
if len(metadata) > width {
    return metadata[:width-3] + "..."
}
```

**Problems**:
1. Uses byte-length check (`len()`) instead of rune count
2. ANSI color codes not counted - metadata appears shorter than it is
3. Truncation point doesn't align with visible width
4. Uses "..." (3 chars) instead of "…" (1 char)

## Text Width Measurement

### utf8.RuneCountInString()

**Source**: Go standard library

**Usage in codebase**:
- `log_line_format.go` - Width measurements for commit line components
- Progressive truncation loop width comparison

**Behavior**:
- Counts Unicode runes (logical characters), not bytes
- Properly handles multi-byte UTF-8 sequences
- Performance: O(n) where n is string length

**Example**:
```go
utf8.RuneCountInString("Hello")     // returns 5
utf8.RuneCountInString("café")      // returns 4 (é is 1 rune)
utf8.RuneCountInString("你好")       // returns 2
```

### lipgloss Width Styling

**Source**: lipgloss library

**Usage**:
- `lipgloss.NewStyle().Width(n)` - Sets fixed-width padding
- Used for fixed-width column alignment in split view

**Note**: lipgloss does NOT expose a function to measure text width - it only allows setting fixed width for styling.

## Progressive Truncation Strategy

**Location**: `internal/ui/states/log_line_format.go` in `formatCommitLine()`

The codebase uses a 10-level progressive truncation strategy for commit lines in the left panel. This demonstrates the pattern of prioritizing information and gracefully degrading when space is limited.

## Recommendations for Design

1. **For subject line wrapping**: Use existing `wrapText()` function
2. **For metadata line truncation**: Create smart truncation with component prioritization
3. **For refs display**: Use `wrapText()` with comma-separated format
4. **For body truncation indicator**: Count remaining lines accurately
5. **For width measurement**: Continue using `utf8.RuneCountInString()`

---

**Research conducted**: 2026-01-06
