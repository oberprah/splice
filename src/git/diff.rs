use std::path::Path;

use super::git_command;

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

pub struct FullFileDiff {
    pub old_content: String,
    pub new_content: String,
    pub diff_output: String,
}

pub fn fetch_full_file_diff(
    repo_path: &Path,
    commit_hash: &str,
    path: &str,
) -> Result<FullFileDiff, String> {
    let old_content = fetch_file_content(repo_path, &format!("{}^", commit_hash), path)?;
    let new_content = fetch_file_content(repo_path, commit_hash, path)?;
    let diff_output = fetch_file_diff(repo_path, commit_hash, path)?;

    Ok(FullFileDiff {
        old_content,
        new_content,
        diff_output,
    })
}
