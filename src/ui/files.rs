use crate::app::FilesView;
use crate::core::{DiffSource, FileStatus};
use crate::domain::filetree::{FolderNode, TreeNode};
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

pub fn render_files_view(f: &mut Frame, files: &FilesView, area: Rect, theme: &Theme) {
    let mut y = area.y;
    let width = area.width as usize;

    render_source_header(f, &files.source, area.x, y, width);
    y += 1;

    let empty = Paragraph::new("");
    f.render_widget(empty, Rect::new(area.x, y, area.width, 1));
    y += 1;

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

    let help =
        Paragraph::new("j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back")
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

fn render_source_header(f: &mut Frame, source: &DiffSource, x: u16, y: u16, width: usize) {
    let header = source.header_text();
    let truncated: String = header.chars().take(width).collect();
    let para = Paragraph::new(truncated).style(Style::default().fg(Color::Gray));
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
        theme.selection
    } else {
        Style::default()
    };
    spans.push(Span::styled(prefix, prefix_style));

    let tree_style = Style::default().fg(Color::DarkGray).bold_if(is_selected);
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

    let name_style = Style::default().fg(Color::Black).bold_if(is_selected);
    spans.push(Span::styled(folder_name, name_style));

    if !folder.is_expanded {
        let stats = folder.stats;
        let file_word = if stats.file_count == 1 {
            "file"
        } else {
            "files"
        };

        let muted_style = theme.text_muted.bold_if(is_selected);
        let additions_style = theme.additions.bold_if(is_selected);
        let deletions_style = theme.deletions.bold_if(is_selected);

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

    let status_style = match file.status {
        FileStatus::Added => theme.file_status_added,
        FileStatus::Modified => theme.file_status_modified,
        FileStatus::Deleted => theme.file_status_deleted,
        FileStatus::Renamed => theme.file_status_renamed,
    }
    .bold_if(is_selected);

    spans.push(Span::styled(
        file.status.status_char().to_string(),
        status_style,
    ));

    let muted_style = theme.text_muted.bold_if(is_selected);
    let additions_style = theme.additions.bold_if(is_selected);
    let deletions_style = theme.deletions.bold_if(is_selected);

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

    let name_style = theme.text.bold_if(is_selected);
    spans.push(Span::styled(file_node.name.clone(), name_style));

    spans
}
