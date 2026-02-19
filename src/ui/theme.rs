use ratatui::style::{Color, Modifier, Style};

pub const HASH: Style = Style::new().fg(Color::Yellow);
pub const MESSAGE: Style = Style::new().fg(Color::White);
pub const REFS: Style = Style::new().fg(Color::Blue);

pub const SELECTED_HASH: Style = Style::new().fg(Color::Yellow).add_modifier(Modifier::BOLD);
pub const SELECTED_MESSAGE: Style = Style::new().fg(Color::White).add_modifier(Modifier::BOLD);
