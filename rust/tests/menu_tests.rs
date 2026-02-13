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
fn test_menu_initial_render() {
    let (mut terminal, app) = setup_terminal();
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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
}

#[test]
fn test_menu_navigation_down() {
    let (mut terminal, mut app) = setup_terminal();

    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │  View git log                        │                    "
    "                    │  View files                          │                    "
    "                    │> Settings                            │                    "
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

#[test]
fn test_menu_navigation_up() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                    ╭ Splice Rust ─────────────────────────╮                    "
    "                    │  View git log                        │                    "
    "                    │  View files                          │                    "
    "                    │> Settings                            │                    "
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

    app.handle_input(key_event(KeyCode::Char('k')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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
