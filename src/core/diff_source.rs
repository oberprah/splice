use crate::core::CommitRange;

#[derive(Clone)]
pub enum DiffSource {
    CommitRange(CommitRange),
    Uncommitted(UncommittedType),
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum UncommittedType {
    Unstaged,
    Staged,
    All,
}

impl DiffSource {
    pub fn header_text(&self) -> String {
        match self {
            DiffSource::CommitRange(range) => {
                if range.is_single_commit() {
                    format!("{} {}", range.end.short_hash(), range.end.message)
                } else {
                    format!(
                        "{}..{} ({} commits)",
                        range.start.short_hash(),
                        range.end.short_hash(),
                        range.count
                    )
                }
            }
            DiffSource::Uncommitted(uncommitted_type) => match uncommitted_type {
                UncommittedType::Staged => "Staged changes".to_string(),
                UncommittedType::Unstaged => "Unstaged changes".to_string(),
                UncommittedType::All => "Uncommitted changes".to_string(),
            },
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::Commit;
    use chrono::{TimeZone, Utc};

    fn test_commit(hash: &str, message: &str) -> Commit {
        Commit {
            hash: hash.to_string(),
            parent_hashes: vec![],
            refs: vec![],
            message: message.to_string(),
            author: "test author".to_string(),
            date: Utc.timestamp_opt(0, 0).unwrap(),
        }
    }

    #[test]
    fn header_text_for_single_commit() {
        let commit = test_commit("abc123def456", "Initial commit");
        let range = CommitRange {
            start: commit.clone(),
            end: commit,
            count: 1,
        };
        let source = DiffSource::CommitRange(range);
        assert_eq!(source.header_text(), "abc123d Initial commit");
    }

    #[test]
    fn header_text_for_commit_range() {
        let range = CommitRange {
            start: test_commit("start123", "Start commit"),
            end: test_commit("end45678", "End commit"),
            count: 3,
        };
        let source = DiffSource::CommitRange(range);
        assert_eq!(source.header_text(), "start12..end4567 (3 commits)");
    }

    #[test]
    fn header_text_for_staged_changes() {
        let source = DiffSource::Uncommitted(UncommittedType::Staged);
        assert_eq!(source.header_text(), "Staged changes");
    }

    #[test]
    fn header_text_for_unstaged_changes() {
        let source = DiffSource::Uncommitted(UncommittedType::Unstaged);
        assert_eq!(source.header_text(), "Unstaged changes");
    }

    #[test]
    fn header_text_for_all_uncommitted_changes() {
        let source = DiffSource::Uncommitted(UncommittedType::All);
        assert_eq!(source.header_text(), "Uncommitted changes");
    }
}
