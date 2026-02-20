use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice_rust::git::{fetch_file_changes_for_source, resolve_diff_source, DiffSpec};

#[test]
#[serial]
fn fetch_file_changes_three_dot_excludes_merge_base_changes() {
    reset_counter();
    let repo = TestRepo::new();

    repo.add_file("base.txt", "base");
    repo.commit("Base commit");

    repo.create_branch("feature");
    repo.checkout("feature");
    repo.add_file("feature.txt", "feature");
    repo.commit("Feature commit");

    repo.checkout("main");
    repo.add_file("main.txt", "main");
    repo.commit("Main commit");

    let source = resolve_diff_source(
        repo.path(),
        DiffSpec {
            raw: Some("main...feature".to_string()),
            uncommitted_type: None,
        },
    )
    .unwrap();

    let changes = fetch_file_changes_for_source(repo.path(), &source).unwrap();
    let paths: Vec<&str> = changes.iter().map(|change| change.path.as_str()).collect();

    assert!(paths.contains(&"feature.txt"));
    assert!(!paths.contains(&"base.txt"));
    assert!(!paths.contains(&"main.txt"));
}
