# ANSI + Lip Gloss Integration Research

## Summary

Lip Gloss is ANSI-aware and can preserve existing ANSI codes when applying new styles. This makes combining Chroma's syntax highlighting (foreground colors) with Lip Gloss backgrounds (for diff line changes) straightforward.

## Key Findings

### 1. Lip Gloss Preserves ANSI Codes

Lip Gloss is built on:
- **Termenv** - for advanced ANSI style support
- **Reflow** - for ANSI-aware text operations (wrapping, padding)

The `Style.Render()` method preserves existing ANSI escape codes while applying new styles. Functions like `Width()` and `Height()` correctly measure text by ignoring ANSI sequences.

### 2. ANSI Foreground + Background Work Independently

ANSI escape codes support independent foreground and background colors:
- Syntax highlighting provides: foreground colors (keywords blue, strings green, etc.)
- Lip Gloss can add: background colors (green bg for additions, red bg for deletions)

These combine in terminal output - foreground colors display on top of background colors. No "blending" occurs; each attribute is independent.

### 3. Charm Ecosystem Example: Glamour

The Glamour markdown renderer (Charm ecosystem) demonstrates this pattern:
- Uses Chroma for syntax highlighting in code blocks
- Applies margins, padding, and styling through Lip Gloss/Reflow
- Handles color profiles through Termenv (NoTTY, ANSI, ANSI256, TrueColor)

See: github.com/charmbracelet/glamour/blob/master/ansi/codeblock.go

### 4. Practical Integration Approach

**Direct Composition (Simplest):**

1. Use Chroma to syntax-highlight code → ANSI string with foreground colors
2. Pass that string to Lip Gloss style with Background() → ANSI string with both
3. Render in terminal → syntax colors displayed over background color

```
chromaOutput := "[ANSI fg codes]highlighted code[reset]"
lipglossOutput := DiffAdditionsStyle.Render(chromaOutput)
// Result: syntax foreground colors + green background
```

### 5. Technical Constraints

**ANSI Color Priority:**
- Multiple foreground colors: last one wins
- Multiple background colors: last one wins
- Foreground + background: both displayed (independent)

**Terminal Capabilities:**
- Lip Gloss auto-detects color profile (TrueColor/ANSI256/ANSI/NoColor)
- Colors outside gamut automatically degraded to closest available
- Use same color profile for both Chroma and Lip Gloss for consistency

**Width Calculations:**
- `lipgloss.JoinHorizontal()` correctly handles ANSI-laden strings
- Reflow's padding/indentation works with pre-styled text
- Text measurements ignore ANSI codes in length calculations

## Decision

> **Decision:** Use direct composition - apply Chroma syntax highlighting first (foreground colors), then wrap with Lip Gloss for backgrounds. This works because ANSI foreground/background are independent attributes, and Lip Gloss preserves existing ANSI codes.
