use crate::core::DiffRef;
use crate::domain::diff::layout::{build_rows, HunkRange, ScreenRow};
use crate::domain::diff::FileDiff;

#[derive(Clone, Copy)]
pub struct Viewport {
    pub scroll_offset: usize,
    pub height: usize,
    pub width: usize,
    pub active_hunk: Option<usize>,
    pub total_rows: usize,
}

pub enum ViewportAction {
    ScrollDown(usize),
    ScrollUp(usize),
    NextHunk,
    PrevHunk,
    JumpToFirstHunk,
    JumpToLastHunk,
}

pub struct VisibleContent<'a> {
    pub rows: &'a [ScreenRow],
    pub row_offset: usize,
    pub active_hunk_range: Option<HunkRange>,
}

pub fn visible_content<'a>(
    rows: &'a [ScreenRow],
    hunks: &[HunkRange],
    viewport: &Viewport,
) -> VisibleContent<'a> {
    let start = viewport.scroll_offset;
    let end = (start + viewport.height).min(rows.len());
    let visible_rows = &rows[start..end];

    let active_hunk_range = viewport.active_hunk.and_then(|idx| hunks.get(idx).copied());

    VisibleContent {
        rows: visible_rows,
        row_offset: viewport.scroll_offset,
        active_hunk_range,
    }
}

// --- pure viewport helpers (free functions) ---

fn max_scroll_offset(viewport: &Viewport) -> usize {
    viewport.total_rows.saturating_sub(viewport.height)
}

fn page_step(viewport: &Viewport) -> usize {
    (viewport.height / 2).max(1)
}

fn clamp_scroll_offset(viewport: &Viewport) -> usize {
    viewport.scroll_offset.min(max_scroll_offset(viewport))
}

fn scroll_for_hunk_start(viewport: &Viewport, hunk_start: usize) -> usize {
    let focus_offset = viewport.height / 4;
    hunk_start
        .saturating_sub(focus_offset)
        .min(max_scroll_offset(viewport))
}

fn max_focus_for_large_hunk(viewport: &Viewport, hunk: HunkRange) -> usize {
    hunk.end.saturating_sub(viewport.height / 2)
}

fn should_advance_within_hunk(viewport: &Viewport, hunk: HunkRange) -> bool {
    viewport.height > 0
        && hunk.len() > viewport.height
        && viewport.scroll_offset < max_focus_for_large_hunk(viewport, hunk)
}

fn should_rewind_within_hunk(viewport: &Viewport, hunk: HunkRange) -> bool {
    viewport.height > 0 && hunk.len() > viewport.height && viewport.scroll_offset > hunk.start
}

/// Pure function: compute the next Viewport state given an action.
/// Returns `(new_viewport, did_navigate)` where `did_navigate` is always `true`
/// for scroll actions and signals whether hunk navigation actually moved for hunk actions.
pub fn update_viewport(
    viewport: &Viewport,
    action: ViewportAction,
    hunks: &[HunkRange],
) -> (Viewport, bool) {
    match action {
        ViewportAction::ScrollDown(amount) => {
            let max_scroll = max_scroll_offset(viewport);
            let new_offset = viewport
                .scroll_offset
                .saturating_add(amount)
                .min(max_scroll);
            (
                Viewport {
                    scroll_offset: new_offset,
                    active_hunk: None,
                    ..*viewport
                },
                true,
            )
        }
        ViewportAction::ScrollUp(amount) => {
            let new_offset = viewport.scroll_offset.saturating_sub(amount);
            (
                Viewport {
                    scroll_offset: new_offset,
                    active_hunk: None,
                    ..*viewport
                },
                true,
            )
        }
        ViewportAction::NextHunk => {
            if hunks.is_empty() {
                return (*viewport, false);
            }

            let (new_offset, new_active) = match viewport.active_hunk {
                None => {
                    let Some(idx) = hunks.iter().position(|h| h.end > viewport.scroll_offset)
                    else {
                        return (*viewport, false);
                    };
                    let hunk = hunks[idx];

                    if viewport.scroll_offset < hunk.start {
                        let target = scroll_for_hunk_start(viewport, hunk.start);
                        return (
                            Viewport {
                                scroll_offset: target,
                                active_hunk: Some(idx),
                                ..*viewport
                            },
                            true,
                        );
                    }

                    if should_advance_within_hunk(viewport, hunk) {
                        let target = viewport
                            .scroll_offset
                            .saturating_add(page_step(viewport))
                            .min(max_focus_for_large_hunk(viewport, hunk))
                            .min(max_scroll_offset(viewport));
                        if viewport.scroll_offset != target {
                            return (
                                Viewport {
                                    scroll_offset: target,
                                    active_hunk: Some(idx),
                                    ..*viewport
                                },
                                true,
                            );
                        }
                    }

                    let next_idx = idx + 1;
                    if next_idx >= hunks.len() {
                        return (*viewport, false);
                    }
                    let next_hunk = hunks[next_idx];
                    let target = scroll_for_hunk_start(viewport, next_hunk.start);
                    (target, Some(next_idx))
                }
                Some(current_idx) => {
                    let current_hunk = hunks[current_idx];

                    if should_advance_within_hunk(viewport, current_hunk) {
                        let target = viewport
                            .scroll_offset
                            .saturating_add(page_step(viewport))
                            .min(max_focus_for_large_hunk(viewport, current_hunk))
                            .min(max_scroll_offset(viewport));
                        if viewport.scroll_offset != target {
                            return (
                                Viewport {
                                    scroll_offset: target,
                                    active_hunk: Some(current_idx),
                                    ..*viewport
                                },
                                true,
                            );
                        }
                    }

                    let next_idx = current_idx + 1;
                    if next_idx >= hunks.len() {
                        return (*viewport, false);
                    }
                    let next_hunk = hunks[next_idx];
                    let target = scroll_for_hunk_start(viewport, next_hunk.start);
                    (target, Some(next_idx))
                }
            };

            (
                Viewport {
                    scroll_offset: new_offset,
                    active_hunk: new_active,
                    ..*viewport
                },
                true,
            )
        }
        ViewportAction::PrevHunk => {
            if hunks.is_empty() {
                return (*viewport, false);
            }

            let (new_offset, new_active) = match viewport.active_hunk {
                None => {
                    let visible_end = viewport.scroll_offset.saturating_add(viewport.height);
                    let Some(idx) = hunks.iter().rposition(|h| h.start < visible_end) else {
                        return (*viewport, false);
                    };
                    let hunk = hunks[idx];

                    if viewport.scroll_offset >= hunk.end {
                        let target = scroll_for_hunk_start(viewport, hunk.start);
                        return (
                            Viewport {
                                scroll_offset: target,
                                active_hunk: Some(idx),
                                ..*viewport
                            },
                            true,
                        );
                    }

                    if viewport.scroll_offset < hunk.start {
                        let target = scroll_for_hunk_start(viewport, hunk.start);
                        return (
                            Viewport {
                                scroll_offset: target,
                                active_hunk: Some(idx),
                                ..*viewport
                            },
                            true,
                        );
                    }

                    if should_rewind_within_hunk(viewport, hunk) {
                        let target = viewport
                            .scroll_offset
                            .saturating_sub(page_step(viewport))
                            .max(hunk.start);
                        if viewport.scroll_offset != target {
                            return (
                                Viewport {
                                    scroll_offset: target,
                                    active_hunk: Some(idx),
                                    ..*viewport
                                },
                                true,
                            );
                        }
                    }

                    if idx == 0 {
                        return (*viewport, false);
                    }
                    let prev_idx = idx - 1;
                    let prev_hunk = hunks[prev_idx];
                    let target = scroll_for_hunk_start(viewport, prev_hunk.start);
                    (target, Some(prev_idx))
                }
                Some(current_idx) => {
                    let current_hunk = hunks[current_idx];

                    if should_rewind_within_hunk(viewport, current_hunk) {
                        let target = viewport
                            .scroll_offset
                            .saturating_sub(page_step(viewport))
                            .max(current_hunk.start);
                        if viewport.scroll_offset != target {
                            return (
                                Viewport {
                                    scroll_offset: target,
                                    active_hunk: Some(current_idx),
                                    ..*viewport
                                },
                                true,
                            );
                        }
                    }

                    if current_idx == 0 {
                        return (*viewport, false);
                    }
                    let prev_idx = current_idx - 1;
                    let prev_hunk = hunks[prev_idx];
                    let target = scroll_for_hunk_start(viewport, prev_hunk.start);
                    (target, Some(prev_idx))
                }
            };

            (
                Viewport {
                    scroll_offset: new_offset,
                    active_hunk: new_active,
                    ..*viewport
                },
                true,
            )
        }
        ViewportAction::JumpToFirstHunk => {
            if hunks.is_empty() {
                return (*viewport, false);
            }
            let hunk = hunks[0];
            let target = scroll_for_hunk_start(viewport, hunk.start);
            (
                Viewport {
                    scroll_offset: target,
                    active_hunk: Some(0),
                    ..*viewport
                },
                true,
            )
        }
        ViewportAction::JumpToLastHunk => {
            if hunks.is_empty() {
                return (*viewport, false);
            }
            let last_idx = hunks.len() - 1;
            let hunk = hunks[last_idx];
            let target = if viewport.height > 0 && hunk.len() > viewport.height {
                max_focus_for_large_hunk(viewport, hunk).min(max_scroll_offset(viewport))
            } else {
                scroll_for_hunk_start(viewport, hunk.start)
            };
            (
                Viewport {
                    scroll_offset: target,
                    active_hunk: Some(last_idx),
                    ..*viewport
                },
                true,
            )
        }
    }
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
        view.update(ViewportAction::NextHunk);
        assert!(view.viewport.active_hunk.is_some());
        view.update(ViewportAction::ScrollDown(1));
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
        view.update(ViewportAction::NextHunk);
        assert!(view.viewport.active_hunk.is_some());
        view.update(ViewportAction::ScrollUp(1));
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
        assert!(view.update(ViewportAction::NextHunk));
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
        assert!(view.update(ViewportAction::NextHunk));
        assert_eq!(view.viewport.active_hunk, Some(1));
        assert_eq!(view.viewport.scroll_offset, 6); // min(15, max_scroll=6)

        assert!(!view.update(ViewportAction::NextHunk));
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
        assert!(view.update(ViewportAction::NextHunk));
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
        view.update(ViewportAction::JumpToLastHunk);
        assert_eq!(view.viewport.active_hunk, Some(1));
        assert_eq!(view.viewport.scroll_offset, 6); // min(15, max_scroll=6)

        assert!(view.update(ViewportAction::PrevHunk));
        assert_eq!(view.viewport.active_hunk, Some(0));
        assert_eq!(view.viewport.scroll_offset, 0);

        assert!(!view.update(ViewportAction::PrevHunk));
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
        assert!(view.update(ViewportAction::NextHunk));
        assert_eq!(view.viewport.scroll_offset, 3);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Second next: advance within (3+3=6, min(7,4)=4)
        assert!(view.update(ViewportAction::NextHunk));
        assert_eq!(view.viewport.scroll_offset, 4);

        // Third next: at max_focus (4+3=7, min(7,4)=4 == current), no next hunk -> false
        assert!(!view.update(ViewportAction::NextHunk));
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
        assert!(view.update(ViewportAction::PrevHunk));
        assert_eq!(view.viewport.scroll_offset, 5);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Prev: rewind to hunk.start=4
        assert!(view.update(ViewportAction::PrevHunk));
        assert_eq!(view.viewport.scroll_offset, 4);
        assert_eq!(view.viewport.active_hunk, Some(0));

        // Prev: at hunk.start, no prev hunk
        assert!(!view.update(ViewportAction::PrevHunk));
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
        view.update(ViewportAction::JumpToLastHunk);
        assert!(view.update(ViewportAction::JumpToFirstHunk));
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
        assert!(view.update(ViewportAction::JumpToLastHunk));
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
        assert!(view.update(ViewportAction::JumpToLastHunk));
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
        view.update(ViewportAction::NextHunk);
        let content = view.visible_content();
        let hunk_range = content.active_hunk_range;
        let in_hunk =
            |abs_row: usize| hunk_range.is_some_and(|r| abs_row >= r.start && abs_row < r.end);
        assert!(!in_hunk(0));
        assert!(!in_hunk(2));
        assert!(in_hunk(3));
        assert!(!in_hunk(4));
    }

    #[test]
    fn no_hunks_returns_false_for_navigate() {
        let mut view = view_with_blocks(
            vec![DiffBlock::Unchanged(vec![unchanged_line(1, 1, "ctx")])],
            10,
        );
        assert!(!view.update(ViewportAction::NextHunk));
        assert!(!view.update(ViewportAction::PrevHunk));
        assert!(!view.update(ViewportAction::JumpToFirstHunk));
        assert!(!view.update(ViewportAction::JumpToLastHunk));
    }
}
