# Phase 2: Design

> **Workflow:** 1. Clarify → 2. **Design** → 3. Implement
> See [overview](00-overview.md) for details.

## Purpose

This phase is a **checkpoint**. The AI works autonomously to propose a solution, then the developer reviews and approves before any code is written.

Why a design doc instead of jumping to code?
- Reviewing a design doc is much easier than reviewing a PR.
- Decisions and tradeoffs are documented so the developer can evaluate the approach.
- Catching issues here saves significant rework later.

**How it works:** The AI proposes, the developer validates. The AI makes decisions and documents rationale. The developer approves, redirects, or rejects. This keeps momentum while maintaining oversight.

**After approval:** Proceed to [Phase 3: Implement](03-implement.md).

## The Work

The AI works autonomously through these steps:

### Understand

- Read `01-requirements.md` and any research documents from phase 1.
- Research the codebase: existing patterns, integration points, constraints.

Do the homework. A great solution requires fully understanding the problem and codebase. Use subagents for research to keep context clean, and document findings in `research/` folder.

### Design

Start from the user-facing layer and work down:
- What does the user see/do?
- What API serves that?
- What services/logic support that API?
- What data structures are needed?

For significant decisions: consider alternatives, choose the best approach, document why.

### Validate

Use a subagent to review the design. A fresh perspective catches issues that accumulated context blinds you to.

The subagent should evaluate:

| Test | Question |
|------|----------|
| **Fitness Test** | Is this the right solution for the problem? Would a simpler approach work? Are we working with the codebase or fighting it? |
| **Vacation Test** | Could a teammate understand and evaluate this approach without contacting you? |
| **Skeptic Test** | Have you addressed likely objections? |
| **Scope Test** | Does this solve the requirements without gold-plating? |

If validation reveals requirement gaps, return to [Phase 1: Clarify](01-clarify.md) before proceeding.

## The Design Document

Create `02-design.md` in the feature folder.

### Level of Detail

Stay at the architectural level:

| Design (this phase) | Implementation (next phase) |
|---------------------|----------------------------|
| WHAT components interact | WHERE in the codebase |
| HOW data flows | WHAT order to code |
| WHY this approach | WHICH files and lines |
| Interface contracts | Method signatures |

Design docs should remain valid even if code is refactored. This means **no implementation code**—if you're writing code blocks, you've gone too deep. The exception: brief snippets for critical interfaces or non-obvious algorithms where prose would be less clear.

### What to Cover

**Context**
- What problem are we solving? Why now?
- Goals: what success looks like
- Non-goals: what we're explicitly not doing

**Current State**
- How things work today (if relevant to understanding the change)

**Solution**
The proposed design. Structure flexibly based on the problem, but cover these areas as applicable:
- Architecture overview
- Components and interfaces
- Data flow
- Data models
- Error handling

Document decisions and tradeoffs inline as they arise—why this approach over alternatives.

**Verification**
- How we'll test it works

**Open Questions** (if any)
- Decisions that need human input before implementation

### Writing for Reviewability

- Good flow: if the reader is surprised by a conclusion, you skipped a step.
- Diagrams for anything with multiple components interacting.
- Verified metrics only, or explicitly marked as estimates.

The document must be standalone. The reviewer shouldn't need prior context.
