# Implementation: Smart Navigation in Diff View

**Requirements:** `01_requirements_smart-navigation-diff-view.md`
**Design:** `02_design_smart-navigation-diff-view.md`

## Steps

### Step 1: Create Block-Based Data Model

**Goal:** Define the new block-based data structures that replace the flat `[]Alignment` model.

**Structure:**
- New file `internal/domain/diff/block.go` containing:
  - `Block` interface (sealed with `block()` marker + `LineCount() int`)
  - `UnchangedBlock` with `Lines []LinePair`
  - `LinePair` struct with `LeftLineNo`, `RightLineNo`, `Tokens`
  - `ChangeBlock` with `Lines []ChangeLine`
  - `ChangeLine` interface (sealed with `changeLine()` marker)
  - `ModifiedLine` with both sides' tokens and inline diff
  - `RemovedLine` with left tokens only
  - `AddedLine` with right tokens only
  - `FileDiff` struct with `Path` and `Blocks []Block`

**Verify:** Unit tests for all new types: `block_test.go`

**Read:** `internal/domain/diff/alignment.go`, `02_design_smart-navigation-diff-view.md`

**Status:** Complete

**Commits:** (pending commit)

**Verification:** Tests passed
- `go build ./internal/domain/diff/` - compilation successful
- `go test ./internal/domain/diff/ -v -run "TestUnchangedBlock|TestChangeBlock|TestFileDiff|TestBlock|TestChangeLineInterface|TestModifiedLine"` - all 7 tests passed
- `go test ./...` - full test suite passed

**Notes:**
- Renamed existing `FileDiff` in `parse.go` to `ParsedFileDiff` to resolve naming collision. The old `FileDiff` was used as an intermediate parsing type for unified diff format; the new `FileDiff` is the canonical block-based structure as specified in the design.
- Updated `builder.go` and `builder_test.go` to use `ParsedFileDiff` for the parsing type.
- All existing tests continue to pass after the rename.

---

### Step 2: Update BuildFileDiff to Produce Block Structure

**Goal:** Modify the builder to produce the new `FileDiff` with blocks instead of `AlignedFileDiff` with flat alignments.

**Structure:**
- Modify `internal/domain/diff/builder.go`:
  - Create new function `BuildFileDiff` that returns `*FileDiff`
  - Group consecutive unchanged lines into `UnchangedBlock`
  - Group consecutive change lines into `ChangeBlock`
  - Keep `BuildFileContent` for syntax highlighting (still needed internally)
- The existing `BuildAlignedFileDiff` can be kept temporarily until all callers are migrated

**Verify:** Unit tests covering:
- Single unchanged block
- Single change block (mixed modified/added/removed)
- Alternating unchanged and change blocks
- Multiple change blocks separated by unchanged
- Edge cases: empty diff, all changes, all unchanged

**Read:** `internal/domain/diff/builder.go`, `internal/domain/diff/alignment.go`

**Status:** Complete

**Commits:** (pending commit)

**Verification:** Tests passed
- `go build ./internal/domain/diff/` - compilation successful
- `go test ./internal/domain/diff/ -v -run TestBuildFileDiff` - all 12 tests passed
- `go test ./...` - full test suite passed
- `go tool golangci-lint run` - 0 issues

**Notes:**
- Added `BuildFileDiff` function as the new entry point for building block-based diffs
- Added `buildBlocks` helper function that walks through both files and groups lines into blocks
- The function reuses existing `pairLines` for line similarity matching and `diffmatchpatch` for inline diffs
- Key difference from `BuildAlignments`: instead of appending to a flat slice, we accumulate lines into current block and flush when block type changes
- Test cases cover: single unchanged block, single change block (only added/removed), mixed blocks, consecutive changes, consecutive changes with pairing, empty diff, multiple change blocks, total line count, token preservation, inline diff verification, and mixed change types

---

### Step 3: Update Core Messages and Navigation Types

**Goal:** Update `PushDiffScreenMsg` and `DiffLoadedMsg` to use the new `FileDiff` type and include file list for navigation.

**Structure:**
- Modify `internal/core/messages.go`:
  - `DiffLoadedMsg`: Change `Diff` field from `*diff.AlignedFileDiff` to `*diff.FileDiff`, remove `ChangeIndices`, add `Files []FileChange`, `FileIndex int`
- Modify `internal/core/navigation.go`:
  - `PushDiffScreenMsg`: Same changes as `DiffLoadedMsg`

**Verify:** Compilation succeeds (tests will fail until other steps complete)

**Read:** `internal/core/messages.go`, `internal/core/navigation.go`

**Status:** Complete

**Verification:**
- `go build ./internal/core/...` - compilation successful
- Full build fails with expected caller errors in `internal/ui/states/files/update.go` - callers will be updated in subsequent steps

**Notes:**
- Removed `ChangeIndices []int` field from both `DiffLoadedMsg` and `PushDiffScreenMsg` - no longer needed with block-based model where `ChangeBlock` types are self-identifying
- Added `Files []FileChange` and `FileIndex int` fields to enable file navigation from diff view
- Changed `Diff` field type from `*diff.AlignedFileDiff` to `*diff.FileDiff` to use the new block-based structure
- Callers in `internal/ui/states/files/update.go` and `internal/ui/states/diff/` will be updated in Steps 4-8

---

### Step 4: Update Diff State with Files Context and New Fields

**Goal:** Expand the diff state to hold file navigation context and use the new `FileDiff` type.

**Structure:**
- Modify `internal/ui/states/diff/state.go`:
  - Add `Files []core.FileChange` field (all files in diff source)
  - Add `FileIndex int` field (position in file list)
  - Change `Diff` field type from `*diff.AlignedFileDiff` to `*diff.FileDiff`
  - Remove `ChangeIndices []int` field
  - Remove `CurrentChangeIdx int` field (replaced by block-based navigation)
  - Add `CurrentBlockIdx int` field (tracks current change block for navigation)
- Update `New()` constructor to accept new parameters and initialize properly

**Verify:** Unit tests for state creation with various configurations

**Read:** `internal/ui/states/diff/state.go`

**Status:** Complete

**Verification:**
- `go build ./internal/ui/states/diff/state.go` - compilation successful
- `go vet ./internal/ui/states/diff/state.go` - no issues
- Full package build fails with expected caller errors in `update.go` and `view.go` - these will be updated in Steps 5-6

**Notes:**
- Added `Files []core.FileChange` field to store all files in the diff source for file-to-file navigation
- Added `FileIndex int` field to track the current file's position in the Files slice
- Changed `Diff` field type from `*diff.AlignedFileDiff` to `*diff.FileDiff` to use the new block-based structure
- Removed `ChangeIndices []int` field - no longer needed with block-based model where ChangeBlock types are self-identifying
- Removed `CurrentChangeIdx int` field - replaced by `CurrentBlockIdx` for block-based navigation
- Added `CurrentBlockIdx int` field to track the index of the current change block (-1 means "not in a change block")
- Updated `New()` constructor signature to accept `files []core.FileChange` and `fileIndex int` parameters
- Constructor now iterates through blocks to find and position the viewport at the first `ChangeBlock`
- Fixed comment: changed "DiffState represents" to "State represents" to match the actual type name

---

### Step 5: Update Diff View to Render Blocks

**Goal:** Modify the view rendering to iterate over blocks instead of flat alignments.

**Structure:**
- Modify `internal/ui/states/diff/view.go`:
  - Iterate over `s.Diff.Blocks` instead of `s.Diff.Alignments`
  - For `UnchangedBlock`: render each `LinePair`
  - For `ChangeBlock`: render each `ChangeLine` (type switch on ModifiedLine/RemovedLine/AddedLine)
  - Update helper functions to work with new types
  - Update `calculateLineNoWidth()` to iterate blocks
  - Update `calculateMaxViewportStart()` to count total lines across blocks

**Verify:** Golden file tests for view rendering (update existing tests to use new structure)

**Read:** `internal/ui/states/diff/view.go`, `internal/ui/states/diff/view_test.go`

**Status:** Complete

**Verification:**
- `go build ./...` - compilation successful
- `go test ./...` - all tests passed
- `go tool golangci-lint run` - 0 issues

**Notes:**
- Updated `View()` method to iterate over `s.Diff.Blocks` instead of `s.Diff.Alignments`, using nested loops: outer loop over blocks, inner loop over lines within each block
- Added `renderLinePair()` helper method to render `diff.LinePair` (unchanged lines) using shared tokens
- Added `renderChangeLine()` helper method to render `diff.ChangeLine` (type switch on `ModifiedLine`, `RemovedLine`, `AddedLine`) with appropriate left/right tokens and inline diff
- Removed old `renderAlignment()` function that referenced the deprecated `AlignedFileDiff` structure
- Updated `calculateLineNoWidth()` to iterate over blocks and extract line numbers from `LinePair` and `ChangeLine` types
- Updated `calculateMaxViewportStart()` in `update.go` to use `s.Diff.TotalLineCount()` instead of `len(s.Diff.Alignments)`
- Updated `jumpToNextChange()` and `jumpToPreviousChange()` in `update.go` to work with block structure (temporary implementation, will be replaced in Step 6)
- Updated `getCurrentFileLineNumber()` in `update.go` to traverse blocks and find the line at the current viewport position
- Added `findNextRightLineNo()` helper for finding the next line with a right line number (used for removed lines)
- Updated all view tests in `view_test.go` to use new `diff.FileDiff` with blocks instead of `AlignedFileDiff` with alignments
- Updated all update tests in `update_test.go` to use new block-based structure
- Updated `internal/ui/states/files/update.go` to use `BuildFileDiff` instead of `BuildAlignedFileDiff`, and include `Files` and `FileIndex` in `DiffLoadedMsg`
- Updated `internal/app/model.go` to pass new parameters to `diff.New()`
- Updated `internal/core/navigation_test.go` to use new `PushDiffScreenMsg` structure with `Files`, `FileIndex`, and `*diff.FileDiff`
- Updated `internal/ui/states/files/update_test.go` to use new `DiffLoadedMsg` structure
- Golden files remain unchanged as the visual rendering is equivalent

---

### Step 6: Implement Smart Change Navigation (n/p Keys)

**Goal:** Replace the current `n`/`N` navigation with smart `n`/`p` navigation that scrolls through multi-screen changes.

**Structure:**
- Modify `internal/ui/states/diff/update.go`:
  - Replace `"N"` key handler with `"p"` for previous change
  - Update `"n"` handler to use new smart logic
  - Implement `navigateToNextChange(height int)`:
    - Calculate current position in block structure
    - If in ChangeBlock and end not visible: scroll half page
    - Otherwise: find and jump to next ChangeBlock (or next file)
  - Implement `navigateToPrevChange(height int)`:
    - If in ChangeBlock and start not visible: scroll half page
    - Otherwise: find and jump to previous ChangeBlock (or previous file)
  - Add helper functions:
    - `findBlockAtPosition(lineOffset int) (blockIdx int, lineInBlock int)`
    - `getBlockStartPosition(blockIdx int) int`
    - `getBlockEndPosition(blockIdx int) int`
    - `findNextChangeBlock(fromBlock int) int` (-1 if none)
    - `findPrevChangeBlock(fromBlock int) int` (-1 if none)
- Remove old `jumpToNextChange` and `jumpToPreviousChange` functions

**Verify:** Unit tests covering:
- Navigation within unchanged content → jumps to next change
- Navigation within small change → jumps to next change
- Navigation within multi-screen change → scrolls first, then jumps
- Navigation at last change → triggers file navigation (tested in Step 7)
- Navigation at first change → triggers file navigation

**Read:** `internal/ui/states/diff/update.go`, `02_design_smart-navigation-diff-view.md` (navigation flowchart)

**Status:** Complete

**Verification:**
- `go build ./internal/ui/states/diff/` - compilation successful
- `go test ./internal/ui/states/diff/ -v -run "TestNavigate"` - all 8 new tests passed
- `go test ./...` - full test suite passed
- `go tool golangci-lint run` - 0 issues

**Notes:**
- Replaced key binding from `N` (previous change) to `p` (previous change) to match requirements
- Removed old `jumpToNextChange` and `jumpToPreviousChange` functions
- Implemented `navigateToNextChange` that:
  - Checks if currently in a ChangeBlock with content extending below viewport
  - If so, scrolls down half page to show more of current change
  - Otherwise, finds and jumps to the next ChangeBlock
  - Returns `(*State, tea.Cmd)` for future file navigation integration (Step 7)
- Implemented `navigateToPrevChange` that:
  - Checks if currently in a ChangeBlock with content extending above viewport
  - If so, scrolls up half page or jumps to block start (if start is visible)
  - Otherwise, finds and jumps to the previous ChangeBlock
  - Returns `(*State, tea.Cmd)` for future file navigation integration (Step 7)
- Added 6 helper functions for block position calculations:
  - `getBlockAtPosition(linePos int)` - returns block index and line offset within block
  - `getBlockStartPosition(blockIdx int)` - returns global line position where block starts
  - `getBlockEndPosition(blockIdx int)` - returns global line position where block ends
  - `findNextChangeBlock(fromBlock int)` - finds next ChangeBlock index (-1 if none)
  - `findPrevChangeBlock(fromBlock int)` - finds previous ChangeBlock index (-1 if none)
  - `isChangeBlockEndVisible(blockIdx, availableHeight int)` - checks if block end is visible
  - `isChangeBlockStartVisible(blockIdx int)` - checks if block start is visible
- Added new helper function `createTestDiffStateWithMultiScreenChange` for testing multi-screen changes
- Cross-file navigation (when at last/first change in file) will be connected in Step 7

---

### Step 7: Implement File Navigation (]/[ Keys)

**Goal:** Add `]` and `[` keys to navigate directly between files, and handle cross-file navigation from `n`/`p`.

**Structure:**
- Modify `internal/ui/states/diff/update.go`:
  - Add `"]"` key handler: call `navigateToNextFile(ctx)`
  - Add `"["` key handler: call `navigateToPrevFile(ctx)`
  - Implement `navigateToNextFile(ctx core.Context) tea.Cmd`:
    - If `FileIndex+1 < len(Files)`: return `loadDiffForFile(FileIndex+1, ctx)`
    - Otherwise: return nil (stay in place)
  - Implement `navigateToPrevFile(ctx core.Context) tea.Cmd`:
    - If `FileIndex > 0`: return `loadDiffForFile(FileIndex-1, ctx)`
    - Otherwise: return nil (stay in place)
  - Implement `loadDiffForFile(fileIndex int, ctx core.Context) tea.Cmd`:
    - Creates async command to fetch diff for `Files[fileIndex]`
    - Returns `DiffLoadedMsg` on completion
  - Update smart navigation to call file navigation when at boundaries:
    - `navigateToNextChange`: at last change block → call `navigateToNextFile`
    - `navigateToPrevChange`: at first change block → call `navigateToPrevFile`
  - Add `DiffLoadedMsg` handler in `Update()` to update state with new file's diff

**Verify:** Unit tests covering:
- `]` at first file → navigates to second file
- `]` at last file → stays in place
- `[` at last file → navigates to previous file
- `[` at first file → stays in place
- `n` at last change → triggers file navigation
- `p` at first change → triggers file navigation

**Read:** `internal/ui/states/diff/update.go`, `internal/ui/states/files/update.go` (for loadDiff pattern)

**Status:** Complete

**Verification:**
- `go build ./internal/ui/states/diff/` - compilation successful
- `go test ./internal/ui/states/diff/ -v -run "TestNavigate|TestDiffLoaded|TestPosition"` - all 21 tests passed
- `go test ./...` - full test suite passed
- `go tool golangci-lint run` - 0 issues

**Notes:**
- Added `DiffLoadedMsg` handler in `Update()` to process async diff loading results and update state
- Added key handlers for `]` (next file) and `[` (previous file) in `Update()`
- Implemented `navigateToNextFile(ctx)` - creates async command to load next file's diff, or stays in place at last file
- Implemented `navigateToPrevFile(ctx)` - creates async command to load previous file's diff, or stays in place at first file
- Implemented `loadDiffForFile(file, fileIndex, ctx)` - creates tea.Cmd that fetches and parses diff, returns `DiffLoadedMsg`
- Copied `fetchFileDiffForSource` helper from files/update.go to diff/update.go to handle different DiffSource types (CommitRange, Uncommitted changes)
- Implemented `positionAtFirstChange()` - resets viewport to first change block after loading new file
- Updated `navigateToNextChange(ctx)` to accept `core.Context` and trigger file navigation when at last change block in a file
- Updated `navigateToPrevChange(ctx)` to accept `core.Context` and trigger file navigation when at first change block in a file
- Added test helper `createTestDiffStateWithFiles` for creating states with multiple files
- Added test helper `createTestDiffStateWithChangesAndFiles` for creating states with changes and multiple files
- Added `MockFetchFullFileDiff` field to `testutils.MockContext` for injectable diff loading in tests
- All cross-file navigation uses async loading pattern consistent with existing codebase architecture

---

### Step 8: Update Files State to Pass File List to Diff

**Goal:** Modify the files state to include the full file list and file index when navigating to diff view.

**Structure:**
- Modify `internal/ui/states/files/update.go`:
  - Update `DiffLoadedMsg` handler to include `Files` and `FileIndex` in `PushDiffScreenMsg`
  - Calculate `FileIndex` by finding the loaded file's position in `s.Files`
  - Update `loadDiff` to work with new `BuildFileDiff` function
- Modify `internal/app/model.go`:
  - Update `PushDiffScreenMsg` handler to pass new fields to `diff.New()`

**Verify:**
- Unit tests for files state diff loading
- Integration test: navigate from files view to diff view, verify file context is passed

**Read:** `internal/ui/states/files/update.go`, `internal/app/model.go`

**Status:** Complete (completed as part of Step 5)

**Verification:**
- Files state already updated in Step 5 to pass `Files` and `FileIndex` to diff screen
- `internal/ui/states/files/update.go` uses `BuildFileDiff` and includes file context in `DiffLoadedMsg`
- `internal/app/model.go` updated to pass new fields to `diff.New()`

---

### Step 9: Remove Old Alignment Types and Clean Up Code

**Goal:** Remove the old `Alignment` types and `AlignedFileDiff` that are no longer used.

**Structure:**
- Delete from `internal/domain/diff/alignment.go`:
  - `AlignedLine` struct (replaced by tokens in LinePair/ChangeLine)
  - `FileContent` struct (no longer needed)
  - `Alignment` interface and all implementations (`UnchangedAlignment`, `ModifiedAlignment`, `RemovedAlignment`, `AddedAlignment`)
  - `AlignedFileDiff` struct (replaced by `FileDiff`)
- Delete from `internal/domain/diff/builder.go`:
  - `BuildAlignedFileDiff` function (replaced by `BuildFileDiff`)
  - Any helper functions only used by the old builder
- Update any remaining imports/references

**Verify:**
- All tests pass
- No compiler errors
- `go tool golangci-lint run` passes

**Read:** `internal/domain/diff/alignment.go`, `internal/domain/diff/builder.go`

**Status:** Complete

**Verification:**
- `go build ./...` - compilation successful
- `go test ./...` - all tests passed
- `go tool golangci-lint run` - 0 issues

**Notes:**
- Removed `Alignment` interface and all 4 implementations (`UnchangedAlignment`, `ModifiedAlignment`, `RemovedAlignment`, `AddedAlignment`) from alignment.go
- Removed `AlignedFileDiff` struct from alignment.go
- Removed `BuildAlignedFileDiff` function from builder.go
- Removed `BuildAlignments` function from builder.go (only used internally by `BuildAlignedFileDiff`)
- Removed all `TestBuildAlignments_*` tests and related integration tests from builder_test.go
- Kept `AlignedLine` and `FileContent` as they are still used internally by `BuildFileDiff` and `buildBlocks`
- Renamed `alignment.go` to `content.go` since it now only contains file content types
- Removed unused `diffmatchpatch` import from content.go (was only used by removed `ModifiedAlignment`)

---

### Step 10: Final Verification and E2E Tests

**Goal:** Verify the complete implementation meets all requirements.

**Structure:**
- Create E2E test in `test/e2e/smart_navigation_test.go`:
  - Test `n` navigation through multi-screen change (scrolls then jumps)
  - Test `p` navigation backwards
  - Test `]` and `[` for direct file navigation
  - Test cross-file navigation with `n`/`p`
  - Test boundary conditions (first/last file, first/last change)
- Run full test suite: `go test ./...`
- Update any affected golden files: `go test ./... -update`
- Verify with tape-runner for visual confirmation

**Verify:**
- [ ] Full test suite passes
- [ ] All requirements from `01_requirements_smart-navigation-diff-view.md` verified
- [ ] Design decisions from `02_design_smart-navigation-diff-view.md` followed
- [ ] Lint passes: `go tool golangci-lint run`

**Read:** `01_requirements_smart-navigation-diff-view.md`, `02_design_smart-navigation-diff-view.md`

**Status:** Pending

---

## Final Verification

- [ ] Full test suite passes
- [ ] All requirements verified:
  - [ ] `n` scrolls through multi-screen changes before jumping
  - [ ] `p` scrolls through multi-screen changes before jumping
  - [ ] `n`/`p` navigate across files at boundaries
  - [ ] `]` jumps to next file
  - [ ] `[` jumps to previous file
  - [ ] No wrapping at first/last file
- [ ] Design decisions followed:
  - [ ] Block-based data model implemented
  - [ ] Old Alignment types removed
  - [ ] File list passed to diff state
  - [ ] Async loading pattern preserved

## Summary

(To be filled after implementation)
