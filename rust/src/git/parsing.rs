use chrono::{DateTime, Utc};
use crate::core::{Commit, RefInfo, RefType};

pub fn parse_log_output(output: &str) -> Result<Vec<Commit>, String> {
    let output = output.trim();
    if output.is_empty() {
        return Ok(Vec::new());
    }

    let mut commits = Vec::new();

    for record in output.split('\x1e') {
        let record = record.trim();
        if record.is_empty() {
            continue;
        }

        let fields: Vec<&str> = record.splitn(6, '\0').collect();
        if fields.len() != 6 {
            continue;
        }

        let hash = fields[0].to_string();
        if hash.is_empty() {
            continue;
        }

        let parent_hashes: Vec<String> = if fields[1].is_empty() {
            Vec::new()
        } else {
            fields[1].split(' ').map(|s| s.to_string()).collect()
        };

        let refs = parse_ref_decorations(fields[2]);
        let author = fields[3].to_string();
        
        let date = match DateTime::parse_from_rfc3339(fields[4]) {
            Ok(dt) => dt.with_timezone(&Utc),
            Err(_) => continue,
        };

        let message = fields[5].to_string();

        commits.push(Commit {
            hash,
            parent_hashes,
            refs,
            message,
            author,
            date,
        });
    }

    Ok(commits)
}

fn parse_ref_decorations(refs_str: &str) -> Vec<RefInfo> {
    let refs_str = refs_str.trim();
    if refs_str.is_empty() {
        return Vec::new();
    }

    if !refs_str.starts_with('(') || !refs_str.ends_with(')') {
        return Vec::new();
    }

    let inner = &refs_str[1..refs_str.len() - 1];
    let mut refs = Vec::new();
    let mut head_branch: Option<&str> = None;
    let mut parts: Vec<&str> = Vec::new();

    for part in inner.split(',') {
        let part = part.trim();
        if part.is_empty() {
            continue;
        }

        if let Some(branch_name) = part.strip_prefix("HEAD -> ") {
            head_branch = Some(branch_name);
        }

        parts.push(part);
    }

    for part in parts {
        if let Some(branch_name) = part.strip_prefix("HEAD -> ") {
            refs.push(RefInfo::new(branch_name.to_string(), RefType::Branch, true));
        } else if part == "HEAD" {
            refs.push(RefInfo::new("HEAD".to_string(), RefType::DetachedHead, true));
        } else if let Some(tag_name) = part.strip_prefix("tag: ") {
            refs.push(RefInfo::tag(tag_name.to_string()));
        } else if part.contains('/') {
            refs.push(RefInfo::remote_branch(part.to_string()));
        } else {
            let is_head = head_branch == Some(part);
            refs.push(RefInfo::new(part.to_string(), RefType::Branch, is_head));
        }
    }

    refs
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_empty_output() {
        let result = parse_log_output("").unwrap();
        assert!(result.is_empty());
    }

    #[test]
    fn test_parse_whitespace_only_output() {
        let result = parse_log_output("   \n\n  ").unwrap();
        assert!(result.is_empty());
    }

    #[test]
    fn test_parse_single_commit() {
        let input = "abc123\0\0\0Alice\02024-01-01T10:00:00Z\0Initial commit";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].hash, "abc123");
        assert_eq!(commits[0].message, "Initial commit");
        assert_eq!(commits[0].author, "Alice");
        assert!(commits[0].refs.is_empty());
        assert!(commits[0].parent_hashes.is_empty());
    }

    #[test]
    fn test_parse_commit_with_single_parent() {
        let input = "child123\0parent456\0\0Bob\02024-01-02T10:00:00Z\0Second commit";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].hash, "child123");
        assert_eq!(commits[0].parent_hashes, vec!["parent456"]);
    }

    #[test]
    fn test_parse_merge_commit_with_multiple_parents() {
        let input = "merge123\0parent1 parent2\0\0Carol\02024-01-03T10:00:00Z\0Merge commit";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].parent_hashes, vec!["parent1", "parent2"]);
    }

    #[test]
    fn test_parse_multiple_commits() {
        let input = "hash1\0\0\0Alice\02024-01-01T10:00:00Z\0First\x1ehash2\0\0\0Bob\02024-01-02T10:00:00Z\0Second";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 2);
        assert_eq!(commits[0].hash, "hash1");
        assert_eq!(commits[0].message, "First");
        assert_eq!(commits[1].hash, "hash2");
        assert_eq!(commits[1].message, "Second");
    }

    #[test]
    fn test_parse_skips_malformed_records() {
        let input = "valid\0\0\0Alice\02024-01-01T10:00:00Z\0Valid\x1einvalid\x1ealso\0\0\0Bob\02024-01-02T10:00:00Z\0Also valid";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 2);
        assert_eq!(commits[0].hash, "valid");
        assert_eq!(commits[1].hash, "also");
    }

    #[test]
    fn test_parse_skips_empty_hash() {
        let input = "\0\0\0Alice\02024-01-01T10:00:00Z\0No hash\x1ehash2\0\0\0Bob\02024-01-02T10:00:00Z\0Valid";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].hash, "hash2");
    }

    #[test]
    fn test_parse_skips_invalid_date() {
        let input = "hash1\0\0\0Alice\0invalid-date\0Message1\x1ehash2\0\0\0Bob\02024-01-01T10:00:00Z\0Message2";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].hash, "hash2");
    }

    #[test]
    fn test_parse_commit_with_refs() {
        let input = "abc123\0\0 (HEAD -> main, tag: v1.0)\0Alice\02024-01-01T10:00:00Z\0Message";
        let commits = parse_log_output(input).unwrap();
        
        assert_eq!(commits.len(), 1);
        assert_eq!(commits[0].refs.len(), 2);
        assert_eq!(commits[0].refs[0].name, "main");
        assert!(commits[0].refs[0].is_head);
        assert_eq!(commits[0].refs[1].name, "v1.0");
        assert_eq!(commits[0].refs[1].ref_type, RefType::Tag);
    }

    #[test]
    fn test_parse_ref_decorations_empty() {
        assert!(parse_ref_decorations("").is_empty());
        assert!(parse_ref_decorations("  ").is_empty());
    }

    #[test]
    fn test_parse_ref_decorations_unwrapped() {
        assert!(parse_ref_decorations("main").is_empty());
        assert!(parse_ref_decorations("(unclosed").is_empty());
        assert!(parse_ref_decorations("unopened)").is_empty());
    }

    #[test]
    fn test_parse_ref_decorations_head_branch() {
        let refs = parse_ref_decorations("(HEAD -> main)");
        assert_eq!(refs.len(), 1);
        assert_eq!(refs[0].name, "main");
        assert!(refs[0].is_head);
        assert_eq!(refs[0].ref_type, RefType::Branch);
    }

    #[test]
    fn test_parse_ref_decorations_branch_only() {
        let refs = parse_ref_decorations("(main)");
        assert_eq!(refs.len(), 1);
        assert_eq!(refs[0].name, "main");
        assert!(!refs[0].is_head);
        assert_eq!(refs[0].ref_type, RefType::Branch);
    }

    #[test]
    fn test_parse_ref_decorations_remote_branch() {
        let refs = parse_ref_decorations("(origin/main)");
        assert_eq!(refs.len(), 1);
        assert_eq!(refs[0].name, "origin/main");
        assert_eq!(refs[0].ref_type, RefType::RemoteBranch);
        assert!(!refs[0].is_head);
    }

    #[test]
    fn test_parse_ref_decorations_tag() {
        let refs = parse_ref_decorations("(tag: v1.0.0)");
        assert_eq!(refs.len(), 1);
        assert_eq!(refs[0].name, "v1.0.0");
        assert_eq!(refs[0].ref_type, RefType::Tag);
        assert!(!refs[0].is_head);
    }

    #[test]
    fn test_parse_ref_decorations_multiple_tags() {
        let refs = parse_ref_decorations("(tag: v1.0.0, tag: latest)");
        assert_eq!(refs.len(), 2);
        assert_eq!(refs[0].name, "v1.0.0");
        assert_eq!(refs[0].ref_type, RefType::Tag);
        assert_eq!(refs[1].name, "latest");
        assert_eq!(refs[1].ref_type, RefType::Tag);
    }

    #[test]
    fn test_parse_ref_decorations_mixed_refs() {
        let refs = parse_ref_decorations("(HEAD -> main, origin/main, feature, tag: v1.0)");
        assert_eq!(refs.len(), 4);
        
        assert_eq!(refs[0].name, "main");
        assert!(refs[0].is_head);
        
        assert_eq!(refs[1].name, "origin/main");
        assert_eq!(refs[1].ref_type, RefType::RemoteBranch);
        
        assert_eq!(refs[2].name, "feature");
        assert_eq!(refs[2].ref_type, RefType::Branch);
        assert!(!refs[2].is_head);
        
        assert_eq!(refs[3].name, "v1.0");
        assert_eq!(refs[3].ref_type, RefType::Tag);
    }

    #[test]
    fn test_parse_ref_decorations_empty_parts() {
        let refs = parse_ref_decorations("(main,  , feature)");
        assert_eq!(refs.len(), 2);
        assert_eq!(refs[0].name, "main");
        assert_eq!(refs[1].name, "feature");
    }

    #[test]
    fn test_parse_ref_decorations_detached_head() {
        let refs = parse_ref_decorations("(HEAD)");
        assert_eq!(refs.len(), 1);
        assert_eq!(refs[0].name, "HEAD");
        assert_eq!(refs[0].ref_type, RefType::DetachedHead);
        assert!(refs[0].is_head);
    }
}
