# Requirements: Improve Log Detail View

## Problem Statement

The detail view (right pane) in the log view has several UX issues:

1. **Flickering on navigation** - When navigating to a new commit, the panel flickers as the metadata line appears after preview loads, pushing all content down
2. **Unnecessary truncation bug** - Metadata line (e.g., `fc5a4c3 · oberprah co...`) is truncated even when it's well under the 80-character panel width
3. **Missing refs information** - Refs (branches, tags) are shown in the left commit list but not in the detail view
4. **Unclear body truncation** - When commit body exceeds 5 lines, it's silently cut off with no indication
5. **Misaligned separator line** - The horizontal line `────────` has alignment issues
6. **Subject line truncation** - Long subject lines are truncated instead of wrapped

## Goals

Improve the detail view to:
- Eliminate flickering during navigation
- Show complete commit information (including refs)
- Provide clear visual feedback for truncated content
- Fix truncation bugs and improve text handling
- Simplify layout by removing unnecessary visual elements

## User Impact

Users navigating through commits in the log view will experience:
- Smoother, flicker-free transitions between commits
- More complete information about each commit (with refs visible)
- Better understanding of commit body content (with truncation indicators)
- Cleaner, more consistent layout

## Functional Requirements

### 1. Eliminate Flickering

**Current behavior:**
- Subject line appears first
- Then metadata line loads and pushes everything down
- Causes visual flicker/jump

**Required behavior:**
- Show all commit information immediately on navigation (hash, author, timestamp, refs, subject, body)
- Only show "Loading files..." for the file list section
- When files load, only that section updates (no content shift above it)

**Rationale:** All commit metadata is already available in the log data; only file changes require async loading.

### 2. Fix Metadata Line Truncation

**Current behavior:**
- Metadata line like `fc5a4c3 · oberprah co...` is truncated even when well under 80 characters
- Uses "..." (3 characters) for truncation

**Required behavior:**
- Fix truncation bug so line only truncates when actually needed
- Use "…" (single ellipsis character) instead of "..."
- Implement smarter truncation logic (prioritize important information)
- Allow wrapping to multiple lines if needed (edge case handling)

### 3. Show Refs Information

**Current behavior:**
- Refs are not shown in detail view at all
- Only visible in left commit list

**Required behavior:**
- Display refs on a separate line after metadata line
- Format: comma-separated list (e.g., `main, origin/main, HEAD, v1.2.0`)
- Wrap to multiple lines if refs list is too long
- Skip the refs line entirely if commit has no refs (don't show blank line)

### 4. Indicate Body Truncation

**Current behavior:**
- Body limited to 5 lines (via `commitBodyMaxLines` constant)
- No indication when body is longer

**Required behavior:**
- Show indicator when body exceeds 5 lines
- Format: `(... N more lines)` where N is the count of remaining lines
- Example: `(... 3 more lines)` if body has 8 lines total

### 5. Remove Separator Line

**Current behavior:**
- Horizontal line `────────` separates message from file list
- Has alignment issues (doesn't extend to left edge)
- File stats shown in metadata line: `3 files · +10 -10`

**Required behavior:**
- Remove horizontal separator line entirely
- Move file stats from metadata line to new line above file list
- Format: `3 files · +10 -10`

### 6. Wrap Subject Line

**Current behavior:**
- Subject line is truncated to panel width with no indication

**Required behavior:**
- Wrap subject line to multiple lines if needed instead of truncating
- No ellipsis needed (full subject always shown)

### 7. Body Line Wrapping (Confirmed OK)

**Current behavior:**
- Long body lines are wrapped to fit 80-character panel width

**Required behavior:**
- Keep current wrapping behavior (no change needed)

**Rationale:** Wrapping is acceptable because showing the message content is more important than preserving formatting for edge cases (code blocks, ASCII art, etc.).

## New Layout

### Example with refs and truncated body:

```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD
(blank)
Use single ellipsis character for truncation
(blank)
Replace triple-dot "..." (3 chars) with single ellipsis "…" (U+2026,
1 char) in message and author truncation. Saves 2 visual characters
per truncation, allowing 2 more content characters to display.
(... 3 more lines)
(blank)
3 files · +10 -10
M +10 -10  internal/ui/states/log_line_format.go
M +29 -41  internal/ui/states/log_line_format_test.go
M +6 -0  internal/ui/states/log_view_test.go
```

### Example without refs, full body:

```
abc123d · Alice committed 11 months ago
(blank)
First commit
(blank)
This is the complete commit body.
(blank)
2 files · +65 -12
M +45 -12  src/main.go
A +20 -0  README.md
```

### Loading state:

```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD
(blank)
Use single ellipsis character for truncation
(blank)
Replace triple-dot "..." (3 chars) with single ellipsis...
(blank)
Loading files...
```

### Error state:

```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD
(blank)
Use single ellipsis character for truncation
(blank)
Replace triple-dot "..." (3 chars) with single ellipsis...
(blank)
Unable to load files
```

## Non-Functional Requirements

- No performance degradation during navigation
- Maintain existing test coverage (update golden files as needed)
- Preserve existing keyboard navigation behavior
- Maintain 80-character fixed width for detail panel

## Open Questions for Design Phase

1. **Metadata line wrapping strategy** - When author name or timestamp is very long, should we wrap or use smarter truncation? (Rare edge case, either approach acceptable)
2. **Refs styling** - Should different ref types (branches, tags, HEAD) have different colors/styling, or keep simple monochrome?
3. **Loading indicator style** - Current "Loading files..." adequate or enhance it?

## Out of Scope

- Changing the 80-character panel width
- Making file list section interactive (scrolling, selecting files)
- Changing the 5-line limit for commit body display
- Supporting rich formatting in commit messages (Markdown, code highlighting)
- Showing full commit diff in detail panel

## References

- [Current implementation research](research/current-implementation.md)
- Key files:
  - `internal/ui/states/log_view.go` - Main rendering logic
  - `internal/ui/states/commit_render.go` - Metadata formatting
  - `internal/ui/states/log_state.go` - State definitions
