use std::collections::HashMap;

use super::{DiffBlock, DiffLine, FileDiff, UnchangedLine};
use crate::core::FileDiffInfo;
use crate::domain::highlight::HighlightedFile;

pub fn build_file_diff(
    info: FileDiffInfo,
    old_content: &str,
    new_content: &str,
    diff_output: &str,
    old_highlights: &HighlightedFile,
    new_highlights: &HighlightedFile,
) -> Result<FileDiff, String> {
    let parsed_diff = parse_unified_diff(diff_output);
    let blocks = build_blocks(
        old_content,
        new_content,
        &parsed_diff,
        old_highlights,
        new_highlights,
    );
    Ok(FileDiff { info, blocks })
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum LineType {
    Context,
    Add,
    Remove,
}

struct ParsedLine {
    line_type: LineType,
    #[allow(dead_code)]
    content: String,
    old_line_no: u32,
    new_line_no: u32,
}

struct ParsedDiff {
    lines: Vec<ParsedLine>,
}

fn parse_unified_diff(raw: &str) -> ParsedDiff {
    let mut lines = Vec::new();
    let mut old_line_no = 0u32;
    let mut new_line_no = 0u32;
    let mut in_hunk = false;

    for line in raw.lines() {
        if line.starts_with("@@ ") {
            if let Some((old_start, new_start)) = parse_hunk_header(line) {
                old_line_no = old_start;
                new_line_no = new_start;
                in_hunk = true;
            } else {
                in_hunk = false;
            }
            continue;
        }

        if !in_hunk {
            continue;
        }

        if line.is_empty() {
            continue;
        }

        if line.starts_with('\\') {
            continue;
        }

        let mut chars = line.chars();
        let prefix = chars.next().unwrap_or(' ');
        let content = chars.as_str().to_string();

        match prefix {
            ' ' => {
                lines.push(ParsedLine {
                    line_type: LineType::Context,
                    content,
                    old_line_no,
                    new_line_no,
                });
                old_line_no += 1;
                new_line_no += 1;
            }
            '-' => {
                lines.push(ParsedLine {
                    line_type: LineType::Remove,
                    content,
                    old_line_no,
                    new_line_no: 0,
                });
                old_line_no += 1;
            }
            '+' => {
                lines.push(ParsedLine {
                    line_type: LineType::Add,
                    content,
                    old_line_no: 0,
                    new_line_no,
                });
                new_line_no += 1;
            }
            _ => {
                in_hunk = false;
            }
        }
    }

    ParsedDiff { lines }
}

fn build_blocks(
    old_content: &str,
    new_content: &str,
    parsed_diff: &ParsedDiff,
    old_highlights: &HighlightedFile,
    new_highlights: &HighlightedFile,
) -> Vec<DiffBlock> {
    let old_lines: Vec<&str> = old_content.lines().collect();
    let new_lines: Vec<&str> = new_content.lines().collect();

    let mut old_diff_map = HashMap::new();
    let mut new_diff_map = HashMap::new();

    for line in &parsed_diff.lines {
        if line.old_line_no > 0 {
            old_diff_map.insert(line.old_line_no, line.line_type);
        }
        if line.new_line_no > 0 {
            new_diff_map.insert(line.new_line_no, line.line_type);
        }
    }

    let mut blocks = Vec::new();
    let mut current_unchanged: Option<Vec<UnchangedLine>> = None;
    let mut current_change: Option<(Vec<DiffLine>, Vec<DiffLine>)> = None;
    let mut hunk_removed: Vec<DiffLine> = Vec::new();
    let mut hunk_added: Vec<DiffLine> = Vec::new();

    let mut left_idx = 0usize;
    let mut right_idx = 0usize;

    let flush_unchanged = |blocks: &mut Vec<DiffBlock>,
                           current: &mut Option<Vec<UnchangedLine>>| {
        if let Some(lines) = current.take() {
            if !lines.is_empty() {
                blocks.push(DiffBlock::Unchanged(lines));
            }
        }
    };

    let flush_hunk =
        |hunk_removed: &mut Vec<DiffLine>,
         hunk_added: &mut Vec<DiffLine>,
         current_change: &mut Option<(Vec<DiffLine>, Vec<DiffLine>)>| {
            if hunk_removed.is_empty() && hunk_added.is_empty() {
                return;
            }

            if current_change.is_none() {
                *current_change = Some((Vec::new(), Vec::new()));
            }

            if let Some((ref mut old_lines, ref mut new_lines)) = current_change {
                old_lines.append(hunk_removed);
                new_lines.append(hunk_added);
            }
        };

    let flush_changed =
        |blocks: &mut Vec<DiffBlock>, current: &mut Option<(Vec<DiffLine>, Vec<DiffLine>)>| {
            if let Some((old, new)) = current.take() {
                if !old.is_empty() || !new.is_empty() {
                    blocks.push(DiffBlock::Change { old, new });
                }
            }
        };

    while left_idx < old_lines.len() || right_idx < new_lines.len() {
        let left_line_no = (left_idx + 1) as u32;
        let right_line_no = (right_idx + 1) as u32;

        let left_type = old_diff_map.get(&left_line_no).copied();
        let right_type = new_diff_map.get(&right_line_no).copied();

        let left_in_diff = left_type.is_some();
        let right_in_diff = right_type.is_some();

        let left_is_unchanged = !left_in_diff || left_type == Some(LineType::Context);
        let right_is_unchanged = !right_in_diff || right_type == Some(LineType::Context);

        if left_idx < old_lines.len()
            && right_idx < new_lines.len()
            && left_is_unchanged
            && right_is_unchanged
        {
            flush_hunk(&mut hunk_removed, &mut hunk_added, &mut current_change);
            flush_changed(&mut blocks, &mut current_change);

            if current_unchanged.is_none() {
                current_unchanged = Some(Vec::new());
            }
            if let Some(ref mut lines) = current_unchanged {
                lines.push(UnchangedLine {
                    old_number: left_line_no,
                    new_number: right_line_no,
                    text: old_lines[left_idx].to_string(),
                    tokens: new_highlights
                        .line_tokens(right_line_no)
                        .unwrap_or_default()
                        .to_vec(),
                });
            }
            left_idx += 1;
            right_idx += 1;
            continue;
        }

        flush_unchanged(&mut blocks, &mut current_unchanged);

        if left_idx < old_lines.len() && left_in_diff && left_type == Some(LineType::Remove) {
            hunk_removed.push(DiffLine {
                number: left_line_no,
                text: old_lines[left_idx].to_string(),
                tokens: old_highlights
                    .line_tokens(left_line_no)
                    .unwrap_or_default()
                    .to_vec(),
            });
            left_idx += 1;
            continue;
        }

        if right_idx < new_lines.len() && right_in_diff && right_type == Some(LineType::Add) {
            hunk_added.push(DiffLine {
                number: right_line_no,
                text: new_lines[right_idx].to_string(),
                tokens: new_highlights
                    .line_tokens(right_line_no)
                    .unwrap_or_default()
                    .to_vec(),
            });
            right_idx += 1;
            continue;
        }

        if left_idx >= old_lines.len() && right_idx < new_lines.len() {
            if right_in_diff && right_type == Some(LineType::Add) {
                hunk_added.push(DiffLine {
                    number: right_line_no,
                    text: new_lines[right_idx].to_string(),
                    tokens: new_highlights
                        .line_tokens(right_line_no)
                        .unwrap_or_default()
                        .to_vec(),
                });
            }
            right_idx += 1;
            continue;
        }

        if right_idx >= new_lines.len() && left_idx < old_lines.len() {
            if left_in_diff && left_type == Some(LineType::Remove) {
                hunk_removed.push(DiffLine {
                    number: left_line_no,
                    text: old_lines[left_idx].to_string(),
                    tokens: old_highlights
                        .line_tokens(left_line_no)
                        .unwrap_or_default()
                        .to_vec(),
                });
            }
            left_idx += 1;
            continue;
        }

        left_idx += 1;
        right_idx += 1;
    }

    flush_hunk(&mut hunk_removed, &mut hunk_added, &mut current_change);
    flush_changed(&mut blocks, &mut current_change);
    flush_unchanged(&mut blocks, &mut current_unchanged);

    blocks
}

fn parse_hunk_header(line: &str) -> Option<(u32, u32)> {
    if !line.starts_with("@@ ") {
        return None;
    }

    let end = line.find(" @@")?;
    let range = &line[3..end];
    let mut parts = range.split(' ');
    let old_part = parts.next()?;
    let new_part = parts.next()?;

    let old_start = parse_range(old_part.trim_start_matches('-'))?;
    let new_start = parse_range(new_part.trim_start_matches('+'))?;

    Some((old_start, new_start))
}

fn parse_range(range: &str) -> Option<u32> {
    let mut parts = range.split(',');
    let start = parts.next()?;
    start.parse().ok()
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::FileDiffInfo;

    fn no_highlights() -> HighlightedFile {
        HighlightedFile::default()
    }

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

    #[test]
    fn builds_full_file_diff() {
        let old_content = "line1\nline2\nline3\nline4\nline5\n";
        let new_content = "line1\nmodified\nline3\nline4\nline5\n";
        let diff_output = "--- a/file.txt\n+++ b/file.txt\n@@ -1,5 +1,5 @@\n line1\n-line2\n+modified\n line3\n line4\n line5\n";

        let info = FileDiffInfo {
            path: "file.txt".to_string(),
            old_path: None,
            status: crate::core::FileStatus::Modified,
            additions: 1,
            deletions: 1,
            is_binary: false,
        };

        let file_diff = build_file_diff(
            info,
            old_content,
            new_content,
            diff_output,
            &no_highlights(),
            &no_highlights(),
        )
        .expect("diff parse should succeed");

        assert_eq!(file_diff.blocks.len(), 3);

        assert_eq!(
            file_diff.blocks[0],
            DiffBlock::Unchanged(vec![unchanged_line(1, 1, "line1")])
        );

        assert_eq!(
            file_diff.blocks[1],
            DiffBlock::Change {
                old: vec![diff_line(2, "line2")],
                new: vec![diff_line(2, "modified")],
            }
        );

        assert_eq!(
            file_diff.blocks[2],
            DiffBlock::Unchanged(vec![
                unchanged_line(3, 3, "line3"),
                unchanged_line(4, 4, "line4"),
                unchanged_line(5, 5, "line5"),
            ])
        );
    }
}
