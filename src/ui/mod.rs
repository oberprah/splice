mod diff;
mod files;
mod log;
mod theme;

pub use theme::*;

use crate::app::{App, ThemeMode, View};
use ratatui::{prelude::*, widgets::Paragraph};
use std::sync::OnceLock;

const LOG_AREA_X: u16 = 2;
const LOG_AREA_Y: u16 = 0;
const LOG_AREA_RIGHT_MARGIN: u16 = 4;
const LOG_AREA_BOTTOM_MARGIN: u16 = 0;
static CACHED_THEME: OnceLock<Theme> = OnceLock::new();

pub fn render(f: &mut Frame, app: &mut App) {
    let size = f.area();

    if let Some(ref error) = app.error {
        let msg = Paragraph::new(format!("Error: {}", error))
            .style(Style::default().fg(Color::Red))
            .alignment(Alignment::Center);
        f.render_widget(msg, size);
        return;
    }

    let area = Rect::new(
        LOG_AREA_X,
        LOG_AREA_Y,
        size.width.saturating_sub(LOG_AREA_RIGHT_MARGIN),
        size.height.saturating_sub(LOG_AREA_BOTTOM_MARGIN),
    );

    let viewport_height = match &app.view {
        View::Log(_) => area.height.saturating_sub(2) as usize,
        _ => area.height.saturating_sub(1) as usize,
    };
    let viewport_width = area.width as usize;
    app.set_viewport_size(viewport_height, viewport_width);

    let theme = match app.theme_mode {
        ThemeMode::Auto => CACHED_THEME.get_or_init(Theme::detect_theme),
        ThemeMode::Dark => &Theme::dark(),
        ThemeMode::Light => &Theme::light(),
    };

    match &app.view {
        View::Log(log) => {
            log::render_log_view(
                f,
                &log.commits,
                &log.graph_layout,
                &log.summary,
                &log.cursor,
                log.scroll_offset,
                app.now(),
                area,
                theme,
            );
        }
        View::Files(files) => {
            files::render_files_view(f, files, area, theme);
        }
        View::Diff(diff) => {
            diff::render_diff_view(f, diff, area, theme);
        }
    }
}
