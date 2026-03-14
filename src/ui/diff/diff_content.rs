use crate::app::DiffView;
use crate::core::{DiffRef, UncommittedType};
use crate::domain::diff::layout::{Cell, CellKind};
use crate::domain::wrap::WrappedSegment;
use crate::ui::theme::Theme;
use ratatui::{prelude::*, widgets::Paragraph};
use unicode_width::UnicodeWidthChar;

pub fn render_diff_view(f: &mut Frame, view: &DiffView, area: Rect, theme: &Theme) {
    let mut y = area.y;
    let width = area.width as usize;

    let header = format!(
        "{} · {} · +{} -{}",
        diff_header_prefix(&view.diff_ref),
        view.file.info.path,
        view.file.info.additions,
        view.file.info.deletions
    );
    let header = pad_or_trim(&header, width);
    f.render_widget(
        Paragraph::new(header).style(theme.text_muted),
        Rect::new(area.x, y, area.width, 1),
    );
    y = y.saturating_add(1);

    let content_height = area.height.saturating_sub(y - area.y).saturating_sub(1) as usize;

    render_diff_rows(f, view, area.x, y, area.width, content_height, theme);

    let help = Paragraph::new("j/k: scroll  n/p: next/prev diff  o: open  q: back")
        .style(theme.text_muted)
        .alignment(Alignment::Left);
    f.render_widget(
        help,
        Rect::new(
            area.x,
            area.y + area.height.saturating_sub(1),
            area.width,
            1,
        ),
    );
}

fn render_diff_rows(
    f: &mut Frame,
    view: &DiffView,
    x: u16,
    y: u16,
    width: u16,
    content_height: usize,
    theme: &Theme,
) {
    let total_width = width as usize;
    let separator = " │ ";
    let left_width = total_width.saturating_sub(separator.len()) / 2;
    let right_width = total_width
        .saturating_sub(separator.len())
        .saturating_sub(left_width);

    let visible = view.visible_rows();
    let row_offset = view.visible_row_offset();

    if visible.is_empty() {
        let msg = Paragraph::new("No changes")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, Rect::new(x, y, width, content_height as u16));
        return;
    }

    for (i, row) in visible.iter().enumerate() {
        if i >= content_height {
            break;
        }
        let abs_row = row_offset + i;
        let in_active_hunk = view.is_row_in_active_hunk(abs_row);

        let left_spans = render_cell(&row.left, left_width, in_active_hunk, true, theme);
        let right_spans = render_cell(&row.right, right_width, in_active_hunk, false, theme);

        let mut spans: Vec<Span> = left_spans;
        spans.push(Span::styled(separator, theme.diff_divider));
        spans.extend(right_spans);

        f.render_widget(
            Paragraph::new(Line::from(spans)),
            Rect::new(x, y + i as u16, width, 1),
        );
    }
}

fn render_cell(
    cell: &Cell,
    width: usize,
    in_active_hunk: bool,
    is_left: bool,
    theme: &Theme,
) -> Vec<Span<'static>> {
    match cell.kind {
        CellKind::Empty => vec![Span::raw(blank_cell(width))],
        _ => {
            let (base_style, sign) = cell_style_and_sign(cell, in_active_hunk, is_left, theme);

            let line_num_str = match cell.line_number {
                Some(n) => format!("{:>3} ", n),
                None => "  \u{21aa} ".to_string(), // "  ↪ "
            };
            let sign_str = sign.to_string();
            let prefix_width = line_num_str.chars().count() + sign_str.chars().count();
            let content_width = width.saturating_sub(prefix_width);

            let mut spans: Vec<Span<'static>> = vec![
                Span::styled(line_num_str, theme.diff_line_number),
                Span::styled(sign_str, base_style),
            ];

            let segment = WrappedSegment {
                text: cell.text.clone(),
                char_offset: 0,
                tokens: cell.tokens.clone(),
            };
            spans.extend(render_text_with_tokens(
                &segment,
                content_width,
                base_style,
                theme,
            ));

            spans
        }
    }
}

fn cell_style_and_sign(
    cell: &Cell,
    in_active_hunk: bool,
    is_left: bool,
    theme: &Theme,
) -> (Style, char) {
    match cell.kind {
        CellKind::Context => (Style::default(), ' '),
        CellKind::Removed => {
            let colors = &theme.diff_removed;
            let bg = if in_active_hunk {
                colors.bg_bright
            } else {
                colors.bg
            };
            (Style::new().bg(bg).fg(colors.fg), '-')
        }
        CellKind::Added => {
            let colors = &theme.diff_added;
            let bg = if in_active_hunk {
                colors.bg_bright
            } else {
                colors.bg
            };
            (Style::new().bg(bg).fg(colors.fg), '+')
        }
        CellKind::Changed => {
            // Both old and new line exist at this pair index — use changed (blue) color
            let colors = &theme.diff_changed;
            let bg = if in_active_hunk {
                colors.bg_bright
            } else {
                colors.bg
            };
            // Left side is the old (removed) half, right side is the new (added) half
            let sign = if is_left { '-' } else { '+' };
            (Style::new().bg(bg).fg(colors.fg), sign)
        }
        CellKind::Empty => (Style::default(), ' '),
    }
}

fn render_text_with_tokens(
    segment: &WrappedSegment,
    content_width: usize,
    base_style: Style,
    theme: &Theme,
) -> Vec<Span<'static>> {
    render_wrapped_segment(segment, content_width, base_style, theme)
}

fn render_wrapped_segment(
    segment: &WrappedSegment,
    content_width: usize,
    base_style: Style,
    theme: &Theme,
) -> Vec<Span<'static>> {
    if content_width == 0 {
        return Vec::new();
    }

    let max_content_cols = content_width.saturating_sub(1);
    let mut chars: Vec<char> = Vec::new();
    let mut cols_used = 0usize;
    for ch in segment.text.chars() {
        let ch_width = UnicodeWidthChar::width(ch).unwrap_or(1);
        if cols_used + ch_width > max_content_cols {
            break;
        }
        cols_used += ch_width;
        chars.push(ch);
    }

    let visible = chars.len();
    let mut char_kinds = vec![None; visible];

    for token in &segment.tokens {
        let start = token.start_col.min(visible);
        let end = token.end_col.min(visible);
        for kind in char_kinds.iter_mut().take(end).skip(start) {
            *kind = Some(token.kind);
        }
    }

    let mut spans = vec![Span::styled(" ", base_style)];
    if visible > 0 {
        let mut run = String::new();
        let mut run_style = char_style(base_style, char_kinds[0], theme);
        for idx in 0..visible {
            let current_style = char_style(base_style, char_kinds[idx], theme);
            if current_style != run_style && !run.is_empty() {
                spans.push(Span::styled(std::mem::take(&mut run), run_style));
                run_style = current_style;
            }
            run.push(chars[idx]);
        }
        if !run.is_empty() {
            spans.push(Span::styled(run, run_style));
        }
    }

    let used_cols = 1 + cols_used;
    if used_cols < content_width {
        spans.push(Span::styled(
            " ".repeat(content_width - used_cols),
            base_style,
        ));
    }

    spans
}

fn char_style(
    base_style: Style,
    token_kind: Option<crate::domain::highlight::HighlightKind>,
    theme: &Theme,
) -> Style {
    match token_kind {
        Some(kind) => base_style.fg(theme.syntax_color(kind)),
        None => base_style,
    }
}

fn blank_cell(width: usize) -> String {
    " ".repeat(width)
}

fn pad_or_trim(input: &str, width: usize) -> String {
    let mut chars: Vec<char> = input.chars().collect();
    if chars.len() > width {
        chars.truncate(width);
        return chars.into_iter().collect();
    }
    let mut result: String = chars.into_iter().collect();
    let result_len = result.chars().count();
    if result_len < width {
        result.push_str(&" ".repeat(width - result_len));
    }
    result
}

fn diff_header_prefix(diff_ref: &DiffRef) -> String {
    match diff_ref {
        DiffRef::CommitRange(range) => {
            if range.is_single_commit() {
                range.end.short_hash().to_string()
            } else {
                format!("{}..{}", range.end.short_hash(), range.start.short_hash())
            }
        }
        DiffRef::Uncommitted(uncommitted_type) => match uncommitted_type {
            UncommittedType::Unstaged => "Unstaged changes".to_string(),
            UncommittedType::Staged => "Staged changes".to_string(),
            UncommittedType::All => "Uncommitted changes".to_string(),
        },
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::highlight::{HighlightKind, TokenSpan};

    #[test]
    fn render_wrapped_segment_cjk_does_not_exceed_content_width() {
        let theme = crate::ui::theme::Theme::dark();
        let segment = crate::domain::wrap::WrappedSegment {
            text: "你好世界".to_string(),
            char_offset: 0,
            tokens: vec![],
        };
        let spans = render_wrapped_segment(&segment, 6, Style::default(), &theme);
        let total_display_width: usize = spans
            .iter()
            .map(|s| unicode_width::UnicodeWidthStr::width(s.content.as_ref()))
            .sum();
        assert!(
            total_display_width <= 6,
            "display width {total_display_width} exceeds content_width 6"
        );
    }

    #[test]
    fn render_wrapped_segment_applies_syntax_foreground() {
        let theme = Theme::dark();
        let tokens = vec![TokenSpan {
            start_col: 0,
            end_col: 2,
            kind: HighlightKind::Keyword,
        }];

        let segment = WrappedSegment {
            text: "fn main".to_string(),
            char_offset: 0,
            tokens,
        };

        let spans = render_wrapped_segment(&segment, 10, Style::default(), &theme);

        assert!(
            spans.iter().any(|span| {
                span.style.fg == Some(theme.syntax.keyword) && span.content.contains("fn")
            }),
            "Expected a keyword-colored span for `fn`"
        );
    }

    #[test]
    fn render_text_with_tokens_delegates_to_wrapped_segment() {
        let theme = Theme::dark();
        let tokens = vec![TokenSpan {
            start_col: 0,
            end_col: 2,
            kind: HighlightKind::Keyword,
        }];
        let segment = WrappedSegment {
            text: "fn foo".to_string(),
            char_offset: 0,
            tokens,
        };
        let spans = render_text_with_tokens(&segment, 10, Style::default(), &theme);
        assert!(
            spans.iter().any(|span| span.content.contains("fn")),
            "Expected fn to appear in spans"
        );
    }
}
