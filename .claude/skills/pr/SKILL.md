---
name: pr
description: Create a pull request following repository conventions
---

# Pull Request Creation

## Core Principles

- **Help the reviewer understand**: Answer what, why, and implications
- **Information density**: Every sentence adds value, no filler

## Workflow

1. Run `git status` to check if current branch tracks a remote
2. Run `git diff main...HEAD` and `git log main...HEAD --format=fuller` to understand ALL changes
3. **Review all commits** that will be included in the PR (not just the latest)
4. Draft PR description using template in `.github/pull_request_template.md`
5. Push to remote with `-u` flag if needed
6. Create PR: `gh pr create --title "..." --body "$(cat <<'EOF'\n[body]\nEOF\n)"`
7. Return PR URL to user

## When to Include Optional Sections

- **How**: Include when technical decisions or implementation approach would help review (architecture choices, non-obvious solutions, trade-offs made)
- **Breaking Changes**: Include when existing functionality changes in ways that affect users or other code
