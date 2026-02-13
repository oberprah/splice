use crossterm::event::{self, KeyCode, KeyModifiers};
use ratatui::{prelude::*, widgets::*};

pub enum View {
    Menu,
    GitLog,
    Files,
}

pub struct App {
    pub view: View,
    pub menu_selected: usize,
    pub git_log_selected: usize,
    pub files_selected: usize,
}

impl App {
    pub fn new() -> Self {
        Self {
            view: View::Menu,
            menu_selected: 0,
            git_log_selected: 0,
            files_selected: 0,
        }
    }

    pub fn menu_items() -> Vec<&'static str> {
        vec!["View git log", "View files", "Settings", "Help", "Quit"]
    }

    pub fn git_log_items() -> Vec<(&'static str, &'static str, &'static str)> {
        vec![
            ("e96269a", "Add Rust experiment with ratatui hello world", "2 hours ago"),
            ("b4a19bc", "Add Cargo.lock for Rust experiment", "2 hours ago"),
            ("9b17fe2", "Fix sandbox conflicts for multiple repo copies", "1 day ago"),
            ("55bc557", "Refactor sandbox to user-configurable Docker", "2 days ago"),
            ("bfa9279", "Fix multi-commit file preview", "3 days ago"),
            ("1da9976", "Restructure git package into layered architecture", "4 days ago"),
            ("951443e", "Bump the go-dependencies group with 2 updates", "5 days ago"),
            ("a1b2c3d", "Add initial TUI implementation", "1 week ago"),
        ]
    }

    pub fn files_items() -> Vec<&'static str> {
        vec![
            "internal/app/model.go",
            "internal/core/messages.go",
            "internal/git/commands.go",
            "internal/ui/states/log/state.go",
            "internal/ui/states/log/view.go",
            "internal/ui/components/commit_list.go",
            "go.mod",
            "go.sum",
        ]
    }

    pub fn handle_input(&mut self, key: event::KeyEvent) -> bool {
        match self.view {
            View::Menu => handle_menu_input(self, key),
            View::GitLog => handle_git_log_input(self, key),
            View::Files => handle_files_input(self, key),
        }
    }
}

pub fn render(f: &mut Frame, app: &App) {
    match app.view {
        View::Menu => render_menu(f, app),
        View::GitLog => render_git_log(f, app),
        View::Files => render_files(f, app),
    }
}

fn render_menu(f: &mut Frame, app: &App) {
    let size = f.area();

    let items: Vec<ListItem> = App::menu_items()
        .iter()
        .enumerate()
        .map(|(i, &item)| {
            let style = if i == app.menu_selected {
                Style::default()
                    .fg(Color::Black)
                    .bg(Color::Cyan)
                    .add_modifier(Modifier::BOLD)
            } else {
                Style::default()
            };
            ListItem::new(item).style(style)
        })
        .collect();

    let list = List::new(items)
        .block(
            Block::default()
                .title(" Splice Rust ")
                .title_style(Style::default().fg(Color::Cyan).add_modifier(Modifier::BOLD))
                .borders(Borders::ALL)
                .border_type(BorderType::Rounded),
        )
        .highlight_symbol("> ");

    let area = centered_rect(40, 12, size);
    f.render_widget(list, area);
}

fn render_git_log(f: &mut Frame, app: &App) {
    let size = f.area();

    let items: Vec<ListItem> = App::git_log_items()
        .iter()
        .enumerate()
        .map(|(i, &(hash, msg, time))| {
            let style = if i == app.git_log_selected {
                Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD)
            } else {
                Style::default()
            };
            let hash_span = Span::styled(hash, Style::default().fg(Color::Green));
            let msg_span = Span::styled(format!(" {}", msg), style);
            let time_span = Span::styled(
                format!(" ({})", time),
                Style::default().fg(Color::DarkGray),
            );
            ListItem::new(Line::from(vec![hash_span, msg_span, time_span]))
        })
        .collect();

    let list = List::new(items)
        .block(
            Block::default()
                .title(" Git Log ")
                .title_style(Style::default().fg(Color::Green).add_modifier(Modifier::BOLD))
                .borders(Borders::ALL)
                .border_type(BorderType::Rounded),
        )
        .highlight_symbol("> ");

    let area = Rect::new(
        size.width / 8,
        size.height / 6,
        size.width * 6 / 8,
        size.height * 4 / 6,
    );
    f.render_widget(list, area);

    let help = Paragraph::new("↑/↓ or j/k: navigate | Esc: back | q: quit")
        .style(Style::default().fg(Color::DarkGray))
        .alignment(Alignment::Center);
    let help_area = Rect::new(size.width / 4, size.height - 2, size.width / 2, 1);
    f.render_widget(help, help_area);
}

fn render_files(f: &mut Frame, app: &App) {
    let size = f.area();

    let items: Vec<ListItem> = App::files_items()
        .iter()
        .enumerate()
        .map(|(i, &item)| {
            let style = if i == app.files_selected {
                Style::default()
                    .fg(Color::Black)
                    .bg(Color::Yellow)
                    .add_modifier(Modifier::BOLD)
            } else {
                Style::default().fg(Color::White)
            };
            ListItem::new(item).style(style)
        })
        .collect();

    let list = List::new(items)
        .block(
            Block::default()
                .title(" Files ")
                .title_style(Style::default().fg(Color::Yellow).add_modifier(Modifier::BOLD))
                .borders(Borders::ALL)
                .border_type(BorderType::Rounded),
        )
        .highlight_symbol("> ");

    let area = Rect::new(
        size.width / 6,
        size.height / 6,
        size.width * 2 / 3,
        size.height * 4 / 6,
    );
    f.render_widget(list, area);

    let help = Paragraph::new("↑/↓ or j/k: navigate | Esc: back | q: quit")
        .style(Style::default().fg(Color::DarkGray))
        .alignment(Alignment::Center);
    let help_area = Rect::new(size.width / 4, size.height - 2, size.width / 2, 1);
    f.render_widget(help, help_area);
}

fn handle_menu_input(app: &mut App, key: event::KeyEvent) -> bool {
    let items = App::menu_items();
    match key.code {
        KeyCode::Char('q') | KeyCode::Esc => return true,
        KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => return true,
        KeyCode::Down | KeyCode::Char('j') => {
            if app.menu_selected < items.len() - 1 {
                app.menu_selected += 1;
            }
        }
        KeyCode::Up | KeyCode::Char('k') => {
            if app.menu_selected > 0 {
                app.menu_selected -= 1;
            }
        }
        KeyCode::Enter => match items[app.menu_selected] {
            "Quit" => return true,
            "View git log" => app.view = View::GitLog,
            "View files" => app.view = View::Files,
            _ => {}
        },
        _ => {}
    }
    false
}

fn handle_git_log_input(app: &mut App, key: event::KeyEvent) -> bool {
    let items = App::git_log_items();
    match key.code {
        KeyCode::Char('q') => return true,
        KeyCode::Esc => app.view = View::Menu,
        KeyCode::Down | KeyCode::Char('j') => {
            if app.git_log_selected < items.len() - 1 {
                app.git_log_selected += 1;
            }
        }
        KeyCode::Up | KeyCode::Char('k') => {
            if app.git_log_selected > 0 {
                app.git_log_selected -= 1;
            }
        }
        _ => {}
    }
    false
}

fn handle_files_input(app: &mut App, key: event::KeyEvent) -> bool {
    let items = App::files_items();
    match key.code {
        KeyCode::Char('q') => return true,
        KeyCode::Esc => app.view = View::Menu,
        KeyCode::Down | KeyCode::Char('j') => {
            if app.files_selected < items.len() - 1 {
                app.files_selected += 1;
            }
        }
        KeyCode::Up | KeyCode::Char('k') => {
            if app.files_selected > 0 {
                app.files_selected -= 1;
            }
        }
        _ => {}
    }
    false
}

fn centered_rect(percent_x: u16, percent_y: u16, r: Rect) -> Rect {
    Rect::new(
        (r.width.saturating_sub(percent_x)) / 2,
        (r.height.saturating_sub(percent_y)) / 2,
        percent_x.min(r.width),
        percent_y.min(r.height),
    )
}
