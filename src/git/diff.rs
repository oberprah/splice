use std::path::Path;

use super::git_command;
use crate::core::{CommitRange, UncommittedType};

pub fn fetch_file_diff(repo_path: &Path, commit_hash: &str, path: &str) -> Result<String, String> {
    let output = git_command(repo_path)
        .args([
            "diff-tree",
            "--no-commit-id",
            "--root",
            "-p",
            commit_hash,
            "--",
            path,
        ])
        .output()
        .map_err(|e| format!("Failed to run git diff-tree: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git diff-tree -p failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}

pub fn fetch_file_content(
    repo_path: &Path,
    commit_hash: &str,
    path: &str,
) -> Result<String, String> {
    let output = git_command(repo_path)
        .args(["show", &format!("{}:{}", commit_hash, path)])
        .output()
        .map_err(|e| format!("Failed to run git show: {}", e))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        if stderr.contains("does not exist") || stderr.contains("exists on disk, but not in") {
            return Ok(String::new());
        }
        return Err(format!("git show failed: {}", stderr));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}

fn fetch_index_file_content(repo_path: &Path, path: &str) -> Result<String, String> {
    let output = git_command(repo_path)
        .args(["show", &format!(":{}", path)])
        .output()
        .map_err(|e| format!("Failed to run git show: {}", e))?;

    if !output.status.success() {
        let stderr = String::from_utf8_lossy(&output.stderr);
        if stderr.contains("does not exist") || stderr.contains("exists on disk, but not in") {
            return Ok(String::new());
        }
        return Err(format!("git show failed: {}", stderr));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}

fn fetch_working_tree_content(repo_path: &Path, path: &str) -> Result<String, String> {
    let file_path = repo_path.join(path);
    match std::fs::read_to_string(&file_path) {
        Ok(content) => Ok(content),
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(String::new()),
        Err(err) => Err(format!(
            "Failed to read file {}: {}",
            file_path.display(),
            err
        )),
    }
}

pub struct FullFileDiff {
    pub old_content: String,
    pub new_content: String,
    pub diff_output: String,
}

pub fn fetch_full_file_diff(
    repo_path: &Path,
    range: &CommitRange,
    path: &str,
) -> Result<FullFileDiff, String> {
    let old_content = fetch_file_content(repo_path, &range.diff_base_spec(), path)?;
    let new_content = fetch_file_content(repo_path, &range.end.hash, path)?;
    let diff_output = fetch_file_diff_range(repo_path, range, path)?;

    Ok(FullFileDiff {
        old_content,
        new_content,
        diff_output,
    })
}

pub fn fetch_full_uncommitted_file_diff(
    repo_path: &Path,
    uncommitted_type: UncommittedType,
    path: &str,
) -> Result<FullFileDiff, String> {
    let (old_content, new_content) = match uncommitted_type {
        UncommittedType::Unstaged => (
            fetch_index_file_content(repo_path, path)?,
            fetch_working_tree_content(repo_path, path)?,
        ),
        UncommittedType::Staged => (
            fetch_file_content(repo_path, "HEAD", path)?,
            fetch_index_file_content(repo_path, path)?,
        ),
        UncommittedType::All => (
            fetch_file_content(repo_path, "HEAD", path)?,
            fetch_working_tree_content(repo_path, path)?,
        ),
    };

    let diff_output = fetch_file_diff_uncommitted(repo_path, uncommitted_type, path)?;

    Ok(FullFileDiff {
        old_content,
        new_content,
        diff_output,
    })
}

fn fetch_file_diff_range(
    repo_path: &Path,
    range: &CommitRange,
    path: &str,
) -> Result<String, String> {
    let diff_spec = range.to_diff_spec();
    let output = git_command(repo_path)
        .args(["diff", &diff_spec, "--", path])
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git diff failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}

fn fetch_file_diff_uncommitted(
    repo_path: &Path,
    uncommitted_type: UncommittedType,
    path: &str,
) -> Result<String, String> {
    let args: Vec<&str> = match uncommitted_type {
        UncommittedType::Unstaged => vec!["diff"],
        UncommittedType::Staged => vec!["diff", "--staged"],
        UncommittedType::All => vec!["diff", "HEAD"],
    };

    let output = git_command(repo_path)
        .args(&args)
        .arg("--")
        .arg(path)
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !output.status.success() {
        return Err(format!(
            "git diff failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ));
    }

    Ok(String::from_utf8_lossy(&output.stdout).to_string())
}
