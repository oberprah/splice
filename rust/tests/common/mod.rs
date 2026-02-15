mod test_repo;

use crossterm::event::KeyCode;
pub use test_repo::TestRepo;

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
        let app = App::with_repo_path(repo.path());
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
