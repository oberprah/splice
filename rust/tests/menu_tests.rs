use crossterm::event::KeyCode;
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{render, App};

struct TestHarness {
    terminal: Terminal<TestBackend>,
    app: App,
}

impl TestHarness {
    fn new() -> Self {
        let backend = TestBackend::new(80, 24);
        let terminal = Terminal::new(backend).unwrap();
        let app = App::new();
        Self { terminal, app }
    }

    fn press(&mut self, key: KeyCode) {
        let event = crossterm::event::KeyEvent::new(key, crossterm::event::KeyModifiers::NONE);
        self.app.handle_input(event);
    }

    fn snapshot(&mut self) -> &TestBackend {
        self.terminal.draw(|f| render(f, &self.app)).unwrap();
        self.terminal.backend()
    }
}

#[test]
fn test_menu_navigation_and_view_transitions() {
    let mut h = TestHarness::new();

    // Initial render
    insta::assert_snapshot!(h.snapshot(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │> View git log                        │                    "
    "                    │  View files                          │                    "
    "                    │  Settings                            │                    "
    "                    │  Help                                │                    "
    "                    │  Quit                                │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    ╰──────────────────────────────────────╯                    "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "###);

    // Press 'j' - navigate down
    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │  View git log                        │                    "
    "                    │> View files                          │                    "
    "                    │  Settings                            │                    "
    "                    │  Help                                │                    "
    "                    │  Quit                                │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    ╰──────────────────────────────────────╯                    "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "###);

    // Press Enter - enter files view
    h.press(KeyCode::Enter);
    insta::assert_snapshot!(h.snapshot(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "             ╭ Files ────────────────────────────────────────────╮              "
    "             │> internal/app/model.go                            │              "
    "             │  internal/core/messages.go                        │              "
    "             │  internal/git/commands.go                         │              "
    "             │  internal/ui/states/log/state.go                  │              "
    "             │  internal/ui/states/log/view.go                   │              "
    "             │  internal/ui/components/commit_list.go            │              "
    "             │  go.mod                                           │              "
    "             │  go.sum                                           │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             ╰───────────────────────────────────────────────────╯              "
    "                                                                                "
    "                                                                                "
    "                    ↑/↓ or j/k: navigate | Esc: back | q: qu                    "
    "                                                                                "
    "###);

    // Press 'j' - navigate within files
    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "             ╭ Files ────────────────────────────────────────────╮              "
    "             │  internal/app/model.go                            │              "
    "             │> internal/core/messages.go                        │              "
    "             │  internal/git/commands.go                         │              "
    "             │  internal/ui/states/log/state.go                  │              "
    "             │  internal/ui/states/log/view.go                   │              "
    "             │  internal/ui/components/commit_list.go            │              "
    "             │  go.mod                                           │              "
    "             │  go.sum                                           │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             │                                                   │              "
    "             ╰───────────────────────────────────────────────────╯              "
    "                                                                                "
    "                                                                                "
    "                    ↑/↓ or j/k: navigate | Esc: back | q: qu                    "
    "                                                                                "
    "###);

    // Press Esc - back to menu
    h.press(KeyCode::Esc);
    insta::assert_snapshot!(h.snapshot(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │  View git log                        │                    "
    "                    │> View files                          │                    "
    "                    │  Settings                            │                    "
    "                    │  Help                                │                    "
    "                    │  Quit                                │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    │                                      │                    "
    "                    ╰──────────────────────────────────────╯                    "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "###);
}
