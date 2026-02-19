# State Management and Navigation Patterns in Splice

## Overview

This document analyzes how Splice manages UI state and navigation across different states (log, files, diff). It documents patterns for cursor navigation, viewport scrolling, update handlers, event flow, and testing to guide implementation of the tree file view.

## Core Architecture

### State Interface (`internal/core/state.go`)

All states implement the `core.State` interface:

```go
type State interface {
    View(ctx Context) ViewRenderer
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
```

**Key principles:**
- States are **immutable** - Update returns a new state, never modifies in place
- States access context via `core.Context` interface (width, height, time, fetch functions)
- States communicate via **typed messages** (not direct calls)
- Navigation uses messages: `PushXScreenMsg`, `PopScreenMsg`

### Context Interface (`internal/core/state.go:28-34`)

```go
type Context interface {
    Width() int
    Height() int
    FetchFileChanges() FetchFileChangesFunc
    FetchFullFileDiff() FetchFullFileDiffFunc
    Now() time.Time
}
```

Implemented by `app.Model` - provides terminal dimensions and injected dependencies.

### Navigation Stack Pattern (`internal/app/model.go`)

The app maintains a navigation stack where:
- Current state = `stack[len-1]`
- Push messages add states to stack
- Pop messages remove from stack
- Loading states are **transient** - they get replaced, not stacked

```go
func (m *Model) pushState(newState core.State) {
    current := m.current()
    _, isLoading := current.(loading.State)
    _, isDirectDiffLoading := current.(directdiff.State)

    if isLoading || isDirectDiffLoading {
        m.stack[len(m.stack)-1] = newState  // Replace
    } else {
        m.stack = append(m.stack, newState)  // Push
    }
}
```

## State Structure Patterns

### 1. Log State (`internal/ui/states/log/state.go`)

**Most complex state - supports cursor modes and async preview loading**

```go
type State struct {
    Commits       []core.GitCommit
    Cursor        core.CursorState        // Sum type: CursorNormal | CursorVisual
    ViewportStart int
    Preview       PreviewState            // Sum type for async state
    GraphLayout   *graph.Layout
}
```

**Pattern: Sum types for state variants**
- `CursorState`: Either `CursorNormal{Pos}` or `CursorVisual{Pos, Anchor}`
- `PreviewState`: `PreviewNone | PreviewLoading | PreviewLoaded | PreviewError`

**Pattern: Async data loading with stale detection**
```go
type PreviewLoading struct {
    ForHash string  // Track what we're loading
}

type PreviewLoaded struct {
    ForHash string  // Track what we loaded
    Files   []core.FileChange
}

// In Update handler - detect stale responses:
if currentRangeHash != msg.ForHash {
    return s, nil  // Discard stale response
}
```

### 2. Files State (`internal/ui/states/files/state.go`)

**Simple state - flat list with cursor**

```go
type State struct {
    Source        core.DiffSource
    Files         []core.FileChange
    Cursor        int                // Simple integer cursor
    ViewportStart int
}
```

**Simpler than log:**
- Single cursor mode (no visual selection)
- No async preview (loads full data upfront)
- Flat list navigation

### 3. Diff State (`internal/ui/states/diff/state.go`)

**Viewport-only state - no cursor, just scrolling**

```go
type State struct {
    Source           core.DiffSource
    File             core.FileChange
    Diff             *diff.AlignedFileDiff
    ViewportStart    int
    CurrentChangeIdx int      // For n/N navigation
    ChangeIndices    []int    // Jump targets for changes
}
```

**Pattern: Index-based navigation**
- No cursor (entire viewport scrolls)
- `ChangeIndices` stores positions of interesting items
- `n`/`N` keys jump between changes

## Cursor Patterns

### Core Cursor Types (`internal/core/cursor.go`)

```go
type CursorState interface {
    Position() int
    cursorState()  // Marker method
}

type CursorNormal struct {
    Pos int
}

type CursorVisual struct {
    Pos    int
    Anchor int
}
```

**Helper functions:**
```go
// Get ordered selection range
func SelectionRange(cursor CursorState) (int, int)

// Check if index is in selection
func IsInSelection(cursor CursorState, index int) bool
```

### Cursor Navigation Pattern (Log State)

**Up/Down navigation:**
```go
case "j", "down":
    pos := s.CursorPosition()
    if pos < len(s.Commits)-1 {
        newPos := pos + 1
        switch cursor := s.Cursor.(type) {
        case core.CursorNormal:
            s.Cursor = core.CursorNormal{Pos: newPos}
        case core.CursorVisual:
            s.Cursor = core.CursorVisual{Pos: newPos, Anchor: cursor.Anchor}
        }
        s.updateViewport(ctx.Height())
        // Trigger side effect (preview reload)
        return s, LoadPreview(...)
    }
```

**Pattern: Immutable state updates**
- Create new cursor struct with updated position
- Don't modify cursor in place
- Update viewport AFTER cursor moves
- Return command for async side effects

### Simpler Cursor Pattern (Files State)

```go
case "j", "down":
    if len(s.Files) > 0 && s.Cursor < len(s.Files)-1 {
        s.Cursor++
        s.updateViewport(ctx.Height())
    }
    return s, nil
```

**Difference:** No cursor modes, so simple integer increment.

## Viewport Scrolling Patterns

### Pattern 1: Keep Cursor Visible (Log & Files)

```go
func (s *State) updateViewport(height int) {
    pos := s.CursorPosition()

    // Scroll down if cursor is below viewport
    if pos >= s.ViewportStart+height {
        s.ViewportStart = pos - height + 1
    }

    // Scroll up if cursor is above viewport
    if pos < s.ViewportStart {
        s.ViewportStart = pos
    }

    // Ensure viewport doesn't go negative
    if s.ViewportStart < 0 {
        s.ViewportStart = 0
    }
}
```

**Files state adjustment:**
```go
func (s *State) updateViewport(height int) {
    headerLines := 2  // commit info + separator
    availableHeight := max(height-headerLines, 1)

    // Same logic but with availableHeight instead of height
    // ...
}
```

**Pattern: Account for header lines**
- Files state reserves space for commit info header
- Subtract header lines from available height
- Prevents viewport from scrolling header off screen

### Pattern 2: Free Scrolling (Diff State)

```go
case "j", "down":
    if s.Diff != nil && len(s.Diff.Alignments) > 0 {
        maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
        if s.ViewportStart < maxViewportStart {
            s.ViewportStart++
        }
    }

func (s *State) calculateMaxViewportStart(height int) int {
    headerLines := 2
    availableHeight := max(height-headerLines, 1)
    maxStart := len(s.Diff.Alignments) - availableHeight
    if maxStart < 0 {
        maxStart = 0
    }
    return maxStart
}
```

**Pattern: Bounded scrolling**
- Calculate max viewport position to prevent over-scrolling
- Ensure at least one screen of content visible

### Jump Navigation Patterns

**Jump to top/bottom (all states):**
```go
case "g":
    // Top
    s.Cursor = 0
    s.ViewportStart = 0

case "G":
    // Bottom
    s.Cursor = len(s.Items) - 1
    s.updateViewport(ctx.Height())  // Scroll to show cursor
```

**Jump to specific positions (diff state):**
```go
case "n":  // Next change
    s.jumpToNextChange(ctx.Height())

func (s *State) jumpToNextChange(height int) {
    for i, changeIdx := range s.ChangeIndices {
        if changeIdx > s.ViewportStart {
            s.CurrentChangeIdx = i
            s.ViewportStart = changeIdx
            // Clamp to max viewport
            maxViewport := s.calculateMaxViewportStart(height)
            if s.ViewportStart > maxViewport {
                s.ViewportStart = maxViewport
            }
            return
        }
    }
}
```

## Update Handler Patterns

### Message Routing Pattern

```go
func (s State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
    switch msg := msg.(type) {
    case CustomLoadedMsg:
        // Handle async data load
        return s.handleDataLoaded(msg)

    case tea.KeyMsg:
        return s.handleKey(msg, ctx)
    }

    return s, nil
}
```

**Pattern: Separate handlers for different message types**
- Async results (custom messages)
- Keyboard input (`tea.KeyMsg`)
- Window resize (`tea.WindowSizeMsg` - handled at app level)

### Keyboard Handler Pattern

```go
case tea.KeyMsg:
    switch msg.String() {
    case "q":
        return s.handleQuit()
    case "ctrl+c", "Q":
        return s, tea.Quit
    case "enter":
        return s.handleSelect(ctx)
    case "j", "down":
        return s.handleDown(ctx)
    // ... more keys
    }
```

**Pattern: String matching for keys**
- Use `msg.String()` for rune-based keys (`"j"`, `"g"`, `"G"`)
- Use `msg.Type` for special keys (`tea.KeyEnter`, `tea.KeyDown`)

### Navigation Message Pattern

**Push to new screen:**
```go
case "enter":
    // Start async load
    return s, s.loadData(...)

// In async handler:
case DataLoadedMsg:
    if msg.Err != nil {
        return s, nil  // Stay on current screen
    }

    // Navigate by returning command that produces navigation message
    return s, func() tea.Msg {
        return core.PushNewScreenMsg{
            Data: msg.Data,
        }
    }
```

**Pop to previous screen:**
```go
case "q":
    return s, func() tea.Msg {
        return core.PopScreenMsg{}
    }
```

**Pattern: Navigation via messages, not direct state changes**
- States don't create other states directly
- States return navigation messages
- App.Model handles the navigation stack

### Async Loading Pattern (Log State)

**Trigger load on navigation:**
```go
case "j":
    // Update cursor
    s.Cursor = newPosition

    // Mark as loading
    commitRange := s.GetSelectedRange()
    rangeHash := getRangeHash(commitRange)
    s.Preview = PreviewLoading{ForHash: rangeHash}

    // Return command to load
    return s, LoadPreview(commitRange, ctx.FetchFileChanges())
```

**Handle load completion:**
```go
case core.FilesPreviewLoadedMsg:
    // Stale detection
    currentRangeHash := getRangeHash(s.GetSelectedRange())
    if currentRangeHash != msg.ForHash {
        return s, nil  // Discard stale response
    }

    // Update preview state
    if msg.Err != nil {
        s.Preview = PreviewError{ForHash: msg.ForHash, Err: msg.Err}
    } else {
        s.Preview = PreviewLoaded{ForHash: msg.ForHash, Files: msg.Files}
    }
    return s, nil
```

**Pattern: Stale response detection**
- Tag async requests with identifier (hash)
- Compare identifier on response
- Discard if doesn't match current selection

## View Rendering Patterns

### Viewport Rendering Pattern

```go
func (s State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // Calculate viewport bounds
    viewportEnd := min(s.ViewportStart+ctx.Height(), len(s.Items))

    // Render only visible items
    for i := s.ViewportStart; i < viewportEnd; i++ {
        item := s.Items[i]
        line := renderItem(item, i == s.Cursor)
        vb.AddLine(line)
    }

    return vb
}
```

**Pattern: Only render visible items**
- Calculate what's visible based on viewport
- Iterate only visible range
- Efficient for large lists

### Line Display State Pattern (Log State)

```go
func (s State) getLineDisplayState(index int) components.LineDisplayState {
    pos := s.CursorPosition()

    switch cursor := s.Cursor.(type) {
    case core.CursorNormal:
        if index == pos {
            return components.LineStateCursor
        }
        return components.LineStateNone
    case core.CursorVisual:
        if index == pos {
            return components.LineStateVisualCursor
        }
        if core.IsInSelection(cursor, index) {
            return components.LineStateSelected
        }
        return components.LineStateNone
    }
}
```

**Pattern: Sum type for display states**
```go
type LineDisplayState int

const (
    LineStateNone LineDisplayState = iota
    LineStateCursor
    LineStateSelected
    LineStateVisualCursor
)
```

Maps cursor state to visual styling:
- `LineStateNone`: Normal styling
- `LineStateCursor`: Highlighted (current item)
- `LineStateSelected`: Selected range styling
- `LineStateVisualCursor`: Cursor within selection

### Header Lines Pattern (Files State)

```go
func (s *State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // 1. Render header (commit info)
    commitInfoLines := components.CommitInfoFromRange(...)
    for _, line := range commitInfoLines {
        vb.AddLine(line)
    }

    // 2. Render file section (blank + stats + files)
    fileSectionLines := components.FileSection(s.Files, ctx.Width(), &s.Cursor)

    // 3. Calculate available height for file list
    headerLinesCount := len(commitInfoLines)
    fileSectionHeaderLines := 2  // blank + stats
    availableHeight := max(ctx.Height()-headerLinesCount-fileSectionHeaderLines, 1)

    // 4. Add file section header
    for i := 0; i < fileSectionHeaderLines; i++ {
        vb.AddLine(fileSectionLines[i])
    }

    // 5. Add visible file lines only
    viewportEnd := min(s.ViewportStart+availableHeight, len(s.Files))
    for i := s.ViewportStart; i < viewportEnd; i++ {
        lineIndex := fileSectionHeaderLines + i
        vb.AddLine(fileSectionLines[lineIndex])
    }

    return vb
}
```

**Pattern: Fixed header + scrollable content**
- Header always visible (commit info + stats)
- Only file list scrolls
- Viewport calculations account for fixed header space

## Testing Patterns

### Unit Test Structure

```go
func TestStateName_Method_Scenario(t *testing.T) {
    // 1. Setup state
    s := State{
        Items:  createTestItems(10),
        Cursor: 0,
    }
    ctx := testutils.MockContext{W: 80, H: 24}

    // 2. Send message
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
    newState, cmd := s.Update(msg, ctx)

    // 3. Assert state changes
    stateAfter := newState.(State)
    if stateAfter.Cursor != 1 {
        t.Errorf("Expected cursor at 1, got %d", stateAfter.Cursor)
    }

    // 4. Assert commands (if any)
    if cmd != nil {
        result := cmd()
        // Assert message type/content
    }
}
```

### Golden File Test Pattern

```go
func TestState_View_Scenario(t *testing.T) {
    testutils.SetupColorProfile()  // For deterministic colors

    s := State{...}
    ctx := testutils.MockContext{W: 80, H: 24}

    output := s.View(ctx).String()
    testutils.AssertGolden(t, output, "testdata/scenario.golden", *update)
}
```

**Update golden files:**
```bash
go test ./... -update
```

### Mock Context Pattern

```go
type MockContext struct {
    W                    int
    H                    int
    MockFetchFileChanges core.FetchFileChangesFunc
}

func (m MockContext) Width() int { return m.W }
func (m MockContext) Height() int { return m.H }
func (m MockContext) Now() time.Time {
    // Fixed time for deterministic tests
    return time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
}
```

**Pattern: Inject mocks via context**
- All external dependencies accessed via context
- Tests provide mock implementations
- Ensures deterministic, fast tests

### Testing Async Patterns (Log State)

```go
func TestLogState_Update_NavigationTriggersPreviewLoading(t *testing.T) {
    s := State{
        Commits: createTestCommits(5),
        Cursor:  core.CursorNormal{Pos: 0},
        Preview: PreviewNone{},
    }

    mockFetch := func(commitRange core.CommitRange) ([]core.FileChange, error) {
        return []core.FileChange{{Path: "test.go"}}, nil
    }

    ctx := testutils.MockContext{
        W: 80, H: 24,
        MockFetchFileChanges: mockFetch,
    }

    // 1. Navigate - should trigger preview load
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
    newState, cmd := s.Update(msg, ctx)

    // 2. Assert preview state changed to loading
    stateAfter := newState.(State)
    previewLoading, ok := stateAfter.Preview.(PreviewLoading)
    if !ok {
        t.Fatal("Expected PreviewLoading")
    }

    // 3. Execute command to get result message
    resultMsg := cmd()

    // 4. Send result back to state
    finalState, _ := stateAfter.Update(resultMsg, ctx)

    // 5. Assert preview is now loaded
    finalAfter := finalState.(State)
    previewLoaded, ok := finalAfter.Preview.(PreviewLoaded)
    if !ok {
        t.Fatal("Expected PreviewLoaded")
    }
}
```

### Testing Stale Response Pattern

```go
func TestLogState_Update_FilesPreviewLoadedMsg_StaleResponse(t *testing.T) {
    commits := createTestCommits(3)
    s := State{
        Commits: commits,
        Cursor:  core.CursorNormal{Pos: 2},  // User moved to commit 2
        Preview: PreviewLoading{ForHash: commits[2].Hash},
    }

    // Stale response for commit 1 arrives
    msg := core.FilesPreviewLoadedMsg{
        ForHash: commits[1].Hash,
        Files:   []core.FileChange{{Path: "file.go"}},
        Err:     nil,
    }

    newState, _ := s.Update(msg, ctx)
    stateAfter := newState.(State)

    // Preview should remain as PreviewLoading for commit 2
    previewLoading, ok := stateAfter.Preview.(PreviewLoading)
    if !ok {
        t.Fatal("Expected Preview to remain PreviewLoading")
    }
    if previewLoading.ForHash != commits[2].Hash {
        t.Error("Stale response was not discarded")
    }
}
```

### Test Helper Patterns

```go
// Create deterministic test data
func createTestCommits(count int) []core.GitCommit {
    commits := make([]core.GitCommit, count)
    baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
    for i := range count {
        commits[i] = core.GitCommit{
            Hash:    string(rune('a' + i)),
            Message: "Commit " + string(rune('0'+i)),
            Date:    baseTime.Add(time.Duration(-i) * time.Hour),
        }
    }
    return commits
}

// Mock fetch functions
func MockFetchFileChanges(files []core.FileChange, err error) func(core.CommitRange) ([]core.FileChange, error) {
    return func(commitRange core.CommitRange) ([]core.FileChange, error) {
        return files, err
    }
}
```

## Key Patterns Summary

### State Management
1. **Immutable state updates** - Never modify state in place, return new state
2. **Sum types for variants** - Use interfaces with marker methods for state variants
3. **Context for dependencies** - Access terminal size and fetch functions via context
4. **Messages for communication** - States communicate via typed messages, not direct calls

### Cursor & Navigation
1. **CursorState sum type** - `CursorNormal | CursorVisual` for selection modes
2. **Position tracking** - Separate cursor position from viewport position
3. **Viewport follows cursor** - Update viewport after cursor moves to keep it visible
4. **Bounded scrolling** - Calculate max viewport to prevent over-scrolling

### Async Operations
1. **Loading state markers** - Mark state as loading before async operation
2. **Stale detection** - Tag requests and responses with identifiers
3. **Command pattern** - Return `tea.Cmd` for async operations
4. **Message-based completion** - Async results delivered via custom messages

### View Rendering
1. **Viewport-based rendering** - Only render visible items
2. **Header line accounting** - Subtract header space from available height
3. **Display state mapping** - Map cursor state to visual styling enum
4. **Component composition** - Use reusable components (FileSection, CommitInfo)

### Testing
1. **Mock context** - Inject dependencies via context for testing
2. **Golden files** - Snapshot test for visual output
3. **Deterministic data** - Use fixed times and test data generators
4. **Async testing** - Execute commands and feed results back to state
5. **Table-driven tests** - Use subtests for multiple scenarios

## Implications for Tree View

Based on these patterns, a tree file view should:

1. **State Structure:**
   - Store tree structure (not flat list)
   - Track cursor position in tree (path to current node)
   - Track expanded/collapsed state per folder
   - Maintain viewport position

2. **Cursor Navigation:**
   - Use tree-aware navigation (j/k moves through visible nodes)
   - Handle expansion/collapse with cursor adjustment
   - Keep cursor visible in viewport after tree changes

3. **Viewport Management:**
   - Flatten tree to visible nodes for rendering
   - Calculate viewport over flattened list
   - Update viewport when tree expands/collapses

4. **Update Handlers:**
   - Toggle expand/collapse on enter/space
   - Navigate between visible nodes with j/k
   - Jump to parent/first/last visible node

5. **View Rendering:**
   - Flatten tree to visible nodes
   - Render viewport slice of visible nodes
   - Show tree structure with indentation/symbols
   - Highlight cursor position

6. **Testing:**
   - Test expand/collapse logic with deterministic trees
   - Golden files for tree rendering
   - Test cursor navigation through tree structure
   - Test viewport updates on tree changes
