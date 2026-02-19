# Phase 3: Implement

> **Workflow:** 1. Clarify → 2. Design → 3. **Implement**
> See [overview](00-overview.md) for details.

## Purpose

This phase executes the approved design through coordinated, autonomous work. The main agent acts as a **coordinator**, delegating implementation steps to subagents who implement, test, commit, and report back.

Why delegate?
- **Clean context** — Main agent coordinates; subagents handle implementation details
- **Autonomous execution** — Subagents are equally capable; they decide implementation within structural boundaries
- **Recoverable failures** — Issues are isolated to individual steps

**After completion:** Human tests the implementation, then creates PR.

## The Work

### 1. Prepare

- Read `02_design_<feature-name>.md` — understand the structural decisions
- Review `01_requirements_<feature-name>.md` — note acceptance criteria
- Review research files from `research/` folder if they exist — prior exploration may answer your questions
- **Explore the codebase as needed** — Use subagents to find specific files, understand current patterns not covered in prior research, or verify information is still current. The design doc tells you *what*; exploration tells you *where* and *how things currently work*.
- Identify component boundaries, interfaces, and key files

### 2. Create Implementation Plan

Write `03_implementation_<feature-name>.md` with steps. Each step is a **logical unit** a subagent executes independently.

**What makes a good step:**

- **Structural clarity** — Which components are involved? New modules to create? Interfaces to define?
- **Clear verification** — How does the subagent prove success? Be specific: "unit tests for X logic", "e2e test for Y flow"—not just "tests pass"
- **Tests included** — New logic includes its tests; never separate "write tests" steps
- **Right level of detail** — Describe *what* and *where*, not *how*. The subagent decides concrete implementation.
- **Self-contained** — App compiles and tests pass after the step

**The coordinator decides structure:** Which component owns this logic? Should we extract a module? What are the interfaces? The subagent decides implementation within those boundaries.

**Self-review:** Would an equally capable developer know the goal, which components to modify, and how to verify success? Too detailed = specifying every line. Too vague = no interfaces or component boundaries.

### 3. Execute Steps

For each step, sequentially — complete one before starting the next:

#### Delegate

Pass to the subagent:
- **Goal** — What to accomplish
- **Structure** — Components, interfaces, boundaries
- **Verification** — Specific tests or checks that prove success
- **Files to read** — Design doc and key source files
- **Where to report** — The step section in `03_implementation_<feature-name>.md`

#### Subagent Work

Subagents work autonomously: read context, implement within structural boundaries, commit, run verification, document progress in the implementation file.

**Default approach:** Write tests first. This clarifies the interface and catches misunderstandings before implementation. Deviate only when testing after makes more sense (e.g., exploratory spikes, UI layout).

If verification fails, iterate. If still failing after reasonable effort, report as blocker.

#### Coordinator Review

After each step, do a **structural review**:

1. Read subagent's notes — what was implemented, any deviations?
2. Check structural alignment — interfaces match design? Components in right places?
3. Skim key changes — new interfaces, public APIs. Don't review every line.

Then decide:
- **Aligned** → Next step
- **Minor deviation** → Note it, proceed
- **Structural issue** → Adjust plan or have subagent revise
- **Blocker** → Stop, document, report to developer

### 4. Complete

**Final verification:**
- Run full test suite
- Check each requirement in `01_requirements_<feature-name>.md`
- Check structural decisions in `02_design_<feature-name>.md` were followed

**Prepare for human testing** (do NOT create PR):
- All changes committed
- Update `03_implementation_<feature-name>.md`
- Report completion to developer

## Implementation Document Template

```markdown
# Implementation: <feature-name>

**Requirements:** `01_requirements_<feature-name>.md`
**Design:** `02_design_<feature-name>.md`

## Steps

### Step 1: <description>
**Goal:** ...
**Structure:** Components X, Y. New interface Z.
**Verify:** Unit tests for X, integration test for X-Y.
**Read:** `src/services/X.ts`, `src/types/Y.ts`

**Status:** Complete
**Commits:** abc123, def456
**Verification:** Passed — unit tests added, integration test green
**Notes:** Implemented X in `src/services/`. Decision: used adapter pattern for Y because...
**Coordinator Review:** Structure matches design. → Step 2

### Step 2: <description>
**Goal:** ...
**Structure:** Component Z, extends interface from Step 1.
**Verify:** E2E test for user flow.
**Read:** `src/components/Z.tsx`, `src/hooks/useZ.ts`

**Status:** In Progress
**Notes:** ...

## Final Verification

- [ ] Full test suite passes
- [ ] All requirements verified
- [ ] Design decisions followed

## Summary

- What was built
- Deviations from design (with rationale)
```

## Principles

- **Coordinator, not implementer** — Main agent orchestrates; subagents execute
- **Context and verification** — Every delegation includes what to build and how to verify
- **Structural guidance** — Coordinator decides boundaries and interfaces; subagents decide implementation
- **Review after each step** — Catch structural drift early
- **Tests with code** — Same step, not separate
- **Document deviations** — With rationale
- **Stop on blockers** — Don't work around fundamental issues
