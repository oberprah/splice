# AGENTS.md

**This is a living document**. Update it often, but keep it lean. Update it when you learn something that would have helped you move faster, when the UI changes, or when guidance becomes outdated.

## Project Overview

Splice is a terminal-native git log/diff viewer. It is built for fast, keyboard-driven inspection of commit history and file changes, with a lean interface and clear visual hierarchy. The goal is to make navigating complex histories easy without leaving the terminal.

## UI Overview

Quick snapshots show how users navigate log, files, and diff views.

**Log View** - Commit graph and merge topology.
```text
"      Uncommitted changes · 2 files                                           "
"  → ├ a1b2c3d (main) Fix log view spacing                                        "
"    ├─╮ 8c9d0e1 Merge feature/search                                            "
"    │ ├ 5f6a7b8 (feature/search) Add fuzzy match                                "
"    ├─╯ 1d2e3f4 Update README examples                                          "
"  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
```

**Files View** - Expanded file tree for the selected commit with status and line counts.
```text
"  e2af8ce Modify and add files                                                  "
"  3 files · +5 -1                                                               "
"  →├── src/                                                                     "
"   │   ├── M +3 -1  main.rs                                                     "
"   │   └── A +1 -0  new.rs                                                      "
"   └── A +1 -0  file_1.txt                                                      "
"  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
```

**Diff View** - Side-by-side diff showing additions and deletions as you scroll.
```text
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"    1   pub struct Calculator;        │   1   pub struct Calculator;            "
"    2                                 │   2                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"   17 -         a * b                 │  17 +         a.checked_mul(b).unwra    "
"  j/k: scroll  n/p: next/prev diff  q: back                                     "
```

More expressive snapshots live inline in E2E tests under `tests/e2e/`. Treat these snapshots as system documentation for expected UI behavior, and check them when changing UI.

## Architecture

### Layout

```
src/
├── cli/        # Argument parsing; dispatches to log or diff entry point
├── core/       # Shared primitives: Commit, FileChange, DiffSource, CursorState
├── git/        # All git operations: fetching, parsing, resolving
├── domain/     # Business logic: diff building, graph layout, file tree
├── app/        # App state and view states (LogView, FilesView, DiffView)
├── lib.rs       # Crate root used by tests and shared exports
├── main.rs      # Binary entry point and event loop
├── input.rs    # Raw events -> Action mapping
└── ui/         # Rendering only; reads app state, draws to terminal
```

`tests/` mirrors this with `integration_tests/` (git layer), `e2e/` (full app via `Harness`), `common/` (test utilities like `Harness` and `TestRepo`), and unit tests inline in source files.

### Flow

`cli` parses args and resolves the diff source if needed, then hands off to `run_log_app` or `run_diff_app`. Inside the event loop, input events map to `Action`, `App::update` mutates state, and `ui::render` draws from state.

### Import Direction

- `core/` depends on nothing internal
- `git/` and `domain/` depend only on `core/`
- `app/` depends on `core/`, `git/`, `domain/`, and `input.rs`
- `ui/` depends on `core/`, `app/`, and `domain/` — never calls `git/` directly

### Principles

- **Deterministic rendering** — `ui/` renders from state and never fetches git data. Given the same state, output stays predictable and testable.
- **Action-driven updates** — all input maps to an `Action` before anything mutates. This keeps the event loop simple and makes behavior easy to trace.
- **View stack navigation** — `App` holds a `Vec<View>`; pushing and popping drives all navigation. Never hard-code transitions between views.
- **Make illegal states unrepresentable** — use enums (`View`, `DiffSource`, `CursorState`) instead of flags or multiple booleans. Invalid combinations should not compile.
- **Deep functions over shallow layers** — prefer one function that does the full job over several thin wrappers that just delegate. Abstraction should earn its place.
- **Controlled redraw** — render only on state change or resize, not every tick. The `should_render` flag in the event loop enforces this.
- **Safe terminal cleanup** — `TerminalGuard` and a panic hook ensure raw mode and the alternate screen are always restored, even on crash.

### Testing

Test behavior, not implementation — assert visible output and user-facing effects.

| Scope | Type | Location |
|-------|------|----------|
| `git/` operations | Integration | `tests/integration_tests/` — real git repos via `TestRepo` |
| Full app flows | E2E snapshots | `tests/e2e/` — drive input via `Harness`, assert rendered output |
| Logic in `core/`, `app/`, etc. | Unit | Inline `#[cfg(test)] mod tests` in source files |

**Key rules:**
- **Use `Harness` for E2E tests** — it wraps the terminal backend, drives key events, and captures rendered output as a string
- **Always use `TestRepo` for git operations** — never touch the current repo
- **`#[serial]` + `reset_counter()`** — tests using `TestRepo` must use both; the commit hash is deterministic only when tests run serially with a reset counter
- **Snapshots are updated manually** — there is no auto-update; run the test, read the diff in the panic output, edit the snapshot string
- **E2E snapshots must be full-screen** — always assert the entire rendered output, never just a single line or substring
- **Prefer deep E2E tests** — one test that simulates a real user flow (navigate, open, scroll, go back) is better than multiple small tests each covering one action

### Commands

```bash
cargo run                  # Run app (uses current directory)
cargo run -- /path/to/repo # Run with specific repo
cargo test                 # Run all tests
cargo test <name>          # Run a specific test
cargo clippy               # Lint
cargo fmt                  # Format
```
