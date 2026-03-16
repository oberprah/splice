use crate::domain::diff::layout::ScreenRow;
use crate::domain::diff::HunkRange;

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

pub fn max_scroll_offset(viewport: &Viewport) -> usize {
    viewport.total_rows.saturating_sub(viewport.height)
}

pub fn page_step(viewport: &Viewport) -> usize {
    (viewport.height / 2).max(1)
}

pub fn clamp_scroll_offset(viewport: &Viewport) -> usize {
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
                    } else if viewport.scroll_offset >= hunk.start
                        && viewport.scroll_offset < hunk.end
                    {
                        // Scroll is inside this hunk but it wasn't activated (user scrolled
                        // manually). Activate it and position to hunk start. Only skip to
                        // the next hunk if we're already at the target position.
                        let target = scroll_for_hunk_start(viewport, hunk.start);
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

#[cfg(test)]
mod tests {
    use super::*;
    use crate::app::diff_view::test_helpers::{diff_line, unchanged_line, view_with_blocks};
    use crate::domain::diff::DiffBlock;

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
    fn next_hunk_activates_current_hunk_when_scrolled_inside_it() {
        // Layout: 5 context lines, 1-line change hunk, 5 more context lines.
        // Total rows = 11, viewport height = 10, max_scroll = 1.
        // Hunk 0: start=5, end=6. scroll_for_hunk_start = min(5 - 10/4, max_scroll=1) = min(3,1) = 1.
        // Manually scroll into the hunk (scroll_offset=1, hunk.start=5? No — let's recalculate.)
        //
        // Actually with 5 unchanged + 1 change + 5 unchanged = 11 rows, max_scroll=1.
        // hunk.start=5, target = scroll_for_hunk_start = min(5-2, 1) = 1.
        // If we set scroll_offset=1 (inside range because scroll >= hunk.start is false — 1 < 5).
        // We need scroll_offset inside [hunk.start, hunk.end), i.e. inside [5,6).
        // But max_scroll=1 prevents reaching 5. Use a taller layout so we can scroll there.
        //
        // Layout: 10 context + 1 change + 10 context = 21 rows, viewport=10, max_scroll=11.
        // Hunk 0: start=10, end=11.
        // scroll_for_hunk_start = min(10 - 2, 11) = 8.
        // Set scroll_offset=10 (inside hunk: 10 >= 10 && 10 < 11), active_hunk=None.
        // Calling NextHunk: should activate hunk 0 and move scroll to 8, NOT jump to next hunk.
        let mut view = view_with_blocks(
            vec![
                DiffBlock::Unchanged((1..=10).map(|n| unchanged_line(n, n, "ctx")).collect()),
                DiffBlock::Change {
                    old: vec![diff_line(11, "a")],
                    new: vec![diff_line(11, "b")],
                },
                DiffBlock::Unchanged((12..=21).map(|n| unchanged_line(n, n, "ctx")).collect()),
            ],
            10,
        );
        // 10 + 1 + 10 = 21 rows, max_scroll=11. Hunk 0: start=10, end=11.
        // Simulate the user having manually scrolled into the hunk.
        view.viewport.scroll_offset = 10;
        view.viewport.active_hunk = None;

        assert!(view.update(ViewportAction::NextHunk));
        // Must activate hunk 0, not skip past it.
        assert_eq!(view.viewport.active_hunk, Some(0));
        // scroll_for_hunk_start(10) = min(10 - 10/4, 11) = min(10-2, 11) = 8
        assert_eq!(view.viewport.scroll_offset, 8);
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
