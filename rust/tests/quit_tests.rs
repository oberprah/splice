use crossterm::event::{KeyCode, KeyModifiers};
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::App;

fn setup_terminal() -> (Terminal<TestBackend>, App) {
    let backend = TestBackend::new(80, 24);
    let terminal = Terminal::new(backend).unwrap();
    let app = App::new();
    (terminal, app)
}

fn key_event(code: KeyCode) -> crossterm::event::KeyEvent {
    crossterm::event::KeyEvent::new(code, KeyModifiers::NONE)
}

fn key_event_with_modifiers(code: KeyCode, modifiers: KeyModifiers) -> crossterm::event::KeyEvent {
    crossterm::event::KeyEvent::new(code, modifiers)
}

#[test]
fn test_quit_from_menu_with_q() {
    let (_, mut app) = setup_terminal();
    let should_quit = app.handle_input(key_event(KeyCode::Char('q')));
    assert!(should_quit);
}

#[test]
fn test_quit_from_menu_with_esc() {
    let (_, mut app) = setup_terminal();
    let should_quit = app.handle_input(key_event(KeyCode::Esc));
    assert!(should_quit);
}

#[test]
fn test_quit_from_menu_with_ctrl_c() {
    let (_, mut app) = setup_terminal();
    let should_quit = app.handle_input(key_event_with_modifiers(KeyCode::Char('c'), KeyModifiers::CONTROL));
    assert!(should_quit);
}

#[test]
fn test_quit_from_menu_with_enter_on_quit_item() {
    let (_, mut app) = setup_terminal();
    for _ in 0..4 {
        app.handle_input(key_event(KeyCode::Char('j')));
    }
    let should_quit = app.handle_input(key_event(KeyCode::Enter));
    assert!(should_quit);
}

#[test]
fn test_quit_from_git_log_with_q() {
    let (_, mut app) = setup_terminal();
    app.handle_input(key_event(KeyCode::Enter));
    let should_quit = app.handle_input(key_event(KeyCode::Char('q')));
    assert!(should_quit);
}

#[test]
fn test_quit_from_files_with_q() {
    let (_, mut app) = setup_terminal();
    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Enter));
    let should_quit = app.handle_input(key_event(KeyCode::Char('q')));
    assert!(should_quit);
}
