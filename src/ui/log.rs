use crate::app::LogSummary;
use crate::core::{format_relative_time, is_in_selection, Commit, CursorState, RefInfo, RefType};
use crate::domain::graph::{render_row, Layout};
use crate::ui::theme::Theme;
use chrono::{DateTime, Utc};
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
    summary: &LogSummary,
    cursor: &CursorState,
    scroll_offset: usize,
    now: DateTime<Utc>,
    area: Rect,
    theme: &Theme,
) {
    let graph_width = graph_layout
        .rows
        .first()
        .map(|row| row.symbols.len())
        .unwrap_or(1);
    let graph_padding = "  ".repeat(graph_width);

    let summary_display_state = get_line_display_state(cursor, 0);
    let summary_prefix = summary_display_state.prefix();
    let summary_highlighted = summary_display_state.is_highlighted();
    let summary_base_style = if summary.is_selectable() {
        theme.message
    } else {
        theme.text_muted
    };
    let summary_selected_style = if summary.is_selectable() {
        theme.message_selected
    } else {
        theme.text_muted_selected
    };
    let summary_style = if summary_highlighted {
        summary_selected_style
    } else {
        summary_base_style
    };
    let summary_text = if summary.is_selectable() {
        let file_word = if summary.file_count == 1 {
            "file"
        } else {
            "files"
        };
        format!("{} · {} {}", summary.label(), summary.file_count, file_word)
    } else {
        summary.label().to_string()
    };
    let summary_line = Line::from(vec![
        Span::styled(summary_prefix, summary_style),
        Span::styled(graph_padding.clone(), summary_style),
        Span::styled(summary_text, summary_style),
    ]);
    let summary_area = Rect::new(area.x, area.y, area.width, 1);
    f.render_widget(Paragraph::new(summary_line), summary_area);

    let list_area = Rect::new(
        area.x,
        area.y.saturating_add(1),
        area.width,
        area.height.saturating_sub(2),
    );
    let visible_height = list_area.height as usize;
    let end = (scroll_offset + visible_height).min(commits.len());

    let items: Vec<ListItem> = (scroll_offset..end)
        .map(|commit_index| {
            let entry_index = commit_index + 1;
            let display_state = get_line_display_state(cursor, entry_index);
            let prefix = display_state.prefix();
            let is_highlighted = display_state.is_highlighted();

            let commit = &commits[commit_index];

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

            let graph_str = if commit_index < graph_layout.rows.len() {
                render_row(&graph_layout.rows[commit_index])
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

            let author_style = theme.author.bold_if(is_highlighted);
            let time_style = theme.time.bold_if(is_highlighted);
            let relative_time = format_relative_time(commit.date, now);

            spans.push(Span::styled(" · ", message_style));
            spans.push(Span::styled(&commit.author, author_style));
            spans.push(Span::styled(" · ", message_style));
            spans.push(Span::styled(relative_time, time_style));

            ListItem::new(Line::from(spans))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, list_area);

    if commits.is_empty() {
        let msg = Paragraph::new("No commits found")
            .style(Style::default().fg(Color::Gray))
            .alignment(Alignment::Center);
        f.render_widget(msg, list_area);
    }

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
