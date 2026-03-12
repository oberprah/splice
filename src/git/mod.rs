mod diff;
mod file_changes;
mod log;
mod resolve;
mod uncommitted;

pub use diff::{
    fetch_file_content, fetch_file_diff, fetch_full_file_diff, fetch_full_uncommitted_file_diff,
    FullFileDiff,
};
pub use file_changes::parse_file_changes;
pub use log::parse_log_output;
pub use resolve::{resolve_commit_range, resolve_diff_source, DiffSpec};
pub use uncommitted::{fetch_uncommitted_file_changes, fetch_uncommitted_summary};

use std::path::{Path, PathBuf};
use std::process::Command;

use crate::core::{Commit, CommitRange, DiffSource, FileChange, LogSpec};

fn git_command(repo_path: &Path) -> Command {
    let mut cmd = Command::new("git");
    cmd.current_dir(repo_path)
        .env_remove("GIT_DIR")
        .env_remove("GIT_WORK_TREE")
        .env_remove("GIT_INDEX_FILE")
        .env_remove("GIT_COMMON_DIR")
        .env_remove("GIT_OBJECT_DIRECTORY")
        .env_remove("GIT_ALTERNATE_OBJECT_DIRECTORIES");
    cmd
}

pub fn fetch_commits(repo_path: &Path, spec: LogSpec) -> Result<Vec<Commit>, String> {
    let mut args = vec![
        "log".to_string(),
        "--topo-order".to_string(),
        "--pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e".to_string(),
        "--date=iso-strict".to_string(),
    ];

    match spec {
        LogSpec::Head => {}
        LogSpec::All => args.push("--all".to_string()),
        LogSpec::Rev(rev) => args.push(rev),
    }

    let output = git_command(repo_path)
        .args(&args)
        .output()
        .map_err(|e| format!("Failed to run git: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git log failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    let stdout = String::from_utf8_lossy(&output.stdout);
    parse_log_output(&stdout)
}

pub fn fetch_file_changes(
    repo_path: &Path,
    range: &CommitRange,
) -> Result<Vec<FileChange>, String> {
    if range.is_single_commit() {
        fetch_file_changes_single(repo_path, &range.end.hash)
    } else {
        fetch_file_changes_range(repo_path, range)
    }
}

fn fetch_file_changes_single(
    repo_path: &Path,
    commit_hash: &str,
) -> Result<Vec<FileChange>, String> {
    let numstat_output = git_command(repo_path)
        .args([
            "diff-tree",
            "--no-commit-id",
            "--numstat",
            "-r",
            "-M",
            "--root",
            commit_hash,
        ])
        .output()
        .map_err(|e| format!("Failed to run git diff-tree: {}", e))?;

    if !numstat_output.status.success() {
        return Err(format!(
            "git diff-tree --numstat failed: {}",
            String::from_utf8_lossy(&numstat_output.stderr)
        ));
    }

    let name_status_output = git_command(repo_path)
        .args([
            "diff-tree",
            "--no-commit-id",
            "--name-status",
            "-r",
            "-M",
            "--root",
            commit_hash,
        ])
        .output()
        .map_err(|e| format!("Failed to run git diff-tree: {}", e))?;

    if !name_status_output.status.success() {
        return Err(format!(
            "git diff-tree --name-status failed: {}",
            String::from_utf8_lossy(&name_status_output.stderr)
        ));
    }

    let numstat = String::from_utf8_lossy(&numstat_output.stdout);
    let name_status = String::from_utf8_lossy(&name_status_output.stdout);

    parse_file_changes(&numstat, &name_status)
}

fn fetch_file_changes_range(
    repo_path: &Path,
    range: &CommitRange,
) -> Result<Vec<FileChange>, String> {
    let range_spec = range.to_diff_spec();

    let numstat_output = git_command(repo_path)
        .args(["diff", "--numstat", "-M", &range_spec])
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !numstat_output.status.success() {
        return Err(format!(
            "git diff --numstat failed: {}",
            String::from_utf8_lossy(&numstat_output.stderr)
        ));
    }

    let name_status_output = git_command(repo_path)
        .args(["diff", "--name-status", "-M", &range_spec])
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !name_status_output.status.success() {
        return Err(format!(
            "git diff --name-status failed: {}",
            String::from_utf8_lossy(&name_status_output.stderr)
        ));
    }

    let numstat = String::from_utf8_lossy(&numstat_output.stdout);
    let name_status = String::from_utf8_lossy(&name_status_output.stdout);

    parse_file_changes(&numstat, &name_status)
}

pub fn fetch_file_changes_for_source(
    repo_path: &Path,
    source: &DiffSource,
) -> Result<Vec<FileChange>, String> {
    match source {
        DiffSource::CommitRange(range) => fetch_file_changes(repo_path, range),
        DiffSource::Uncommitted(uncommitted_type) => {
            fetch_uncommitted_file_changes(repo_path, *uncommitted_type)
        }
    }
}

pub fn repository_root(repo_path: &Path) -> Result<PathBuf, String> {
    let output = git_command(repo_path)
        .args(["rev-parse", "--show-toplevel"])
        .output()
        .map_err(|e| format!("Failed to run git rev-parse: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git rev-parse --show-toplevel failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    let stdout = String::from_utf8_lossy(&output.stdout);
    Ok(PathBuf::from(stdout.trim()))
}

pub fn fetch_full_file_diff_for_source(
    repo_path: &Path,
    source: &DiffSource,
    path: &str,
) -> Result<FullFileDiff, String> {
    match source {
        DiffSource::CommitRange(range) => fetch_full_file_diff(repo_path, range, path),
        DiffSource::Uncommitted(uncommitted_type) => {
            fetch_full_uncommitted_file_diff(repo_path, *uncommitted_type, path)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::UncommittedType;

    #[test]
    fn fetch_file_changes_for_source_dispatches_to_commit_range() {
        let repo_path = Path::new("/nonexistent");
        let commit = crate::core::Commit {
            hash: "abc123".to_string(),
            parent_hashes: vec![],
            refs: vec![],
            message: "test".to_string(),
            body: None,
            author: "test".to_string(),
            date: chrono::Utc::now(),
        };
        let range = CommitRange {
            start: commit.clone(),
            end: commit,
            count: 1,
            include_start: true,
        };
        let source = DiffSource::CommitRange(range);
        let result = fetch_file_changes_for_source(repo_path, &source);
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("git diff-tree"));
    }

    #[test]
    fn fetch_file_changes_for_source_dispatches_to_uncommitted() {
        let repo_path = Path::new("/nonexistent");
        let source = DiffSource::Uncommitted(UncommittedType::Staged);
        let result = fetch_file_changes_for_source(repo_path, &source);
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("git diff"));
    }

    #[test]
    fn repository_root_reports_rev_parse_error() {
        let repo_path = Path::new("/nonexistent");
        let result = repository_root(repo_path);
        assert!(result.is_err());
        assert!(result.unwrap_err().contains("git rev-parse"));
    }
}
