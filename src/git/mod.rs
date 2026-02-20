mod diff;
mod file_changes;
mod log;

pub use diff::{fetch_file_content, fetch_file_diff, fetch_full_file_diff, FullFileDiff};
pub use file_changes::parse_file_changes;
pub use log::parse_log_output;

use std::path::Path;
use std::process::Command;

use crate::core::{Commit, CommitRange, FileChange};

const MAX_COMMITS: usize = 100;

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

pub fn fetch_commits(repo_path: &Path) -> Result<Vec<Commit>, String> {
    let output = git_command(repo_path)
        .args([
            "log",
            "--topo-order",
            "--pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x1e",
            "--date=iso-strict",
            "-n",
            &MAX_COMMITS.to_string(),
        ])
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
        fetch_file_changes_range(repo_path, &range.start.hash, &range.end.hash)
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
    start_hash: &str,
    end_hash: &str,
) -> Result<Vec<FileChange>, String> {
    let range_spec = format!("{}^..{}", start_hash, end_hash);

    let numstat_output = git_command(repo_path)
        .args(["diff", "--numstat", &range_spec])
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !numstat_output.status.success() {
        return Err(format!(
            "git diff --numstat failed: {}",
            String::from_utf8_lossy(&numstat_output.stderr)
        ));
    }

    let name_status_output = git_command(repo_path)
        .args(["diff", "--name-status", &range_spec])
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
