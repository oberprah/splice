# Design Doc: Commit Multiselect (Visual Mode)

## Overview

Implement vim-style visual mode for selecting multiple commits in the log view. This enables viewing combined diffs across a commit range.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Selection model | Visual mode (vim-style) | Familiar to vim users |
| Preview panel | Not included | Keep scope minimal, matches user preference |
| Range representation | CommitRange struct | Normalizes older→newer ordering, single struct for both cases |

## Domain Objects

### CursorState (new)

```rust
pub enum CursorState {
    Normal { pos: usize },
    Visual { pos: usize, anchor: usize },
}
```

- `pos`: Current cursor position (moving)
- `anchor`: Fixed starting position (where visual mode was entered)
- Selection range = `[min(pos, anchor), max(pos, anchor)]`

### CommitRange (new)

```rust
pub struct CommitRange {
    pub start: Commit,  // Older commit (higher index in log)
    pub end: Commit,    // Newer commit (lower index in log)
    pub count: usize,   // Number of commits in range (1 = single)
}
```

**Important:** Git log index 0 = newest commit, so:
- `start` = older commit = higher index (max of selection)
- `end` = newer commit = lower index (min of selection)

### LineDisplayState (new)

```rust
pub enum LineDisplayState {
    None,           // "  " - not selected, not cursor
    Cursor,         // "→ " - normal mode cursor
    Selected,       // "▌ " - in visual selection range
    VisualCursor,   // "█ " - visual mode cursor position
}
```

## File Changes

### src/core/cursor.rs (new)

```rust
pub enum CursorState { ... }

impl CursorState {
    pub fn position(&self) -> usize;
}

pub fn selection_range(cursor: &CursorState) -> (usize, usize);
pub fn is_in_selection(cursor: &CursorState, index: usize) -> bool;
```

### src/core/commit_range.rs (new)

```rust
pub struct CommitRange { ... }

impl CommitRange {
    pub fn is_single_commit(&self) -> bool;
    pub fn to_diff_spec(&self) -> String;  // "hash" or "start^..end"
}
```

### src/app/log_view.rs

**Current:**
```rust
pub selected: usize,
```

**After:**
```rust
pub cursor: CursorState,
```

**Methods to add:**
- `fn is_visual_mode(&self) -> bool`
- `fn get_selected_range(&self) -> Option<CommitRange>`
- `fn enter_visual_mode(&mut self)`
- `fn exit_visual_mode(&mut self)`

### src/input.rs

**New Actions:**
```rust
pub enum Action {
    // ... existing
    ToggleVisualMode,
}
```

**Keybinding:** `v` key

### src/ui/log.rs

**Current:**
```rust
let is_selected = i == selected;
let prefix = if is_selected { "→ " } else { "  " };
```

**After:**
```rust
let display_state = get_line_display_state(&cursor, i);
let prefix = display_state.selector_string();
// Apply different styles for Selected/VisualCursor
```

### src/git/file_changes.rs

**Update signature:**
```rust
pub fn fetch_file_changes(range: &CommitRange) -> Result<Vec<FileChange>>
```

**Logic:**
- Single commit: `commit^..commit`
- Range: `start^..end` (includes changes from all commits in range)

### src/app/files_view.rs

**Current:**
```rust
pub commit: Commit,
```

**After:**
```rust
pub range: CommitRange,
```

**UI rendering:** Show range header for multi-commit selections:
- Single: `abc123 Add feature A`
- Range: `abc123..def456 (3 commits)`

### src/app/mod.rs

**Update `open_selected()`:**
```rust
// Current:
let commit = log.selected_commit();
let files_view = FilesView::new(commit, files);

// After:
let range = log.get_selected_range();
let files_view = FilesView::new(range, files);
```

**Update navigation handlers:** Preserve anchor in visual mode

## Interface: Log → Files

```
LogView.get_selected_range() → CommitRange
              ↓
        FilesView.range: CommitRange
              ↓
        git::fetch_file_changes(&CommitRange)
```

The `CommitRange` struct serves as the contract between views. Unlike Go, we don't need a `DiffSource` enum yet - `CommitRange` handles both single and multi-commit cases.

## User Flow

```
┌─────────────────────────────────────────┐
│ Log View (Normal Mode)                  │
│ → abc123 Add feature A                  │  ← cursor
│   def456 Fix bug B                      │
│   ghi789 Update docs                    │
└─────────────────────────────────────────┘
         │ Press 'v'
         ▼
┌─────────────────────────────────────────┐
│ Log View (Visual Mode)                  │
│ █ abc123 Add feature A                  │  ← cursor + anchor
│   def456 Fix bug B                      │
│   ghi789 Update docs                    │
└─────────────────────────────────────────┘
         │ Press 'j' twice
         ▼
┌─────────────────────────────────────────┐
│ Log View (Visual Mode)                  │
│ ▌ abc123 Add feature A                  │  ← selected
│ ▌ def456 Fix bug B                      │  ← selected
│ █ ghi789 Update docs                    │  ← cursor (anchor at abc123)
└─────────────────────────────────────────┘
         │ Press Enter
         ▼
┌─────────────────────────────────────────┐
│ Files View                              │
│ abc123..ghi789 (3 commits)              │
│                                         │
│ 5 files · +42 -18                       │
│ →├── src/main.rs                        │
│  ├── src/lib.rs                         │
└─────────────────────────────────────────┘
```

## Edge Cases

1. **Entering visual mode**: Anchor set to current position
2. **Boundary navigation**: Clamp to valid index range
3. **Empty commit list**: No-op for visual mode toggle
4. **DiffView with range**: Uses `range.start` for single-file diff lookup

## Testing Strategy

1. **Unit tests**: `CursorState`, `selection_range()`, `is_in_selection()`
2. **Integration tests**: `LogView::get_selected_range()`, navigation in visual mode
3. **E2E tests**: Visual mode entry, selection extension, exit, file view with range

## Implementation Order

1. Add `CursorState` to `src/core/cursor.rs`
2. Add `CommitRange` to `src/core/commit_range.rs`
3. Update `LogView` to use `CursorState`, add `get_selected_range()`
4. Add `ToggleVisualMode` action and `v` keybinding
5. Update `ui::log` rendering with display states
6. Update `fetch_file_changes` to accept `&CommitRange`
7. Update `FilesView` to store `CommitRange` instead of `Commit`
8. Update `DiffView` to work with `CommitRange`
9. Add tests
