# Phase 2: Design

> **Workflow:** 1. Clarify → 2. **Design** → 3. Implement
> See [overview](00-overview.md) for full context.

## Purpose

Create a design document that lets the developer review and approve the approach before any code is written. Reviewing a design doc is much faster than reviewing a PR, and catching issues here saves significant rework.

**The core question:** "What solution did I choose and why?" — not "How will I build this?"

| Design Is | Design Is Not |
|-----------|---------------|
| Exploring solution space | Describing THE solution in implementation detail |
| Making and documenting decisions | Documenting architecture without rationale |
| Why this approach over alternatives | How the pieces connect at code level |
| Risks and unknowns | Implementation order or file-by-file changes |

**After approval:** Proceed to [Phase 3: Implement](03-implement.md).

## Workflow

The AI works autonomously through these steps:

### 1. Understand Requirements

Read `01-requirements.md` and any research documents from Phase 1. Be clear on what problem we're solving and what success looks like.

### 2. Research the Codebase

Use subagents to explore the current codebase. Each subagent investigates a specific area and documents findings in the `research/` folder.

Why subagents?
- Clean context produces better analysis
- Each research doc is focused and coherent
- Main agent synthesizes without getting lost in details

Example research tasks:
- "How does the current X system work?"
- "What patterns does this codebase use for Y?"
- "What are the integration points for Z?"

### 3. Explore Solution Space

Use subagents to explore different approaches. Each subagent can investigate one potential solution direction and document pros/cons.

The main agent then synthesizes: compare approaches, make decisions, document rationale.

This works better than a single agent iteratively updating a document — fresh context with complete information produces more coherent output.

### 4. Write the Design Document

Create `02-design.md` in the feature folder using the template below.

## Design Document Template

The document should flow naturally — the reader should never be surprised. Every sentence should follow obviously from the previous ones.

### Executive Summary

*Write this section last.*

2-3 paragraphs maximum. A developer should be able to read this and decide whether they need to read the full document or can approve based on the summary.

### Context & Problem Statement

What problem are we solving? Why now?

Include scope boundaries if there's risk of confusion: "This design covers X. It does not cover Y."

### Current State

How things work today. Important for understanding what changes and why.

### Solution

Structure this section to best match the problem. The goal is clear communication, not filling in a template.

**Requirements:**

- **Good flow.** The reader should think "this is entirely straightforward" by the end.
- **Decisions marked clearly.** Use this pattern so reviewers can spot decisions quickly:

  > **Decision:** We chose X over Y because Z. The tradeoff is...

- **Visual representations.** Diagrams for component interactions, data flow, state changes. A picture often communicates better than paragraphs.

**Areas to cover** (as relevant):

- How do the major components interact? (high-level diagram)
- How does data flow through the system? What state changes?
- What data models or interfaces are involved? (types/schemas are fine — they clarify the design)

**Stay at the right level:** Data models and interfaces are design. Implementation details (which files to change, what order to code things, method bodies) belong in Phase 3.

### Open Questions

Decisions that need developer input before implementation can proceed. These are blockers — the developer must answer them as part of approval.

If there are no open questions, state that explicitly.

## Principles

- **Decide, don't defer.** The AI makes decisions and documents rationale. The developer validates, redirects, or approves — but shouldn't have to make every decision.
- **Mark decisions clearly.** The developer reviewing should be able to scan for decision points and evaluate the reasoning.
- **Stay autonomous.** Don't ask the developer for input on every question. Research, decide, document. The review is where feedback happens.
- **No implementation planning.** That's Phase 3. Focus on what the solution is and why, not how to build it.
- **Document stands alone.** The reviewer shouldn't need prior context or conversation history.
