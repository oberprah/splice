use crossterm::event::KeyCode;
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{render, App};

fn setup_terminal() -> (Terminal<TestBackend>, App) {
    let backend = TestBackend::new(80, 24);
    let terminal = Terminal::new(backend).unwrap();
    let app = App::new();
    (terminal, app)
}

fn key_event(code: KeyCode) -> crossterm::event::KeyEvent {
    crossterm::event::KeyEvent::new(code, crossterm::event::KeyModifiers::NONE)
}

#[test]
fn test_git_log_initial_render() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "          ╭ Git Log ─────────────────────────────────────────────────╮          "
    "          │e96269a Add Rust experiment with ratatui hello world (2 ho│          "
    "          │b4a19bc Add Cargo.lock for Rust experiment (2 hours ago)  │          "
    "          │9b17fe2 Fix sandbox conflicts for multiple repo copies (1 │          "
    "          │55bc557 Refactor sandbox to user-configurable Docker (2 da│          "
    "          │bfa9279 Fix multi-commit file preview (3 days ago)        │          "
    "          │1da9976 Restructure git package into layered architecture │          "
    "          │951443e Bump the go-dependencies group with 2 updates (5 d│          "
    "          │a1b2c3d Add initial TUI implementation (1 week ago)       │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          ╰──────────────────────────────────────────────────────────╯          "
    "                                                                                "
    "                                                                                "
    "                    ↑/↓ or j/k: navigate | Esc: back | q: qu                    "
    "                                                                                "
    "###);
}

#[test]
fn test_git_log_navigation() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "          ╭ Git Log ─────────────────────────────────────────────────╮          "
    "          │e96269a Add Rust experiment with ratatui hello world (2 ho│          "
    "          │b4a19bc Add Cargo.lock for Rust experiment (2 hours ago)  │          "
    "          │9b17fe2 Fix sandbox conflicts for multiple repo copies (1 │          "
    "          │55bc557 Refactor sandbox to user-configurable Docker (2 da│          "
    "          │bfa9279 Fix multi-commit file preview (3 days ago)        │          "
    "          │1da9976 Restructure git package into layered architecture │          "
    "          │951443e Bump the go-dependencies group with 2 updates (5 d│          "
    "          │a1b2c3d Add initial TUI implementation (1 week ago)       │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          ╰──────────────────────────────────────────────────────────╯          "
    "                                                                                "
    "                                                                                "
    "                    ↑/↓ or j/k: navigate | Esc: back | q: qu                    "
    "                                                                                "
    "###);

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "          ╭ Git Log ─────────────────────────────────────────────────╮          "
    "          │e96269a Add Rust experiment with ratatui hello world (2 ho│          "
    "          │b4a19bc Add Cargo.lock for Rust experiment (2 hours ago)  │          "
    "          │9b17fe2 Fix sandbox conflicts for multiple repo copies (1 │          "
    "          │55bc557 Refactor sandbox to user-configurable Docker (2 da│          "
    "          │bfa9279 Fix multi-commit file preview (3 days ago)        │          "
    "          │1da9976 Restructure git package into layered architecture │          "
    "          │951443e Bump the go-dependencies group with 2 updates (5 d│          "
    "          │a1b2c3d Add initial TUI implementation (1 week ago)       │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          │                                                          │          "
    "          ╰──────────────────────────────────────────────────────────╯          "
    "                                                                                "
    "                                                                                "
    "                    ↑/↓ or j/k: navigate | Esc: back | q: qu                    "
    "                                                                                "
    "###);
}

#[test]
fn test_git_log_escape_back_to_menu() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();

    app.handle_input(key_event(KeyCode::Esc));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │View git log                          │                    "
    "                    │View files                            │                    "
    "                    │Settings                              │                    "
    "                    │Help                                  │                    "
    "                    │Quit                                  │                    "
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
