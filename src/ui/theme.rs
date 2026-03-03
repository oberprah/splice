use ratatui::style::{Color, Modifier, Style};
use terminal_colorsaurus::{theme_mode, QueryOptions, ThemeMode};

use crate::domain::highlight::HighlightKind;

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
    pub text_selected: Style,
    pub text_muted: Style,
    pub text_muted_selected: Style,
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
    pub syntax: SyntaxColors,
    pub additions: Style,
    pub additions_selected: Style,
    pub deletions: Style,
    pub deletions_selected: Style,
    pub file_status_added: Style,
    pub file_status_added_selected: Style,
    pub file_status_modified: Style,
    pub file_status_modified_selected: Style,
    pub file_status_deleted: Style,
    pub file_status_deleted_selected: Style,
    pub file_status_renamed: Style,
    pub file_status_renamed_selected: Style,
}

#[derive(Debug, Clone, Copy)]
pub struct DiffColors {
    pub bg: Color,
    pub bg_bright: Color,
    pub fg: Color,
}

#[derive(Debug, Clone, Copy)]
pub struct SyntaxColors {
    pub keyword: Color,
    pub string: Color,
    pub comment: Color,
    pub r#type: Color,
    pub function: Color,
    pub constant: Color,
    pub variable: Color,
    pub number: Color,
    pub operator: Color,
    pub property: Color,
    pub punctuation: Color,
}

impl Theme {
    pub fn detect_theme() -> Self {
        match theme_mode(QueryOptions::default()) {
            Ok(ThemeMode::Dark) => Self::dark(),
            Ok(ThemeMode::Light) => Self::light(),
            Err(_) => Self::dark(),
        }
    }

    pub fn dark() -> Self {
        Self {
            variant: ThemeVariant::Dark,
            text: Style::default().fg(Color::Rgb(252, 252, 252)),
            text_selected: Style::default()
                .fg(Color::Rgb(255, 255, 255))
                .add_modifier(Modifier::BOLD),
            text_muted: Style::default().fg(Color::Rgb(128, 128, 128)),
            text_muted_selected: Style::default()
                .fg(Color::Rgb(200, 200, 200))
                .add_modifier(Modifier::BOLD),
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
            syntax: SyntaxColors {
                keyword: Color::Rgb(0xc6, 0x92, 0xff),
                string: Color::Rgb(0xa5, 0xe0, 0x75),
                comment: Color::Rgb(0x7f, 0x84, 0x90),
                r#type: Color::Rgb(0x66, 0xd9, 0xef),
                function: Color::Rgb(0x82, 0xaa, 0xff),
                constant: Color::Rgb(0xff, 0xcb, 0x6b),
                variable: Color::Rgb(0xe6, 0xe6, 0xe6),
                number: Color::Rgb(0xff, 0x9e, 0x64),
                operator: Color::Rgb(0x89, 0xdd, 0xff),
                property: Color::Rgb(0xc3, 0xe8, 0xff),
                punctuation: Color::Rgb(0xb0, 0xb8, 0xc0),
            },
            additions: Style::default().fg(Color::Rgb(0, 255, 0)),
            additions_selected: Style::default().fg(Color::Rgb(64, 255, 64)),
            deletions: Style::default().fg(Color::Rgb(255, 0, 0)),
            deletions_selected: Style::default().fg(Color::Rgb(255, 96, 96)),
            file_status_added: Style::default().fg(Color::Green),
            file_status_added_selected: Style::default().fg(Color::Rgb(120, 255, 120)),
            file_status_modified: Style::default().fg(Color::Yellow),
            file_status_modified_selected: Style::default().fg(Color::Rgb(255, 235, 120)),
            file_status_deleted: Style::default().fg(Color::Red),
            file_status_deleted_selected: Style::default().fg(Color::Rgb(255, 120, 120)),
            file_status_renamed: Style::default().fg(Color::Cyan),
            file_status_renamed_selected: Style::default().fg(Color::Rgb(120, 255, 255)),
        }
    }

    pub fn light() -> Self {
        Self {
            variant: ThemeVariant::Light,
            text: Style::default().fg(Color::Rgb(58, 58, 58)),
            text_selected: Style::default()
                .fg(Color::Rgb(0, 0, 0))
                .add_modifier(Modifier::BOLD),
            text_muted: Style::default().fg(Color::Rgb(108, 108, 108)),
            text_muted_selected: Style::default()
                .fg(Color::Rgb(58, 58, 58))
                .add_modifier(Modifier::BOLD),
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
            syntax: SyntaxColors {
                keyword: Color::Rgb(0x6a, 0x1b, 0x9a),
                string: Color::Rgb(0x2e, 0x7d, 0x32),
                comment: Color::Rgb(0x78, 0x7d, 0x86),
                r#type: Color::Rgb(0x00, 0x79, 0x96),
                function: Color::Rgb(0x15, 0x65, 0xc0),
                constant: Color::Rgb(0xb2, 0x87, 0x04),
                variable: Color::Rgb(0x42, 0x42, 0x42),
                number: Color::Rgb(0xd8, 0x43, 0x15),
                operator: Color::Rgb(0x00, 0x83, 0xa3),
                property: Color::Rgb(0x00, 0x61, 0x6b),
                punctuation: Color::Rgb(0x8a, 0x8f, 0x98),
            },
            additions: Style::default().fg(Color::Rgb(0, 128, 0)),
            additions_selected: Style::default().fg(Color::Rgb(0, 110, 0)),
            deletions: Style::default().fg(Color::Rgb(178, 34, 34)),
            deletions_selected: Style::default().fg(Color::Rgb(150, 20, 20)),
            file_status_added: Style::default().fg(Color::Green),
            file_status_added_selected: Style::default().fg(Color::Rgb(0, 120, 0)),
            file_status_modified: Style::default().fg(Color::Yellow),
            file_status_modified_selected: Style::default().fg(Color::Rgb(180, 140, 0)),
            file_status_deleted: Style::default().fg(Color::Red),
            file_status_deleted_selected: Style::default().fg(Color::Rgb(160, 20, 20)),
            file_status_renamed: Style::default().fg(Color::Cyan),
            file_status_renamed_selected: Style::default().fg(Color::Rgb(0, 150, 150)),
        }
    }

    pub fn default_variant(variant: ThemeVariant) -> Self {
        match variant {
            ThemeVariant::Dark => Self::dark(),
            ThemeVariant::Light => Self::light(),
        }
    }

    pub fn syntax_color(&self, kind: HighlightKind) -> Color {
        match kind {
            HighlightKind::Keyword => self.syntax.keyword,
            HighlightKind::String => self.syntax.string,
            HighlightKind::Comment => self.syntax.comment,
            HighlightKind::Type => self.syntax.r#type,
            HighlightKind::Function => self.syntax.function,
            HighlightKind::Constant => self.syntax.constant,
            HighlightKind::Variable => self.syntax.variable,
            HighlightKind::Number => self.syntax.number,
            HighlightKind::Operator => self.syntax.operator,
            HighlightKind::Property => self.syntax.property,
            HighlightKind::Punctuation => self.syntax.punctuation,
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
