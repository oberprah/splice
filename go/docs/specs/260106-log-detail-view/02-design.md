# Design: Improve Log Detail View

## Executive Summary

The log detail view (right panel) and files view both display commit information with similar bugs: missing refs, faulty truncation, and unclear body truncation. Additionally, the log detail view flickers during navigation because it waits for file loading before showing metadata.

The solution introduces a shared **commit info component** that both views can use, fixing the bugs once for both. For the log detail view specifically, we decouple commit info rendering from file loading—showing all commit data immediately since it's already in memory, with only the file list waiting for async load.

## Context & Problem Statement

### Problems in Log Detail View

1. **Flickering** - Metadata line appears after files load, pushing subject down
2. **Truncation bug** - Metadata truncated even when under 80 chars
3. **Missing refs** - Branches/tags not shown (visible in left list though)
4. **Unclear body truncation** - Body silently cut at 5 lines with no indicator
5. **Misaligned separator** - Horizontal line has alignment issues
6. **Subject truncation** - Long subjects truncated instead of wrapped

### Same Problems in Files View

The files view (`files_view.go`) displays nearly identical commit information:

```
abc123d · John Doe committed 2 hours ago · 3 files · +45 -12

Subject line here

Body text here...
────────────────
[file list]
```

It shares the same bugs:
- Missing refs
- Same truncation issues (uses `RenderCommitMetadata()`)
- Same separator line

### Opportunity

Both views render commit info using overlapping code. Fixing these issues in both places would mean duplicating work. Instead, we should extract a shared component.

## Current State

### Data Structures

**GitCommit** (what we have for each commit):
```
GitCommit {
    Hash         string      // Full 40-char hash
    Refs         []RefInfo   // Branches, tags, HEAD ← already loaded!
    Message      string      // Subject line
    Body         string      // Body text
    Author       string
    Date         time.Time
}
```

**PreviewState** (file loading status in log view):
```
PreviewNone                      // Initial state
PreviewLoading { ForHash }       // Fetching files for commit X
PreviewLoaded { ForHash, Files } // Files ready
PreviewError { ForHash, Err }    // Load failed
```

### What's Loaded When

| Data | When Available | Where From |
|------|----------------|------------|
| Hash, Author, Date | App start | `FetchCommits()` → `LogState.Commits[]` |
| Subject, Body | App start | ↑ same |
| **Refs** | App start | ↑ same (already there!) |
| File list, stats | Per-commit async | `FetchFileChanges()` → `Preview` state |

**Key insight**: Everything except files is already in memory. The current code waits for files before showing metadata—that's unnecessary and causes flickering.

### Views Using Commit Info

**Log Detail View** (right panel in split view):
- Metadata line (currently waits for files—wrong!)
- Subject (truncated—should wrap)
- Body (max 5 lines, "..." indicator—should show count)
- Separator
- File list (async)

**Files View** (full-screen file browser):
- Metadata line
- Subject
- Body (full, no truncation)
- Separator
- File list (already loaded when entering this view)

Both use `RenderCommitMetadata()` for the metadata line. Neither shows refs.

## Solution

### Shared Components

We extract two shared components that both views use:

#### 1. Commit Info Component

> **Decision:** Extract a shared **commit info component** that renders: metadata line, refs line, subject, and body. Both views use the exact same component.
>
> **Rationale:**
> - Fixes bugs once for both views
> - Ensures visual consistency
> - Reduces maintenance burden
> - Natural boundary—commit info is conceptually separate from file list

**Parameters:**
- `commit: GitCommit` - The commit to display
- `width: int` - Panel width for wrapping/truncation
- `bodyMaxLines: int` - 0 for unlimited (files view), 5 for log detail view
- `ctx: Context` - For time formatting

**Structure:**
```
CommitInfo(commit, width, bodyMaxLines, ctx) → []string
    │
    ├── Metadata line: {hash} · {author} committed {time}
    ├── Refs line: main, origin/main, HEAD (if refs exist)
    ├── (blank)
    ├── Subject (wrapped if needed)
    ├── (blank, if body exists)
    └── Body (wrapped, truncated to bodyMaxLines with indicator)
```

Note: File stats are NOT in the commit info. They belong with the file list (see below).

#### 2. File Section Component

> **Decision:** Extract a shared **file section component** that renders: stats line and file list.

**Parameters:**
- `files: []FileChange` - Files to display
- `width: int` - Panel width
- `cursor: int` - Selected file index (-1 for no selection)
- `showSelector: bool` - Whether to show `>` selection indicator

**Output:** Blank line + stats line + all file lines. The leading blank line separates the file section from commit info above. No file list truncation—callers handle that.

Log detail view truncates to available space and adds overflow indicator ("... and N more files"). Files view uses its own viewport scrolling.

**Loading/Error States:** `FileSection` only renders when files are available. The caller (`log_view.go`) handles loading and error states directly, rendering "Loading files..." or "Unable to load files" instead of calling `FileSection`.

### File Organization

The shared components live in `internal/ui/states/` alongside the views that use them:

| File | Contents |
|------|----------|
| `commit_info.go` | New. `CommitInfo()` component |
| `file_section.go` | New. `FileSection()` component with all file rendering utilities |
| `commit_render.go` | Delete. Functions migrate to `file_section.go` or are removed |
| `log_view.go` | Uses both components |
| `files_view.go` | Uses both components |

> **Decision:** Keep components in `internal/ui/states/` rather than creating a new `components/` package.
>
> **Rationale:** The components are only used by these two views. No need for a separate package until we have more shared UI components.

> **Decision:** Merge file rendering utilities from `commit_render.go` into `file_section.go`, then delete `commit_render.go`.
>
> **Rationale:** Creates a deep module — `FileSection` has a simple interface but encapsulates all file rendering complexity internally. Functions like `FormatFileLine()`, `CalculateTotalStats()`, and `CalculateMaxStatWidth()` become private helpers. `RenderCommitMetadata()` and `TruncatePathFromLeft()` are deleted (replaced by `CommitInfo` and unused, respectively).

### How Each View Uses the Components

Both views compose the same two components:

```
┌──────────────────────────────────────────────┐
│ CommitInfo(commit, width, bodyMaxLines, ctx)
│                                              │
│ FileSection(files, width, cursor, showSelector)
└──────────────────────────────────────────────┘
```

**Log Detail View** (preview, read-only):
- `CommitInfo` with `bodyMaxLines=5`
- `FileSection` with `cursor=-1, showSelector=false`
- Handles loading/error states, truncates file list to available space

**Files View** (full, interactive):
- `CommitInfo` with `bodyMaxLines=0` (unlimited)
- `FileSection` with actual cursor, `showSelector=true`
- Uses viewport scrolling for long file lists

> **Decision:** Both views use identical components. Parameters handle contextual differences (preview vs. full, read-only vs. interactive).

### Why This Eliminates Flickering

**Current flow** (log detail view):
```
Navigate → Show subject + body (no metadata yet)
              ↓
         Files load (50-100ms)
              ↓
         Insert metadata line above subject ← LAYOUT SHIFT!
         Show files
```

**New flow:**
```
Navigate → Show CommitInfo immediately (all data in memory)
         + Show "Loading files..."
              ↓
         Files load (50-100ms)
              ↓
         Replace "Loading..." with stats + file list
         (CommitInfo unchanged—no shift)
```

The commit info section is stable from first render. Only the file section below it updates.

### Visual Layout

Both views have the same structure. The differences are: body length and file list interactivity.

**Log detail view (preview, body truncated):**
```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD

Use single ellipsis character for truncation

Replace triple-dot "..." (3 chars) with single ellipsis "…"
in message and author truncation. Saves 2 visual characters.
(... 3 more lines)

2 files · +39 -51
M +10 -10  internal/ui/states/log_line_format.go
M +29 -41  internal/ui/states/log_line_format_test.go
```

**Log detail view, loading state:**
```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD

Use single ellipsis character for truncation

Replace triple-dot "..." (3 chars) with single ellipsis "…"
in message and author truncation. Saves 2 visual characters.
(... 3 more lines)

Loading files...
```

**Files view (full body, interactive file list):**
```
fc5a4c3 · oberprah committed 9 hours ago
main, origin/main, HEAD

Use single ellipsis character for truncation

Replace triple-dot "..." (3 chars) with single ellipsis "…" (U+2026,
1 char) in message and author truncation. Saves 2 visual characters
per truncation, allowing 2 more content characters to display.

Changes:
- Updated capMessage() function
- Updated truncateAuthor() function

2 files · +39 -51
> M +10 -10  internal/ui/states/log_line_format.go
  M +29 -41  internal/ui/states/log_line_format_test.go
```

Both views now have identical structure. The file stats line naturally separates commit info from file list.

## Design Decisions

### 1. Metadata Line

> **Decision:** Metadata line contains only commit info: `{hash} · {author} committed {time}`. File stats move to the file section in both views.
>
> **Rationale:** Clean separation—commit info is always available, file stats come with files. Both views now render identical commit info. This also eliminates flickering in log detail view since commit info never waits for file data.

> **Decision:** Fix truncation by measuring plain text width before styling. Use `utf8.RuneCountInString()` instead of `len()`.
>
> **Rationale:** Current bug: `len()` counts bytes, not characters. Multi-byte UTF-8 characters (accented names, emoji) inflate the count. A 60-character line might report as 80+ bytes. Solution: use `utf8.RuneCountInString()` to count actual characters.

> **Decision:** Smart truncation priority: hash (keep) > time (keep) > author (truncate/drop).
>
> **Rationale:** Hash identifies the commit. Time gives crucial context. Author is also visible in left panel. Progressive levels:
> 1. Full: `{hash} · {author} committed {time}`
> 2. Truncate author: `{hash} · {auth…} committed {time}`
> 3. Drop "committed": `{hash} · {auth…} {time}`
> 4. Drop author: `{hash} · {time}`

### 2. Refs Display

> **Decision:** Show refs on a new line after metadata. Comma-separated, wrap if needed, no truncation. Monochrome styling.
>
> **Rationale:** Refs are already loaded—no reason not to show them. Wrapping is fine since most commits have 1-3 refs. Complete info more valuable than compact display. Monochrome keeps it simple; can add per-type colors later if needed.

### 3. Subject Line

> **Decision:** Wrap to multiple lines instead of truncating.
>
> **Rationale:** Most subjects are 50-72 chars (git best practice), fitting on one line. When longer, showing complete info is worth the extra line. We have vertical space.

### 4. Body Truncation

> **Decision:** Show `(... N more lines)` indicator when body exceeds limit. Count must account for wrapped lines.
>
> **Rationale:** Users should know content was truncated and roughly how much. "(... 2 more lines)" vs "(... 15 more lines)" sets different expectations.

### 5. Separator Line

> **Decision:** Remove the horizontal separator line from both views.
>
> **Rationale:** The file stats line naturally separates commit info from the file list. The separator added visual noise and had alignment issues. Removing it creates a cleaner, more consistent layout across both views.

## Open Questions

**None.** All design decisions have been made with clear rationale.

## References

- Requirements: `01-requirements.md`
- Research: `research/current-implementation.md`, `research/commit-data-structure.md`, `research/text-utilities.md`
- Key files:
  - `internal/ui/states/log_view.go` - Log detail rendering
  - `internal/ui/states/files_view.go` - Files view rendering
  - `internal/ui/states/commit_render.go` - Current `RenderCommitMetadata()`
