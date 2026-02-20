use crate::core::{Commit, RefInfo, RefType};
use crate::domain::graph::{render_row, Layout};
use crate::ui::theme::Theme;
use ratatui::{
    prelude::*,
    widgets::{List, ListItem, Paragraph},
};

pub fn render_log_view(
    f: &mut Frame,
    commits: &[Commit],
    graph_layout: &Layout,
    selected: usize,
    scroll_offset: usize,
    area: Rect,
    theme: &Theme,
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
                theme.hash_selected
            } else {
                theme.hash
            };

            let message_style = if is_selected {
                theme.message_selected
            } else {
                theme.message
            };

            let mut spans = vec![Span::styled(prefix, message_style)];

            let graph_str = if i < graph_layout.rows.len() {
                render_row(&graph_layout.rows[i])
            } else {
                String::new()
            };
            spans.push(Span::styled(graph_str, Style::default().fg(Color::Blue)));

            spans.push(Span::styled(commit.short_hash(), hash_style));

            if !commit.refs.is_empty() {
                let refs_str = format_refs(&commit.refs);
                spans.push(Span::raw(" "));
                spans.push(Span::styled(refs_str, theme.refs));
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
