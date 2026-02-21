use crate::core::{selection_range, Commit, CommitRange, CursorState, UncommittedType};
use crate::domain::graph::{compute_layout, GraphCommit, Layout};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct LogSummary {
    pub uncommitted_type: Option<UncommittedType>,
    pub file_count: usize,
}

impl LogSummary {
    pub fn clean() -> Self {
        Self {
            uncommitted_type: None,
            file_count: 0,
        }
    }

    pub fn is_selectable(&self) -> bool {
        self.uncommitted_type.is_some()
    }

    pub fn label(&self) -> &'static str {
        match self.uncommitted_type {
            Some(UncommittedType::Staged) => "Staged changes",
            Some(UncommittedType::Unstaged) => "Unstaged changes",
            Some(UncommittedType::All) => "Uncommitted changes",
            None => "Working tree clean",
        }
    }
}

pub struct LogView {
    pub commits: Vec<Commit>,
    pub graph_layout: Layout,
    pub cursor: CursorState,
    pub scroll_offset: usize,
    pub viewport_height: usize,
    pub summary: LogSummary,
}

impl LogView {
    pub fn new(commits: Vec<Commit>) -> Self {
        let graph_commits: Vec<GraphCommit> = commits
            .iter()
            .map(|c| GraphCommit {
                hash: c.hash.clone(),
                parents: c.parent_hashes.clone(),
            })
            .collect();
        let graph_layout = compute_layout(&graph_commits);
        let summary = LogSummary::clean();
        let cursor = if commits.is_empty() {
            CursorState::Normal { pos: 0 }
        } else {
            CursorState::Normal { pos: 1 }
        };
        Self {
            commits,
            graph_layout,
            cursor,
            scroll_offset: 0,
            viewport_height: 0,
            summary,
        }
    }

    pub fn set_summary(&mut self, summary: LogSummary) {
        self.summary = summary;
        if self.summary.is_selectable() {
            self.cursor = CursorState::Normal { pos: 0 };
        } else if !self.commits.is_empty() && self.cursor.position() == 0 {
            self.cursor = CursorState::Normal { pos: 1 };
        }
        self.clamp_scroll_offset();
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
    }

    fn clamp_scroll_offset(&mut self) {
        let total = self.entry_count();
        if total == 0 {
            self.cursor = CursorState::Normal { pos: 0 };
            self.scroll_offset = 0;
            return;
        }

        if self.commits.is_empty() {
            let pos = if self.summary.is_selectable() { 0 } else { 0 };
            self.cursor = CursorState::Normal { pos };
            self.scroll_offset = 0;
            return;
        }

        let min_pos = self.min_selectable_index();
        let max_pos = total.saturating_sub(1);
        let pos = self.cursor.position().clamp(min_pos, max_pos);

        self.cursor = match self.cursor {
            CursorState::Normal { .. } => CursorState::Normal { pos },
            CursorState::Visual { anchor, .. } => CursorState::Visual {
                pos,
                anchor: anchor.clamp(min_pos, max_pos),
            },
        };

        if let Some(commit_index) = self.selected_commit_index() {
            if commit_index < self.scroll_offset {
                self.scroll_offset = commit_index;
            } else if self.viewport_height > 0
                && commit_index >= self.scroll_offset + self.viewport_height
            {
                self.scroll_offset = commit_index - self.viewport_height + 1;
            }
        }
    }

    pub fn move_down(&mut self, amount: usize) {
        let total = self.entry_count();
        if total == 0 {
            return;
        }
        let last = total.saturating_sub(1);
        let new_pos = self.cursor.position().saturating_add(amount).min(last);
        self.cursor = match self.cursor {
            CursorState::Normal { .. } => CursorState::Normal { pos: new_pos },
            CursorState::Visual { anchor, .. } => CursorState::Visual {
                pos: new_pos,
                anchor,
            },
        };
        self.clamp_scroll_offset();
    }

    pub fn move_up(&mut self, amount: usize) {
        let total = self.entry_count();
        if total == 0 {
            return;
        }
        let new_pos = self.cursor.position().saturating_sub(amount);
        self.cursor = match self.cursor {
            CursorState::Normal { .. } => CursorState::Normal { pos: new_pos },
            CursorState::Visual { anchor, .. } => CursorState::Visual {
                pos: new_pos,
                anchor,
            },
        };
        self.clamp_scroll_offset();
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    pub fn cursor_position(&self) -> usize {
        self.cursor.position()
    }

    pub fn selected_commit(&self) -> Option<&Commit> {
        let index = self.selected_commit_index()?;
        self.commits.get(index)
    }

    pub fn selected_uncommitted_type(&self) -> Option<UncommittedType> {
        if self.cursor.position() == 0 {
            self.summary.uncommitted_type
        } else {
            None
        }
    }

    pub fn is_visual_mode(&self) -> bool {
        matches!(self.cursor, CursorState::Visual { .. })
    }

    pub fn enter_visual_mode(&mut self) {
        if self.selected_commit().is_none() {
            return;
        }
        let pos = self.cursor.position();
        self.cursor = CursorState::Visual { pos, anchor: pos };
    }

    pub fn exit_visual_mode(&mut self) {
        let pos = self.cursor.position();
        self.cursor = CursorState::Normal { pos };
    }

    pub fn get_selected_range(&self) -> Option<CommitRange> {
        let (min, max) = selection_range(&self.cursor);
        if min == 0 || max == 0 {
            return None;
        }
        let end = self.commits.get(min - 1)?.clone();
        let start = self.commits.get(max - 1)?.clone();
        let count = max - min + 1;
        Some(CommitRange {
            start,
            end,
            count,
            include_start: true,
        })
    }

    fn entry_count(&self) -> usize {
        1 + self.commits.len()
    }

    fn min_selectable_index(&self) -> usize {
        if self.summary.is_selectable() || self.commits.is_empty() {
            0
        } else {
            1
        }
    }

    fn selected_commit_index(&self) -> Option<usize> {
        let pos = self.cursor.position();
        if pos == 0 {
            None
        } else {
            Some(pos - 1)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{TimeZone, Utc};

    fn test_commit(hash: &str) -> Commit {
        Commit {
            hash: hash.to_string(),
            parent_hashes: vec![],
            refs: vec![],
            message: "test message".to_string(),
            author: "test author".to_string(),
            date: Utc.timestamp_opt(0, 0).unwrap(),
        }
    }

    fn make_commits(n: usize) -> Vec<Commit> {
        (0..n).map(|i| test_commit(&format!("hash{i}"))).collect()
    }

    #[test]
    fn new_initializes_cursor_to_first_commit() {
        let view = LogView::new(make_commits(3));
        assert_eq!(view.cursor_position(), 1);
        assert!(!view.is_visual_mode());
    }

    #[test]
    fn cursor_position_returns_current_position() {
        let mut view = LogView::new(make_commits(3));
        view.cursor = CursorState::Normal { pos: 2 };
        assert_eq!(view.cursor_position(), 2);
    }

    #[test]
    fn selected_commit_returns_commit_at_cursor() {
        let view = LogView::new(make_commits(3));
        assert_eq!(view.selected_commit().unwrap().hash, "hash0");
    }

    #[test]
    fn selected_commit_returns_none_for_empty_commits() {
        let view = LogView::new(Vec::new());
        assert!(view.selected_commit().is_none());
    }

    #[test]
    fn is_visual_mode_returns_false_initially() {
        let view = LogView::new(make_commits(3));
        assert!(!view.is_visual_mode());
    }

    #[test]
    fn is_visual_mode_returns_true_after_enter_visual_mode() {
        let mut view = LogView::new(make_commits(3));
        view.enter_visual_mode();
        assert!(view.is_visual_mode());
    }

    #[test]
    fn enter_visual_mode_sets_anchor_at_current_position() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Normal { pos: 2 };
        view.enter_visual_mode();
        assert!(view.is_visual_mode());
        assert_eq!(view.cursor_position(), 2);
        if let CursorState::Visual { pos, anchor } = view.cursor {
            assert_eq!(pos, 2);
            assert_eq!(anchor, 2);
        } else {
            panic!("expected Visual mode");
        }
    }

    #[test]
    fn exit_visual_mode_returns_to_normal_mode() {
        let mut view = LogView::new(make_commits(3));
        view.enter_visual_mode();
        assert!(view.is_visual_mode());
        view.exit_visual_mode();
        assert!(!view.is_visual_mode());
    }

    #[test]
    fn exit_visual_mode_preserves_cursor_position() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 3, anchor: 1 };
        view.exit_visual_mode();
        assert_eq!(view.cursor_position(), 3);
    }

    #[test]
    fn move_up_in_visual_mode_preserves_anchor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 3, anchor: 1 };
        view.move_up(1);
        assert_eq!(view.cursor_position(), 2);
        if let CursorState::Visual { pos, anchor } = view.cursor {
            assert_eq!(pos, 2);
            assert_eq!(anchor, 1);
        } else {
            panic!("expected Visual mode");
        }
    }

    #[test]
    fn move_down_in_visual_mode_preserves_anchor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 1, anchor: 3 };
        view.move_down(1);
        assert_eq!(view.cursor_position(), 2);
        if let CursorState::Visual { pos, anchor } = view.cursor {
            assert_eq!(pos, 2);
            assert_eq!(anchor, 3);
        } else {
            panic!("expected Visual mode");
        }
    }

    #[test]
    fn move_up_in_normal_mode_updates_cursor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Normal { pos: 2 };
        view.move_up(1);
        assert_eq!(view.cursor_position(), 1);
        assert!(!view.is_visual_mode());
    }

    #[test]
    fn move_down_in_normal_mode_updates_cursor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Normal { pos: 2 };
        view.move_down(1);
        assert_eq!(view.cursor_position(), 3);
        assert!(!view.is_visual_mode());
    }

    #[test]
    fn get_selected_range_returns_none_for_empty_commits() {
        let view = LogView::new(Vec::new());
        assert!(view.get_selected_range().is_none());
    }

    #[test]
    fn get_selected_range_normal_mode_returns_single_commit() {
        let view = LogView::new(make_commits(5));
        let range = view.get_selected_range().unwrap();
        assert_eq!(range.count, 1);
        assert_eq!(range.start.hash, "hash0");
        assert_eq!(range.end.hash, "hash0");
    }

    #[test]
    fn get_selected_range_visual_mode_returns_correct_range() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 1, anchor: 3 };
        let range = view.get_selected_range().unwrap();
        assert_eq!(range.count, 3);
        assert_eq!(range.start.hash, "hash2");
        assert_eq!(range.end.hash, "hash0");
    }

    #[test]
    fn get_selected_range_visual_mode_with_reversed_pos_anchor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 5, anchor: 3 };
        let range = view.get_selected_range().unwrap();
        assert_eq!(range.count, 3);
        assert_eq!(range.start.hash, "hash4");
        assert_eq!(range.end.hash, "hash2");
    }

    #[test]
    fn enter_visual_mode_does_nothing_on_empty_commits() {
        let mut view = LogView::new(Vec::new());
        view.enter_visual_mode();
        assert!(!view.is_visual_mode());
    }
}
