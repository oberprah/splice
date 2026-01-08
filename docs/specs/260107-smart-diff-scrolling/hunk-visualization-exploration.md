# Hunk Visualization Exploration

This document captures the exploration of how to visually indicate hunk correspondence in a side-by-side diff view **without using blank lines**.

## The Problem

With the "no blank lines" approach, each panel displays continuous content. However, this makes it difficult to see which hunks on the left correspond to which hunks on the right, especially when:

- Hunks have different sizes on each side
- Multiple hunks appear consecutively
- Some hunks are pure additions or pure deletions

### Example Diff

**Left (old):**
```
1  package main
2  func old1() {}
3  func old2() {}
4  func old3() {}
5  func old4() {}
6  func middle() {}
7  func another() {}
8  func update() { old }
9  func end() {}
```

**Right (new):**
```
1  package main
2  func middle() {}
3  func inserted() {}
4  func another() {}
5  func update() { new }
6  func extra1() {}
7  func extra2() {}
8  func end() {}
```

**Hunks:**
1. **Hunk 1** (pure removal): Lines 2-5 removed (old1-old4)
2. **Hunk 2** (pure addition): Line 3 added (inserted)
3. **Hunk 3** (modification): Line 8 changed + lines 6-7 added (update modified, extra1/extra2 added)

### Current "No Blank Lines" Rendering (The Problem)

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
  1   package main            │   1   package main
  2 - func old1() {}          │   2   func middle()
  3 - func old2() {}          │   3 + func inserted() {}
  4 - func old3() {}          │   4   func another()
  5 - func old4() {}          │   5 + func update() { new }
  6   func middle()           │   6 + func extra1() {}
  7   func another()          │   7 + func extra2() {}
  8 - func update() { old }   │   8   func end()
  9   func end()              │
```

**Problem:** You can't tell which removals correspond to which additions. The hunks visually merge together.

## IDE Reference (IntelliJ)

IntelliJ uses:
- **Grey background**: Removed lines (left side)
- **Green background**: Added lines (right side)
- **Blue background**: Modified lines (both sides, connected)
- **Colored connector strips** in the gutter between panels showing which hunks correspond

The key insight: The **continuous colored blocks** show what belongs together, and **connector strips** in the gutter visually link corresponding hunks across panels.

**Terminal limitation:** We cannot draw graphical connector strips in the gutter.

## Explored Approaches

### 1. Numbered Markers

Mark the start of each hunk with a number:

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
  1   package main            │   1   package main
① 2 - func old1() {}          │   2   func middle()
  3 - func old2() {}          │ ② 3 + func inserted() {}
  4 - func old3() {}          │   4   func another()
  5 - func old4() {}          │ ③ 5 + func update() { new }
  6   func middle()           │   6 + func extra1() {}
  7   func another()          │   7 + func extra2() {}
③ 8 - func update() { old }   │   8   func end()
  9   func end()              │
```

### 2. Vertical Bars with Hunk Numbers

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
    1   package main          │     1   package main
1 │ 2 - func old1() {}        │     2   func middle()
1 │ 3 - func old2() {}        │ 2 │ 3 + func inserted() {}
1 │ 4 - func old3() {}        │     4   func another()
1 │ 5 - func old4() {}        │ 3 │ 5 + func update() { new }
    6   func middle()         │ 3 │ 6 + func extra1() {}
    7   func another()        │ 3 │ 7 + func extra2() {}
3 │ 8 - func update() { old } │     8   func end()
    9   func end()            │
```

### 3. Bracket-Style Hunk Boundaries

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
    1   package main          │     1   package main
┌   2 - func old1() {}        │     2   func middle()
│   3 - func old2() {}        │ ─   3 + func inserted() {}
│   4 - func old3() {}        │     4   func another()
└   5 - func old4() {}        │ ┌   5 + func update() { new }
    6   func middle()         │ │   6 + func extra1() {}
    7   func another()        │ └   7 + func extra2() {}
─   8 - func update() { old } │     8   func end()
    9   func end()            │
```

### 4. Underscore/Overline Markers on Anchor Lines

Use `▔` (overline) or `▁` (underscore) to mark where content was removed/added:

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
  1   package main            │   1   package main▁▁▁▁▁▁▁▁▁▁
  2 - func old1() {}          │   2   func middle()
  3 - func old2() {}          │   3 + func inserted() {}
  ...
```

The underscore under "package main" on the right indicates "content was removed after this line".

### 5. Diagonal/Block Connectors

Using corner triangles and block characters:

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
  1   package main            │   1   package main
  2 - func old1() {}          ◤
  3 - func old2() {}          █   2   func middle()
  4 - func old3() {}          █
  5 - func old4() {}          ◣
  6   func middle()           │
```

### 6. Git-Log Style Branching

```
Left Panel                    │ Right Panel
──────────────────────────────│──────────────────────────────
  1   package main            │     1   package main
  2 - func old1() {}          ●─┐
  3 - func old2() {}          │ │   2   func middle()
  4 - func old3() {}          │ │ ●─3 + func inserted() {}
  5 - func old4() {}          │ │   4   func another()
  6   func middle()           │ └─●─5 + func update() { new }
  7   func another()          │   │ 6 + func extra1() {}
  8 - func update() { old }   ●───┤ 7 + func extra2() {}
  9   func end()              │     8   func end()
```

## Useful Unicode Characters

### Squares & Blocks
```
█ ▓ ▒ ░       Full → Light shading
■ □ ▪ ▫       Black/white squares
▀ ▄           Upper/lower half
▌ ▐           Left/right half
▔ ▁           Upper/lower eighth (thin lines)
```

### Triangles (Corner/Diagonal Fills)
```
◢   Black lower-right triangle
◣   Black lower-left triangle
◤   Black upper-left triangle
◥   Black upper-right triangle
```

### Diagonal Fills (Symbols for Legacy Computing)
```
🭀 🭁 🭂 🭃 🭄 🭅 🭆 🭇 🭈 🭉 🭊 🭋   Upper-left to lower-right
🭌 🭍 🭎 🭏 🭐 🭑 🭒 🭓 🭔 🭕 🭖 🭗   Lower-left to upper-right
```

### Box Drawing
```
─ │ ┌ ┐ └ ┘   Basic
├ ┤ ┬ ┴ ┼     T-junctions
╭ ╮ ╯ ╰       Rounded corners
```

### Arrows
```
→ ← ↑ ↓       Basic
▶ ◀ ▲ ▼       Filled triangular
```

## Terminal Drawing Capabilities

1. **Character-based**: Unicode blocks, box drawing, braille - works everywhere
2. **ANSI colors**: Background colors on characters - works in most terminals
3. **Braille patterns**: 2x4 dot grid per character for higher resolution
4. **Graphics protocols** (terminal-dependent): Sixel, Kitty, iTerm2

## Open Questions

1. Which approach provides the best balance of clarity and visual appeal?
2. Should we combine approaches (e.g., colored backgrounds + markers)?
3. How do these render across different terminal emulators?
4. What's the performance impact of complex Unicode rendering?

## Next Steps

- Choose an approach to prototype
- Implement and test across terminals
- Gather user feedback on clarity
