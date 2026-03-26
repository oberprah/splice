use crate::app::FilesView;
use crate::core::{format_relative_time, DiffRef, FileStatus};
use crate::domain::filetree::{FolderNode, TreeNode};
use crate::ui::theme::Theme;
use chrono::{DateTime, Utc};
use ratatui::{
    prelude::*,
    style::Modifier,
    widgets::{List, ListItem, Paragraph},
};

fn source_header_line(
    diff_ref: &DiffRef,
    now: DateTime<Utc>,
    width: usize,
    theme: &Theme,
) -> Line<'static> {
    if let DiffRef::CommitRange(range) = diff_ref {
        if range.is_single_commit() {
            let hash = range.end.short_hash().to_owned();
            let author = range.end.author.clone();
            let relative_time = format_relative_time(range.end.date, now);
            let suffix = format!(" · {} · {}", author, relative_time);
            let hash_len = hash.chars().count();
            let suffix_len = suffix.chars().count();
            let msg: String = range
                .end
                .message
                .chars()
                .take(width.saturating_sub(hash_len + 1 + suffix_len))
                .collect();
            return Line::from(vec![
                Span::styled(hash, theme.hash),
                Span::raw(" "),
                Span::styled(msg, theme.text.add_modifier(Modifier::BOLD)),
                Span::styled(" · ", theme.text_muted),
                Span::styled(author, theme.author),
                Span::styled(" · ", theme.text_muted),
                Span::styled(relative_time, theme.text_muted),
            ]);
        }
    }
    let header = diff_ref.header_text();
    let truncated: String = header.chars().take(width).collect();
    Line::from(Span::styled(truncated, theme.text_muted))
}

pub fn render_files_view(
    f: &mut Frame,
    files: &FilesView,
    now: DateTime<Utc>,
    area: Rect,
    theme: &Theme,
) {
    let mut y = area.y;
    let width = area.width as usize;

    render_source_header(f, &files.diff_ref, now, area.x, y, width, theme);
    y += 1;

    match files.body_display_info() {
        None => {
            f.render_widget(Paragraph::new(""), Rect::new(area.x, y, area.width, 1));
            y += 1;
        }
        Some(body_info) => {
            f.render_widget(Paragraph::new(""), Rect::new(area.x, y, area.width, 1));
            y += 1;
            let last = body_info.shown_lines.len().saturating_sub(1);
            for (i, line) in body_info.shown_lines.iter().enumerate() {
                let content: String = format!(
                    "  {}",
                    line.chars()
                        .take(width.saturating_sub(2))
                        .collect::<String>()
                );
                let is_hint = body_info.last_line_is_hint && i == last;
                let style = if is_hint {
                    theme.text_muted
                } else {
                    theme.text
                };
                f.render_widget(
                    Paragraph::new(content).style(style),
                    Rect::new(area.x, y, area.width, 1),
                );
                y += 1;
            }
            f.render_widget(Paragraph::new(""), Rect::new(area.x, y, area.width, 1));
            y += 1;
        }
    }

    let total_additions = files.total_additions();
    let total_deletions = files.total_deletions();
    let file_count = files.files.len();
    let stats_line = Span::styled(format!("{} files · ", file_count), theme.text_muted);
    let additions = Span::styled(format!("+{}", total_additions), theme.additions);
    let deletions = Span::styled(format!(" -{}", total_deletions), theme.deletions);
    let stats = Paragraph::new(Line::from(vec![stats_line, additions, deletions]));
    f.render_widget(stats, Rect::new(area.x, y, area.width, 1));
    y += 1;

    let list_area = Rect::new(
        area.x,
        y,
        area.width,
        area.height.saturating_sub(y - area.y).saturating_sub(1),
    );
    render_tree_list(
        f,
        &files.visible_items,
        files.selected,
        files.scroll_offset,
        list_area,
        theme,
    );

    let help_text = if files.message_expanded {
        "j/k: navigate  Enter/space: toggle/open  ←/→: fold  m: collapse  q: back"
    } else if files.has_expandable_body() {
        "j/k: navigate  Enter/space: toggle/open  ←/→: fold  m: expand  q: back"
    } else {
        "j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back"
    };
    let help = Paragraph::new(help_text)
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

fn render_source_header(
    f: &mut Frame,
    diff_ref: &DiffRef,
    now: DateTime<Utc>,
    x: u16,
    y: u16,
    width: usize,
    theme: &Theme,
) {
    let line = source_header_line(diff_ref, now, width, theme);
    let para = Paragraph::new(line);
    f.render_widget(para, Rect::new(x, y, width as u16, 1));
}

fn render_tree_list(
    f: &mut Frame,
    visible_items: &[crate::domain::filetree::VisibleTreeItem],
    selected: usize,
    scroll_offset: usize,
    area: Rect,
    theme: &Theme,
) {
    if visible_items.is_empty() {
        let msg = Paragraph::new("No files changed")
            .style(theme.text_muted)
            .alignment(Alignment::Center);
        f.render_widget(msg, area);
        return;
    }

    let visible_height = area.height as usize;
    let end = (scroll_offset + visible_height).min(visible_items.len());

    let items: Vec<ListItem> = visible_items
        .iter()
        .enumerate()
        .skip(scroll_offset)
        .take(end - scroll_offset)
        .map(|(i, item)| {
            let is_selected = i == selected;
            let line = format_tree_line(item, is_selected, theme);
            ListItem::new(line)
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, area);
}

fn format_tree_line(
    item: &crate::domain::filetree::VisibleTreeItem,
    is_selected: bool,
    theme: &Theme,
) -> Line<'static> {
    let mut spans = Vec::new();

    let prefix = if is_selected { "→" } else { " " };
    let prefix_style = if is_selected {
        theme.text_selected
    } else {
        theme.text_muted
    };
    spans.push(Span::styled(prefix, prefix_style));

    let tree_style = Style::default().fg(Color::DarkGray);
    for &has_more_siblings in &item.parent_lines {
        if has_more_siblings {
            spans.push(Span::styled("│   ", tree_style));
        } else {
            spans.push(Span::raw("    "));
        }
    }

    let connector = if item.is_last_child {
        "└── "
    } else {
        "├── "
    };
    spans.push(Span::styled(connector, tree_style));

    match &item.node {
        TreeNode::Folder(folder) => {
            spans.extend(format_folder_spans(folder, is_selected, theme));
        }
        TreeNode::File(file_node) => {
            spans.extend(format_file_spans(file_node, is_selected, theme));
        }
    }

    if is_selected {
        for span in &mut spans {
            span.style = span.style.add_modifier(Modifier::BOLD);
        }
    }

    Line::from(spans)
}

fn format_folder_spans(
    folder: &FolderNode,
    is_selected: bool,
    theme: &Theme,
) -> Vec<Span<'static>> {
    let mut spans = Vec::new();

    let folder_name = if folder.name.ends_with('/') {
        folder.name.clone()
    } else {
        format!("{}/", folder.name)
    };

    let name_style = if is_selected {
        theme.text_selected
    } else {
        theme.text
    };
    spans.push(Span::styled(folder_name, name_style));

    if !folder.is_expanded {
        let stats = folder.stats;
        let file_word = if stats.file_count == 1 {
            "file"
        } else {
            "files"
        };

        let muted_style = if is_selected {
            theme.text_muted_selected
        } else {
            theme.text_muted
        };
        let additions_style = if is_selected {
            theme.additions_selected
        } else {
            theme.additions
        };
        let deletions_style = if is_selected {
            theme.deletions_selected
        } else {
            theme.deletions
        };

        spans.push(Span::styled(" +", muted_style));
        spans.push(Span::styled(stats.additions.to_string(), additions_style));
        spans.push(Span::styled(" -", muted_style));
        spans.push(Span::styled(stats.deletions.to_string(), deletions_style));
        spans.push(Span::styled(
            format!(" ({} {})", stats.file_count, file_word),
            muted_style,
        ));
    }

    spans
}

fn format_file_spans(
    file_node: &crate::domain::filetree::FileNode,
    is_selected: bool,
    theme: &Theme,
) -> Vec<Span<'static>> {
    let mut spans = Vec::new();
    let file = &file_node.file;

    let status_style = if is_selected {
        match file.status {
            FileStatus::Added => theme.file_status_added_selected,
            FileStatus::Modified => theme.file_status_modified_selected,
            FileStatus::Deleted => theme.file_status_deleted_selected,
            FileStatus::Renamed => theme.file_status_renamed_selected,
        }
    } else {
        match file.status {
            FileStatus::Added => theme.file_status_added,
            FileStatus::Modified => theme.file_status_modified,
            FileStatus::Deleted => theme.file_status_deleted,
            FileStatus::Renamed => theme.file_status_renamed,
        }
    };

    spans.push(Span::styled(
        file.status.status_char().to_string(),
        status_style,
    ));

    let muted_style = if is_selected {
        theme.text_muted_selected
    } else {
        theme.text_muted
    };
    let additions_style = if is_selected {
        theme.additions_selected
    } else {
        theme.additions
    };
    let deletions_style = if is_selected {
        theme.deletions_selected
    } else {
        theme.deletions
    };

    if file.is_binary {
        spans.push(Span::styled(" (binary)  ", muted_style));
    } else {
        spans.push(Span::styled(" +", muted_style));
        spans.push(Span::styled(file.additions.to_string(), additions_style));
        spans.push(Span::styled(" -", muted_style));
        spans.push(Span::styled(
            format!("{}  ", file.deletions),
            deletions_style,
        ));
    }

    let name_style = if is_selected {
        theme.text_selected
    } else {
        theme.text
    };
    spans.push(Span::styled(file_node.name.clone(), name_style));

    if let Some(old_path) = &file.old_path {
        spans.push(Span::styled(
            format!(" (moved from {})", old_path),
            muted_style,
        ));
    }

    spans
}
