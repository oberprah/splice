use crate::core::{DiffSource, FileChange};
use crate::domain::filetree::{self, TreeNode, VisibleTreeItem};

const FIXED_ROWS: usize = 4;
const COLLAPSED_BODY_LIMIT: usize = 3;
const MIN_VISIBLE_FILES: usize = 3;

pub struct BodyDisplayInfo {
    pub shown_lines: Vec<String>,
    pub last_line_is_hint: bool,
}

pub struct FilesView {
    pub source: DiffSource,
    pub files: Vec<FileChange>,
    pub root: TreeNode,
    pub visible_items: Vec<VisibleTreeItem>,
    pub selected: usize,
    pub scroll_offset: usize,
    total_height: usize,
    pub message_expanded: bool,
}

impl FilesView {
    pub fn new(source: DiffSource, files: Vec<FileChange>) -> Self {
        let (root, visible_items) = filetree::build_visible_tree(&files);

        Self {
            source,
            files,
            root,
            visible_items,
            selected: 0,
            scroll_offset: 0,
            total_height: 0,
            message_expanded: false,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.total_height = height;
        self.clamp_scroll_offset();
    }

    pub fn list_viewport_height(&self) -> usize {
        self.total_height
            .saturating_sub(FIXED_ROWS + self.body_section_lines())
    }

    fn body(&self) -> Option<&str> {
        match &self.source {
            DiffSource::CommitRange(range) if range.is_single_commit() => range.end.body.as_deref(),
            _ => None,
        }
    }

    fn body_cap(&self) -> usize {
        if self.message_expanded {
            let blank_before_body = 1;
            let show_less_line = 1;
            let reserved = FIXED_ROWS + MIN_VISIBLE_FILES + blank_before_body + show_less_line;
            self.total_height.saturating_sub(reserved).max(1)
        } else {
            COLLAPSED_BODY_LIMIT
        }
    }

    pub fn body_section_lines(&self) -> usize {
        match self.body_display_info() {
            None => 0,
            Some(info) => {
                let blank_before_body = 1;
                blank_before_body + info.shown_lines.len()
            }
        }
    }

    pub fn body_display_info(&self) -> Option<BodyDisplayInfo> {
        if self.total_height == 0 {
            return None;
        }
        let body = self.body()?;
        let all_lines: Vec<&str> = body.lines().collect();
        if all_lines.is_empty() {
            return None;
        }

        let cap = self.body_cap();
        let collapsed_cap = COLLAPSED_BODY_LIMIT;
        let is_truncated = all_lines.len() > cap;
        let has_expandable = all_lines.len() > collapsed_cap;

        let (shown_lines, last_line_is_hint) = if is_truncated {
            let shown_count = cap.saturating_sub(1).max(1);
            let mut lines: Vec<String> = all_lines[..shown_count.min(all_lines.len())]
                .iter()
                .map(|l| l.to_string())
                .collect();
            let remaining = all_lines.len() - shown_count.min(all_lines.len());
            lines.push(format!("↓ {} more lines  (m: expand)", remaining));
            (lines, true)
        } else {
            let mut lines: Vec<String> = all_lines.iter().map(|l| l.to_string()).collect();
            if has_expandable && self.message_expanded {
                lines.push("↑ show less (m: collapse)".to_string());
                (lines, true)
            } else {
                (lines, false)
            }
        };

        Some(BodyDisplayInfo {
            shown_lines,
            last_line_is_hint,
        })
    }

    pub fn has_expandable_body(&self) -> bool {
        let Some(body) = self.body() else {
            return false;
        };
        body.lines().count() > COLLAPSED_BODY_LIMIT
    }

    pub fn toggle_message(&mut self) -> bool {
        if !self.has_expandable_body() {
            return false;
        }
        self.message_expanded = !self.message_expanded;
        self.clamp_scroll_offset();
        true
    }

    fn clamp_scroll_offset(&mut self) {
        if self.visible_items.is_empty() {
            self.selected = 0;
            self.scroll_offset = 0;
            return;
        }
        let lvh = self.list_viewport_height();
        if self.selected < self.scroll_offset {
            self.scroll_offset = self.selected;
        } else if lvh > 0 && self.selected >= self.scroll_offset + lvh {
            self.scroll_offset = self.selected - lvh + 1;
        }
    }

    pub fn move_down(&mut self, amount: usize) {
        if self.visible_items.is_empty() {
            return;
        }
        let last = self.visible_items.len().saturating_sub(1);
        self.selected = self.selected.saturating_add(amount).min(last);
        self.clamp_scroll_offset();
    }

    pub fn move_up(&mut self, amount: usize) {
        if self.visible_items.is_empty() {
            return;
        }
        self.selected = self.selected.saturating_sub(amount);
        self.clamp_scroll_offset();
    }

    pub fn page_step(&self) -> usize {
        (self.list_viewport_height() / 2).max(1)
    }

    pub fn selected_file(&self) -> Option<&FileChange> {
        let item = self.visible_items.get(self.selected)?;
        if let TreeNode::File(file_node) = &item.node {
            Some(&file_node.file)
        } else {
            None
        }
    }

    pub fn selected_is_folder(&self) -> bool {
        let Some(item) = self.visible_items.get(self.selected) else {
            return false;
        };
        matches!(item.node, TreeNode::Folder(_))
    }

    pub fn toggle_folder(&mut self, expand_only: bool, collapse_only: bool) -> bool {
        let Some((new_visible, new_cursor)) = filetree::toggle_folder_at_cursor(
            &self.visible_items,
            self.selected,
            expand_only,
            collapse_only,
        ) else {
            return false;
        };

        self.visible_items = new_visible;
        self.selected = new_cursor;
        self.clamp_scroll_offset();
        true
    }

    pub fn select_file_path(&mut self, path: &str) -> bool {
        let Some(idx) = self.visible_items.iter().position(|item| match &item.node {
            TreeNode::File(file_node) => file_node.file.path == path,
            TreeNode::Folder(_) => false,
        }) else {
            return false;
        };

        self.selected = idx;
        self.clamp_scroll_offset();
        true
    }

    pub fn adjacent_visible_file(
        &self,
        current_path: &str,
        direction: isize,
    ) -> Option<FileChange> {
        let visible_files: Vec<&FileChange> = self
            .visible_items
            .iter()
            .filter_map(|item| match &item.node {
                TreeNode::File(file_node) => Some(&file_node.file),
                TreeNode::Folder(_) => None,
            })
            .collect();

        let current_idx = visible_files
            .iter()
            .position(|file| file.path == current_path)? as isize;
        let target_idx = current_idx + direction;
        if target_idx < 0 {
            return None;
        }

        visible_files
            .get(target_idx as usize)
            .map(|file| (*file).clone())
    }

    pub fn total_additions(&self) -> u32 {
        self.files.iter().map(|f| f.additions).sum()
    }

    pub fn total_deletions(&self) -> u32 {
        self.files.iter().map(|f| f.deletions).sum()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::{Commit, CommitRange};
    use chrono::{TimeZone, Utc};

    fn commit_with_body(body: &str) -> Commit {
        Commit {
            hash: "abc123def456".to_string(),
            parent_hashes: vec!["parent".to_string()],
            refs: vec![],
            message: "Fix thing".to_string(),
            body: Some(body.to_string()),
            author: "Test".to_string(),
            date: Utc.timestamp_opt(0, 0).unwrap(),
        }
    }

    fn commit_no_body() -> Commit {
        Commit {
            hash: "abc123def456".to_string(),
            parent_hashes: vec!["parent".to_string()],
            refs: vec![],
            message: "Fix thing".to_string(),
            body: None,
            author: "Test".to_string(),
            date: Utc.timestamp_opt(0, 0).unwrap(),
        }
    }

    fn single_commit_source(commit: Commit) -> DiffSource {
        DiffSource::CommitRange(CommitRange {
            start: commit.clone(),
            end: commit,
            count: 1,
            include_start: true,
        })
    }

    fn view_with_body(body: &str, total_height: usize) -> FilesView {
        let source = single_commit_source(commit_with_body(body));
        let mut view = FilesView::new(source, vec![]);
        view.set_viewport_height(total_height);
        view
    }

    #[test]
    fn body_section_lines_is_zero_when_no_body() {
        let source = single_commit_source(commit_no_body());
        let mut view = FilesView::new(source, vec![]);
        view.set_viewport_height(24);
        assert_eq!(view.body_section_lines(), 0);
    }

    #[test]
    fn body_section_lines_is_zero_for_uncommitted_source() {
        use crate::core::UncommittedType;
        let source = DiffSource::Uncommitted(UncommittedType::All);
        let mut view = FilesView::new(source, vec![]);
        view.set_viewport_height(24);
        assert_eq!(view.body_section_lines(), 0);
    }

    #[test]
    fn body_section_lines_counts_shown_lines_plus_blank_before() {
        // 3 body lines → 3 shown + 1 blank before = 4
        let view = view_with_body("Line 1\nLine 2\nLine 3", 24);
        assert_eq!(view.body_section_lines(), 4);
    }

    #[test]
    fn body_section_lines_caps_at_three_lines() {
        // collapsed cap=3. Body has 12 lines → shown=[l1,l2,indicator]=3, +1 blank before = 4
        let body = (0..12)
            .map(|i| format!("Line {}", i))
            .collect::<Vec<_>>()
            .join("\n");
        let view = view_with_body(&body, 24);
        assert_eq!(view.body_section_lines(), 4);
    }

    #[test]
    fn list_viewport_height_accounts_for_body() {
        // total=24, body=3 lines → body_section=4 (3 shown + 1 blank before), list=24-4-4=16
        let view = view_with_body("Line 1\nLine 2\nLine 3", 24);
        assert_eq!(view.list_viewport_height(), 16);
    }

    #[test]
    fn list_viewport_height_unchanged_when_no_body() {
        let source = single_commit_source(commit_no_body());
        let mut view = FilesView::new(source, vec![]);
        view.set_viewport_height(24);
        assert_eq!(view.list_viewport_height(), 20); // 24 - 4 - 0
    }

    #[test]
    fn has_expandable_body_false_when_body_fits_in_cap() {
        // collapsed cap=3. Body has 3 lines → fits exactly, not expandable
        let view = view_with_body("Line 1\nLine 2\nLine 3", 24);
        assert!(!view.has_expandable_body());
    }

    #[test]
    fn has_expandable_body_true_when_body_exceeds_cap() {
        // collapsed cap=3. Body has 12 lines → expandable
        let body = (0..12)
            .map(|i| format!("Line {}", i))
            .collect::<Vec<_>>()
            .join("\n");
        let view = view_with_body(&body, 24);
        assert!(view.has_expandable_body());
    }

    #[test]
    fn toggle_message_returns_false_when_not_expandable() {
        let mut view = view_with_body("Short body", 24);
        assert!(!view.toggle_message());
        assert!(!view.message_expanded);
    }

    #[test]
    fn toggle_message_expands_and_collapses() {
        let body = (0..12)
            .map(|i| format!("Line {}", i))
            .collect::<Vec<_>>()
            .join("\n");
        let mut view = view_with_body(&body, 24);
        assert!(view.toggle_message());
        assert!(view.message_expanded);
        assert!(view.toggle_message());
        assert!(!view.message_expanded);
    }

    #[test]
    fn body_display_info_none_when_no_body() {
        let source = single_commit_source(commit_no_body());
        let mut view = FilesView::new(source, vec![]);
        view.set_viewport_height(24);
        assert!(view.body_display_info().is_none());
    }

    #[test]
    fn body_display_info_shows_all_lines_when_short() {
        let view = view_with_body("Line 1\nLine 2", 24);
        let info = view.body_display_info().unwrap();
        assert_eq!(info.shown_lines, vec!["Line 1", "Line 2"]);
        assert!(!info.last_line_is_hint);
    }

    #[test]
    fn body_display_info_truncates_with_indicator_when_long() {
        let body = (0..12)
            .map(|i| format!("Line {}", i))
            .collect::<Vec<_>>()
            .join("\n");
        let view = view_with_body(&body, 24);
        let info = view.body_display_info().unwrap();
        // collapsed cap=3, shown_count=2, shown=[l0, l1, indicator] = 3 total
        assert_eq!(info.shown_lines.len(), 3);
        assert!(info.shown_lines.last().unwrap().contains("more lines"));
        assert!(info.shown_lines.last().unwrap().contains("(m: expand)"));
        assert!(info.last_line_is_hint);
    }

    #[test]
    fn body_display_info_expanded_shows_more_lines_with_collapse_hint() {
        let body = (0..12)
            .map(|i| format!("Line {}", i))
            .collect::<Vec<_>>()
            .join("\n");
        let mut view = view_with_body(&body, 24);
        view.toggle_message();
        let info = view.body_display_info().unwrap();
        // expanded cap = 24-9 = 15, body has 12 lines → all fit + "show less" appended
        assert_eq!(info.shown_lines.len(), 13);
        assert_eq!(
            info.shown_lines.last().unwrap(),
            "↑ show less (m: collapse)"
        );
        assert!(info.last_line_is_hint);
    }
}
