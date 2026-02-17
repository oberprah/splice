use crate::core::Commit;
use crate::domain::graph::{compute_layout, GraphCommit, Layout};

pub struct LogView {
    pub commits: Vec<Commit>,
    pub graph_layout: Layout,
    pub selected: usize,
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
        if self.commits.is_empty() {
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
        if self.commits.is_empty() {
            return;
        }
        let last = self.commits.len().saturating_sub(1);
        self.selected = self.selected.saturating_add(amount).min(last);
        self.clamp_scroll_offset();
    }

    pub fn move_up(&mut self, amount: usize) {
        if self.commits.is_empty() {
            return;
        }
        self.selected = self.selected.saturating_sub(amount);
        self.clamp_scroll_offset();
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    pub fn selected_commit(&self) -> Option<&Commit> {
        self.commits.get(self.selected)
    }
}
