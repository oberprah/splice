# Line Pairing Algorithms for Side-by-Side Diff Views

## Overview

This document researches algorithms for pairing removed and added lines in side-by-side diff views. The goal is to match lines that are "modifications" of each other (similar content) so they can be displayed on the same row, enabling direct visual comparison.

## 1. Common Algorithms for Line Similarity

### 1.1 Levenshtein Distance (Edit Distance)

**Definition**: The minimum number of single-character edits (insertions, deletions, substitutions) required to transform one string into another.

**Example**: The Levenshtein distance between "kitten" and "sitting" is 3:
- kitten → sitten (substitute "s" for "k")
- sitten → sittin (substitute "i" for "e")
- sittin → sitting (insert "g" at end)

**Characteristics**:
- Produces a numeric distance (lower = more similar)
- Character-level granularity
- Time complexity: O(m × n) using dynamic programming
- Space complexity: O(m × n) or O(min(m, n)) with optimization

**Applications**:
- Spell checkers
- Fuzzy text search
- DNA/RNA sequencing
- Duplicate detection
- Plagiarism detection

**Advantages**:
- Well-understood and widely implemented
- Deterministic and precise results
- Works well for small character-level changes

**Disadvantages**:
- Computationally expensive for long strings
- Treats all characters equally (no semantic understanding)
- May not work well for lines with reordered words

### 1.2 Longest Common Subsequence (LCS)

**Definition**: Finds the longest subsequence common to both strings. Unlike substrings, subsequences don't need to occupy consecutive positions.

**Characteristics**:
- Forms the foundation of most diff utilities
- Time complexity: O(m × n) with dynamic programming
- The LCS corresponds exactly to the unchanged parts of a diff

**How it works**:
1. Build a matrix comparing all characters/tokens
2. Use dynamic programming to find optimal subsequence
3. Unchanged parts = LCS, everything else = changes

**Applications**:
- Core algorithm for Unix `diff` utility
- Used by Git and other version control systems
- File comparison tools

**Advantages**:
- Finds semantically meaningful matches
- Works well for identifying blocks of unchanged content
- Can operate on lines, words, or characters

**Disadvantages**:
- O(m × n) complexity can be slow for large inputs
- May not find the most intuitive match in all cases
- Doesn't provide a normalized similarity score

### 1.3 Token-Based Similarity

**Definition**: Treats strings as sets of tokens (typically words or bi-grams) and measures overlap.

**Common approaches**:
1. **Word tokenization**: Split on whitespace/punctuation
2. **N-gram tokenization**: Break into overlapping sequences (e.g., bi-grams: "differ" → "di", "if", "ff", "fe", "er")

#### Dice Coefficient (Sørensen-Dice)

**Formula**: `Dice(A, B) = 2 × |A ∩ B| / (|A| + |B|)`

**Characteristics**:
- Produces score between 0 (no overlap) and 1 (identical)
- Weights commonalities more than differences
- Generally more forgiving than Jaccard
- Dice ≥ 0.7 often considered acceptable threshold

**Example**:
```
A = "fmt.Println(name)"      tokens: [fmt, Println, name]
B = "fmt.Println(userName)"  tokens: [fmt, Println, userName]
Common: [fmt, Println]
Dice = 2 × 2 / (3 + 3) = 0.667
```

#### Jaccard Similarity

**Formula**: `Jaccard(A, B) = |A ∩ B| / |A ∪ B|`

**Characteristics**:
- Also produces score between 0 and 1
- Penalizes differences more than Dice
- Can be converted to/from Dice: `Jaccard = Dice / (2 - Dice)`

**Example** (same strings as above):
```
Common: [fmt, Println]
Union: [fmt, Println, name, userName]
Jaccard = 2 / 4 = 0.5
```

**Advantages** (both metrics):
- Fast to compute
- Language-aware (considers word boundaries)
- Works well when word order changes
- Position-independent

**Disadvantages**:
- Ignores sequence/order information
- Requires tokenization strategy
- May not capture small character-level changes within words

### 1.4 Comparison Summary

| Algorithm | Granularity | Complexity | Best For | Produces |
|-----------|-------------|------------|----------|----------|
| Levenshtein | Character | O(m×n) | Small edits, typos | Distance (0+) |
| LCS | Flexible | O(m×n) | Finding unchanged blocks | Diff script |
| Dice/Jaccard | Token/Word | O(m+n) | Semantic similarity, reordering | Similarity (0-1) |

## 2. How Other Tools Solve This Problem

### 2.1 Git's Diff Algorithms

Git supports four primary diff algorithms that determine how lines are matched between file versions:

#### Myers Algorithm (Default)

- The basic greedy diff algorithm, developed by Eugene W. Myers in 1986
- Currently the default in Git
- Fast but may produce less intuitive diffs in some cases
- No special heuristics for matching lines

#### Patience Diff

- Invented by Bram Cohen (creator of BitTorrent)
- **Key innovation**: Performs a full scan to match up unique lines first
- These unique lines are more likely to be meaningful content than boilerplate
- Uses these anchor points to break the diff into smaller pieces
- Then applies Myers algorithm to the smaller pieces

**Process**:
1. Identify lines that appear exactly once in both versions
2. Match these unique lines as anchor points
3. Recursively diff the regions between anchors

**Advantages**: Often produces more intuitive diffs for code
**Disadvantages**: Slower than Myers; only works well when unique lines exist

#### Histogram Diff

- Extension of patience diff developed by Shawn Pearce (2010) for JGit
- "Support low-occurrence common elements"
- If unique common elements exist, behaves identically to patience diff
- Otherwise, selects elements with lowest occurrence count as anchors

**Process**:
1. Build histogram of line appearances in first file
2. Match elements in second file, count occurrences
3. Select lowest-occurrence LCS as separator
4. Recursively process regions

**Performance**: Faster than patience while maintaining similar quality
**Research finding**: "histogram does better at indicating the intended changes between files" (2019 study)

#### Configuration

```bash
# Set globally
git config --global diff.algorithm histogram

# Per-command
git diff --histogram
git diff --patience
```

### 2.2 GitHub, GitLab, VS Code

#### VS Code

- **Default**: Side-by-side diff editor
- **Line highlighting**: Red for removed lines, green for added lines
- **Intra-line highlighting**: Character-level differences highlighted with brighter background
- **Implementation**: Uses proprietary diff engine (available as `vscode-diff.nvim` port)
  - Line-level diffs using standard algorithms
  - Character-level highlighting: 1.4× brighter for dark themes, 0.92× darker for light themes
  - C-based computation with multi-core parallelization (OpenMP)
- **Configuration**: `"diffEditor.renderSideBySide": true/false`

**Challenge noted by developers**: "It was really hard to make [VS Code diff] understand that each line should be in front of each other in both panes."

#### GitHub/GitLab

- Both use Git's diff algorithms under the hood
- Likely use patience or histogram diff for better readability
- Add syntax highlighting on top of diff output
- Inline (unified) view by default, with side-by-side option
- Intra-line highlighting shows character/word-level changes

### 2.3 Other Diff Tools

#### diffr (Command-line)

- Post-processes diff output for intra-line highlighting
- Produces clean output by default
- Lightweight and fast

#### delta

- Syntax highlighting plus intra-line diff
- Highly customizable
- Written in Rust

#### diff-highlight (Ships with Git)

- Simple intra-line highlighting
- Character/word-based highlighting
- Easy to set up: `git config --global pager.diff "diff-highlight | less"`

#### wikEd diff (JavaScript)

- Visual diff library for inline text comparisons
- Detects and highlights block moves
- Works at word and character level
- Uses Paul Heckel's "technique for isolating differences between files"

## 3. Pairing Strategies

### 3.1 Sequential Pairing (Greedy)

**How it works**: Match the first removed line with the first added line, second with second, etc.

**Advantages**:
- Extremely simple to implement (O(1) per pair)
- Predictable behavior
- No computation overhead
- Works well when changes are mostly modifications in place

**Disadvantages**:
- Produces poor matches when lines are inserted/deleted
- No consideration of similarity
- May pair completely unrelated lines

**Example**:
```
Removed:        Added:
1. foo()        1. bar()
2. baz()        2. qux()
                3. baz()

Sequential pairing:
foo() ↔ bar()   (poor match)
baz() ↔ qux()   (poor match)
∅     ↔ baz()   (unmatched)
```

**When to use**: Only as a fallback when no similarity threshold is met

### 3.2 Greedy Similarity-Based Pairing

**How it works**:
1. Calculate similarity score for all removed-added pairs
2. Sort pairs by similarity score (descending)
3. Greedily match highest-scoring pair
4. Remove matched lines from consideration
5. Repeat until no more pairs above threshold

**Advantages**:
- Simple to implement
- Fast: O(m × n) to compute scores, O(k log k) to sort
- Often produces good results
- Allows threshold to prevent bad matches

**Disadvantages**:
- First matches might be optimal, but later matches can be poor
- "Greedy nearest neighbor matching may result in poor quality matches overall"
- Doesn't consider global optimization
- Sensitive to threshold choice

**Example**:
```
Removed:              Added:
1. fmt.Println(x)     1. fmt.Println(y)
2. return x           2. calculateX()
                      3. return x

Similarity scores:
fmt.Println(x) ↔ fmt.Println(y): 0.9
fmt.Println(x) ↔ calculateX(): 0.1
fmt.Println(x) ↔ return x: 0.3
return x ↔ fmt.Println(y): 0.3
return x ↔ calculateX(): 0.1
return x ↔ return x: 1.0

Greedy matching:
1. return x ↔ return x (score 1.0)
2. fmt.Println(x) ↔ fmt.Println(y) (score 0.9)
3. ∅ ↔ calculateX() (unmatched)
```

**Time complexity**:
- Compute all scores: O(m × n × s) where s is similarity computation cost
- Sort: O(k log k) where k = m × n pairs
- Match: O(k)
- **Total**: O(m × n × s + k log k)

### 3.3 Optimal Assignment (Hungarian Algorithm)

**How it works**:
1. Create cost matrix: similarity scores for all removed-added pairs
2. Use Hungarian algorithm to find optimal global matching
3. Maximizes total similarity across all pairs

**Characteristics**:
- Finds the globally optimal solution
- Time complexity: O(n³) where n = max(removed_count, added_count)
- Named after Hungarian mathematicians Dénes Kőnig and Jenő Egerváry
- Also called Kuhn-Munkres algorithm

**Advantages**:
- "Takes into account the entire system before making any matches"
- Provably optimal (maximizes total similarity)
- Better when there's "competition for controls"
- Handles rectangular matrices (unequal line counts)

**Disadvantages**:
- More complex to implement
- O(n³) can be slow for large hunks (though rarely a problem in practice)
- May produce non-intuitive results if similarity metric is poor
- Overkill for simple cases

**When optimal outperforms greedy**:
- When there's high competition (many similar-looking lines)
- When choosing one match affects quality of other matches
- Research: "When there is a lot of competition for controls, greedy matching performs poorly and optimal matching performs well"

**Example** (same as greedy):
```
Removed:              Added:
1. fmt.Println(x)     1. fmt.Println(y)
2. return x           2. calculateX()
                      3. return x

Hungarian algorithm considers all possible pairings:
- Pairing 1: {1↔1, 2↔3} = 0.9 + 1.0 = 1.9
- Pairing 2: {1↔1, 2↔2} = 0.9 + 0.1 = 1.0
- Pairing 3: {1↔2, 2↔3} = 0.1 + 1.0 = 1.1
- Pairing 4: {1↔3, 2↔1} = 0.3 + 0.3 = 0.6
... etc

Selects Pairing 1 as it maximizes total similarity (1.9)
```

### 3.4 Hybrid Approaches

**Sequential with threshold**:
- Try sequential pairing first
- If similarity is below threshold, try greedy matching
- Fallback to sequential if no good matches found

**Greedy with position bias**:
- Add small bonus to similarity score based on position proximity
- Encourages matching nearby lines when similarity is similar
- Formula: `final_score = similarity + (position_bonus × proximity)`

**Windowed matching**:
- Only consider pairing lines within a certain distance (e.g., ±3 lines)
- Reduces computation for large hunks
- Assumes most modifications are local

## 4. Recommendations for This Project

### 4.1 Context

**Technology stack**:
- Go (excellent string handling, good performance)
- Bubbletea (terminal UI, needs responsive rendering)
- Target: Terminal users expecting fast, intuitive diffs

**Current state**:
- Side-by-side layout already implemented
- Syntax highlighting working
- Need to add line pairing + intra-line highlighting

**Constraints**:
- Keep implementation simple (this is v1)
- Maintain fast rendering performance
- Layer on top of existing diff structure

### 4.2 Recommended Algorithm: Token-Based Similarity (Dice Coefficient)

**Rationale**:

1. **Simplicity**: Most straightforward to implement and reason about
2. **Performance**: O(m + n) tokenization + O(min(m,n)) comparison = very fast
3. **Code-aware**: Word-level tokenization respects identifiers, keywords
4. **Normalized score**: 0-1 range makes threshold selection intuitive
5. **Robust**: Handles reordered words, partial matches gracefully

**Implementation sketch**:
```go
// Tokenize into words/identifiers
func tokenize(line string) []string {
    // Split on whitespace + keep alphanumeric/underscore together
    // "fmt.Println(userName)" → ["fmt", "Println", "userName"]
}

// Compute Dice coefficient
func diceCoefficient(line1, line2 string) float64 {
    tokens1 := tokenize(line1)
    tokens2 := tokenize(line2)

    // Build sets (use map[string]bool for deduplication)
    set1 := toSet(tokens1)
    set2 := toSet(tokens2)

    // Count intersection
    intersection := 0
    for token := range set1 {
        if set2[token] {
            intersection++
        }
    }

    // Dice = 2 * |intersection| / (|set1| + |set2|)
    return 2.0 * float64(intersection) / float64(len(set1) + len(set2))
}
```

**Threshold recommendation**: 0.4 - 0.6
- Research suggests 0.7 for medical imaging (high precision required)
- Code diffs can tolerate more false positives
- Start with 0.5, tune based on real-world testing

### 4.3 Recommended Pairing Strategy: Greedy with Threshold

**Rationale**:

1. **Good enough**: Greedy performs well in practice for diff use cases
2. **Simple**: Straightforward to implement and debug
3. **Fast**: O(m × n) is acceptable for typical hunk sizes (< 100 lines)
4. **Quality control**: Threshold prevents poor matches

**When to consider Hungarian**:
- If user feedback shows poor pairings in practice
- For hunks with many competing similar lines (rare in real code)
- As a future optimization, not v1

**Implementation approach**:
```go
func pairLines(removed, added []Line) []Pair {
    // 1. Compute similarity matrix
    scores := make([][]float64, len(removed))
    for i, r := range removed {
        scores[i] = make([]float64, len(added))
        for j, a := range added {
            scores[i][j] = diceCoefficient(r.Content, a.Content)
        }
    }

    // 2. Create scored pairs and sort
    type ScoredPair struct {
        removedIdx, addedIdx int
        score float64
    }
    pairs := []ScoredPair{}
    for i := range removed {
        for j := range added {
            if scores[i][j] >= threshold {
                pairs = append(pairs, ScoredPair{i, j, scores[i][j]})
            }
        }
    }
    sort.Slice(pairs, func(i, j int) bool {
        return pairs[i].score > pairs[j].score
    })

    // 3. Greedy matching
    usedRemoved := make(map[int]bool)
    usedAdded := make(map[int]bool)
    result := []Pair{}

    for _, sp := range pairs {
        if !usedRemoved[sp.removedIdx] && !usedAdded[sp.addedIdx] {
            result = append(result, Pair{
                Removed: removed[sp.removedIdx],
                Added: added[sp.addedIdx],
            })
            usedRemoved[sp.removedIdx] = true
            usedAdded[sp.addedIdx] = true
        }
    }

    return result
}
```

### 4.4 Intra-Line Highlighting Algorithm

For highlighting changes within paired lines, recommend character-level LCS:

**Rationale**:
- LCS is the standard for diff highlighting
- Character-level provides fine granularity
- Can reuse Go library implementations

**Go libraries available**:
- `github.com/agnivade/levenshtein` - Simple, performant
- `github.com/texttheater/golang-levenshtein` - Provides edit scripts
- `github.com/hbollon/go-edlib` - Multiple algorithms including LCS

**Recommended**: `github.com/agnivade/levenshtein`
- Most straightforward API
- Good performance (handles up to 65K characters)
- Active maintenance
- No complex dependencies

**Alternative if LCS is too complex**: Simple word-level diff
1. Tokenize both lines
2. Find common prefix/suffix of tokens
3. Highlight the middle portion that differs
4. Much simpler but less precise

### 4.5 Trade-offs Summary

| Approach | Accuracy | Complexity | Performance | Maintenance |
|----------|----------|------------|-------------|-------------|
| **Recommended (Dice + Greedy)** | Good | Low | Fast | Easy |
| Levenshtein + Greedy | Better | Medium | Medium | Medium |
| Token + Hungarian | Best | High | Slower | Complex |
| Sequential only | Poor | Trivial | Fastest | Trivial |

**V1 recommendation**: Dice + Greedy
**Future consideration**: Add Hungarian as an option if needed

### 4.6 Implementation Plan

**Phase 1: Line pairing**
1. Implement Dice coefficient similarity function
2. Add greedy pairing algorithm with threshold (0.5)
3. Update data structures to track paired lines
4. Modify rendering to display pairs on same row

**Phase 2: Intra-line highlighting**
1. Add `github.com/agnivade/levenshtein` dependency
2. For paired lines, compute character-level diff
3. Mark changed ranges with brighter background
4. Update rendering to apply multi-level styles

**Phase 3: Tuning**
1. Test on real-world diffs
2. Adjust threshold if needed
3. Consider position bias for ambiguous cases
4. Add unit tests for edge cases

**Estimated complexity**: 200-300 lines of new code, plus tests

### 4.7 Edge Cases to Consider

1. **Empty lines**: Treat carefully (many empties can create false matches)
2. **Very long lines**: Dice coefficient handles well, but may want length penalty
3. **Identical lines**: Perfect score (1.0), but may want position proximity bonus
4. **No good matches**: All scores below threshold - leave unpaired
5. **Unequal counts**: Greedy naturally handles (some lines stay unpaired)

## Sources

### Git Diff Algorithms
- [Git diff-options Documentation](https://git-scm.com/docs/diff-options/2.6.7)
- [More on "histogram diff", and a working program](https://www.raygard.net/2025/01/29/a-histogram-diff-implementation/)
- [Line Based Diffs - Difftastic Wiki](https://github.com/Wilfred/difftastic/wiki/Line-Based-Diffs)
- [How different are different diff algorithms in Git?](https://link.springer.com/article/10.1007/s10664-019-09772-z)
- [When to Use Each of the Git Diff Algorithms](https://luppeng.wordpress.com/2020/10/10/when-to-use-each-of-the-git-diff-algorithms/)
- [The patience diff algorithm](https://blog.jcoglan.com/2017/09/19/the-patience-diff-algorithm/)
- [Git Source Code Review: Diff Algorithms](https://www.fabiensanglard.net/git_code_review/diff.php)

### Side-by-Side Diff Implementation
- [GNU diff Options](https://www.gnu.org/software/diffutils/manual/html_node/diff-Options.html)
- [Paul Heckel's Diff Algorithm](https://gist.github.com/ndarville/3166060)
- [Python difflib documentation](https://docs.python.org/3/library/difflib.html)
- [Side-by-side diff for source code - Phabricator](https://secure.phabricator.com/T6791)

### Levenshtein Distance
- [Levenshtein distance - Wikipedia](https://en.wikipedia.org/wiki/Levenshtein_distance)
- [Understanding the Levenshtein Distance Equation for Beginners](https://medium.com/@ethannam/understanding-the-levenshtein-distance-equation-for-beginners-c4285a5604f0)
- [Levenshtein Distance Computation - Baeldung](https://www.baeldung.com/cs/levenshtein-distance-computation)
- [Introduction to Levenshtein distance - GeeksforGeeks](https://www.geeksforgeeks.org/dsa/introduction-to-levenshtein-distance/)
- [Edit distance - Wikipedia](https://en.wikipedia.org/wiki/Edit_distance)

### VS Code and Other Tools
- [Source Control in VS Code](https://code.visualstudio.com/docs/sourcecontrol/overview)
- [Support inline and side by side view for Diff files - VS Code Issue](https://github.com/Microsoft/vscode/issues/34623)
- [Comparing Files in Visual Studio Code](https://semanticdiff.com/blog/visual-studio-code-compare-files/)
- [vscode-diff.nvim - VS Code diff algorithm for Neovim](https://github.com/esmuellert/vscode-diff.nvim)

### Longest Common Subsequence
- [Longest Common Subsequence - GeeksforGeeks](https://www.geeksforgeeks.org/dsa/longest-common-subsequence-dp-4/)
- [diff-lcs - Ruby implementation](https://github.com/halostatue/diff-lcs)
- [lcs_diff - GitHub implementation](https://github.com/wk-cof/lcs_diff)
- [Longest common subsequence - Wikipedia](https://en.wikipedia.org/wiki/Longest_common_subsequence)
- [Longest Common Subsequence Diff Part 1](https://nghiatran.me/longest-common-subsequence-diff-part-1)
- [Write your own diff for fun](https://alex.dzyoba.com/blog/writing-diff/)

### Token-Based Similarity
- [String Similarity Metrics: Token Methods - Baeldung](https://www.baeldung.com/cs/string-similarity-token-methods)
- [Similarity Coefficients: A Beginner's Guide](https://medium.com/@igniobydigitate/similarity-coefficients-a-beginners-guide-to-measuring-string-similarity-d84da77e8c5a)
- [Text-Sorensen - Dice/Jaccard coefficient library](https://github.com/thundergnat/Text-Sorensen)
- [Set-based (Jaccard) similarity](https://ds1.datascience.uchicago.edu/08/1/Set_Based_Similarity.html)
- [Understanding the Dice Similarity Coefficient](https://medium.com/@caring_smitten_gerbil_914/understanding-the-dice-similarity-coefficient-a-practical-guide-for-data-scientists-a83fef69dbf4)
- [String metric - Wikipedia](https://en.wikipedia.org/wiki/String_metric)

### Hungarian Algorithm
- [Hungarian algorithm - Wikipedia](https://en.wikipedia.org/wiki/Hungarian_algorithm)
- [Hungarian Maximum Matching Algorithm - Brilliant](https://brilliant.org/wiki/hungarian-matching/)
- [Hungarian algorithm - CP-Algorithms](https://cp-algorithms.com/graph/hungarian-algorithm.html)
- [Exactly how the Hungarian Algorithm Works](https://www.thinkautonomous.ai/blog/hungarian-algorithm/)
- [Optimum Assignment and the Hungarian Algorithm](https://medium.com/data-science/optimum-assignment-and-the-hungarian-algorithm-8b1027628028)
- [The Hungarian Method - TU München](https://algorithms.discrete.ma.tum.de/graph-algorithms/matchings-hungarian-method/index_en.html)

### Go Libraries
- [agnivade/levenshtein - Go package](https://github.com/agnivade/levenshtein)
- [texttheater/golang-levenshtein - Go implementation](https://github.com/texttheater/golang-levenshtein)
- [hbollon/go-edlib - Multiple string distance algorithms](https://github.com/hbollon/go-edlib)
- [agnivade/levenshtein - pkg.go.dev](https://pkg.go.dev/github.com/agnivade/levenshtein)
- [Introduction to string edit distance in Golang](https://medium.com/@hugo.bollon/introduction-to-string-edit-distance-and-levenshtein-implementation-in-golang-8ed81b5b06d9)

### Intra-Line Diff Highlighting
- [Vim PR #16881 - Per-character/word diff](https://github.com/vim/vim/pull/16881)
- [Intra-line highlighting - Difftastic Issue](https://github.com/Wilfred/difftastic/issues/33)
- [Word-wise Git Diff highlighting](https://maximsmol.medium.com/improving-git-diffs-4519a6541cd1)
- [vscode-diff.nvim](https://github.com/esmuellert/vscode-diff.nvim)
- [Dress Up Your Git Diffs With Word-level Highlights](https://www.viget.com/articles/dress-up-your-git-diffs-with-word-level-highlights/)
- [diffchar.vim - Highlight exact differences](https://github.com/rickhowe/diffchar.vim)
- [Line or Word Diffs - diff-match-patch Wiki](https://github.com/google/diff-match-patch/wiki/Line-or-Word-Diffs)

### Greedy vs Optimal Matching
- [Greedy Algorithm & Greedy Matching in Statistics](https://www.statisticshowto.com/greedy-algorithm-matching/)
- [Data Matching – Optimal and Greedy - NCSS](https://www.ncss.com/wp-content/themes/ncss/pdf/Procedures/NCSS/Data_Matching-Optimal_and_Greedy.pdf)
- [Optimal Matching - Harvard](https://r.iq.harvard.edu/docs/matchit/2.4-15/Optimal_Matching.html)
- [A comparison of 12 algorithms for matching on the propensity score](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC4285163/)
- [An Optimal Algorithm for On-line Bipartite Matching](https://people.eecs.berkeley.edu/~vazirani/pubs/online.pdf)
