use crossterm::event::{Event, KeyCode, KeyEvent, KeyModifiers};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Action {
    Quit,
    MoveDown,
    MoveUp,
    PageDown,
    PageUp,
    Resize { width: u16, height: u16 },
    None,
}

pub fn action_from_event(event: Event) -> Action {
    match event {
        Event::Key(key) => action_from_key(key),
        Event::Resize(width, height) => Action::Resize { width, height },
        _ => Action::None,
    }
}

fn action_from_key(key: KeyEvent) -> Action {
    match key.code {
        KeyCode::Char('q') => Action::Quit,
        KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::Quit,
        KeyCode::Down | KeyCode::Char('j') => Action::MoveDown,
        KeyCode::Up | KeyCode::Char('k') => Action::MoveUp,
        KeyCode::Char('d') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::PageDown,
        KeyCode::Char('u') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::PageUp,
        _ => Action::None,
    }
}
