use crate::app::DiffView;
use crate::domain::diff::{ChangeBlock, DiffBlock, UnchangedBlock};
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
    scroll_offset: usize,
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

    let range_display = if diff.range.is_single_commit() {
        diff.range.end.short_hash().to_string()
    } else {
        format!(
            "{}..{}",
            diff.range.end.short_hash(),
            diff.range.start.short_hash()
        )
    };

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

    let help = Paragraph::new("j/k: scroll  q: back")
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
            scroll_offset: diff.scroll_offset,
        },
        state: RenderState {
            y: area.y,
            rendered: 0,
            row_index: 0,
        },
    };

    for block in &diff.diff.blocks {
        match block {
            DiffBlock::Unchanged(unchanged) => {
                if !render_unchanged_block(f, unchanged, ctx) {
                    return;
                }
            }
            DiffBlock::Change(change) => {
                if !render_change_block(f, change, ctx, theme) {
                    return;
                }
            }
        }
    }
}

fn render_unchanged_block(f: &mut Frame, block: &UnchangedBlock, ctx: &mut RenderContext) -> bool {
    for (i, line) in block.lines.iter().enumerate() {
        if ctx.state.row_index < ctx.viewport.scroll_offset {
            ctx.state.row_index += 1;
            continue;
        }
        if ctx.state.rendered >= ctx.viewport.height {
            return false;
        }

        let old_num = block.old_start + i as u32;
        let new_num = block.new_start + i as u32;
        let left = format_cell(old_num, ' ', line, ctx.layout.left_width);
        let right = format_cell(new_num, ' ', line, ctx.layout.right_width);
        render_row(f, ctx.state.y, left, right, ctx.layout.x, ctx.layout.width);

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
    theme: &Theme,
) -> bool {
    let max_len = block.old_lines.len().max(block.new_lines.len());

    for i in 0..max_len {
        if ctx.state.row_index < ctx.viewport.scroll_offset {
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
            format_cell_styled(line_num, '-', text, ctx.layout.left_width, colors)
        });

        let right_spans = block.new_lines.get(i).map(|text| {
            let line_num = block.new_start + i as u32;
            let has_old_line = block.old_lines.get(i).is_some();
            let colors = if has_old_line {
                &theme.diff_changed
            } else {
                &theme.diff_added
            };
            format_cell_styled(line_num, '+', text, ctx.layout.right_width, colors)
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

fn format_cell(line_num: u32, sign: char, text: &str, width: usize) -> String {
    let cell = format!("{:>3} {} {}", line_num, sign, text);
    pad_or_trim(&cell, width)
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
    if result.len() < width {
        result.push_str(&" ".repeat(width - result.len()));
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
) -> Vec<Span<'static>> {
    let line_num_str = format!("{:>3} ", line_num);
    let sign_str = sign.to_string();
    let text_str = format!(" {}", text);

    let total_len = line_num_str.len() + sign_str.len() + text_str.len();
    let padded_text = if total_len < width {
        format!("{}{}", text_str, " ".repeat(width - total_len))
    } else if total_len > width {
        text_str
            .chars()
            .take(width - line_num_str.len() - sign_str.len())
            .collect()
    } else {
        text_str
    };

    let style = Style::new().bg(colors.bg).fg(colors.fg);

    vec![
        Span::raw(line_num_str),
        Span::styled(sign_str, style),
        Span::styled(padded_text, style),
    ]
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
