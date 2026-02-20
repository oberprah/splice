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
