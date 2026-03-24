use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice::core::LogSpec;
use splice::git::fetch_commits;

#[test]
#[serial]
fn fetch_commits_includes_branch_ref() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");
    repo.create_branch("feature");

    let commits = fetch_commits(repo.path(), LogSpec::Head).unwrap();

    assert_eq!(commits.len(), 1);
    let branch_refs: Vec<_> = commits[0]
        .refs
        .iter()
        .filter(|r| r.name == "feature")
        .collect();
    assert_eq!(branch_refs.len(), 1);
}

#[test]
#[serial]
fn fetch_commits_includes_tag_ref() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");
    repo.create_tag("v1.0.0");

    let commits = fetch_commits(repo.path(), LogSpec::Head).unwrap();

    assert_eq!(commits.len(), 1);
    let tag_refs: Vec<_> = commits[0]
        .refs
        .iter()
        .filter(|r| r.name == "v1.0.0")
        .collect();
    assert_eq!(tag_refs.len(), 1);
    assert!(tag_refs
        .iter()
        .all(|r| r.ref_type == splice::core::RefType::Tag));
}

#[test]
#[serial]
fn fetch_commits_marks_head_branch() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");

    let commits = fetch_commits(repo.path(), LogSpec::Head).unwrap();

    assert_eq!(commits.len(), 1);
    let head_refs: Vec<_> = commits[0].refs.iter().filter(|r| r.is_head).collect();
    assert_eq!(head_refs.len(), 1);
}
