# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Splice is a terminal-native git log/diff viewer. Lean interface, intuitive navigation, fast keyboard-driven workflow.

## Development Commands

```bash
go run .                        # Run application
go build -o splice .            # Build binary
go tool golangci-lint run       # Lint
go test ./...                   # Run all tests
go test ./internal/ui/states/log/...  # Run tests for a specific package
go test ./... -update           # Update golden files
```

Setup git hooks (runs lint, tests, build on commit):
```bash
git config core.hooksPath .githooks
```

## Testing the Compiled Binary (For AI Agents)

**DO NOT run `./splice` directly** - it requires a real terminal and will fail. Instead, use the test-app wrapper:

```bash
./test-app -c initial 'jjj' -c after-nav '<enter>' -c 'q' -c final
```

This creates screenshots at each checkpoint in `test-output/TIMESTAMP/`:
- `001-initial.png` - Initial state after app loads
- `002-after-nav.png` - After navigation ('jjj')
- `003.png` - After pressing enter
- `004-final.png` - After quit

**Options**:
- `-c` or `--checkpoint [name]` - Take screenshot (optional name)
- `--width 1200` - Screenshot width in pixels (default: 1200)
- `--height 800` - Screenshot height in pixels (default: 800)
- `--font-size 12` - Terminal font size (default: 12)

**Special keys**:
- `<enter>`, `<esc>`, `<tab>`, `<space>`, `<backspace>`
- `<up>`, `<down>`, `<left>`, `<right>`
- `<ctrl-c>`

**Examples**:
```bash
# Simple test with 2 checkpoints
./test-app -c 'j' -c 'q' -c

# Named checkpoints for clarity
./test-app -c initial 'jjj' -c after-nav '<enter>' -c selected 'q' -c

# Custom dimensions for more content
./test-app --width 1600 --height 1000 --font-size 14 -c 'jj' -c

# Build first if needed
go build -o test-app ./cmd/test-app
```

**Understanding the syntax**:
Actions come BEFORE the checkpoint:
```bash
./test-app 'jjj' -c after-nav 'q' -c
#          ^^^^^ screenshot  ^^^ screenshot
#          actions first     actions then screenshot
```

**Verification**:
Review screenshots to verify the app behaved correctly. Screenshots are ~500KB PNG files at 1200×800.

## Package Architecture

The project follows a layered architecture with strict import rules:

```
internal/
├── core/       # Interfaces and messages (no dependencies on other internal packages)
├── domain/     # Pure business logic (diff parsing, graph layout, syntax highlighting)
├── git/        # Git command execution
├── ui/         # UI layer
│   ├── states/     # Screen states (loading, log, files, diff, error)
│   ├── components/ # Reusable view components
│   ├── styles/     # Lip Gloss styling
│   ├── format/     # Formatting utilities
│   └── testutils/  # Test helpers and mocks
└── app/        # Application model and Bubbletea integration
```

**Import direction**: `app` → `ui/states` → `ui/components` → `domain` → `git` (`core/` is shared by all layers)

## State Machine Architecture

The app is a navigation stack where each screen is a separate state implementing `core.State`:

```
LoadingState → LogState ⇄ FilesState ⇄ DiffState
```

Navigation uses typed messages (`core.Push*ScreenMsg`, `core.PopScreenMsg`) handled by `app.Model`.

**State file organization** (`internal/ui/states/<name>/`):
- `state.go` - State struct and constructor
- `view.go` - Rendering logic (View method)
- `update.go` - Event handling (Update method)

**Async data loading pattern**:
1. User action triggers `tea.Cmd` (e.g., `loadDiff()`)
2. Cmd executes async, returns a message (e.g., `DiffLoadedMsg`)
3. Message routed to current state's `Update()` method
4. Update returns `Push*ScreenMsg` to navigate or new state + command

## Coding Principles

- **Deep functions**: Prefer fewer, more capable functions over many shallow ones
- **Pure functions**: Favor pure functions; isolate side effects at boundaries
- **Minimal comments**: Code should be self-explanatory; comment only for non-obvious "why"
- **Make illegal states unrepresentable**: Use sum types and type design to prevent invalid states
- **Enums over multiple booleans**: When multiple booleans represent related states, use an enum instead. Multiple booleans create 2^n possible states, most of which are usually invalid.
- **TDD**: Write tests first when implementing new functionality

## Testing

**Read `docs/guidelines/testing-guidelines.md` before implementing tests.**

- **Unit tests**: Most functionality
- **Golden file tests**: TUI rendering (run with `-update` to regenerate)
- **E2E tests**: Full user workflows (`test/e2e/`)

All external dependencies (git commands, time) are mocked via functional options on `app.Model`. Test helpers are in `internal/ui/testutils/`.

**Golden file updates**: After running `go test ./... -update`, always review the git diff of `.golden` files to verify changes are intentional before committing.
