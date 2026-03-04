mod diff_view;
mod files_view;
mod log_view;

pub use diff_view::DiffView;
pub use files_view::FilesView;
pub use log_view::{LogSummary, LogView};

use std::path::PathBuf;
use std::process::Command;

use crate::core::{DiffSource, FileChange, LogSpec};
use crate::git;
use crate::input::Action;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum ThemeMode {
    #[default]
    Auto,
    Dark,
    Light,
}

pub enum View {
    Log(LogView),
    Files(FilesView),
    Diff(DiffView),
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum DiffEntryPoint {
    Top,
    FirstDiff,
    LastDiff,
}

pub struct App {
    pub repo_path: Option<PathBuf>,
    pub view: View,
    pub view_stack: Vec<View>,
    pub error: Option<String>,
    pub theme_mode: ThemeMode,
    viewport_width: usize,
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
            theme_mode: ThemeMode::Auto,
            viewport_width: 0,
        }
    }

    pub fn with_repo_path(path: impl Into<PathBuf>) -> Self {
        Self::with_repo_path_and_log_spec(path, LogSpec::Head)
    }

    pub fn with_repo_path_and_log_spec(path: impl Into<PathBuf>, spec: LogSpec) -> Self {
        let repo_path = path.into();
        match git::fetch_commits(&repo_path, spec) {
            Ok(commits) => Self {
                view: {
                    let mut log = LogView::new(commits);
                    if let Ok((uncommitted_type, file_count)) =
                        git::fetch_uncommitted_summary(&repo_path)
                    {
                        log.set_summary(LogSummary {
                            uncommitted_type,
                            file_count,
                        });
                    }
                    View::Log(log)
                },
                repo_path: Some(repo_path),
                view_stack: Vec::new(),
                error: None,
                theme_mode: ThemeMode::Auto,
                viewport_width: 0,
            },
            Err(e) => Self {
                repo_path: Some(repo_path),
                view: View::Log(LogView::new(Vec::new())),
                view_stack: Vec::new(),
                error: Some(e),
                theme_mode: ThemeMode::Auto,
                viewport_width: 0,
            },
        }
    }

    pub fn with_diff_source(
        repo_path: PathBuf,
        source: DiffSource,
        files: Vec<FileChange>,
    ) -> Self {
        let files_view = FilesView::new(source, files);
        Self {
            repo_path: Some(repo_path),
            view: View::Files(files_view),
            view_stack: vec![],
            error: None,
            theme_mode: ThemeMode::Auto,
            viewport_width: 0,
        }
    }

    pub fn set_theme_mode(&mut self, theme_mode: ThemeMode) {
        self.theme_mode = theme_mode;
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        match &mut self.view {
            View::Log(log) => log.set_viewport_height(height),
            View::Files(files) => files.set_viewport_height(height),
            View::Diff(diff) => diff.set_viewport_dimensions(height, self.viewport_width),
        }
    }

    pub fn set_viewport_size(&mut self, height: usize, width: usize) {
        self.viewport_width = width;
        match &mut self.view {
            View::Log(log) => log.set_viewport_height(height),
            View::Files(files) => files.set_viewport_height(height),
            View::Diff(diff) => diff.set_viewport_dimensions(height, width),
        }
    }

    pub fn update(&mut self, action: Action) -> bool {
        match action {
            Action::Quit => {
                if let View::Log(log) = &mut self.view {
                    if log.is_visual_mode() {
                        log.exit_visual_mode();
                        return false;
                    }
                }
                return true;
            }
            Action::Back => {
                if let View::Log(log) = &mut self.view {
                    if log.is_visual_mode() {
                        log.exit_visual_mode();
                    } else if self.go_back() {
                        return true;
                    }
                } else if self.go_back() {
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
            Action::ToggleFolder => self.toggle_folder(false, false),
            Action::ExpandFolder => self.toggle_folder(true, false),
            Action::CollapseFolder => self.toggle_folder(false, true),
            Action::ToggleVisualMode => {
                if let View::Log(log) = &mut self.view {
                    if log.is_visual_mode() {
                        log.exit_visual_mode();
                    } else {
                        log.enter_visual_mode();
                    }
                }
            }
            Action::NextDiff => self.navigate_diff(1),
            Action::PrevDiff => self.navigate_diff(-1),
            Action::OpenInEditor => {}
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

    fn viewport_height(&self) -> usize {
        match &self.view {
            View::Log(log) => log.viewport_height,
            View::Files(files) => files.viewport_height,
            View::Diff(diff) => diff.viewport_height,
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
            if let Some(uncommitted_type) = log.selected_uncommitted_type() {
                let repo_path = match &self.repo_path {
                    Some(p) => p.clone(),
                    None => return,
                };

                let source = DiffSource::Uncommitted(uncommitted_type);
                match git::fetch_file_changes_for_source(&repo_path, &source) {
                    Ok(files) => {
                        let files_view = FilesView::new(source, files);
                        let old_view = std::mem::replace(&mut self.view, View::Files(files_view));
                        self.view_stack.push(old_view);
                    }
                    Err(e) => {
                        self.error = Some(e);
                    }
                }
            } else if let Some(range) = log.get_selected_range() {
                let repo_path = match &self.repo_path {
                    Some(p) => p.clone(),
                    None => return,
                };

                match git::fetch_file_changes(&repo_path, &range) {
                    Ok(files) => {
                        let files_view = FilesView::new(DiffSource::CommitRange(range), files);
                        let old_view = std::mem::replace(&mut self.view, View::Files(files_view));
                        self.view_stack.push(old_view);
                    }
                    Err(e) => {
                        self.error = Some(e);
                    }
                }
            }
        } else if let View::Files(files) = &mut self.view {
            if files.selected_is_folder() {
                files.toggle_folder(false, false);
            } else if let Some(file) = files.selected_file() {
                let file = file.clone();
                let source = files.source.clone();
                self.open_file_diff(source, file, true, DiffEntryPoint::Top);
            }
        }
    }

    fn toggle_folder(&mut self, expand_only: bool, collapse_only: bool) {
        if let View::Files(files) = &mut self.view {
            files.toggle_folder(expand_only, collapse_only);
        }
    }

    fn navigate_diff(&mut self, direction: isize) {
        let moved_in_file = match &mut self.view {
            View::Diff(diff) => {
                if direction > 0 {
                    diff.navigate_next_diff()
                } else {
                    diff.navigate_prev_diff()
                }
            }
            _ => false,
        };

        if !moved_in_file {
            self.open_adjacent_diff_file(direction);
        }
    }

    fn open_adjacent_diff_file(&mut self, direction: isize) {
        let (source, current_path) = match &self.view {
            View::Diff(diff) => (diff.source.clone(), diff.file.path.clone()),
            _ => return,
        };

        let Some(file) = self.adjacent_file(&current_path, direction) else {
            return;
        };

        let entry = if direction > 0 {
            DiffEntryPoint::FirstDiff
        } else {
            DiffEntryPoint::LastDiff
        };

        if self.open_file_diff(source, file.clone(), false, entry) {
            self.sync_files_selection(&file.path);
        }
    }

    fn adjacent_file(&self, current_path: &str, direction: isize) -> Option<FileChange> {
        let files = self.view_stack.iter().rev().find_map(|view| match view {
            View::Files(files) => Some(files),
            _ => None,
        })?;

        files.adjacent_visible_file(current_path, direction)
    }

    fn sync_files_selection(&mut self, path: &str) {
        for view in self.view_stack.iter_mut().rev() {
            if let View::Files(files) = view {
                files.select_file_path(path);
                return;
            }
        }
    }

    fn open_file_diff(
        &mut self,
        source: DiffSource,
        file: FileChange,
        push_previous: bool,
        entry_point: DiffEntryPoint,
    ) -> bool {
        let repo_path = match &self.repo_path {
            Some(path) => path.clone(),
            None => return false,
        };

        let full_diff = match git::fetch_full_file_diff_for_source(&repo_path, &source, &file.path)
        {
            Ok(diff) => diff,
            Err(e) => {
                self.error = Some(e);
                return false;
            }
        };

        let meta = crate::domain::diff::DiffMeta {
            path: file.path.clone(),
            additions: file.additions,
            deletions: file.deletions,
        };

        let diff = match crate::domain::diff::build_file_diff_full(
            meta,
            &full_diff.old_content,
            &full_diff.new_content,
            &full_diff.diff_output,
        ) {
            Ok(diff) => diff,
            Err(e) => {
                self.error = Some(e);
                return false;
            }
        };

        let highlights = crate::domain::highlight::highlight_diff_sides(
            &file.path,
            &full_diff.old_content,
            &full_diff.new_content,
        );

        let mut diff_view = DiffView::new(source, file, diff, highlights);
        diff_view.set_viewport_dimensions(self.viewport_height(), self.viewport_width);
        let old_view = std::mem::replace(&mut self.view, View::Diff(diff_view));
        if push_previous {
            self.view_stack.push(old_view);
        }

        if let View::Diff(diff) = &mut self.view {
            match entry_point {
                DiffEntryPoint::Top => {}
                DiffEntryPoint::FirstDiff => {
                    diff.jump_to_first_diff();
                }
                DiffEntryPoint::LastDiff => {
                    diff.jump_to_last_diff();
                }
            }
        }

        true
    }

    pub fn open_current_diff_in_editor(&self) -> Result<(), String> {
        let View::Diff(diff) = &self.view else {
            return Ok(());
        };

        if diff.file.is_binary {
            return Err("cannot open binary file in editor".to_string());
        }

        if diff.file.status == crate::core::FileStatus::Deleted {
            return Err("cannot open: file has been deleted".to_string());
        }

        let editor = configured_editor()
            .ok_or_else(|| "no editor configured (set $EDITOR or $VISUAL)".to_string())?;
        let line = diff
            .current_file_line_number()
            .map_err(|err| format!("failed to determine line number: {err}"))?;

        let repo_path = self
            .repo_path
            .as_ref()
            .ok_or_else(|| "missing repository path".to_string())?;
        let repo_root = git::repository_root(repo_path)
            .map_err(|err| format!("failed to determine repository root: {err}"))?;
        let absolute_path = repo_root.join(&diff.file.path);

        if !absolute_path.exists() {
            return Err("cannot open: file not found".to_string());
        }

        let status = Command::new(editor)
            .arg(format!("+{line}"))
            .arg(&absolute_path)
            .status()
            .map_err(|err| format!("failed to launch editor: {err}"))?;

        if status.success() {
            Ok(())
        } else {
            Err(format!("editor exited with status: {status}"))
        }
    }
}

fn configured_editor() -> Option<String> {
    std::env::var("EDITOR")
        .ok()
        .filter(|value| !value.is_empty())
        .or_else(|| {
            std::env::var("VISUAL")
                .ok()
                .filter(|value| !value.is_empty())
        })
}
