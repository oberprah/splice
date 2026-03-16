use chrono::{DateTime, TimeZone, Utc};
use crossterm::event::{Event, KeyCode, KeyEvent, KeyModifiers};
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{action_from_event, git, render, Action, App, DiffRef, LogSpec};

use super::{snapshot::assert_snapshot, TestRepo};

pub struct Harness {
    terminal: Terminal<TestBackend>,
    app: App,
}

impl Harness {
    fn fixed_now() -> DateTime<Utc> {
        Utc.with_ymd_and_hms(2020, 1, 3, 0, 0, 0).unwrap()
    }

    pub fn with_repo(repo: &TestRepo) -> Self {
        Self::with_repo_and_log_spec_and_screen_size(repo, LogSpec::Head, 80, 24)
    }

    pub fn with_repo_and_screen_size(repo: &TestRepo, width: u16, height: u16) -> Self {
        Self::with_repo_and_log_spec_and_screen_size(repo, LogSpec::Head, width, height)
    }

    pub fn with_repo_and_log_spec(repo: &TestRepo, spec: LogSpec) -> Self {
        Self::with_repo_and_log_spec_and_screen_size(repo, spec, 80, 24)
    }

    pub fn with_repo_and_log_spec_and_screen_size(
        repo: &TestRepo,
        spec: LogSpec,
        width: u16,
        height: u16,
    ) -> Self {
        let backend = TestBackend::new(width, height);
        let terminal = Terminal::new(backend).unwrap();
        let mut app =
            App::with_repo_path_and_log_spec_and_now(repo.path(), spec, Self::fixed_now());
        app.set_viewport_size(height.saturating_sub(1) as usize, width as usize);
        Self { terminal, app }
    }

    pub fn with_diff_source(repo: &TestRepo, diff_ref: DiffRef) -> Result<Self, String> {
        Self::with_diff_source_and_screen_size(repo, diff_ref, 80, 24)
    }

    pub fn with_diff_source_and_screen_size(
        repo: &TestRepo,
        diff_ref: DiffRef,
        width: u16,
        height: u16,
    ) -> Result<Self, String> {
        let files = git::fetch_file_changes_for_ref(repo.path(), &diff_ref)?;
        let backend = TestBackend::new(width, height);
        let terminal = Terminal::new(backend).unwrap();
        let mut app = App::with_diff_source_and_now(
            repo.path().to_path_buf(),
            diff_ref,
            files,
            Self::fixed_now(),
        );
        app.set_viewport_size(height.saturating_sub(1) as usize, width as usize);
        Ok(Self { terminal, app })
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

    fn apply_event(&mut self, event: Event) -> bool {
        let action = action_from_event(event);
        if action != Action::None {
            self.app.update(action)
        } else {
            false
        }
    }

    pub fn should_exit(&mut self) -> bool {
        self.app.update(Action::Quit)
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

    pub fn assert_snapshot(&mut self, expected: &str) -> &mut Self {
        let actual = self.snapshot();
        assert_snapshot(&actual, expected);
        self
    }
}
