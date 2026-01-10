# Design: Direct Diff View

## Executive Summary

This design adds a `splice diff <spec>` subcommand that allows users to view any git diff directly, bypassing the log view. The solution introduces a sealed interface pattern (`DiffSource`) to represent both commit ranges and uncommitted changes, uses manual CLI argument parsing, and reuses existing files/diff view components with minimal modifications.

**Key decisions:**
- **Data model**: Replace `CommitRange` with sealed `DiffSource` interface (matches existing `Alignment` pattern)
- **CLI parsing**: Manual parsing in main.go (~50 lines, zero dependencies)
- **State flow**: New `DirectDiffLoadingState` → `FilesState` (with exit-on-quit behavior) → `DiffState`
- **Git commands**: New functions for uncommitted changes, abstracted through `DiffSource` interface

The design maintains backward compatibility (`splice` continues to work), reuses all existing diff viewing UI, and follows the codebase's principle of "make illegal states unrepresentable."

## Context & Problem Statement

Currently, Splice only supports viewing commit history. Users must start with the log view, select a commit, then view its files and diffs. However, many git workflows involve viewing diffs that don't correspond to historical commits:
- Uncommitted changes (staged/unstaged)
- Branch comparisons (`main..feature`)
- Arbitrary commit ranges (`HEAD~5..HEAD`)

These workflows force users to use other tools, breaking their flow.

**Scope:** This design covers adding `splice diff <spec>` to view any git diff directly. It does NOT cover:
- Changing existing `splice` (log view) behavior
- Adding new diff visualization features
- Direct jump to specific file diffs
- File-to-file navigation within diff view

## Current State

### Architecture

Splice uses a state machine with navigation stack:
```
LoadingState → LogState ⇄ FilesState ⇄ DiffState
```

Each state implements:
```go
type State interface {
    View(ctx Context) ViewRenderer
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
```

Navigation uses typed messages (`PushFilesScreenMsg`, `PushDiffScreenMsg`, `PopScreenMsg`).

### Data Flow

1. **LoadingState**: Fetches commits → transitions to LogState
2. **LogState**: Shows commit list, user selects → transitions to FilesState
3. **FilesState**: Shows file list, user selects → transitions to DiffState
4. **DiffState**: Shows side-by-side diff, user can go back or quit

### Current Constraints

1. **CommitRange everywhere**: All states use `CommitRange{Start, End GitCommit, Count int}`
2. **No CLI arguments**: main.go hardcodes `LoadingState` as entry point
3. **Git commands assume commits**: `git diff <hash>^..<hash>` pattern requires actual commits

## Solution

### High-Level Architecture

```
main.go (parse CLI args)
    ↓
  ┌─────────────────┬──────────────────┐
  │                 │                  │
splice          splice diff       splice diff <spec>
  │                 │                  │
  ↓                 ↓                  ↓
LoadingState  DirectDiffLoadingState  DirectDiffLoadingState
  ↓                 ↓                  ↓
LogState        FilesState          FilesState
  ↓                 ↓                  ↓
FilesState      DiffState           DiffState
  ↓
DiffState
```

**Key insight**: `FilesState` and `DiffState` work unchanged; we just change the entry point and data model.

### Component Interactions

```
┌──────────────┐
│   main.go    │ Parse args → create initial state
└──────┬───────┘
       │
       ↓
┌──────────────────────────┐
│ DirectDiffLoadingState   │ Fetch files for DiffSource
└──────┬───────────────────┘
       │ (FilesLoadedMsg)
       ↓
┌──────────────────────┐
│    FilesState        │ Display files, handle selection
│ - DiffSource         │ - Uses DiffSource for header
│ - exitOnPop = true   │ - Exits instead of popping
└──────┬───────────────┘
       │ (DiffLoadedMsg)
       ↓
┌──────────────────────┐
│     DiffState        │ Display side-by-side diff
│ - DiffSource         │ - Same as current behavior
│ - File               │
│ - AlignedFileDiff    │
└──────────────────────┘
```

## Design Decisions

### Decision 1: Data Model - Sealed Interface Pattern

**Decision:** Replace `CommitRange` with a sealed `DiffSource` interface. This is a sum type with two implementations: `CommitRangeDiffSource` and `UncommittedChangesDiffSource`.

**Alternatives considered:**
1. **Sealed interface** (chosen)
2. Extended `CommitRange` with optional fields
3. Wrapper type with tagged union

**Rationale:**
- **Type safety**: Compiler prevents invalid states (vs extended CommitRange which allows 6+ invalid combinations)
- **Precedent**: Codebase already uses this pattern for `Alignment` type
- **Extensibility**: Future diff sources (stash, cherry-pick) add cleanly
- **Self-documenting**: Code explicitly shows valid diff source types
- **Follows principle**: "Make illegal states unrepresentable" (from CLAUDE.md)

**Tradeoff:** Requires updating more files initially, but prevents entire classes of runtime bugs.

**Structure:**
```go
// internal/core/diff_source.go

type DiffSource interface {
    diffSource() // Sealed marker
}

type CommitRangeDiffSource struct {
    Start GitCommit
    End   GitCommit
    Count int
}

type UncommittedChangesDiffSource struct {
    Type UncommittedType
}

type UncommittedType int
const (
    UncommittedTypeUnstaged UncommittedType = iota
    UncommittedTypeStaged
    UncommittedTypeAll
}

func (CommitRangeDiffSource) diffSource() {}
func (UncommittedChangesDiffSource) diffSource() {}
```

**Usage pattern:**
```go
func DisplayHeader(source DiffSource) string {
    switch s := source.(type) {
    case CommitRangeDiffSource:
        return fmt.Sprintf("%s..%s", shortHash(s.Start), shortHash(s.End))
    case UncommittedChangesDiffSource:
        switch s.Type {
        case UncommittedTypeUnstaged:
            return "Uncommitted changes (unstaged)"
        case UncommittedTypeStaged:
            return "Uncommitted changes (staged)"
        case UncommittedTypeAll:
            return "Uncommitted changes"
        }
    }
    return ""
}
```

### Decision 2: CLI Parsing - Manual Implementation

**Decision:** Use manual argument parsing in main.go rather than a CLI framework.

**Alternatives considered:**
1. **Manual parsing** (chosen)
2. Cobra framework
3. Standard library `flag` package

**Rationale:**
- Aligns with Splice's "lean" philosophy
- Only 2 commands needed (log, diff)
- ~50 lines of clear, testable code
- Zero new dependencies
- Full control over parsing behavior

**Tradeoff:** More code to write and maintain vs pulling in a framework, but matches project values.

**Implementation:**
```go
func parseArgs() (cmd string, args []string) {
    if len(os.Args) < 2 {
        return "log", []string{}
    }
    if os.Args[1] == "diff" {
        return "diff", os.Args[2:]
    }
    return "log", []string{} // or error
}
```

Spec parsing validates:
- `splice diff` → unstaged
- `splice diff --staged` or `--cached` → staged
- `splice diff HEAD` → all uncommitted
- `splice diff <range>` → commit range (e.g., `main..feature`)

### Decision 3: State Architecture - New DirectDiffLoadingState

**Decision:** Create `DirectDiffLoadingState` separate from `LoadingState`, which fetches files for a `DiffSource` instead of commits.

**Alternatives considered:**
1. **New DirectDiffLoadingState** (chosen)
2. Extend existing LoadingState with optional diff source
3. Skip loading state and go directly to FilesState

**Rationale:**
- **Separation of concerns**: Loading commits vs loading file changes are different operations
- **Async pattern**: Maintains existing pattern of loading state → data state
- **Error handling**: Can show loading spinner and handle errors consistently
- **Code reuse**: LoadingState is ~50 lines, mostly boilerplate

**Tradeoff:** One more state type, but cleaner separation and consistent UX.

**Data flow:**
```
DirectDiffLoadingState.Init()
  → FetchFileChangesCmd(diffSource)
    → async git operation
      → FilesLoadedMsg
        → Push FilesState or ErrorState
```

### Decision 4: FilesState Navigation - Exit on Pop Flag

**Decision:** Add `exitOnPop bool` field to `FilesState`. When true, pressing `q` exits the app instead of popping back to log view.

**Alternatives considered:**
1. **ExitOnPop flag** (chosen)
2. Separate DirectFilesState type
3. Check stack depth in Update

**Rationale:**
- **Minimal change**: One boolean field vs entire new state type
- **Explicit**: Behavior is clear in constructor
- **Testable**: Easy to verify with unit tests
- **Precedent**: Similar pattern used in LogState cursor modes

**Tradeoff:** FilesState has slightly more complexity, but avoids code duplication.

**Implementation:**
```go
// internal/ui/states/files/state.go
type State struct {
    Source       core.DiffSource  // Changed from CommitRange
    Files        []core.FileChange
    Cursor       int
    ViewportStart int
    ExitOnPop    bool  // New field
}

// In Update method:
case tea.KeyMsg:
    if key.String() == "q" {
        if s.ExitOnPop {
            return s, tea.Quit
        }
        return s, func() tea.Msg { return core.PopScreenMsg{} }
    }
```

### Decision 5: Git Commands - Source-Specific Functions

**Decision:** Add new git functions for uncommitted changes, accessed through `DiffSource` interface methods.

**Alternatives considered:**
1. **Source-specific functions** (chosen)
2. Parallel `FetchFileChangesFromSpec(string)` functions
3. Modify existing functions with conditionals

**Rationale:**
- **Type safety**: DiffSource determines which git commands to use
- **Clarity**: Separate functions for different workflows
- **Testability**: Easy to mock and test each case
- **Maintainability**: Clear what commands handle what scenarios

**Tradeoff:** More functions in git.go, but each is focused and clear.

**New functions:**
```go
// internal/git/git.go

// For file lists
func FetchFileChangesForCommitRange(start, end GitCommit) ([]FileChange, error)
func FetchUnstagedFileChanges() ([]FileChange, error)
func FetchStagedFileChanges() ([]FileChange, error)
func FetchAllUncommittedFileChanges() ([]FileChange, error)

// For file diffs
func FetchFileDiffForCommitRange(start, end GitCommit, file FileChange) (*FullFileDiffResult, error)
func FetchUnstagedFileDiff(file FileChange) (*FullFileDiffResult, error)
func FetchStagedFileDiff(file FileChange) (*FullFileDiffResult, error)
func FetchAllUncommittedFileDiff(file FileChange) (*FullFileDiffResult, error)
```

**Git commands:**
```bash
# Unstaged file list
git diff --name-status
git diff --numstat

# Staged file list
git diff --staged --name-status
git diff --staged --numstat

# All uncommitted file list
git diff HEAD --name-status
git diff HEAD --numstat

# Unstaged file content
git show :path           # Old (from index)
cat path                 # New (from working tree)
git diff -- path         # Unified diff

# Staged file content
git show HEAD:path       # Old (from HEAD)
git show :path           # New (from index)
git diff --staged -- path # Unified diff

# All uncommitted file content
git show HEAD:path       # Old (from HEAD)
cat path                 # New (from working tree)
git diff HEAD -- path    # Unified diff
```

### Decision 6: Error Handling - Validate Before TUI

**Decision:** Validate diff spec in main.go before entering TUI. Invalid specs or empty diffs show error message and exit immediately.

**Alternatives considered:**
1. **Validate in main.go** (chosen)
2. Show error in TUI (ErrorState)
3. Validate in loading state

**Rationale:**
- **Fail fast**: User gets immediate feedback without TUI overhead
- **Simpler UX**: Clear error message on terminal vs full-screen error state
- **Consistency**: Matches git CLI behavior (quick errors for bad input)

**Tradeoff:** Can't use TUI error display, but errors are clearer.

**Validation:**
```go
func validateDiffSpec(spec DiffSource) error {
    // Run: git diff --quiet <args>
    // Exit 0 = no changes → error
    // Exit 1 = has changes → ok
    // Exit 128+ = invalid spec → error
}
```

## Data Flow Diagrams

### CLI Argument Flow

```
User runs: splice diff main..feature
                    ↓
           parseArgs() in main.go
                    ↓
            cmd = "diff"
            args = ["main..feature"]
                    ↓
         parseDiffSpec(args)
                    ↓
     DiffSource = CommitRangeDiffSource{
         Start: <main commit>,
         End: <feature commit>,
         Count: N
     }
                    ↓
         validateDiffSpec()
                    ↓
    DirectDiffLoadingState{source}
                    ↓
              tea.Program
```

### Uncommitted Changes Flow

```
User runs: splice diff
                    ↓
        DiffSource = UncommittedChangesDiffSource{
            Type: UncommittedTypeUnstaged
        }
                    ↓
    DirectDiffLoadingState
                    ↓
        Init() → FetchFileChanges(source)
                    ↓
         Type switch on DiffSource
                    ↓
     FetchUnstagedFileChanges()
         ├─ git diff --name-status
         └─ git diff --numstat
                    ↓
         FilesLoadedMsg{files}
                    ↓
     FilesState{source, files, exitOnPop=true}
                    ↓
        User selects file
                    ↓
     FetchUnstagedFileDiff(file)
         ├─ git show :file (old)
         ├─ cat file (new)
         └─ git diff -- file (unified)
                    ↓
         DiffLoadedMsg{diff}
                    ↓
     DiffState{source, file, diff}
```

### State Transitions

```
                     main.go
                        │
        ┌───────────────┴────────────────┐
        │                                │
   splice (no args)              splice diff <spec>
        │                                │
        ↓                                ↓
  LoadingState              DirectDiffLoadingState
        ↓                                ↓
  (CommitsLoadedMsg)          (FilesLoadedMsg)
        ↓                                ↓
    LogState            ┌────────────────┘
        ↓               │
   (FilesLoadedMsg)     ↓
        │           FilesState (exitOnPop=true)
        └───────────────┤
                        ↓
                   (DiffLoadedMsg)
                        ↓
                    DiffState
                        ↓
                    (q pressed)
                        ↓
                ┌───────┴───────┐
                │               │
         exitOnPop=false   exitOnPop=true
                │               │
                ↓               ↓
           (PopScreenMsg)   (tea.Quit)
                │               │
                ↓               ↓
           FilesState        Exit app
```

## Type Definitions

### Core Types

```go
// internal/core/diff_source.go

// DiffSource represents any source of a git diff (sum type)
type DiffSource interface {
    diffSource()
}

// CommitRangeDiffSource: diff between two commits
type CommitRangeDiffSource struct {
    Start GitCommit
    End   GitCommit
    Count int // Number of commits in range
}

// UncommittedChangesDiffSource: diff of uncommitted changes
type UncommittedChangesDiffSource struct {
    Type UncommittedType
}

type UncommittedType int

const (
    UncommittedTypeUnstaged UncommittedType = iota // Working tree vs index
    UncommittedTypeStaged                          // Index vs HEAD
    UncommittedTypeAll                             // Working tree vs HEAD
)

// Marker methods
func (CommitRangeDiffSource) diffSource() {}
func (UncommittedChangesDiffSource) diffSource() {}
```

### Navigation Messages

```go
// internal/core/navigation.go

// Updated: CommitRange → DiffSource
type PushFilesScreenMsg struct {
    Source DiffSource      // Changed from CommitRange
    Files  []FileChange
    ExitOnPop bool         // New field
}

// Updated: CommitRange → DiffSource
type PushDiffScreenMsg struct {
    Source        DiffSource  // Changed from CommitRange
    File          FileChange
    Diff          *diff.AlignedFileDiff
    ChangeIndices []int
}
```

### State Types

```go
// internal/ui/states/directdiff/state.go (new file)

// DirectDiffLoadingState loads file changes for a DiffSource
type State struct {
    Source core.DiffSource
}

func New(source core.DiffSource) State {
    return State{Source: source}
}

func (s State) Init(ctx core.Context) tea.Cmd {
    return fetchFileChangesCmd(s.Source, ctx.FetchFileChanges())
}
```

```go
// internal/ui/states/files/state.go (updated)

type State struct {
    Source       core.DiffSource  // Changed from CommitRange
    Files        []core.FileChange
    Cursor       int
    ViewportStart int
    ExitOnPop    bool  // New field
}
```

## Open Questions

None. All design decisions have been made based on research and existing codebase patterns.

The implementation is ready to proceed to Phase 3.
