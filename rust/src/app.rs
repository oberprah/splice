use std::path::PathBuf;

use crate::core::Commit;
use crate::git;
use crate::input::Action;

pub struct App {
    pub repo_path: Option<PathBuf>,
    pub commits: Vec<Commit>,
    pub selected: usize,
    pub scroll_offset: usize,
    pub viewport_height: usize,
    pub error: Option<String>,
}

impl App {
    pub fn new() -> Self {
        Self {
            repo_path: None,
            commits: Vec::new(),
            selected: 0,
            scroll_offset: 0,
            viewport_height: 0,
            error: None,
        }
    }

    pub fn with_repo_path(path: impl Into<PathBuf>) -> Self {
        let repo_path = path.into();
        match git::fetch_commits(&repo_path) {
            Ok(commits) => Self {
                repo_path: Some(repo_path),
                commits,
                selected: 0,
                scroll_offset: 0,
                viewport_height: 0,
                error: None,
            },
            Err(e) => Self {
                repo_path: Some(repo_path),
                commits: Vec::new(),
                selected: 0,
                scroll_offset: 0,
                viewport_height: 0,
                error: Some(e),
            },
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

    pub fn update(&mut self, action: Action) -> bool {
        match action {
            Action::Quit => return true,
            Action::MoveDown => self.move_down(1),
            Action::MoveUp => self.move_up(1),
            Action::PageDown => {
                let step = self.page_step();
                self.move_down(step);
            }
            Action::PageUp => {
                let step = self.page_step();
                self.move_up(step);
            }
            Action::Resize { .. } | Action::None => {}
        }

        false
    }

    fn move_down(&mut self, amount: usize) {
        if self.commits.is_empty() {
            return;
        }
        let last = self.commits.len().saturating_sub(1);
        self.selected = self.selected.saturating_add(amount).min(last);
        self.clamp_scroll_offset();
    }

    fn move_up(&mut self, amount: usize) {
        if self.commits.is_empty() {
            return;
        }
        self.selected = self.selected.saturating_sub(amount);
        self.clamp_scroll_offset();
    }

    fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }
}
