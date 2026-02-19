# Requirements: Multi-Commit Selection

## Problem Statement

Splice currently only allows viewing one commit at a time. Users reviewing feature branches or understanding a series of related changes must either view each commit individually (tedious) or fall back to external tools like `git diff branch1..branch2` (defeats the purpose of using Splice).

## Goals

- Enable users to select a range of commits in LogState and view their combined diff
- Maintain consistency with existing navigation (LogState → FilesState → DiffState)
- Provide clear visual feedback during range selection using cursor styles
- Support the primary use case: viewing cumulative changes across multiple commits

## User Impact

Users can now select any two commits in the log view and instantly see all changes between them, streamlining feature branch reviews and multi-commit analysis.

## Key Requirements

### Selection Interaction

| Action | Key | Behavior |
|--------|-----|----------|
| Start selection | `v` | Anchors current commit, enters visual mode |
| Extend selection | Navigation keys | Move cursor, all commits between anchor and cursor are selected |
| View combined diff | `Enter` | Opens FilesState showing combined diff of selected range |
| Cancel selection | `Escape` | Exits visual mode, clears selection |

### Visual Feedback (Cursor Styles)

| Mode | Element | Cursor |
|------|---------|--------|
| Normal | Current commit | `→` |
| Visual | Selected commits | `▌` |
| Visual | Current cursor | `█` |

- Selected commits (the range between anchor and cursor) receive line highlighting
- Cursor shape alone indicates the current mode - no additional UI elements needed

### Diff Behavior

- Uses `git diff <older>..<newer>` under the hood
- Commit order is always normalized to chronological (older..newer), regardless of selection direction
- If anchor and cursor are the same commit, shows that single commit's diff (graceful fallback)

### FilesState / DiffState Header

- When viewing a range, header displays the commit range, e.g., `abc123..def456 (5 commits)`
- Replaces single commit hash display for range diffs

## Visual Examples

### Normal Mode

Cursor on second commit:

```
  → ├ abc1234 Fix authentication bug - Test Author 1 year ago
    ├ def5678 Add dark mode toggle - Test Author 1 year ago
    ├ 9012abc Update dependencies - Test Author 1 year ago
    ├ 345def6 Refactor git parsing - Test Author 1 year ago
    ├ 789abc0 Initial commit - Test Author 1 year ago
```

### Visual Mode (Linear History)

Pressed `v` on "Fix authentication bug", moved cursor down to "Refactor git parsing":

```
  ▌ ├ abc1234 Fix authentication bug - Test Author 1 year ago   ← anchor
  ▌ ├ def5678 Add dark mode toggle - Test Author 1 year ago     ← selected
  ▌ ├ 9012abc Update dependencies - Test Author 1 year ago      ← selected
  █ ├ 345def6 Refactor git parsing - Test Author 1 year ago     ← cursor
    ├ 789abc0 Initial commit - Test Author 1 year ago
```

All 4 rows would have line highlight styling.

### FilesState Header (Range)

After pressing Enter on the selection above:

```
abc1234..345def6 (4 commits)

Feature: Authentication and refactoring changes

3 files · +120 -45
→ M +45 -12  internal/auth/login.go
  M +50 -20  internal/ui/theme.go
  M +25 -13  internal/git/parser.go
```

### Visual Mode (Branching History)

Selecting from merge commit to initial commit:

```
  ▌ ├─╮ eeeeeee Merge feature-x - Alice 11 months ago        ← anchor
  ▌ │ ├ ddddddd Add feature X part 2 - Bob 11 months ago     ← selected
  ▌ │ ├ ccccccc Add feature X part 1 - Bob 11 months ago     ← selected
  ▌ ├ │ bbbbbbb Fix bug on main - Alice 11 months ago        ← selected
  █ ├─╯ aaaaaaa Initial commit - Alice 11 months ago         ← cursor
```

Selection is purely visual/positional - it highlights all rows between anchor and cursor regardless of branch topology. The actual `git diff eeeeeee..aaaaaaa` compares the two endpoint trees.

### Visual Mode (Feature Branch Only)

Selecting just two commits on a feature branch:

```
    ├─╮ eeeeeee Merge feature-x - Alice 11 months ago
  ▌ │ ├ ddddddd Add feature X part 2 - Bob 11 months ago     ← anchor
  █ │ ├ ccccccc Add feature X part 1 - Bob 11 months ago     ← cursor
    ├ │ bbbbbbb Fix bug on main - Alice 11 months ago
    ├─╯ aaaaaaa Initial commit - Alice 11 months ago
```

This would show `git diff ccccccc..ddddddd` - just the changes between those two feature commits.

## Open Questions

None - ready for design phase.

## References

- Existing navigation flow: LogState → FilesState → DiffState (see `internal/ui/states/`)
- Git diff range syntax: `git diff <commit>..<commit>`
