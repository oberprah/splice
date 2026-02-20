use crossterm::event::{Event, KeyCode, KeyEvent, KeyModifiers};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Action {
    Quit,
    MoveDown,
    MoveUp,
    PageDown,
    PageUp,
    Open,
    Back,
    ExpandFolder,
    CollapseFolder,
    ToggleFolder,
    ToggleVisualMode,
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
        KeyCode::Char('q') => Action::Back,
        KeyCode::Char('Q') => Action::Quit,
        KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::Quit,
        KeyCode::Down | KeyCode::Char('j') => Action::MoveDown,
        KeyCode::Up | KeyCode::Char('k') => Action::MoveUp,
        KeyCode::Char('d') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::PageDown,
        KeyCode::Char('u') if key.modifiers.contains(KeyModifiers::CONTROL) => Action::PageUp,
        KeyCode::Enter => Action::Open,
        KeyCode::Char(' ') => Action::ToggleFolder,
        KeyCode::Right => Action::ExpandFolder,
        KeyCode::Left => Action::CollapseFolder,
        KeyCode::Char('v') => Action::ToggleVisualMode,
        _ => Action::None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crossterm::event::KeyEvent;

    #[test]
    fn test_v_key_maps_to_toggle_visual_mode() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('v')));
        assert_eq!(action_from_event(event), Action::ToggleVisualMode);
    }
}
