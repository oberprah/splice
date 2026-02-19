# Requirements: Files Preview Tree Structure

## Problem Statement

The files preview panel (shown in the log view's split mode) currently displays files as a flat list, while the dedicated files view shows a hierarchical tree structure with collapsible folders. This inconsistency creates a disjointed user experience when viewing the same commit's files in different contexts.

Users expect the same visual representation of file structure regardless of where they're viewing it, making it easier to understand the scope of changes and navigate mentally between the two views.

## Goals

- **Consistency**: Use the same tree structure display in both the files preview and files view
- **Clarity**: Show folder hierarchy in the preview to help users understand file organization at a glance
- **Simplicity**: Keep the preview read-only without adding interaction complexity

## Non-Goals

- Making the preview interactive (navigation, expand/collapse)
- Changing the files view behavior
- Adding any new keyboard shortcuts or user interactions
- Implementing smart folder collapse heuristics or state persistence

## User Impact

**Before**:
- Log view preview shows: `src/components/App.tsx`
- Files view shows: `📁 src/ → 📁 components/ → App.tsx`
- Users see two different representations of the same data

**After**:
- Both views show the same tree structure
- Preview displays full tree with all folders expanded (read-only)
- Files view remains interactive with collapse/expand
- Consistent visual representation across contexts

## Key Requirements

### Functional Requirements

1. **Tree Structure Display**
   - Files preview must display files using a tree structure with folder hierarchy
   - Use the same tree symbols and indentation as the files view

2. **Default Folder State**
   - All folders should be expanded by default in the preview
   - Show the complete tree structure without any collapsed folders

3. **Read-Only Display**
   - No cursor or selection indicator in the preview
   - No keyboard interaction in the preview panel
   - Preview remains a visual-only component

4. **Component Usage**
   - Replace `FileSection` component with `TreeSection` component in the log view preview
   - Reuse existing `filetree` package logic for building tree structure

### Non-Functional Requirements

1. **Performance**
   - Tree building should not noticeably impact preview rendering performance
   - Should handle typical commit file counts (dozens to hundreds of files) smoothly

2. **Visual Consistency**
   - Tree symbols, indentation, and styling must match the files view exactly
   - File stats display (status, additions, deletions) should remain unchanged

## Technical Constraints

- Must use existing `filetree` package for tree structure building
- Must use existing `TreeSection` component for rendering
- Preview panel has fixed width (80 characters) in split view
- Must handle truncation when tree doesn't fit in available height

## Open Questions

None - ready for design phase.

## References

- [Current Implementation Research](research/current-implementation.md)
- Relevant components: `FileSection`, `TreeSection`
- Relevant packages: `internal/domain/filetree`
- Log view rendering: `internal/ui/states/log/view.go`
