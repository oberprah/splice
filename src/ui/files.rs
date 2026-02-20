use crate::app::FilesView;
use crate::domain::filetree::{FolderNode, TreeNode};
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
    let file_count = files.files.len();
    let stats_line = format!(
        "{} files · +{} -{}",
        file_count, total_additions, total_deletions
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
    render_tree_list(
        f,
        &files.visible_items,
        files.selected,
        files.scroll_offset,
        list_area,
    );

    let help =
        Paragraph::new("j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back")
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

fn render_commit_header(f: &mut Frame, commit: &crate::core::Commit, x: u16, y: u16, width: usize) {
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

fn render_tree_list(
    f: &mut Frame,
    visible_items: &[crate::domain::filetree::VisibleTreeItem],
    selected: usize,
    scroll_offset: usize,
    area: Rect,
) {
    if visible_items.is_empty() {
        let msg = Paragraph::new("No files changed")
            .style(Style::default().fg(Color::Gray))
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
            let line = format_tree_line(item, is_selected);
            ListItem::new(Line::from(line))
        })
        .collect();

    let list = List::new(items);
    f.render_widget(list, area);
}

fn format_tree_line(item: &crate::domain::filetree::VisibleTreeItem, is_selected: bool) -> String {
    let mut line = String::new();

    if is_selected {
        line.push('→');
    } else {
        line.push(' ');
    }

    for &has_more_siblings in &item.parent_lines {
        if has_more_siblings {
            line.push_str("│   ");
        } else {
            line.push_str("    ");
        }
    }

    if item.is_last_child {
        line.push_str("└── ");
    } else {
        line.push_str("├── ");
    }

    match &item.node {
        TreeNode::Folder(folder) => {
            line.push_str(&format_folder_node(folder, is_selected));
        }
        TreeNode::File(file_node) => {
            line.push_str(&format_file_node(file_node, is_selected));
        }
    }

    line
}

fn format_folder_node(folder: &FolderNode, _is_selected: bool) -> String {
    let mut content = String::new();

    let folder_name = if folder.name.ends_with('/') {
        folder.name.clone()
    } else {
        format!("{}/", folder.name)
    };

    content.push_str(&folder_name);

    if !folder.is_expanded {
        let stats = folder.stats;
        let file_word = if stats.file_count == 1 {
            "file"
        } else {
            "files"
        };
        content.push_str(&format!(
            " +{} -{} ({} {})",
            stats.additions, stats.deletions, stats.file_count, file_word
        ));
    }

    content
}

fn format_file_node(file_node: &crate::domain::filetree::FileNode, _is_selected: bool) -> String {
    let file = &file_node.file;
    let mut content = String::new();

    let status_char = file.status.status_char();
    content.push(status_char);

    if file.is_binary {
        content.push_str(" (binary)  ");
    } else {
        content.push_str(&format!(" +{} -{}  ", file.additions, file.deletions));
    }

    content.push_str(&file_node.name);

    content
}
