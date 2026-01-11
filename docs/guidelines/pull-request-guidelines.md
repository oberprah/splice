# Pull Request Guidelines

## Core Principles

**Help the reviewer understand the change**. Every PR should answer: what changed, why it changed, and what implications it has. Include the sections that provide value for understanding this specific change.

**Information density matters**. Every sentence should add value. Avoid filler words and unnecessary elaboration. Be direct and concise.

## Workflow

1. Run `git status` to check if current branch tracks a remote
2. Run `git diff main...HEAD` and `git log main...HEAD --format=fuller` to understand all changes
3. Review all commits that will be included in the PR, not just the latest
4. Draft PR description using the template below (include optional sections when relevant)
5. Push to remote with `-u` flag if needed
6. Create PR using GitHub CLI: `gh pr create --title "..." --body "..."`
7. Return the PR URL

## When to Include Optional Sections

- **How**: Include when technical decisions or implementation approach would help review (architecture choices, non-obvious solutions, trade-offs made)
- **Breaking Changes**: Include when existing functionality changes in ways that affect users or other code

## PR Description Template

```md
**TL;DR**: One sentence summary - what changed and why

**Why**
What problem does this solve? What's the impact or benefit?

[OPTIONAL] **How**
Implementation approach and key technical decisions

[OPTIONAL] **Breaking Changes**
What breaks, migration strategy, and rollback plan
```

## Example PR Descriptions

### Example 1: Feature Addition (Simple)
```md
**TL;DR**: Add --format flag to support JSON output

**Why**
Users need structured output for scripting and integration with other tools. Plain text format requires fragile parsing.
```

### Example 2: Refactoring (With Implementation Details)
```md
**TL;DR**: Refactor state machine to use typed messages instead of string commands

**Why**
String-based commands cause runtime errors and make the codebase harder to maintain. Type safety catches bugs at compile time.

**How**
- Replaced command strings with typed message structs (Push*ScreenMsg, PopScreenMsg)
- Updated app.Model to route messages based on type instead of string matching
- Added compile-time verification of state transitions
```

### Example 3: Bug Fix
```md
**TL;DR**: Fix crash when viewing diff for merge commits

**Why**
App panics when selecting merge commits because diff parsing assumes single parent. Affects ~15% of commits in typical repos.

**How**
Detect merge commits (>1 parent) and show combined diff using `git diff-tree --cc` instead of regular diff.
```

### Example 4: Breaking Change
```md
**TL;DR**: Remove deprecated --legacy-format flag

**Why**
Legacy format hasn't been used since v0.3 (released 6 months ago). Removing it simplifies output formatting code and reduces maintenance burden.

**Breaking Changes**
- `--legacy-format` flag removed - scripts using it will fail with "unknown flag" error
- **Migration**: Remove the flag from scripts (new format has been the default since v0.3)
- **Rollback**: Pin to v0.9.x if legacy format is still needed
```
