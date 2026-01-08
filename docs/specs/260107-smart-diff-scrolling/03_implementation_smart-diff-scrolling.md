# Implementation: Smart Diff Scrolling

**Requirements:** `01_requirements_smart-diff-scrolling.md`
**Design:** `02_design_smart-diff-scrolling.md`

## Steps

### Step 1: Create Segment Data Model

**Goal:** Define the new segment-based data types in `internal/domain/diff/` that will replace the flat `Alignments` array. This is the foundation for all subsequent work.

**Structure:**
- New file: `internal/domain/diff/segment.go`
- Types:
  - `Segment` interface (sealed sum type marker)
  - `UnchangedSegment` struct with `LeftStart`, `RightStart`, `Count` fields
  - `HunkSegment` struct with `LeftLines`, `RightLines` slices
  - `HunkLine` struct with `SourceIdx`, `Type` fields
  - `LineType` enum: `LineTypeAdded`, `LineTypeRemoved`, `LineTypeModified`
- Extend `AlignedFileDiff` to include `Segments []Segment` field (coexist with `Alignments` initially for migration)

**Verify:**
- Unit tests for segment types (construction, type assertions)
- Tests pass: `go test ./internal/domain/diff/...`

**Read:**
- `internal/domain/diff/alignment.go` (existing types)
- `02_design_smart-diff-scrolling.md` (data types section)

**Status:** Complete
**Commits:** 61dbcc9
**Verification:** All tests pass (`go test ./internal/domain/diff/...`), build succeeds, lint passes
**Notes:**
- Renamed `LineType` to `HunkLineType` (and constants to `HunkLineAdded`, `HunkLineRemoved`, `HunkLineModified`) to avoid conflict with existing `LineType` in parse.go which represents diff line types (Context, Add, Remove)
- The distinction is semantic: `LineType` in parse.go represents raw diff line types, while `HunkLineType` represents change types within a hunk segment
**Coordinator Review:** Structure matches design. Naming deviation to `HunkLineType` is appropriate to avoid collision. Types are well-documented. → Step 2

---

### Step 2: Build Segment Builder

**Goal:** Implement the algorithm that builds `[]Segment` from parsed diff data. This runs alongside the existing `BuildAlignments` logic initially.

**Structure:**
- New function in `internal/domain/diff/builder.go` or new file `segment_builder.go`
- Function: `BuildSegments(left, right FileContent, parsedDiff *FileDiff) []Segment`
- Algorithm:
  1. Walk through files using line mapping from parsed diff
  2. Accumulate consecutive unchanged lines → `UnchangedSegment`
  3. Accumulate consecutive changed lines → `HunkSegment` with `LeftLines`/`RightLines`
  4. Flush segments when transitioning between changed/unchanged regions
- Update `BuildAlignedFileDiff` to also populate `Segments` field

**Verify:**
- Unit tests with known diffs verifying segment boundaries are correct
- Test cases: pure additions, pure deletions, modifications, mixed changes, multiple hunks
- Tests pass: `go test ./internal/domain/diff/...`

**Read:**
- `internal/domain/diff/builder.go` (existing BuildAlignments logic)
- `internal/domain/diff/parser.go` (FileDiff structure)
- `02_design_smart-diff-scrolling.md` (segment building section)

**Status:** Complete
**Commits:** 4629a57
**Verification:** All tests pass (`go test ./...`), build succeeds (`go build ./...`), lint passes (`go tool golangci-lint run`)
**Notes:**
- Created `segment_builder.go` with `BuildSegments` function
- Algorithm follows the same pattern as `BuildAlignments`: builds line type maps, walks both files with two pointers
- Accumulates consecutive unchanged lines into `UnchangedSegment`, changed lines into `HunkSegment`
- All removed lines are marked `HunkLineRemoved`, all added lines are marked `HunkLineAdded` (no pairing/modified detection - deferred per design)
- Updated `BuildAlignedFileDiff` to call `BuildSegments` and populate the `Segments` field
- Comprehensive TDD test coverage: 14 test cases covering all scenarios (empty files, pure additions, pure deletions, mixed changes, multiple hunks, changes at start/end, lines outside diff context)
**Coordinator Review:** Algorithm structure is clean and follows established patterns. Correct use of line type maps and two-pointer walk. Edge cases handled properly. → Step 3

---

### Step 3: Update Diff State for Segment-Based Scrolling

**Goal:** Modify `DiffState` to use segment-based scroll position tracking instead of single `ViewportStart`.

**Structure:**
- Modify `internal/ui/states/diff/state.go`:
  - New fields: `SegmentIndex`, `LeftOffset`, `RightOffset` (replace or augment `ViewportStart`)
  - Preserve `ChangeIndices` for now (will be updated in Step 6)
- Update constructor `New()` to initialize segment-based position
- Helper methods:
  - `totalLeftLines()` / `totalRightLines()` - count lines per side
  - `lineAtOffset(segIdx, offset)` - get line at position for rendering

**Verify:**
- State initializes correctly with segment-based position
- Compile succeeds
- Existing tests continue to pass (may need adjustment)

**Read:**
- `internal/ui/states/diff/state.go` (current State struct)
- `02_design_smart-diff-scrolling.md` (state section)

**Status:** Complete
**Commits:** cab6cdb
**Verification:** All tests pass (`go test ./...`), build succeeds (`go build ./...`), lint passes (`go tool golangci-lint run`)
**Notes:**
- Added three new fields to State struct: `SegmentIndex`, `LeftOffset`, `RightOffset` for segment-based scroll position tracking
- Kept existing `ViewportStart` field for backward compatibility during migration
- Updated `New()` constructor to initialize segment position at first HunkSegment (or index 0 if no hunks)
- Added helper function `findFirstHunkSegmentIndex()` to locate first hunk
- Added four helper methods: `segmentLeftLineCount()`, `segmentRightLineCount()`, `totalLeftLines()`, `totalRightLines()`
- Created comprehensive unit tests in `state_test.go` covering all new functionality
- Deferred `lineAtOffset()` helper to Step 4 (rendering) as it wasn't needed for this step
**Coordinator Review:** State structure is clean with good separation of legacy vs new scroll tracking. Helper methods are well-documented and tested. Initialization correctly finds first hunk. → Step 4

---

### Step 4: Implement Segment-Based Rendering

**Goal:** Update `View()` to render using segments instead of alignments. Each panel collects lines independently.

**Structure:**
- Modify `internal/ui/states/diff/view.go`:
  - New method: `collectViewportLines(viewportHeight int) (leftLines, rightLines []renderedLine)`
    - Walks segments from current position
    - For `UnchangedSegment`: adds same content to both sides
    - For `HunkSegment`: adds left lines to left, right lines to right (may differ)
  - Update `View()` to use `collectViewportLines()` instead of looping over alignments
  - Handle line type styling (removed=red, added=green, unchanged=neutral)
- Simplification: No inline (word-level) diff highlighting in this step

**Verify:**
- Visual rendering matches expected output (hunks show without blank padding)
- Golden file tests for various diff scenarios
- Tests pass: `go test ./internal/ui/states/diff/...`

**Read:**
- `internal/ui/states/diff/view.go` (current rendering)
- `internal/ui/components/viewbuilder.go` (ViewBuilder API)
- `02_design_smart-diff-scrolling.md` (rendering section)

**Status:** Complete
**Commits:** 2eafb8c
**Verification:** All tests pass (`go test ./...`), build succeeds (`go build ./...`), lint passes (`go tool golangci-lint run`)
**Notes:**
- Added `collectViewportLines()` method that walks segments from current position and collects rendered lines for both panels
- Added `renderedLine` type to hold the formatted content for each line
- Added `hunkLineStyle()` helper to determine indicator and background style based on line type
- Added `formatFillerLine()` helper to create empty filler rows when one panel has fewer lines than the other
- Added `calculateSegmentLineNoWidth()` to compute line number width from segments
- Refactored `View()` to dispatch between segment-based rendering (`renderWithSegments`) and legacy alignment-based rendering (`renderWithAlignments`) based on whether segments are available
- When segments are available, the new rendering eliminates blank line padding - each panel renders its content independently with filler rows where needed
- Updated e2e golden files to reflect the new segment-based rendering output
- Created 5 new golden file tests for segment-based rendering:
  - `segment_pure_additions.golden` - hunk with only additions (right side has more lines)
  - `segment_pure_deletions.golden` - hunk with only deletions (left side has more lines)
  - `segment_mixed_changes.golden` - hunk with both additions and deletions (different line counts)
  - `segment_multiple_hunks.golden` - multiple hunks separated by unchanged regions
  - `segment_start_at_hunk.golden` - viewport starting at a hunk segment
**Coordinator Review:** Rendering structure is clean with good separation of concerns. `collectViewportLines` properly handles both segment types with filler lines for asymmetric hunks. Backward compatibility preserved via `renderWithAlignments`. Golden file tests provide good coverage. → Step 5

---

### Step 5: Implement Differential Scrolling Logic

**Goal:** Update scrolling to use differential rates when hunks are centered in viewport.

**Structure:**
- Modify `internal/ui/states/diff/update.go`:
  - New helper: `isHunkCentered(viewportHeight int) bool` - checks if current segment is a hunk at viewport center
  - New helper: `scrollDown(viewportHeight int)` - implements differential scroll logic
  - New helper: `scrollUp(viewportHeight int)` - symmetric for scrolling up
  - Update key handlers (`j`, `k`, `ctrl+d`, `ctrl+u`) to use new scroll methods
- Differential scroll rate:
  - When hunk centered: larger side scrolls every step, smaller side scrolls every N steps (N = ratio)
  - Track "scroll accumulator" for the slower side to determine when to advance

**Verify:**
- Single-line scroll (`j`/`k`) works with differential rates
- Half-page scroll (`ctrl+d`/`ctrl+u`) applies differential logic correctly
- Golden file tests showing scroll positions at various points
- Tests pass: `go test ./internal/ui/states/diff/...`

**Read:**
- `internal/ui/states/diff/update.go` (current scroll handling)
- `02_design_smart-diff-scrolling.md` (scrolling behavior section)
- `01_requirements_smart-diff-scrolling.md` (visual examples)

**Status:** Complete
**Commits:** feb426f
**Verification:** All tests pass (`go test ./...`), build succeeds (`go build -o splice .`), lint passes (`go tool golangci-lint run`)
**Notes:**
- Added `ScrollAccumulator` field to State struct to track fractional scroll progress
- Implemented `scrollDownSegment()` with differential scrolling logic:
  - For unchanged segments: both panels advance together
  - For hunks: larger side advances every step, smaller side advances every `ratio` steps
  - Transition to next segment when current segment is exhausted
- Implemented `scrollUpSegment()` with symmetric differential scrolling for scrolling up
- Implemented `isHunkCentered()` to detect when a hunk overlaps the viewport center zone (30%-70%)
- Implemented `isAtStart()` and `isAtEnd()` for bounds checking
- Implemented `resetToStart()` and `scrollToEnd()` for g/G navigation
- Added `calculateViewportHeight()` helper
- Updated all key handlers (`j`/`down`, `k`/`up`, `ctrl+d`, `ctrl+u`, `g`, `G`) to use segment-based scrolling when segments are available, with fallback to legacy alignment-based scrolling
- Comprehensive unit tests for all new functionality (13 new test functions)
- Differential scrolling uses integer ratio calculation: `ratio = max(leftCount, rightCount) / min(leftCount, rightCount)` rounded up
**Coordinator Review:** Differential scrolling logic is well-implemented. Clean separation between unchanged and hunk scrolling. `isHunkCentered()` correctly calculates center zone overlap. Scroll accumulator properly tracks when slower side should advance. Comprehensive test coverage with 13 new tests. → Step 6

---

### Step 6: Update Navigation (n/N Jump to Change)

**Goal:** Update change navigation to work with segment-based model.

**Structure:**
- Modify `internal/ui/states/diff/state.go` or `update.go`:
  - `ChangeIndices` becomes indices of `HunkSegment`s in `Segments` array (or recompute)
  - `jumpToNextHunk()` - finds next `HunkSegment`, sets `SegmentIndex`, resets offsets
  - `jumpToPreviousHunk()` - finds previous `HunkSegment`
- Update `n`/`N` key handlers to use new methods

**Verify:**
- `n` jumps to next hunk, `N` jumps to previous
- Works correctly at boundaries (first hunk, last hunk)
- Tests pass: `go test ./internal/ui/states/diff/...`

**Read:**
- `internal/ui/states/diff/update.go` (current jump logic)
- `02_design_smart-diff-scrolling.md` (navigation section)

**Status:** Complete
**Commits:** 3a3068b
**Verification:** All tests pass (`go test ./...`), build succeeds (`go build ./...`), lint passes (`go tool golangci-lint run`)
**Notes:**
- Added `jumpToNextHunkSegment()` method to state.go that searches forward from current segment for the next HunkSegment
- Added `jumpToPreviousHunkSegment()` method to state.go that searches backward from current segment for the previous HunkSegment
- Both methods reset `LeftOffset`, `RightOffset`, and `ScrollAccumulator` to 0 when jumping to a hunk
- Updated `n` and `N` key handlers in update.go to use segment-based navigation when segments are available, with fallback to legacy `jumpToNextChange`/`jumpToPreviousChange` methods
- Comprehensive test coverage: 15 new test functions covering all edge cases (finding next/previous hunk, boundary conditions, no hunks, starting from unchanged segment, key handler integration, legacy fallback)
**Coordinator Review:** Navigation implementation is clean. Both methods correctly search for HunkSegments and reset all offsets. Legacy fallback preserved. Test coverage is comprehensive. → Final Verification

---

## Final Verification

- [x] Full test suite passes: `go test ./...`
- [x] Lint passes: `go tool golangci-lint run`
- [x] Build succeeds: `go build -o splice .`
- [ ] Manual testing (to be done by developer):
  - [ ] Side-by-side diff shows no blank lines
  - [ ] Scrolling through unchanged regions: both panels scroll together
  - [ ] Scrolling through hunks: differential scrolling when hunk centered
  - [ ] `n`/`N` navigation jumps between hunks correctly
  - [ ] Works with various diff types (pure add, pure delete, mixed)

## Requirements Checklist

From `01_requirements_smart-diff-scrolling.md`:

- [x] Remove blank lines from the side-by-side diff view (Step 4: segment-based rendering with filler lines instead of blanks)
- [x] Implement differential scrolling so corresponding content stays aligned (Step 5: scrollHunkDifferential with ratio-based scrolling)
- [x] Keep the hunk centered in the viewport during differential scrolling (Step 5: isHunkCentered with 30%-70% center zone)
- [x] Maintain a smooth, intuitive scrolling experience (Step 5: integer ratio with accumulator for smooth progression)
- [x] Normal scrolling: Both panels scroll together when outside of hunks (Step 5: UnchangedSegment handling)
- [x] Differential scrolling at hunks: When a hunk has different sizes on each side (Step 5: HunkSegment handling with differential rates)
- [x] Symmetric behavior: Same logic applies when scrolling up (Step 5: scrollUpSegment and scrollHunkDifferentialUp)
- [x] Multiple hunks: Each hunk independently triggers differential scrolling (Step 5: per-segment state tracking)
- [x] Large hunks: Smaller side stays frozen while larger side scrolls through (Step 5: accumulator pattern freezes smaller side every N steps)
- [x] Page up/down: Same differential scrolling logic applies (Step 5: ctrl+d/ctrl+u call scroll methods multiple times)

## Summary

### What Was Built

A complete segment-based data model and rendering system for the side-by-side diff view that eliminates blank line padding and implements smart differential scrolling:

1. **Segment Data Model** (Step 1): `Segment` interface with `UnchangedSegment` and `HunkSegment` types, replacing the flat `Alignments` array for scroll position tracking.

2. **Segment Builder** (Step 2): `BuildSegments()` function that converts parsed diff data into the segment representation.

3. **Segment-Based State** (Step 3): Extended `DiffState` with `SegmentIndex`, `LeftOffset`, `RightOffset`, and `ScrollAccumulator` for independent panel tracking.

4. **Segment-Based Rendering** (Step 4): `collectViewportLines()` method that walks segments and renders each panel independently, using filler lines instead of blank padding.

5. **Differential Scrolling** (Step 5): Core scrolling logic that applies different scroll rates when hunks are centered in the viewport, based on the ratio of line counts between sides.

6. **Hunk Navigation** (Step 6): Updated `n`/`N` keys to jump between `HunkSegment`s instead of using the legacy `ChangeIndices` array.

### Deviations from Design

1. **HunkLineType naming**: Renamed from `LineType` to `HunkLineType` to avoid collision with existing `LineType` in `parse.go`. The distinction is semantic: parse `LineType` represents raw diff line types, while `HunkLineType` represents change types within a hunk segment.

2. **No inline diff highlighting**: As specified in the design, inline (word-level) diff highlighting was omitted from the initial implementation to focus on correct scrolling behavior. All changed lines use simple line-level styling (red for removed, green for added).

3. **Legacy compatibility**: Both alignment-based and segment-based systems coexist, with automatic fallback to legacy behavior when segments aren't available. This allows for safe migration and easier testing.
