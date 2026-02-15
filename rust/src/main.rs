use crossterm::event::{self, Event};
use crossterm::execute;
use crossterm::terminal::{self, EnterAlternateScreen, LeaveAlternateScreen};
use ratatui::{backend::CrosstermBackend, prelude::Backend, Terminal};
use splice_rust::{render, App};
use std::env;
use std::io;
use std::path::Path;

struct TerminalGuard;

impl Drop for TerminalGuard {
    fn drop(&mut self) {
        let _ = terminal::disable_raw_mode();
        let _ = execute!(io::stdout(), LeaveAlternateScreen);
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let repo_path = match env::args().nth(1) {
        Some(path) => {
            let p = Path::new(&path);
            if !p.exists() {
                eprintln!("Error: Path does not exist: {}", path);
                std::process::exit(1);
            }
            if !p.is_dir() {
                eprintln!("Error: Path is not a directory: {}", path);
                std::process::exit(1);
            }
            p.to_path_buf()
        }
        None => env::current_dir()?,
    };

    terminal::enable_raw_mode()?;
    let mut stdout = io::stdout();
    execute!(stdout, EnterAlternateScreen)?;
    let _guard = TerminalGuard;

    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let res = run_app(&mut terminal, repo_path);

    if let Err(err) = res {
        eprintln!("Error: {err}");
    }

    Ok(())
}

fn run_app<B: Backend>(terminal: &mut Terminal<B>, repo_path: std::path::PathBuf) -> io::Result<()> {
    let mut app = App::with_repo_path(repo_path);

    loop {
        terminal.draw(|f| {
            render(f, &app);
        })?;

        if let Event::Key(key) = event::read()? {
            if app.handle_input(key) {
                return Ok(());
            }
        }
    }
}
