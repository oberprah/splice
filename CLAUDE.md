# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Splice is a terminal-native git log/diff viewer. Lean interface, intuitive navigation, fast keyboard-driven workflow.

This repository contains two implementations:
- `go/` - Original Go implementation (stable)
- `rust/` - Rust experiment with ratatui (in progress)

See the respective CLAUDE.md files in each directory for implementation-specific guidance.

## Development Commands

### Go

```bash
cd go && go run .                        # Run application
cd go && go test ./...                   # Run tests
```

### Rust

```bash
cd rust && cargo run                     # Run application
cd rust && cargo test                    # Run tests
```

Setup git hooks (runs lint, tests, build on commit):
```bash
git config core.hooksPath .githooks
```
