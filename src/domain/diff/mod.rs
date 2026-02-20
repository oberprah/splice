#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DiffMeta {
    pub path: String,
    pub additions: u32,
    pub deletions: u32,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct FileDiff {
    pub meta: DiffMeta,
    pub blocks: Vec<DiffBlock>,
}

impl FileDiff {
    pub fn total_line_count(&self) -> usize {
        self.blocks
            .iter()
            .map(|block| match block {
                DiffBlock::Unchanged(unchanged) => unchanged.lines.len(),
                DiffBlock::Change(change) => change.old_lines.len().max(change.new_lines.len()),
            })
            .sum()
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum DiffBlock {
    Unchanged(UnchangedBlock),
    Change(ChangeBlock),
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnchangedBlock {
    pub old_start: u32,
    pub new_start: u32,
    pub lines: Vec<String>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct ChangeBlock {
    pub old_start: u32,
    pub new_start: u32,
    pub old_lines: Vec<String>,
    pub new_lines: Vec<String>,
}

pub fn build_file_diff(meta: DiffMeta, diff_output: &str) -> Result<FileDiff, String> {
    let blocks = parse_blocks(diff_output)?;
    Ok(FileDiff { meta, blocks })
}

fn parse_blocks(diff_output: &str) -> Result<Vec<DiffBlock>, String> {
    let mut blocks = Vec::new();
    let mut in_hunk = false;

    let mut old_line = 0u32;
    let mut new_line = 0u32;

    let mut current: Option<CurrentBlock> = None;

    for line in diff_output.lines() {
        if line.starts_with("@@ ") {
            if let Some(block) = current.take() {
                push_block(&mut blocks, block);
            }
            if let Some((old_start, new_start)) = parse_hunk_header(line) {
                in_hunk = true;
                old_line = old_start;
                new_line = new_start;
            } else {
                in_hunk = false;
            }
            continue;
        }

        if !in_hunk {
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
                let start_new_block = !matches!(current, Some(CurrentBlock::Unchanged { .. }));
                if start_new_block {
                    if let Some(block) = current.take() {
                        push_block(&mut blocks, block);
                    }
                    current = Some(CurrentBlock::Unchanged {
                        old_start: old_line,
                        new_start: new_line,
                        lines: Vec::new(),
                    });
                }
                if let Some(CurrentBlock::Unchanged { lines, .. }) = &mut current {
                    lines.push(content);
                }
                old_line = old_line.saturating_add(1);
                new_line = new_line.saturating_add(1);
            }
            '-' => {
                let start_new_block = !matches!(current, Some(CurrentBlock::Change { .. }));
                if start_new_block {
                    if let Some(block) = current.take() {
                        push_block(&mut blocks, block);
                    }
                    current = Some(CurrentBlock::Change {
                        old_start: old_line,
                        new_start: new_line,
                        old_lines: Vec::new(),
                        new_lines: Vec::new(),
                    });
                }
                if let Some(CurrentBlock::Change { old_lines, .. }) = &mut current {
                    old_lines.push(content);
                }
                old_line = old_line.saturating_add(1);
            }
            '+' => {
                let start_new_block = !matches!(current, Some(CurrentBlock::Change { .. }));
                if start_new_block {
                    if let Some(block) = current.take() {
                        push_block(&mut blocks, block);
                    }
                    current = Some(CurrentBlock::Change {
                        old_start: old_line,
                        new_start: new_line,
                        old_lines: Vec::new(),
                        new_lines: Vec::new(),
                    });
                }
                if let Some(CurrentBlock::Change { new_lines, .. }) = &mut current {
                    new_lines.push(content);
                }
                new_line = new_line.saturating_add(1);
            }
            _ => {}
        }
    }

    if let Some(block) = current.take() {
        push_block(&mut blocks, block);
    }

    Ok(blocks)
}

enum CurrentBlock {
    Unchanged {
        old_start: u32,
        new_start: u32,
        lines: Vec<String>,
    },
    Change {
        old_start: u32,
        new_start: u32,
        old_lines: Vec<String>,
        new_lines: Vec<String>,
    },
}

fn push_block(blocks: &mut Vec<DiffBlock>, block: CurrentBlock) {
    match block {
        CurrentBlock::Unchanged {
            old_start,
            new_start,
            lines,
        } => {
            if !lines.is_empty() {
                blocks.push(DiffBlock::Unchanged(UnchangedBlock {
                    old_start,
                    new_start,
                    lines,
                }));
            }
        }
        CurrentBlock::Change {
            old_start,
            new_start,
            old_lines,
            new_lines,
        } => {
            if !old_lines.is_empty() || !new_lines.is_empty() {
                blocks.push(DiffBlock::Change(ChangeBlock {
                    old_start,
                    new_start,
                    old_lines,
                    new_lines,
                }));
            }
        }
    }
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

    #[test]
    fn parses_unchanged_and_change_blocks() {
        let diff = "@@ -1,3 +1,3 @@\n line1\n-line2\n+line two\n line3\n";
        let meta = DiffMeta {
            path: "file.txt".to_string(),
            additions: 1,
            deletions: 1,
        };

        let file_diff = build_file_diff(meta, diff).expect("diff parse should succeed");

        assert_eq!(file_diff.blocks.len(), 3);
        assert_eq!(
            file_diff.blocks[0],
            DiffBlock::Unchanged(UnchangedBlock {
                old_start: 1,
                new_start: 1,
                lines: vec!["line1".to_string()],
            })
        );
        assert_eq!(
            file_diff.blocks[1],
            DiffBlock::Change(ChangeBlock {
                old_start: 2,
                new_start: 2,
                old_lines: vec!["line2".to_string()],
                new_lines: vec!["line two".to_string()],
            })
        );
        assert_eq!(
            file_diff.blocks[2],
            DiffBlock::Unchanged(UnchangedBlock {
                old_start: 3,
                new_start: 3,
                lines: vec!["line3".to_string()],
            })
        );
    }

    #[test]
    fn preserves_blank_lines_in_unchanged_blocks() {
        let diff = "@@ -1,3 +1,3 @@\n line1\n \n line3\n";
        let meta = DiffMeta {
            path: "file.txt".to_string(),
            additions: 0,
            deletions: 0,
        };

        let file_diff = build_file_diff(meta, diff).expect("diff parse should succeed");

        assert_eq!(file_diff.blocks.len(), 1);
        assert_eq!(
            file_diff.blocks[0],
            DiffBlock::Unchanged(UnchangedBlock {
                old_start: 1,
                new_start: 1,
                lines: vec!["line1".to_string(), "".to_string(), "line3".to_string()],
            })
        );
    }

    #[test]
    fn parses_add_only_and_remove_only_hunks() {
        let diff = "@@ -1,2 +1,1 @@\n-removed a\n-removed b\n@@ -5,0 +4,2 @@\n+added a\n+added b\n";
        let meta = DiffMeta {
            path: "file.txt".to_string(),
            additions: 2,
            deletions: 2,
        };

        let file_diff = build_file_diff(meta, diff).expect("diff parse should succeed");

        assert_eq!(file_diff.blocks.len(), 2);
        assert_eq!(
            file_diff.blocks[0],
            DiffBlock::Change(ChangeBlock {
                old_start: 1,
                new_start: 1,
                old_lines: vec!["removed a".to_string(), "removed b".to_string()],
                new_lines: Vec::new(),
            })
        );
        assert_eq!(
            file_diff.blocks[1],
            DiffBlock::Change(ChangeBlock {
                old_start: 5,
                new_start: 4,
                old_lines: Vec::new(),
                new_lines: vec!["added a".to_string(), "added b".to_string()],
            })
        );
    }
}
