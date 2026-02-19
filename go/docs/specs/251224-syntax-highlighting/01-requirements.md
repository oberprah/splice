# Requirements: Syntax Highlighting in Diff View

## Problem Statement

The diff view currently displays file content without syntax highlighting. All code appears in a single color (gray for unchanged lines), making it harder to read and understand the code structure. Developers expect syntax-aware coloring when viewing code, similar to their editors and other diff tools.

## Goals

1. **Add syntax highlighting** to the diff view so code is displayed with language-aware coloring (keywords, strings, comments, etc.)
2. **Preserve diff indicators** - changed lines should show syntax colors on top of the existing subtle green/red background colors
3. **Support common languages** - Go, JavaScript, TypeScript, Python, Rust, Java, C/C++, and other mainstream languages
4. **Adapt to terminal theme** - automatically use appropriate colors for light and dark terminal backgrounds

## Non-Goals

- User-configurable themes (single auto light/dark theme is sufficient)
- Content-based language detection for unknown file types
- Syntax highlighting in other views (log list, file list)

## User Impact

All users viewing diffs will see improved code readability through syntax-highlighted content. This brings the diff viewing experience closer to modern code editors.

## Functional Requirements

### FR1: Syntax Highlighting
- Code in the diff view must be rendered with syntax highlighting
- Highlighting must be based on the file's language (detected from file extension)
- Both the left (old) and right (new) columns must be highlighted

### FR2: Changed Line Rendering
- Added lines: syntax colors applied on top of green background
- Removed lines: syntax colors applied on top of red background
- Unchanged lines: syntax colors with no background

### FR3: Language Support
- Must support common programming languages including at minimum:
  - Go
  - JavaScript / TypeScript
  - Python
  - Rust
  - Java
  - C / C++
- Language detection via file extension

### FR4: Fallback Behavior
- Files with unrecognized extensions: display without syntax highlighting (current behavior)
- No errors or warnings for unsupported file types

### FR5: Theme Adaptation
- Automatically select appropriate syntax colors for light terminal backgrounds
- Automatically select appropriate syntax colors for dark terminal backgrounds
- Must work with the existing Lip Gloss `AdaptiveColor` pattern used in the project

## Non-Functional Requirements

### NFR1: Performance
- Syntax highlighting should not introduce noticeable lag when opening diffs
- Must handle typical source files without degradation

### NFR2: Consistency
- Syntax colors should be visually harmonious with existing UI colors
- Changed line backgrounds must remain visible beneath syntax colors

## Open Questions for Design Phase

1. Which Go library should be used for syntax highlighting?
2. How to combine syntax highlighting ANSI codes with Lip Gloss background styles?
3. What specific color palette to use for syntax tokens?

## References

- [Current Diff View Implementation](research/current-diff-view.md)
