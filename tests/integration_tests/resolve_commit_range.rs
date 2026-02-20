use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice_rust::git::resolve_commit_range;

#[test]
#[serial]
fn resolve_commit_range_single_commit() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");

    let first_hash = repo.rev_parse("HEAD");

    let range = resolve_commit_range(repo.path(), &first_hash).unwrap();

    assert_eq!(range.count, 1);
    assert_eq!(range.end.hash, first_hash);
    assert_eq!(range.end.message, "Initial commit");
}

#[test]
#[serial]
fn resolve_commit_range_two_dot() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("First commit");
    repo.commit("Second commit");
    repo.commit("Third commit");

    let first_hash = repo.rev_parse("HEAD~2");
    let third_hash = repo.rev_parse("HEAD");

    let range =
        resolve_commit_range(repo.path(), &format!("{}..{}", first_hash, third_hash)).unwrap();

    assert_eq!(range.count, 2);
    assert_eq!(range.start.hash, first_hash);
    assert_eq!(range.end.hash, third_hash);
}

#[test]
#[serial]
fn resolve_commit_range_three_dot() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");

    repo.create_branch("feature");
    repo.checkout("feature");
    repo.commit("Feature commit");

    repo.checkout("main");
    repo.commit("Main commit");

    let main_hash = repo.rev_parse("main");
    let feature_hash = repo.rev_parse("feature");

    let range = resolve_commit_range(repo.path(), "main...feature").unwrap();

    assert!(range.count >= 1);
    assert_ne!(range.start.hash, main_hash);
    assert_eq!(range.end.hash, feature_hash);
}

#[test]
#[serial]
fn resolve_commit_range_invalid_ref() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");

    let result = resolve_commit_range(repo.path(), "nonexistent");

    assert!(result.is_err());
    assert!(result.unwrap_err().contains("error resolving"));
}

#[test]
#[serial]
fn resolve_commit_range_with_branch_name() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");
    repo.commit("Second commit");

    let range = resolve_commit_range(repo.path(), "HEAD~1..main").unwrap();

    assert_eq!(range.count, 1);
    assert_eq!(range.end.message, "Second commit");
}
