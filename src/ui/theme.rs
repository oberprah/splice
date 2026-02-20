use ratatui::style::{Color, Modifier, Style};
use terminal_colorsaurus::{color_scheme, ColorScheme, QueryOptions};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum ThemeVariant {
    #[default]
    Dark,
    Light,
}

#[derive(Debug, Clone, Copy)]
pub struct Theme {
    pub variant: ThemeVariant,
    pub text: Style,
    pub text_muted: Style,
    pub selection: Style,
    pub hash: Style,
    pub hash_selected: Style,
    pub message: Style,
    pub message_selected: Style,
    pub refs: Style,
    pub author: Style,
    pub time: Style,
    pub diff_added: DiffColors,
    pub diff_removed: DiffColors,
    pub diff_changed: DiffColors,
    pub additions: Style,
    pub deletions: Style,
    pub file_status_added: Style,
    pub file_status_modified: Style,
    pub file_status_deleted: Style,
    pub file_status_renamed: Style,
}

#[derive(Debug, Clone, Copy)]
pub struct DiffColors {
    pub bg: Color,
    pub bg_bright: Color,
    pub fg: Color,
}

impl Theme {
    pub fn detect_theme() -> Self {
        match color_scheme(QueryOptions::default()) {
            Ok(ColorScheme::Dark) => Self::dark(),
            Ok(ColorScheme::Light) => Self::light(),
            Err(_) => Self::dark(),
        }
    }

    pub fn dark() -> Self {
        Self {
            variant: ThemeVariant::Dark,
            text: Style::default().fg(Color::Rgb(252, 252, 252)),
            text_muted: Style::default().fg(Color::Rgb(128, 128, 128)),
            selection: Style::default().add_modifier(Modifier::BOLD),
            hash: Style::default().fg(Color::Rgb(255, 175, 0)),
            hash_selected: Style::default()
                .fg(Color::Rgb(255, 215, 0))
                .add_modifier(Modifier::BOLD),
            message: Style::default().fg(Color::Rgb(210, 210, 210)),
            message_selected: Style::default()
                .fg(Color::Rgb(255, 255, 255))
                .add_modifier(Modifier::BOLD),
            refs: Style::default().fg(Color::Blue),
            author: Style::default().fg(Color::Rgb(95, 215, 175)),
            time: Style::default().fg(Color::Rgb(128, 128, 128)),
            diff_added: DiffColors {
                bg: Color::Rgb(0x1e, 0x3a, 0x1e),
                bg_bright: Color::Rgb(0x26, 0x4d, 0x26),
                fg: Color::Rgb(0x52, 0xff, 0x52),
            },
            diff_removed: DiffColors {
                bg: Color::Rgb(0x3a, 0x1e, 0x1e),
                bg_bright: Color::Rgb(0x4d, 0x26, 0x26),
                fg: Color::Rgb(0xff, 0x52, 0x52),
            },
            diff_changed: DiffColors {
                bg: Color::Rgb(0x1e, 0x2a, 0x3a),
                bg_bright: Color::Rgb(0x26, 0x36, 0x4d),
                fg: Color::Rgb(0x52, 0x9c, 0xff),
            },
            additions: Style::default().fg(Color::Rgb(0, 255, 0)),
            deletions: Style::default().fg(Color::Rgb(255, 0, 0)),
            file_status_added: Style::default().fg(Color::Green),
            file_status_modified: Style::default().fg(Color::Yellow),
            file_status_deleted: Style::default().fg(Color::Red),
            file_status_renamed: Style::default().fg(Color::Cyan),
        }
    }

    pub fn light() -> Self {
        Self {
            variant: ThemeVariant::Light,
            text: Style::default().fg(Color::Rgb(58, 58, 58)),
            text_muted: Style::default().fg(Color::Rgb(108, 108, 108)),
            selection: Style::default().add_modifier(Modifier::BOLD),
            hash: Style::default().fg(Color::Rgb(215, 135, 0)),
            hash_selected: Style::default()
                .fg(Color::Rgb(255, 135, 0))
                .add_modifier(Modifier::BOLD),
            message: Style::default().fg(Color::Rgb(58, 58, 58)),
            message_selected: Style::default()
                .fg(Color::Rgb(0, 0, 0))
                .add_modifier(Modifier::BOLD),
            refs: Style::default().fg(Color::Blue),
            author: Style::default().fg(Color::Rgb(0, 175, 135)),
            time: Style::default().fg(Color::Rgb(108, 108, 108)),
            diff_added: DiffColors {
                bg: Color::Rgb(0xe8, 0xf5, 0xe9),
                bg_bright: Color::Rgb(0xc8, 0xe6, 0xc9),
                fg: Color::Rgb(0x1b, 0x5e, 0x20),
            },
            diff_removed: DiffColors {
                bg: Color::Rgb(0xff, 0xeb, 0xee),
                bg_bright: Color::Rgb(0xff, 0xcd, 0xd2),
                fg: Color::Rgb(0xb7, 0x1c, 0x1c),
            },
            diff_changed: DiffColors {
                bg: Color::Rgb(0xe3, 0xf2, 0xfd),
                bg_bright: Color::Rgb(0xbb, 0xde, 0xfb),
                fg: Color::Rgb(0x0d, 0x47, 0xa1),
            },
            additions: Style::default().fg(Color::Rgb(0, 128, 0)),
            deletions: Style::default().fg(Color::Rgb(178, 34, 34)),
            file_status_added: Style::default().fg(Color::Green),
            file_status_modified: Style::default().fg(Color::Yellow),
            file_status_deleted: Style::default().fg(Color::Red),
            file_status_renamed: Style::default().fg(Color::Cyan),
        }
    }

    pub fn default_variant(variant: ThemeVariant) -> Self {
        match variant {
            ThemeVariant::Dark => Self::dark(),
            ThemeVariant::Light => Self::light(),
        }
    }
}

impl Default for Theme {
    fn default() -> Self {
        Self::dark()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn dark_theme_has_green_diff_added_colors() {
        let theme = Theme::dark();
        assert_eq!(theme.diff_added.bg, Color::Rgb(0x1e, 0x3a, 0x1e));
        assert_eq!(theme.diff_added.bg_bright, Color::Rgb(0x26, 0x4d, 0x26));
        assert_eq!(theme.diff_added.fg, Color::Rgb(0x52, 0xff, 0x52));
    }

    #[test]
    fn dark_theme_has_red_diff_removed_colors() {
        let theme = Theme::dark();
        assert_eq!(theme.diff_removed.bg, Color::Rgb(0x3a, 0x1e, 0x1e));
        assert_eq!(theme.diff_removed.bg_bright, Color::Rgb(0x4d, 0x26, 0x26));
        assert_eq!(theme.diff_removed.fg, Color::Rgb(0xff, 0x52, 0x52));
    }

    #[test]
    fn dark_theme_has_blue_diff_changed_colors() {
        let theme = Theme::dark();
        assert_eq!(theme.diff_changed.bg, Color::Rgb(0x1e, 0x2a, 0x3a));
        assert_eq!(theme.diff_changed.bg_bright, Color::Rgb(0x26, 0x36, 0x4d));
        assert_eq!(theme.diff_changed.fg, Color::Rgb(0x52, 0x9c, 0xff));
    }

    #[test]
    fn light_theme_has_green_diff_added_colors() {
        let theme = Theme::light();
        assert_eq!(theme.diff_added.bg, Color::Rgb(0xe8, 0xf5, 0xe9));
        assert_eq!(theme.diff_added.bg_bright, Color::Rgb(0xc8, 0xe6, 0xc9));
        assert_eq!(theme.diff_added.fg, Color::Rgb(0x1b, 0x5e, 0x20));
    }

    #[test]
    fn light_theme_has_red_diff_removed_colors() {
        let theme = Theme::light();
        assert_eq!(theme.diff_removed.bg, Color::Rgb(0xff, 0xeb, 0xee));
        assert_eq!(theme.diff_removed.bg_bright, Color::Rgb(0xff, 0xcd, 0xd2));
        assert_eq!(theme.diff_removed.fg, Color::Rgb(0xb7, 0x1c, 0x1c));
    }

    #[test]
    fn light_theme_has_blue_diff_changed_colors() {
        let theme = Theme::light();
        assert_eq!(theme.diff_changed.bg, Color::Rgb(0xe3, 0xf2, 0xfd));
        assert_eq!(theme.diff_changed.bg_bright, Color::Rgb(0xbb, 0xde, 0xfb));
        assert_eq!(theme.diff_changed.fg, Color::Rgb(0x0d, 0x47, 0xa1));
    }

    #[test]
    fn dark_theme_is_default() {
        let theme = Theme::default();
        assert_eq!(theme.variant, ThemeVariant::Dark);
    }

    #[test]
    fn default_variant_returns_correct_theme() {
        let dark = Theme::default_variant(ThemeVariant::Dark);
        assert_eq!(dark.variant, ThemeVariant::Dark);

        let light = Theme::default_variant(ThemeVariant::Light);
        assert_eq!(light.variant, ThemeVariant::Light);
    }

    #[test]
    fn diff_colors_distinct_between_types() {
        let theme = Theme::dark();
        assert_ne!(theme.diff_added.bg, theme.diff_removed.bg);
        assert_ne!(theme.diff_added.bg, theme.diff_changed.bg);
        assert_ne!(theme.diff_removed.bg, theme.diff_changed.bg);
    }

    #[test]
    fn dark_theme_has_darker_backgrounds_than_light() {
        let dark = Theme::dark();
        let light = Theme::light();

        fn sum_rgb(c: Color) -> u8 {
            match c {
                Color::Rgb(r, g, b) => r.wrapping_add(g).wrapping_add(b),
                _ => 0,
            }
        }

        assert!(sum_rgb(dark.diff_added.bg) < sum_rgb(light.diff_added.bg));
        assert!(sum_rgb(dark.diff_removed.bg) < sum_rgb(light.diff_removed.bg));
        assert!(sum_rgb(dark.diff_changed.bg) < sum_rgb(light.diff_changed.bg));
    }
}
