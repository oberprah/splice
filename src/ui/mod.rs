mod diff;
mod files;
mod log;
mod theme;

pub use theme::*;

use crate::app::{App, View};
use ratatui::{prelude::*, widgets::Paragraph};

const LOG_AREA_X: u16 = 2;
const LOG_AREA_Y: u16 = 0;
const LOG_AREA_RIGHT_MARGIN: u16 = 4;
const LOG_AREA_BOTTOM_MARGIN: u16 = 0;

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

    let viewport_height = area.height.saturating_sub(1) as usize;
    app.set_viewport_height(viewport_height);

    let theme = Theme::default();

    match &app.view {
        View::Log(log) => {
            log::render_log_view(
                f,
                &log.commits,
                &log.graph_layout,
                log.selected,
                log.scroll_offset,
                area,
            );
        }
        View::Files(files) => {
            files::render_files_view(f, files, area);
        }
        View::Diff(diff) => {
            diff::render_diff_view(f, diff, area, &theme);
        }
    }
}
