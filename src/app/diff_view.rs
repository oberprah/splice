use crate::core::DiffRef;
use crate::domain::diff::layout::{build_rows, ScreenRow};
use crate::domain::diff::FileDiff;

use super::viewport::{
    clamp_scroll_offset, page_step, update_viewport, visible_content, Viewport, ViewportAction,
    VisibleContent,
};

pub struct DiffView {
    pub diff_ref: DiffRef,
    pub file: FileDiff,
    rows: Vec<ScreenRow>,
    hunks: Vec<crate::domain::diff::layout::HunkRange>,
    pub viewport: Viewport,
}

impl DiffView {
    pub fn new(diff_ref: DiffRef, file: FileDiff) -> Self {
        Self {
            diff_ref,
            file,
            rows: Vec::new(),
            hunks: Vec::new(),
            viewport: Viewport {
                scroll_offset: 0,
                height: 0,
                width: 0,
                active_hunk: None,
                total_rows: 0,
            },
        }
    }

    pub fn set_viewport_dimensions(&mut self, height: usize, width: usize) {
        let width_changed = self.viewport.width != width;
        self.viewport.height = height;
        self.viewport.width = width;
        if width_changed {
            let (rows, hunks) = build_rows(&self.file, width);
            self.viewport.total_rows = rows.len();
            self.rows = rows;
            self.hunks = hunks;
        }
        let clamped = clamp_scroll_offset(&self.viewport);
        self.viewport.scroll_offset = clamped;
    }

    /// Apply a viewport action and store the new viewport state.
    /// Returns `true` if navigation happened (for hunk actions: whether it moved;
    /// for scroll actions: always `true`).
    pub fn update(&mut self, action: ViewportAction) -> bool {
        let (new_viewport, did_navigate) = update_viewport(&self.viewport, action, &self.hunks);
        self.viewport = new_viewport;
        did_navigate
    }

    pub fn visible_content(&self) -> VisibleContent<'_> {
        visible_content(&self.rows, &self.hunks, &self.viewport)
    }

    pub fn page_step(&self) -> usize {
        page_step(&self.viewport)
    }

    pub fn viewport_height(&self) -> usize {
        self.viewport.height
    }

    pub fn viewport_width(&self) -> usize {
        self.viewport.width
    }

    pub fn scroll_offset(&self) -> usize {
        self.viewport.scroll_offset
    }

    pub fn current_file_line_number(&self) -> Result<u32, String> {
        if self.rows.is_empty() {
            return Err("diff has no rows".to_string());
        }

        let start = self.viewport.scroll_offset;

        // Walk forward from scroll_offset to find a row with a new-side line number
        for row in self.rows.get(start..).unwrap_or(&[]) {
            if let Some(n) = row.right.line_number {
                return Ok(n);
            }
            if let Some(n) = row.left.line_number {
                return Ok(n);
            }
        }

        // Walk from beginning if nothing found forward
        for row in &self.rows {
            if let Some(n) = row.right.line_number {
                return Ok(n);
            }
            if let Some(n) = row.left.line_number {
                return Ok(n);
            }
        }

        Err("no line number found in diff".to_string())
    }
}

#[cfg(test)]
impl DiffView {
    pub fn set_scroll_offset_for_test(&mut self, offset: usize) {
        self.viewport.scroll_offset = offset;
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::{FileDiffInfo, FileStatus};
    use crate::domain::diff::{DiffBlock, DiffLine, UnchangedLine};

    fn unchanged_line(old_number: u32, new_number: u32, text: &str) -> UnchangedLine {
        UnchangedLine {
            old_number,
            new_number,
            text: text.to_string(),
            tokens: vec![],
        }
    }

    fn diff_line(number: u32, text: &str) -> DiffLine {
        DiffLine {
            number,
            text: text.to_string(),
            tokens: vec![],
        }
    }

    fn view_with_blocks(blocks: Vec<DiffBlock>, viewport_height: usize) -> DiffView {
        let mut view = DiffView::new(
            DiffRef::Uncommitted(crate::core::UncommittedType::All),
            FileDiff {
                info: FileDiffInfo {
                    path: "src/main.rs".to_string(),
                    old_path: None,
                    status: FileStatus::Modified,
                    additions: 1,
                    deletions: 1,
                    is_binary: false,
                },
                blocks,
            },
        );
        view.set_viewport_dimensions(viewport_height, 80);
        view
    }

    #[test]
    fn current_file_line_number_returns_right_side_line_number() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged(vec![
                    unchanged_line(10, 20, "a"),
                    unchanged_line(11, 21, "b"),
                ]),
                DiffBlock::Change {
                    old: vec![diff_line(12, "old")],
                    new: vec![diff_line(22, "new"), diff_line(23, "newer")],
                },
            ],
            8,
        );

        view.viewport.scroll_offset = 1;
        assert_eq!(view.current_file_line_number().unwrap(), 21);
    }

    #[test]
    fn current_file_line_number_for_removed_row_falls_forward() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Change {
                    old: vec![diff_line(1, "removed")],
                    new: vec![],
                },
                DiffBlock::Unchanged(vec![unchanged_line(2, 1, "kept")]),
            ],
            8,
        );

        view.viewport.scroll_offset = 0;
        // Row 0: left=Removed(1), right=Empty -> no right line_number, falls through to row 1
        // Row 1: right has line_number=1
        assert_eq!(view.current_file_line_number().unwrap(), 1);
    }
}
