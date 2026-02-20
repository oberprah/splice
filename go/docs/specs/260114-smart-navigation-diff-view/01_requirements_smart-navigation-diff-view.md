# Requirements: Smart Navigation in Diff View

## Problem Statement

Currently, the diff view has basic change navigation (`n`/`N`) that jumps between changes within a single file, but it has limitations:
1. When a change is taller than the viewport, users can miss parts of it by jumping past
2. No way to navigate directly between files from the diff view
3. When reaching the end of a file, navigation stops - users must manually go back to files view to see the next file

This makes it cumbersome to review diffs with many files or large changes.

## Goals

Implement smart keyboard-driven navigation in the diff view that:
1. Intelligently handles multi-screen changes by scrolling through them before moving to the next change
2. Automatically continues to the next/previous file when reaching the end of the current file's changes
3. Provides direct file-to-file navigation shortcuts for quickly skipping between files

## Non-Goals

- Visual indicators showing which change you're on (e.g., "change 3 of 15") - keep UI minimal
- Configurable navigation behavior - use single, sensible behavior
- Navigation history or "jump back" functionality
- Filtering or searching for specific changes

## User Impact

Users reviewing multi-file diffs will be able to navigate the entire changeset without leaving the diff view, with confidence they won't miss any parts of tall changes.

## Functional Requirements

### 1. Smart Change Navigation (`n` and `p`)

**Replace current shortcuts:** `n` (next change) and `N` (previous change) with `n` (next) and `p` (previous)

**Behavior for `n` (next change):**
1. If currently viewing a change that extends below the viewport:
   - Scroll down half a page to show more of the current change
   - Continue doing this on subsequent `n` presses until the entire change is visible
2. Once the current change is fully visible (or if it was already fully visible):
   - Jump to the start of the next change in the current file
3. If at the last change in the current file:
   - Jump to the first change in the next file
4. If at the last change in the last file:
   - Stay in place (no wrapping)

**Behavior for `p` (previous change):**
1. If currently viewing a change that extends above the viewport:
   - Scroll up half a page to show more of the current change
   - Continue doing this on subsequent `p` presses until viewing the start of the change
2. Once at the start of the current change (or if already there):
   - Jump to the start of the previous change in the current file
3. If at the first change in the current file:
   - Jump to the first change in the previous file
4. If at the first change in the first file:
   - Stay in place (no wrapping)

**Multi-screen change detection:**
- A change is considered "multi-screen" if it extends beyond the current viewport boundaries
- Use half-page scrolling (consistent with existing `ctrl+d`/`ctrl+u` behavior) to navigate through tall changes

### 2. Direct File Navigation (`]` and `[`)

**New shortcuts:** `]` (next file) and `[` (previous file)

**Behavior for `]` (next file):**
1. Jump to the first change in the next file in the file list
2. If at the last file, stay in place (no wrapping)

**Behavior for `[` (previous file):**
1. Jump to the first change in the previous file in the file list
2. If at the first file, stay in place (no wrapping)

**File ordering:**
- Use the same file order as displayed in the files view (alphabetical by path)

### 3. State Management

**The diff state must track:**
- Current file being viewed (needs to be added if not already present)
- List of all files in the current diff source (for file navigation)
- Current position within multi-screen changes (to know when to scroll vs jump)

**State transitions:**
- When navigating to a new file, load its diff if not already loaded
- Preserve the async loading pattern used by the files view

## Non-Functional Requirements

### Performance
- File-to-file navigation should feel instant when diff is already loaded
- Async loading of diffs should show appropriate loading state (consistent with current behavior)

### Consistency
- Half-page scrolling should use the same logic as existing `ctrl+d`/`ctrl+u` shortcuts
- File navigation should maintain the same navigation stack behavior (can still press `q` to go back)

### Maintainability
- Keep the stateful navigation logic testable with unit tests
- Reuse existing diff loading and parsing infrastructure

## Open Questions

None - ready for design phase.

## References

- [Current diff view implementation](research/current-diff-view.md)
- Existing shortcuts: `n`/`N` for change navigation (to be replaced)
- Existing shortcuts: `ctrl+d`/`ctrl+u` for half-page scrolling (reuse same logic)
