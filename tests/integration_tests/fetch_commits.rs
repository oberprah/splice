use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice_rust::git::fetch_commits;

#[test]
#[serial]
fn fetch_commits_returns_single_commit() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");
    let commits = fetch_commits(repo.path()).unwrap();

    assert_eq!(commits.len(), 1);
    assert_eq!(commits[0].message, "Initial commit");
    assert!(!commits[0].hash.is_empty());
    assert!(commits[0].parent_hashes.is_empty());
}

#[test]
#[serial]
fn fetch_commits_returns_commits_in_reverse_chronological_order() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("First commit");
    repo.commit("Second commit");
    repo.commit("Third commit");

    let commits = fetch_commits(repo.path()).unwrap();

    assert_eq!(commits.len(), 3);
    assert_eq!(commits[0].message, "Third commit");
    assert_eq!(commits[1].message, "Second commit");
    assert_eq!(commits[2].message, "First commit");
}

#[test]
#[serial]
fn fetch_commits_includes_parent_hashes() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("First");
    repo.commit("Second");

    let commits = fetch_commits(repo.path()).unwrap();

    assert_eq!(commits.len(), 2);
    assert!(!commits[0].parent_hashes.is_empty());
    assert_eq!(commits[0].parent_hashes[0], commits[1].hash);
}

#[test]
#[serial]
fn fetch_commits_includes_author_info() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Test commit");
    let commits = fetch_commits(repo.path()).unwrap();

    assert_eq!(commits.len(), 1);
    assert!(!commits[0].author.is_empty());
}

#[test]
#[serial]
fn fetch_commits_returns_full_40_char_hash() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Test commit");
    let commits = fetch_commits(repo.path()).unwrap();

    assert_eq!(commits.len(), 1);
    assert_eq!(commits[0].hash.len(), 40);
    assert!(commits[0].hash.chars().all(|c| c.is_ascii_hexdigit()));
}
