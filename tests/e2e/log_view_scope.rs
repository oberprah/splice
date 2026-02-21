use crate::common::{reset_counter, Harness, TestRepo};
use serial_test::serial;
use splice_rust::LogSpec;

#[test]
#[serial]
fn log_view_all_branches_includes_unmerged_branch() {
    reset_counter();

    let repo = TestRepo::new();
    repo.commit("A");
    repo.commit("B");

    repo.create_branch("feature");
    repo.checkout("feature");
    repo.commit("C");

    repo.checkout("main");
    repo.commit("D");

    let mut head_harness = Harness::with_repo_and_screen_size(&repo, 80, 10);
    let head_snapshot = head_harness.snapshot();
    assert!(!head_snapshot.contains("(feature)"));

    let mut all_harness =
        Harness::with_repo_and_log_spec_and_screen_size(&repo, LogSpec::All, 80, 10);
    let all_snapshot = all_harness.snapshot();
    assert!(all_snapshot.contains("(feature)"));
}
