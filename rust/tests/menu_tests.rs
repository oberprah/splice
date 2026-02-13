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
fn test_menu_navigation_and_view_transitions() {
    let (mut terminal, mut app) = setup_terminal();

    // Initial render - first item selected
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

    // Navigate down - second item selected
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

    // Enter files view
    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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

    // Navigate within files
    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend(), @r###"
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

    // Escape back to menu - selection preserved at "View files"
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
