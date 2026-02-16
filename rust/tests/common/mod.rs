#![allow(unused_imports, dead_code)]

pub mod test_repo;

use crossterm::event::KeyCode;
pub use test_repo::{reset_counter, TestRepo};

use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{render, App};

#[allow(dead_code)]
pub struct Harness {
    terminal: Terminal<TestBackend>,
    app: App,
}

#[allow(dead_code)]
impl Harness {
    pub fn with_repo(repo: &TestRepo) -> Self {
        let backend = TestBackend::new(80, 24);
        let terminal = Terminal::new(backend).unwrap();
        let mut app = App::with_repo_path(repo.path());
        app.set_viewport_height(23);
        Self { terminal, app }
    }

    pub fn press(&mut self, key: KeyCode) -> &mut Self {
        let event = crossterm::event::KeyEvent::new(key, crossterm::event::KeyModifiers::NONE);
        self.app.handle_input(event);
        self
    }

    pub fn press_ctrl(&mut self, key: KeyCode) -> &mut Self {
        let event = crossterm::event::KeyEvent::new(key, crossterm::event::KeyModifiers::CONTROL);
        self.app.handle_input(event);
        self
    }

    pub fn snapshot(&mut self) -> String {
        self.terminal.draw(|f| render(f, &mut self.app)).unwrap();
        let buffer = self.terminal.backend().buffer();
        let mut lines = Vec::new();
        for y in 0..buffer.area.height {
            let mut line = String::new();
            for x in 0..buffer.area.width {
                let cell = buffer.cell((x, y)).unwrap();
                line.push_str(cell.symbol());
            }
            lines.push(format!("\"{}\"", line));
        }
        lines.join("\n")
    }

    pub fn selected(&self) -> usize {
        self.app.selected
    }

    pub fn scroll_offset(&self) -> usize {
        self.app.scroll_offset
    }
}
