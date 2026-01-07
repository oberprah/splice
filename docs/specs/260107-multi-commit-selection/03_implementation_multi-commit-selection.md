# Implementation: Multi-Commit Selection

**Requirements:** `01_requirements_multi-commit-selection.md`
**Design:** `02_design_multi-commit-selection.md`

## Steps

### Step 1: Core Types (CursorState, CommitRange, LineDisplayState)

**Goal:** Introduce the foundational types that enable multi-commit selection without changing behavior yet.

**Structure:**
- New file `internal/core/cursor.go`: `CursorState` interface with `CursorNormal` and `CursorVisual` implementations
- New file `internal/core/commit_range.go`: `CommitRange` type with `Start` and `End` commits
- New file `internal/ui/components/line_display_state.go`: `LineDisplayState` enum

**Verify:** Unit tests for:
- `CursorState` types: Position() method, type assertions
- `CommitRange`: Single commit detection (IsSingleCommit()), commit ordering helpers
- `LineDisplayState`: Enum values and string representation

**Read:**
- `internal/core/state.go` (existing core types)
- `internal/git/git.go` (GitCommit type)

**Status:** Complete

**Implementation Notes:**
- Created `internal/core/cursor.go` with `CursorState` interface and sum type implementations
  - `CursorNormal` for single cursor position
  - `CursorVisual` for visual mode with anchor and cursor position
  - Helper functions: `SelectionRange()` and `IsInSelection()`
- Created `internal/core/commit_range.go` with `CommitRange` struct
  - Contains `Start`, `End` (GitCommit), and `Count` (int) fields
  - `IsSingleCommit()` method checks if Count == 1
  - Constructor functions: `NewSingleCommitRange()` and `NewCommitRange()`
- Created `internal/ui/components/line_display_state.go` with `LineDisplayState` enum
  - Four states: `LineStateNone`, `LineStateCursor`, `LineStateSelected`, `LineStateVisualCursor`
  - `String()` method for debugging
  - `SelectorString()` method returns visual indicators: "  ", "→ ", "▌ ", "█ "
- All tests pass (15 tests in cursor_test.go, 5 tests in commit_range_test.go, 3 tests in line_display_state_test.go)
- Build successful: `go build ./...` completes without errors
- Lint successful: `go tool golangci-lint run` reports 0 issues

---

### Step 2: Git Function Updates

**Goal:** Update git functions to accept range parameters (fromHash, toHash) instead of single commitHash.

**Structure:**
- `internal/git/git.go`:
  - `FetchFileChanges(commitHash string)` → `FetchFileChanges(fromHash, toHash string)`
  - `FetchFullFileDiff(commitHash string, change FileChange)` → `FetchFullFileDiff(fromHash, toHash string, change FileChange)`
  - Add helper `FetchFileChangesForCommit(commitHash string)` that calls `FetchFileChanges(commitHash+"^", commitHash)` for backward compatibility during transition

**Verify:** Unit tests for:
- `FetchFileChanges` with range: verifies git command uses correct syntax
- `FetchFileChangesForCommit` wrapper produces same output as before

**Read:**
- `internal/git/git.go` (current implementations at lines 290-360, 405-450)

**Status:** Complete

**Commit:** 07e89d7791c70f3c31baa9070c5807e9f1e8e561

**Implementation Notes:**
- Updated `FetchFileChanges(fromHash, toHash string)` to use `git diff` with range syntax instead of `git diff-tree`
- Updated `FetchFullFileDiff(fromHash, toHash string, change FileChange)` to fetch content at fromHash/toHash and call new `FetchFileDiffRange()` helper
- Added `FetchFileChangesForCommit(commitHash)` wrapper: calls `FetchFileChanges(commitHash+"^", commitHash)`
- Added `FetchFullFileDiffForCommit(commitHash, change)` wrapper: calls `FetchFullFileDiff(commitHash+"^", commitHash, change)`
- Added `FetchFileDiffRange(rangeSpec, filePath)` helper to fetch diffs for ranges using `git diff`
- Updated function type signatures in `internal/core/state.go`
- Updated all callers in `log/update.go`, `files/update.go` to use explicit `hash+"^", hash` syntax
- Updated test helpers in `internal/ui/testutils/helpers.go` and all test files
- All existing tests pass without modification (15 git tests, all UI state tests, all e2e tests)
- Build successful, linting clean

**Verification:**
- ✓ `go build ./...` succeeds
- ✓ `go test ./...` passes (all 15 git tests + all UI tests)
- ✓ `golangci-lint run` passes with 0 issues
- ✓ Integration tests verify git commands work with actual repository
- ✓ Backward compatibility maintained through wrapper functions

---

### Step 3: Navigation Message Updates

**Goal:** Update `PushFilesScreenMsg` and `PushDiffScreenMsg` to use `CommitRange` instead of single `Commit`.

**Structure:**
- `internal/core/navigation.go`:
  - `PushFilesScreenMsg.Commit` → `PushFilesScreenMsg.Range CommitRange`
  - `PushDiffScreenMsg.Commit` → `PushDiffScreenMsg.Range CommitRange`
- `internal/core/messages.go`:
  - `FilesLoadedMsg.Commit` → `FilesLoadedMsg.Range CommitRange`
  - `DiffLoadedMsg.Commit` → `DiffLoadedMsg.Range CommitRange`

**Verify:** Compile succeeds. All existing tests pass after updating test code to use `CommitRange`.

**Read:**
- `internal/core/navigation.go` (lines 20-32)
- `internal/core/messages.go`
- All files that reference these messages (log, files, diff states)

**Status:** Complete

**Implementation Notes:**
- Updated message structs in `internal/core/navigation.go`:
  - Changed `PushFilesScreenMsg.Commit` to `PushFilesScreenMsg.Range CommitRange`
  - Changed `PushDiffScreenMsg.Commit` to `PushDiffScreenMsg.Range CommitRange`
- Updated message structs in `internal/core/messages.go`:
  - Changed `FilesLoadedMsg.Commit` to `FilesLoadedMsg.Range CommitRange`
  - Changed `DiffLoadedMsg.Commit` to `DiffLoadedMsg.Range CommitRange`
- Updated state structs:
  - `internal/ui/states/files/state.go`: Changed `Commit` field to `Range CommitRange`, updated constructor `New()`
  - `internal/ui/states/diff/state.go`: Changed `Commit` field to `Range CommitRange`, updated constructor `New()`
- Updated all message producers to wrap single commits in `core.NewSingleCommitRange()`:
  - `internal/ui/states/log/update.go`: Wraps commit when creating `FilesLoadedMsg` and `PushFilesScreenMsg`
  - `internal/ui/states/files/update.go`: Uses `s.Range` instead of `s.Commit`, wraps in messages
- Updated view files to access commit via `Range.End`:
  - `internal/ui/states/files/view.go`: Uses `s.Range.End` for displaying commit info
  - `internal/ui/states/diff/view.go`: Uses `s.Range.End.Hash` in header rendering
- Updated navigation handler in `internal/app/model.go`:
  - Changed to pass `msg.Range` instead of `msg.Commit` to state constructors
- Updated all test files to wrap commits in `core.NewSingleCommitRange()`:
  - `internal/ui/states/files/update_test.go`: All State literals updated
  - `internal/ui/states/files/view_test.go`: All State literals updated
  - `internal/ui/states/diff/update_test.go`: All State literals updated
  - `internal/ui/states/diff/view_test.go`: All State literals updated
  - `internal/app/model_navigation_test.go`: Updated PushFilesScreenMsg and PushDiffScreenMsg usage

**Verification:**
- ✓ `go build ./...` succeeds
- ✓ `go test ./...` passes (all tests in app, core, domain, git, ui packages)
- ✓ `go tool golangci-lint run` passes with 0 issues
- ✓ All state tests pass with CommitRange
- ✓ Navigation tests verify correct range passing
- ✓ Backward compatibility maintained: single commits wrapped in NewSingleCommitRange()
- ✓ View rendering works correctly using Range.End for single commits

---

### Step 4: LogState Visual Mode

**Goal:** Implement visual mode selection in LogState with keyboard interaction and visual feedback.

**Structure:**
- `internal/ui/states/log/state.go`:
  - Replace `Cursor int` with `Cursor CursorState`
  - Add helper methods: `IsVisualMode()`, `GetSelectedRange()`, `CursorPosition()`
- `internal/ui/states/log/update.go`:
  - Handle `v` key: toggle visual mode
  - Handle `Escape`: exit visual mode
  - Update navigation (j/k): work with CursorState
  - Update Enter: create CommitRange from selection
- `internal/ui/states/log/view.go`:
  - Update `buildCommitLineComponents()` to compute `LineDisplayState` for each line
  - Update rendering to show correct cursor symbols (→, ▌, █)
- `internal/ui/components/log_line_format.go`:
  - Update `CommitLineComponents.IsSelected bool` → `CommitLineComponents.DisplayState LineDisplayState`
  - Update `FormatCommitLine` to render based on DisplayState

**Verify:**
- Unit tests for CursorState transitions (normal→visual, visual→normal)
- Unit tests for range calculation (anchor..cursor normalization)
- Golden file tests for visual mode rendering (all cursor states)

**Read:**
- `internal/ui/states/log/state.go`
- `internal/ui/states/log/update.go`
- `internal/ui/states/log/view.go`
- `internal/ui/components/log_line_format.go`

**Status:** Complete

**Implementation Notes:**
- Updated `internal/ui/states/log/state.go`:
  - Changed `Cursor int` to `Cursor core.CursorState`
  - Added helper methods: `CursorPosition()`, `IsVisualMode()`, `GetSelectedRange()`
  - `GetSelectedRange()` properly handles git log ordering (index 0 = newest, higher index = older)
- Updated `internal/ui/states/log/update.go`:
  - Added `v` key handler to toggle between `CursorNormal` and `CursorVisual`
  - Added `esc` key handler to exit visual mode
  - Updated all navigation keys (j/k/g/G) to work with CursorState (maintain Anchor in visual mode)
  - Updated `Enter` key to call `GetSelectedRange()` and pass CommitRange to FilesLoadedMsg
  - Updated `updateViewport()` to use `CursorPosition()` method
- Updated `internal/ui/states/log/view.go`:
  - Added `getLineDisplayState()` method to compute LineDisplayState for each line
  - Updated `buildCommitLineComponents()` to use `DisplayState` instead of `IsSelected`
  - Updated all cursor position access to use `CursorPosition()` method
- Updated `internal/ui/components/log_line_format.go`:
  - Changed `CommitLineComponents.IsSelected bool` to `CommitLineComponents.DisplayState LineDisplayState`
  - Updated `assembleLine()` to accept DisplayState and apply selected styles for both `LineStateSelected` and `LineStateVisualCursor`
  - Updated `FormatCommitLine()` to use `DisplayState.SelectorString()` for cursor indicators
- Updated all tests:
  - `internal/ui/states/log/update_test.go`: Updated all State literals to use `core.CursorNormal{Pos: n}`
  - `internal/ui/states/log/view_test.go`: Added core import, updated all State literals
  - `internal/ui/components/log_line_format_test.go`: Updated all CommitLineComponents to use DisplayState
  - Updated golden files for new cursor character (→ instead of >)
  - E2E golden files also updated

**Verification:**
- ✓ `go build ./...` succeeds
- ✓ `go test ./...` passes (all tests in all packages)
- ✓ `go tool golangci-lint run` passes with 0 issues
- ✓ All golden files updated and verified
- ✓ Visual mode toggle works (v key)
- ✓ Navigation extends selection in visual mode (j/k/g/G)
- ✓ Escape cancels visual mode
- ✓ Enter opens files view with range
- ✓ Cursor indicators render correctly: → (normal), ▌ (selected), █ (visual cursor)

---

### Step 5: CommitInfo Range Support

**Goal:** Update CommitInfo component to display range information when CommitRange spans multiple commits.

**Structure:**
- `internal/ui/components/commit_info.go`:
  - Add `CommitInfoFromRange(range CommitRange, files []git.FileChange, width int, ctx core.Context) []string`
  - Single commit: delegates to existing `CommitInfo()`
  - Range: renders `{startHash}..{endHash} (N commits)` header without subject/body

**Verify:**
- Unit tests for single commit rendering (unchanged behavior)
- Unit tests for range rendering (shows hash range, commit count)
- Golden file tests for both single and range display

**Read:**
- `internal/ui/components/commit_info.go`

**Status:** Complete

**Implementation Notes:**
- Added `CommitInfoFromRange(commitRange CommitRange, width int, bodyMaxLines int, ctx Context)` function to `internal/ui/components/commit_info.go`
  - For single commits (IsSingleCommit() == true), delegates to existing `CommitInfo()` function
  - For ranges (multiple commits), renders a header line with format: `{startHash}..{endHash} (N commits)`
  - Uses `format.ToShortHash()` to format hashes (7 characters)
  - Uses `styles.HashStyle` for styling the range header
- Updated `internal/ui/states/files/view.go`:
  - Changed from `CommitInfo(s.Range.End, ...)` to `CommitInfoFromRange(s.Range, ...)`
  - Now properly handles both single commits and ranges
- Added comprehensive tests in `internal/ui/components/commit_info_test.go`:
  - `TestCommitInfoFromRange_SingleCommit`: Verifies delegation to CommitInfo for single commits
  - `TestCommitInfoFromRange_MultipleCommits`: Verifies range header format and content
  - `TestCommitInfoFromRange_RangeFormat`: Verifies exact formatting of range header
  - All three tests verify correct handling of hash formatting, commit counts, and message structure
- Updated imports in commit_info_test.go to include `core` package for CommitRange types

**Verification:**
- ✓ `go build ./...` succeeds
- ✓ `go test ./...` passes (all tests in all packages)
- ✓ `go tool golangci-lint run` passes with 0 issues
- ✓ Single commit rendering unchanged (delegates to CommitInfo)
- ✓ Range rendering shows proper format: `abc123d..def456a (3 commits)`
- ✓ Hash formatting uses ToShortHash (7 characters)
- ✓ Range header uses HashStyle for consistent styling

---

### Step 6: FilesState and DiffState Updates

**Goal:** Update FilesState and DiffState to work with CommitRange.

**Structure:**
- `internal/ui/states/files/state.go`:
  - Replace `Commit git.GitCommit` with `Range CommitRange`
  - Update constructor `New(range CommitRange, files []git.FileChange)`
- `internal/ui/states/files/view.go`:
  - Use `CommitInfoFromRange()` instead of `CommitInfo()`
- `internal/ui/states/files/update.go`:
  - Pass Range to DiffLoadedMsg
  - Update loadDiff to use range hashes
- `internal/ui/states/diff/state.go`:
  - Replace `Commit git.GitCommit` with `Range CommitRange`
- `internal/ui/states/diff/view.go`:
  - Update header rendering for range display
- `internal/app/model.go`:
  - Update navigation handlers for new message shapes

**Verify:**
- Existing golden file tests pass (with updated test data)
- E2E tests pass for single commit flow
- New tests for range flow

**Read:**
- `internal/ui/states/files/state.go`, `view.go`, `update.go`
- `internal/ui/states/diff/state.go`, `view.go`, `update.go`
- `internal/app/model.go`

**Status:** Complete

**Implementation Notes:**
- Updated `internal/ui/states/files/update.go`:
  - Modified `loadDiff()` to properly handle both single commits and ranges
  - For single commits: uses `commitRange.End.Hash + "^"` to `commitRange.End.Hash`
  - For ranges: uses `commitRange.Start.Hash + "^"` to `commitRange.End.Hash`
  - This ensures the diff spans the correct range of commits
- Updated `internal/ui/states/diff/view.go`:
  - Modified `renderHeader()` to display range format for multi-commit ranges
  - Single commit: `abc123d · path/to/file.go · +15 -8`
  - Range: `abc123d..def456e · path/to/file.go · +15 -8`
  - Uses `format.ToShortHash()` for consistent 7-character hash display
  - Applies `styles.HashStyle` to entire range string for consistent styling
- Added test in `internal/ui/states/diff/view_test.go`:
  - `TestDiffState_View_RangeHeader`: Verifies range header format with 4-commit range
  - Creates golden file `range_header.golden` showing proper range display
  - Uses `core.NewCommitRange()` to create multi-commit range for testing
- Note: FilesState and DiffState already had Range field from Step 3
- Note: FilesState view already used CommitInfoFromRange from Step 5

**Verification:**
- ✓ `go build ./...` succeeds
- ✓ `go test ./...` passes (all tests in all packages)
- ✓ `go tool golangci-lint run` passes with 0 issues
- ✓ New golden file created and verified: `internal/ui/states/diff/testdata/range_header.golden`
- ✓ Range header shows format: `abc123d..def456a · internal/auth/handler.go · +25 -13`
- ✓ Single commit behavior unchanged
- ✓ FilesState correctly determines hash range for diff loading
- ✓ DiffState displays appropriate header based on range type

---

## Final Verification

- [ ] Full test suite passes (`go test ./...`)
- [ ] Lint passes (`go tool golangci-lint run`)
- [ ] Build succeeds (`go build -o splice .`)
- [ ] All requirements verified:
  - [ ] `v` enters visual mode, anchors current commit
  - [ ] Navigation keys extend selection
  - [ ] `Enter` opens FilesState with combined diff
  - [ ] `Escape` cancels selection
  - [ ] Cursor indicators: → (normal), ▌ (selected), █ (visual cursor)
  - [ ] Range header displays `abc123..def456 (N commits)`
  - [ ] Diff uses `git diff <older>..<newer>`
- [ ] Design decisions followed:
  - [ ] CursorState sum type (no illegal states)
  - [ ] CommitRange for all navigation
  - [ ] LineDisplayState enum (no invalid boolean combos)
  - [ ] Git functions unified to range syntax

## Summary

*To be completed after implementation*
