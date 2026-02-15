mod parsing;

pub use parsing::parse_log_output;

use std::path::Path;

const MAX_COMMITS: usize = 100;
use std::process::Command;

use crate::core::Commit;

pub fn fetch_commits(repo_path: &Path) -> Result<Vec<Commit>, String> {
    let output = Command::new("git")
        .current_dir(repo_path)
        .args([
            "log",
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
