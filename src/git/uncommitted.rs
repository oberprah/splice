use std::path::Path;

use crate::core::{FileChange, UncommittedType};

use super::{git_command, parse_file_changes};

pub fn fetch_uncommitted_file_changes(
    repo_path: &Path,
    uncommitted_type: UncommittedType,
) -> Result<Vec<FileChange>, String> {
    let args: Vec<&str> = match uncommitted_type {
        UncommittedType::Unstaged => vec!["diff"],
        UncommittedType::Staged => vec!["diff", "--staged"],
        UncommittedType::All => vec!["diff", "HEAD"],
    };

    let numstat_output = git_command(repo_path)
        .args(&args)
        .arg("--numstat")
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if !numstat_output.status.success() {
        return Err(format!(
            "git diff --numstat failed: {}",
            String::from_utf8_lossy(&numstat_output.stderr)
        ));
    }

    let name_status_output = git_command(repo_path)
        .args(&args)
        .arg("--name-status")
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
