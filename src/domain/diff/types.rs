use crate::core::FileDiffInfo;
use crate::domain::highlight::TokenSpan;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DiffLine {
    pub number: u32,
    pub text: String,
    pub tokens: Vec<TokenSpan>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct UnchangedLine {
    pub old_number: u32,
    pub new_number: u32,
    pub text: String,
    pub tokens: Vec<TokenSpan>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum DiffBlock {
    Unchanged(Vec<UnchangedLine>),
    Change {
        old: Vec<DiffLine>,
        new: Vec<DiffLine>,
    },
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct FileDiff {
    pub info: FileDiffInfo,
    pub blocks: Vec<DiffBlock>,
}

impl FileDiff {
    pub fn total_line_count(&self) -> usize {
        self.blocks
            .iter()
            .map(|block| match block {
                DiffBlock::Unchanged(lines) => lines.len(),
                DiffBlock::Change { old, new } => old.len().max(new.len()),
            })
            .sum()
    }
}
