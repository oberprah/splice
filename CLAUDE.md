# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Splice is a terminal-based git diff viewer built with Go and Bubbletea. The goal is to provide a superior diff viewing experience compared to existing tools, with easy distribution as a single binary.

## Technology Stack

- **Go** (primary language)
- **Bubbletea** (TUI framework using The Elm Architecture: Model-Update-View pattern)
- **Charm Bracelet ecosystem** (Bubbles, Lip Gloss, Glamour for UI components)

See `docs/adr/adr-002-acc-go-bubbletea-stack.md` for the rationale behind this stack choice.

## Project Structure

The project uses a simplified structure with `main.go` at the root and `/internal` for private application code.

UI components follow the Elm Architecture (Model-Update-View) with states organized into subpackages:
- Each state (loading, list, error) has its own package under `internal/ui/state/`
- Within each state package: `state.go` (struct), `view.go` (rendering), `update.go` (event handling)
- States implement the `state.State` interface and receive a `state.Context` for accessing model properties

## Architecture & Data Flow

The app is a state machine where each screen (loading, log, files, diff) is a separate state:

```
LoadingState → LogState → FilesState → DiffState
                  ↑______________|          |
                  |_________________________|
```

**State file organization** (`internal/ui/states/`):
- `*_state.go` - State struct definition
- `*_view.go` - Rendering logic (View method)
- `*_update.go` - Event handling (Update method)

**Async data loading pattern**:
1. User action triggers `tea.Cmd` (e.g., `loadDiff()` in `files_update.go`)
2. Cmd executes async, returns a message (e.g., `DiffLoadedMsg`)
3. Message routed to current state's `Update()` method
4. Update returns new state + optional new command

**Message definitions**: `internal/ui/messages/messages.go`

When modifying a state's data structure, typically need to update:
- The message struct in `messages.go`
- The state struct in `*_state.go`
- View rendering in `*_view.go`
- The code that creates the message (e.g., `loadDiff()`)
- Corresponding `*_test.go` files

## Development Commands

```bash
go run .                        # Run application
go build -o splice .            # Build binary

go tool golangci-lint run       # Lint

go mod tidy                     # Update dependencies

go test ./...                   # Run tests
go test ./... -update           # Update golden files
```
