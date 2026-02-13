use crossterm::event::{KeyCode, KeyModifiers};
use ratatui::{backend::TestBackend, Terminal};
use splice_rust::{render, App};

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
fn test_menu_initial_render() {
    let (mut terminal, app) = setup_terminal();
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!(terminal.backend());
}

#[test]
fn test_menu_navigation_down() {
    let (mut terminal, mut app) = setup_terminal();

    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("menu_nav_down_initial", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("menu_nav_down_after_1", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("menu_nav_down_after_2", terminal.backend());
}

#[test]
fn test_menu_navigation_up() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("menu_nav_up_at_item_2", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('k')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("menu_nav_up_after_k", terminal.backend());
}

#[test]
fn test_menu_enter_git_log() {
    let (mut terminal, mut app) = setup_terminal();

    terminal.draw(|f| render(f, &app)).unwrap();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();

    insta::assert_snapshot!(terminal.backend());
}

#[test]
fn test_menu_enter_files() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();

    insta::assert_snapshot!(terminal.backend());
}

#[test]
fn test_git_log_navigation() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("git_log_nav_initial", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("git_log_nav_after_1", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("git_log_nav_after_2", terminal.backend());
}

#[test]
fn test_git_log_escape_back_to_menu() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();

    app.handle_input(key_event(KeyCode::Esc));
    terminal.draw(|f| render(f, &app)).unwrap();

    insta::assert_snapshot!(terminal.backend());
}

#[test]
fn test_files_navigation() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("files_nav_initial", terminal.backend());

    app.handle_input(key_event(KeyCode::Char('j')));
    terminal.draw(|f| render(f, &app)).unwrap();
    insta::assert_snapshot!("files_nav_after_1", terminal.backend());
}

#[test]
fn test_files_escape_back_to_menu() {
    let (mut terminal, mut app) = setup_terminal();

    app.handle_input(key_event(KeyCode::Char('j')));
    app.handle_input(key_event(KeyCode::Enter));
    terminal.draw(|f| render(f, &app)).unwrap();

    app.handle_input(key_event(KeyCode::Esc));
    terminal.draw(|f| render(f, &app)).unwrap();

    insta::assert_snapshot!(terminal.backend());
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

#[test]
fn test_quit_from_menu_with_enter() {
    let (_, mut app) = setup_terminal();
    for _ in 0..4 {
        app.handle_input(key_event(KeyCode::Char('j')));
    }
    let should_quit = app.handle_input(key_event(KeyCode::Enter));
    assert!(should_quit);
}

#[test]
fn test_ctrl_c_quits_from_menu() {
    let (_, mut app) = setup_terminal();
    let should_quit =
        app.handle_input(key_event_with_modifiers(KeyCode::Char('c'), KeyModifiers::CONTROL));
    assert!(should_quit);
}
