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

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct HunkRange {
    pub start: usize,
    pub end: usize,
}

impl HunkRange {
    pub fn len(&self) -> usize {
        self.end.saturating_sub(self.start)
    }

    pub fn is_empty(&self) -> bool {
        self.len() == 0
    }
}
