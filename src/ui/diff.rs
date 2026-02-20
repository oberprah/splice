use crate::app::DiffView;
use crate::domain::diff::{ChangeBlock, DiffBlock, UnchangedBlock};
use ratatui::{prelude::*, widgets::Paragraph};

pub fn render_diff_view(f: &mut Frame, diff: &DiffView, area: Rect) {
    let mut y = area.y;
    let width = area.width as usize;

    let header = format!(
        "{} · {} · +{} -{}",
        diff.commit.short_hash(),
        diff.file.path,
        diff.file.additions,
        diff.file.deletions
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

    render_diff_lines(f, diff, content_area);

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

fn render_diff_lines(f: &mut Frame, diff: &DiffView, area: Rect) {
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

    let mut y = area.y;
    let mut rendered = 0usize;
    let mut row_index = 0usize;

    for block in &diff.diff.blocks {
        match block {
            DiffBlock::Unchanged(unchanged) => {
                if !render_unchanged_block(
                    f,
                    unchanged,
                    &mut y,
                    &mut rendered,
                    &mut row_index,
                    diff.scroll_offset,
                    area.height as usize,
                    left_width,
                    right_width,
                    area.x,
                    area.width,
                ) {
                    return;
                }
            }
            DiffBlock::Change(change) => {
                if !render_change_block(
                    f,
                    change,
                    &mut y,
                    &mut rendered,
                    &mut row_index,
                    diff.scroll_offset,
                    area.height as usize,
                    left_width,
                    right_width,
                    area.x,
                    area.width,
                ) {
                    return;
                }
            }
        }
    }
}

fn render_unchanged_block(
    f: &mut Frame,
    block: &UnchangedBlock,
    y: &mut u16,
    rendered: &mut usize,
    row_index: &mut usize,
    scroll_offset: usize,
    height: usize,
    left_width: usize,
    right_width: usize,
    x: u16,
    width: u16,
) -> bool {
    for (i, line) in block.lines.iter().enumerate() {
        if *row_index < scroll_offset {
            *row_index += 1;
            continue;
        }
        if *rendered >= height {
            return false;
        }

        let old_num = block.old_start + i as u32;
        let new_num = block.new_start + i as u32;
        let left = format_cell(old_num, ' ', line, left_width);
        let right = format_cell(new_num, ' ', line, right_width);
        render_row(f, *y, left, right, x, width);

        *y = y.saturating_add(1);
        *rendered += 1;
        *row_index += 1;
    }

    true
}

fn render_change_block(
    f: &mut Frame,
    block: &ChangeBlock,
    y: &mut u16,
    rendered: &mut usize,
    row_index: &mut usize,
    scroll_offset: usize,
    height: usize,
    left_width: usize,
    right_width: usize,
    x: u16,
    width: u16,
) -> bool {
    let max_len = block.old_lines.len().max(block.new_lines.len());

    for i in 0..max_len {
        if *row_index < scroll_offset {
            *row_index += 1;
            continue;
        }
        if *rendered >= height {
            return false;
        }

        let left = block.old_lines.get(i).map(|text| {
            let line_num = block.old_start + i as u32;
            format_cell(line_num, '-', text, left_width)
        });
        let right = block.new_lines.get(i).map(|text| {
            let line_num = block.new_start + i as u32;
            format_cell(line_num, '+', text, right_width)
        });

        render_row(
            f,
            *y,
            left.unwrap_or_else(|| blank_cell(left_width)),
            right.unwrap_or_else(|| blank_cell(right_width)),
            x,
            width,
        );

        *y = y.saturating_add(1);
        *rendered += 1;
        *row_index += 1;
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
