use crate::core::DiffRef;
use crate::domain::diff::layout::{build_rows, HunkRange, ScreenRow};
use crate::domain::diff::FileDiff;

pub struct Viewport {
    pub scroll_offset: usize,
    pub height: usize,
    pub width: usize,
    pub active_hunk: Option<usize>,
}

pub struct DiffView {
    pub diff_ref: DiffRef,
    pub file: FileDiff,
    rows: Vec<ScreenRow>,
    hunks: Vec<HunkRange>,
    viewport: Viewport,
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
            },
        }
    }

    pub fn set_viewport_dimensions(&mut self, height: usize, width: usize) {
        let width_changed = self.viewport.width != width;
        self.viewport.height = height;
        self.viewport.width = width;
        if width_changed {
            let (rows, hunks) = build_rows(&self.file, width);
            self.rows = rows;
            self.hunks = hunks;
        }
        self.clamp_scroll_offset();
    }

    pub fn visible_rows(&self) -> &[ScreenRow] {
        let start = self.viewport.scroll_offset;
        let end = (start + self.viewport.height).min(self.rows.len());
        &self.rows[start..end]
    }

    pub fn visible_row_offset(&self) -> usize {
        self.viewport.scroll_offset
    }

    pub fn is_row_in_active_hunk(&self, absolute_row_idx: usize) -> bool {
        if let Some(hunk_idx) = self.viewport.active_hunk {
            if let Some(hunk) = self.hunks.get(hunk_idx) {
                return absolute_row_idx >= hunk.start && absolute_row_idx < hunk.end;
            }
        }
        false
    }

    pub fn move_down(&mut self, amount: usize) {
        let max_scroll = self.max_scroll_offset();
        self.viewport.scroll_offset = self
            .viewport
            .scroll_offset
            .saturating_add(amount)
            .min(max_scroll);
        self.viewport.active_hunk = None;
    }

    pub fn move_up(&mut self, amount: usize) {
        self.viewport.scroll_offset = self.viewport.scroll_offset.saturating_sub(amount);
        self.viewport.active_hunk = None;
    }

    pub fn page_step(&self) -> usize {
        (self.viewport.height / 2).max(1)
    }

    pub fn max_scroll_offset(&self) -> usize {
        self.rows.len().saturating_sub(self.viewport.height)
    }

    pub fn navigate_next_hunk(&mut self) -> bool {
        if self.hunks.is_empty() {
            return false;
        }

        match self.viewport.active_hunk {
            None => {
                // Find the first hunk whose end is strictly after scroll_offset.
                // This matches the old "focus < range.end" logic where focus = scroll_offset.
                let Some(idx) = self
                    .hunks
                    .iter()
                    .position(|h| h.end > self.viewport.scroll_offset)
                else {
                    return false;
                };
                let hunk = self.hunks[idx];

                if self.viewport.scroll_offset < hunk.start {
                    // scroll is before this hunk; jump to its start
                    let target = self.scroll_for_hunk_start(hunk.start);
                    self.viewport.scroll_offset = target;
                    self.viewport.active_hunk = Some(idx);
                    return true;
                }

                // scroll is within or past this hunk's start — treat as "already at hunk",
                // check for advance-within-hunk or skip to next hunk.
                if self.should_advance_within_hunk(hunk) {
                    let target = self
                        .viewport
                        .scroll_offset
                        .saturating_add(self.page_step())
                        .min(self.max_focus_for_large_hunk(hunk))
                        .min(self.max_scroll_offset());
                    if self.viewport.scroll_offset != target {
                        self.viewport.scroll_offset = target;
                        self.viewport.active_hunk = Some(idx);
                        return true;
                    }
                }

                // Skip to next hunk
                let next_idx = idx + 1;
                if next_idx >= self.hunks.len() {
                    return false;
                }
                let next_hunk = self.hunks[next_idx];
                let target = self.scroll_for_hunk_start(next_hunk.start);
                self.viewport.scroll_offset = target;
                self.viewport.active_hunk = Some(next_idx);
                true
            }
            Some(current_idx) => {
                let current_hunk = self.hunks[current_idx];

                // If large hunk and we haven't scrolled to its end yet, scroll within it
                if self.should_advance_within_hunk(current_hunk) {
                    let target = self
                        .viewport
                        .scroll_offset
                        .saturating_add(self.page_step())
                        .min(self.max_focus_for_large_hunk(current_hunk))
                        .min(self.max_scroll_offset());
                    if self.viewport.scroll_offset == target {
                        // Already at max within this hunk; try next
                    } else {
                        self.viewport.scroll_offset = target;
                        return true;
                    }
                }

                // Move to next hunk
                let next_idx = current_idx + 1;
                if next_idx >= self.hunks.len() {
                    return false;
                }
                let next_hunk = self.hunks[next_idx];
                let target = self.scroll_for_hunk_start(next_hunk.start);
                self.viewport.scroll_offset = target;
                self.viewport.active_hunk = Some(next_idx);
                true
            }
        }
    }

    pub fn navigate_prev_hunk(&mut self) -> bool {
        if self.hunks.is_empty() {
            return false;
        }

        match self.viewport.active_hunk {
            None => {
                // Find the last hunk visible in or before the viewport window.
                // Use "start < scroll_offset + viewport_height" so that hunks visible
                // at the bottom of the viewport are reachable even when scroll is clamped.
                let visible_end = self
                    .viewport
                    .scroll_offset
                    .saturating_add(self.viewport.height);
                let Some(idx) = self.hunks.iter().rposition(|h| h.start < visible_end) else {
                    return false;
                };
                let hunk = self.hunks[idx];

                if self.viewport.scroll_offset >= hunk.end {
                    // scroll is at or past the end of this hunk; jump back to its start
                    let target = self.scroll_for_hunk_start(hunk.start);
                    self.viewport.scroll_offset = target;
                    self.viewport.active_hunk = Some(idx);
                    return true;
                }

                if self.viewport.scroll_offset < hunk.start {
                    // scroll is before this hunk; jump to it
                    let target = self.scroll_for_hunk_start(hunk.start);
                    self.viewport.scroll_offset = target;
                    self.viewport.active_hunk = Some(idx);
                    return true;
                }

                // scroll is within this hunk — check rewind-within-hunk or go to prev hunk
                if self.should_rewind_within_hunk(hunk) {
                    let target = self
                        .viewport
                        .scroll_offset
                        .saturating_sub(self.page_step())
                        .max(hunk.start);
                    if self.viewport.scroll_offset != target {
                        self.viewport.scroll_offset = target;
                        self.viewport.active_hunk = Some(idx);
                        return true;
                    }
                    // Already at hunk.start; fall through to prev hunk
                }

                // At hunk.start — go to previous hunk
                if idx == 0 {
                    return false;
                }
                let prev_idx = idx - 1;
                let prev_hunk = self.hunks[prev_idx];
                let target = self.scroll_for_hunk_start(prev_hunk.start);
                self.viewport.scroll_offset = target;
                self.viewport.active_hunk = Some(prev_idx);
                true
            }
            Some(current_idx) => {
                let current_hunk = self.hunks[current_idx];

                // If large hunk and scroll_offset > hunk.start, scroll back within it
                if self.should_rewind_within_hunk(current_hunk) {
                    let target = self
                        .viewport
                        .scroll_offset
                        .saturating_sub(self.page_step())
                        .max(current_hunk.start);
                    if self.viewport.scroll_offset == target {
                        // Already at start; fall through to prev hunk
                    } else {
                        self.viewport.scroll_offset = target;
                        return true;
                    }
                }

                // Move to previous hunk
                if current_idx == 0 {
                    return false;
                }
                let prev_idx = current_idx - 1;
                let prev_hunk = self.hunks[prev_idx];
                let target = self.scroll_for_hunk_start(prev_hunk.start);
                self.viewport.scroll_offset = target;
                self.viewport.active_hunk = Some(prev_idx);
                true
            }
        }
    }

    pub fn jump_to_first_hunk(&mut self) -> bool {
        if self.hunks.is_empty() {
            return false;
        }
        let hunk = self.hunks[0];
        self.viewport.scroll_offset = self.scroll_for_hunk_start(hunk.start);
        self.viewport.active_hunk = Some(0);
        true
    }

    pub fn jump_to_last_hunk(&mut self) -> bool {
        if self.hunks.is_empty() {
            return false;
        }
        let last_idx = self.hunks.len() - 1;
        let hunk = self.hunks[last_idx];
        let target = if self.viewport.height > 0 && hunk.len() > self.viewport.height {
            self.max_focus_for_large_hunk(hunk)
                .min(self.max_scroll_offset())
        } else {
            self.scroll_for_hunk_start(hunk.start)
        };
        self.viewport.scroll_offset = target;
        self.viewport.active_hunk = Some(last_idx);
        true
    }

    pub fn focused_change_idx(&self) -> Option<usize> {
        self.viewport.active_hunk
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

    pub fn viewport_height(&self) -> usize {
        self.viewport.height
    }

    pub fn viewport_width(&self) -> usize {
        self.viewport.width
    }

    pub fn scroll_offset(&self) -> usize {
        self.viewport.scroll_offset
    }

    // --- private helpers ---

    fn clamp_scroll_offset(&mut self) {
        let max = self.max_scroll_offset();
        if self.viewport.scroll_offset > max {
            self.viewport.scroll_offset = max;
        }
    }

    /// Compute the scroll offset that positions hunk_start near the top of the viewport
    /// with focus_offset rows of context above it (matching old renderer behavior).
    fn scroll_for_hunk_start(&self, hunk_start: usize) -> usize {
        let focus_offset = self.viewport.height / 4;
        hunk_start
            .saturating_sub(focus_offset)
            .min(self.max_scroll_offset())
    }

    fn should_advance_within_hunk(&self, hunk: HunkRange) -> bool {
        self.viewport.height > 0
            && hunk.len() > self.viewport.height
            && self.viewport.scroll_offset < self.max_focus_for_large_hunk(hunk)
    }

    fn should_rewind_within_hunk(&self, hunk: HunkRange) -> bool {
        self.viewport.height > 0
            && hunk.len() > self.viewport.height
            && self.viewport.scroll_offset > hunk.start
    }

    fn max_focus_for_large_hunk(&self, hunk: HunkRange) -> usize {
        hunk.end.saturating_sub(self.viewport.height / 2)
    }

    // Keep old navigate_next_diff / navigate_prev_diff as aliases for app/mod.rs compatibility
    pub fn navigate_next_diff(&mut self) -> bool {
        self.navigate_next_hunk()
    }

    pub fn navigate_prev_diff(&mut self) -> bool {
        self.navigate_prev_hunk()
    }

    pub fn jump_to_first_diff(&mut self) -> bool {
        self.jump_to_first_hunk()
    }

    pub fn jump_to_last_diff(&mut self) -> bool {
        self.jump_to_last_hunk()
    }
}

// Expose scroll_offset as a public field-like accessor for tests that set it directly
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
    fn move_down_clears_active_hunk() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=5).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(6, "a")],
                    new: vec![diff_line(6, "b")],
                },
            ],
            10,
        );
        view.navigate_next_hunk();
        assert!(view.viewport.active_hunk.is_some());
        view.move_down(1);
        assert_eq!(view.viewport.active_hunk, None);
    }

    #[test]
    fn move_up_clears_active_hunk() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=5).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(6, "a")],
                    new: vec![diff_line(6, "b")],
                },
            ],
            10,
        );
        view.navigate_next_hunk();
        assert!(view.viewport.active_hunk.is_some());
        view.move_up(1);
        assert_eq!(view.viewport.active_hunk, None);
    }

    #[test]
    fn navigate_next_hunk_scrolls_to_hunk_before_viewport() {
        // 20 unchanged + 1 change = 21 rows, viewport=10, max_scroll=11
        // scroll=0, hunk.start=20, hunk.end=21
        // focus=0 < hunk.end=21 → current hunk; focus(0) < hunk.start(20) → jump there
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=20).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(21, "a")],
                    new: vec![diff_line(21, "b")],
                },
            ],
            10,
        );
        assert!(view.navigate_next_hunk());
        assert_eq!(view.viewport.scroll_offset, 11); // min(20, max_scroll=11)
        assert_eq!(view.viewport.active_hunk, Some(0));
    }

    #[test]
    fn navigate_next_hunk_jumps_to_next_when_scroll_at_hunk_start() {
        // Hunk 0 at row 0, hunk 1 at row 15. scroll=0.
        // focus=0, hunk0.end=1 > 0 → current. focus(0) < hunk0.start(0)? No.
        // No advance-within (len=1 < 10). → skip to hunk 1.
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Change {
                    old: vec![diff_line(1, "a")],
                    new: vec![diff_line(1, "b")],
                },
                DiffBlock::Unchanged((2..=15).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(16, "c")],
                    new: vec![diff_line(16, "d")],
                },
            ],
            10,
        );
        // 1 + 14 + 1 = 16 rows, max_scroll = 16 - 10 = 6
        // Hunk 0 at row 0, hunk 1 at row 15
        // First navigate: scroll=0, hunk0.start=0, hunk0.end=1 > 0. focus<end. focus>=start.
        // Not large. → skip to hunk 1 at min(15,6)=6
        assert!(view.navigate_next_hunk());
        assert_eq!(view.viewport.active_hunk, Some(1));
        assert_eq!(view.viewport.scroll_offset, 6); // min(15, max_scroll=6)

        assert!(!view.navigate_next_hunk());
    }

    #[test]
    fn navigate_next_hunk_jumps_to_hunk_in_future() {
        // scroll=0, hunk at row 5. Should jump there.
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=5).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(6, "a")],
                    new: vec![diff_line(6, "b")],
                },
            ],
            10,
        );
        // 5 + 1 = 6 rows, max_scroll=0. hunk.start=5, hunk.end=6
        // focus=0 < end=6 → current. focus(0) < hunk.start(5) → jump to min(5,0)=0
        assert!(view.navigate_next_hunk());
        assert_eq!(view.viewport.scroll_offset, 0); // clamped by max_scroll=0
        assert_eq!(view.viewport.active_hunk, Some(0));
    }

    #[test]
    fn navigate_prev_hunk_works_backwards() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Change {
                    old: vec![diff_line(1, "a")],
                    new: vec![diff_line(1, "b")],
                },
                DiffBlock::Unchanged((2..=15).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(16, "c")],
                    new: vec![diff_line(16, "d")],
                },
            ],
            10,
        );
        // 1 + 14 + 1 = 16 rows, max_scroll = 6
        // Hunk 0 at row 0, hunk 1 at row 15
        view.jump_to_last_hunk();
        assert_eq!(view.viewport.active_hunk, Some(1));
        assert_eq!(view.viewport.scroll_offset, 6); // min(15, max_scroll=6)

        assert!(view.navigate_prev_hunk());
        assert_eq!(view.viewport.active_hunk, Some(0));
        assert_eq!(view.viewport.scroll_offset, 0);

        assert!(!view.navigate_prev_hunk());
    }

    #[test]
    fn large_hunk_scrolls_within_before_jumping_next() {
        // 10 rows of change starting at row 0, viewport=6
        // max_scroll = 10-6=4, page_step=3, max_focus=10-3=7 (clamped by max_scroll=4)
        // scroll=0, hunk.start=0, hunk.end=10. focus=0 < end=10. focus(0) >= start(0).
        // should_advance_within_hunk: len=10 > 6, scroll(0) < max_focus(7). target=0+3=3 min(7,4)=3
        let mut view = view_with_blocks(
            vec![DiffBlock::Change {
                old: (1..=10).map(|n| diff_line(n, "a")).collect(),
                new: (1..=10).map(|n| diff_line(n, "b")).collect(),
            }],
            6,
        );
        // First next: advance within large hunk (0+3=3)
        assert!(view.navigate_next_hunk());
        assert_eq!(view.viewport.scroll_offset, 3);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Second next: advance within (3+3=6, min(7,4)=4)
        assert!(view.navigate_next_hunk());
        assert_eq!(view.viewport.scroll_offset, 4);

        // Third next: at max_focus (4+3=7, min(7,4)=4 == current), no next hunk -> false
        assert!(!view.navigate_next_hunk());
    }

    #[test]
    fn large_hunk_scrolls_within_before_jumping_prev() {
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
        // Hunk 0 starts at row 4. max_focus = 14-3=11, but clamped to max_scroll=14-6=8
        // Set scroll to within the hunk
        view.viewport.scroll_offset = 8;
        view.viewport.active_hunk = Some(0);

        // Prev: rewind within hunk (page_step=3, target=max(4, 8-3)=5)
        assert!(view.navigate_prev_hunk());
        assert_eq!(view.viewport.scroll_offset, 5);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Prev: rewind to hunk.start=4
        assert!(view.navigate_prev_hunk());
        assert_eq!(view.viewport.scroll_offset, 4);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Prev: at hunk.start, no prev hunk
        assert!(!view.navigate_prev_hunk());
    }

    #[test]
    fn jump_to_first_hunk_scrolls_to_first_hunk() {
        // 30 unchanged + 2 changes so scrolling is meaningful
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=30).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(31, "a")],
                    new: vec![diff_line(31, "b")],
                },
                DiffBlock::Change {
                    old: vec![diff_line(32, "c")],
                    new: vec![diff_line(32, "d")],
                },
            ],
            10,
        );
        // 32 rows, max_scroll=22, hunk[0].start=30, target=min(30,22)=22
        view.jump_to_last_hunk();
        assert!(view.jump_to_first_hunk());
        assert_eq!(view.viewport.active_hunk, Some(0));
        assert_eq!(view.viewport.scroll_offset, 22);
    }

    #[test]
    fn jump_to_last_hunk_scrolls_to_last_hunk() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Change {
                    old: vec![diff_line(1, "a")],
                    new: vec![diff_line(1, "b")],
                },
                DiffBlock::Unchanged((2..=20).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(21, "c")],
                    new: vec![diff_line(21, "d")],
                },
            ],
            10,
        );
        // 1 + 19 + 1 = 21 rows, max_scroll=11, hunk[1].start=20
        assert!(view.jump_to_last_hunk());
        assert_eq!(view.viewport.active_hunk, Some(1));
        assert_eq!(view.viewport.scroll_offset, 11); // min(20, max_scroll=11)
    }

    #[test]
    fn jump_to_last_hunk_large_hunk_scrolls_to_near_end() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Change {
                old: (1..=20).map(|n| diff_line(n, "a")).collect(),
                new: (1..=20).map(|n| diff_line(n, "b")).collect(),
            }],
            13,
        );
        assert!(view.jump_to_last_hunk());
        // max_focus = 20 - 13/2 = 20 - 6 = 14, clamped to max_scroll = 20 - 13 = 7
        assert_eq!(view.viewport.scroll_offset, 7);
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

    #[test]
    fn is_row_in_active_hunk_returns_true_for_hunk_rows() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=3).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(4, "a")],
                    new: vec![diff_line(4, "b")],
                },
            ],
            10,
        );
        view.navigate_next_hunk();
        assert!(!view.is_row_in_active_hunk(0));
        assert!(!view.is_row_in_active_hunk(2));
        assert!(view.is_row_in_active_hunk(3));
        assert!(!view.is_row_in_active_hunk(4));
    }

    #[test]
    fn no_hunks_returns_false_for_navigate() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Unchanged(vec![unchanged_line(1, 1, "ctx")])],
            10,
        );
        assert!(!view.navigate_next_hunk());
        assert!(!view.navigate_prev_hunk());
        assert!(!view.jump_to_first_hunk());
        assert!(!view.jump_to_last_hunk());
    }
}
