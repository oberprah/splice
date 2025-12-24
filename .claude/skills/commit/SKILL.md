---
name: commit
description: Always use this skill when committing in this repository. Never use the standard commit workflow.
---

# Commit

## Workflow

1. Check changes by running `git status` and review the diffs
2. Analyze changes considering conversation context for rationale and intent
3. Generate commit message answering: **What behavior changes and why?**
4. Create the commit

## Message Guidelines

- **Imperative voice**: "Add feature" not "Added feature"
- **Behavior focus**: Describe outcomes, not code mechanics
- **Include rationale**: Briefly explain why the change was necessary
- **High information density**: Avoid obvious details
- **Verified facts only**: No unverified performance claims
- **Line length**: Aim for ≤50 chars in subject, ≤72 chars per line in body
- **No attribution**: Omit Claude Code co-author tags
