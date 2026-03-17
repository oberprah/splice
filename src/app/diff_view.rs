use crate::core::DiffRef;
use crate::domain::diff::layout::{build_rows, ScreenRow};
use crate::domain::diff::FileDiff;

use super::viewport::{
    clamp_scroll_offset, max_scroll_offset, page_step, update_viewport, visible_content, Viewport,
    ViewportAction, VisibleContent,
};

pub struct DiffView {
    pub diff_ref: DiffRef,
    pub file: FileDiff,
    rows: Vec<ScreenRow>,
    hunks: Vec<crate::domain::diff::HunkRange>,
    pub viewport: Viewport,
    /// When set, the viewport is animating toward this scroll offset.
    /// Each call to `advance_animation` steps `viewport.scroll_offset` closer.
    animation_target: Option<usize>,
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
            animation_target: None,
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

        // Clamp or cancel animation target after resize — total_rows may have changed
        if let Some(target) = self.animation_target {
            let max = max_scroll_offset(&self.viewport);
            let clamped_target = target.min(max);
            if clamped_target == self.viewport.scroll_offset {
                self.animation_target = None;
            } else {
                self.animation_target = Some(clamped_target);
            }
        }
    }

    /// Apply a viewport action and store the new viewport state.
    /// Returns `true` if navigation happened (for hunk actions: whether it moved;
    /// for scroll actions: always `true`).
    ///
    /// Single-line moves (`ScrollDown(1)` / `ScrollUp(1)`) are applied immediately
    /// and cancel any running animation. All other actions set an animation target
    /// that `advance_animation` steps toward over multiple frames.
    pub fn update(&mut self, action: ViewportAction) -> bool {
        let is_single_step = matches!(
            action,
            ViewportAction::ScrollDown(1) | ViewportAction::ScrollUp(1)
        );

        if is_single_step {
            // j/k: cancel animation, apply immediately from current position
            self.animation_target = None;
            let (new_viewport, did_navigate) = update_viewport(&self.viewport, action, &self.hunks);
            self.viewport = new_viewport;
            did_navigate
        } else {
            // Compute from target state so rapid presses accumulate correctly
            let base_offset = self.animation_target.unwrap_or(self.viewport.scroll_offset);
            let virtual_viewport = Viewport {
                scroll_offset: base_offset,
                ..self.viewport
            };
            let (new_viewport, did_navigate) =
                update_viewport(&virtual_viewport, action, &self.hunks);

            // Apply non-scroll state immediately (hunk highlight appears right away)
            self.viewport.active_hunk = new_viewport.active_hunk;

            // Set animation target (or clear if already there)
            if new_viewport.scroll_offset != self.viewport.scroll_offset {
                self.animation_target = Some(new_viewport.scroll_offset);
            } else {
                self.animation_target = None;
            }

            did_navigate
        }
    }

    /// Step the scroll offset toward the animation target.
    /// Returns `true` if the offset changed (and a re-render is needed).
    pub fn advance_animation(&mut self) -> bool {
        let target = match self.animation_target {
            Some(t) => t,
            None => return false,
        };

        let current = self.viewport.scroll_offset;
        if current == target {
            self.animation_target = None;
            return false;
        }

        let distance = current.abs_diff(target);
        let step = (distance / 3).max(1);

        self.viewport.scroll_offset = if current < target {
            (current + step).min(target)
        } else {
            current.saturating_sub(step).max(target)
        };

        if self.viewport.scroll_offset == target {
            self.animation_target = None;
        }

        true
    }

    /// Returns `true` when a scroll animation is in progress.
    pub fn is_animating(&self) -> bool {
        self.animation_target.is_some()
    }

    /// Instantly complete any running animation (useful for tests).
    pub fn settle_animation(&mut self) {
        if let Some(target) = self.animation_target.take() {
            self.viewport.scroll_offset = target;
        }
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
pub(crate) mod test_helpers {
    use crate::core::{DiffRef, FileDiffInfo, FileStatus, UncommittedType};
    use crate::domain::diff::{DiffBlock, DiffLine, FileDiff, UnchangedLine};

    use super::DiffView;

    pub fn unchanged_line(old_number: u32, new_number: u32, text: &str) -> UnchangedLine {
        UnchangedLine {
            old_number,
            new_number,
            text: text.to_string(),
            tokens: vec![],
        }
    }

    pub fn diff_line(number: u32, text: &str) -> DiffLine {
        DiffLine {
            number,
            text: text.to_string(),
            tokens: vec![],
        }
    }

    pub fn view_with_blocks(blocks: Vec<DiffBlock>, viewport_height: usize) -> DiffView {
        let mut view = DiffView::new(
            DiffRef::Uncommitted(UncommittedType::All),
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
}

#[cfg(test)]
mod tests {
    use super::test_helpers::{diff_line, unchanged_line, view_with_blocks};
    use crate::app::viewport::ViewportAction;
    use crate::domain::diff::DiffBlock;

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
    fn animation_steps_toward_target_with_ease_out() {
        // 30 unchanged + 1 change = 31 rows, viewport=10, max_scroll=21
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=30).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(31, "a")],
                    new: vec![diff_line(31, "b")],
                },
            ],
            10,
        );

        // Navigate to hunk — sets animation target
        view.update(ViewportAction::NextHunk);
        assert!(view.is_animating());
        // active_hunk is set immediately
        assert!(view.viewport.active_hunk.is_some());
        // scroll_offset hasn't moved yet
        assert_eq!(view.viewport.scroll_offset, 0);

        // Advance animation — should take multiple frames with decreasing step
        let mut frames = 0;
        let mut offsets = vec![0usize];
        while view.advance_animation() {
            frames += 1;
            offsets.push(view.viewport.scroll_offset);
            assert!(frames < 100, "animation should converge");
        }

        // Must have taken multiple frames (ease-out)
        assert!(frames > 1, "expected multi-frame animation, got {frames}");
        // Must have arrived at the target
        assert!(!view.is_animating());
        // Steps should be monotonically increasing (scrolling down)
        for w in offsets.windows(2) {
            assert!(w[1] >= w[0], "offsets should be non-decreasing");
        }
        // Steps should decrease over time (ease-out characteristic)
        let steps: Vec<usize> = offsets.windows(2).map(|w| w[1] - w[0]).collect();
        let first_step = steps[0];
        let last_step = *steps.last().unwrap();
        assert!(
            first_step >= last_step,
            "first step ({first_step}) should be >= last step ({last_step})"
        );
    }

    #[test]
    fn single_line_scroll_cancels_animation() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=30).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(31, "a")],
                    new: vec![diff_line(31, "b")],
                },
            ],
            10,
        );

        // Start animation
        view.update(ViewportAction::NextHunk);
        assert!(view.is_animating());

        // Single-line scroll cancels animation and moves immediately
        view.update(ViewportAction::ScrollDown(1));
        assert!(!view.is_animating());
        assert_eq!(view.viewport.scroll_offset, 1);
    }

    #[test]
    fn rapid_hunk_navigation_retargets_correctly() {
        // Layout: hunk at row 10, hunk at row 25
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=10).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(11, "a")],
                    new: vec![diff_line(11, "b")],
                },
                DiffBlock::Unchanged((12..=25).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(26, "c")],
                    new: vec![diff_line(26, "d")],
                },
            ],
            10,
        );
        // 10 + 1 + 14 + 1 = 26 rows, max_scroll=16

        // First n: targets hunk 0
        view.update(ViewportAction::NextHunk);
        assert_eq!(view.viewport.active_hunk, Some(0));
        assert!(view.is_animating());

        // Second n without settling: retargets to hunk 1
        view.update(ViewportAction::NextHunk);
        assert_eq!(view.viewport.active_hunk, Some(1));

        // Settle to final position
        view.settle_animation();
        assert!(!view.is_animating());
    }

    #[test]
    fn settle_animation_jumps_to_target() {
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=30).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(31, "a")],
                    new: vec![diff_line(31, "b")],
                },
            ],
            10,
        );

        view.update(ViewportAction::NextHunk);
        assert!(view.is_animating());

        view.settle_animation();
        assert!(!view.is_animating());
        // Should be at the computed target position
        assert!(view.viewport.scroll_offset > 0);
    }

    #[test]
    fn resize_during_animation_clamps_target() {
        // 30 unchanged + 1 change = 31 rows, viewport=10, max_scroll=21
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=30).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(31, "a")],
                    new: vec![diff_line(31, "b")],
                },
            ],
            10,
        );

        // Start animation toward the hunk (target is near max_scroll=21)
        view.update(ViewportAction::NextHunk);
        assert!(view.is_animating());

        // Advance a few frames so scroll_offset moves partway
        view.advance_animation();
        view.advance_animation();
        let mid_offset = view.viewport.scroll_offset;
        assert!(mid_offset > 0);

        // Simulate resize to a much taller viewport — reduces max_scroll
        // With height=28, max_scroll = 31-28 = 3, which is less than mid_offset
        view.set_viewport_dimensions(28, 80);

        // Animation target should be clamped to new max_scroll
        let max = view.viewport.total_rows.saturating_sub(28);
        if view.is_animating() {
            view.settle_animation();
            assert!(view.viewport.scroll_offset <= max);
        } else {
            // Animation was cancelled because clamped target == clamped offset
            assert!(view.viewport.scroll_offset <= max);
        }
    }
}
