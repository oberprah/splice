use crate::core::{Commit, FileChange};

pub struct FilesView {
    pub commit: Commit,
    pub files: Vec<FileChange>,
    pub selected: usize,
    pub scroll_offset: usize,
    pub viewport_height: usize,
}

impl FilesView {
    pub fn new(commit: Commit, files: Vec<FileChange>) -> Self {
        Self {
            commit,
            files,
            selected: 0,
            scroll_offset: 0,
            viewport_height: 0,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
    }

    fn clamp_scroll_offset(&mut self) {
        if self.files.is_empty() {
            self.selected = 0;
            self.scroll_offset = 0;
            return;
        }
        if self.selected < self.scroll_offset {
            self.scroll_offset = self.selected;
        } else if self.viewport_height > 0
            && self.selected >= self.scroll_offset + self.viewport_height
        {
            self.scroll_offset = self.selected - self.viewport_height + 1;
        }
    }

    pub fn move_down(&mut self, amount: usize) {
        if self.files.is_empty() {
            return;
        }
        let last = self.files.len().saturating_sub(1);
        self.selected = self.selected.saturating_add(amount).min(last);
        self.clamp_scroll_offset();
    }

    pub fn move_up(&mut self, amount: usize) {
        if self.files.is_empty() {
            return;
        }
        self.selected = self.selected.saturating_sub(amount);
        self.clamp_scroll_offset();
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    pub fn selected_file(&self) -> Option<&FileChange> {
        self.files.get(self.selected)
    }

    pub fn total_additions(&self) -> u32 {
        self.files.iter().map(|f| f.additions).sum()
    }

    pub fn total_deletions(&self) -> u32 {
        self.files.iter().map(|f| f.deletions).sum()
    }
}
