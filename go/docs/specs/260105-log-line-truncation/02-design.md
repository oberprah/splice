# Design: Log Line Truncation Strategy

## Executive Summary

This design implements a predictable, progressive truncation strategy for git log lines that prioritizes the most important information (commit message) while handling edge cases gracefully. The solution replaces the current fixed-ratio truncation with a 10-level progressive strategy that degrades refs through three separate levels (shorten individual names → show first + count → show total count only), progressively shortens author names (25 chars → 5 chars → drop), drops time when needed, and continues truncating the message until the line fits.

The implementation modifies the `formatCommitLine()` function in `internal/ui/states/log_view.go` to apply truncation rules in sequence, measuring line width after each level until the line fits within terminal width. All truncation logic is implemented as **pure functions** (no side effects, deterministic) to enable comprehensive unit testing. This ensures predictable behavior across all terminal sizes while maintaining visual quality (balanced parentheses, clean ellipsis placement).

**Key decisions**: (1) Apply truncation levels sequentially rather than pre-calculating optimal distribution, for simplicity and predictability. (2) Make all functions pure by extracting side effects to the caller - `formatCommitLine()` receives a `CommitLineComponents` struct with pre-computed values, enabling testing without mocks. (3) Use simple string operations initially (not ANSI-aware measurement), since current implementation doesn't style individual components before assembly. (4) Implement refs truncation as a three-level state machine for clarity.

## Context & Problem Statement

When displaying git commit log lines, various components (hash, refs, message, author, time) can exceed available terminal width. The current implementation truncates message and author using a fixed 2:1 ratio allocation, treating all other components as "fixed width." This creates problems:

- Long branch names in refs cause unbalanced parentheses: `(feature/very-long-branch-name-that-descr...`
- No graceful degradation for refs (all or nothing)
- Time is never dropped even in severely narrow terminals
- Message/author get squeezed into tiny spaces when refs or graph are large

Users need a predictable truncation strategy that prioritizes information importance while handling edge cases (very narrow terminals, many/long refs, large graphs) gracefully.

**Scope**: This design covers the truncation strategy for log lines in the LogState view. It does not cover the details panel in split view, which already has separate rendering logic.

## Current State

### Current Truncation Logic

The `formatCommitLine()` function (internal/ui/states/log_view.go:135-220) currently:

1. Treats selector, graph, hash, refs, separator, time prefix, and time as "fixed width"
2. Calculates remaining width for message and author
3. Allocates remaining width: 2/3 to message, 1/3 to author
4. Truncates each with simple string slicing + "..." if over limit

**Problems**:
- Refs can be arbitrarily long but are treated as "fixed"
- No component priority - all "fixed" components get full space regardless of importance
- Time never drops even when space is critical
- Author truncation is single-level (no progressive shortening or dropping)
- Message isn't capped initially - very long messages compete proportionally with author

### Component Structure

A log line consists of these components (in order):
```
[selector][graph][hash] [refs] [message] - [author] [time]
```

Example:
```
> ├ abc123d (HEAD -> main, tag: v1.0) Implement user auth - Alice Johnson (2 days ago)
```

**Width characteristics**:
- Selector: Fixed 2 chars (`"> "` or `"  "`)
- Graph: Variable (2 chars per symbol, e.g., `"│ "`, `"├─╮ "`)
- Hash: Fixed 7 chars
- Refs: Variable (0 chars if none, otherwise `(...)` with trailing space)
- Message: Variable (current: proportional allocation)
- Separator: Fixed 3 chars (`" - "`)
- Author: Variable (current: proportional allocation)
- Time: Variable (e.g., `"just now"`, `"2 days ago"`, `"Jan 2, 2006"`)

## Solution

### Overview

Replace the current fixed-ratio truncation with a 10-level progressive strategy that applies truncation rules in priority order until the line fits within terminal width.

**Core approach**:
1. Start with all components at full/initial size
2. Measure total width
3. If too wide, apply the next truncation level
4. Repeat until the line fits or all levels exhausted

This creates predictable, consistent behavior: users always see the same truncation at a given terminal width, regardless of content.

### Truncation Priority Levels

These levels are applied sequentially in order:

1. **Cap message at 72 chars** - Prevents very long commit messages from dominating
2. **Truncate author to 25 chars** - First-level author shortening with "..."
3. **Shorten refs Level 1** - Truncate individual long ref names with "…"
4. **Shorten refs Level 2** - Show first ref + count (e.g., "main +2 more")
5. **Shorten refs Level 3** - Show total count only (e.g., "3 refs")
6. **Truncate author to 5 chars** - Second-level author shortening (minimal but present)
7. **Drop time** - Remove time component entirely (space less critical than author)
8. **Shorten message to 40 chars** - Significant message reduction while keeping core info
9. **Drop author** - Remove author entirely (message is most important)
10. **Truncate entire line** - Drop refs, assemble minimal line, truncate to fit (handles extreme cases like 20 parallel branches)

> **Decision:** We chose sequential truncation over optimal distribution because it's simpler to implement, easier to test, and provides predictable behavior. Users can reason about what they'll see at different terminal widths. The tradeoff is less optimal space usage in some cases, but the benefit of simplicity and predictability outweighs this.

### Refs Truncation (3 Levels)

Refs have special handling with three degradation levels:

**Level 1 - Shorten individual ref names:**
```
(HEAD -> feature/implement-advanced-user-auth, origin/feature/implement-advanced-user-auth, tag: v2.1.0)
↓
(HEAD -> feature/implement-adva…, origin/feature/implement-adva…, tag: v2.1.0)
```
- Truncate individual ref names that exceed a threshold (e.g., 30 chars)
- Use "…" (single ellipsis char) to save space vs "..."
- Maintain balanced parentheses

**Level 2 - Show first ref + count:**
```
(HEAD -> feature/implement-adva…, origin/feature/implement-adva…, tag: v2.1.0)
↓
(HEAD -> feature/implement-adva… +2 more)
```
- Show only the first ref (current branch if present, otherwise first in list)
- Append "+ N more" to indicate additional refs
- Still apply individual name truncation if needed

**Level 3 - Show total count only:**
```
(HEAD -> feature/implement-adva… +2 more)
↓
(3 refs)
```
- Show only `(N refs)`
- Minimal space usage while indicating refs exist

> **Decision:** We chose three-level degradation for refs because it balances information preservation with space savings. Level 1 keeps all ref names visible (important for multi-branch workflows), Level 2 preserves current branch + count (most common need), Level 3 provides minimal indication. The tradeoff is implementation complexity vs a simpler two-level approach, but the middle level (first ref + count) is valuable in practice for medium-width terminals.

### Implementation Structure

**Design Principle: Pure Functions for Testability**

All truncation logic will be implemented as pure functions (no side effects, deterministic output for given input). This enables:
- Unit testing individual truncation functions in isolation
- Predictable, reproducible behavior
- Easy reasoning about correctness

**Function Purity Classification:**

All functions are pure (no side effects, deterministic):
- `capMessage(message string, maxLen int) string`
- `truncateAuthor(author string, maxLen int) string`
- `truncateEntireLine(line string, maxWidth int) string`
- `buildRefs(refs []git.RefInfo, level RefsLevel) string`
- `measureLineWidth(selector, graph, hash, refs, message, author, time string) int`
- `assembleLine(selector, graph, hash, refs, message, author, time string, isSelected bool) string`
- `formatCommitLine(components CommitLineComponents, availableWidth int, isSelected bool) string`

To make `formatCommitLine()` pure, we introduce a struct that holds all pre-computed components:

```go
type CommitLineComponents struct {
    Selector string
    Graph    string
    Hash     string
    Refs     []git.RefInfo
    Message  string
    Author   string
    Time     string
}
```

The caller (impure) prepares these components, then `formatCommitLine()` (pure) applies truncation logic. This makes the entire truncation pipeline testable without mocks.

The implementation changes the signature and makes it pure:

```go
// Pure function - all inputs provided, no side effects
func formatCommitLine(components CommitLineComponents, availableWidth int, isSelected bool) string {
    // 1. Extract components (already computed by caller)
    selector := components.Selector
    graph := components.Graph
    hash := components.Hash
    message := components.Message
    author := components.Author
    time := components.Time

    // Build refs at full level initially
    refs := buildRefs(components.Refs, RefsLevelFull)

    // 2. Apply truncation levels sequentially
    level := 0
    for measureLineWidth(...) > availableWidth && level < 10 {
        switch level {
        case 0:
            message = capMessage(message, 72)
        case 1:
            author = truncateAuthor(author, 25)
        case 2:
            refs = buildRefs(components.Refs, RefsLevelShortenIndividual)
        case 3:
            refs = buildRefs(components.Refs, RefsLevelFirstPlusCount)
        case 4:
            refs = buildRefs(components.Refs, RefsLevelCountOnly)
        case 5:
            author = truncateAuthor(author, 5)
        case 6:
            time = ""
        case 7:
            message = capMessage(message, 40)
        case 8:
            author = ""
        case 9:
            // Drop refs if present, assemble minimal line, then truncate entire line to fit
            refs = ""
            assembledLine := assembleLine(selector, graph, hash, refs, message, author, time, isSelected)
            if len(assembledLine) > availableWidth {
                return truncateEntireLine(assembledLine, availableWidth)
            }
            return assembledLine
        }
        level++
    }

    // 3. Assemble and style the line
    return assembleLine(selector, graph, hash, refs, message, author, time, isSelected)
}

// Caller (impure) prepares components - example from LogState.View()
func (s LogState) buildCommitListColumn(width int, ctx Context) *ViewBuilder {
    // ... viewport logic ...
    for i := 0; i < ctx.Height(); i++ {
        commit := s.Commits[logIdx]

        // Prepare all components (this is where impure operations happen)
        components := CommitLineComponents{
            Selector: buildSelector(logIdx == s.Cursor),
            Graph:    s.buildGraphForCommit(logIdx),
            Hash:     format.ToShortHash(commit.Hash),
            Refs:     commit.Refs,
            Message:  commit.Message,
            Author:   commit.Author,
            Time:     format.ToRelativeTimeFrom(commit.Date, ctx.Now()),
        }

        // Call pure function with all components
        line := formatCommitLine(components, width, logIdx == s.Cursor)
        vb.AddLine(colStyle.Render(line))
    }
    return vb
}
```

**Key aspects**:
- `formatCommitLine()` is now completely pure - receives all inputs, no side effects
- Caller handles impure operations (time formatting, graph lookup)
- `measureLineWidth()` calculates total width including all spacing
- `buildRefs()` takes a level parameter (Full, ShortenIndividual, FirstPlusCount, CountOnly)
- Refs degrade through 3 separate levels (cases 2, 3, 4) for progressive space savings
- Level 9 handles extreme cases: drops refs, assembles line, then truncates entire line if needed
- `truncateEntireLine()` hard-truncates the assembled string to fit (handles huge graphs)
- `assembleLine()` handles final spacing, separator logic, and styling

### Component Details

#### Message Truncation

**Initial cap (72 chars)**:
```go
func capMessage(message string, maxLen int) string {
    if len(message) <= maxLen {
        return message
    }
    return message[:maxLen-3] + "..."
}
```

**Entire line truncation (extreme cases)**:

For level 9, when even mandatory components (selector + large graph + hash) consume most available width:

```go
func truncateEntireLine(line string, maxWidth int) string {
    if len(line) <= maxWidth {
        return line
    }

    if maxWidth < 3 {
        // Terminal too narrow, show what we can
        if maxWidth <= 0 {
            return ""
        }
        return line[:maxWidth]
    }

    // Truncate with ellipsis
    return line[:maxWidth-3] + "..."
}
```

This handles extreme cases like 20 parallel branches where the graph itself may be 40+ characters.

#### Author Truncation

Simple string truncation with "...":
```go
func truncateAuthor(author string, maxLen int) string {
    if len(author) <= maxLen {
        return author
    }
    if maxLen < 3 {
        return ""
    }
    return author[:maxLen-3] + "..."
}
```

#### Refs Building

```go
type RefsLevel int

const (
    RefsLevelFull RefsLevel = iota
    RefsLevelShortenIndividual
    RefsLevelFirstPlusCount
    RefsLevelCountOnly
)

func buildRefs(refs []git.RefInfo, level RefsLevel) string {
    if len(refs) == 0 {
        return ""
    }

    switch level {
    case RefsLevelFull:
        return formatRefsFull(refs)
    case RefsLevelShortenIndividual:
        return formatRefsShortenedIndividual(refs, 30)
    case RefsLevelFirstPlusCount:
        return formatRefsFirstPlusCount(refs, 30)
    case RefsLevelCountOnly:
        return fmt.Sprintf("(%d refs)", len(refs))
    }
}

func formatRefsFull(refs []git.RefInfo) string {
    // Current implementation - join all refs with ", "
    // Returns: "(HEAD -> main, origin/main, tag: v1.0) "
}

func formatRefsShortenedIndividual(refs []git.RefInfo, maxLen int) string {
    // Truncate each ref name > maxLen with "…"
    // Returns: "(HEAD -> main, origin/mai…, tag: v1.0) "
}

func formatRefsFirstPlusCount(refs []git.RefInfo, maxLen int) string {
    // Show first ref (prefer current branch) + count
    // Returns: "(HEAD -> main +2 more) " or "(main +2 more) "
}
```

> **Decision:** We chose to implement refs truncation as a state machine with explicit levels rather than a more dynamic approach (e.g., progressively removing refs one at a time) because it provides predictable behavior and simpler code. The tradeoff is less granularity in space optimization, but three levels are sufficient for practical cases.

#### Width Measurement

Initially, use simple `len()` for width measurement since components aren't styled before assembly:

```go
func measureLineWidth(selector, graph, hash, refs, message, author, time string) int {
    width := len(selector) + len(graph) + len(hash)

    if refs != "" {
        width += 1 + len(refs)  // space + refs
    }

    width += len(message)

    if author != "" {
        width += 3 + len(author)  // " - " + author
    }

    if time != "" {
        width += 1 + len(time)  // space + time (includes time prefix)
    }

    return width
}
```

**Future consideration**: If we add per-component styling in the future (colors applied before assembly), we'll need to switch to `ansi.StringWidth()` from the `github.com/charmbracelet/x/ansi` package to measure visible width excluding ANSI escape codes.

> **Decision:** We chose to keep using `len()` initially because the current implementation applies styling after assembly, so components are plain strings during truncation. Switching to ANSI-aware measurement would add complexity without immediate benefit. The tradeoff is that future refactoring (if we add per-component styling earlier) would require updating this code, but that's acceptable given current needs.

### Data Flow Diagram

```
Caller prepares CommitLineComponents
        ↓
formatCommitLine(components, width, isSelected)
        ↓
Extract component strings from struct
        ↓
Build refs at full level
        ↓
    ┌───────────────────┐
    │ Measure width     │
    └───────────────────┘
            ↓
    ┌───────────────────┐
    │ Fits?             │──Yes──→ Skip to assembly
    └───────────────────┘
            ↓ No
    ┌───────────────────┐
    │ Apply next level  │
    │ (switch on level) │
    └───────────────────┘
            ↓
    ┌───────────────────┐
    │ level++           │
    └───────────────────┘
            ↓
    (loop until fits or level >= 10)
            ↓
    ┌───────────────────┐
    │ Assemble line     │
    │ - add spacing     │
    │ - add separator   │
    │ - apply styling   │
    └───────────────────┘
            ↓
        return string
```

### Edge Case Handling

**Very narrow terminals (< 40 chars)**:
- All truncation levels applied
- Message truncated to minimal size
- Still show: selector, graph, hash, minimal message
- Example (30 chars): `> ├ abc123d Impl...`

**No refs**:
- Refs string is empty, skipped in assembly
- More space available for message/author

**Very long graph (e.g., 20 parallel branches)**:
- Graph width accepted as mandatory (never truncated)
- Other components compete for remaining space
- In extreme cases (graph + hash + selector > availableWidth), level 9 truncates entire line
- Example (normal): `> ├─┬─╮ abc123d Fix... - Alice`
- Example (extreme, 30 chars with huge graph): `> ├─┬─╮─┬─╮─┬─╮ abc123d ...`

**Empty message**:
- Show empty string for message
- Don't show separator/author/time if message is empty

**Time dropped**:
- Simply skip time component in assembly
- Adjust spacing logic (no trailing space after author)

### Visual Quality Rules

1. **Balanced parentheses**: Refs always have matching `()` or are omitted entirely
2. **Ellipsis consistency**:
   - Use "..." for message and author (3 chars, standard)
   - Use "…" for refs (1 char, space-saving)
3. **Separator logic**: Only show " - " separator if both message and author are present
4. **Trailing spaces**: No trailing spaces in final line

## Testing Strategy

All functions are pure and straightforward to test:
- **Unit tests**: Table-driven tests for each truncation function
- **Integration tests**: `formatCommitLine()` with various widths (no mocks needed - all pure)
- **Golden file tests**: Visual regression testing at different terminal widths

Test files are colocated with implementation in `internal/ui/states/log_view_test.go`.

## Open Questions

There are no open questions. The design is complete and ready for implementation.

The truncation strategy is fully specified in requirements, the existing codebase patterns are clear from research, and all implementation decisions have been made with documented rationale.
