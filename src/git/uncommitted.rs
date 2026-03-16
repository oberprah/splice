use std::path::Path;

use crate::core::{FileDiffInfo, UncommittedType};

use super::{git_command, parse_file_changes};

pub fn fetch_uncommitted_file_changes(
    repo_path: &Path,
    uncommitted_type: UncommittedType,
) -> Result<Vec<FileDiffInfo>, String> {
    let args: Vec<&str> = match uncommitted_type {
        UncommittedType::Unstaged => vec!["diff", "-M"],
        UncommittedType::Staged => vec!["diff", "--staged", "-M"],
        UncommittedType::All => vec!["diff", "HEAD", "-M"],
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

fn diff_has_changes(repo_path: &Path, args: &[&str]) -> Result<bool, String> {
    let output = git_command(repo_path)
        .args(args)
        .arg("--quiet")
        .output()
        .map_err(|e| format!("Failed to run git diff: {}", e))?;

    if output.status.success() {
        Ok(false)
    } else if output.status.code() == Some(1) {
        Ok(true)
    } else {
        Err(format!(
            "git diff --quiet failed: {}",
            String::from_utf8_lossy(&output.stderr)
        ))
    }
}

pub fn fetch_uncommitted_summary(
    repo_path: &Path,
) -> Result<(Option<UncommittedType>, usize), String> {
    let has_unstaged = diff_has_changes(repo_path, &["diff"])?;
    let has_staged = diff_has_changes(repo_path, &["diff", "--staged"])?;

    let uncommitted_type = match (has_unstaged, has_staged) {
        (false, false) => return Ok((None, 0)),
        (true, false) => UncommittedType::Unstaged,
        (false, true) => UncommittedType::Staged,
        (true, true) => UncommittedType::All,
    };

    let files = fetch_uncommitted_file_changes(repo_path, uncommitted_type)?;
    Ok((Some(uncommitted_type), files.len()))
}
