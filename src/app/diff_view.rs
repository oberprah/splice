use crate::core::{Commit, FileChange};
use crate::domain::diff::FileDiff;

pub struct DiffView {
    pub commit: Commit,
    pub file: FileChange,
    pub diff: FileDiff,
    pub scroll_offset: usize,
    pub viewport_height: usize,
}

impl DiffView {
    pub fn new(commit: Commit, file: FileChange, diff: FileDiff) -> Self {
        Self {
            commit,
            file,
            diff,
            scroll_offset: 0,
            viewport_height: 0,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
    }

    pub fn move_down(&mut self, amount: usize) {
        let max_scroll = self.max_scroll_offset();
        self.scroll_offset = self.scroll_offset.saturating_add(amount).min(max_scroll);
    }

    pub fn move_up(&mut self, amount: usize) {
        self.scroll_offset = self.scroll_offset.saturating_sub(amount);
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    fn clamp_scroll_offset(&mut self) {
        let max_scroll = self.max_scroll_offset();
        if self.scroll_offset > max_scroll {
            self.scroll_offset = max_scroll;
        }
    }

    fn max_scroll_offset(&self) -> usize {
        let total = self.diff.total_line_count();
        if total <= self.viewport_height {
            0
        } else {
            total - self.viewport_height
        }
    }
}
