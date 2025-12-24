# Phase 1: Clarify

> **Workflow:** 1. **Clarify** → 2. Design → 3. Implement  
> See [overview](00-overview.md) for details.

Transform a rough idea into clear, validated requirements through structured conversation and research.

## Goal

Establish shared understanding of **what** we're building and **why**. End with documented requirements that both human and AI agree on.

## Inputs

- Rough idea or feature description from user
- Access to codebase for context

## Outputs

Created in `docs/specs/<yymmdd-feature-name>/`:
- `01-requirements.md` - Validated requirements document
- `research/<topic>.md` - Research findings (as needed)

## Process

### 1. Setup

- Create spec folder: `docs/specs/<yymmdd-feature-name>/`
- Capture the rough idea

### 2. Iterative Clarification

Flexibly combine asking questions and doing research based on what's needed:

- If input is too vague → ask for more details first
- If extending existing feature → research current implementation first
- If questions reveal unknowns → research before continuing
- If research raises new questions → ask them

Continue until both parties have sufficient understanding. No artificial limits.

**Ask questions**
- Ask only ONE question per message - wait for the answer before asking the next
- Provide concrete options when possible (e.g., "Should we: A) cache results, B) compute on demand, or C) something else?")
- Let the user pick or provide their own answer

Example topics to clarify:
- What problem are we solving?
- Who is affected?
- What does success look like?
- What are the constraints?
- What is explicitly out of scope?

**Research** (use subagents to keep main context clean)
- Understand existing code/features being extended
- Investigate technical constraints
- Explore patterns, APIs, data models
- Document findings in `research/<topic>.md`

### 3. Document Requirements

Write `01-requirements.md` containing:
- Problem statement
- Goals and non-goals
- User impact
- Key requirements (functional and non-functional)
- Open questions (if any remain for design phase)
- References to research documents

## Completion Gate

- User explicitly approves the requirements
- Both parties have shared understanding
- Ready to proceed to design phase

## Principles

- **What/Why, not How** - Clarify requirements and constraints, not implementation details. Save technical decisions for the design phase.
- **Ask, don't assume** - When uncertain, ask rather than guess. Never invent requirements the user didn't state.
- **One question at a time** - Ask ONE question, wait for answer, then ask the next.
- **No invented numbers** - Never make up specific metrics, thresholds, or performance targets (e.g., "<100ms", "up to 1000 users").
- **Subagents for research** - Keep main conversation context clean.
- **Document as you go** - Research findings go into files, not just chat.
- **User drives scope** - AI proposes, user decides what's in/out.
- **Flexible flow** - Adapt the order of activities to what the situation needs.
