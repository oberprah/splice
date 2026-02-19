# ANSI-Aware Text Operations Research

## Summary

The Charm ecosystem provides comprehensive ANSI-aware text utilities. For per-file syntax highlighting with line splitting and truncation, use `charmbracelet/x/ansi` alongside Chroma and Lip Gloss.

## Key Libraries

### charmbracelet/x/ansi (Recommended)

Modern, comprehensive ANSI utilities:
- `StringWidth()` - Calculate visible width ignoring escape sequences
- `Truncate(s, maxWidth, tail)` - Truncate to N visible chars, preserving ANSI codes
- `Cut(s, start, end)` - Extract substring range while preserving codes
- `Wrap()` / `Wordwrap()` - Line wrapping with ANSI awareness
- `Strip()` - Remove all ANSI codes

### muesli/reflow

ANSI-aware text transformation library:
- `truncate.String(s, width)` - Truncate preserving ANSI
- `wrap.String(s, width)` - Line wrapping
- `ansi` package - Low-level ANSI sequence handling

### charmbracelet/lipgloss (Already Used)

- `Width()`, `Height()` - Measure text accounting for ANSI codes
- `JoinHorizontal()` - Combines strings properly
- Preserves existing ANSI codes when applying new styles

## Chroma Output Behavior

Chroma's terminal formatters:
- Output ANSI escape sequences per token
- Example: `\033[38;5;82mfunc\033[0m \033[38;5;220mMain\033[0m`
- Include reset codes (`\033[0m`) between tokens
- Preserve newlines in output

**Key finding:** When highlighting multi-line code, each token gets its own escape codes. Splitting by `\n` after highlighting works correctly.

## Operations Needed

### 1. ANSI-Aware Width Calculation

```go
import "github.com/charmbracelet/x/ansi"

width := ansi.StringWidth(highlightedLine)  // Ignores escape sequences
```

### 2. ANSI-Aware Truncation

```go
import "github.com/charmbracelet/x/ansi"

if ansi.StringWidth(line) > maxWidth {
    line = ansi.Truncate(line, maxWidth, "…")
}
```

### 3. Line Splitting

Simple `strings.Split()` works after Chroma highlighting:
```go
highlightedContent := chroma.Highlight(fileContent, "go")
lines := strings.Split(highlightedContent, "\n")
```

Each line retains its ANSI codes because Chroma adds codes per-token.

## Integration Pattern

1. Highlight entire file with Chroma → ANSI string with newlines
2. Split by `\n` → slice of highlighted lines
3. For each line during render:
   - Calculate visible width with `ansi.StringWidth()`
   - Truncate if needed with `ansi.Truncate()`
   - Apply Lip Gloss background (preserves syntax colors)

## Gotchas

1. **Width vs Bytes:** Use `ansi.StringWidth()`, not `len(string)`
2. **Wide characters:** Emojis/CJK count as 2 cells - `StringWidth()` handles this
3. **Reset codes:** Truncation may leave "open" color codes; add `\033[0m` if needed
4. **Color profile:** Use same profile for Chroma and Lip Gloss for consistency

## Decision

> **Decision:** Use `charmbracelet/x/ansi` for ANSI-aware text operations. It's part of the Charm ecosystem (already using Lip Gloss), provides all needed operations, and is used by production tools like moar and Glamour.
