use crate::core::{FileChange, FileStatus};
use std::collections::HashMap;

pub fn parse_file_changes(numstat: &str, name_status: &str) -> Result<Vec<FileChange>, String> {
    let mut stats: HashMap<String, (Option<u32>, Option<u32>, bool)> = HashMap::new();

    for line in numstat.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        let parts: Vec<&str> = line.splitn(3, '\t').collect();
        if parts.len() != 3 {
            continue;
        }

        let additions_str = parts[0];
        let deletions_str = parts[1];
        let path = parts[2].to_string();

        let is_binary = additions_str == "-" && deletions_str == "-";
        let additions = if is_binary {
            0
        } else {
            additions_str.parse().unwrap_or(0)
        };
        let deletions = if is_binary {
            0
        } else {
            deletions_str.parse().unwrap_or(0)
        };

        stats.insert(path, (Some(additions), Some(deletions), is_binary));
    }

    let mut changes = Vec::new();

    for line in name_status.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        let parts: Vec<&str> = line.splitn(2, '\t').collect();
        if parts.len() != 2 {
            continue;
        }

        let status_char = parts[0].trim();
        let path = parts[1].to_string();

        let status = match status_char {
            "M" => FileStatus::Modified,
            "A" => FileStatus::Added,
            "D" => FileStatus::Deleted,
            "R" => FileStatus::Renamed,
            _ => continue,
        };

        let (additions, deletions, is_binary) = stats
            .get(&path)
            .map(|(a, d, b)| (a.unwrap_or(0), d.unwrap_or(0), *b))
            .unwrap_or((0, 0, false));

        changes.push(FileChange {
            path,
            status,
            additions,
            deletions,
            is_binary,
        });
    }

    Ok(changes)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_file_changes_empty() {
        let result = parse_file_changes("", "").unwrap();
        assert!(result.is_empty());
    }

    #[test]
    fn test_parse_file_changes_modified() {
        let numstat = "10\t5\tsrc/main.rs";
        let name_status = "M\tsrc/main.rs";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].path, "src/main.rs");
        assert_eq!(changes[0].status, FileStatus::Modified);
        assert_eq!(changes[0].additions, 10);
        assert_eq!(changes[0].deletions, 5);
        assert!(!changes[0].is_binary);
    }

    #[test]
    fn test_parse_file_changes_added() {
        let numstat = "20\t0\tnew_file.rs";
        let name_status = "A\tnew_file.rs";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].status, FileStatus::Added);
        assert_eq!(changes[0].additions, 20);
        assert_eq!(changes[0].deletions, 0);
    }

    #[test]
    fn test_parse_file_changes_deleted() {
        let numstat = "0\t15\tdeleted_file.rs";
        let name_status = "D\tdeleted_file.rs";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].status, FileStatus::Deleted);
        assert_eq!(changes[0].additions, 0);
        assert_eq!(changes[0].deletions, 15);
    }

    #[test]
    fn test_parse_file_changes_binary() {
        let numstat = "-\t-\timage.png";
        let name_status = "M\timage.png";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert!(changes[0].is_binary);
        assert_eq!(changes[0].additions, 0);
        assert_eq!(changes[0].deletions, 0);
    }

    #[test]
    fn test_parse_file_changes_multiple() {
        let numstat = "10\t5\tsrc/a.rs\n20\t0\tsrc/b.rs\n0\t3\tsrc/c.rs";
        let name_status = "M\tsrc/a.rs\nA\tsrc/b.rs\nD\tsrc/c.rs";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 3);
        assert_eq!(changes[0].status, FileStatus::Modified);
        assert_eq!(changes[1].status, FileStatus::Added);
        assert_eq!(changes[2].status, FileStatus::Deleted);
    }
}
