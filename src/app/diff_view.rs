use crate::core::{DiffSource, FileChange};
use crate::domain::diff::{DiffBlock, FileDiff};
use crate::domain::highlight::DiffHighlights;

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
    pub source: DiffSource,
    pub file: FileChange,
    pub diff: FileDiff,
    pub highlights: DiffHighlights,
    pub scroll_offset: usize,
    pub viewport_height: usize,
}

impl DiffView {
    pub fn new(
        source: DiffSource,
        file: FileChange,
        diff: FileDiff,
        highlights: DiffHighlights,
    ) -> Self {
        Self {
            source,
            file,
            diff,
            highlights,
            scroll_offset: 0,
            viewport_height: 0,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
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
            return self.set_focus_line(self.focus_line_for_range_start(current_range));
        }

        if self.should_advance_within_range(current_range) {
            return self.set_focus_line(self.next_focus_line_within_range(current_range));
        }

        if let Some(next_range) = ranges.get(current_idx + 1) {
            return self.set_focus_line(self.focus_line_for_range_start(*next_range));
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
            return self.set_focus_line(self.focus_line_for_range_start(current_range));
        }

        if focus < current_range.end && self.should_rewind_within_range(current_range) {
            return self.set_focus_line(self.prev_focus_line_within_range(current_range));
        }

        if current_idx == 0 {
            return false;
        }

        let previous_range = ranges[current_idx - 1];

        self.set_focus_line(self.focus_line_for_range_start(previous_range))
    }

    pub fn jump_to_first_diff(&mut self) -> bool {
        let Some(first_range) = self.change_ranges().first().copied() else {
            return false;
        };
        self.set_focus_line(self.focus_line_for_range_start(first_range))
    }

    pub fn jump_to_last_diff(&mut self) -> bool {
        let Some(last_range) = self.change_ranges().last().copied() else {
            return false;
        };

        let target = if self.viewport_height > 0 && last_range.len() > self.viewport_height {
            self.max_focus_line_for_large_range(last_range)
        } else {
            self.focus_line_for_range_start(last_range)
        };

        self.set_focus_line(target)
    }

    fn clamp_scroll_offset(&mut self) {
        let max_scroll = self.max_scroll_offset();
        if self.scroll_offset > max_scroll {
            self.scroll_offset = max_scroll;
        }
    }

    fn max_scroll_offset(&self) -> usize {
        self.diff.total_line_count()
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

    fn change_ranges(&self) -> Vec<ChangeRange> {
        let mut ranges = Vec::new();
        let mut row = 0;

        for block in &self.diff.blocks {
            let len = match block {
                DiffBlock::Unchanged(unchanged) => unchanged.lines.len(),
                DiffBlock::Change(change) => {
                    let block_len = change.old_lines.len().max(change.new_lines.len());
                    ranges.push(ChangeRange {
                        start: row,
                        end: row + block_len,
                    });
                    block_len
                }
            };
            row += len;
        }

        ranges
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
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::FileStatus;
    use crate::domain::diff::{ChangeBlock, DiffMeta, UnchangedBlock};

    fn file_change(path: &str) -> FileChange {
        FileChange {
            path: path.to_string(),
            status: FileStatus::Modified,
            additions: 1,
            deletions: 1,
            is_binary: false,
        }
    }

    fn view_with_blocks(blocks: Vec<DiffBlock>, viewport_height: usize) -> DiffView {
        let mut view = DiffView::new(
            DiffSource::Uncommitted(crate::core::UncommittedType::All),
            file_change("src/main.rs"),
            FileDiff {
                meta: DiffMeta {
                    path: "src/main.rs".to_string(),
                    additions: 1,
                    deletions: 1,
                },
                blocks,
            },
            DiffHighlights::default(),
        );
        view.set_viewport_height(viewport_height);
        view
    }

    #[test]
    fn next_diff_advances_within_large_change_block_before_jumping() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Change(ChangeBlock {
                old_start: 1,
                new_start: 1,
                old_lines: vec!["a".to_string(); 10],
                new_lines: vec!["b".to_string(); 10],
            })],
            6,
        );

        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 3);
        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 6);
        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 7);
        assert!(!view.navigate_next_diff());
    }

    #[test]
    fn prev_diff_rewinds_within_large_change_block_before_previous() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 1,
                    new_start: 1,
                    lines: vec!["ctx".to_string(); 4],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 5,
                    new_start: 5,
                    old_lines: vec!["a".to_string(); 10],
                    new_lines: vec!["b".to_string(); 10],
                }),
            ],
            6,
        );
        view.scroll_offset = 9;

        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 6);
        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
        assert!(!view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
    }

    #[test]
    fn jump_to_last_diff_lands_on_trailing_page_of_large_hunk() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Change(ChangeBlock {
                old_start: 1,
                new_start: 1,
                old_lines: vec!["a".to_string(); 20],
                new_lines: vec!["b".to_string(); 20],
            })],
            13,
        );

        assert!(view.jump_to_last_diff());
        assert_eq!(view.scroll_offset, 14);
    }

    #[test]
    fn next_diff_can_still_align_bottom_visible_hunk_with_padding() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 1,
                    new_start: 1,
                    lines: vec!["ctx".to_string(); 20],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 21,
                    new_start: 21,
                    old_lines: Vec::new(),
                    new_lines: vec!["added".to_string()],
                }),
            ],
            13,
        );

        view.scroll_offset = 8;
        assert!(view.navigate_next_diff());
        assert_eq!(view.scroll_offset, 20);
    }

    #[test]
    fn next_diff_aligns_hunk_near_quarter_screen() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 1,
                    new_start: 1,
                    lines: vec!["ctx".to_string(); 20],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 21,
                    new_start: 21,
                    old_lines: vec!["old".to_string()],
                    new_lines: vec!["new".to_string()],
                }),
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 22,
                    new_start: 22,
                    lines: vec!["tail".to_string(); 20],
                }),
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
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 1,
                    new_start: 1,
                    lines: vec!["ctx".to_string(); 2],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 3,
                    new_start: 3,
                    old_lines: vec!["a".to_string()],
                    new_lines: vec!["b".to_string()],
                }),
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 4,
                    new_start: 4,
                    lines: vec!["ctx".to_string(); 1],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 5,
                    new_start: 5,
                    old_lines: vec!["c".to_string()],
                    new_lines: vec!["d".to_string()],
                }),
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 6,
                    new_start: 6,
                    lines: vec!["tail".to_string(); 20],
                }),
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
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 1,
                    new_start: 1,
                    lines: vec!["ctx".to_string(); 4],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 5,
                    new_start: 5,
                    old_lines: vec!["a".to_string()],
                    new_lines: vec!["b".to_string()],
                }),
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 6,
                    new_start: 6,
                    lines: vec!["ctx".to_string(); 4],
                }),
                DiffBlock::Change(ChangeBlock {
                    old_start: 10,
                    new_start: 10,
                    old_lines: vec!["c".to_string()],
                    new_lines: vec!["d".to_string()],
                }),
                DiffBlock::Unchanged(UnchangedBlock {
                    old_start: 11,
                    new_start: 11,
                    lines: vec!["tail".to_string(); 10],
                }),
            ],
            12,
        );

        view.scroll_offset = 20;
        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 9);
    }
}
