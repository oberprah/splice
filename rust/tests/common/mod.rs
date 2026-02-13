use crossterm::event::KeyCode;
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{render, App};

pub struct Harness {
    terminal: Terminal<TestBackend>,
    app: App,
}

impl Harness {
    pub fn new() -> Self {
        let backend = TestBackend::new(80, 24);
        let terminal = Terminal::new(backend).unwrap();
        let app = App::new();
        Self { terminal, app }
    }

    pub fn press(&mut self, key: KeyCode) -> &mut Self {
        let event = crossterm::event::KeyEvent::new(key, crossterm::event::KeyModifiers::NONE);
        self.app.handle_input(event);
        self
    }

    pub fn snapshot(&mut self) -> &TestBackend {
        self.terminal.draw(|f| render(f, &self.app)).unwrap();
        self.terminal.backend()
    }
}

#[macro_export]
macro_rules! assert_snapshot {
    ($h:expr, @$snapshot:literal) => {
        insta::assert_snapshot!($h.snapshot(), @$snapshot);
    };
}
