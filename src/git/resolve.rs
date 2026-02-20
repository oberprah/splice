use std::path::Path;

use crate::core::{Commit, CommitRange, DiffSource, UncommittedType};

use super::git_command;

pub fn resolve_commit_range(repo_path: &Path, spec: &str) -> Result<CommitRange, String> {
    let is_single_commit = !spec.contains("..");
    let include_start = !spec.contains("...");

    let (start_ref, end_ref) = parse_range_spec(spec)?;

    let start_ref = if spec.contains("...") {
        let parts: Vec<&str> = spec.splitn(2, "...").collect();
        if parts.len() == 2 {
            find_merge_base(repo_path, parts[0], parts[1])?
        } else {
            start_ref
        }
    } else {
        start_ref
    };

    let start_commit = resolve_ref(repo_path, &start_ref)
        .map_err(|e| format!("error resolving start ref {:?}: {}", start_ref, e))?;
    let end_commit = resolve_ref(repo_path, &end_ref)
        .map_err(|e| format!("error resolving end ref {:?}: {}", end_ref, e))?;

    let count = if is_single_commit {
        1
    } else {
        count_commits_in_range(repo_path, &start_ref, &end_ref)?
    };

    let start_commit = if is_single_commit {
        end_commit.clone()
    } else {
        start_commit
    };

    Ok(CommitRange {
        start: start_commit,
        end: end_commit,
        count,
        include_start,
    })
}

fn parse_range_spec(spec: &str) -> Result<(String, String), String> {
    if spec.contains("...") {
        let parts: Vec<&str> = spec.splitn(2, "...").collect();
        if parts.len() != 2 {
            return Err(format!("invalid range spec: {}", spec));
        }
        Ok((parts[0].to_string(), parts[1].to_string()))
    } else if spec.contains("..") {
        let parts: Vec<&str> = spec.splitn(2, "..").collect();
        if parts.len() != 2 {
            return Err(format!("invalid range spec: {}", spec));
        }
        Ok((parts[0].to_string(), parts[1].to_string()))
    } else {
        Ok((spec.to_string(), "HEAD".to_string()))
    }
}

fn find_merge_base(repo_path: &Path, ref1: &str, ref2: &str) -> Result<String, String> {
    let output = git_command(repo_path)
        .args(["merge-base", ref1.trim(), ref2.trim()])
        .output()
        .map_err(|e| format!("Failed to run git merge-base: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git merge-base failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    Ok(String::from_utf8(output.stdout)
        .map_err(|e| format!("Invalid UTF-8 output: {}", e))?
        .trim()
        .to_string())
}

fn count_commits_in_range(
    repo_path: &Path,
    start_ref: &str,
    end_ref: &str,
) -> Result<usize, String> {
    let range_spec = format!("{}..{}", start_ref.trim(), end_ref.trim());

    let output = git_command(repo_path)
        .args(["rev-list", "--count", &range_spec])
        .output()
        .map_err(|e| format!("Failed to run git rev-list: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git rev-list failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    let stdout =
        String::from_utf8(output.stdout).map_err(|e| format!("Invalid UTF-8 output: {}", e))?;

    stdout
        .trim()
        .parse::<usize>()
        .map_err(|e| format!("error parsing commit count: {}", e))
}

fn resolve_ref(repo_path: &Path, ref_name: &str) -> Result<Commit, String> {
    let output = git_command(repo_path)
        .args([
            "log",
            "-1",
            "--format=%H%n%s%n%an%n%aI%n%P",
            ref_name.trim(),
        ])
        .output()
        .map_err(|e| format!("Failed to run git log: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git log failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    parse_commit_output(&String::from_utf8_lossy(&output.stdout))
}

fn parse_commit_output(output: &str) -> Result<Commit, String> {
    let lines: Vec<&str> = output.trim().lines().collect();
    if lines.len() < 4 {
        return Err("unexpected git log output".to_string());
    }

    let hash = lines[0].to_string();
    let message = lines[1].to_string();
    let author = lines[2].to_string();

    let date = chrono::DateTime::parse_from_rfc3339(lines[3].trim())
        .map_err(|e| format!("error parsing date: {}", e))?
        .with_timezone(&chrono::Utc);

    let parent_hashes = if lines.len() >= 5 && !lines[4].is_empty() {
        lines[4].split_whitespace().map(String::from).collect()
    } else {
        vec![]
    };

    Ok(Commit {
        hash,
        parent_hashes,
        refs: vec![],
        message,
        author,
        date,
    })
}

pub struct DiffSpec {
    pub raw: Option<String>,
    pub uncommitted_type: Option<UncommittedType>,
}

pub fn resolve_diff_source(repo_path: &Path, spec: DiffSpec) -> Result<DiffSource, String> {
    if let Some(uncommitted_type) = spec.uncommitted_type {
        return Ok(DiffSource::Uncommitted(uncommitted_type));
    }

    if let Some(raw) = spec.raw {
        let range = resolve_commit_range(repo_path, &raw)?;
        return Ok(DiffSource::CommitRange(range));
    }

    Err("no spec provided".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_range_spec_single_commit() {
        let (start, end) = parse_range_spec("abc123").unwrap();
        assert_eq!(start, "abc123");
        assert_eq!(end, "HEAD");
    }

    #[test]
    fn parse_range_spec_two_dot() {
        let (start, end) = parse_range_spec("main..feature").unwrap();
        assert_eq!(start, "main");
        assert_eq!(end, "feature");
    }

    #[test]
    fn parse_range_spec_three_dot() {
        let (start, end) = parse_range_spec("main...feature").unwrap();
        assert_eq!(start, "main");
        assert_eq!(end, "feature");
    }
}
