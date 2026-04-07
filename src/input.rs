use crossterm::event::{Event, KeyCode, KeyEvent, KeyModifiers, MouseEvent, MouseEventKind};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Action {
    Quit,
    MoveDown,
    MoveUp,
    ScrollDown,
    ScrollUp,
    PageDown,
    PageUp,
    Open,
    Back,
    ExpandFolder,
    CollapseFolder,
    ToggleFolder,
    ToggleViewMode,
    NextDiff,
    PrevDiff,
    OpenInEditor,
    ToggleMessage,
    Resize { width: u16, height: u16 },
    None,
}

pub fn action_from_event(event: Event) -> Action {
    match event {
        Event::Key(key) => action_from_key(key),
        Event::Resize(width, height) => Action::Resize { width, height },
        Event::Mouse(mouse) => action_from_mouse(mouse),
        _ => Action::None,
    }
}

fn action_from_mouse(mouse: MouseEvent) -> Action {
    match mouse.kind {
        MouseEventKind::ScrollDown => Action::ScrollDown,
        MouseEventKind::ScrollUp => Action::ScrollUp,
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
        KeyCode::Char('v') => Action::ToggleViewMode,
        KeyCode::Char('n') => Action::NextDiff,
        KeyCode::Char('p') => Action::PrevDiff,
        KeyCode::Char('o') => Action::OpenInEditor,
        KeyCode::Char('m') => Action::ToggleMessage,
        _ => Action::None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crossterm::event::{KeyEvent, KeyModifiers as KM, MouseEvent, MouseEventKind};

    fn mouse_event(kind: MouseEventKind) -> Event {
        Event::Mouse(MouseEvent {
            kind,
            column: 0,
            row: 0,
            modifiers: KM::empty(),
        })
    }

    #[test]
    fn test_v_key_maps_to_toggle_view_mode() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('v')));
        assert_eq!(action_from_event(event), Action::ToggleViewMode);
    }

    #[test]
    fn test_n_key_maps_to_next_diff() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('n')));
        assert_eq!(action_from_event(event), Action::NextDiff);
    }

    #[test]
    fn test_p_key_maps_to_prev_diff() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('p')));
        assert_eq!(action_from_event(event), Action::PrevDiff);
    }

    #[test]
    fn test_o_key_maps_to_open_in_editor() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('o')));
        assert_eq!(action_from_event(event), Action::OpenInEditor);
    }

    #[test]
    fn test_m_key_maps_to_toggle_message() {
        let event = Event::Key(KeyEvent::from(KeyCode::Char('m')));
        assert_eq!(action_from_event(event), Action::ToggleMessage);
    }

    #[test]
    fn test_scroll_down_maps_to_scroll_down() {
        assert_eq!(
            action_from_event(mouse_event(MouseEventKind::ScrollDown)),
            Action::ScrollDown
        );
    }

    #[test]
    fn test_scroll_up_maps_to_scroll_up() {
        assert_eq!(
            action_from_event(mouse_event(MouseEventKind::ScrollUp)),
            Action::ScrollUp
        );
    }

    #[test]
    fn test_mouse_click_maps_to_none() {
        assert_eq!(
            action_from_event(mouse_event(MouseEventKind::Down(
                crossterm::event::MouseButton::Left
            ))),
            Action::None
        );
    }
}
