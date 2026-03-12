use crate::core::Commit;

#[derive(Clone, Debug)]
pub struct CommitRange {
    pub start: Commit,
    pub end: Commit,
    pub count: usize,
    pub include_start: bool,
}

impl CommitRange {
    pub fn is_single_commit(&self) -> bool {
        self.count == 1
    }

    pub fn diff_base_spec(&self) -> String {
        if self.is_single_commit() {
            format!("{}^", self.end.hash)
        } else if self.include_start {
            format!("{}^", self.start.hash)
        } else {
            self.start.hash.clone()
        }
    }

    pub fn to_diff_spec(&self) -> String {
        format!("{}..{}", self.diff_base_spec(), self.end.hash)
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::{TimeZone, Utc};

    fn test_commit(hash: &str) -> Commit {
        Commit {
            hash: hash.to_string(),
            parent_hashes: vec![],
            refs: vec![],
            message: "test message".to_string(),
            body: None,
            author: "test author".to_string(),
            date: Utc.timestamp_opt(0, 0).unwrap(),
        }
    }

    #[test]
    fn is_single_commit_returns_true_for_count_1() {
        let commit = test_commit("abc123");
        let range = CommitRange {
            start: commit.clone(),
            end: commit,
            count: 1,
            include_start: true,
        };
        assert!(range.is_single_commit());
    }

    #[test]
    fn is_single_commit_returns_false_for_count_greater_than_1() {
        let range = CommitRange {
            start: test_commit("abc123"),
            end: test_commit("def456"),
            count: 3,
            include_start: true,
        };
        assert!(!range.is_single_commit());
    }

    #[test]
    fn to_diff_spec_for_single_commit() {
        let commit = test_commit("abc123def456");
        let range = CommitRange {
            start: commit.clone(),
            end: commit,
            count: 1,
            include_start: true,
        };
        assert_eq!(range.to_diff_spec(), "abc123def456^..abc123def456");
    }

    #[test]
    fn to_diff_spec_for_range() {
        let range = CommitRange {
            start: test_commit("ghi789"),
            end: test_commit("abc123"),
            count: 3,
            include_start: true,
        };
        assert_eq!(range.to_diff_spec(), "ghi789^..abc123");
    }
}
