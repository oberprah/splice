use crossterm::event::{Event, KeyCode, KeyEvent, KeyModifiers};
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{action_from_event, render, Action, App};

use super::{snapshot::assert_snapshot, TestRepo};

pub struct Harness {
    terminal: Terminal<TestBackend>,
    app: App,
}

impl Harness {
    pub fn with_repo(repo: &TestRepo) -> Self {
        let backend = TestBackend::new(80, 24);
        let terminal = Terminal::new(backend).unwrap();
        let mut app = App::with_repo_path(repo.path());
        app.set_viewport_height(23);
        Self { terminal, app }
    }

    pub fn press(&mut self, key: KeyCode) -> &mut Self {
        let event = Event::Key(KeyEvent::new(key, KeyModifiers::NONE));
        self.apply_event(event);
        self
    }

    pub fn press_ctrl(&mut self, key: KeyCode) -> &mut Self {
        let event = Event::Key(KeyEvent::new(key, KeyModifiers::CONTROL));
        self.apply_event(event);
        self
    }

    fn apply_event(&mut self, event: Event) {
        let action = action_from_event(event);
        if action != Action::None {
            self.app.update(action);
        }
    }

    fn snapshot(&mut self) -> String {
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

    pub fn assert_snapshot(&mut self, expected: &str) -> &mut Self {
        let actual = self.snapshot();
        assert_snapshot(&actual, expected);
        self
    }
}
