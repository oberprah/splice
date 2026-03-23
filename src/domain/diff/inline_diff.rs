#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct InlineSpan {
    pub start_col: usize,
    pub end_col: usize,
}

/// A token from splitting a line into words and separators.
#[derive(Debug, Clone, PartialEq, Eq)]
struct Token {
    text: Vec<char>,
    /// Character offset in the original line.
    offset: usize,
}

/// Computes inline diff spans using a two-level approach:
///
/// 1. **Word-level**: tokenize both lines, run LCS on tokens to find which
///    words changed.
/// 2. **Character-level**: for each pair of changed words, run character LCS
///    to pinpoint exactly which characters differ.
///
/// This avoids the fragmented highlights that pure character-level LCS
/// produces on word renames while keeping precision for small edits.
pub fn compute_inline_spans(old_text: &str, new_text: &str) -> (Vec<InlineSpan>, Vec<InlineSpan>) {
    if old_text.is_empty() || new_text.is_empty() {
        return (Vec::new(), Vec::new());
    }

    let old_chars: Vec<char> = old_text.chars().collect();
    let new_chars: Vec<char> = new_text.chars().collect();

    if old_chars.len() > 500 || new_chars.len() > 500 {
        return (Vec::new(), Vec::new());
    }

    // Character-level similarity check upfront: if the lines share fewer than
    // 20% common characters, they're too different for inline highlighting.
    let char_lcs_len = char_lcs_length(&old_chars, &new_chars);
    let min_char_len = old_chars.len().min(new_chars.len());
    if min_char_len > 0 && char_lcs_len * 5 < min_char_len {
        return (Vec::new(), Vec::new());
    }

    let old_tokens = tokenize(&old_chars);
    let new_tokens = tokenize(&new_chars);

    // Word-level LCS
    let n = old_tokens.len();
    let m = new_tokens.len();

    let mut dp = vec![vec![0usize; m + 1]; n + 1];
    for i in 1..=n {
        for j in 1..=m {
            if old_tokens[i - 1].text == new_tokens[j - 1].text {
                dp[i][j] = dp[i - 1][j - 1] + 1;
            } else {
                dp[i][j] = dp[i - 1][j].max(dp[i][j - 1]);
            }
        }
    }

    // Backtrack to classify each token as common or changed
    let mut old_states = vec![TokenState::Changed; n];
    let mut new_states = vec![TokenState::Changed; m];

    let mut i = n;
    let mut j = m;
    while i > 0 && j > 0 {
        if old_tokens[i - 1].text == new_tokens[j - 1].text {
            old_states[i - 1] = TokenState::Common;
            new_states[j - 1] = TokenState::Common;
            i -= 1;
            j -= 1;
        } else if dp[i - 1][j] >= dp[i][j - 1] {
            i -= 1;
        } else {
            j -= 1;
        }
    }

    // Collect changed token groups and pair them for character-level refinement.
    // A "group" is a maximal run of consecutive Changed tokens on each side,
    // between the same pair of Common anchors.
    let old_groups = collect_changed_groups(&old_tokens, &old_states);
    let new_groups = collect_changed_groups(&new_tokens, &new_states);

    let mut old_spans = Vec::new();
    let mut new_spans = Vec::new();

    // Pair groups positionally: group 0 on old with group 0 on new, etc.
    let paired = old_groups.len().min(new_groups.len());
    for idx in 0..paired {
        let (og_start, og_end) = old_groups[idx];
        let (ng_start, ng_end) = new_groups[idx];
        refine_group(
            &old_chars[og_start..og_end],
            og_start,
            &new_chars[ng_start..ng_end],
            ng_start,
            &mut old_spans,
            &mut new_spans,
        );
    }

    // Unpaired groups on old side: highlight entirely
    for &(start, end) in &old_groups[paired..] {
        if start < end {
            old_spans.push(InlineSpan {
                start_col: start,
                end_col: end,
            });
        }
    }

    // Unpaired groups on new side: highlight entirely
    for &(start, end) in &new_groups[paired..] {
        if start < end {
            new_spans.push(InlineSpan {
                start_col: start,
                end_col: end,
            });
        }
    }

    (old_spans, new_spans)
}

/// Computes just the LCS length between two char slices (no backtracking needed).
fn char_lcs_length(a: &[char], b: &[char]) -> usize {
    let n = a.len();
    let m = b.len();
    let mut prev = vec![0usize; m + 1];
    let mut curr = vec![0usize; m + 1];
    for i in 1..=n {
        for j in 1..=m {
            if a[i - 1] == b[j - 1] {
                curr[j] = prev[j - 1] + 1;
            } else {
                curr[j] = prev[j].max(curr[j - 1]);
            }
        }
        std::mem::swap(&mut prev, &mut curr);
        curr.fill(0);
    }
    prev[m]
}

/// Tokenizes a char slice into words and whitespace/punctuation separators.
/// Each token records its starting offset in the original line.
fn tokenize(chars: &[char]) -> Vec<Token> {
    let mut tokens = Vec::new();
    let mut i = 0;

    while i < chars.len() {
        let start = i;
        if chars[i].is_alphanumeric() || chars[i] == '_' {
            // Word token: alphanumeric + underscores
            while i < chars.len() && (chars[i].is_alphanumeric() || chars[i] == '_') {
                i += 1;
            }
        } else if chars[i].is_whitespace() {
            // Whitespace token
            while i < chars.len() && chars[i].is_whitespace() {
                i += 1;
            }
        } else {
            // Single punctuation character as its own token
            i += 1;
        }
        tokens.push(Token {
            text: chars[start..i].to_vec(),
            offset: start,
        });
    }

    tokens
}

/// Collects contiguous runs of Changed tokens as (start_char_offset, end_char_offset) ranges.
fn collect_changed_groups(tokens: &[Token], states: &[TokenState]) -> Vec<(usize, usize)> {
    let mut groups = Vec::new();
    let mut i = 0;
    while i < tokens.len() {
        if states[i] == TokenState::Changed {
            let start = tokens[i].offset;
            let mut end = tokens[i].offset + tokens[i].text.len();
            i += 1;
            while i < tokens.len() && states[i] == TokenState::Changed {
                end = tokens[i].offset + tokens[i].text.len();
                i += 1;
            }
            groups.push((start, end));
        } else {
            i += 1;
        }
    }
    groups
}

/// Refines a paired group of changed characters using character-level LCS.
/// Appends resulting spans (in full-line coordinates) to old_spans / new_spans.
fn refine_group(
    old_slice: &[char],
    old_offset: usize,
    new_slice: &[char],
    new_offset: usize,
    old_spans: &mut Vec<InlineSpan>,
    new_spans: &mut Vec<InlineSpan>,
) {
    let n = old_slice.len();
    let m = new_slice.len();

    if n == 0 && m == 0 {
        return;
    }
    if n == 0 {
        new_spans.push(InlineSpan {
            start_col: new_offset,
            end_col: new_offset + m,
        });
        return;
    }
    if m == 0 {
        old_spans.push(InlineSpan {
            start_col: old_offset,
            end_col: old_offset + n,
        });
        return;
    }

    // Character-level LCS within the group
    let mut dp = vec![vec![0usize; m + 1]; n + 1];
    for i in 1..=n {
        for j in 1..=m {
            if old_slice[i - 1] == new_slice[j - 1] {
                dp[i][j] = dp[i - 1][j - 1] + 1;
            } else {
                dp[i][j] = dp[i - 1][j].max(dp[i][j - 1]);
            }
        }
    }

    let lcs_len = dp[n][m];
    let min_len = n.min(m);

    // If chars are too different within this group, highlight the whole group
    if min_len > 0 && lcs_len * 5 < min_len {
        old_spans.push(InlineSpan {
            start_col: old_offset,
            end_col: old_offset + n,
        });
        new_spans.push(InlineSpan {
            start_col: new_offset,
            end_col: new_offset + m,
        });
        return;
    }

    // Backtrack
    let mut old_in_lcs = vec![false; n];
    let mut new_in_lcs = vec![false; m];

    let mut i = n;
    let mut j = m;
    while i > 0 && j > 0 {
        if old_slice[i - 1] == new_slice[j - 1] {
            old_in_lcs[i - 1] = true;
            new_in_lcs[j - 1] = true;
            i -= 1;
            j -= 1;
        } else if dp[i - 1][j] >= dp[i][j - 1] {
            i -= 1;
        } else {
            j -= 1;
        }
    }

    let old_local = group_non_common_spans(&old_in_lcs);
    let new_local = group_non_common_spans(&new_in_lcs);

    // Count how many characters are highlighted on each side
    let old_highlighted: usize = old_local.iter().map(|s| s.end_col - s.start_col).sum();
    let new_highlighted: usize = new_local.iter().map(|s| s.end_col - s.start_col).sum();

    // If more than half of either side is highlighted, the char-level diff is too
    // noisy — just highlight the whole group on both sides for a clean result.
    let old_noisy = n > 0 && old_highlighted * 2 > n;
    let new_noisy = m > 0 && new_highlighted * 2 > m;
    if old_noisy || new_noisy {
        old_spans.push(InlineSpan {
            start_col: old_offset,
            end_col: old_offset + n,
        });
        new_spans.push(InlineSpan {
            start_col: new_offset,
            end_col: new_offset + m,
        });
        return;
    }

    // Char-level result is clean — use it with full-line offsets
    for span in old_local {
        old_spans.push(InlineSpan {
            start_col: old_offset + span.start_col,
            end_col: old_offset + span.end_col,
        });
    }
    for span in new_local {
        new_spans.push(InlineSpan {
            start_col: new_offset + span.start_col,
            end_col: new_offset + span.end_col,
        });
    }
}

/// Groups consecutive `false` entries into contiguous `InlineSpan` ranges.
fn group_non_common_spans(in_lcs: &[bool]) -> Vec<InlineSpan> {
    let mut spans = Vec::new();
    let mut i = 0;
    while i < in_lcs.len() {
        if !in_lcs[i] {
            let start = i;
            while i < in_lcs.len() && !in_lcs[i] {
                i += 1;
            }
            spans.push(InlineSpan {
                start_col: start,
                end_col: i,
            });
        } else {
            i += 1;
        }
    }
    spans
}

/// Remaps emphasis spans from full-line character coordinates to a wrapped
/// segment's local coordinates.
///
/// Spans that don't overlap the segment are excluded. Spans that partially
/// overlap are clipped to segment bounds and rebased to local coordinates.
pub fn map_emphasis_for_segment(
    spans: &[InlineSpan],
    segment_start: usize,
    segment_len: usize,
) -> Vec<InlineSpan> {
    if spans.is_empty() || segment_len == 0 {
        return Vec::new();
    }

    let segment_end = segment_start + segment_len;
    let mut result = Vec::new();

    for span in spans {
        if span.end_col <= segment_start || span.start_col >= segment_end {
            continue;
        }

        let new_start = span.start_col.saturating_sub(segment_start);
        let new_end = (span.end_col.saturating_sub(segment_start)).min(segment_len);

        if new_start < new_end {
            result.push(InlineSpan {
                start_col: new_start,
                end_col: new_end,
            });
        }
    }

    result
}

// Need this in scope for collect_changed_groups
#[derive(Debug, Clone, Copy, PartialEq)]
enum TokenState {
    Common,
    Changed,
}

#[cfg(test)]
mod tests {
    use super::*;

    // ── Basic edge cases ────────────────────────────────────────────

    #[test]
    fn identical_strings_return_empty_spans() {
        let (old, new) = compute_inline_spans("hello world", "hello world");
        assert!(old.is_empty());
        assert!(new.is_empty());
    }

    #[test]
    fn single_char_change() {
        // "hello" and "hallo" are each one word-token. Word-level diff sees them
        // as different, then char-level refinement pinpoints the 'e' vs 'a'.
        let (old, new) = compute_inline_spans("hello", "hallo");
        assert_eq!(
            old,
            vec![InlineSpan {
                start_col: 1,
                end_col: 2
            }]
        );
        assert_eq!(
            new,
            vec![InlineSpan {
                start_col: 1,
                end_col: 2
            }]
        );
    }

    #[test]
    fn suffix_change() {
        let (old, new) = compute_inline_spans("foo bar", "foo baz");
        // "bar" vs "baz" — word-level finds the changed words, char-level pinpoints "r" vs "z"
        assert_eq!(
            old,
            vec![InlineSpan {
                start_col: 6,
                end_col: 7
            }]
        );
        assert_eq!(
            new,
            vec![InlineSpan {
                start_col: 6,
                end_col: 7
            }]
        );
    }

    #[test]
    fn prefix_change() {
        let (old, new) = compute_inline_spans("abc def", "xyz def");
        // Whole word "abc" vs "xyz" — completely different chars, highlights the full word
        assert_eq!(
            old,
            vec![InlineSpan {
                start_col: 0,
                end_col: 3
            }]
        );
        assert_eq!(
            new,
            vec![InlineSpan {
                start_col: 0,
                end_col: 3
            }]
        );
    }

    #[test]
    fn completely_different_strings_return_empty_spans() {
        let (old, new) = compute_inline_spans("abcdef", "ghijkl");
        assert!(old.is_empty());
        assert!(new.is_empty());
    }

    #[test]
    fn one_empty_string_returns_empty_spans() {
        let (old, new) = compute_inline_spans("hello", "");
        assert!(old.is_empty());
        assert!(new.is_empty());

        let (old, new) = compute_inline_spans("", "hello");
        assert!(old.is_empty());
        assert!(new.is_empty());
    }

    #[test]
    fn both_empty_strings_return_empty_spans() {
        let (old, new) = compute_inline_spans("", "");
        assert!(old.is_empty());
        assert!(new.is_empty());
    }

    #[test]
    fn long_line_returns_empty_spans() {
        let long = "a".repeat(501);
        let (old, new) = compute_inline_spans(&long, "hello");
        assert!(old.is_empty());
        assert!(new.is_empty());

        let (old, new) = compute_inline_spans("hello", &long);
        assert!(old.is_empty());
        assert!(new.is_empty());
    }

    // ── map_emphasis_for_segment ────────────────────────────────────

    #[test]
    fn map_emphasis_span_fully_within_segment() {
        let spans = vec![InlineSpan {
            start_col: 2,
            end_col: 5,
        }];
        let result = map_emphasis_for_segment(&spans, 0, 10);
        assert_eq!(
            result,
            vec![InlineSpan {
                start_col: 2,
                end_col: 5
            }]
        );
    }

    #[test]
    fn map_emphasis_span_crossing_boundary_is_clipped() {
        let spans = vec![InlineSpan {
            start_col: 3,
            end_col: 8,
        }];
        let result = map_emphasis_for_segment(&spans, 5, 10);
        assert_eq!(
            result,
            vec![InlineSpan {
                start_col: 0,
                end_col: 3
            }]
        );
    }

    #[test]
    fn map_emphasis_span_fully_outside_is_excluded() {
        let spans = vec![InlineSpan {
            start_col: 0,
            end_col: 3,
        }];
        let result = map_emphasis_for_segment(&spans, 5, 10);
        assert!(result.is_empty());
    }

    #[test]
    fn map_emphasis_multiple_spans() {
        let spans = vec![
            InlineSpan {
                start_col: 0,
                end_col: 2,
            },
            InlineSpan {
                start_col: 3,
                end_col: 7,
            },
            InlineSpan {
                start_col: 8,
                end_col: 12,
            },
            InlineSpan {
                start_col: 13,
                end_col: 18,
            },
            InlineSpan {
                start_col: 20,
                end_col: 25,
            },
        ];
        let result = map_emphasis_for_segment(&spans, 5, 10);
        assert_eq!(
            result,
            vec![
                InlineSpan {
                    start_col: 0,
                    end_col: 2
                },
                InlineSpan {
                    start_col: 3,
                    end_col: 7
                },
                InlineSpan {
                    start_col: 8,
                    end_col: 10
                },
            ]
        );
    }

    #[test]
    fn map_emphasis_empty_spans() {
        let result = map_emphasis_for_segment(&[], 0, 10);
        assert!(result.is_empty());
    }

    #[test]
    fn map_emphasis_zero_segment_len() {
        let spans = vec![InlineSpan {
            start_col: 0,
            end_col: 5,
        }];
        let result = map_emphasis_for_segment(&spans, 0, 0);
        assert!(result.is_empty());
    }

    // ── Realistic code change scenarios ─────────────────────────────

    #[test]
    fn word_rename_multiply_to_mul() {
        // Word-level diff should highlight both whole words
        let (old, new) = compute_inline_spans(
            "    pub fn multiply(a: i32, b: i32) -> i32 {",
            "    pub fn mul(a: i32, b: i32) -> i32 {",
        );
        // "multiply" (col 11..19) vs "mul" (col 11..14) — both highlighted as changed words
        assert_eq!(
            old,
            vec![InlineSpan {
                start_col: 11,
                end_col: 19
            }]
        );
        assert_eq!(
            new,
            vec![InlineSpan {
                start_col: 11,
                end_col: 14
            }]
        );
    }

    #[test]
    fn variable_rename_clean_words() {
        // Two word renames: "result" -> "output", "calculate" -> "compute"
        let (old, new) = compute_inline_spans(
            "    let result = calculate_total(items);",
            "    let output = compute_total(items);",
        );
        // Should produce clean word-level spans, not scattered fragments
        // "result" is at col 8..14, "calculate_total" at col 17..32
        // "output" is at col 8..14, "compute_total" at col 17..30
        assert!(!old.is_empty());
        assert!(!new.is_empty());
        // Should be exactly 2 spans on each side (the two renamed words)
        // and no scattered single-char fragments
        for span in &old {
            assert!(
                span.end_col - span.start_col >= 2,
                "old span {:?} is too small — should be word-level",
                span
            );
        }
        for span in &new {
            assert!(
                span.end_col - span.start_col >= 2,
                "new span {:?} is too small — should be word-level",
                span
            );
        }
    }

    #[test]
    fn string_literal_change() {
        let (old, new) = compute_inline_spans(
            r#"    println!("Hello, world!");"#,
            r#"    println!("Goodbye, world!");"#,
        );
        // "Hello" vs "Goodbye" — word-level change within the string
        assert!(!old.is_empty());
        assert!(!new.is_empty());
    }

    #[test]
    fn type_change() {
        let (old, new) = compute_inline_spans(
            "    fn process(data: Vec<String>) -> Result<(), Error> {",
            "    fn process(data: Vec<&str>) -> Result<(), AppError> {",
        );
        // Two change sites: "String" -> "&str" and "Error" -> "AppError"
        assert!(!old.is_empty());
        assert!(!new.is_empty());
    }

    #[test]
    fn added_argument() {
        let (old, new) =
            compute_inline_spans("    let x = foo(a, b);", "    let x = foo(a, b, c);");
        // Only the new side should have a span for the added ", c"
        assert!(!new.is_empty());
    }

    #[test]
    fn numeric_constant_change() {
        let (old, new) = compute_inline_spans(
            "    const MAX_RETRIES: u32 = 3;",
            "    const MAX_RETRIES: u32 = 5;",
        );
        // Precisely highlights the single changed digit
        // "    const MAX_RETRIES: u32 = 3;" — '3' is at column 29
        assert_eq!(
            old,
            vec![InlineSpan {
                start_col: 29,
                end_col: 30
            }]
        );
        assert_eq!(
            new,
            vec![InlineSpan {
                start_col: 29,
                end_col: 30
            }]
        );
    }

    // ── Tokenizer tests ─────────────────────────────────────────────

    #[test]
    fn tokenize_splits_words_and_punctuation() {
        let chars: Vec<char> = "fn foo(x: i32)".chars().collect();
        let tokens = tokenize(&chars);
        let texts: Vec<String> = tokens.iter().map(|t| t.text.iter().collect()).collect();
        assert_eq!(
            texts,
            vec!["fn", " ", "foo", "(", "x", ":", " ", "i32", ")"]
        );
    }

    #[test]
    fn tokenize_handles_underscores_in_words() {
        let chars: Vec<char> = "my_var = 42".chars().collect();
        let tokens = tokenize(&chars);
        let texts: Vec<String> = tokens.iter().map(|t| t.text.iter().collect()).collect();
        assert_eq!(texts, vec!["my_var", " ", "=", " ", "42"]);
    }

    #[test]
    fn tokenize_preserves_offsets() {
        let chars: Vec<char> = "a + b".chars().collect();
        let tokens = tokenize(&chars);
        assert_eq!(tokens[0].offset, 0); // "a"
        assert_eq!(tokens[1].offset, 1); // " "
        assert_eq!(tokens[2].offset, 2); // "+"
        assert_eq!(tokens[3].offset, 3); // " "
        assert_eq!(tokens[4].offset, 4); // "b"
    }
}
