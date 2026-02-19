# Go Syntax Highlighting Libraries Research

## Summary

Based on comprehensive research, **Chroma** (github.com/alecthomas/chroma/v2) is the clear choice for terminal-based syntax highlighting in Go applications.

## Chroma (Recommended)

**Repository:** github.com/alecthomas/chroma/v2

### Architecture

Lexer-Formatter-Style pattern (inspired by Pygments):
- **Lexers** tokenize source code into token streams
- **Formatters** convert tokens to various output formats
- **Styles** define color mappings for token types

### Language Support

200+ languages including all required languages:
- Go, JavaScript, TypeScript, Python, Rust, Java, C, C++

Language detection by:
- Filename matching (primary)
- Explicit name specification
- Content analysis (limited to ~5 languages)

Fallback to plain text if language undetected.

### ANSI Terminal Output

Multiple terminal formatters available:
- **TTY** (8-color): Maps RGB to nearest index color
- **TTY16** (16-color): Uses ANSI codes `\033[3xm`
- **TTY256** (256-color): Full 8-bit color support
- **TTY16m** (true-color): Full RGB 24-bit support

All formatters output ANSI escape codes directly.

### Theme Support

50+ built-in styles including:

**Light themes:** Autumn, Friendly, GitHub, GruvboxLight, SolarizedLight, Xcode, Tango, ParaisoLight, RosePineDawn

**Dark themes:** Dracula, Monokai, Nord, Gruvbox, SolarizedDark, XcodeDark, RosePine, RosePineMoon, Catppuccin variants, DoomOne, GitHubDark, Tokyo Night variants, OneDark

Programmatic access:
- `styles.Names()` - list all available styles
- `styles.Get(name)` - get style by name
- `styles.Register(style)` - register custom styles

### Usage Example

```go
import "github.com/alecthomas/chroma/v2/quick"

// Quick highlighting
err := quick.Highlight(os.Stdout, sourceCode, "go", "terminal256", "monokai")

// Full control
lexer := lexers.Match("file.go")          // Detect by filename
style := styles.Get("dracula")            // Get theme
formatter := formatters.Get("terminal256") // Get formatter
iterator, _ := lexer.Tokenise(nil, sourceCode)
formatter.Format(os.Stdout, style, iterator)
```

### Advantages

- Excellent language coverage (250+ languages)
- Multiple terminal color modes for different terminal capabilities
- Extensive, well-maintained style library
- Very fast (10-20x faster than JavaScript alternatives per Hugo benchmarks)
- Pure Go, no external dependencies
- Used in production: Hugo, Gitea, moar (terminal pager)
- Mature, actively maintained (4.8k+ GitHub stars)

### Gotchas & Limitations

1. **Diff highlighting issue**: Known issue where diff lexer doesn't properly handle final newlines
2. **Token specificity**: Less granular than VSCode/Atom tokenization
3. **Limited auto-detection**: Only ~5 languages support content-based detection
4. **Lexer chattiness**: Some lexers emit many tokens; use `chroma.Coalesce(lexer)` to mitigate

## Alternative: zyedidia/highlight

**Not recommended** for production diff viewer.

- Outputs highlighting metadata, not ANSI codes directly
- Requires additional library (fatih/color) for colorization
- Less mature (56 GitHub stars vs Chroma's 4.8k)
- Smaller language coverage and theme ecosystem
- Would require more custom integration work

## Decision

> **Decision:** Use Chroma for syntax highlighting. It has the best language coverage, native terminal output support, and is proven in production (used by moar terminal pager and Gitea for diff viewing).
