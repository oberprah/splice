# Current Implementation Research

Based on my analysis of the codebase, here's a comprehensive overview of how Splice currently handles the navigation flow:

## 1. Entry Point and Initialization

**File**: `/home/user/splice/main.go`

The application starts with a simple entry point that initializes the Bubbletea TUI framework:

```go
func main() {
    initialModel := app.NewModel(
        app.WithInitialState(loading.State{}),
    )
    p := tea.NewProgram(initialModel, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Printf("Error running program: %v\n", err)
        os.Exit(1)
    }
}
```

**Key points**:
- Creates a Model using functional options pattern
- Sets initial state to `loading.State{}`
- Runs in alternate screen mode (fullscreen)
- No CLI arguments are currently supported

## 2. Application Model and State Machine

**File**: `/home/user/splice/internal/app/model.go`

The application uses a navigation stack architecture:

```go
type Model struct {
    stack             []core.State  // Navigation stack
    width             int
    height            int
    fetchCommits      FetchCommitsFunc
    fetchFileChanges  core.FetchFileChangesFunc
    fetchFullFileDiff core.FetchFullFileDiffFunc
    nowFunc           func() time.Time
}
```

**Navigation Stack Pattern**:
- Current state is always `stack[len-1]`
- `pushState()` adds new states (replaces LoadingState instead of stacking)
- `PopScreenMsg` removes the top state and returns to previous
- Navigation is type-safe via specific message types

## 3. State Machine Flow

The complete navigation flow is:

```
LoadingState → LogState ⇄ FilesState ⇄ DiffState
     ↓
  (CommitsLoadedMsg)
     ↓
  LogState (shows commit list with preview panel)
     ↓
  (Enter on log entry or FilesLoadedMsg)
     ↓
  FilesState (shows list of changed files)
     ↓
  (Enter on file or DiffLoadedMsg)
     ↓
  DiffState (shows unified diff with line-by-line comparison)
```

### LoadingState
**Files**: `internal/ui/states/loading/state.go`, `internal/ui/states/loading/update.go`

- **Purpose**: Fetches commits on app initialization
- **Init message**: App.Init() runs the fetchCommits command
- **Data flow**:
  1. `Model.Init()` calls `FetchCommits(500)` - fetches up to 500 commits from git log
  2. Returns `CommitsLoadedMsg` with all commits
  3. On success, computes graph layout via `graph.ComputeLayout()`
  4. Transitions to LogState with:
     - All commits
     - Computed graph layout
     - Initial command to load preview for first commit
  5. On error, transitions to ErrorState

### LogState
**Files**: `internal/ui/states/log/state.go`, `internal/ui/states/log/update.go`, `internal/ui/states/log/view.go`

**Structure**:
```go
type State struct {
    Commits       []core.GitCommit
    Cursor        core.CursorState    // Normal or Visual mode
    ViewportStart int
    Preview       PreviewState        // PreviewNone, PreviewLoading, PreviewLoaded, or PreviewError
    GraphLayout   *graph.Layout
}
```

**Preview System** (sum type for different preview states):
- `PreviewNone`: No preview loaded
- `PreviewLoading`: Fetching files for a commit hash
- `PreviewLoaded`: Successfully loaded files for preview
- `PreviewError`: Error occurred while loading

**Key interactions**:
- **Navigation**:
  - `j/k` (up/down arrows): Move cursor, trigger preview load for new selection
  - `v`: Toggle visual mode for multi-commit selection
  - `q`: Exit visual mode or quit app
  - `Enter`: Load full files list (triggers `FilesLoadedMsg`)
  - `g/G`: Jump to first/last commit

- **Visual rendering**:
  - Split view when terminal width >= 160 chars (left: commit list, right: files preview)
  - Simple single-column view on narrower terminals
  - Graph visualization with parent/child relationships

**Preview Loading Pattern**:
```
User presses 'j' (move down)
  ↓
Cursor position changes
  ↓
New CommitRange is selected via GetSelectedRange()
  ↓
LoadPreview() command is triggered
  ↓
FetchFileChanges(commitRange) is called asynchronously
  ↓
FilesPreviewLoadedMsg is returned
  ↓
Preview state is updated (PreviewLoaded or PreviewError)
```

### FilesState
**Files**: `internal/ui/states/files/state.go`, `internal/ui/states/files/update.go`, `internal/ui/states/files/view.go`

**Structure**:
```go
type State struct {
    CommitRange   core.CommitRange
    Files         []core.FileChange
    Cursor        int
    ViewportStart int
}
```

**Key interactions**:
- **Navigation**:
  - `j/k`: Move cursor through files
  - `Enter`: Load diff for selected file (triggers `DiffLoadedMsg`)
  - `q`: Pop back to LogState
  - `g/G`: Jump to first/last file

**Diff Loading Pattern**:
```
User presses Enter on a file
  ↓
loadDiff(file) is called
  ↓
FetchFullFileDiff(commitRange, file) gets:
  - Old file content via git show <hash>^:<path>
  - New file content via git show <hash>:<path>
  - Unified diff via git diff <hash>^..<hash>
  ↓
BuildAlignedFileDiff() processes the diff:
  - Parses unified diff into hunks
  - Builds FileContent with syntax highlighting
  - Creates Alignments (UnchangedAlignment, ModifiedAlignment, etc.)
  - Calculates change indices for navigation
  ↓
DiffLoadedMsg is returned with:
  - CommitRange
  - File
  - AlignedFileDiff
  - ChangeIndices (indices of actual changes)
```

### DiffState
**Files**: `internal/ui/states/diff/state.go`, `internal/ui/states/diff/update.go`, `internal/ui/states/diff/view.go`

**Structure**:
```go
type State struct {
    CommitRange      core.CommitRange
    File             core.FileChange
    Diff             *diff.AlignedFileDiff
    ViewportStart    int
    CurrentChangeIdx int
    ChangeIndices    []int  // Indices of alignments with actual changes
}
```

**Display Format**:
- Split view: left (old version) | separator | right (new version)
- Each alignment represents one display row
- Lines can be: UnchangedAlignment, ModifiedAlignment, RemovedAlignment, AddedAlignment
- Color coding: red for deletions, green for additions, neutral for unchanged

**Key interactions**:
- `j/k`: Scroll line by line
- `ctrl+d/u`: Page down/up
- `n/N`: Jump to next/previous change
- `g/G`: Jump to top/bottom
- `q`: Pop back to FilesState

## 4. Data Structures

**Core Types** (`internal/core/git_types.go`):

```go
type GitCommit struct {
    Hash         string
    ParentHashes []string
    Refs         []RefInfo  // Branch/tag decorations
    Message      string     // First line only
    Body         string     // Full message body
    Author       string
    Date         time.Time
}

type FileChange struct {
    Path      string
    Status    string  // M, A, D, R, etc.
    Additions int
    Deletions int
    IsBinary  bool
}

type CommitRange struct {
    Start GitCommit
    End   GitCommit
    Count int  // Number of commits in range
}
```

**Diff Types** (`internal/domain/diff/alignment.go`):

```go
type AlignedFileDiff struct {
    Left       FileContent  // Old version
    Right      FileContent  // New version
    Alignments []Alignment  // Display rows
}

type FileContent struct {
    Path  string
    Lines []AlignedLine  // Lines with syntax highlighting tokens
}

// Alignment is a sum type:
type UnchangedAlignment struct { LeftIdx, RightIdx int }
type ModifiedAlignment struct { LeftIdx, RightIdx int; InlineDiff []diff.Diff }
type RemovedAlignment struct { LeftIdx int }
type AddedAlignment struct { RightIdx int }
```

## 5. Navigation Messages

**File**: `internal/core/navigation.go`

Type-safe navigation via message types:

```go
type PushLogScreenMsg struct {
    Commits     []GitCommit
    GraphLayout *graph.Layout
    InitCmd     tea.Cmd  // Optional initial command
}

type PushFilesScreenMsg struct {
    CommitRange CommitRange
    Files       []FileChange
}

type PushDiffScreenMsg struct {
    CommitRange   CommitRange
    File          FileChange
    Diff          *diff.AlignedFileDiff
    ChangeIndices []int
}

type PushErrorScreenMsg struct {
    Err error
}

type PopScreenMsg struct{}
```

## 6. Git Command Execution

**File**: `/home/user/splice/internal/git/git.go`

**FetchCommits** (called during Loading state):
```bash
git log --pretty=format:'%H\x00%P\x00%d\x00%an\x00%ad\x00%s\x00%b\x1e' \
    --date=iso-strict -n 500
```
- Fetches up to 500 commits with: hash, parents, refs, author, date, subject, body
- Uses NULL byte as field separator, record separator (0x1e) between commits

**FetchFileChanges** (called when navigating log or entering files state):
```bash
git diff --name-status <hash>^..<hash>
git diff --numstat <hash>^..<hash>
```
- Gets file status (added/modified/deleted/renamed)
- Gets line change counts (additions/deletions)
- Returns list of `FileChange` structs

**FetchFullFileDiff** (called when opening diff view):
```bash
git show <hash>^:<path>      # Old content
git show <hash>:<path>        # New content
git diff <hash>^..<hash> -- <path>  # Unified diff
```
- Fetches complete file content before and after
- Fetches unified diff for diff parsing
- Returns `FullFileDiffResult` with both contents and diff

**FetchFileContent**:
```bash
git show <hash>:<path>
```
- Retrieves full file content at a specific commit
- Returns empty string (no error) if file doesn't exist at that commit

## 7. Current CLI Arguments

**Status**: No CLI arguments are currently supported

The application always:
- Starts with LoadingState
- Fetches commits from current git repository
- Fetches 500 most recent commits
- Operates in fullscreen mode

Future direct diff view would require:
- Arguments for specifying commit/hash
- Optional file path argument
- Flag for showing diff directly without log view

## 8. Key Architecture Decisions

**State Pattern**: Each screen is a separate state implementing `core.State` interface
- View(ctx Context) ViewRenderer
- Update(msg tea.Msg, ctx Context) (State, tea.Cmd)

**Navigation Stack**: Typed messages for compile-time safety
- LoadingState is special - replaced instead of stacked
- Symmetric push/pop (except initial state)

**Async Data Loading**: Tea commands for non-blocking operations
- FetchFileChanges runs async and returns FilesLoadedMsg
- FetchFullFileDiff runs async and returns DiffLoadedMsg
- Stale response detection via range hash tracking in LogState

**Context Interface**: States access model through `core.Context`
- Width(), Height()
- FetchFileChanges(), FetchFullFileDiff()
- Now()

**Diff Pipeline** (in `internal/domain/diff/`):
1. Parse unified diff (hunks, changed lines)
2. Build file content with syntax highlighting (per-token highlighting)
3. Build alignments (pair lines, compute inline diffs, mark changes)
4. Calculate change indices for navigation

This structure makes it straightforward to extend with direct diff viewing - you could add a `DirectDiffState` that bypasses LogState and FilesState, directly loading a diff from specified commit hash and file path.
