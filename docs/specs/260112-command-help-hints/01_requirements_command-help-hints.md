# Requirements: Command Help Hints

**Date:** 2026-01-12
**Status:** Awaiting approval

## Problem Statement

Splice currently has no built-in help system. Users must discover keyboard shortcuts and commands through trial and error, external documentation, or reading the code. This creates a poor first-time user experience and makes the application less accessible.

As documented in the research findings (`research/current-commands.md`), Splice has a rich set of Vim-style keyboard shortcuts across three main screens (LogState, FilesState, DiffState), but there is no way for users to discover these commands within the application itself.

## Goals

- Provide in-app help that shows available keyboard shortcuts and commands
- Make the help system easily discoverable through a common convention (`?` key)
- Keep the help system unobtrusive and context-aware
- Maintain Splice's lean, keyboard-driven interface philosophy

## Non-Goals

- Persistent UI hints or footer bars (help is on-demand only)
- Interactive tutorials or walkthroughs
- Help for LoadingState or ErrorState (these are transient/edge cases)
- Configuration or customization of key bindings

## User Impact

**Who benefits:**
- New users discovering Splice for the first time
- Existing users who forget less-common shortcuts
- Users familiar with other terminal apps (who expect `?` to show help)

**How they benefit:**
- Faster onboarding and command discovery
- Less context-switching to external documentation
- More confidence in exploring the application's features

## Functional Requirements

### FR1: Help Toggle Key
- Pressing `?` in any main navigation screen (LogState, FilesState, DiffState) toggles the help overlay
- The help key should work in both normal and visual modes (where applicable)

### FR2: Help Overlay Display
- Help appears as a centered modal/box overlay on top of the current view
- The overlay should be clearly distinguishable from the background content
- The overlay should be appropriately sized to fit the terminal but not obstruct critical information unnecessarily

### FR3: Context-Sensitive Content
- Help overlay shows only commands relevant to the current screen:
  - **LogState**: Log view commands (navigation, visual mode, file loading, quit)
  - **FilesState**: File list commands (navigation, diff loading, back, quit)
  - **DiffState**: Diff view commands (scrolling, navigation, change jumping, back, quit)
- Each command entry shows the key(s) and a concise description
  - Format: `key / alternate-key` - Brief description

### FR4: Dismissing the Overlay
- Multiple keys can close the help overlay:
  - `?` (toggle off)
  - `q` (quit/close)
  - `esc` (escape)
- Closing the help returns to the exact state before it was opened

### FR5: Screen Availability
- Help overlay is available only in main navigation screens:
  - LogState ✓
  - FilesState ✓
  - DiffState ✓
  - LoadingState ✗ (transient, no user interaction)
  - ErrorState ✗ (already shows error info, minimal commands)

### FR6: Organic Discovery
- No persistent hints or "Press ? for help" messages in the UI
- Users discover the help feature organically or through external documentation (README, etc.)

## Non-Functional Requirements

### NFR1: Performance
- Help overlay should appear instantly (no async loading)
- Toggling help should not interrupt or reset the current state

### NFR2: Visual Consistency
- Help overlay styling should match Splice's existing aesthetic (Lip Gloss styles)
- Should follow the terminal-native, clean design philosophy

### NFR3: Maintainability
- Command descriptions should be easy to update as features change
- Design should accommodate future addition of new commands without major refactoring

## Open Questions for Design Phase

None - all requirements are clear and validated.

## References

- Research: [Current Commands and Shortcuts](research/current-commands.md)
- Architecture: [State Machine Architecture](../../../CLAUDE.md#state-machine-architecture)
- Codebase patterns: Each state implements `core.State` with `Update()` and `View()` methods

## Success Criteria

- Users can press `?` in LogState, FilesState, or DiffState to view context-appropriate help
- Help overlay displays all available commands for the current screen
- Help can be dismissed with `?`, `q`, or `esc`
- Help overlay is visually clear and doesn't permanently obstruct important content
- Implementation follows Splice's architectural patterns (state-based, pure functions, minimal comments)
