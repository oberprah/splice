# CLAUDE.md

This file provides guidance for the Rust implementation of Splice.

## Overview

This is a port of the Go implementation (`../go/`). The Go codebase is the source of truth for behavior. See `ARCHITECTURE.md` for architectural decisions and patterns.

**Important:** Read `rust/ARCHITECTURE.md` before working on the Rust implementation.

## Development Commands

```bash
cargo run                  # Run application
cargo test                 # Run all tests
cargo test --test <name>   # Run specific test file
cargo clippy               # Lint
```

## Testing

- **Domain tests**: Pure unit tests, no git needed
- **Git tests**: Use `tempfile` to create real git repos
- **UI tests**: Snapshot tests with `insta`, mocked git

## Project Structure

```
src/
├── core/           # Shared types (enums, traits)
├── domain/         # Pure business logic (diff, graph, filetree)
├── git/            # Git operations
└── ui/             # UI layer
```

## Key Principles

- **Port from Go**: Reference `../go/` for behavior
- **Use enums**: Rust enums replace Go's sealed interfaces
- **Real git for integration tests**: Better than mocking
- **Snapshot UI tests**: Deterministic with mocked git
