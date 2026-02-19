# ADR-002: Technology Stack Revision - Go + Bubbletea

* **Status**: Accepted
* **Date**: 2025-10-15
* **Supersedes**: ADR-001
* **Problem**: Need to reconsider technology stack with emphasis on distribution and performance
* **Decision**: Use Go + Bubbletea as the primary technology stack

## Context

After the initial decision to use TypeScript + Ink (ADR-001), we reconsidered the requirements with emphasis on:
- **Distribution**: Easy installation via Homebrew with minimal footprint
- **Performance**: Handling large git repositories and complex diffs efficiently
- **User Experience**: Zero-dependency installation

## Decision

Technology stack:
- **Go** - Primary programming language
- **Bubbletea** - TUI framework based on The Elm Architecture
- **Charm Bracelet ecosystem** - Supporting libraries (Bubbles, Lip Gloss, Glamour)

## Key Comparisons

### Stability
Both Bubbletea (v2.0, 9k+ projects) and Ink (v6.3, used by GitHub/Claude Code/Prisma) are production-ready and stable.

### Performance
- Go provides ~34% better performance in benchmarks, 2x in stress tests
- Significantly faster for large git logs (10k+ commits) and complex diffs (1000+ lines)
- Lower memory footprint and faster startup time

### Distribution
- **Go**: Single binary (~8-15MB), zero dependencies, trivial Homebrew formula
- **Ink**: Requires Node.js runtime (~60-100MB total footprint)

### Ecosystem
- **Go**: Focused, cohesive tools from Charm Bracelet; all needed components available
- **Ink**: Larger npm ecosystem with more options but variable quality

### AI Assistance
- **TypeScript**: Excellent AI support ("on distribution")
- **Go**: Good AI support; hypothesis that modern AI can compensate for learning curve

## Rationale

### Why Go + Bubbletea

1. **Distribution Excellence**: Single binary with zero dependencies provides best user experience
2. **Performance**: Better handling of large repos and complex diffs
3. **Professional Polish**: Fast startup, no dependency conflicts, aligns with system tool expectations
4. **AI-Assisted Learning**: Betting on AI assistance to overcome Go learning curve
5. **Cohesive Ecosystem**: Charm Bracelet provides integrated, high-quality tooling

### Trade-offs Accepted

- Learning curve for Go and Elm Architecture (mitigated by AI assistance)
- Potentially slower initial development (expected to catch up quickly)
- Smaller ecosystem than npm (all needed components available)

## Consequences

### Positive
- Superior distribution story (single binary, ~8-15MB vs 60-100MB)
- Better performance for large repos and diffs
- Simpler deployment and cross-compilation
- Production-ready, cohesive ecosystem

### Negative
- Learning curve for Go and Bubbletea
- TypeScript would have slightly better AI support
- Development velocity initially uncertain

## Success Criteria

1. AI assistance proves effective for learning Go/Bubbletea
2. Development velocity remains acceptable
3. Distribution via Homebrew is straightforward
4. Performance meets expectations for large repositories

## Alternatives Considered

- **TypeScript + Ink**: Rejected due to Node.js dependency and performance concerns
- **Rust + Ratatui**: Rejected due to steeper learning curve
- **Go + tview**: Rejected in favor of Bubbletea's Elm Architecture for maintainability
