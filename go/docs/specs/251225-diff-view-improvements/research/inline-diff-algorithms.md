# Inline Diff Algorithms Research

This document researches algorithms and approaches for computing inline/character-level diffs within pairs of lines, to highlight what changed between two similar lines.

## 1. Character-Level Diff Algorithms

### 1.1 Myers Diff Algorithm

The Myers diff algorithm, published by Eugene Myers in 1986 ("An O(ND) Difference Algorithm and Its Variations"), is the gold standard for computing minimal edit sequences between two strings.

**How it works:**
- Models the problem as finding the shortest path over an "edit graph"
- The (x, y) coordinates in the grid correspond to steps in the editing process
- Moving rightward (increasing x) = deleting a character
- Moving downward (increasing y) = inserting a character
- Diagonal moves = consuming an equal character from both strings (no edit)
- When strings have the same character at position indexes, there's a diagonal in the graph

**Key characteristics:**
- Fast and produces good quality diffs most of the time
- Greedy approach: tries to consume as many equal characters before making a change
- Prefers deletions over insertions when given a choice
- **Time complexity:** O(min(N,M) * D) where N,M are string lengths and D is number of differences
- **Space complexity:** O(min(N,M) + D)

**Character-level application:**
- Can be applied directly to strings (treating each character as an element)
- Produces minimal edit sequences
- Well-suited for small to medium sized strings (like code lines)

### 1.2 LCS-Based Approach

The Longest Common Subsequence (LCS) algorithm provides a simpler alternative to Myers diff.

**How it relates to diffing:**
- Finding the LCS is equivalent to finding the unchanged parts of a diff
- Everything not in the LCS is either removed or added
- Maximizes unchanged parts = minimizes change markers

**Algorithm approach:**
- Uses dynamic programming with memoization
- Has optimal substructure and overlapping subproblems
- **Time complexity:** O(M * N) where M,N are string lengths
- **Space complexity:** O(M * N) for the DP table

**Pros:**
- Conceptually simpler than Myers
- Easier to implement from scratch
- Reliable results

**Cons:**
- Slower than Myers for typical cases
- Uses more memory
- Less optimized than Myers for minimizing edits

### 1.3 Ratcliff-Obershelp (Gestalt Pattern Matching)

This algorithm from the late 1980s is used in Python's difflib and go-difflib.

**How it works:**
- Finds the longest contiguous matching subsequence (no "junk" elements)
- Recursively applies the same idea to sequences left and right of the match
- Different from Myers - not focused on minimal edits

**Key characteristics:**
- Does NOT yield minimal edit sequences
- Tends to yield matches that "look right" to people
- Produces "human-friendly diffs"
- More intuitive for natural language

**Trade-offs:**
- Better for human readability
- Not optimal for code where precision matters
- Slower than Myers

## 2. Word-Level Diff Algorithms

### 2.1 Concept

Word-level diffing uses the same algorithms (Myers, LCS, etc.) but operates on tokenized sequences rather than character sequences.

**Process:**
1. Tokenize both strings into words/symbols
2. Apply diff algorithm to token sequences
3. Produce diff output in terms of tokens

**Tokenization strategies:**
- **Whitespace-based:** Split on spaces, tabs, newlines (simplest)
- **Punctuation-aware:** Split on whitespace AND punctuation boundaries
- **Syntax-aware:** Use programming language lexer to tokenize into semantic units
- **Regex-based:** Use custom regex patterns (e.g., Git's `--word-diff-regex`)

### 2.2 Git's Word-Diff Implementation

Git supports word-level diffing with several options:

```bash
# Basic word diff
git diff --word-diff

# Colored word diff
git diff --word-diff=color

# Custom word definition via regex
git diff --word-diff-regex='[^[:space:]]+'

# Character-level diff (each char is a "word")
git diff --word-diff-regex=.
```

**Modes:**
- `color`: Highlights changed words using colors only
- `plain`: Shows words as `[-removed-]` and `{+added+}`
- `porcelain`: Special line-based format for script consumption

**Underlying algorithms:** Git supports multiple:
- `myers`: Basic greedy diff (default)
- `minimal`: Extra time for smallest possible diff
- `patience`: Better for large files with unique lines
- `histogram`: Extension of patience algorithm

### 2.3 Pros and Cons

**Pros of word-level:**
- More readable for prose and long lines
- Reduces "chaff" problem (spurious single-character matches)
- Better for natural language documents
- Larger tokens = fewer comparisons = potentially faster
- Matches human intuition better for paragraphs

**Cons of word-level:**
- Less precise for code where single character changes matter
- Requires tokenization logic (more complexity)
- May miss important small changes (e.g., `==` vs `===`)
- Token boundaries may not align with semantic changes
- Extra step: need to map tokens back to character positions for highlighting

**Pros of character-level:**
- Maximum precision - catches every change
- Simpler implementation (no tokenization needed)
- Better for code where operators/symbols matter
- Works consistently across all content types
- Direct mapping to visual presentation

**Cons of character-level:**
- Can produce noisy diffs with many small matches
- May not align with human perception of changes
- Slower for very long lines (more comparisons)
- Can be less readable for natural language

## 3. Go Libraries Available

### 3.1 sergi/go-diff

**Repository:** https://github.com/sergi/go-diff
**Package:** `github.com/sergi/go-diff/diffmatchpatch`
**Stars:** 1,445+
**Status:** Active, well-maintained

**Description:**
- Go port of Neil Fraser's google-diff-match-patch library
- Implements Myers 1986 algorithm
- Most popular Go diffing library

**Key functions:**
```go
// Main diff function
DiffMain(text1, text2 string, checklines bool) []Diff

// Helper functions
DiffCommonPrefix(text1, text2 string) int
DiffCommonSuffix(text1, text2 string) int
```

**Diff struct:**
```go
type Diff struct {
    Type Operation  // DiffDelete, DiffInsert, or DiffEqual
    Text string     // The text segment
}
```

**Pros:**
- Battle-tested implementation (based on Google's library)
- Well-documented
- Supports character-level, word-level, and line-level
- Includes match and patch operations
- Clean, simple API

**Cons:**
- Larger library (includes features beyond basic diffing)
- External dependency

**Ease of use:** Excellent - simple API, good examples

### 3.2 pmezard/go-difflib

**Repository:** https://github.com/pmezard/go-difflib
**Package:** `github.com/pmezard/go-difflib/difflib`

**Description:**
- Partial port of Python's difflib module
- Uses Ratcliff-Obershelp algorithm (NOT Myers)
- Focused on unified/context diff output

**Key functions:**
```go
// Unified diff
UnifiedDiff struct {
    A, B     []string  // Input sequences (lines)
    FromFile, ToFile string
    Context  int
}
GetUnifiedDiffString(diff UnifiedDiff) (string, error)

// Sequence matcher
type SequenceMatcher struct { ... }
```

**Pros:**
- Produces "human-friendly" diffs
- Good for testing (comparing expected vs actual)
- Native unified diff format

**Cons:**
- NOT minimal edits (uses different algorithm than Myers)
- **Maintainer has abandoned the project** ("no longer the time nor interest")
- Designed for line-level, not character-level
- Less flexible API

**Ease of use:** Moderate - requires understanding of unified diff format

### 3.3 golang.org/x/tools/internal/diff

**Package:** `golang.org/x/tools/internal/diff`
**Subpackage:** `golang.org/x/tools/internal/diff/myers`
**Status:** Internal package (not stable API)

**Description:**
- Official Go tools package
- Implements Myers diff algorithm
- Used internally by Go language server

**Key functions:**
```go
// Strings computes differences respecting rune boundaries
func Strings(before, after string) []Edit

type Edit struct {
    Start, End int    // Byte offsets
    New        string // Replacement text
}
```

**Pros:**
- From official Go project
- Clean, simple API
- Rune-aware (respects UTF-8 boundaries)
- Lightweight

**Cons:**
- **Internal package** - not meant for external use
- API may change without notice
- Limited documentation
- Basic functionality only

**Note:** There's a proposal (golang/go#58893) to make this a public API, but it hasn't been accepted yet.

### 3.4 Other Options

**cj1128/myers-diff**
- Standalone Myers implementation
- Supports character mode: `myers-diff -char src dst`
- CLI tool rather than library

**github.com/MFAshby/myers**
- Another Myers implementation
- Minimal list of differences
- O(min(len(e),len(f))) space

**akedrou/textdiff**
- Similar to x/tools/diff
- May be based on the same code

### 3.5 Library Recommendation

**For this project: sergi/go-diff**

**Reasoning:**
1. **Most mature and battle-tested** - based on Google's implementation
2. **Active maintenance** - unlike pmezard/go-difflib
3. **Stable API** - unlike golang.org/x/tools/internal packages
4. **Flexible** - supports character, word, and line level
5. **Good documentation** - clear examples and API docs
6. **Clean output format** - simple Diff structs with Type and Text
7. **MIT licensed** - permissive

**Alternative:** If we want minimal dependencies and a truly simple implementation, we could implement a basic Myers or LCS algorithm ourselves (200-300 lines of code).

## 4. Visual Presentation

### 4.1 Data Structures for Inline Changes

The goal is to represent which spans within a line are changed vs unchanged, so they can be styled differently.

**Option 1: Diff Segments (from sergi/go-diff)**
```go
type Operation int
const (
    DiffDelete Operation = -1
    DiffEqual  Operation = 0
    DiffInsert Operation = 1
)

type Diff struct {
    Type Operation
    Text string
}

// Example result for "Hello World" -> "Hello Go"
[]Diff{
    {Type: DiffEqual, Text: "Hello "},
    {Type: DiffDelete, Text: "World"},
    {Type: DiffInsert, Text: "Go"},
}
```

**Pros:**
- Clean separation of diff logic from rendering
- Easy to test diff computation independently
- Flexible - can be rendered in multiple ways
- Natural output from most diff algorithms

**Option 2: Styled Spans**
```go
type SpanStyle int
const (
    SpanNormal SpanStyle = iota
    SpanAdded
    SpanRemoved
)

type Span struct {
    Text  string
    Style SpanStyle
}

// For rendering with Lip Gloss
func (s Span) Render() string {
    switch s.Style {
    case SpanAdded:
        return addedStyle.Render(s.Text)
    case SpanRemoved:
        return removedStyle.Render(s.Text)
    default:
        return s.Text
    }
}
```

**Pros:**
- Directly renderable
- Can include rich styling info (colors, bold, etc.)
- Matches TUI framework patterns (similar to ratatui::text::Span)

**Cons:**
- Couples diff logic to rendering
- Need separate spans for old and new lines

**Option 3: Change Ranges (Position-based)**
```go
type ChangeType int
const (
    ChangeNone ChangeType = iota
    ChangeAdded
    ChangeRemoved
)

type CharChange struct {
    Start, End int        // Character positions
    Type       ChangeType
}

type LineChanges struct {
    Line    string
    Changes []CharChange
}
```

**Pros:**
- Compact representation
- Easy to overlay on original text
- Good for applying multiple styles

**Cons:**
- More complex to render
- Need to handle overlapping ranges
- Requires position tracking

### 4.2 Recommended Approach

**Use Option 1 (Diff Segments) as the internal representation:**
- Compute diff using sergi/go-diff (or custom implementation)
- Store as `[]Diff` with DiffEqual/DiffInsert/DiffDelete operations
- Keep this in the state data structure

**Convert to Option 2 (Styled Spans) at render time:**
- In the view layer, convert `[]Diff` to styled strings
- Use Lip Gloss styles for coloring
- Apply different styles for added/removed/unchanged spans

**Example flow:**
```go
// In data loading (e.g., loadDiff)
oldLine := "Hello World"
newLine := "Hello Go"
oldDiffs := computeInlineDiff(oldLine, newLine, true)  // for old line
newDiffs := computeInlineDiff(oldLine, newLine, false) // for new line

// Store in state
type LineChange struct {
    OldLine    string
    NewLine    string
    OldDiffs   []diffmatchpatch.Diff
    NewDiffs   []diffmatchpatch.Diff
}

// In view rendering
func renderLine(diffs []diffmatchpatch.Diff) string {
    var result strings.Builder
    for _, diff := range diffs {
        switch diff.Type {
        case diffmatchpatch.DiffDelete:
            result.WriteString(removedStyle.Render(diff.Text))
        case diffmatchpatch.DiffInsert:
            result.WriteString(addedStyle.Render(diff.Text))
        case diffmatchpatch.DiffEqual:
            result.WriteString(diff.Text)
        }
    }
    return result.String()
}
```

### 4.3 Styling Approach

**Terminal capabilities:**
- Background colors for changed spans
- Foreground colors for added/removed markers
- Bold/underline for emphasis

**Recommended styles (following GitHub/GitLab conventions):**

For removed sections (old line):
- Red background for deleted text
- Darker red for unchanged text in deleted line

For added sections (new line):
- Green background for inserted text
- Darker green for unchanged text in added line

**Example Lip Gloss styles:**
```go
var (
    // For deleted line
    deletedBgStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#3f1f1f")).  // Dark red
        Foreground(lipgloss.Color("#ff8888"))   // Light red

    deletedTextStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#6f2020")).  // Brighter red
        Foreground(lipgloss.Color("#ffcccc"))

    // For added line
    addedBgStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#1f3f1f")).  // Dark green
        Foreground(lipgloss.Color("#88ff88"))   // Light green

    addedTextStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("#206f20")).  // Brighter green
        Foreground(lipgloss.Color("#ccffcc"))
)
```

### 4.4 Handling Edge Cases

**Empty changes:**
- If lines are identical, no need for inline diff
- Just show as equal (can skip computation)

**Completely different lines:**
- If diff has no DiffEqual segments, treat as full replacement
- Highlight entire old line as removed, entire new line as added

**Very long lines:**
- Consider truncating or wrapping
- May want to limit diff computation to first N characters for performance

**Whitespace changes:**
- Decision: show whitespace changes or ignore?
- Git default: show them (they can be significant in code)
- Could use visible whitespace characters (·, →) for clarity

## 5. Recommendations

### 5.1 Algorithm Choice: Character-Level Myers (via sergi/go-diff)

**Recommendation: Use character-level Myers diff via sergi/go-diff**

**Reasoning:**

1. **Character-level is better for code:**
   - Code diffs require precision - even single character changes matter
   - Operators, punctuation, and symbols are semantically important
   - Examples: `==` vs `===`, `!==` vs `!=`, `->` vs `.`
   - Word-level would miss or misrepresent these changes

2. **Myers algorithm is optimal:**
   - Produces minimal edit sequences
   - Fast enough for line-length strings
   - Industry standard (used by Git)
   - Well-understood and documented

3. **sergi/go-diff provides:**
   - Battle-tested implementation
   - Clean API with simple Diff structs
   - Active maintenance
   - Good documentation
   - Character-level support out of the box

4. **Simplicity:**
   - No tokenization logic needed
   - Direct application to strings
   - Straightforward rendering
   - Fewer edge cases

5. **Consistency:**
   - Character-level works for all content (code, prose, config, etc.)
   - No need for language-specific tokenization
   - Predictable behavior

### 5.2 Implementation Approach

**Phase 1: Basic character-level inline diff**
1. Add sergi/go-diff dependency: `go get github.com/sergi/go-diff/diffmatchpatch`
2. Implement `computeInlineDiff(oldLine, newLine string) []Diff`
3. Store diffs in state for modified line pairs
4. Render with basic styling (red/green backgrounds)

**Phase 2: Optimization**
1. Only compute inline diff for pairs of changed lines (not all lines)
2. Cache results if lines are viewed multiple times
3. Add threshold - skip if lines are too different (no common parts)

**Phase 3: Enhancement**
1. Fine-tune styling based on user feedback
2. Add whitespace visualization options
3. Consider word-level as an alternative mode (togglable)

### 5.3 Alternative Consideration: Word-Level as Optional Mode

While character-level is recommended as the default, we could add word-level as an optional mode for specific use cases:

**When word-level might be useful:**
- Long lines with many changes
- Natural language files (documentation, comments)
- User preference

**Implementation:**
- Add a toggle in the UI (e.g., `w` key to switch modes)
- Use same sergi/go-diff library with tokenization wrapper
- Store mode preference in app config

**Simple tokenization for word-level:**
```go
func tokenize(s string) []string {
    // Split on whitespace and punctuation boundaries
    re := regexp.MustCompile(`\w+|[^\w\s]`)
    return re.FindAllString(s, -1)
}

func computeWordDiff(old, new string) []Diff {
    oldWords := tokenize(old)
    newWords := tokenize(new)

    // Join with special delimiter
    oldText := strings.Join(oldWords, "\x00")
    newText := strings.Join(newWords, "\x00")

    dmp := diffmatchpatch.New()
    diffs := dmp.DiffMain(oldText, newText, false)

    // Post-process to restore spaces
    // ...
}
```

### 5.4 Summary

**Primary recommendation:**
- **Algorithm:** Character-level Myers diff
- **Library:** sergi/go-diff (github.com/sergi/go-diff)
- **Data structure:** Store `[]diffmatchpatch.Diff` in state
- **Rendering:** Convert to styled spans in view layer using Lip Gloss

**Rationale:**
- Character-level is more appropriate for code
- Myers is optimal and industry-standard
- sergi/go-diff is mature, maintained, and easy to use
- Simple implementation with room for enhancement

**Optional enhancement:**
- Add word-level mode as a toggleable alternative
- Use same library with tokenization wrapper
- Implement only if user feedback indicates need

## References

### Algorithm Resources
- [The Myers diff algorithm: part 1 – The If Works](https://blog.jcoglan.com/2017/02/12/the-myers-diff-algorithm-part-1/)
- [The Myers Difference Algorithm](https://www.nathaniel.ai/myers-diff/)
- [Comparing Strings: An Analysis of Diff Algorithms](https://www.somethinkodd.com/oddthinking/2006/01/16/comparing-strings-an-analysis-of-diff-algorithms/)
- [Longest Common Subsequence (LCS) - GeeksforGeeks](https://www.geeksforgeeks.org/dsa/longest-common-subsequence-dp-4/)
- [Longest Common Subsequence Diff [Part 1] · Noah Tran's mad lab](https://nghiatran.me/longest-common-subsequence-diff-part-1)
- [Diffing - florian.github.io](https://florian.github.io/diffing/)

### Go Libraries
- [GitHub - sergi/go-diff: Diff, match and patch text in Go](https://github.com/sergi/go-diff)
- [go-diff module - github.com/sergi/go-diff - Go Packages](https://pkg.go.dev/github.com/sergi/go-diff)
- [GitHub - pmezard/go-difflib: Partial port of Python difflib package to Go](https://github.com/pmezard/go-difflib)
- [difflib package - github.com/pmezard/go-difflib/difflib - Go Packages](https://pkg.go.dev/github.com/pmezard/go-difflib/difflib)
- [myers package - golang.org/x/tools/internal/diff/myers - Go Packages](https://pkg.go.dev/golang.org/x/tools/internal/diff/myers)
- [diff package - golang.org/x/tools/internal/diff - Go Packages](https://pkg.go.dev/golang.org/x/tools/internal/diff)
- [proposal: x/tools/diff: a package for computing text differences · Issue #58893 · golang/go](https://github.com/golang/go/issues/58893)
- [GitHub - cj1128/myers-diff: Golang implementation for myers diff algorithm](https://github.com/cj1128/myers-diff)

### Word-Level vs Character-Level
- [Neil Fraser: Writing: Diff Strategies](https://neil.fraser.name/writing/diff/)
- [Line or Word Diffs · google/diff-match-patch Wiki](https://github.com/google/diff-match-patch/wiki/Line-or-Word-Diffs)
- [Word level comparison instead of character level · Issue #45](https://github.com/google/diff-match-patch/issues/45)

### Git Implementation
- [Git - git-diff Documentation](https://git-scm.com/docs/git-diff)
- [git diff by word · Code with Hugo](https://codewithhugo.com/git-diff-by-word/)
- [Word-wise Git Diff highlighting | Medium](https://maximsmol.medium.com/improving-git-diffs-4519a6541cd1)
- [Git - diff-options Documentation](https://git-scm.com/docs/diff-options/2.6.7)

### Terminal Presentation
- [GitHub - dandavison/delta: A syntax-highlighting pager for git, diff, grep, and blame output](https://github.com/dandavison/delta)
- [Diff Terminal Tools - Terminal Trove](https://terminaltrove.com/categories/diff/)
- [diffr - word-by-word diff highlighting utility - LinuxLinks](https://www.linuxlinks.com/diffr-word-by-word-diff-highlighting-utility/)
- [Better Diff Highlighting in Git](https://joelclermont.com/post/2021-02/better-diff-highlighting-in-git/)
- [GitHub - da-x/fancydiff: Colorful Git diffs for terminal and web, including source syntax highlighting](https://github.com/da-x/fancydiff)
- [Syntax Highlighting Diffs In Code - Steven Hicks](https://www.stevenhicks.me/blog/2021/05/syntax-highlighting-diffs/)
