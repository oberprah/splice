use crate::app::FilesView;
use crate::core::{Commit, FileChange, FileStatus};
use ratatui::{
    prelude::*,
    widgets::{List, ListItem, Paragraph},
};

pub fn render_files_view(f: &mut Frame, files: &FilesView, area: Rect) {
    let mut y = area.y;
    let width = area.width as usize;

    render_commit_header(f, &files.commit, area.x, y, width);
    y += 1;

    let empty = Paragraph::new("");
    f.render_widget(empty, Rect::new(area.x, y, area.width, 1));
    y += 1;

    let subject =
        Paragraph::new(files.commit.message.as_str()).style(Style::default().fg(Color::White));
    f.render_widget(subject, Rect::new(area.x, y, area.width, 1));
    y += 1;

    let empty = Paragraph::new("");
    f.render_widget(empty, Rect::new(area.x, y, area.width, 1));
    y += 1;

    let total_additions = files.total_additions();
    let total_deletions = files.total_deletions();
    let stats_line = format!(
        "{} files · +{} -{}",
        files.files.len(),
        total_additions,
        total_deletions
    );
    let stats = Paragraph::new(stats_line).style(Style::default().fg(Color::Gray));
    f.render_widget(stats, Rect::new(area.x, y, area.width, 1));
    y += 1;

    let list_area = Rect::new(
        area.x,
        y,
        area.width,
        area.height.saturating_sub(y - area.y).saturating_sub(1),
    );
    render_files_list(
        f,
        &files.files,
        files.selected,
        files.scroll_offset,
        list_area,
    );

    let help = Paragraph::new("j/k: navigate  Enter: open diff  q: back")
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

fn render_commit_header(f: &mut Frame, commit: &Commit, x: u16, y: u16, width: usize) {
    let time_ago = format_time_ago(&commit.date);
    let header = format!(
        "{} · {} committed {}",
        commit.short_hash(),
        commit.author,
        time_ago
    );

    let truncated: String = header.chars().take(width).collect();
    let para = Paragraph::new(truncated).style(Style::default().fg(Color::Gray));
    f.render_widget(para, Rect::new(x, y, width as u16, 1));
}

fn format_time_ago(date: &chrono::DateTime<chrono::Utc>) -> String {
    let now = chrono::Utc::now();
    let duration = now.signed_duration_since(*date);

    if duration.num_seconds() < 60 {
        "just now".to_string()
    } else if duration.num_minutes() < 60 {
        format!("{} minutes ago", duration.num_minutes())
    } else if duration.num_hours() < 24 {
        format!("{} hours ago", duration.num_hours())
    } else if duration.num_days() < 30 {
        format!("{} days ago", duration.num_days())
    } else if duration.num_weeks() < 52 {
        format!("{} months ago", duration.num_weeks() / 4)
    } else {
        format!("{} years ago", duration.num_weeks() / 52)
    }
}

fn render_files_list(
    f: &mut Frame,
    files: &[FileChange],
    selected: usize,
    scroll_offset: usize,
    area: Rect,
) {
    if files.is_empty() {
        let msg = Paragraph::new("No files changed")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, area);
        return;
    }

    let visible_height = area.height as usize;
    let end = (scroll_offset + visible_height).min(files.len());

    let items: Vec<ListItem> = files
        .iter()
        .enumerate()
        .skip(scroll_offset)
        .take(end - scroll_offset)
        .map(|(i, file)| {
            let is_selected = i == selected;
            let prefix = if is_selected { "→ " } else { "  " };

            let status_style = match file.status {
                FileStatus::Added => Style::default().fg(Color::Green),
                FileStatus::Modified => Style::default().fg(Color::Yellow),
                FileStatus::Deleted => Style::default().fg(Color::Red),
                FileStatus::Renamed => Style::default().fg(Color::Cyan),
            };

            let additions_style = Style::default().fg(Color::Green);
            let deletions_style = Style::default().fg(Color::Red);

            let mut spans = vec![Span::styled(prefix, Style::default())];
            spans.push(Span::styled(
                file.status.status_char().to_string(),
                status_style,
            ));
            spans.push(Span::raw(" "));

            let add_str = format!("+{}", file.additions);
            let del_str = format!("-{}", file.deletions);
            spans.push(Span::styled(add_str, additions_style));
            spans.push(Span::raw(" "));
            spans.push(Span::styled(del_str, deletions_style));
            spans.push(Span::raw("  "));
            spans.push(Span::styled(&file.path, Style::default()));

            ListItem::new(Line::from(spans))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, area);
}
