use crate::core::{selection_range, Commit, CommitRange, CursorState};
use crate::domain::graph::{compute_layout, GraphCommit, Layout};

pub struct LogView {
    pub commits: Vec<Commit>,
    pub graph_layout: Layout,
    pub cursor: CursorState,
    pub scroll_offset: usize,
    pub viewport_height: usize,
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
        Self {
            commits,
            graph_layout,
            cursor: CursorState::Normal { pos: 0 },
            scroll_offset: 0,
            viewport_height: 0,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
    }

    fn clamp_scroll_offset(&mut self) {
        if self.commits.is_empty() {
            self.cursor = CursorState::Normal { pos: 0 };
            self.scroll_offset = 0;
            return;
        }
        let pos = self.cursor.position();
        if pos < self.scroll_offset {
            self.scroll_offset = pos;
        } else if self.viewport_height > 0 && pos >= self.scroll_offset + self.viewport_height {
            self.scroll_offset = pos - self.viewport_height + 1;
        }
    }

    pub fn move_down(&mut self, amount: usize) {
        if self.commits.is_empty() {
            return;
        }
        let last = self.commits.len().saturating_sub(1);
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
        if self.commits.is_empty() {
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
        self.commits.get(self.cursor.position())
    }

    pub fn is_visual_mode(&self) -> bool {
        matches!(self.cursor, CursorState::Visual { .. })
    }

    pub fn enter_visual_mode(&mut self) {
        if self.commits.is_empty() {
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
        if self.commits.is_empty() {
            return None;
        }
        let (min, max) = selection_range(&self.cursor);
        let end = self.commits.get(min)?.clone();
        let start = self.commits.get(max)?.clone();
        let count = max - min + 1;
        Some(CommitRange {
            start,
            end,
            count,
            include_start: true,
        })
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
    fn new_initializes_cursor_to_normal_pos_0() {
        let view = LogView::new(make_commits(3));
        assert_eq!(view.cursor_position(), 0);
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
        view.cursor = CursorState::Visual { pos: 0, anchor: 2 };
        let range = view.get_selected_range().unwrap();
        assert_eq!(range.count, 3);
        assert_eq!(range.start.hash, "hash2");
        assert_eq!(range.end.hash, "hash0");
    }

    #[test]
    fn get_selected_range_visual_mode_with_reversed_pos_anchor() {
        let mut view = LogView::new(make_commits(5));
        view.cursor = CursorState::Visual { pos: 4, anchor: 2 };
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
