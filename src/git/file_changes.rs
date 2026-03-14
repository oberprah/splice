use crate::core::{FileDiffInfo, FileStatus};
use std::collections::HashMap;

pub fn parse_file_changes(numstat: &str, name_status: &str) -> Result<Vec<FileDiffInfo>, String> {
    let entries = parse_name_status_entries(name_status);
    let rename_new_paths: std::collections::HashSet<&str> = entries
        .iter()
        .filter(|entry| entry.status == FileStatus::Renamed)
        .map(|entry| entry.path.as_str())
        .collect();

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
        let raw_path = parts[2].to_string();

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

        stats.insert(
            raw_path.clone(),
            (Some(additions), Some(deletions), is_binary),
        );

        if let Some((_old, new)) = parse_rename_path(&raw_path) {
            if rename_new_paths.contains(new.as_str()) {
                stats.insert(new, (Some(additions), Some(deletions), is_binary));
            }
        }
    }

    let mut changes = Vec::new();

    for entry in entries {
        let (additions, deletions, is_binary) = stats
            .get(&entry.path)
            .map(|(a, d, b)| (a.unwrap_or(0), d.unwrap_or(0), *b))
            .unwrap_or((0, 0, false));

        changes.push(FileDiffInfo {
            path: entry.path,
            old_path: entry.old_path,
            status: entry.status,
            additions,
            deletions,
            is_binary,
        });
    }

    Ok(changes)
}

struct NameStatusEntry {
    status: FileStatus,
    path: String,
    old_path: Option<String>,
}

fn parse_name_status_entries(name_status: &str) -> Vec<NameStatusEntry> {
    let mut entries = Vec::new();

    for line in name_status.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        let parts: Vec<&str> = line.splitn(3, '\t').collect();
        if parts.len() < 2 {
            continue;
        }

        let status_char = parts[0].trim();
        if status_char.starts_with('R') {
            if parts.len() < 3 {
                continue;
            }

            entries.push(NameStatusEntry {
                status: FileStatus::Renamed,
                path: parts[2].to_string(),
                old_path: Some(parts[1].to_string()),
            });
            continue;
        }

        let status = match status_char {
            "M" => FileStatus::Modified,
            "A" => FileStatus::Added,
            "D" => FileStatus::Deleted,
            _ => continue,
        };

        entries.push(NameStatusEntry {
            status,
            path: parts[1].to_string(),
            old_path: None,
        });
    }

    entries
}

fn parse_rename_path(path: &str) -> Option<(String, String)> {
    if let Some(open) = path.find('{') {
        if let Some(close_rel) = path[open..].find('}') {
            let close = open + close_rel;
            let prefix = &path[..open];
            let middle = &path[open + 1..close];
            let suffix = &path[close + 1..];

            if let Some((old_middle, new_middle)) = middle.split_once(" => ") {
                return Some((
                    format!("{}{}{}", prefix, old_middle, suffix),
                    format!("{}{}{}", prefix, new_middle, suffix),
                ));
            }
        }
    }

    if let Some((old, new)) = path.split_once(" => ") {
        return Some((old.to_string(), new.to_string()));
    }

    None
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

    #[test]
    fn test_parse_file_changes_renamed() {
        let numstat = "0\t0\told_name.txt => new_name.txt";
        let name_status = "R100\told_name.txt\tnew_name.txt";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].status, FileStatus::Renamed);
        assert_eq!(changes[0].path, "new_name.txt");
        assert_eq!(changes[0].old_path, Some("old_name.txt".to_string()));
        assert_eq!(changes[0].additions, 0);
        assert_eq!(changes[0].deletions, 0);
    }

    #[test]
    fn test_parse_file_changes_renamed_with_brace_path_and_edits() {
        let numstat = "5\t2\t{src => ui}/components/Button.tsx";
        let name_status = "R087\tsrc/components/Button.tsx\tui/components/Button.tsx";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].status, FileStatus::Renamed);
        assert_eq!(changes[0].path, "ui/components/Button.tsx");
        assert_eq!(
            changes[0].old_path,
            Some("src/components/Button.tsx".to_string())
        );
        assert_eq!(changes[0].additions, 5);
        assert_eq!(changes[0].deletions, 2);
    }

    #[test]
    fn test_parse_file_changes_modified_file_name_with_arrow_keeps_stats() {
        let numstat = "7\t3\tdocs/a => b.md";
        let name_status = "M\tdocs/a => b.md";
        let changes = parse_file_changes(numstat, name_status).unwrap();

        assert_eq!(changes.len(), 1);
        assert_eq!(changes[0].status, FileStatus::Modified);
        assert_eq!(changes[0].path, "docs/a => b.md");
        assert_eq!(changes[0].additions, 7);
        assert_eq!(changes[0].deletions, 3);
    }
}
