mod theme;

pub use theme::*;

use crate::core::{Commit, RefInfo, RefType};
use crate::App;
use ratatui::{
    prelude::*,
    widgets::{List, ListItem, Paragraph},
};

const LOG_AREA_X: u16 = 2;
const LOG_AREA_Y: u16 = 0;
const LOG_AREA_RIGHT_MARGIN: u16 = 4;
const LOG_AREA_BOTTOM_MARGIN: u16 = 0;

pub fn render(f: &mut Frame, app: &mut App) {
    let size = f.area();

    if let Some(ref error) = app.error {
        let msg = Paragraph::new(format!("Error: {}", error))
            .style(Style::default().fg(Color::Red))
            .alignment(Alignment::Center);
        f.render_widget(msg, size);
        return;
    }

    let log_area = Rect::new(
        LOG_AREA_X,
        LOG_AREA_Y,
        size.width.saturating_sub(LOG_AREA_RIGHT_MARGIN),
        size.height.saturating_sub(LOG_AREA_BOTTOM_MARGIN),
    );
    app.set_viewport_height(log_area.height.saturating_sub(1) as usize);
    render_log_view(f, &app.commits, app.selected, app.scroll_offset, log_area);
}

pub fn render_log_view(
    f: &mut Frame,
    commits: &[Commit],
    selected: usize,
    scroll_offset: usize,
    area: Rect,
) {
    if commits.is_empty() {
        let msg = Paragraph::new("No commits found")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, area);
        return;
    }

    let visible_height = area.height.saturating_sub(1) as usize;
    let end = (scroll_offset + visible_height).min(commits.len());

    let items: Vec<ListItem> = commits
        .iter()
        .enumerate()
        .skip(scroll_offset)
        .take(end - scroll_offset)
        .map(|(i, commit)| {
            let is_selected = i == selected;
            let prefix = if is_selected { "→ " } else { "  " };

            let hash_style = if is_selected {
                theme::SELECTED_HASH
            } else {
                theme::HASH
            };

            let message_style = if is_selected {
                theme::SELECTED_MESSAGE
            } else {
                theme::MESSAGE
            };

            let mut spans = vec![Span::styled(prefix, message_style)];
            spans.push(Span::styled(commit.short_hash(), hash_style));

            if !commit.refs.is_empty() {
                let refs_str = format_refs(&commit.refs);
                spans.push(Span::raw(" "));
                spans.push(Span::styled(refs_str, theme::REFS));
            }

            spans.push(Span::raw(" "));
            spans.push(Span::styled(&commit.message, message_style));

            ListItem::new(Line::from(spans))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, area);

    let help = Paragraph::new("j/k: navigate  Ctrl+d/u: half-page  q: quit")
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

fn format_refs(refs: &[RefInfo]) -> String {
    if refs.is_empty() {
        return String::new();
    }

    let parts: Vec<String> = refs
        .iter()
        .map(|r| {
            let prefix = match r.ref_type {
                RefType::Tag => "tag: ",
                RefType::Branch | RefType::RemoteBranch | RefType::DetachedHead => "",
            };
            format!("{}{}", prefix, r.name)
        })
        .collect();

    format!("({})", parts.join(", "))
}
