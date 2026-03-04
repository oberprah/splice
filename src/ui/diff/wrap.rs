use crate::domain::highlight::TokenSpan;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct WrappedSegment {
    pub text: String,
    pub char_offset: usize,
    pub tokens: Vec<TokenSpan>,
}

pub fn wrap_line(text: &str, tokens: &[TokenSpan], width: usize) -> Vec<WrappedSegment> {
    if text.is_empty() || width == 0 {
        return vec![WrappedSegment {
            text: String::new(),
            char_offset: 0,
            tokens: Vec::new(),
        }];
    }

    let options =
        textwrap::Options::new(width).word_splitter(textwrap::WordSplitter::NoHyphenation);

    let wrapped = textwrap::wrap(text, &options);

    if wrapped.is_empty() {
        return vec![WrappedSegment {
            text: String::new(),
            char_offset: 0,
            tokens: Vec::new(),
        }];
    }

    let mut segments = Vec::with_capacity(wrapped.len());
    let mut char_offset = 0;

    for wrapped_line in wrapped {
        let wrapped_text = wrapped_line.to_string();
        let wrapped_len = wrapped_text.chars().count();

        let segment_tokens = map_tokens_for_segment(tokens, char_offset, wrapped_len);

        segments.push(WrappedSegment {
            text: wrapped_text,
            char_offset,
            tokens: segment_tokens,
        });

        char_offset += wrapped_len;

        if char_offset < text.chars().count() {
            char_offset += 1;
        }
    }

    segments
}

fn map_tokens_for_segment(
    tokens: &[TokenSpan],
    segment_start: usize,
    segment_len: usize,
) -> Vec<TokenSpan> {
    if tokens.is_empty() || segment_len == 0 {
        return Vec::new();
    }

    let segment_end = segment_start + segment_len;
    let mut segment_tokens = Vec::new();

    for token in tokens {
        if token.end_col <= segment_start || token.start_col >= segment_end {
            continue;
        }

        let new_start = token.start_col.saturating_sub(segment_start);
        let new_end = (token.end_col.saturating_sub(segment_start)).min(segment_len);

        if new_start < new_end {
            segment_tokens.push(TokenSpan {
                start_col: new_start,
                end_col: new_end,
                kind: token.kind,
            });
        }
    }

    segment_tokens
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::highlight::HighlightKind;

    #[test]
    fn wrap_empty_text_returns_single_empty_segment() {
        let tokens: Vec<TokenSpan> = Vec::new();
        let result = wrap_line("", &tokens, 10);

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].text, "");
        assert_eq!(result[0].char_offset, 0);
    }

    #[test]
    fn wrap_with_zero_width_returns_single_empty_segment() {
        let result = wrap_line("hello world", &[], 0);

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].text, "");
    }

    #[test]
    fn wrap_short_text_returns_single_segment() {
        let result = wrap_line("hello", &[], 20);

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].text, "hello");
        assert_eq!(result[0].char_offset, 0);
    }

    #[test]
    fn wrap_long_text_splits_into_segments() {
        let result = wrap_line("hello world this is a test", &[], 10);

        assert!(result.len() > 1);
        assert_eq!(result[0].char_offset, 0);
    }

    #[test]
    fn map_tokens_correctly_adjusts_offsets() {
        let tokens = vec![
            TokenSpan {
                start_col: 0,
                end_col: 5,
                kind: HighlightKind::Keyword,
            },
            TokenSpan {
                start_col: 6,
                end_col: 11,
                kind: HighlightKind::String,
            },
        ];

        let result = wrap_line("hello world", &tokens, 6);

        assert!(result.len() >= 1);

        if result.len() >= 1 {
            assert!(result[0]
                .tokens
                .iter()
                .any(|t| t.kind == HighlightKind::Keyword));
        }

        if result.len() >= 2 {
            assert!(result[1]
                .tokens
                .iter()
                .any(|t| t.kind == HighlightKind::String));
        }
    }

    #[test]
    fn tokens_clipped_to_segment_bounds() {
        let tokens = vec![TokenSpan {
            start_col: 3,
            end_col: 8,
            kind: HighlightKind::Keyword,
        }];

        let result = map_tokens_for_segment(&tokens, 0, 5);

        assert_eq!(result.len(), 1);
        assert_eq!(result[0].start_col, 3);
        assert_eq!(result[0].end_col, 5);
    }
}
