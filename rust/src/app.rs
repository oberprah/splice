use std::path::PathBuf;

use crate::core::Commit;

const LOG_AREA_X: u16 = 2;
const LOG_AREA_Y: u16 = 0;
const LOG_AREA_RIGHT_MARGIN: u16 = 4;
const LOG_AREA_BOTTOM_MARGIN: u16 = 0;
use crate::git;
use crate::ui::render_log_view;
use crossterm::event::{self, KeyCode, KeyModifiers};
use ratatui::prelude::*;

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
        self.viewport_height = height.saturating_sub(1);
        self.clamp_scroll_offset();
    }

    fn clamp_scroll_offset(&mut self) {
        if self.selected < self.scroll_offset {
            self.scroll_offset = self.selected;
        } else if self.viewport_height > 0 && self.selected >= self.scroll_offset + self.viewport_height {
            self.scroll_offset = self.selected - self.viewport_height + 1;
        }
    }

    pub fn handle_input(&mut self, key: event::KeyEvent) -> bool {
        match key.code {
            KeyCode::Char('q') => return true,
            KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => return true,
            KeyCode::Down | KeyCode::Char('j') => {
                if self.selected < self.commits.len().saturating_sub(1) {
                    self.selected += 1;
                    self.clamp_scroll_offset();
                }
            }
            KeyCode::Up | KeyCode::Char('k') => {
                if self.selected > 0 {
                    self.selected -= 1;
                    self.clamp_scroll_offset();
                }
            }
            KeyCode::Char('d') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                let half = (self.viewport_height / 2).max(1);
                let new_selected = self.selected.saturating_add(half).min(self.commits.len().saturating_sub(1));
                self.selected = new_selected;
                self.clamp_scroll_offset();
            }
            KeyCode::Char('u') if key.modifiers.contains(KeyModifiers::CONTROL) => {
                let half = (self.viewport_height / 2).max(1);
                self.selected = self.selected.saturating_sub(half);
                self.clamp_scroll_offset();
            }
            _ => {}
        }
        false
    }
}

pub fn render(f: &mut Frame, app: &mut App) {
    let size = f.area();

    if let Some(ref error) = app.error {
        let msg = ratatui::widgets::Paragraph::new(format!("Error: {}", error))
            .style(Style::default().fg(Color::Red))
            .alignment(Alignment::Center);
        f.render_widget(msg, size);
        return;
    }

    let log_area = Rect::new(
        LOG_AREA_X,
        LOG_AREA_Y,
        size.width.saturating_sub(LOG_AREA_RIGHT_MARGIN),
        size.height.saturating_sub(LOG_AREA_BOTTOM_MARGIN),
    );
    app.set_viewport_height(log_area.height as usize);
    render_log_view(f, &app.commits, app.selected, app.scroll_offset, log_area);
}
