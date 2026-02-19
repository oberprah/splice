# Requirements: Git Graph in Logs View

## Problem Statement

When viewing git history in Splice's logs view, users cannot visualize branch relationships or merge history. This makes it difficult to:
- Understand how branches were merged and when
- See the parallel development history across branches
- Quickly identify which branch a commit belongs to

## Goals

1. Display a visual git graph alongside commits in the logs view
2. Show branch/tag refs to identify important commits
3. Use color coding to distinguish different branches
4. Maintain readability of existing commit information

## Non-Goals

- Graph toggling or configuration options (always visible)
- ASCII fallback for non-Unicode terminals
- Interactive graph manipulation (e.g., collapsing branches)

## User Impact

Users viewing the logs will see a representation like:

```
├──╮ (HEAD -> main) abc123d Merge feature-x - Alice (4 min ago)
│  ├ def456a Add feature X - Bob (5 min ago)
│  ├ 111222c Another change - Bob (6 min ago)
├  │ 222333d Fix bug - Alice (7 min ago)
├──╯ 901234d Initial release - Alice (1 day ago)
```

This provides immediate visual context for:
- Merge points and branch splits
- Parallel development across branches
- Which commits are on which branch

## Key Requirements

### Functional

1. **Graph Display**
   - Graph appears to the left of the existing commit line
   - Uses Unicode box-drawing characters (`│ ├ ╮ ╯ ─` etc.)
   - Always visible in both simple view and split view modes

2. **Ref Decorations**
   - Show branch names, tags, and HEAD indicator
   - Display refs inline with commit information (e.g., `(HEAD -> main, origin/main)`)

3. **Color Coding**
   - Different branches use different colors for their graph lines
   - Colors help visually track branch lineage through the history

### Non-Functional

4. **Performance**
   - Graph rendering should not noticeably slow down the logs view
   - Handle repositories with complex branching history

5. **Compatibility**
   - Requires Unicode-capable terminal (acceptable limitation)

## Open Questions for Design Phase

- Maximum graph width and behavior when exceeded
- Specific color palette for branches
- How to fetch graph/topology data from git
- Handling of edge cases (octopus merges, very long ref names)

## Research References

- [Current Logs View Implementation](research/current-logs-view.md)
