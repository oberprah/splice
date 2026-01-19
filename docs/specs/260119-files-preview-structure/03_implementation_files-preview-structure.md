# Implementation: Files Preview Tree Structure

**Requirements:** `01_requirements_files-preview-structure.md`
**Design:** `02_design_files-preview-structure.md`

## Steps

### Step 1: Refactor TreeSection to use nullable cursor

**Goal:** Change `TreeSection` component from using magic number (-1) to nullable cursor pointer (`cursor *int`), improving type safety and code clarity.

**Structure:**
- Component: `internal/ui/components/tree_section.go`
- Tests: `internal/ui/components/tree_section_test.go`
- Consumer: `internal/ui/states/files/view.go` (update to pass `&s.Cursor`)
- Consumer tests: `internal/ui/states/files/view_test.go` (golden files may need update)

**Verify:**
- All existing TreeSection tests pass with updated signature
- Add test case for `nil` cursor (no selection)
- Files state view tests pass (golden files updated if needed)
- Build succeeds: `go build -o splice .`

**Read:**
- `02_design_files-preview-structure.md` (refactoring section)
- `internal/ui/components/tree_section.go`
- `internal/ui/components/tree_section_test.go`
- `internal/ui/states/files/view.go`

**Status:** Complete

**Commits:** d07ca1d

**Verification:**
- All TreeSection tests pass
- Added new test case `TestTreeSection_NilCursor` that verifies nil cursor displays no selection
- All files state view tests pass (no golden file changes needed)
- Build succeeds: `go build -o splice .`
- Pre-commit hooks pass (lint, tests, build)

**Notes:**
- Changed `TreeSection` signature from `cursor int` to `cursor *int`
- Updated implementation to check `cursor != nil && *cursor == i` instead of `i == cursor`
- Updated all test cases to pass cursor as `&cursorValue` instead of `cursorValue`
- Added `TestTreeSection_NilCursor` test with golden file verification
- Updated files state to pass `&s.Cursor` instead of `s.Cursor`
- No changes to golden files were needed for existing tests (signature change doesn't affect output)
- The nil cursor test shows tree items with no selection indicator, as expected

---

### Step 2: Implement tree structure in log preview

**Goal:** Replace `FileSection` with `TreeSection` in the log view's preview panel, adding tree structure display with all folders expanded.

**Structure:**
- Component: `internal/ui/states/log/view.go`
  - Add `buildTreeForPreview()` helper function
  - Update `renderFileList()` method to use TreeSection
  - Update truncation message from "files" to "items"
- Tests: `internal/ui/states/log/view_test.go` (golden files will need update)

**Verify:**
- Log view tests pass with updated golden files
- Visual verification: Tree structure appears in preview panel
- Build succeeds: `go build -o splice .`
- Manual test with `./run-tape` (if available) to verify rendering

**Read:**
- `02_design_files-preview-structure.md` (tree building and rendering sections)
- `internal/ui/states/log/view.go` (especially `renderFileList()` method around line 191)
- `internal/domain/filetree/` package (understand BuildTree, CollapsePaths, ApplyStats, FlattenVisible)
- Step 1 completion notes (TreeSection signature)

**Status:** Pending

---

### Step 3: Delete FileSection component

**Goal:** Remove the now-unused `FileSection` component and its tests, reducing code maintenance burden.

**Structure:**
- Delete: `internal/ui/components/file_section.go`
- Delete: `internal/ui/components/file_section_test.go`

**Verify:**
- No compilation errors after deletion
- No references to `FileSection` remain in codebase: `git grep -i "filesection"`
- Full test suite passes: `go test ./...`
- Build succeeds: `go build -o splice .`

**Read:**
- Verify no other references exist before deleting

**Status:** Pending

---

## Final Verification

- [ ] Full test suite passes: `go test ./...`
- [ ] Build succeeds: `go build -o splice .`
- [ ] All requirements from `01_requirements_files-preview-structure.md` verified:
  - [ ] Tree structure displays in preview panel
  - [ ] All folders expanded by default in preview
  - [ ] No cursor/selection in preview (read-only)
  - [ ] Same tree symbols and indentation as files view
  - [ ] File stats display unchanged
- [ ] Design decisions from `02_design_files-preview-structure.md` followed:
  - [ ] TreeSection uses `cursor *int` (no magic number)
  - [ ] Log preview uses `buildTreeForPreview()` helper
  - [ ] FileSection deleted entirely
  - [ ] Performance acceptable (tree building < 1ms for typical commits)

## Summary

_To be completed after implementation_
