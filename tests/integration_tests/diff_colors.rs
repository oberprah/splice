use crate::common::{reset_counter, TestRepo};
use ratatui::{backend::TestBackend, style::Color, Terminal};
use serial_test::serial;
use splice::{render, Action, App, ThemeMode};

fn navigate_to_diff_view(app: &mut App) {
    app.update(Action::Open);
    app.update(Action::Open);
}

fn collect_background_colors(
    buffer: &ratatui::buffer::Buffer,
) -> std::collections::HashSet<Option<Color>> {
    let mut colors = std::collections::HashSet::new();
    for y in 0..buffer.area.height {
        for x in 0..buffer.area.width {
            let cell = buffer.cell((x, y)).unwrap();
            colors.insert(cell.style().bg);
        }
    }
    colors
}

#[test]
#[serial]
fn diff_view_shows_removed_lines_with_red_background() {
    reset_counter();

    let repo = TestRepo::new();

    repo.add_file("file.txt", "line1\nremoved line\nline3\n");
    repo.commit("Initial");
    repo.modify_file("file.txt", "line1\nline3\n");
    repo.commit("Remove line");

    let backend = TestBackend::new(80, 24);
    let mut terminal = Terminal::new(backend).unwrap();
    let mut app = App::with_repo_path(repo.path());
    app.set_viewport_size(23, 80);
    app.set_theme_mode(ThemeMode::Dark);

    navigate_to_diff_view(&mut app);

    terminal.draw(|f| render(f, &mut app)).unwrap();

    let buffer = terminal.backend().buffer();
    let diff_removed_bg = Color::Rgb(0x4d, 0x26, 0x26);

    let colors = collect_background_colors(buffer);

    assert!(
        colors.contains(&Some(diff_removed_bg)),
        "Expected removed background color {diff_removed_bg:?} to be present. Found colors: {colors:?}"
    );
}

#[test]
#[serial]
fn diff_view_shows_added_lines_with_green_background() {
    reset_counter();

    let repo = TestRepo::new();

    repo.add_file("file.txt", "line1\nline2\n");
    repo.commit("Initial");
    repo.modify_file("file.txt", "line1\nline2\nnew line\n");
    repo.commit("Add line");

    let backend = TestBackend::new(80, 24);
    let mut terminal = Terminal::new(backend).unwrap();
    let mut app = App::with_repo_path(repo.path());
    app.set_viewport_size(23, 80);
    app.set_theme_mode(ThemeMode::Dark);

    navigate_to_diff_view(&mut app);

    terminal.draw(|f| render(f, &mut app)).unwrap();

    let buffer = terminal.backend().buffer();
    let diff_added_bg = Color::Rgb(0x26, 0x4d, 0x26);

    let colors = collect_background_colors(buffer);

    assert!(
        colors.contains(&Some(diff_added_bg)),
        "Expected added background color {diff_added_bg:?} to be present. Found colors: {colors:?}"
    );
}

#[test]
#[serial]
fn diff_view_shows_changed_lines_with_blue_background() {
    reset_counter();

    let repo = TestRepo::new();

    repo.add_file("file.txt", "line1\noriginal text\nline3\n");
    repo.commit("Initial");
    repo.modify_file("file.txt", "line1\nmodified text\nline3\n");
    repo.commit("Modify line");

    let backend = TestBackend::new(80, 24);
    let mut terminal = Terminal::new(backend).unwrap();
    let mut app = App::with_repo_path(repo.path());
    app.set_viewport_size(23, 80);
    app.set_theme_mode(ThemeMode::Dark);

    navigate_to_diff_view(&mut app);

    terminal.draw(|f| render(f, &mut app)).unwrap();

    let buffer = terminal.backend().buffer();
    let diff_changed_bg = Color::Rgb(0x26, 0x36, 0x4d);

    let colors = collect_background_colors(buffer);

    assert!(
        colors.contains(&Some(diff_changed_bg)),
        "Expected changed background color {diff_changed_bg:?} to be present. Found colors: {colors:?}"
    );
}

#[test]
#[serial]
fn diff_view_has_removed_and_added_colors_for_pure_changes() {
    reset_counter();

    let repo = TestRepo::new();

    repo.add_file("file.txt", "line1\nline2\nline3\n");
    repo.commit("Initial");
    repo.modify_file("file.txt", "line1\nline3\nnew line\n");
    repo.commit("Delete and add");

    let backend = TestBackend::new(80, 24);
    let mut terminal = Terminal::new(backend).unwrap();
    let mut app = App::with_repo_path(repo.path());
    app.set_viewport_size(23, 80);
    app.set_theme_mode(ThemeMode::Dark);

    navigate_to_diff_view(&mut app);

    terminal.draw(|f| render(f, &mut app)).unwrap();

    let buffer = terminal.backend().buffer();
    let colors = collect_background_colors(buffer);

    let diff_added_bg = Color::Rgb(0x1e, 0x3a, 0x1e);
    let diff_removed_bg = Color::Rgb(0x4d, 0x26, 0x26);

    assert!(
        colors.contains(&Some(diff_removed_bg)),
        "Expected removed background color. Found: {colors:?}"
    );
    assert!(
        colors.contains(&Some(diff_added_bg)),
        "Expected added background color. Found: {colors:?}"
    );
}

#[test]
#[serial]
fn diff_view_uses_changed_color_for_line_modification() {
    reset_counter();

    let repo = TestRepo::new();

    repo.add_file("file.txt", "line1\noriginal\nline3\n");
    repo.commit("Initial");
    repo.modify_file("file.txt", "line1\nmodified\nline3\n");
    repo.commit("Modify");

    let backend = TestBackend::new(80, 24);
    let mut terminal = Terminal::new(backend).unwrap();
    let mut app = App::with_repo_path(repo.path());
    app.set_viewport_size(23, 80);
    app.set_theme_mode(ThemeMode::Dark);

    navigate_to_diff_view(&mut app);

    terminal.draw(|f| render(f, &mut app)).unwrap();

    let buffer = terminal.backend().buffer();
    let colors = collect_background_colors(buffer);

    let diff_changed_bg = Color::Rgb(0x26, 0x36, 0x4d);

    assert!(
        colors.contains(&Some(diff_changed_bg)),
        "Expected changed background color for modified lines. Found: {colors:?}"
    );
}
