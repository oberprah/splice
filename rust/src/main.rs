use crossterm::event::{self, Event, KeyCode, KeyModifiers};
use crossterm::execute;
use crossterm::terminal::{self, EnterAlternateScreen, LeaveAlternateScreen};
use ratatui::{prelude::*, widgets::*};
use std::io;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    terminal::enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen)?;
    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let res = run_app(&mut terminal);

    terminal::disable_raw_mode()?;
    execute!(terminal.backend_mut(), LeaveAlternateScreen)?;
    terminal.show_cursor()?;

    if let Err(err) = res {
        eprintln!("Error: {err}");
    }

    Ok(())
}

fn run_app<B: Backend>(terminal: &mut Terminal<B>) -> io::Result<()> {
    loop {
        terminal.draw(|f| {
            let size = f.area();
            let text = "Hello, Ratatui!";
            let paragraph = Paragraph::new(text)
                .style(Style::default().fg(Color::Cyan))
                .alignment(Alignment::Center);
            let block = Block::default()
                .title("Splice Rust")
                .borders(Borders::ALL);
            let paragraph = paragraph.block(block);
            f.render_widget(
                paragraph,
                Rect::new(
                    (size.width.saturating_sub(30)) / 2,
                    (size.height.saturating_sub(5)) / 2,
                    30.min(size.width),
                    5.min(size.height),
                ),
            );
        })?;

        if let Event::Key(key) = event::read()? {
            if key.code == KeyCode::Char('q') || key.code == KeyCode::Esc
                || (key.modifiers.contains(KeyModifiers::CONTROL) && key.code == KeyCode::Char('c'))
            {
                return Ok(());
            }
        }
    }
}
