use crate::core::DiffRef;
use crate::domain::diff::{DiffBlock, FileDiff};
use crate::domain::wrap::wrap_line;

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
struct ChangeRange {
    start: usize,
    end: usize,
}

impl ChangeRange {
    fn len(&self) -> usize {
        self.end.saturating_sub(self.start)
    }
}

pub struct DiffView {
    pub diff_ref: DiffRef,
    pub file: FileDiff,
    pub scroll_offset: usize,
    pub viewport_height: usize,
    pub viewport_width: usize,
    cumulative_screen_rows: Vec<usize>,
    cached_change_ranges: Vec<ChangeRange>,
    focused_change_idx: Option<usize>,
}

impl DiffView {
    pub fn new(diff_ref: DiffRef, file: FileDiff) -> Self {
        Self {
            diff_ref,
            file,
            scroll_offset: 0,
            viewport_height: 0,
            viewport_width: 0,
            cumulative_screen_rows: Vec::new(),
            cached_change_ranges: Vec::new(),
            focused_change_idx: None,
        }
    }

    pub fn set_viewport_dimensions(&mut self, height: usize, width: usize) {
        let width_changed = self.viewport_width != width;
        self.viewport_height = height;
        self.viewport_width = width;
        if width_changed {
            self.recompute_screen_rows();
        }
        self.clamp_scroll_offset();
    }

    fn recompute_screen_rows(&mut self) {
        if self.viewport_width == 0 {
            self.cumulative_screen_rows.clear();
            self.cached_change_ranges.clear();
            return;
        }

        let separator_width = 3;
        let available = self.viewport_width.saturating_sub(separator_width);
        let cell_width = available / 2;
        let max_line_num = self
            .file
            .blocks
            .iter()
            .map(|b| match b {
                DiffBlock::Unchanged(lines) => lines.last().map(|l| l.new_number).unwrap_or(0),
                DiffBlock::Change { old, new } => {
                    let old_max = old.last().map(|l| l.number).unwrap_or(0);
                    let new_max = new.last().map(|l| l.number).unwrap_or(0);
                    old_max.max(new_max)
                }
            })
            .max()
            .unwrap_or(0);
        // Mirror format_cell: "{:>3} " + sign char
        let prefix_width = format!("{:>3} ", max_line_num).chars().count() + 1;
        let content_width = cell_width.saturating_sub(prefix_width);

        let mut cumulative = Vec::new();
        let mut total = 0usize;
        let mut change_ranges = Vec::new();

        for block in &self.file.blocks {
            let line_count = match block {
                DiffBlock::Unchanged(lines) => lines.len(),
                DiffBlock::Change { old, new } => {
                    let len = old.len().max(new.len());
                    let start_screen = total;
                    for i in 0..len {
                        let rows = if content_width == 0 {
                            1
                        } else {
                            let old_rows = old
                                .get(i)
                                .map(|l| wrap_line(&l.text, &[], content_width).len())
                                .unwrap_or(0);
                            let new_rows = new
                                .get(i)
                                .map(|l| wrap_line(&l.text, &[], content_width).len())
                                .unwrap_or(0);
                            old_rows.max(new_rows).max(1)
                        };
                        total += rows;
                        cumulative.push(total);
                    }
                    change_ranges.push(ChangeRange {
                        start: start_screen,
                        end: total,
                    });
                    continue;
                }
            };

            for i in 0..line_count {
                let line = self.get_line_text(block, i);
                let rows = if content_width == 0 {
                    1
                } else {
                    wrap_line(line, &[], content_width).len().max(1)
                };
                total += rows;
                cumulative.push(total);
            }
        }

        self.cumulative_screen_rows = cumulative;
        self.cached_change_ranges = change_ranges;
    }

    fn get_line_text<'a>(&self, block: &'a DiffBlock, index: usize) -> &'a str {
        match block {
            DiffBlock::Unchanged(lines) => lines.get(index).map(|l| l.text.as_str()).unwrap_or(""),
            DiffBlock::Change { old, new } => new
                .get(index)
                .map(|l| l.text.as_str())
                .or_else(|| old.get(index).map(|l| l.text.as_str()))
                .unwrap_or(""),
        }
    }

    fn screen_row_to_logical_line(&self, screen_row: usize) -> usize {
        self.cumulative_screen_rows
            .partition_point(|&cum| cum <= screen_row)
    }

    fn total_screen_rows(&self) -> usize {
        self.cumulative_screen_rows.last().copied().unwrap_or(0)
    }

    pub fn move_down(&mut self, amount: usize) {
        let max_scroll = self.max_scroll_offset();
        self.scroll_offset = self.scroll_offset.saturating_add(amount).min(max_scroll);
    }

    pub fn move_up(&mut self, amount: usize) {
        self.scroll_offset = self.scroll_offset.saturating_sub(amount);
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    pub fn navigate_next_diff(&mut self) -> bool {
        let ranges = self.change_ranges();
        let focus = self.focus_line();
        let Some(current_idx) = ranges.iter().position(|range| focus < range.end) else {
            return false;
        };
        let current_range = ranges[current_idx];

        if focus < current_range.start {
            let target = self.focus_line_for_range_start(current_range);
            if self.set_focus_line(target) {
                self.focused_change_idx = Some(current_idx);
                return true;
            }
            // Scroll was already clamped at max and couldn't reach range.start.
            // If we haven't yet "arrived" at this hunk, lock in now and return true
            // so the caller doesn't cross to the next file prematurely.
            if self.focused_change_idx != Some(current_idx) {
                self.focused_change_idx = Some(current_idx);
                return true;
            }
            return false;
        }

        if self.should_advance_within_range(current_range) {
            let target = self.next_focus_line_within_range(current_range);
            if self.set_focus_line(target) {
                self.focused_change_idx = Some(current_idx);
                return true;
            }
            return false;
        }

        if let Some(next_range) = ranges.get(current_idx + 1).copied() {
            let target = self.focus_line_for_range_start(next_range);
            if self.set_focus_line(target) {
                self.focused_change_idx = Some(current_idx + 1);
                return true;
            }
            return false;
        }

        false
    }

    pub fn navigate_prev_diff(&mut self) -> bool {
        let ranges = self.change_ranges();
        if ranges.is_empty() {
            return false;
        }

        let focus = self.focus_line();
        let Some(current_idx) = ranges.iter().rposition(|range| range.start <= focus) else {
            return false;
        };

        let current_range = ranges[current_idx];
        if focus >= current_range.end {
            let target = self.focus_line_for_range_start(current_range);
            if self.set_focus_line(target) {
                self.focused_change_idx = Some(current_idx);
                return true;
            }
            return false;
        }

        if focus < current_range.end && self.should_rewind_within_range(current_range) {
            let target = self.prev_focus_line_within_range(current_range);
            if self.set_focus_line(target) {
                self.focused_change_idx = Some(current_idx);
                return true;
            }
            return false;
        }

        if current_idx == 0 {
            return false;
        }

        let previous_range = ranges[current_idx - 1];
        if self.set_focus_line(self.focus_line_for_range_start(previous_range)) {
            self.focused_change_idx = Some(current_idx - 1);
            return true;
        }
        false
    }

    pub fn jump_to_first_diff(&mut self) -> bool {
        let Some(first_range) = self.change_ranges().first().copied() else {
            return false;
        };
        let target = self.focus_line_for_range_start(first_range);
        self.set_focus_line(target);
        self.focused_change_idx = Some(0);
        true
    }

    pub fn jump_to_last_diff(&mut self) -> bool {
        let ranges = self.change_ranges();
        let Some(last_range) = ranges.last().copied() else {
            return false;
        };
        let last_idx = ranges.len() - 1;

        let target = if self.viewport_height > 0 && last_range.len() > self.viewport_height {
            self.max_focus_line_for_large_range(last_range)
        } else {
            self.focus_line_for_range_start(last_range)
        };

        self.set_focus_line(target);
        self.focused_change_idx = Some(last_idx);
        true
    }

    pub fn focused_change_idx(&self) -> Option<usize> {
        self.focused_change_idx
    }

    pub fn current_file_line_number(&self) -> Result<u32, String> {
        if self.file.blocks.is_empty() {
            return Err("diff has no blocks".to_string());
        }

        let total_lines = self.file.total_line_count();
        if total_lines == 0 {
            return Err("diff has no lines".to_string());
        }

        let logical_line = self.screen_row_to_logical_line(self.scroll_offset);
        if logical_line >= total_lines {
            return Err("scroll position out of range".to_string());
        }

        let mut row = 0;
        for block in &self.file.blocks {
            match block {
                DiffBlock::Unchanged(lines) => {
                    for line in lines {
                        if row == logical_line {
                            return Ok(line.new_number);
                        }
                        row += 1;
                    }
                }
                DiffBlock::Change { old, new } => {
                    let len = old.len().max(new.len());
                    for i in 0..len {
                        if row == logical_line {
                            if let Some(new_line) = new.get(i) {
                                return Ok(new_line.number);
                            }
                            return self.find_next_new_line_number(row + 1);
                        }
                        row += 1;
                    }
                }
            }
        }

        Err("scroll position out of range".to_string())
    }

    fn clamp_scroll_offset(&mut self) {
        let max_scroll = self.max_scroll_offset();
        if self.scroll_offset > max_scroll {
            self.scroll_offset = max_scroll;
        }
    }

    fn max_scroll_offset(&self) -> usize {
        // At max scroll the last real line sits at the focus point (¼ from top).
        let focus_offset = self.viewport_height / 4;
        (self.total_screen_rows() + focus_offset).saturating_sub(self.viewport_height)
    }

    fn should_advance_within_range(&self, range: ChangeRange) -> bool {
        self.viewport_height > 0
            && range.len() > self.viewport_height
            && self.focus_line() < self.max_focus_line_for_large_range(range)
    }

    fn next_focus_line_within_range(&self, range: ChangeRange) -> usize {
        self.focus_line()
            .saturating_add(self.page_step())
            .min(self.max_focus_line_for_large_range(range))
    }

    fn should_rewind_within_range(&self, range: ChangeRange) -> bool {
        self.viewport_height > 0
            && range.len() > self.viewport_height
            && self.focus_line() > range.start
    }

    fn prev_focus_line_within_range(&self, range: ChangeRange) -> usize {
        self.focus_line()
            .saturating_sub(self.page_step())
            .max(range.start)
    }

    fn change_ranges(&self) -> &[ChangeRange] {
        &self.cached_change_ranges
    }

    fn focus_line_for_range_start(&self, range: ChangeRange) -> usize {
        range.start
    }

    fn max_focus_line_for_large_range(&self, range: ChangeRange) -> usize {
        range.end.saturating_sub(self.viewport_height / 2)
    }

    fn focus_line(&self) -> usize {
        self.scroll_offset
    }

    fn set_focus_line(&mut self, line: usize) -> bool {
        self.set_scroll_offset(line.min(self.max_scroll_offset()))
    }

    fn set_scroll_offset(&mut self, target: usize) -> bool {
        if self.scroll_offset == target {
            return false;
        }
        self.scroll_offset = target;
        true
    }

    fn find_next_new_line_number(&self, start_logical_row: usize) -> Result<u32, String> {
        let mut row = 0;
        for block in &self.file.blocks {
            match block {
                DiffBlock::Unchanged(lines) => {
                    for line in lines {
                        if row >= start_logical_row {
                            return Ok(line.new_number);
                        }
                        row += 1;
                    }
                }
                DiffBlock::Change { old, new } => {
                    let len = old.len().max(new.len());
                    for i in 0..len {
                        if row >= start_logical_row {
                            if let Some(new_line) = new.get(i) {
                                return Ok(new_line.number);
                            }
                        }
                        row += 1;
                    }
                }
            }
        }

        Ok(1)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::{FileDiffInfo, FileStatus};
    use crate::domain::diff::{DiffLine, UnchangedLine};

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
    fn next_diff_advances_within_large_change_block_before_jumping() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Change {
                old: (1..=10).map(|n| diff_line(n, "a")).collect(),
                new: (1..=10).map(|n| diff_line(n, "b")).collect(),
            }],
            6,
        );

        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 3);
        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 5);
        assert!(!view.navigate_next_diff());
    }

    #[test]
    fn prev_diff_rewinds_within_large_change_block_before_previous() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=4).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: (5..=14).map(|n| diff_line(n, "a")).collect(),
                    new: (5..=14).map(|n| diff_line(n, "b")).collect(),
                },
            ],
            6,
        );
        view.scroll_offset = 8;

        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 5);
        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
        assert!(!view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
    }

    #[test]
    fn jump_to_last_diff_lands_on_trailing_page_of_large_hunk() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Change {
                old: (1..=20).map(|n| diff_line(n, "a")).collect(),
                new: (1..=20).map(|n| diff_line(n, "b")).collect(),
            }],
            13,
        );

        assert!(view.jump_to_last_diff());
        assert_eq!(view.scroll_offset, 10);
    }

    #[test]
    fn next_diff_can_still_align_bottom_visible_hunk_with_padding() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=20).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![],
                    new: vec![diff_line(21, "added")],
                },
            ],
            13,
        );

        view.scroll_offset = 8;
        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 11);
    }

    #[test]
    fn next_diff_aligns_hunk_near_quarter_screen() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=20).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(21, "old")],
                    new: vec![diff_line(21, "new")],
                },
                DiffBlock::Unchanged((22..=41).map(|n| unchanged_line(n, n, "tail")).collect()),
            ],
            12,
        );

        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 20);
    }

    #[test]
    fn next_diff_does_not_jump_back_when_hunks_are_close() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=2).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(3, "a")],
                    new: vec![diff_line(3, "b")],
                },
                DiffBlock::Unchanged(vec![unchanged_line(4, 4, "ctx")]),
                DiffBlock::Change {
                    old: vec![diff_line(5, "c")],
                    new: vec![diff_line(5, "d")],
                },
                DiffBlock::Unchanged((6..=25).map(|n| unchanged_line(n, n, "tail")).collect()),
            ],
            12,
        );

        assert!(view.navigate_next_diff());
        let first = view.scroll_offset;
        assert!(view.navigate_next_diff());
        let second = view.scroll_offset;

        assert!(second > first);
        assert!(!view.navigate_next_diff());
    }

    #[test]
    fn prev_diff_from_below_last_hunk_lands_on_last_hunk() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=4).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(5, "a")],
                    new: vec![diff_line(5, "b")],
                },
                DiffBlock::Unchanged((6..=9).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(10, "c")],
                    new: vec![diff_line(10, "d")],
                },
                DiffBlock::Unchanged((11..=20).map(|n| unchanged_line(n, n, "tail")).collect()),
            ],
            12,
        );

        view.scroll_offset = 8;
        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
    }

    #[test]
    fn current_file_line_number_uses_new_side_for_unchanged_and_added_rows() {
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

        view.scroll_offset = 1;
        assert_eq!(view.current_file_line_number().unwrap(), 21);

        view.scroll_offset = 3;
        assert_eq!(view.current_file_line_number().unwrap(), 23);
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

        view.scroll_offset = 0;
        assert_eq!(view.current_file_line_number().unwrap(), 1);
    }
}
