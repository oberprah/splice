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
