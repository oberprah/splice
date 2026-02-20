mod diff_view;
mod files_view;
mod log_view;

pub use diff_view::DiffView;
pub use files_view::FilesView;
pub use log_view::LogView;

use std::path::PathBuf;

use crate::git;
use crate::input::Action;

pub enum View {
    Log(LogView),
    Files(FilesView),
    Diff(DiffView),
}

pub struct App {
    pub repo_path: Option<PathBuf>,
    pub view: View,
    pub view_stack: Vec<View>,
    pub error: Option<String>,
}

impl Default for App {
    fn default() -> Self {
        Self::new()
    }
}

impl App {
    pub fn new() -> Self {
        Self {
            repo_path: None,
            view: View::Log(LogView::new(Vec::new())),
            view_stack: Vec::new(),
            error: None,
        }
    }

    pub fn with_repo_path(path: impl Into<PathBuf>) -> Self {
        let repo_path = path.into();
        match git::fetch_commits(&repo_path) {
            Ok(commits) => Self {
                repo_path: Some(repo_path),
                view: View::Log(LogView::new(commits)),
                view_stack: Vec::new(),
                error: None,
            },
            Err(e) => Self {
                repo_path: Some(repo_path),
                view: View::Log(LogView::new(Vec::new())),
                view_stack: Vec::new(),
                error: Some(e),
            },
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        match &mut self.view {
            View::Log(log) => log.set_viewport_height(height),
            View::Files(files) => files.set_viewport_height(height),
            View::Diff(diff) => diff.set_viewport_height(height),
        }
    }

    pub fn update(&mut self, action: Action) -> bool {
        match action {
            Action::Quit => return true,
            Action::Back => {
                if self.go_back() {
                    return true;
                }
            }
            Action::Open => self.open_selected(),
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

    fn page_step(&self) -> usize {
        match &self.view {
            View::Log(log) => log.page_step(),
            View::Files(files) => files.page_step(),
            View::Diff(diff) => diff.page_step(),
        }
    }

    fn move_down(&mut self, amount: usize) {
        match &mut self.view {
            View::Log(log) => log.move_down(amount),
            View::Files(files) => files.move_down(amount),
            View::Diff(diff) => diff.move_down(amount),
        }
    }

    fn move_up(&mut self, amount: usize) {
        match &mut self.view {
            View::Log(log) => log.move_up(amount),
            View::Files(files) => files.move_up(amount),
            View::Diff(diff) => diff.move_up(amount),
        }
    }

    fn go_back(&mut self) -> bool {
        if let Some(prev_view) = self.view_stack.pop() {
            self.view = prev_view;
            false
        } else {
            true
        }
    }

    fn open_selected(&mut self) {
        if let View::Log(log) = &self.view {
            if let Some(commit) = log.selected_commit() {
                let repo_path = match &self.repo_path {
                    Some(p) => p.clone(),
                    None => return,
                };

                match git::fetch_file_changes(&repo_path, &commit.hash) {
                    Ok(files) => {
                        let files_view = FilesView::new(commit.clone(), files);
                        let old_view = std::mem::replace(&mut self.view, View::Files(files_view));
                        self.view_stack.push(old_view);
                    }
                    Err(e) => {
                        self.error = Some(e);
                    }
                }
            }
        } else if let View::Files(files) = &self.view {
            let file = match files.selected_file() {
                Some(file) => file.clone(),
                None => return,
            };

            let repo_path = match &self.repo_path {
                Some(p) => p.clone(),
                None => return,
            };

            match git::fetch_file_diff(&repo_path, &files.commit.hash, &file.path) {
                Ok(diff_output) => {
                    let meta = crate::domain::diff::DiffMeta {
                        path: file.path.clone(),
                        additions: file.additions,
                        deletions: file.deletions,
                    };
                    match crate::domain::diff::build_file_diff(meta, &diff_output) {
                        Ok(diff) => {
                            let diff_view = DiffView::new(files.commit.clone(), file, diff);
                            let old_view = std::mem::replace(&mut self.view, View::Diff(diff_view));
                            self.view_stack.push(old_view);
                        }
                        Err(e) => {
                            self.error = Some(e);
                        }
                    }
                }
                Err(e) => {
                    self.error = Some(e);
                }
            }
        }
    }
}
