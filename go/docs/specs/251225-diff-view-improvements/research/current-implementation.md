# Current Diff View Implementation

## Overview

The diff view renders side-by-side columns using `lipgloss.JoinHorizontal()`. Each column has a fixed width calculated as `(terminalWidth - 3) / 2`.

## Key Files

- `internal/ui/states/diff_view.go` - Rendering logic (View, renderFullFileLine, formatColumnContent, renderTokens)
- `internal/diff/merge.go` - Data structures (FullFileLine, FullFileDiff) and merge algorithm
- `internal/highlight/highlight.go` - Syntax tokenization using Chroma

## Data Structure

```
FullFileLine
├── LeftLineNo: int (0 if line doesn't exist on left)
├── RightLineNo: int (0 if line doesn't exist on right)
├── LeftTokens: []Token (syntax tokens for left side)
├── RightTokens: []Token (syntax tokens for right side)
└── Change: ChangeType (Unchanged, Added, Removed)
```

## Current Rendering Behavior

| Change Type | Left Column | Right Column |
|-------------|-------------|--------------|
| Unchanged   | Shows line with gray background | Shows same line with gray background |
| Removed     | Shows line with red background | Empty (whitespace padding) |
| Added       | Empty (whitespace padding) | Shows line with green background |

## Alignment Mechanism

Each column is padded to fixed width using `bgStyle.Width(columnWidth).Render()`. This ensures vertical alignment but means:
- Removed lines have empty right columns
- Added lines have empty left columns

## Syntax Highlighting

- Uses Chroma for tokenization (per-file, preserving multi-line context)
- Tokens are rendered character-by-character
- Syntax foreground colors combined with diff background colors via `.Inherit(bgStyle)`
- NO word-level diff highlighting exists (only syntax highlighting)

## Relevant Tests

- `internal/ui/states/diff_view_test.go` - 15+ test cases
- `internal/diff/merge_test.go` - Merge algorithm tests
