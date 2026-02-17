// TestRepo creates git repositories with deterministic commit hashes.
//
// Git commit hashes are derived from: tree content, parent hashes, author/committer
// info, and timestamps. By fixing all these values, we get predictable hashes:
//
// - COMMIT_COUNTER ensures unique file content per commit (unique tree hash)
// - GIT_AUTHOR_* and GIT_COMMITTER_* env vars fix identity and timestamps
// - Fixed date (2020-01-01T00:00:00+0000) ensures same timestamp across runs
//
// This allows snapshot tests to use actual commit hashes instead of regex replacements.
//
// NOTE: Tests using TestRepo should use `serial_test` and call `reset_counter()` at
// the start to ensure deterministic commit hashes across parallel test execution.
use std::process::Command;
use std::sync::atomic::{AtomicU64, Ordering};

use tempfile::TempDir;

static COMMIT_COUNTER: AtomicU64 = AtomicU64::new(0);

pub fn reset_counter() {
    COMMIT_COUNTER.store(0, Ordering::SeqCst);
}

pub struct TestRepo {
    _temp_dir: TempDir,
    path: std::path::PathBuf,
}

impl TestRepo {
    pub fn new() -> Self {
        let temp_dir = TempDir::new().expect("Failed to create temp dir");
        let path = temp_dir.path().to_path_buf();

        Self::run_git(&path, &["init", "-b", "main"]);
        Self::run_git(&path, &["config", "user.name", "Test Author"]);
        Self::run_git(&path, &["config", "user.email", "test@example.com"]);

        Self {
            _temp_dir: temp_dir,
            path,
        }
    }

    pub fn path(&self) -> &std::path::Path {
        &self.path
    }

    pub fn add_file(&self, path: &str, content: &str) {
        let file_path = self.path.join(path);
        if let Some(parent) = file_path.parent() {
            std::fs::create_dir_all(parent).expect("Failed to create dir");
        }
        std::fs::write(&file_path, content).expect("Failed to write file");
        Self::run_git(&self.path, &["add", path]);
    }

    pub fn modify_file(&self, path: &str, content: &str) {
        let file_path = self.path.join(path);
        std::fs::write(&file_path, content).expect("Failed to write file");
        Self::run_git(&self.path, &["add", path]);
    }

    pub fn commit(&self, message: &str) {
        let counter = COMMIT_COUNTER.fetch_add(1, Ordering::SeqCst);
        let file_name = format!("file_{}.txt", counter);
        let file_path = self.path.join(&file_name);
        std::fs::write(&file_path, format!("content_{}", counter)).expect("Failed to write file");

        Self::run_git(&self.path, &["add", &file_name]);
        Self::run_git_with_env(
            &self.path,
            &["commit", "-m", message],
            &[
                ("GIT_AUTHOR_NAME", "Test"),
                ("GIT_AUTHOR_EMAIL", "test@test.com"),
                ("GIT_COMMITTER_NAME", "Test"),
                ("GIT_COMMITTER_EMAIL", "test@test.com"),
                ("GIT_AUTHOR_DATE", "2020-01-01T00:00:00+0000"),
                ("GIT_COMMITTER_DATE", "2020-01-01T00:00:00+0000"),
            ],
        );
    }

    pub fn create_branch(&self, name: &str) {
        Self::run_git(&self.path, &["branch", name]);
    }

    pub fn checkout(&self, ref_name: &str) {
        Self::run_git(&self.path, &["checkout", ref_name]);
    }

    pub fn merge(&self, branch: &str) {
        Self::run_git_with_env(
            &self.path,
            &["merge", branch, "--no-ff", "-m", &format!("Merge {}", branch)],
            &[
                ("GIT_AUTHOR_NAME", "Test"),
                ("GIT_AUTHOR_EMAIL", "test@test.com"),
                ("GIT_COMMITTER_NAME", "Test"),
                ("GIT_COMMITTER_EMAIL", "test@test.com"),
                ("GIT_AUTHOR_DATE", "2020-01-01T00:00:00+0000"),
                ("GIT_COMMITTER_DATE", "2020-01-01T00:00:00+0000"),
            ],
        );
    }

    pub fn create_tag(&self, name: &str) {
        Self::run_git(&self.path, &["tag", name]);
    }

    fn run_git(path: &std::path::Path, args: &[&str]) {
        let output = Command::new("git")
            .current_dir(path)
            .args(args)
            .output()
            .expect("Failed to run git");

        if !output.status.success() {
            panic!(
                "git {:?} failed: {}",
                args,
                String::from_utf8_lossy(&output.stderr)
            );
        }
    }

    fn run_git_with_env(path: &std::path::Path, args: &[&str], env_vars: &[(&str, &str)]) {
        let mut cmd = Command::new("git");
        cmd.current_dir(path).args(args);
        for (key, value) in env_vars {
            cmd.env(key, value);
        }
        let output = cmd.output().expect("Failed to run git");

        if !output.status.success() {
            panic!(
                "git {:?} failed: {}",
                args,
                String::from_utf8_lossy(&output.stderr)
            );
        }
    }
}
