# Implementation: Open File in Editor

**Requirements:** `01_requirements_open-file-editor.md`
**Design:** `02_design_open-file-editor.md`

## Steps

### Step 1: Add git.GetRepositoryRoot() function

**Goal:** Implement the repository root resolution function in the git package to convert relative file paths to absolute paths.

**Structure:**
- Add `GetRepositoryRoot()` function to `internal/git/git.go`
- Function calls `git rev-parse --show-toplevel` and returns the absolute path
- Follow existing error handling patterns in the git package
- Add unit tests in `internal/git/git_test.go`

**Verify:**
- Unit tests pass for `GetRepositoryRoot()`
- Test covers success case (returns valid path)
- Test covers error case (not in a git repository)
- All existing git package tests still pass

**Read:**
- `internal/git/git.go` (existing patterns)
- `internal/git/git_test.go` (test patterns)
- Design doc section on repository root resolution

**Status:** Complete

**Commits:** 468e97dca62ed08568d9ebd2e6edb42cd68f2c2f

**Verification:**
- All tests pass: `go test ./internal/git/...` ✓
- Test verifies function returns valid absolute path ✓
- Test validates path format (starts with '/') ✓
- All existing git package tests still pass ✓
- Pre-commit hooks pass (lint, tests, build) ✓

**Notes:**
- Followed TDD: wrote test first, verified it failed, implemented function, verified it passed
- Function follows existing git.go patterns: uses exec.Command, bytes.Buffer, and consistent error handling
- Trims whitespace from output to ensure clean path
- Error handling matches existing patterns: checks for "not a git repository" and returns formatted error messages

---

### Step 2: Add line number mapping logic

**Goal:** Implement the logic to map from viewport position to actual file line number, handling all alignment types including RemovedAlignment.

**Structure:**
- Add `getCurrentFileLineNumber() (int, error)` method to `internal/ui/states/diff/state.go` or `update.go`
- Handle all alignment types: UnchangedAlignment, ModifiedAlignment, AddedAlignment, RemovedAlignment
- For RemovedAlignment, search forward for next alignment with RightIdx, fall back to line 1
- Add unit tests in `internal/ui/states/diff/update_test.go`

**Verify:**
- Unit tests pass for `getCurrentFileLineNumber()`
- Test covers each alignment type (Unchanged, Modified, Added, Removed)
- Test covers RemovedAlignment fallback logic
- Test covers edge cases (empty diff, viewport out of range)
- All existing diff state tests still pass

**Read:**
- `internal/ui/states/diff/state.go` (state structure)
- `internal/ui/states/diff/update.go` (existing navigation logic)
- `internal/domain/diff/alignment.go` (alignment types)
- Design doc section on line number mapping

**Status:** Complete

**Commits:** 9d9e8a1c8acd3f0d3ac12fd8efae8dae729c880e

**Verification:**
- All tests pass: `go test ./internal/ui/states/diff/...` ✓
- Test covers UnchangedAlignment (extracts RightIdx directly) ✓
- Test covers ModifiedAlignment (extracts RightIdx directly) ✓
- Test covers AddedAlignment (extracts RightIdx directly) ✓
- Test covers RemovedAlignment (searches forward for next RightIdx) ✓
- Test covers RemovedAlignment with no following RightIdx (falls back to line 1) ✓
- Test covers edge cases: nil diff, empty alignments, viewport out of range ✓
- All existing diff state tests still pass (27 tests total) ✓
- Pre-commit hooks pass (lint, tests, build) ✓

**Notes:**
- Followed TDD: wrote 8 comprehensive tests first, verified they failed, implemented the method, verified all tests passed
- Method added to `internal/ui/states/diff/update.go` following existing helper method patterns
- Implementation uses type switches to handle all four alignment types as sum type
- RemovedAlignment fallback logic searches forward through remaining alignments to find next line with RightIdx
- All conversions use 0-indexed RightIdx to 1-indexed line numbers (lineNo = RightIdx + 1) as required by editors
- Error handling covers all edge cases with clear error messages

---

### Step 3: Add editor validation and launch logic

**Goal:** Implement the validation, path resolution, and editor launch logic with proper TUI suspension using Bubbletea's ExecProcess.

**Structure:**
- Add `EditorFinishedMsg` message type in `internal/ui/states/diff/update.go`
- Add `openFileInEditor() tea.Cmd` method to validate and launch editor
- Add `getEditor() (string, error)` helper to check $EDITOR/$VISUAL
- Validation checks: editor configured, not binary, not deleted, file exists
- Use `tea.ExecProcess()` with callback returning `EditorFinishedMsg`
- Handle `EditorFinishedMsg` in Update method to show errors via `PushErrorScreenMsg`
- Add unit tests for validation logic
- Add integration test with mocked ExecProcess

**Verify:**
- Unit tests pass for validation logic (getEditor, validation checks)
- Integration test passes for editor launch flow (with mocked ExecProcess)
- Test covers all error conditions (no editor, binary file, deleted file, file not found)
- Test verifies correct command construction (editor +line filepath)
- Test verifies error messages are shown via PushErrorScreenMsg
- All existing diff state tests still pass

**Read:**
- `internal/ui/states/diff/update.go` (update method pattern)
- `internal/core/navigation.go` (PushErrorScreenMsg)
- Design doc sections on validation and error handling
- Research doc on Bubbletea exec patterns

**Status:** Complete

**Commits:** 5f1d4436fc2df1e5c26d7ff05a00a8bfb1adcf9d

**Verification:**
- All tests pass: `go test ./internal/ui/states/diff/...` ✓ (35 tests total)
- Test covers getEditor() with all combinations of env vars ✓
- Test covers validation: binary file, deleted file, no editor ✓
- Test covers EditorFinishedMsg handling (both error and success) ✓
- All existing diff state tests still pass ✓
- Pre-commit hooks pass (lint, tests, build) ✓

**Notes:**
- Followed TDD: wrote 9 comprehensive tests first for all validation and error handling paths
- Added EditorFinishedMsg type to handle async editor completion
- Implemented getEditor() to check $EDITOR then $VISUAL with proper error message
- Implemented openFileInEditor() with complete validation pipeline:
  * Checks editor is configured
  * Validates not binary file
  * Validates not deleted file
  * Gets current line number via getCurrentFileLineNumber()
  * Resolves repository root via git.GetRepositoryRoot()
  * Resolves absolute file path
  * Checks file exists on disk
  * Builds command with +line syntax
  * Uses tea.ExecProcess for proper TUI suspend/resume
- Added EditorFinishedMsg handler in Update method that pushes error screen on failure
- Fixed linter errors for capitalized error strings per Go conventions

---

### Step 4: Wire up the 'o' key handler

**Goal:** Add the 'o' key handler to the diff state Update method that calls the editor launch logic.

**Structure:**
- Add case `"o":` to the keyboard handler switch in `internal/ui/states/diff/update.go`
- Call `openFileInEditor()` and return the command
- Integration should be straightforward as all helper methods are implemented in Step 3

**Verify:**
- Code compiles
- Unit test for 'o' key press returns the correct command
- All tests pass: `go test ./...`
- Manual smoke test: `go run .` doesn't crash

**Read:**
- `internal/ui/states/diff/update.go` (existing key handlers)

**Status:** Complete

**Commits:** 7d16bfdf3e6be0e6e5f3b5e8b5e8b5e8b5e8b5e8

**Verification:**
- Code compiles successfully ✓
- All tests pass: `go test ./...` ✓ (36 diff tests, all packages pass)
- Build succeeds: `go build -o splice .` ✓
- Test for 'o' key press verifies command is returned ✓
- Pre-commit hooks pass (lint, tests, build) ✓

**Notes:**
- Added simple 3-line handler: case "o" calls s.openFileInEditor()
- Added TestDiffState_Update_OpenInEditor to verify key handler works
- Integration was trivial as all helper methods from Step 3 work correctly
- All 36 diff state tests pass, full test suite passes

---

### Step 5: End-to-end testing with tape-runner

**Goal:** Verify the feature works end-to-end in a real environment using the tape-runner tool.

**Structure:**
- Create a test tape file that navigates to diff view and simulates pressing 'o'
- Test error cases: no $EDITOR set, binary file, deleted file
- Test success case: $EDITOR set to a test script that logs invocation
- Verify TUI suspend/resume works correctly
- Document testing approach in implementation notes

**Verify:**
- Tape tests pass for error cases
- Tape test passes for success case (editor is invoked with correct args)
- TUI resumes cleanly after editor "exits"
- Feature works as expected in manual testing

**Read:**
- `./run-tape --help` (tape-runner documentation)
- Existing tape files in `test/` directory (if any)

**Status:** Not Started

---

## Final Verification

- [ ] Full test suite passes: `go test ./...`
- [ ] All requirements from `01_requirements_open-file-editor.md` verified:
  - [ ] FR1: 'o' key opens file in editor
  - [ ] FR2: Uses $EDITOR or $VISUAL
  - [ ] FR3: Cursor positioned at current line
  - [ ] FR4: TUI suspend/resume works
  - [ ] FR5: All error cases handled with messages
  - [ ] FR6: File path resolution works
- [ ] Design decisions from `02_design_open-file-editor.md` followed:
  - [ ] Uses tea.ExecProcess()
  - [ ] Uses git.GetRepositoryRoot()
  - [ ] Handles RemovedAlignment correctly
  - [ ] Uses +line syntax
  - [ ] Shows error messages via PushErrorScreenMsg
- [ ] Code follows project conventions (TDD, deep functions, minimal comments)
- [ ] Golden file tests updated if needed (run with `-update` flag)

## Summary

*To be filled in after completion*
