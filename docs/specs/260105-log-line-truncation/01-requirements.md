# Requirements: Log Line Truncation Strategy

## Problem Statement

When displaying git commit log lines, various components (hash, refs, message, author, time) can exceed the available terminal width, especially with very long branch names or commit messages. The current implementation truncates refs by cutting them mid-string with "..." which creates unbalanced parentheses like `(feature/very-long-branch-name-that-descr...` - poor visual appearance.

Additionally, there's no clear strategy for which components should be shortened first based on their importance to the user.

## Goals

- Implement a clear, predictable truncation strategy that prioritizes the most important information
- Handle edge cases gracefully: very long messages, long author names, many/long branch names, large git graphs
- Maintain visual balance (e.g., matching parentheses, clean truncation indicators)
- Ensure lines never exceed terminal width (no wrapping or escaped newlines in golden files)

## Non-Goals

- Complex name parsing logic (e.g., "Maximilian von Mustermann" → "M. v. M.")
- Dynamic per-component optimization algorithms
- User configuration of truncation preferences (for now)

## Key Requirements

### Truncation Priority (in order)

1. **Cap message at 72 chars** (handle paragraphs in commit titles)
2. **Truncate author at 25 chars** (simple string truncation with "...")
3. **Shorten refs** (three-level degradation, see below)
6. **Truncate author to 5 chars** (with "...")
4. **Drop time**
5. **Shorten message to 40 chars**
7. **Drop author**
8. **Continue truncating message from right** until fits

### Component Rules

**Mandatory (never truncate):**
- Selector indicator (`"> "` or `"  "`)
- Graph symbols (can be large with many branches - accepted)
- Commit hash (7 chars)
- Spacing/separators

**Author truncation:**
- 25 chars max initially (with "...")
- 5 chars max if still too long (with "...")
- Simple string truncation, no name parsing

**Refs truncation (three levels):**
1. First, shorten individual long ref names: `(main, this-is-a-very-long-branch-name, tag: v1.0)` → `(main, this-is-a-very-lo…, tag: v1.0)`
2. Then, show current branch + count: `(main +2 more)` or `(this-is-a… +2 more)`
3. Finally, just count: `(3 refs)`

**Message truncation:**
- 72 chars max initially
- 40 chars if space constrained
- Continue from right if still needed

### Visual Quality

- Refs must have balanced parentheses (or be omitted entirely)
- Use "..." for standard truncation
- Use "…" (ellipsis character) for refs truncation to save space
- All lines must fit exactly within terminal width

## User Impact

**Benefits:**
- More predictable and readable truncated lines
- Important info (message, branch) shown when possible
- Cleaner visual appearance

**Changes:**
- Lines with long refs will now show "(3 refs)" instead of cut-off branch names
- Very long author names will be truncated more aggressively
- Time may be hidden in narrow terminals
