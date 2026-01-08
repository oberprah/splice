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
**Coordinator Review:**

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

**Status:** Pending
**Commits:**
**Verification:**
**Notes:**
**Coordinator Review:**

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

**Status:** Pending
**Commits:**
**Verification:**
**Notes:**
**Coordinator Review:**

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

**Status:** Pending
**Commits:**
**Verification:**
**Notes:**
**Coordinator Review:**

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

**Status:** Pending
**Commits:**
**Verification:**
**Notes:**
**Coordinator Review:**

---

## Final Verification

- [ ] Full test suite passes: `go test ./...`
- [ ] Lint passes: `go tool golangci-lint run`
- [ ] Build succeeds: `go build -o splice .`
- [ ] Manual testing:
  - [ ] Side-by-side diff shows no blank lines
  - [ ] Scrolling through unchanged regions: both panels scroll together
  - [ ] Scrolling through hunks: differential scrolling when hunk centered
  - [ ] `n`/`N` navigation jumps between hunks correctly
  - [ ] Works with various diff types (pure add, pure delete, mixed)

## Requirements Checklist

From `01_requirements_smart-diff-scrolling.md`:

- [ ] Remove blank lines from the side-by-side diff view
- [ ] Implement differential scrolling so corresponding content stays aligned
- [ ] Keep the hunk centered in the viewport during differential scrolling
- [ ] Maintain a smooth, intuitive scrolling experience
- [ ] Normal scrolling: Both panels scroll together when outside of hunks
- [ ] Differential scrolling at hunks: When a hunk has different sizes on each side
- [ ] Symmetric behavior: Same logic applies when scrolling up
- [ ] Multiple hunks: Each hunk independently triggers differential scrolling
- [ ] Large hunks: Smaller side stays frozen while larger side scrolls through
- [ ] Page up/down: Same differential scrolling logic applies

## Summary

(To be completed after implementation)

- What was built
- Deviations from design (with rationale)
