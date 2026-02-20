# Requirements: Direct Diff View

## Problem Statement

Currently, Splice only supports viewing commit history via `git log`. Users must:
1. Start with the log view (list of commits)
2. Select a commit to view its files
3. Select a file to view its diff

However, many git workflows involve viewing diffs that don't correspond to a single commit in history:
- Uncommitted changes (staged and unstaged)
- Branch comparisons (`main..feature-branch`, `main...feature-branch`)
- Arbitrary commit ranges (`HEAD~5..HEAD`, `abc123..def456`)

These workflows are currently unsupported, forcing users to use other tools or the command line for these common tasks.

## Goals

Enable users to view any git diff directly by:
- Adding a `splice diff <spec>` subcommand that accepts git diff specifications
- Jumping directly to the files view, skipping the log view entirely
- Reusing all existing diff viewing logic (files view, diff view, syntax highlighting, navigation)
- Supporting the same diff viewing features users already know (n/N for next/prev change, split view, etc.)

## Non-Goals

- **Not** changing how existing `splice` command works (backward compatibility)
- **Not** adding new diff visualization features (use existing diff view as-is)
- **Not** supporting direct jump to a specific file's diff (e.g., `splice diff main..feature src/main.go`)
- **Not** adding file-to-file navigation within diff view (e.g., ]f/[f to jump between files)

These may be considered for future enhancements but are explicitly out of scope for this feature.

## User Impact

### Who Benefits
- Developers reviewing uncommitted changes before committing
- Developers comparing branches during code review
- Developers examining specific commit ranges

### Use Cases

**Use Case 1: Review uncommitted changes**
```bash
# View all unstaged changes
splice diff

# View staged changes only
splice diff --staged

# View all uncommitted changes (staged + unstaged)
splice diff HEAD
```

**Use Case 2: Compare branches**
```bash
# See what's in feature branch that's not in main
splice diff main..feature-branch

# See changes since branching point
splice diff main...feature-branch
```

**Use Case 3: Review recent work**
```bash
# See changes in last 5 commits
splice diff HEAD~5..HEAD

# Compare specific commits
splice diff abc123..def456
```

### Expected Workflow
1. User runs `splice diff <spec>` from command line
2. App loads and fetches file changes for the specified diff
3. Files view appears showing list of changed files (same UI as current files view)
4. User navigates and selects a file to view
5. Diff view appears showing side-by-side diff (same UI as current diff view)
6. User presses `q` to return to files view, or `q` again to exit app

## Key Requirements

### Functional Requirements

**FR1: Subcommand invocation**
- Add `splice diff <spec>` subcommand
- `<spec>` accepts any git diff specification that produces a diff
- Examples: commit ranges, branch comparisons, flags like `--staged`

**FR2: Uncommitted changes support**
- `splice diff` (no arguments) → unstaged changes only (equivalent to `git diff`)
- `splice diff HEAD` → all uncommitted changes (staged + unstaged)
- `splice diff --staged` → staged changes only (equivalent to `git diff --staged`)

**FR3: Commit range support**
- Two-dot ranges: `commit1..commit2`, `HEAD~5..HEAD`, `main..feature`
- Three-dot ranges: `main...feature-branch`
- Any other valid git diff specifications

**FR4: Files view as entry point**
- Always start with files view showing list of changed files
- Never jump directly to a single file's diff
- If diff spec is invalid → show error message and exit
- If diff spec is valid but has no changes → show error message and exit

**FR5: Reuse existing diff viewing**
- Use same files view UI (list, cursor navigation, status indicators)
- Use same diff view UI (split view, syntax highlighting, change navigation)
- Use same keybindings (j/k, Enter, q, n/N, g/G, etc.)

**FR6: Navigation behavior**
- From files view: `q` exits the application (no log view to return to)
- From diff view: `q` returns to files view
- No file-to-file navigation in diff view (must return to files view to select another file)

**FR7: Backward compatibility**
- `splice` (no arguments) continues to work as before → log view
- Existing workflows are not affected

### Non-Functional Requirements

**NFR1: Error handling**
- Invalid diff specs (nonexistent branches, invalid hashes) → error message and exit
- Empty diffs (no changes) → error message and exit
- Git command failures → clear error messages

**NFR2: Consistency**
- UI/UX should feel identical to existing file/diff viewing
- Keybindings should match existing navigation patterns
- Error messages should follow existing style

**NFR3: Testing**
- TDD approach (write tests first)
- Unit tests for new functionality
- Golden file tests for UI rendering
- E2E tests for complete workflows
- Follow guidelines in `docs/guidelines/testing-guidelines.md`

## Open Questions for Design Phase

**Q1: Data model refactoring**
The current architecture uses `CommitRange` (with Start and End commits) throughout the codebase. Uncommitted changes don't have actual commits.

Design decisions needed:
- Should we use a sum type to represent different diff sources (CommitRange vs UncommittedChanges)?
- Or can we design a structure that handles both cases elegantly?
- How should the files view header display uncommitted changes vs commit ranges?

**Q2: Generic git diff argument support**
Should the implementation be generic enough to support ANY git diff command/flags?
- If yes, how do we parse and validate arbitrary git diff arguments?
- Should we pass through flags like `--ignore-whitespace`, `--word-diff`, etc.?
- Or maintain an explicit allowlist of supported specifications?

**Q3: Flag handling**
Git accepts both `--staged` and `--cached` as synonyms. Should we:
- Support both (more git-compatible)
- Pick one (simpler)
- Design generically to pass through any valid git diff flags?

## References

- [Current Implementation Research](research/current-implementation.md) - Analysis of existing log/files/diff flow
- See `CLAUDE.md` for architecture patterns (state machine, navigation messages, async loading)
- See `docs/guidelines/testing-guidelines.md` for testing approach
