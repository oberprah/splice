# Phase 3: Implement

> **Workflow:** 1. Clarify → 2. Design → 3. **Implement**
> See [overview](00-overview.md) for details.

## Purpose

This phase executes the approved design through coordinated, autonomous work. The main agent acts as a **coordinator**, delegating high-level implementation steps to subagents who implement, test, commit, and report back.

Why delegate to subagents?
- **Clean context** — Main agent stays focused on the big picture, subagents handle implementation details
- **Smart execution** — Subagents are equally capable; they receive full design context and make decisions autonomously
- **Logical units** — Each step is a coherent piece of work with multiple commits if needed
- **Fresh perspective** — Each subagent approaches their step without accumulated mental clutter
- **Recoverable failures** — Issues are isolated to individual steps

**After completion:** Human tests the implementation, then creates PR.

## The Work

The coordinator works through these stages:

### 1. Prepare

- Read `02-design.md` and understand the implementation approach
- Review `01-requirements.md` for verification criteria
- Read relevant `CLAUDE.md` files for affected components
- Note the components to modify and testing strategy

### 2. Create Implementation Plan

Write `03-implementation.md` with a checklist of high-level steps. Each step should be a **logical unit** that a smart, autonomous subagent can execute independently.

**Step characteristics:**
- **Logical scope**: Represents a coherent piece of work
- **Tests included**: New logic includes its tests—don't create separate "write tests" steps
- **Not too detailed**: Subagents are equally capable—they don't need line-by-line instructions
- **Self-contained**: Each step leaves the app compiling and tests passing
- **Can contain multiple commits**: A step can include several related commits
- **Clear outcome**: Success is obvious (tests pass, feature works)

Include **validation steps** at strategic points—not after every step, but at natural checkpoints and always at the end. Validation steps use the running application (e.g., Playwright MCP for UI verification).

**Self-review the plan**: Would an equally capable developer understand the goal and know how to execute it? If you're specifying every file and every line, you're going too deep.

### 3. Execute Steps

For each step in the plan:

**Delegate to subagent** with:
- The full design document (`02-design.md`) — always, for complete context
- The step to accomplish (high-level, not detailed)
- Relevant `CLAUDE.md` files for affected components

Subagents are equally capable—they figure out implementation details, decide what files to change, make commits as needed, and ensure tests pass before reporting back with what was done, key decisions made, and any issues encountered.

**Evaluate the report** and decide next action:
- **Success** → Proceed to next step
- **Minor deviation** → Note in `03-implementation.md`, proceed if result is acceptable
- **Plan adjustment** → Update `03-implementation.md` with new approach and rationale, continue
- **Blocking issue** → Stop immediately, document blocker, report to developer (may need to return to design phase)

### 4. Final Verification

Before declaring complete:
- Run all checks: `./dev backend check`, `./dev frontend check`
- Run all tests: `./dev backend test`, `./dev frontend e2e`
- Review `01-requirements.md` — verify each requirement is addressed
- Run validation step: start application, verify key flows work
- Document verification results in `03-implementation.md`

### 5. Prepare for Human Testing

Do NOT create a PR. Instead:
- Ensure all changes are committed
- Update `03-implementation.md` with:
  - Summary of what was implemented
  - Any deviations from the design
  - Known issues or limitations
  - Instructions for testing
- Report completion to developer for manual testing

The developer will test, then create the PR when satisfied.

## The Implementation Document

Create `03-implementation.md` in the feature folder. The number of steps depends on feature size: small features may have 1-2 steps, larger features may have 5+. Each step is a logical unit.

### Structure

```markdown
# Implementation Plan

## Steps

- [ ] Step 1: High-level description of logical unit
- [ ] Step 2: ...
- [ ] Validation: Test key user flows with running app

## Progress

### Step 1: Description
Status: ✅ Complete
Commits: abc123, def456
Notes: Subagent implemented X and Y, made design decision Z...

### Step 2: Description
Status: 🔄 In progress
Notes: ...

## Discoveries

- Found that X works differently than expected...
- Changed approach for Y because...

## Verification

- [ ] All tests pass
- [ ] Requirements verified
- [ ] Manual validation complete
```

## Completion Gate

- All implementation steps completed
- All tests pass
- Requirements verified against `01-requirements.md`
- Final validation passed
- `03-implementation.md` documents any deviations
- Branch is ready for human testing (no PR yet)

## Principles

- **Coordinator, not implementer** — Main agent orchestrates; subagents execute
- **High-level steps** — Logical units, not detailed instructions; trust subagent intelligence
- **Always pass design doc** — Subagents need full context to make good decisions
- **Multiple commits per step** — Each step can include several commits; tests pass after each step
- **Clean main context** — Subagents handle details so coordinator stays focused on big picture
- **Sequential execution** — One step at a time; later steps may change based on earlier results
- **Tests with code** — Tests belong in the same step as the code they verify, not as separate steps
- **Document deviations** — Note anything different from the design
- **Stop on blockers** — Don't guess or work around fundamental issues
- **Human tests first** — No PR until developer verifies the implementation
