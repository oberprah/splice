use crate::domain::diff::inline_diff::{
    compute_inline_spans, map_emphasis_for_segment, InlineSpan,
};
use crate::domain::diff::types::HunkRange;
use crate::domain::diff::{DiffBlock, FileDiff};
use crate::domain::highlight::TokenSpan;
use crate::domain::wrap::{wrap_line, WrappedSegment};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CellKind {
    Context,
    /// Pure addition: new side only (no old counterpart at this pair index)
    Added,
    /// Pure deletion: old side only (no new counterpart at this pair index)
    Removed,
    /// Modified line: both old and new exist at this pair index
    Changed,
    Empty,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Cell {
    pub kind: CellKind,
    pub line_number: Option<u32>,
    pub text: String,
    pub tokens: Vec<TokenSpan>,
    pub emphasis: Vec<InlineSpan>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ScreenRow {
    pub left: Cell,
    pub right: Cell,
}

fn wrap_or_empty(text: &str, tokens: &[TokenSpan], width: usize) -> Vec<WrappedSegment> {
    if width == 0 {
        vec![WrappedSegment {
            text: String::new(),
            char_offset: 0,
            tokens: Vec::new(),
        }]
    } else {
        wrap_line(text, tokens, width)
    }
}

fn empty_cell() -> Cell {
    Cell {
        kind: CellKind::Empty,
        line_number: None,
        text: String::new(),
        tokens: Vec::new(),
        emphasis: Vec::new(),
    }
}

pub fn build_rows(file: &FileDiff, width: usize) -> (Vec<ScreenRow>, Vec<HunkRange>) {
    if width == 0 {
        return (Vec::new(), Vec::new());
    }

    let separator_width = 3; // " │ "
    let available = width.saturating_sub(separator_width);
    let left_width = available / 2;
    let right_width = available.saturating_sub(left_width);

    // Find max line number to determine prefix width
    let max_line_num = file
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

    // "{:>3} " (4 chars) — line number + trailing space
    let prefix_width = format!("{:>3} ", max_line_num).chars().count();

    // Left and right use their respective cell widths for wrapping (matching the
    // original renderer which wrapped each side at its own visual width).
    let left_content_width = left_width.saturating_sub(prefix_width);
    let right_content_width = right_width.saturating_sub(prefix_width);

    let mut rows: Vec<ScreenRow> = Vec::new();
    let mut hunks: Vec<HunkRange> = Vec::new();

    for block in &file.blocks {
        match block {
            DiffBlock::Unchanged(lines) => {
                for line in lines {
                    let left_segs = wrap_or_empty(&line.text, &line.tokens, left_content_width);
                    let right_segs = wrap_or_empty(&line.text, &line.tokens, right_content_width);

                    let max_segs = left_segs.len().max(right_segs.len());
                    for seg_idx in 0..max_segs {
                        let left_cell = if let Some(seg) = left_segs.get(seg_idx) {
                            Cell {
                                kind: CellKind::Context,
                                line_number: if seg_idx == 0 {
                                    Some(line.old_number)
                                } else {
                                    None
                                },
                                text: seg.text.clone(),
                                tokens: seg.tokens.clone(),
                                emphasis: Vec::new(),
                            }
                        } else {
                            empty_cell()
                        };
                        let right_cell = if let Some(seg) = right_segs.get(seg_idx) {
                            Cell {
                                kind: CellKind::Context,
                                line_number: if seg_idx == 0 {
                                    Some(line.new_number)
                                } else {
                                    None
                                },
                                text: seg.text.clone(),
                                tokens: seg.tokens.clone(),
                                emphasis: Vec::new(),
                            }
                        } else {
                            empty_cell()
                        };
                        rows.push(ScreenRow {
                            left: left_cell,
                            right: right_cell,
                        });
                    }
                }
            }
            DiffBlock::Change { old, new } => {
                let hunk_start = rows.len();

                // A block is a "modification" when both sides have content.
                // In that case every cell (including unpaired extras) uses
                // CellKind::Changed (blue) to keep the block visually unified.
                let is_modification = !old.is_empty() && !new.is_empty();

                // Compute emphasis across the whole block: join all old/new
                // line texts so that multi-line reformats get proper inline
                // highlighting instead of only the first positional pair.
                let (block_old_emphasis, block_new_emphasis) = if is_modification {
                    let old_joined: String = old
                        .iter()
                        .map(|l| l.text.as_str())
                        .collect::<Vec<_>>()
                        .join("\n");
                    let new_joined: String = new
                        .iter()
                        .map(|l| l.text.as_str())
                        .collect::<Vec<_>>()
                        .join("\n");
                    compute_inline_spans(&old_joined, &new_joined)
                } else {
                    (Vec::new(), Vec::new())
                };

                // Build a mapping from block-joined char offsets to per-line
                // offsets so we can slice emphasis spans for each line.
                // Each line starts at cumulative offset (sum of preceding
                // line lengths + 1 for the '\n' separator).
                let old_line_offsets: Vec<usize> = {
                    let mut offsets = Vec::with_capacity(old.len());
                    let mut pos = 0;
                    for line in old.iter() {
                        offsets.push(pos);
                        pos += line.text.chars().count() + 1; // +1 for '\n'
                    }
                    offsets
                };
                let new_line_offsets: Vec<usize> = {
                    let mut offsets = Vec::with_capacity(new.len());
                    let mut pos = 0;
                    for line in new.iter() {
                        offsets.push(pos);
                        pos += line.text.chars().count() + 1;
                    }
                    offsets
                };

                let pair_count = old.len().max(new.len());

                for i in 0..pair_count {
                    let old_line = old.get(i);
                    let new_line = new.get(i);

                    // Slice block-level emphasis to this line's range
                    let old_emphasis = if let Some(line) = old_line {
                        let line_start = old_line_offsets[i];
                        let line_len = line.text.chars().count();
                        map_emphasis_for_segment(&block_old_emphasis, line_start, line_len)
                    } else {
                        Vec::new()
                    };
                    let new_emphasis = if let Some(line) = new_line {
                        let line_start = new_line_offsets[i];
                        let line_len = line.text.chars().count();
                        map_emphasis_for_segment(&block_new_emphasis, line_start, line_len)
                    } else {
                        Vec::new()
                    };

                    let left_segs: Vec<WrappedSegment> = if let Some(line) = old_line {
                        wrap_or_empty(&line.text, &line.tokens, left_content_width)
                    } else {
                        Vec::new()
                    };

                    let right_segs: Vec<WrappedSegment> = if let Some(line) = new_line {
                        wrap_or_empty(&line.text, &line.tokens, right_content_width)
                    } else {
                        Vec::new()
                    };

                    let max_rows = left_segs.len().max(right_segs.len()).max(1);

                    // Cell kinds: if it's a modification block, everything
                    // is Changed (blue). Otherwise pure add or remove.
                    let left_kind = if is_modification {
                        CellKind::Changed
                    } else {
                        CellKind::Removed
                    };
                    let right_kind = if is_modification {
                        CellKind::Changed
                    } else {
                        CellKind::Added
                    };

                    for row_idx in 0..max_rows {
                        let left_cell = if old_line.is_some() {
                            if let Some(seg) = left_segs.get(row_idx) {
                                Cell {
                                    kind: left_kind,
                                    line_number: if row_idx == 0 {
                                        old_line.map(|l| l.number)
                                    } else {
                                        None
                                    },
                                    text: seg.text.clone(),
                                    tokens: seg.tokens.clone(),
                                    emphasis: map_emphasis_for_segment(
                                        &old_emphasis,
                                        seg.char_offset,
                                        seg.text.chars().count(),
                                    ),
                                }
                            } else if is_modification {
                                Cell {
                                    kind: CellKind::Changed,
                                    line_number: None,
                                    text: String::new(),
                                    tokens: Vec::new(),
                                    emphasis: Vec::new(),
                                }
                            } else {
                                empty_cell()
                            }
                        } else {
                            Cell {
                                kind: if is_modification {
                                    CellKind::Changed
                                } else {
                                    CellKind::Empty
                                },
                                line_number: None,
                                text: String::new(),
                                tokens: Vec::new(),
                                emphasis: Vec::new(),
                            }
                        };

                        let right_cell = if new_line.is_some() {
                            if let Some(seg) = right_segs.get(row_idx) {
                                Cell {
                                    kind: right_kind,
                                    line_number: if row_idx == 0 {
                                        new_line.map(|l| l.number)
                                    } else {
                                        None
                                    },
                                    text: seg.text.clone(),
                                    tokens: seg.tokens.clone(),
                                    emphasis: map_emphasis_for_segment(
                                        &new_emphasis,
                                        seg.char_offset,
                                        seg.text.chars().count(),
                                    ),
                                }
                            } else if is_modification {
                                Cell {
                                    kind: CellKind::Changed,
                                    line_number: None,
                                    text: String::new(),
                                    tokens: Vec::new(),
                                    emphasis: Vec::new(),
                                }
                            } else {
                                empty_cell()
                            }
                        } else {
                            Cell {
                                kind: if is_modification {
                                    CellKind::Changed
                                } else {
                                    CellKind::Empty
                                },
                                line_number: None,
                                text: String::new(),
                                tokens: Vec::new(),
                                emphasis: Vec::new(),
                            }
                        };

                        rows.push(ScreenRow {
                            left: left_cell,
                            right: right_cell,
                        });
                    }
                }

                hunks.push(HunkRange {
                    start: hunk_start,
                    end: rows.len(),
                });
            }
        }
    }

    (rows, hunks)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::{FileDiffInfo, FileStatus};
    use crate::domain::diff::{DiffLine, UnchangedLine};

    fn make_file(blocks: Vec<DiffBlock>) -> FileDiff {
        FileDiff {
            info: FileDiffInfo {
                path: "test.rs".to_string(),
                old_path: None,
                status: FileStatus::Modified,
                additions: 0,
                deletions: 0,
                is_binary: false,
            },
            blocks,
        }
    }

    fn unchanged(old: u32, new: u32, text: &str) -> UnchangedLine {
        UnchangedLine {
            old_number: old,
            new_number: new,
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

    #[test]
    fn simple_unchanged_block_produces_context_rows() {
        let file = make_file(vec![DiffBlock::Unchanged(vec![
            unchanged(1, 1, "hello"),
            unchanged(2, 2, "world"),
        ])]);
        let (rows, hunks) = build_rows(&file, 80);
        assert_eq!(rows.len(), 2);
        assert!(hunks.is_empty());
        assert_eq!(rows[0].left.kind, CellKind::Context);
        assert_eq!(rows[0].right.kind, CellKind::Context);
        assert_eq!(rows[0].left.line_number, Some(1));
        assert_eq!(rows[0].right.line_number, Some(1));
        assert_eq!(rows[1].left.line_number, Some(2));
        assert_eq!(rows[1].right.line_number, Some(2));
    }

    #[test]
    fn change_block_produces_correct_kinds_with_padding() {
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "old line")],
            new: vec![diff_line(1, "new line"), diff_line(2, "extra")],
        }]);
        let (rows, hunks) = build_rows(&file, 80);
        // max(1, 2) = 2 rows
        assert_eq!(rows.len(), 2);
        assert_eq!(hunks.len(), 1);
        assert_eq!(hunks[0].start, 0);
        assert_eq!(hunks[0].end, 2);

        // First row: both old[0] and new[0] exist -> Changed
        assert_eq!(rows[0].left.kind, CellKind::Changed);
        assert_eq!(rows[0].right.kind, CellKind::Changed);
        assert_eq!(rows[0].left.line_number, Some(1));
        assert_eq!(rows[0].right.line_number, Some(1));

        // Second row: old side has no more lines but block is a modification,
        // so both sides stay Changed (blue) to keep the block visually unified
        assert_eq!(rows[1].left.kind, CellKind::Changed);
        assert_eq!(rows[1].right.kind, CellKind::Changed);
        assert_eq!(rows[1].left.line_number, None);
        assert_eq!(rows[1].right.line_number, Some(2));
    }

    #[test]
    fn pure_addition_produces_empty_left() {
        let file = make_file(vec![DiffBlock::Change {
            old: vec![],
            new: vec![diff_line(5, "added")],
        }]);
        let (rows, hunks) = build_rows(&file, 80);
        assert_eq!(rows.len(), 1);
        assert_eq!(rows[0].left.kind, CellKind::Empty);
        assert_eq!(rows[0].right.kind, CellKind::Added);
        assert_eq!(hunks[0], HunkRange { start: 0, end: 1 });
    }

    #[test]
    fn pure_deletion_produces_empty_right() {
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(5, "removed")],
            new: vec![],
        }]);
        let (rows, hunks) = build_rows(&file, 80);
        assert_eq!(rows.len(), 1);
        assert_eq!(rows[0].left.kind, CellKind::Removed);
        assert_eq!(rows[0].right.kind, CellKind::Empty);
        assert_eq!(hunks[0], HunkRange { start: 0, end: 1 });
    }

    #[test]
    fn wrapped_lines_produce_continuation_rows_with_no_line_number() {
        // With width=40, prefix=5, left_content_width = (40-3)/2 - 5 = 18-5 = ~13ish
        // Use a very narrow width so the long text wraps
        let long_text = "hello world foo bar baz qux";
        let file = make_file(vec![DiffBlock::Unchanged(vec![unchanged(1, 1, long_text)])]);
        let (rows, _) = build_rows(&file, 30);
        // Should have more than 1 row due to wrapping
        assert!(rows.len() > 1, "expected wrapping to produce multiple rows");
        // First row has line number
        assert_eq!(rows[0].left.line_number, Some(1));
        // Continuation rows have None
        for row in &rows[1..] {
            assert_eq!(row.left.line_number, None);
        }
    }

    #[test]
    fn hunk_ranges_correctly_track_start_end_indices() {
        let file = make_file(vec![
            DiffBlock::Unchanged(vec![unchanged(1, 1, "ctx"), unchanged(2, 2, "ctx")]),
            DiffBlock::Change {
                old: vec![diff_line(3, "a")],
                new: vec![diff_line(3, "b")],
            },
            DiffBlock::Unchanged(vec![unchanged(4, 4, "ctx")]),
            DiffBlock::Change {
                old: vec![diff_line(5, "c"), diff_line(6, "d")],
                new: vec![diff_line(5, "e")],
            },
        ]);
        let (rows, hunks) = build_rows(&file, 80);
        assert_eq!(hunks.len(), 2);
        // First 2 rows are unchanged, then 1 row change
        assert_eq!(hunks[0].start, 2);
        assert_eq!(hunks[0].end, 3);
        // 1 more unchanged, then 2-row change (max(2,1)=2)
        assert_eq!(hunks[1].start, 4);
        assert_eq!(hunks[1].end, 6);
        assert_eq!(rows.len(), 6);
    }

    #[test]
    fn empty_diff_produces_empty_rows() {
        let file = make_file(vec![]);
        let (rows, hunks) = build_rows(&file, 80);
        assert!(rows.is_empty());
        assert!(hunks.is_empty());
    }

    #[test]
    fn zero_width_produces_empty_rows() {
        let file = make_file(vec![DiffBlock::Unchanged(vec![unchanged(1, 1, "hello")])]);
        let (rows, hunks) = build_rows(&file, 0);
        assert!(rows.is_empty());
        assert!(hunks.is_empty());
    }

    #[test]
    fn changed_cells_have_emphasis_spans() {
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "hello world")],
            new: vec![diff_line(1, "hello earth")],
        }]);
        let (rows, _) = build_rows(&file, 80);
        assert_eq!(rows[0].left.kind, CellKind::Changed);
        assert_eq!(rows[0].right.kind, CellKind::Changed);
        // "world" vs "earth" — emphasis should be non-empty on both sides
        assert!(!rows[0].left.emphasis.is_empty());
        assert!(!rows[0].right.emphasis.is_empty());
    }

    #[test]
    fn pure_add_remove_cells_have_no_emphasis() {
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "removed")],
            new: vec![],
        }]);
        let (rows, _) = build_rows(&file, 80);
        assert!(rows[0].left.emphasis.is_empty());
    }

    #[test]
    fn context_cells_have_no_emphasis() {
        let file = make_file(vec![DiffBlock::Unchanged(vec![unchanged(1, 1, "context")])]);
        let (rows, _) = build_rows(&file, 80);
        assert!(rows[0].left.emphasis.is_empty());
        assert!(rows[0].right.emphasis.is_empty());
    }

    #[test]
    fn asymmetric_modification_uses_changed_kind_for_all_rows() {
        // 1 old line, 3 new lines — all should be Changed (blue), not Added
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "foo(a, b, c)")],
            new: vec![
                diff_line(1, "foo("),
                diff_line(2, "    a, b, c,"),
                diff_line(3, ")"),
            ],
        }]);
        let (rows, _) = build_rows(&file, 80);
        assert_eq!(rows.len(), 3);
        for (i, row) in rows.iter().enumerate() {
            assert_eq!(
                row.left.kind,
                CellKind::Changed,
                "row {i} left should be Changed"
            );
            assert_eq!(
                row.right.kind,
                CellKind::Changed,
                "row {i} right should be Changed"
            );
        }
        // First row has line numbers on both sides
        assert_eq!(rows[0].left.line_number, Some(1));
        assert_eq!(rows[0].right.line_number, Some(1));
        // Extra rows on new side have line numbers, old side is a spacer
        assert_eq!(rows[1].left.line_number, None);
        assert_eq!(rows[1].right.line_number, Some(2));
        assert!(
            rows[1].left.text.is_empty(),
            "spacer should have empty text"
        );
    }

    #[test]
    fn wrapped_modification_overflow_rows_use_changed_kind() {
        // When a modification pair wraps to different row counts, the
        // shorter side's overflow rows should be Changed (blue), not Empty.
        // Use a narrow width so one side wraps more than the other.
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "short")],
            new: vec![diff_line(
                1,
                "this text is long enough to wrap at narrow width",
            )],
        }]);
        // Width 30: available=27, left=13, right=14, prefix=4,
        // left_content=9, right_content=10
        // "short" fits in 1 row on the left. The right text wraps to multiple rows.
        let (rows, _) = build_rows(&file, 30);
        assert!(
            rows.len() > 1,
            "expected wrapping to produce multiple rows, got {}",
            rows.len()
        );
        // All rows should be Changed on both sides — not Empty on the left
        for (i, row) in rows.iter().enumerate() {
            assert_eq!(
                row.left.kind,
                CellKind::Changed,
                "row {i} left should be Changed, not {:?}",
                row.left.kind
            );
            assert_eq!(
                row.right.kind,
                CellKind::Changed,
                "row {i} right should be Changed, not {:?}",
                row.right.kind
            );
        }
    }

    #[test]
    fn asymmetric_modification_has_block_level_emphasis() {
        // Emphasis should be computed across all lines, not just the first pair
        let file = make_file(vec![DiffBlock::Change {
            old: vec![diff_line(1, "hello world")],
            new: vec![diff_line(1, "hello"), diff_line(2, "earth")],
        }]);
        let (rows, _) = build_rows(&file, 80);
        // "earth" on the new side should get emphasis (it replaced "world")
        assert!(
            !rows[1].right.emphasis.is_empty(),
            "second new line should have emphasis from block-level diff"
        );
    }
}
