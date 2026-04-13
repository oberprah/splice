use crate::app::viewport::VisibleRows;
use crate::app::{DiffLayout, DiffView};
use crate::core::{DiffRef, UncommittedType};
use crate::domain::diff::inline_diff::InlineSpan;
use crate::domain::diff::layout::{Cell, CellKind, UnifiedRow};
use crate::domain::wrap::WrappedSegment;
use crate::ui::theme::Theme;
use ratatui::{prelude::*, widgets::Paragraph};
use unicode_width::UnicodeWidthChar;

/// Fixed rows used by the diff view chrome (header + footer).
pub const DIFF_CHROME_ROWS: u16 = 2;

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

    let content_height = view.viewport_height();

    render_diff_rows(f, view, area.x, y, area.width, content_height, theme);

    let layout_hint = match view.layout() {
        DiffLayout::SideBySide => "v: unified",
        DiffLayout::Unified => "v: split",
    };
    let help = Paragraph::new(format!(
        "j/k: scroll  n/p: next/prev diff  {layout_hint}  y: copy path  o: open  q: back"
    ))
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
    let content = view.visible_content();

    let is_empty = match &content.rows {
        VisibleRows::SideBySide(rows) => rows.is_empty(),
        VisibleRows::Unified { rows, .. } => rows.is_empty(),
    };

    if is_empty {
        let msg = Paragraph::new("No changes")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, Rect::new(x, y, width, content_height as u16));
        return;
    }

    match content.rows {
        VisibleRows::SideBySide(rows) => {
            render_side_by_side_rows(
                f,
                rows,
                content.row_offset,
                content.active_hunk_range,
                x,
                y,
                width,
                content_height,
                theme,
            );
        }
        VisibleRows::Unified { rows, prefix_width } => {
            render_unified_rows(
                f,
                rows,
                content.row_offset,
                content.active_hunk_range,
                x,
                y,
                width,
                content_height,
                prefix_width,
                theme,
            );
        }
    }
}

fn render_side_by_side_rows(
    f: &mut Frame,
    rows: &[crate::domain::diff::ScreenRow],
    row_offset: usize,
    active_hunk_range: Option<crate::domain::diff::HunkRange>,
    x: u16,
    y: u16,
    width: u16,
    content_height: usize,
    theme: &Theme,
) {
    let total_width = width as usize;
    let separator = " \u{2502} ";
    let left_width = total_width.saturating_sub(separator.len()) / 2;
    let right_width = total_width
        .saturating_sub(separator.len())
        .saturating_sub(left_width);

    for (i, row) in rows.iter().enumerate() {
        if i >= content_height {
            break;
        }
        let abs_row = row_offset + i;
        let in_active_hunk =
            active_hunk_range.is_some_and(|r| abs_row >= r.start && abs_row < r.end);

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

fn render_unified_rows(
    f: &mut Frame,
    rows: &[UnifiedRow],
    row_offset: usize,
    active_hunk_range: Option<crate::domain::diff::HunkRange>,
    x: u16,
    y: u16,
    width: u16,
    content_height: usize,
    prefix_width: usize,
    theme: &Theme,
) {
    let total_width = width as usize;
    let single_gutter = prefix_width / 2;

    for (i, row) in rows.iter().enumerate() {
        if i >= content_height {
            break;
        }
        let abs_row = row_offset + i;
        let in_active_hunk =
            active_hunk_range.is_some_and(|r| abs_row >= r.start && abs_row < r.end);

        let is_continuation = row.old_line_number.is_none() && row.new_line_number.is_none();

        let prefix = if is_continuation {
            // (prefix_width - 2) spaces + ↪ + space
            format!("{}\u{21aa} ", " ".repeat(prefix_width.saturating_sub(2)))
        } else {
            let blank_gutter = " ".repeat(single_gutter);
            let old_str = match row.old_line_number {
                Some(n) => format!("{:>width$} ", n, width = single_gutter - 1),
                None => blank_gutter.clone(),
            };
            let new_str = match row.new_line_number {
                Some(n) => format!("{:>width$} ", n, width = single_gutter - 1),
                None => blank_gutter,
            };
            format!("{old_str}{new_str}")
        };

        let content_width = total_width.saturating_sub(prefix_width);

        let (base_style, emphasis_bg) =
            unified_cell_style(&row.cell, row.is_old_side, in_active_hunk, theme);

        let mut spans: Vec<Span<'static>> = vec![Span::styled(prefix, theme.diff_line_number)];

        if row.cell.kind == CellKind::Empty {
            spans = vec![Span::raw(" ".repeat(total_width))];
        } else if row.cell.text.is_empty() {
            // Empty text with a valid prefix (e.g. spacer row)
            spans.push(Span::styled(" ".repeat(content_width), base_style));
        } else {
            let segment = WrappedSegment {
                text: row.cell.text.clone(),
                char_offset: 0,
                tokens: row.cell.tokens.clone(),
            };
            spans.extend(render_wrapped_segment(
                &segment,
                content_width,
                base_style,
                &row.cell.emphasis,
                emphasis_bg,
                theme,
            ));
        }

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
            let (base_style, emphasis_bg) = cell_style(cell, in_active_hunk, is_left, theme);

            // Spacer cell in a modification block: keep line-number area
            // unstyled so the background doesn't bleed into it.
            if cell.line_number.is_none() && cell.text.is_empty() {
                let prefix_width = 4; // matches "{:>3} " / "  ↪ "
                let content_width = width.saturating_sub(prefix_width);
                return vec![
                    Span::raw(" ".repeat(prefix_width)),
                    Span::styled(" ".repeat(content_width), base_style),
                ];
            }

            let line_num_str = match cell.line_number {
                Some(n) => format!("{:>3} ", n),
                None => "  \u{21aa} ".to_string(), // "  ↪ "
            };
            let prefix_width = line_num_str.chars().count();
            let content_width = width.saturating_sub(prefix_width);

            let mut spans: Vec<Span<'static>> =
                vec![Span::styled(line_num_str, theme.diff_line_number)];

            let segment = WrappedSegment {
                text: cell.text.clone(),
                char_offset: 0,
                tokens: cell.tokens.clone(),
            };
            spans.extend(render_wrapped_segment(
                &segment,
                content_width,
                base_style,
                &cell.emphasis,
                emphasis_bg,
                theme,
            ));

            spans
        }
    }
}

fn cell_style(
    cell: &Cell,
    in_active_hunk: bool,
    _is_left: bool,
    theme: &Theme,
) -> (Style, Option<Color>) {
    match cell.kind {
        CellKind::Context => (Style::default(), None),
        CellKind::Removed => {
            let colors = &theme.diff_removed;
            let bg = if in_active_hunk {
                colors.bg_bright
            } else {
                colors.bg
            };
            (Style::new().bg(bg).fg(colors.fg), None)
        }
        CellKind::Added => {
            let colors = &theme.diff_added;
            let bg = if in_active_hunk {
                colors.bg_bright
            } else {
                colors.bg
            };
            (Style::new().bg(bg).fg(colors.fg), None)
        }
        CellKind::Changed => {
            let colors = &theme.diff_changed;
            let (bg, emph_bg) = if in_active_hunk {
                (colors.bg_bright, colors.bg_bright_emphasis)
            } else {
                (colors.bg, colors.bg_emphasis)
            };
            (Style::new().bg(bg).fg(colors.fg), Some(emph_bg))
        }
        CellKind::Empty => (Style::default(), None),
    }
}

/// In unified mode, `Changed` cells should use red/green based on which side
/// the row belongs to, rather than the blue "changed" color that works in
/// side-by-side where both sides are visible simultaneously.
fn unified_cell_style(
    cell: &Cell,
    is_old_side: bool,
    in_active_hunk: bool,
    theme: &Theme,
) -> (Style, Option<Color>) {
    if cell.kind == CellKind::Changed {
        let colors = if is_old_side {
            &theme.diff_removed
        } else {
            &theme.diff_added
        };
        let (bg, emph_bg) = if in_active_hunk {
            (colors.bg_bright, colors.bg_bright_emphasis)
        } else {
            (colors.bg, colors.bg_emphasis)
        };
        (Style::new().bg(bg).fg(colors.fg), Some(emph_bg))
    } else {
        cell_style(cell, in_active_hunk, true, theme)
    }
}

fn render_wrapped_segment(
    segment: &WrappedSegment,
    content_width: usize,
    base_style: Style,
    emphasis: &[InlineSpan],
    emphasis_bg: Option<Color>,
    theme: &Theme,
) -> Vec<Span<'static>> {
    if content_width == 0 {
        return Vec::new();
    }

    let mut chars: Vec<char> = Vec::new();
    let mut cols_used = 0usize;
    for ch in segment.text.chars() {
        let ch_width = UnicodeWidthChar::width(ch).unwrap_or(1);
        if cols_used + ch_width > content_width {
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

    let mut char_emphasis = vec![false; visible];
    for span in emphasis {
        let start = span.start_col.min(visible);
        let end = span.end_col.min(visible);
        for em in char_emphasis.iter_mut().take(end).skip(start) {
            *em = true;
        }
    }

    let mut spans: Vec<Span<'static>> = Vec::new();
    if visible > 0 {
        let first_base = if char_emphasis[0] {
            if let Some(emph_bg) = emphasis_bg {
                base_style.bg(emph_bg)
            } else {
                base_style
            }
        } else {
            base_style
        };
        let mut run = String::new();
        let mut run_style = char_style(first_base, char_kinds[0], theme);
        for idx in 0..visible {
            let em_base = if char_emphasis[idx] {
                if let Some(emph_bg) = emphasis_bg {
                    base_style.bg(emph_bg)
                } else {
                    base_style
                }
            } else {
                base_style
            };
            let current_style = char_style(em_base, char_kinds[idx], theme);
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

    if cols_used < content_width {
        spans.push(Span::styled(
            " ".repeat(content_width - cols_used),
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
        let spans = render_wrapped_segment(&segment, 6, Style::default(), &[], None, &theme);
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

        let spans = render_wrapped_segment(&segment, 10, Style::default(), &[], None, &theme);

        assert!(
            spans.iter().any(|span| {
                span.style.fg == Some(theme.syntax.keyword) && span.content.contains("fn")
            }),
            "Expected a keyword-colored span for `fn`"
        );
    }

    #[test]
    fn spacer_cell_does_not_color_line_number_area() {
        let theme = Theme::dark();
        // Spacer cell: Changed kind, no line number, empty text.
        // This is what the layout layer produces for the shorter side
        // of an asymmetric modification block.
        let cell = Cell {
            kind: CellKind::Changed,
            line_number: None,
            text: String::new(),
            tokens: vec![],
            emphasis: vec![],
        };
        let width = 14; // 4 prefix + 10 content
        let spans = render_cell(&cell, width, true, true, &theme);

        // The line number area (first 4 chars) should NOT have the
        // changed background color. Only the content area should.
        assert!(
            spans.len() >= 2,
            "Expected separate spans for line-number area and content, got {} span(s): {:?}",
            spans.len(),
            spans.iter().map(|s| s.content.as_ref()).collect::<Vec<_>>()
        );

        // First span (line number area) should have default/no background
        let line_num_span = &spans[0];
        assert_eq!(
            line_num_span.style.bg, None,
            "Line number area should have no background color, got {:?}",
            line_num_span.style.bg
        );
    }

    #[test]
    fn render_cell_shows_all_wrapped_text() {
        let theme = Theme::dark();
        // Text is exactly 10 chars
        let text = "abcdefghij";
        let cell = Cell {
            kind: CellKind::Context,
            line_number: Some(1),
            text: text.to_string(),
            tokens: vec![],
            emphasis: vec![],
        };
        // prefix = 4 chars ("  1 "), so content_width = width - 4 = 10
        // The layout layer wraps text at content_width=10, so the cell can contain up to 10 chars.
        // The renderer should display all 10 chars.
        let width = 14; // 4 prefix + 10 content
        let spans = render_cell(&cell, width, false, true, &theme);
        let rendered: String = spans.iter().map(|s| s.content.as_ref()).collect();
        // The rendered output should contain all characters of the original text
        assert!(
            rendered.contains(text),
            "Expected rendered output to contain '{}', but got '{}'",
            text,
            rendered
        );
    }
}
