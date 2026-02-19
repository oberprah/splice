# ADR-001: Technology Stack Selection

* **Status**: Superseded by ADR-002
* **Date**: 2025-09-27
* **Superseded**: 2025-10-15
* **Problem**: We need to choose a technology stack for the project.
* **Decision**: ~~Use TypeScript + React + Ink as the primary technology stack~~ (Superseded)

## Context

We are building a terminal-based git diff viewer called Splice. The application needs to provide a superior diff viewing experience compared to existing tools like delta, tig, and IntelliJ's diff view, while maintaining the flexibility and customization options that come with building from scratch.

Key requirements:
- Terminal User Interface (TUI) application
- Good diff viewing capabilities with side-by-side view
- Overview of all changed files
- Keyboard shortcuts for navigation
- Fast and responsive performance

## Decision

We will use the following technology stack (similar to what Claude code is using)
- **TypeScript** - Primary programming language
- **React with Ink** - UI framework for building terminal interfaces
- **Bun** - Build tool and package manager

## Rationale

### TypeScript + React + Ink

- **"On distribution" for AI models**: TypeScript and React are technologies that AI models like Claude are very capable with, enabling better development assistance and potentially allowing the tool to "build itself"
- **Familiar technology**: TypeScript is well-known and widely adopted
- **Proven in similar applications**: Claude Code successfully uses this exact stack for their terminal application
- **React paradigm**: Allows for component-based UI development even in terminal environments
- **Ink framework**: Specifically designed for building interactive command-line applications with React

### Alternative Considerations Rejected

- **bubbletea (Go)**: While powerful, Go is less familiar and would be "off distribution" for AI assistance
- **ratatui (Rust)**: Also unfamiliar territory with less AI model support
- Both alternatives would require more manual learning and implementation effort

### Build Tools

- **Bun**: Chosen for build speed compared to Webpack, Vite, and other build systems

## Consequences

### Positive

- Leverages AI development capabilities effectively
- Uses familiar, well-documented technologies
- Proven stack in similar terminal applications
- Strong ecosystem and community support
- Component-based architecture enables maintainable code

### Negative

- JavaScript/TypeScript ecosystem can be complex and heavy - compared to Go and rust.
- Ink framework has learning curve for terminal-specific patterns
- Dependency on Node.js runtime

### Neutral

- Will require learning Ink-specific patterns for terminal UI development

## Reason for Superseding

After further analysis of distribution requirements and performance characteristics, the decision was reconsidered. See ADR-002 for the updated technology stack decision.
