# Data Model Options Research

Analysis of different approaches for representing diff sources (commit ranges vs uncommitted changes).

## Problem Statement

The current architecture uses `CommitRange` throughout:
```go
type CommitRange struct {
    Start GitCommit
    End   GitCommit
    Count int
}
```

This works for commit-to-commit diffs but cannot represent uncommitted changes (no actual commits exist).

**Affected Areas:**
- `FilesState` and `DiffState` store `CommitRange`
- Navigation messages (`PushFilesScreenMsg`, `PushDiffScreenMsg`) carry `CommitRange`
- Git functions (`FetchFileChanges`, `FetchFullFileDiff`) accept `CommitRange`
- View components (`CommitInfoFromRange`) display commit info

## Option 1: Sealed Interface Pattern (Sum Type)

### Structure

```go
// DiffSource is a sum type representing the source of a diff
type DiffSource interface {
    diffSource() // Sealed interface marker
    DisplayName() string
}

// CommitRangeDiffSource represents a diff between two commits
type CommitRangeDiffSource struct {
    Start GitCommit
    End   GitCommit
    Count int
}

// UncommittedChangesDiffSource represents uncommitted changes
type UncommittedChangesDiffSource struct {
    Type UncommittedType // Unstaged, Staged, or All
}

type UncommittedType int
const (
    UncommittedTypeUnstaged UncommittedType = iota
    UncommittedTypeStaged
    UncommittedTypeAll
)

// Marker methods (seal the interface)
func (CommitRangeDiffSource) diffSource()      {}
func (UncommittedChangesDiffSource) diffSource() {}
```

### Usage Pattern

```go
// Type switch in functions
func FetchFileChanges(source DiffSource) ([]FileChange, error) {
    switch s := source.(type) {
    case CommitRangeDiffSource:
        return fetchForCommitRange(s)
    case UncommittedChangesDiffSource:
        return fetchForUncommitted(s.Type)
    default:
        panic("unreachable: unknown DiffSource type")
    }
}

// View rendering
func DisplayHeader(source DiffSource) string {
    switch s := source.(type) {
    case CommitRangeDiffSource:
        return fmt.Sprintf("%s..%s", s.Start.Hash[:7], s.End.Hash[:7])
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
}
```

### Impact

**Files to update:**
- `internal/core/diff_source.go` (new file)
- `internal/core/navigation.go` (change message types)
- `internal/git/git.go` (update function signatures)
- `internal/ui/states/files/state.go` (replace CommitRange field)
- `internal/ui/states/diff/state.go` (replace CommitRange field)
- `internal/ui/components/*.go` (update view functions)

**Pros:**
- Type-safe: compiler prevents invalid combinations
- Follows existing codebase pattern (like `Alignment`)
- Clear separation of concerns
- Easy to extend (add more diff source types)
- Type switches force handling all cases

**Cons:**
- More verbose than current CommitRange
- Requires changes throughout codebase
- Type assertions needed in some places
- Higher initial implementation effort

## Option 2: Extended CommitRange with Optional Fields

### Structure

```go
type CommitRange struct {
    Start       *GitCommit // nil for uncommitted changes
    End         *GitCommit // nil for uncommitted changes
    Count       int        // 0 for uncommitted
    IsUncommitted bool
    UncommittedType string // "unstaged", "staged", "all"
}

// Helper methods
func (r CommitRange) IsUncommittedChanges() bool {
    return r.IsUncommitted
}

func NewUncommittedRange(changeType string) CommitRange {
    return CommitRange{
        IsUncommitted: true,
        UncommittedType: changeType,
    }
}

func NewCommitRange(start, end GitCommit, count int) CommitRange {
    return CommitRange{
        Start: &start,
        End: &end,
        Count: count,
        IsUncommitted: false,
    }
}
```

### Usage Pattern

```go
func FetchFileChanges(commitRange CommitRange) ([]FileChange, error) {
    if commitRange.IsUncommitted {
        return fetchUncommittedChanges(commitRange.UncommittedType)
    }
    // Existing logic for commit ranges
    return fetchForCommitRange(*commitRange.Start, *commitRange.End)
}

func DisplayHeader(commitRange CommitRange) string {
    if commitRange.IsUncommitted {
        return fmt.Sprintf("Uncommitted changes (%s)", commitRange.UncommittedType)
    }
    return fmt.Sprintf("%s..%s", commitRange.Start.Hash[:7], commitRange.End.Hash[:7])
}
```

### Impact

**Files to update:**
- `internal/core/git_types.go` (extend CommitRange)
- `internal/git/git.go` (add conditionals)
- View components (add conditionals for display)

**Pros:**
- Minimal signature changes
- Backward compatible with existing code
- Simpler diff (fewer files touched)
- Faster to implement
- Existing tests mostly work unchanged

**Cons:**
- Multiple invalid state combinations possible
  - `Start=nil, End=valid, IsUncommitted=false` (broken)
  - `Start=valid, End=nil, IsUncommitted=true` (inconsistent)
  - etc. (6+ invalid states)
- Violates "make illegal states unrepresentable" principle
- Requires runtime validation
- Less type-safe
- Documentation burden

## Option 3: Wrapper Type with Tagged Union

### Structure

```go
type DiffSpec struct {
    Type        DiffSpecType
    CommitRange *CommitRange     // Non-nil for commit ranges
    Uncommitted *UncommittedSpec // Non-nil for uncommitted
}

type DiffSpecType int
const (
    DiffSpecTypeCommitRange DiffSpecType = iota
    DiffSpecTypeUncommitted
)

type UncommittedSpec struct {
    Type string // "unstaged", "staged", "all"
}

// Constructors ensure only one field is set
func NewDiffSpecCommitRange(start, end GitCommit, count int) DiffSpec {
    return DiffSpec{
        Type: DiffSpecTypeCommitRange,
        CommitRange: &CommitRange{Start: start, End: end, Count: count},
    }
}

func NewDiffSpecUncommitted(typ string) DiffSpec {
    return DiffSpec{
        Type: DiffSpecTypeUncommitted,
        Uncommitted: &UncommittedSpec{Type: typ},
    }
}
```

### Usage Pattern

```go
func FetchFileChanges(spec DiffSpec) ([]FileChange, error) {
    switch spec.Type {
    case DiffSpecTypeCommitRange:
        return fetchForCommitRange(*spec.CommitRange)
    case DiffSpecTypeUncommitted:
        return fetchForUncommitted(spec.Uncommitted.Type)
    }
}
```

**Pros:**
- Type-safe constructors
- Clear separation with explicit fields
- No invalid combinations (constructors prevent it)
- Middle ground between Options 1 and 2

**Cons:**
- Awkward field access (check Type, then access pointer)
- Still requires type switching
- Nullable pointers less clean than sealed interface
- More boilerplate than Option 1

## Comparison Table

| Aspect | Option 1: Sealed Interface | Option 2: Extended CommitRange | Option 3: Wrapper Type |
|--------|---------------------------|------------------------------|----------------------|
| **Type Safety** | Excellent | Poor | Good |
| **Invalid States** | None possible | 6+ invalid combinations | None possible |
| **Codebase Changes** | High (many files) | Low (focused areas) | Medium |
| **Implementation Effort** | High | Low | Medium |
| **Go Idioms** | Excellent | Fair | Good |
| **Extensibility** | Excellent | Fair | Medium |
| **Pattern Match** | Alignment pattern | Ad-hoc | Tagged union |
| **Testing Complexity** | Medium | Low | Medium |
| **Documentation Needs** | Low | High | Medium |
| **Backward Compat** | Breaking | Preserved | Breaking |

## Recommendation

**Option 1: Sealed Interface Pattern** is recommended.

### Rationale

1. **Aligns with codebase**: The `Alignment` type already uses this exact pattern (sealed interface with marker method)

2. **Type safety**: Compiler prevents invalid combinations. Option 2 allows broken states like `Start=nil, End=valid, IsUncommitted=false`

3. **Extensibility**: Future requirements (stash diffs, cherry-pick, etc.) extend naturally

4. **Go idioms**: Sealed interfaces via marker methods is idiomatic for sum types in Go

5. **Self-documenting**: Code clearly shows which diff sources are valid

### Implementation Strategy

1. Create `internal/core/diff_source.go` with sealed interface
2. Update core types first (navigation.go, git_types.go)
3. Adapt git commands to use type switches
4. Update states (FilesState, DiffState)
5. Update view components
6. Update app model

### Precedent in Codebase

From `internal/domain/diff/alignment.go`:
```go
type Alignment interface {
    alignment()
}

type UnchangedAlignment struct { ... }
type ModifiedAlignment struct { ... }
type RemovedAlignment struct { ... }
type AddedAlignment struct { ... }

func (UnchangedAlignment) alignment() {}
func (ModifiedAlignment) alignment() {}
// etc.
```

This proves the team understands and values this pattern for sum types.
