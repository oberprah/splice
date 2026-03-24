use crate::common::{reset_counter, Harness, TestRepo};
use serial_test::serial;
use splice::LogSpec;

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
    head_harness.assert_snapshot(
        r#"
"      Working tree clean                                                        "
"  → ├ b53b778 (main) D · 2d ago                                                 "
"    ├ cc4032c B · 2d ago                                                        "
"    ├ fe76018 A · 2d ago                                                        "
"                                                                                "
"                                                                                "
"                                                                                "
"                                                                                "
"                                                                                "
"  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
"#,
    );

    let mut all_harness =
        Harness::with_repo_and_log_spec_and_screen_size(&repo, LogSpec::All, 80, 10);
    all_harness.assert_snapshot(
        r#"
"      Working tree clean                                                        "
"  → ├ e8faeba (feature) C · 2d ago                                              "
"    │ ├ b53b778 (main) D · 2d ago                                               "
"    ├─╯ cc4032c B · 2d ago                                                      "
"    ├ fe76018 A · 2d ago                                                        "
"                                                                                "
"                                                                                "
"                                                                                "
"                                                                                "
"  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
"#,
    );
}
