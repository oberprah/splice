use crossterm::event;
use crossterm::execute;
use crossterm::terminal::{self, EnterAlternateScreen, LeaveAlternateScreen};
use ratatui::{backend::CrosstermBackend, Terminal};
use splice_rust::{action_from_event, cli, git, render, Action, App};
use std::env;
use std::io;
use std::path::PathBuf;
use std::time::{Duration, Instant};

struct TerminalGuard;

impl Drop for TerminalGuard {
    fn drop(&mut self) {
        let _ = terminal::disable_raw_mode();
        let _ = execute!(io::stdout(), LeaveAlternateScreen);
    }
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let args: Vec<String> = env::args().collect();
    let command = cli::parse_args(&args);

    match command {
        cli::Command::Help => {
            println!("{}", cli::help_text());
            return Ok(());
        }
        cli::Command::Version => {
            println!("{}", cli::version_text());
            return Ok(());
        }
        _ => {}
    }

    let repo_path = resolve_repo_path(&command)?;

    match command {
        cli::Command::Log(log_args) => run_log_app(repo_path, log_args.spec),
        cli::Command::Diff(spec) => {
            let diff_spec = git::DiffSpec {
                raw: spec.raw,
                uncommitted_type: spec.uncommitted_type,
            };

            let diff_ref = match git::resolve_diff_ref(&repo_path, diff_spec) {
                Ok(s) => s,
                Err(e) => {
                    eprintln!("Error: {}", e);
                    std::process::exit(1);
                }
            };

            let files = match git::fetch_file_changes_for_ref(&repo_path, &diff_ref) {
                Ok(f) => f,
                Err(e) => {
                    eprintln!("Error: {}", e);
                    std::process::exit(1);
                }
            };

            if files.is_empty() {
                eprintln!("No changes found");
                std::process::exit(1);
            }

            run_diff_app(repo_path, diff_ref, files)
        }
        cli::Command::Help | cli::Command::Version => unreachable!(),
    }
}

fn resolve_repo_path(command: &cli::Command) -> Result<PathBuf, Box<dyn std::error::Error>> {
    let path_arg = match command {
        cli::Command::Log(log_args) => log_args.path.clone(),
        cli::Command::Diff(_) => None,
        cli::Command::Help | cli::Command::Version => None,
    };

    match path_arg {
        Some(path) => {
            let p = PathBuf::from(&path);
            if !p.exists() {
                eprintln!("Error: Path does not exist: {}", path);
                std::process::exit(1);
            }
            if !p.is_dir() {
                eprintln!("Error: Path is not a directory: {}", path);
                std::process::exit(1);
            }
            Ok(p)
        }
        None => Ok(env::current_dir()?),
    }
}

fn run_log_app(
    repo_path: PathBuf,
    log_spec: splice_rust::core::LogSpec,
) -> Result<(), Box<dyn std::error::Error>> {
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

    let res = run_app(
        &mut terminal,
        App::with_repo_path_and_log_spec(repo_path, log_spec),
    );

    if let Err(err) = res {
        eprintln!("Error: {err}");
    }

    Ok(())
}

fn run_diff_app(
    repo_path: PathBuf,
    diff_ref: splice_rust::core::DiffRef,
    files: Vec<splice_rust::core::FileDiffInfo>,
) -> Result<(), Box<dyn std::error::Error>> {
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

    let app = App::with_diff_source(repo_path, diff_ref, files);
    let res = run_app(&mut terminal, app);

    if let Err(err) = res {
        eprintln!("Error: {err}");
    }

    Ok(())
}

fn run_app(terminal: &mut Terminal<CrosstermBackend<io::Stdout>>, mut app: App) -> io::Result<()> {
    let tick_rate = Duration::from_millis(250);
    let animation_rate = Duration::from_millis(16); // ~60 fps during animation
    let mut last_tick = Instant::now();
    let mut should_render = true;

    loop {
        if should_render {
            terminal.draw(|f| {
                render(f, &mut app);
            })?;
            should_render = false;
        }

        let timeout = if app.is_animating() {
            animation_rate
        } else {
            tick_rate.saturating_sub(last_tick.elapsed())
        };

        if event::poll(timeout)? {
            let action = action_from_event(event::read()?);
            if action != Action::None {
                should_render = true;
                if app.error.take().is_some() {
                    continue;
                }

                if action == Action::OpenInEditor {
                    if let Some(err) = open_diff_in_editor(terminal, &mut app)? {
                        app.error = Some(err);
                    }
                    continue;
                }

                if app.update(action) {
                    return Ok(());
                }
            }
        }

        // Advance scroll animation each tick
        if app.advance_scroll_animation() {
            should_render = true;
        }

        if last_tick.elapsed() >= tick_rate {
            last_tick = Instant::now();
        }
    }
}

fn open_diff_in_editor(
    terminal: &mut Terminal<CrosstermBackend<io::Stdout>>,
    app: &mut App,
) -> io::Result<Option<String>> {
    terminal::disable_raw_mode()?;
    execute!(io::stdout(), LeaveAlternateScreen)?;

    let editor_result = app.open_current_diff_in_editor();

    execute!(io::stdout(), EnterAlternateScreen)?;
    terminal::enable_raw_mode()?;
    terminal.clear()?;

    Ok(editor_result.err())
}
