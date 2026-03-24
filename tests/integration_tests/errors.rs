use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice::core::LogSpec;
use splice::git::fetch_commits;
use std::path::PathBuf;

#[test]
#[serial]
fn fetch_commits_returns_error_for_empty_repo() {
    reset_counter();
    let repo = TestRepo::new();
    let result = fetch_commits(repo.path(), LogSpec::Head);
    assert!(result.is_err());
}

#[test]
#[serial]
fn fetch_commits_returns_error_for_nonexistent_path() {
    reset_counter();
    let invalid_path = PathBuf::from("/nonexistent/path/that/does/not/exist");
    let result = fetch_commits(&invalid_path, LogSpec::Head);

    assert!(result.is_err());
}

#[test]
#[serial]
fn fetch_commits_returns_error_for_non_git_directory() {
    reset_counter();
    let temp_dir = tempfile::TempDir::new().unwrap();
    let result = fetch_commits(temp_dir.path(), LogSpec::Head);

    assert!(result.is_err());
}
