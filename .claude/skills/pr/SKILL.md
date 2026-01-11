---
name: pr
description: Create a pull request following repository conventions
---

# Pull Request Creation

**Read `docs/guidelines/pull-request-guidelines.md` for complete guidance.**

## Core Principles

- **Help the reviewer understand**: Answer what, why, and implications
- **Information density**: Every sentence adds value, no filler

## Workflow

1. Run `git status` to check if current branch tracks a remote
2. Run `git diff main...HEAD` and `git log main...HEAD --format=fuller` to understand ALL changes
3. **Review all commits** that will be included in the PR (not just the latest)
4. Draft PR description using template below
5. Push to remote with `-u` flag if needed
6. Create PR: `gh pr create --title "..." --body "$(cat <<'EOF'\n[body]\nEOF\n)"`
7. Return PR URL to user

## PR Description Template

```md
**TL;DR**: One sentence summary - what changed and why

**Why**
What problem does this solve? What's the impact or benefit?

[OPTIONAL] **How**
Implementation approach and key technical decisions
(Include when: architecture choices, non-obvious solutions, trade-offs)

[OPTIONAL] **Breaking Changes**
What breaks, migration strategy, and rollback plan
(Include when: existing functionality changes affecting users/code)
```

## Example

```md
**TL;DR**: Add --format flag to support JSON output

**Why**
Users need structured output for scripting and integration with other tools. Plain text format requires fragile parsing.
```
