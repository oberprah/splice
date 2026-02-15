use std::path::PathBuf;

use crate::core::Commit;

const LOG_AREA_X: u16 = 2;
const LOG_AREA_Y: u16 = 3;
const LOG_AREA_RIGHT_MARGIN: u16 = 4;
const LOG_AREA_BOTTOM_MARGIN: u16 = 5;
use crate::git;
use crate::ui::render_log_view;
use crossterm::event::{self, KeyCode, KeyModifiers};
use ratatui::prelude::*;

pub struct App {
    pub repo_path: Option<PathBuf>,
    pub commits: Vec<Commit>,
    pub selected: usize,
    pub error: Option<String>,
}

impl App {
    pub fn new() -> Self {
        Self {
            repo_path: None,
            commits: Vec::new(),
            selected: 0,
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
                error: None,
            },
            Err(e) => Self {
                repo_path: Some(repo_path),
                commits: Vec::new(),
                selected: 0,
                error: Some(e),
            },
        }
    }

    pub fn handle_input(&mut self, key: event::KeyEvent) -> bool {
        match key.code {
            KeyCode::Char('q') => return true,
            KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => return true,
            KeyCode::Down | KeyCode::Char('j') => {
                if self.selected < self.commits.len().saturating_sub(1) {
                    self.selected += 1;
                }
            }
            KeyCode::Up | KeyCode::Char('k') => {
                if self.selected > 0 {
                    self.selected -= 1;
                }
            }
            KeyCode::Char('G') => {
                self.selected = self.commits.len().saturating_sub(1);
            }
            KeyCode::Char('g') => {
                self.selected = 0;
            }
            _ => {}
        }
        false
    }
}

pub fn render(f: &mut Frame, app: &App) {
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
    render_log_view(f, &app.commits, app.selected, log_area);
}
