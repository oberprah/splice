# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Splice is a terminal-based git diff viewer built with Go and Bubbletea. The goal is to provide a superior diff viewing experience compared to existing tools, with easy distribution as a single binary.

## Technology Stack

- **Go** (primary language)
- **Bubbletea** (TUI framework using The Elm Architecture: Model-Update-View pattern)
- **Charm Bracelet ecosystem** (Bubbles, Lip Gloss, Glamour for UI components)

See `docs/adr/adr-002-acc-go-bubbletea-stack.md` for the rationale behind this stack choice.

## Development Commands

```bash
# Run the application
go run main.go

# Build binary
go build -o splice main.go

# Add dependencies
go get <package>

# Update dependencies
go mod tidy
```

## Architecture

The application follows the Bubbletea (Elm Architecture) pattern:
- **Model**: Application state
- **Init**: Initialize the program
- **Update**: Handle messages (keyboard input, events) and update state
- **View**: Render the UI based on current state

Currently, the codebase contains a minimal hello world implementation in `main.go` that demonstrates this pattern.
