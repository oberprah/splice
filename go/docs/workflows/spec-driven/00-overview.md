# Spec-Driven Development Workflow

A structured approach for implementing features with AI assistance. Human stays strategic (requirements, design review), AI handles tactical work (research, implementation).

## Philosophy

- **Documents as communication** - Artifacts over constant back-and-forth
- **Two human checkpoints** - After clarification, after design
- **AI proposes, human validates** - AI makes decisions and documents rationale; human corrects if needed
- **Fresh context per phase** - Each phase can start with new conversation, reading artifacts
- **Research is continuous** - Can happen in any phase, always documented

## Phases

| Phase | Purpose | Who Leads | Output | Gate |
|-------|---------|-----------|--------|------|
| 1. Clarify | Define what and why | AI (with user Q&A) | `01_requirements_<feature-name>.md` | User approves understanding |
| 2. Design | Define how | AI | `02_design_<feature-name>.md` | User approves approach |
| 3. Implement | Build it | AI (coordinator + subagents) | Code + Tests + `03_implementation_<feature-name>.md` | Human tests, then creates PR |

See detailed instructions:
- [01-clarify.md](01-clarify.md) - Phase 1: Requirements clarification
- [02-design.md](02-design.md) - Phase 2: Design document creation
- [03-implement.md](03-implement.md) - Phase 3: Implementation

## Folder Structure

Each feature gets a folder: `docs/specs/<yymmdd-feature-name>/`

```
docs/specs/241219-user-export/
├── 01_requirements_user-export.md   # What we're building and why
├── 02_design_user-export.md         # How we're building it (standalone)
├── 03_implementation_user-export.md # Plan, progress, discoveries
└── research/                        # Research documents (any phase)
    ├── existing-api.md
    └── auth-patterns.md
```

## When to Use

- New features requiring design decisions
- Complex changes touching multiple components
- Work that benefits from documented rationale
