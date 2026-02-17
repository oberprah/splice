use crossterm::event;
use crossterm::execute;
use crossterm::terminal::{self, EnterAlternateScreen, LeaveAlternateScreen};
use ratatui::{backend::CrosstermBackend, prelude::Backend, Terminal};
use splice_rust::{action_from_event, render, Action, App};
use std::env;
use std::io;
use std::path::Path;
use std::time::{Duration, Instant};

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

    let default_panic = std::panic::take_hook();
    std::panic::set_hook(Box::new(move |info| {
        let _ = terminal::disable_raw_mode();
        let _ = execute!(io::stdout(), LeaveAlternateScreen);
        default_panic(info);
    }));

    let backend = CrosstermBackend::new(stdout);
    let mut terminal = Terminal::new(backend)?;

    let res = run_app(&mut terminal, repo_path);

    if let Err(err) = res {
        eprintln!("Error: {err}");
    }

    Ok(())
}

fn run_app<B: Backend>(
    terminal: &mut Terminal<B>,
    repo_path: std::path::PathBuf,
) -> io::Result<()> {
    let mut app = App::with_repo_path(repo_path);
    let tick_rate = Duration::from_millis(250);
    let mut last_tick = Instant::now();
    let mut should_render = true;

    loop {
        if should_render {
            terminal.draw(|f| {
                render(f, &mut app);
            })?;
            should_render = false;
        }

        let timeout = tick_rate.saturating_sub(last_tick.elapsed());
        if event::poll(timeout)? {
            let action = action_from_event(event::read()?);
            if action != Action::None {
                should_render = true;
                if app.update(action) {
                    return Ok(());
                }
            }
        }

        if last_tick.elapsed() >= tick_rate {
            last_tick = Instant::now();
        }
    }
}
