use crate::core::{DiffSource, FileChange};
use crate::domain::diff::{DiffBlock, FileDiff};

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
    pub scroll_offset: usize,
    pub viewport_height: usize,
}

impl DiffView {
    pub fn new(source: DiffSource, file: FileChange, diff: FileDiff) -> Self {
        Self {
            source,
            file,
            diff,
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
        let Some(first_range) = ranges.first() else {
            return false;
        };

        let current = self.scroll_offset;
        if current < first_range.start {
            self.scroll_offset = first_range.start.min(self.max_scroll_offset());
            return true;
        }

        for (idx, range) in ranges.iter().enumerate() {
            if current < range.start {
                self.scroll_offset = range.start.min(self.max_scroll_offset());
                return true;
            }

            if current >= range.end {
                continue;
            }

            if self.should_advance_within_range(*range) {
                self.scroll_offset = self.next_offset_within_range(*range);
                return true;
            }

            if let Some(next_range) = ranges.get(idx + 1) {
                self.scroll_offset = next_range.start.min(self.max_scroll_offset());
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

        let current = self.scroll_offset;

        if let Some((_, range)) = ranges
            .iter()
            .enumerate()
            .find(|(_, range)| current >= range.start && current < range.end)
        {
            if self.should_rewind_within_range(*range) {
                self.scroll_offset = self.prev_offset_within_range(*range);
                return true;
            }
        }

        let Some(previous_range) = ranges.iter().rfind(|range| range.start < current) else {
            return false;
        };

        self.scroll_offset = previous_range.start.min(self.max_scroll_offset());
        true
    }

    fn clamp_scroll_offset(&mut self) {
        let max_scroll = self.max_scroll_offset();
        if self.scroll_offset > max_scroll {
            self.scroll_offset = max_scroll;
        }
    }

    fn max_scroll_offset(&self) -> usize {
        let total = self.diff.total_line_count();
        total.saturating_sub(self.viewport_height)
    }

    fn should_advance_within_range(&self, range: ChangeRange) -> bool {
        self.viewport_height > 0
            && range.len() > self.viewport_height
            && self.scroll_offset < range.end.saturating_sub(self.viewport_height)
    }

    fn next_offset_within_range(&self, range: ChangeRange) -> usize {
        let max_offset = range.end.saturating_sub(self.viewport_height);
        self.scroll_offset
            .saturating_add(self.page_step())
            .min(max_offset)
    }

    fn should_rewind_within_range(&self, range: ChangeRange) -> bool {
        self.viewport_height > 0
            && range.len() > self.viewport_height
            && self.scroll_offset > range.start
    }

    fn prev_offset_within_range(&self, range: ChangeRange) -> usize {
        self.scroll_offset
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
        assert_eq!(view.scroll_offset, 4);
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
        view.scroll_offset = 8;

        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 5);
        assert!(view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
        assert!(!view.navigate_prev_diff());
        assert_eq!(view.scroll_offset, 4);
    }
}
