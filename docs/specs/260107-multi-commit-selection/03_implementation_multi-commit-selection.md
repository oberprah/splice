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

**Status:** Pending

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

**Status:** Pending

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

**Status:** Pending

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

**Status:** Pending

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
