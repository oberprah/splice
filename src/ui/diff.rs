use crate::app::DiffView;
use crate::core::{DiffSource, UncommittedType};
use crate::domain::diff::{ChangeBlock, DiffBlock, UnchangedBlock};
use crate::domain::highlight::TokenSpan;
use crate::ui::theme::{DiffColors, Theme};
use ratatui::{prelude::*, widgets::Paragraph};

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

pub fn render_diff_view(f: &mut Frame, diff: &DiffView, area: Rect, theme: &Theme) {
    let mut y = area.y;
    let width = area.width as usize;

    let range_display = diff_header_prefix(&diff.source);

    let header = format!(
        "{} · {} · +{} -{}",
        range_display, diff.file.path, diff.file.additions, diff.file.deletions
    );
    let header = truncate_to_width(&header, width);
    let header_widget = Paragraph::new(header).style(Style::default().fg(Color::Gray));
    f.render_widget(header_widget, Rect::new(area.x, y, area.width, 1));
    y = y.saturating_add(1);

    let empty = Paragraph::new("");
    f.render_widget(empty, Rect::new(area.x, y, area.width, 1));
    y = y.saturating_add(1);

    let content_height = area.height.saturating_sub(y - area.y).saturating_sub(1) as usize;
    let content_area = Rect::new(area.x, y, area.width, content_height as u16);

    render_diff_lines(f, diff, content_area, theme);

    let help = Paragraph::new("j/k: scroll  n/p: next/prev diff  o: open  q: back")
        .style(Style::default().fg(Color::DarkGray))
        .alignment(Alignment::Left);
    let help_area = Rect::new(
        area.x,
        area.y + area.height.saturating_sub(1),
        area.width,
        1,
    );
    f.render_widget(help, help_area);
}

fn diff_header_prefix(source: &DiffSource) -> String {
    match source {
        DiffSource::CommitRange(range) => {
            if range.is_single_commit() {
                range.end.short_hash().to_string()
            } else {
                format!("{}..{}", range.end.short_hash(), range.start.short_hash())
            }
        }
        DiffSource::Uncommitted(uncommitted_type) => match uncommitted_type {
            UncommittedType::Unstaged => "Unstaged changes".to_string(),
            UncommittedType::Staged => "Staged changes".to_string(),
            UncommittedType::All => "Uncommitted changes".to_string(),
        },
    }
}

fn render_diff_lines(f: &mut Frame, diff: &DiffView, area: Rect, theme: &Theme) {
    if diff.diff.blocks.is_empty() {
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

    let ctx = &mut RenderContext {
        layout: Layout {
            x: area.x,
            width: area.width,
            left_width,
            right_width,
        },
        viewport: Viewport {
            height: area.height as usize,
            content_start: diff.scroll_offset.saturating_sub(diff.viewport_height / 4),
        },
        state: RenderState {
            y: area.y,
            rendered: 0,
            row_index: 0,
        },
    };

    let focus_offset = diff.viewport_height / 4;
    let top_padding = focus_offset.saturating_sub(diff.scroll_offset);
    while ctx.state.rendered < ctx.viewport.height && ctx.state.rendered < top_padding {
        render_row(
            f,
            ctx.state.y,
            blank_cell(ctx.layout.left_width),
            blank_cell(ctx.layout.right_width),
            ctx.layout.x,
            ctx.layout.width,
        );
        ctx.state.y = ctx.state.y.saturating_add(1);
        ctx.state.rendered += 1;
    }

    for block in &diff.diff.blocks {
        match block {
            DiffBlock::Unchanged(unchanged) => {
                if !render_unchanged_block(f, unchanged, ctx, diff, theme) {
                    return;
                }
            }
            DiffBlock::Change(change) => {
                if !render_change_block(f, change, ctx, diff, theme) {
                    return;
                }
            }
        }
    }

    while ctx.state.rendered < ctx.viewport.height {
        render_row(
            f,
            ctx.state.y,
            blank_cell(ctx.layout.left_width),
            blank_cell(ctx.layout.right_width),
            ctx.layout.x,
            ctx.layout.width,
        );
        ctx.state.y = ctx.state.y.saturating_add(1);
        ctx.state.rendered += 1;
    }
}

fn render_unchanged_block(
    f: &mut Frame,
    block: &UnchangedBlock,
    ctx: &mut RenderContext,
    diff: &DiffView,
    theme: &Theme,
) -> bool {
    for (i, line) in block.lines.iter().enumerate() {
        if ctx.state.row_index < ctx.viewport.content_start {
            ctx.state.row_index += 1;
            continue;
        }
        if ctx.state.rendered >= ctx.viewport.height {
            return false;
        }

        let old_num = block.old_start + i as u32;
        let new_num = block.new_start + i as u32;
        let left_tokens = diff.highlights.old.line_tokens(old_num);
        let right_tokens = diff.highlights.new.line_tokens(new_num);
        let left = format_cell_with_tokens(
            old_num,
            ' ',
            line,
            ctx.layout.left_width,
            Style::default(),
            left_tokens,
            theme,
        );
        let right = format_cell_with_tokens(
            new_num,
            ' ',
            line,
            ctx.layout.right_width,
            Style::default(),
            right_tokens,
            theme,
        );
        render_styled_row(f, ctx.state.y, left, right, ctx.layout.x, ctx.layout.width);

        ctx.state.y = ctx.state.y.saturating_add(1);
        ctx.state.rendered += 1;
        ctx.state.row_index += 1;
    }

    true
}

fn render_change_block(
    f: &mut Frame,
    block: &ChangeBlock,
    ctx: &mut RenderContext,
    diff: &DiffView,
    theme: &Theme,
) -> bool {
    let max_len = block.old_lines.len().max(block.new_lines.len());

    for i in 0..max_len {
        if ctx.state.row_index < ctx.viewport.content_start {
            ctx.state.row_index += 1;
            continue;
        }
        if ctx.state.rendered >= ctx.viewport.height {
            return false;
        }

        let left_spans = block.old_lines.get(i).map(|text| {
            let line_num = block.old_start + i as u32;
            let has_new_line = block.new_lines.get(i).is_some();
            let colors = if has_new_line {
                &theme.diff_changed
            } else {
                &theme.diff_removed
            };
            let tokens = diff.highlights.old.line_tokens(line_num);
            format_cell_styled(
                line_num,
                '-',
                text,
                ctx.layout.left_width,
                colors,
                tokens,
                theme,
            )
        });

        let right_spans = block.new_lines.get(i).map(|text| {
            let line_num = block.new_start + i as u32;
            let has_old_line = block.old_lines.get(i).is_some();
            let colors = if has_old_line {
                &theme.diff_changed
            } else {
                &theme.diff_added
            };
            let tokens = diff.highlights.new.line_tokens(line_num);
            format_cell_styled(
                line_num,
                '+',
                text,
                ctx.layout.right_width,
                colors,
                tokens,
                theme,
            )
        });

        render_styled_row(
            f,
            ctx.state.y,
            left_spans.unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.left_width))]),
            right_spans.unwrap_or_else(|| vec![Span::raw(blank_cell(ctx.layout.right_width))]),
            ctx.layout.x,
            ctx.layout.width,
        );

        ctx.state.y = ctx.state.y.saturating_add(1);
        ctx.state.rendered += 1;
        ctx.state.row_index += 1;
    }

    true
}

fn render_row(f: &mut Frame, y: u16, left: String, right: String, x: u16, width: u16) {
    let line = format!("{} │ {}", left, right);
    let para = Paragraph::new(line);
    let area = Rect::new(x, y, width, 1);
    f.render_widget(para, area);
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

fn truncate_to_width(input: &str, width: usize) -> String {
    pad_or_trim(input, width)
}

fn format_cell_styled(
    line_num: u32,
    sign: char,
    text: &str,
    width: usize,
    colors: &DiffColors,
    tokens: Option<&[TokenSpan]>,
    theme: &Theme,
) -> Vec<Span<'static>> {
    let line_num_str = format!("{:>3} ", line_num);
    let sign_str = sign.to_string();
    let style = Style::new().bg(colors.bg).fg(colors.fg);
    let prefix_width = line_num_str.chars().count() + sign_str.chars().count();
    let mut spans = vec![Span::raw(line_num_str), Span::styled(sign_str, style)];
    spans.extend(render_text_with_tokens(
        text,
        tokens,
        width.saturating_sub(prefix_width),
        style,
        theme,
    ));
    spans
}

fn format_cell_with_tokens(
    line_num: u32,
    sign: char,
    text: &str,
    width: usize,
    base_style: Style,
    tokens: Option<&[TokenSpan]>,
    theme: &Theme,
) -> Vec<Span<'static>> {
    let line_num_str = format!("{:>3} ", line_num);
    let sign_str = sign.to_string();
    let prefix_width = line_num_str.chars().count() + sign_str.chars().count();
    let mut spans = vec![Span::raw(line_num_str), Span::styled(sign_str, base_style)];
    spans.extend(render_text_with_tokens(
        text,
        tokens,
        width.saturating_sub(prefix_width),
        base_style,
        theme,
    ));
    spans
}

fn render_text_with_tokens(
    text: &str,
    tokens: Option<&[TokenSpan]>,
    text_area_width: usize,
    base_style: Style,
    theme: &Theme,
) -> Vec<Span<'static>> {
    if text_area_width == 0 {
        return Vec::new();
    }

    let chars: Vec<char> = text.chars().collect();
    let max_visible_text = text_area_width.saturating_sub(1);
    let visible = chars.len().min(max_visible_text);
    let mut char_kinds = vec![None; visible];

    if let Some(line_tokens) = tokens {
        for token in line_tokens {
            let start = token.start_col.min(visible);
            let end = token.end_col.min(visible);
            for kind in char_kinds.iter_mut().take(end).skip(start) {
                *kind = Some(token.kind);
            }
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

    let used_width = 1 + visible;
    if used_width < text_area_width {
        spans.push(Span::styled(
            " ".repeat(text_area_width - used_width),
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
) {
    let mut spans: Vec<Span<'_>> = left;
    spans.push(Span::raw(" │ "));
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
    fn render_text_with_tokens_applies_syntax_foreground() {
        let theme = Theme::dark();
        let tokens = vec![TokenSpan {
            start_col: 0,
            end_col: 2,
            kind: HighlightKind::Keyword,
        }];

        let spans = render_text_with_tokens("fn main", Some(&tokens), 10, Style::default(), &theme);

        assert!(
            spans.iter().any(|span| {
                span.style.fg == Some(theme.syntax.keyword) && span.content.contains("fn")
            }),
            "Expected a keyword-colored span for `fn`"
        );
    }
}
