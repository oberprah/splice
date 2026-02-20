use crate::core::{is_in_selection, Commit, CursorState, RefInfo, RefType};
use crate::domain::graph::{render_row, Layout};
use crate::ui::theme::Theme;
use ratatui::{
    prelude::*,
    widgets::{List, ListItem, Paragraph},
};

trait StyleExt {
    fn bold_if(self, condition: bool) -> Self;
}

impl StyleExt for Style {
    fn bold_if(self, condition: bool) -> Self {
        if condition {
            self.bold()
        } else {
            self
        }
    }
}

enum LineDisplayState {
    None,
    Cursor,
    Selected,
    VisualCursor,
}

impl LineDisplayState {
    fn prefix(&self) -> &'static str {
        match self {
            LineDisplayState::None => "  ",
            LineDisplayState::Cursor => "→ ",
            LineDisplayState::Selected => "▌ ",
            LineDisplayState::VisualCursor => "█ ",
        }
    }

    fn is_highlighted(&self) -> bool {
        matches!(
            self,
            LineDisplayState::Cursor | LineDisplayState::Selected | LineDisplayState::VisualCursor
        )
    }
}

fn get_line_display_state(cursor: &CursorState, index: usize) -> LineDisplayState {
    match cursor {
        CursorState::Normal { pos } if *pos == index => LineDisplayState::Cursor,
        CursorState::Normal { .. } => LineDisplayState::None,
        CursorState::Visual { pos, .. } if *pos == index => LineDisplayState::VisualCursor,
        CursorState::Visual { .. } if is_in_selection(cursor, index) => LineDisplayState::Selected,
        CursorState::Visual { .. } => LineDisplayState::None,
    }
}

pub fn render_log_view(
    f: &mut Frame,
    commits: &[Commit],
    graph_layout: &Layout,
    cursor: &CursorState,
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
            let display_state = get_line_display_state(cursor, i);
            let prefix = display_state.prefix();
            let is_highlighted = display_state.is_highlighted();

            let hash_style = if is_highlighted {
                theme.hash_selected
            } else {
                theme.hash
            };

            let message_style = if is_highlighted {
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
            spans.push(Span::styled(
                graph_str,
                Style::default().fg(Color::Blue).bold_if(is_highlighted),
            ));

            spans.push(Span::styled(commit.short_hash(), hash_style));

            if !commit.refs.is_empty() {
                let refs_str = format_refs(&commit.refs);
                spans.push(Span::styled(" ", message_style));
                spans.push(Span::styled(refs_str, theme.refs.bold_if(is_highlighted)));
            }

            spans.push(Span::styled(" ", message_style));
            spans.push(Span::styled(&commit.message, message_style));

            ListItem::new(Line::from(spans))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, area);

    let help = Paragraph::new("j/k: navigate  Ctrl+d/u: half-page  q: quit")
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
