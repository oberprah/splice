# Design: Command Help Hints

**Date:** 2026-01-12
**Status:** Awaiting approval

## Executive Summary

This design adds a context-sensitive help overlay to Splice, allowing users to press `?` to see available commands for the current screen. The solution uses a pure component function (`HelpOverlay`) that renders a centered modal box, with each state maintaining a simple `showHelp` boolean flag. This approach follows Splice's existing component patterns exactly, requiring minimal code per state (~6 lines) while keeping all help rendering logic in one reusable function.

The design prioritizes architectural consistency and simplicity over eliminating all duplication. For 3 states with straightforward help overlays, explicit integration is clearer and more maintainable than wrapper patterns or complex abstractions.

Key aspects:
- New `components.HelpOverlay()` pure function for rendering
- New `ViewBuilder.AddOverlay()` method for composing overlays over content
- Each state adds: `showHelp bool` field, `?` key handler, conditional overlay rendering
- Help content defined as static command lists per state
- Text-based box using Unicode box-drawing characters, consistent with Splice's minimal aesthetic

## Context & Problem Statement

Splice currently has no built-in help system. Users must discover keyboard shortcuts through trial and error, external documentation, or reading code. With Vim-style shortcuts across LogState (10+ commands), FilesState (7+ commands), and DiffState (10+ commands), the lack of in-app help creates a poor first-time user experience.

The requirements specify:
- Toggle overlay with `?` key
- Context-sensitive commands for each screen
- Centered modal/box display over current view
- Dismissible with `?`, `q`, or `esc`
- Available only in main navigation screens (not LoadingState or ErrorState)
- Organic discovery (no persistent hints)

**Design constraint:** Must be a generic, reusable solution—not separate implementations per screen.

## Current State

### No Help System

Splice has **no existing help, modals, or overlay components**. Current transient states (LoadingState, ErrorState) show minimal text with no decoration.

### State Architecture

Each screen implements `core.State` interface:
```go
type State interface {
    View(ctx Context) ViewRenderer
    Update(msg tea.Msg, ctx Context) (State, tea.Cmd)
}
```

States are organized in a navigation stack. Key handling is per-state via direct string matching:
```go
case tea.KeyMsg:
    switch msg.String() {
    case "q": // Back
    case "j", "down": // Navigate
    }
```

### Component Patterns

Existing components are pure functions:
- `CommitInfo(commit, width, bodyMaxLines, ctx) []string`
- `FileSection(files, width, cursor) []string`
- `FormatCommitLine(components, width) string`

Components use `ViewBuilder` for rendering:
```go
vb := components.NewViewBuilder()
vb.AddLine(styledLine)
return vb
```

### Styling

Splice uses minimal Lip Gloss styling:
- Color-only (Foreground/Background with AdaptiveColor)
- No borders, padding, or margins via Lip Gloss
- Box-drawing characters for structure (`─`, `│`, `┌`, `┐`, `└`, `┘`)
- Manual layout calculations

All styles defined in `internal/ui/styles/styles.go`.

## Solution

### Architecture Decision

> **Decision:** Use a pure component function (`components.HelpOverlay`) with explicit per-state integration, following Splice's existing component patterns.

**Rationale:** After analyzing 4 approaches (see `research/solution-approaches.md`):
1. **State wrapper**: Too complex, foreign pattern, breaks state identity
2. **Embedded helper**: Halfway abstraction without enough benefit
3. **Help screen state**: Wrong abstraction—help is a modal, not a navigation screen
4. **Pure component**: Matches existing patterns exactly, simple and maintainable

The pure component approach trades ~6 lines of duplication per state for significant simplicity and architectural consistency. For 3 states, this is the right tradeoff.

**Benefits:**
- Zero new architectural concepts
- Obvious behavior when reading state code
- Easy to test (pure function)
- Enables future per-state customization
- Matches `CommitInfo`, `FileSection` pattern exactly

**Tradeoff:** Each state explicitly integrates help (field + key handler + view call) rather than magic injection. This is acceptable—explicit is better than implicit for this use case.

### Component Structure

#### 1. Command Data Model

```go
// internal/ui/components/help_overlay.go

type Command struct {
    Keys        []string  // e.g., ["j", "down"]
    Description string    // Brief description
}
```

Each state defines its command list as a package variable:

```go
// internal/ui/states/log/commands.go

var helpCommands = []components.Command{
    {Keys: []string{"j", "down"}, Description: "Move cursor down"},
    {Keys: []string{"k", "up"}, Description: "Move cursor up"},
    {Keys: []string{"g"}, Description: "Jump to first commit"},
    {Keys: []string{"G"}, Description: "Jump to last commit"},
    {Keys: []string{"v"}, Description: "Toggle visual mode"},
    {Keys: []string{"esc"}, Description: "Exit visual mode"},
    {Keys: []string{"enter"}, Description: "Load files for commit(s)"},
    {Keys: []string{"?", "q", "esc"}, Description: "Close help"},
    {Keys: []string{"q"}, Description: "Exit visual mode or quit"},
    {Keys: []string{"ctrl+c"}, Description: "Quit application"},
}
```

#### 2. Help Overlay Component

Pure rendering function:

```go
// internal/ui/components/help_overlay.go

// HelpOverlay renders a centered modal box with command list
func HelpOverlay(commands []Command, width, height int) string {
    // Calculate dimensions
    maxKeyWidth := calculateMaxKeyWidth(commands)
    contentWidth := maxKeyWidth + 3 + maxDescriptionWidth(commands)
    boxWidth := min(contentWidth+4, width-4) // +4 for borders and padding

    // Build box content
    lines := []string{}
    lines = append(lines, buildTitle("COMMANDS"))
    lines = append(lines, "")
    for _, cmd := range commands {
        lines = append(lines, formatCommandLine(cmd, boxWidth-4, maxKeyWidth))
    }

    // Wrap in box with borders
    boxed := wrapInBox(lines, boxWidth)

    // Center vertically and horizontally
    centered := centerBox(boxed, width, height)

    return centered
}

func formatCommandLine(cmd Command, width, keyWidth int) string {
    keyText := strings.Join(cmd.Keys, " / ")
    keyStyled := styles.HelpKeyStyle.Render(keyText)

    // Pad key section to align descriptions
    padding := strings.Repeat(" ", keyWidth-len(keyText))

    descStyled := styles.HelpDescriptionStyle.Render(cmd.Description)

    return fmt.Sprintf("  %s%s - %s", keyStyled, padding, descStyled)
}

func wrapInBox(lines []string, width int) []string {
    bordered := []string{}
    bordered = append(bordered, buildBoxTop(width))
    for _, line := range lines {
        bordered = append(bordered, "│ "+line+strings.Repeat(" ", width-2-runewidth.StringWidth(line))+" │")
    }
    bordered = append(bordered, buildBoxBottom(width))
    return bordered
}

func buildBoxTop(width int) string {
    return "┌" + strings.Repeat("─", width-2) + "┐"
}

func buildBoxBottom(width int) string {
    return "└" + strings.Repeat("─", width-2) + "┘"
}

func centerBox(boxLines []string, termWidth, termHeight int) string {
    boxHeight := len(boxLines)

    // Calculate centering
    topPadding := (termHeight - boxHeight) / 2
    if topPadding < 0 {
        topPadding = 0
    }

    // Center each line horizontally
    centered := []string{}
    for i := 0; i < topPadding; i++ {
        centered = append(centered, "")
    }

    for _, line := range boxLines {
        lineWidth := runewidth.StringWidth(line)
        leftPadding := (termWidth - lineWidth) / 2
        if leftPadding < 0 {
            leftPadding = 0
        }
        centered = append(centered, strings.Repeat(" ", leftPadding)+line)
    }

    return strings.Join(centered, "\n")
}
```

#### 3. ViewBuilder Extension

Add overlay composition method:

```go
// internal/ui/components/viewbuilder.go

// AddOverlay renders overlay content on top of existing content
func (vb *ViewBuilder) AddOverlay(overlay string) {
    overlayLines := strings.Split(overlay, "\n")

    // Replace non-empty overlay lines over content lines
    // Empty overlay lines are transparent (show content below)
    for i, overlayLine := range overlayLines {
        if i < len(vb.lines) {
            if strings.TrimSpace(overlayLine) != "" {
                vb.lines[i] = overlayLine
            }
        } else {
            vb.lines = append(vb.lines, overlayLine)
        }
    }
}
```

#### 4. Styling Additions

Add to `internal/ui/styles/styles.go`:

```go
var (
    // Help overlay styles
    HelpKeyStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "172", Dark: "220"}).
        Bold(true)

    HelpDescriptionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "252"})

    HelpTitleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "236", Dark: "231"}).
        Bold(true)

    HelpBorderStyle = lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "243", Dark: "248"})
)
```

Colors align with existing palette:
- Keys: Yellow/amber like commit hashes
- Descriptions: Subtle gray like timestamps
- Title: White/high contrast
- Border: Muted gray like header separators

### State Integration Pattern

Each state (LogState, FilesState, DiffState) integrates help with 3 changes:

#### 1. Add Field

```go
// In internal/ui/states/log/state.go
type State struct {
    Commits       []core.GitCommit
    Cursor        core.CursorState
    ViewportStart int
    Preview       PreviewState
    GraphLayout   *graph.Layout
    showHelp      bool  // New
}
```

#### 2. Handle Keys

```go
// In internal/ui/states/log/update.go
case tea.KeyMsg:
    switch msg.String() {
    case "?":
        s.showHelp = !s.showHelp
        return s, nil
    case "q", "esc":
        if s.showHelp {
            s.showHelp = false
            return s, nil
        }
        // Existing quit/exit logic
    }

    // Ignore other keys when help visible
    if s.showHelp {
        return s, nil
    }

    // Existing key handlers...
```

#### 3. Render Overlay

```go
// In internal/ui/states/log/view.go
func (s State) View(ctx core.Context) core.ViewRenderer {
    vb := components.NewViewBuilder()

    // Normal rendering logic...
    // (commit list, split view, etc.)

    // Overlay help if visible
    if s.showHelp {
        overlay := components.HelpOverlay(helpCommands, ctx.Width(), ctx.Height())
        vb.AddOverlay(overlay)
    }

    return vb
}
```

**Per-state cost:** ~6 lines (1 field, 2 key cases, 3 view lines)

### Visual Design

**Box Style:**
```
┌─────────────────────────────────────┐
│  COMMANDS                           │
│                                     │
│  j / down  - Move cursor down      │
│  k / up    - Move cursor up        │
│  g         - Jump to first commit  │
│  G         - Jump to last commit   │
│  v         - Toggle visual mode    │
│  esc       - Exit visual mode      │
│  enter     - Load files            │
│  ? / q / esc - Close help          │
│  q         - Exit or quit          │
│  ctrl+c    - Quit application      │
└─────────────────────────────────────┘
```

**Layout:**
- Centered horizontally and vertically
- Max width: 60 characters (readable, doesn't dominate screen)
- Max height: Terminal height - 4 (leaves space at edges)
- Scrolling: Not needed (command lists fit in typical terminal)
- Background: Transparent (empty lines show content below)
- Border: Box-drawing characters styled with `HelpBorderStyle`

**Responsive Behavior:**
- Terminal < 60 chars: Box shrinks, descriptions may wrap or truncate
- Terminal < 40 chars: Minimal layout (keys only)
- Narrow mode details handled in `HelpOverlay()` function

### Behavior Specification

**Help Toggle:**
- Press `?` → `showHelp = !showHelp`
- Help visible → all keys except `?`, `q`, `esc` ignored

**Help Dismiss:**
- Press `?` → Toggle off (if visible)
- Press `q` → Close help (if visible), else execute normal `q` behavior
- Press `esc` → Close help (if visible), else execute normal `esc` behavior

**State Preservation:**
- Toggling help does not change state (cursor, viewport, selections preserved)
- Returning from help shows exactly the same view as before

### Testing Strategy

**Component Tests:**
- Unit test `HelpOverlay()` with various terminal sizes
- Golden file test for visual output
- Test key alignment, border rendering, centering

**Integration Tests:**
- Test `?` key toggle in each state
- Test dismiss keys (`q`, `esc`) when help visible
- Test that other keys ignored when help visible
- Test state preservation (cursor position unchanged)

**Golden Files:**
- `help_overlay_log_80x24.golden`
- `help_overlay_files_80x24.golden`
- `help_overlay_diff_80x24.golden`
- `help_overlay_narrow_40x24.golden`

### Implementation Checklist

1. **Component implementation:**
   - [ ] `internal/ui/components/help_overlay.go` - Pure function
   - [ ] `Command` struct definition
   - [ ] Box rendering helpers
   - [ ] Centering logic
   - [ ] Component unit tests

2. **ViewBuilder extension:**
   - [ ] `AddOverlay()` method
   - [ ] Overlay composition logic
   - [ ] Tests for overlay behavior

3. **Styling:**
   - [ ] Add help styles to `styles.go`
   - [ ] Verify color palette consistency

4. **LogState integration:**
   - [ ] Define `helpCommands` list
   - [ ] Add `showHelp` field
   - [ ] Handle `?`, `q`, `esc` keys
   - [ ] Render overlay conditionally
   - [ ] Golden file test

5. **FilesState integration:**
   - [ ] Define `helpCommands` list
   - [ ] Add `showHelp` field
   - [ ] Handle `?`, `q`, `esc` keys
   - [ ] Render overlay conditionally
   - [ ] Golden file test

6. **DiffState integration:**
   - [ ] Define `helpCommands` list
   - [ ] Add `showHelp` field
   - [ ] Handle `?`, `q`, `esc` keys
   - [ ] Render overlay conditionally
   - [ ] Golden file test

7. **Documentation:**
   - [ ] Update README to mention `?` for help
   - [ ] Add implementation notes to CLAUDE.md if pattern should be followed

## Open Questions

None—all design decisions are made and ready for implementation.

## References

- **Requirements:** [01_requirements_command-help-hints.md](01_requirements_command-help-hints.md)
- **Research:**
  - [Current Commands](research/current-commands.md)
  - [State Architecture](research/state-architecture.md)
  - [UI Components](research/ui-components.md)
  - [Styling Patterns](research/styling-patterns.md)
  - [Solution Approaches Analysis](research/solution-approaches.md)

## Success Criteria

- [ ] User can press `?` in LogState, FilesState, or DiffState to see help
- [ ] Help overlay shows context-appropriate commands for each screen
- [ ] Help dismisses with `?`, `q`, or `esc`
- [ ] Other keys ignored when help visible
- [ ] State preserved when toggling help (cursor, viewport unchanged)
- [ ] Visual styling matches Splice aesthetic (minimal, text-based)
- [ ] All tests pass (unit, integration, golden files)
- [ ] No regression in existing functionality
