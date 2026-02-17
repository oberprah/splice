# CLAUDE.md

This file provides guidance for the Rust implementation of Splice.

## Overview

This is a port of the Go implementation (`../go/`). The Go codebase is the source of truth for behavior. See `ARCHITECTURE.md` for architectural decisions and patterns.

**Important:** Read `rust/ARCHITECTURE.md` before working on the Rust implementation.

## Development Commands

```bash
cargo run                  # Run application (uses current directory)
cargo run -- /path/to/repo # Run with specific repo path
cargo test                 # Run all tests
cargo clippy               # Lint
```

## Testing

- **Unit tests**: In source files via `#[cfg(test)] mod tests`
- **Integration tests**: Real git repos via `TestRepo` helper
- **E2E tests**: Snapshot tests with `insta`

## Project Structure

```
src/
├── core/           # Shared types (enums, traits)
├── git/            # Git operations
├── input.rs        # Event → Action mapping
├── app.rs          # App state + reducer/update logic
└── ui/             # UI layer (pure rendering)
```

## Key Principles

- **Port from Go**: Reference `../go/` for behavior
- **Use enums**: Rust enums replace Go's sealed interfaces
- **Real git for tests**: `tempfile` creates isolated repos
- **Deterministic test data**: `TestRepo` uses fixed git env vars for predictable hashes

## Ratatui Best Practices (for this repo)

- **Pure rendering**: `ui::render` should be read-only and deterministic
- **Action mapping**: map keys/events → `Action`, then apply `App::update`
- **Controlled redraw**: render on state change or resize, not every loop
- **Safe terminal cleanup**: guard + panic hook to restore raw mode/screen
- **Layout clarity**: compute list height from area minus footer rows
