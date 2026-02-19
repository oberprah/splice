# Commit Data Structure Research

Research into commit data availability and structure in the Splice codebase.

## GitCommit Struct Definition

**Location**: `internal/git/git.go` (lines 27-36)

```go
type GitCommit struct {
    Hash         string    // Full 40-char hash
    ParentHashes []string  // Parent commit hashes (empty for root commits)
    Refs         []RefInfo // Branch/tag decorations
    Message      string    // First line of commit message (subject)
    Body         string    // Commit message body (everything after subject line)
    Author       string    // Author name (not email)
    Date         time.Time // Commit timestamp
}
```

**All fields** are populated during the initial `FetchCommits()` call - they are NOT lazily loaded.

## RefInfo Structure

**Location**: `internal/git/git.go` (lines 11-25)

```go
type RefType int

const (
    RefTypeBranch       RefType = iota // Local branch (e.g., "main")
    RefTypeRemoteBranch                // Remote branch (e.g., "origin/main")
    RefTypeTag                         // Tag (e.g., "v1.0")
)

type RefInfo struct {
    Name   string  // e.g., "main", "v1.0", "origin/main"
    Type   RefType // Branch, RemoteBranch, or Tag
    IsHead bool    // true if this is the current HEAD
}
```

Refs are parsed by the `parseRefDecorations()` function from git's `%d` format and are stored directly in the `Refs` field of each commit.

## Message Structure

The commit message is split into two parts:

- **Message**: First line of commit message (subject line)
- **Body**: Everything after the subject line (commit message body)

These are parsed separately by git log using `%s` (subject) and `%b` (body) format specifiers.

## Data Availability Timeline

All commit data is available at the same time:

**Timeline**:
1. **App initialization** (`main.go`, line 66): `go run .` triggers `NewModel().Init()`
2. **Load command** (`internal/ui/app.go`, lines 64-69): Single async command calls `FetchCommits(500)` with a 500-commit limit
3. **Commits loaded** (`internal/ui/states/loading_update.go`, lines 20-56): When `CommitsLoadedMsg` is received:
   - All `GitCommit` structs are fully populated with all fields
   - Transition to `LogState` with all commit data ready
   - Graph layout is computed immediately
   - Preview loading is triggered asynchronously for the first commit

**Available BEFORE preview loads:**
- Hash ✓
- Author ✓
- Timestamp ✓
- Subject line (Message) ✓
- Body ✓
- Refs (branches, tags) ✓

**Available AFTER preview loads (separately):**
- File changes list (loaded via `FetchFileChanges()`)
- File stats (additions/deletions)

## LogState Structure

**Location**: `internal/ui/states/log_state.go` (lines 42-49)

```go
type LogState struct {
    Commits       []git.GitCommit
    Cursor        int
    ViewportStart int
    Preview       PreviewState
    GraphLayout   *graph.Layout
}
```

All commits in the `Commits` slice have **complete data** when the LogState is created. The `Preview` field tracks whether files have been loaded for the currently selected commit.

## Key Insight

**Refs are NOT loaded separately** - they are part of the initial git log query and come with every commit from the first fetch. The git command uses `%d` format specifier which outputs refs like `" (HEAD -> main, origin/main, tag: v1.0)"` and these are parsed into the `Refs` field for each commit.

The only data that loads asynchronously is the **file changes** (which populate the preview panel), but commit metadata, message, author, and refs are all immediately available.

---

**Research conducted**: 2026-01-06
