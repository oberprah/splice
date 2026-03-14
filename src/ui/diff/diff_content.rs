use crate::app::DiffView;
use crate::core::{DiffRef, UncommittedType};
use crate::domain::diff::DiffBlock;
use crate::domain::highlight::TokenSpan;
use crate::domain::wrap::{wrap_line, WrappedSegment};
use crate::ui::theme::{DiffColors, Theme};
use ratatui::{prelude::*, widgets::Paragraph};
use unicode_width::UnicodeWidthChar;

struct Layout {
    x: u16,
    width: u16,
    left_width: usize,
    right_width: usize,
}

struct Viewport {
    height: usize,
    content_start: usize,
}

struct RenderState {
    y: u16,
    rendered: usize,
    row_index: usize,
}

struct RenderContext {
    layout: Layout,
    viewport: Viewport,
    state: RenderState,
}

pub fn render_diff_view(f: &mut Frame, view: &DiffView, area: Rect, theme: &Theme) {
    let mut y = area.y;
    let width = area.width as usize;

    let range_display = diff_header_prefix(&view.diff_ref);

    let header = format!(
        "{} · {} · +{} -{}",
        range_display, view.file.info.path, view.file.info.additions, view.file.info.deletions
    );
    let header = pad_or_trim(&header, width);
    let header_widget = Paragraph::new(header).style(theme.text_muted);
    f.render_widget(header_widget, Rect::new(area.x, y, area.width, 1));
    y = y.saturating_add(1);

    let content_height = area.height.saturating_sub(y - area.y).saturating_sub(1) as usize;
    let content_area = Rect::new(area.x, y, area.width, content_height as u16);

    render_diff_lines(f, view, content_area, theme);

    let help = Paragraph::new("j/k: scroll  n/p: next/prev diff  o: open  q: back")
        .style(theme.text_muted)
        .alignment(Alignment::Left);
    let help_area = Rect::new(
        area.x,
        area.y + area.height.saturating_sub(1),
        area.width,
        1,
    );
    f.render_widget(help, help_area);
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

fn render_diff_lines(f: &mut Frame, view: &DiffView, area: Rect, theme: &Theme) {
    if view.file.blocks.is_empty() {
        let msg = Paragraph::new("No changes")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, area);
        return;
    }

    let width = area.width as usize;
    let separator = " │ ";
    let left_width = width.saturating_sub(separator.len()) / 2;
    let right_width = width
        .saturating_sub(separator.len())
        .saturating_sub(left_width);

    let focus_offset = view.viewport_height / 4;
    let ctx = &mut RenderContext {
        layout: Layout {
            x: area.x,
            width: area.width,
            left_width,
            right_width,
        },
        viewport: Viewport {
            height: area.height as usize,
            content_start: view.scroll_offset.saturating_sub(focus_offset),
        },
        state: RenderState {
            y: area.y,
            rendered: 0,
            row_index: 0,
        },
    };

    let focused_change_idx = view.focused_change_idx();
    let mut change_block_counter = 0usize;
    for block in &view.file.blocks {
        match block {
            DiffBlock::Unchanged(lines) => {
                if !render_unchanged_block(f, lines, ctx, theme) {
                    return;
                }
            }
            DiffBlock::Change { old, new } => {
                let idx = change_block_counter;
                change_block_counter += 1;
                if !render_change_block(f, old, new, ctx, theme, idx, focused_change_idx) {
                    return;
                }
            }
        }
    }
}

fn render_unchanged_block(
    f: &mut Frame,
    lines: &[crate::domain::diff::UnchangedLine],
    ctx: &mut RenderContext,
    theme: &Theme,
) -> bool {
    for line in lines {
        let left_rows = format_cell(
            line.old_number,
            ' ',
            &line.text,
            ctx.layout.left_width,
            Style::default(),
            Some(&line.tokens),
            theme,
        );
        let right_rows = format_cell(
            line.new_number,
            ' ',
            &line.text,
            ctx.layout.right_width,
            Style::default(),
            Some(&line.tokens),
            theme,
        );

        let max_rows = left_rows.len().max(right_rows.len());

        for row_idx in 0..max_rows {
            if ctx.state.row_index < ctx.viewport.content_start {
                ctx.state.row_index += 1;
                continue;
            }
            if ctx.state.rendered >= ctx.viewport.height {
                return false;
            }

            let left = left_rows
                .get(row_idx)
                .cloned()
                .unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.left_width))]);
            let right = right_rows
                .get(row_idx)
                .cloned()
                .unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.right_width))]);

            render_styled_row(
                f,
                ctx.state.y,
                left,
                right,
                ctx.layout.x,
                ctx.layout.width,
                theme,
            );

            ctx.state.y = ctx.state.y.saturating_add(1);
            ctx.state.rendered += 1;
            ctx.state.row_index += 1;
        }
    }

    true
}

fn render_change_block(
    f: &mut Frame,
    old_lines: &[crate::domain::diff::DiffLine],
    new_lines: &[crate::domain::diff::DiffLine],
    ctx: &mut RenderContext,
    theme: &Theme,
    change_block_idx: usize,
    focused_idx: Option<usize>,
) -> bool {
    let is_focused = focused_idx == Some(change_block_idx);
    let max_len = old_lines.len().max(new_lines.len());

    for i in 0..max_len {
        let left_rows = old_lines.get(i).map(|line| {
            let has_new_line = new_lines.get(i).is_some();
            let base_colors = if has_new_line {
                &theme.diff_changed
            } else {
                &theme.diff_removed
            };
            let colors = if is_focused {
                use crate::ui::theme::DiffColors;
                &DiffColors {
                    bg: base_colors.bg_bright,
                    bg_bright: base_colors.bg_bright,
                    fg: base_colors.fg,
                }
            } else {
                base_colors
            };
            format_cell_styled(
                line.number,
                '-',
                &line.text,
                ctx.layout.left_width,
                colors,
                Some(&line.tokens),
                theme,
            )
        });

        let right_rows = new_lines.get(i).map(|line| {
            let has_old_line = old_lines.get(i).is_some();
            let base_colors = if has_old_line {
                &theme.diff_changed
            } else {
                &theme.diff_added
            };
            let colors = if is_focused {
                use crate::ui::theme::DiffColors;
                &DiffColors {
                    bg: base_colors.bg_bright,
                    bg_bright: base_colors.bg_bright,
                    fg: base_colors.fg,
                }
            } else {
                base_colors
            };
            format_cell_styled(
                line.number,
                '+',
                &line.text,
                ctx.layout.right_width,
                colors,
                Some(&line.tokens),
                theme,
            )
        });

        let left_rows = left_rows.unwrap_or_default();
        let right_rows = right_rows.unwrap_or_default();
        let max_rows = left_rows.len().max(right_rows.len());

        for row_idx in 0..max_rows {
            if ctx.state.row_index < ctx.viewport.content_start {
                ctx.state.row_index += 1;
                continue;
            }
            if ctx.state.rendered >= ctx.viewport.height {
                return false;
            }

            let left = left_rows
                .get(row_idx)
                .cloned()
                .unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.left_width))]);
            let right = right_rows
                .get(row_idx)
                .cloned()
                .unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.right_width))]);

            render_styled_row(
                f,
                ctx.state.y,
                left,
                right,
                ctx.layout.x,
                ctx.layout.width,
                theme,
            );

            ctx.state.y = ctx.state.y.saturating_add(1);
            ctx.state.rendered += 1;
            ctx.state.row_index += 1;
        }
    }

    true
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

fn format_cell_styled(
    line_num: u32,
    sign: char,
    text: &str,
    width: usize,
    colors: &DiffColors,
    tokens: Option<&[TokenSpan]>,
    theme: &Theme,
) -> Vec<Vec<Span<'static>>> {
    format_cell(
        line_num,
        sign,
        text,
        width,
        Style::new().bg(colors.bg).fg(colors.fg),
        tokens,
        theme,
    )
}

fn format_cell(
    line_num: u32,
    sign: char,
    text: &str,
    width: usize,
    style: Style,
    tokens: Option<&[TokenSpan]>,
    theme: &Theme,
) -> Vec<Vec<Span<'static>>> {
    let line_num_str = format!("{:>3} ", line_num);
    let sign_str = sign.to_string();
    let prefix_width = line_num_str.chars().count() + sign_str.chars().count();
    let content_width = width.saturating_sub(prefix_width);

    let segments = wrap_line(text, tokens.unwrap_or(&[]), content_width);

    let mut rows = Vec::new();
    for (i, segment) in segments.into_iter().enumerate() {
        let mut spans = Vec::new();

        if i == 0 {
            spans.push(Span::styled(line_num_str.clone(), theme.diff_line_number));
        } else {
            spans.push(Span::styled("  ↪ ", theme.diff_line_number));
        }
        spans.push(Span::styled(sign_str.clone(), style));
        spans.extend(render_wrapped_segment(
            &segment,
            content_width,
            style,
            theme,
        ));

        rows.push(spans);
    }

    rows
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

fn render_styled_row(
    f: &mut Frame,
    y: u16,
    left: Vec<Span<'_>>,
    right: Vec<Span<'_>>,
    x: u16,
    width: u16,
    theme: &Theme,
) {
    let mut spans: Vec<Span<'_>> = left;
    spans.push(Span::styled(" │ ", theme.diff_divider));
    spans.extend(right);

    let line = Line::from(spans);
    let para = Paragraph::new(line);
    let area = Rect::new(x, y, width, 1);
    f.render_widget(para, area);
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::domain::highlight::HighlightKind;

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
}
