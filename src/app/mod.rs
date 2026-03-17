mod diff_view;
mod files_view;
mod log_view;
pub mod viewport;

pub use diff_view::DiffView;
pub use files_view::FilesView;
pub use log_view::{LogSummary, LogView};
pub use viewport::{Viewport, ViewportAction, VisibleContent};

use std::path::PathBuf;
use std::process::Command;

use chrono::{DateTime, Utc};

use crate::core::{DiffRef, FileDiffInfo, LogSpec};
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
    now: Option<DateTime<Utc>>,
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
            now: None,
        }
    }

    pub fn new_with_now(now: DateTime<Utc>) -> Self {
        Self {
            repo_path: None,
            view: View::Log(LogView::new(Vec::new())),
            view_stack: Vec::new(),
            error: None,
            theme_mode: ThemeMode::Auto,
            viewport_width: 0,
            now: Some(now),
        }
    }

    pub fn with_repo_path(path: impl Into<PathBuf>) -> Self {
        Self::with_repo_path_and_log_spec_with_now(path, LogSpec::Head, None)
    }

    pub fn with_repo_path_and_log_spec(path: impl Into<PathBuf>, spec: LogSpec) -> Self {
        Self::with_repo_path_and_log_spec_with_now(path, spec, None)
    }

    pub fn with_repo_path_and_log_spec_and_now(
        path: impl Into<PathBuf>,
        spec: LogSpec,
        now: DateTime<Utc>,
    ) -> Self {
        Self::with_repo_path_and_log_spec_with_now(path, spec, Some(now))
    }

    fn with_repo_path_and_log_spec_with_now(
        path: impl Into<PathBuf>,
        spec: LogSpec,
        now: Option<DateTime<Utc>>,
    ) -> Self {
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
                now,
            },
            Err(e) => Self {
                repo_path: Some(repo_path),
                view: View::Log(LogView::new(Vec::new())),
                view_stack: Vec::new(),
                error: Some(e),
                theme_mode: ThemeMode::Auto,
                viewport_width: 0,
                now,
            },
        }
    }

    pub fn with_diff_source(
        repo_path: PathBuf,
        diff_ref: DiffRef,
        files: Vec<FileDiffInfo>,
    ) -> Self {
        Self::with_diff_source_with_now(repo_path, diff_ref, files, None)
    }

    pub fn with_diff_source_and_now(
        repo_path: PathBuf,
        diff_ref: DiffRef,
        files: Vec<FileDiffInfo>,
        now: DateTime<Utc>,
    ) -> Self {
        Self::with_diff_source_with_now(repo_path, diff_ref, files, Some(now))
    }

    fn with_diff_source_with_now(
        repo_path: PathBuf,
        diff_ref: DiffRef,
        files: Vec<FileDiffInfo>,
        now: Option<DateTime<Utc>>,
    ) -> Self {
        let files_view = FilesView::new(diff_ref, files);
        Self {
            repo_path: Some(repo_path),
            view: View::Files(files_view),
            view_stack: vec![],
            error: None,
            theme_mode: ThemeMode::Auto,
            viewport_width: 0,
            now,
        }
    }

    /// Returns `true` when the diff view has a scroll animation in progress.
    pub fn is_animating(&self) -> bool {
        matches!(&self.view, View::Diff(diff) if diff.is_animating())
    }

    /// Advance the scroll animation by one frame. Returns `true` if the
    /// scroll offset changed (caller should re-render).
    pub fn advance_scroll_animation(&mut self) -> bool {
        match &mut self.view {
            View::Diff(diff) => diff.advance_animation(),
            _ => false,
        }
    }

    /// Instantly complete any running scroll animation.
    pub fn settle_animation(&mut self) {
        if let View::Diff(diff) = &mut self.view {
            diff.settle_animation();
        }
    }

    pub fn set_theme_mode(&mut self, theme_mode: ThemeMode) {
        self.theme_mode = theme_mode;
    }

    pub fn set_now(&mut self, now: DateTime<Utc>) {
        self.now = Some(now);
    }

    pub fn now(&self) -> DateTime<Utc> {
        self.now.unwrap_or_else(Utc::now)
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
        if self.error.take().is_some() {
            return false;
        }

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
            Action::ToggleMessage => {
                if let View::Files(files) = &mut self.view {
                    files.toggle_message();
                }
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

    fn viewport_height(&self) -> usize {
        match &self.view {
            View::Log(log) => log.viewport_height,
            View::Files(files) => files.list_viewport_height(),
            View::Diff(diff) => diff.viewport_height(),
        }
    }

    fn move_down(&mut self, amount: usize) {
        match &mut self.view {
            View::Log(log) => log.move_down(amount),
            View::Files(files) => files.move_down(amount),
            View::Diff(diff) => {
                diff.update(ViewportAction::ScrollDown(amount));
            }
        }
    }

    fn move_up(&mut self, amount: usize) {
        match &mut self.view {
            View::Log(log) => log.move_up(amount),
            View::Files(files) => files.move_up(amount),
            View::Diff(diff) => {
                diff.update(ViewportAction::ScrollUp(amount));
            }
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

                let diff_ref = DiffRef::Uncommitted(uncommitted_type);
                match git::fetch_file_changes_for_ref(&repo_path, &diff_ref) {
                    Ok(files) => {
                        let files_view = FilesView::new(diff_ref, files);
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
                        let files_view = FilesView::new(DiffRef::CommitRange(range), files);
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
                let diff_ref = files.diff_ref.clone();
                self.open_file_diff(diff_ref, file, true, DiffEntryPoint::Top);
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
                    diff.update(ViewportAction::NextHunk)
                } else {
                    diff.update(ViewportAction::PrevHunk)
                }
            }
            _ => false,
        };

        if !moved_in_file {
            self.open_adjacent_diff_file(direction);
        }
    }

    fn open_adjacent_diff_file(&mut self, direction: isize) {
        let (diff_ref, current_path) = match &self.view {
            View::Diff(diff) => (diff.diff_ref.clone(), diff.file.info.path.clone()),
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

        if self.open_file_diff(diff_ref, file.clone(), false, entry) {
            self.sync_files_selection(&file.path);
        }
    }

    fn adjacent_file(&self, current_path: &str, direction: isize) -> Option<FileDiffInfo> {
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
        diff_ref: DiffRef,
        file: FileDiffInfo,
        push_previous: bool,
        entry_point: DiffEntryPoint,
    ) -> bool {
        let repo_path = match &self.repo_path {
            Some(path) => path.clone(),
            None => return false,
        };

        let full_diff = match git::fetch_full_file_diff_for_ref(
            &repo_path,
            &diff_ref,
            &file.path,
            file.old_path.as_deref(),
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

        let file_diff = match crate::domain::diff::build_file_diff(
            file,
            &full_diff.old_content,
            &full_diff.new_content,
            &full_diff.diff_output,
            &highlights.old,
            &highlights.new,
        ) {
            Ok(diff) => diff,
            Err(e) => {
                self.error = Some(e);
                return false;
            }
        };

        let mut diff_view = DiffView::new(diff_ref, file_diff);
        diff_view.set_viewport_dimensions(self.viewport_height(), self.viewport_width);
        let old_view = std::mem::replace(&mut self.view, View::Diff(diff_view));
        if push_previous {
            self.view_stack.push(old_view);
        }

        if let View::Diff(diff) = &mut self.view {
            match entry_point {
                DiffEntryPoint::Top => {}
                DiffEntryPoint::FirstDiff => {
                    diff.update(ViewportAction::JumpToFirstHunk);
                    diff.settle_animation();
                }
                DiffEntryPoint::LastDiff => {
                    diff.update(ViewportAction::JumpToLastHunk);
                    diff.settle_animation();
                }
            }
        }

        true
    }

    pub fn open_current_diff_in_editor(&self) -> Result<(), String> {
        let View::Diff(diff) = &self.view else {
            return Ok(());
        };

        if diff.file.info.is_binary {
            return Err("cannot open binary file in editor".to_string());
        }

        if diff.file.info.status == crate::core::FileStatus::Deleted {
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
        let absolute_path = repo_root.join(&diff.file.info.path);

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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::input::Action;

    #[test]
    fn error_is_cleared_on_any_keypress() {
        let mut app = App::new();
        app.error = Some("some error".to_string());

        let quit = app.update(Action::MoveDown);

        assert!(!quit, "should not quit");
        assert!(
            app.error.is_none(),
            "error should be cleared after keypress"
        );
    }

    #[test]
    fn error_clears_without_performing_action() {
        let mut app = App::new();
        app.set_viewport_size(10, 80);
        app.error = Some("some error".to_string());

        // Quit action should be consumed to dismiss the error, not actually quit
        let quit = app.update(Action::Quit);

        assert!(
            !quit,
            "should not quit — keypress should only dismiss error"
        );
        assert!(app.error.is_none());
    }
}
