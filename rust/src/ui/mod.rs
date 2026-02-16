mod theme;

pub use theme::*;

use crate::core::{Commit, RefInfo, RefType};
use ratatui::{
    prelude::*,
    widgets::{List, ListItem, Paragraph},
};

pub fn render_log_view(f: &mut Frame, commits: &[Commit], selected: usize, scroll_offset: usize, area: Rect) {
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
    let help_area = Rect::new(area.x, area.y + area.height.saturating_sub(1), area.width, 1);
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
