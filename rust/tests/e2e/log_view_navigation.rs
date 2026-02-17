use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn log_view_navigation() {
    reset_counter();

    let repo = TestRepo::new();
    for i in 0..30 {
        repo.commit(&format!("Commit {i}"));
    }
    repo.create_branch("feature");
    repo.create_tag("v1.0.0");

    let mut h = Harness::with_repo(&repo);

    insta::assert_snapshot!(h.snapshot(), @r###"
    "  → ├ f00429c (main, tag: v1.0.0, feature) Commit 29                            "
    "    ├ d975d5b Commit 28                                                         "
    "    ├ d6fcefe Commit 27                                                         "
    "    ├ 4c8fe93 Commit 26                                                         "
    "    ├ 9feaeb4 Commit 25                                                         "
    "    ├ 2e50686 Commit 24                                                         "
    "    ├ cbbad51 Commit 23                                                         "
    "    ├ dc6b3d8 Commit 22                                                         "
    "    ├ 984d749 Commit 21                                                         "
    "    ├ 8aeb3df Commit 20                                                         "
    "    ├ 15df8f3 Commit 19                                                         "
    "    ├ ca24928 Commit 18                                                         "
    "    ├ 50c9441 Commit 17                                                         "
    "    ├ b37126e Commit 16                                                         "
    "    ├ 3a1959a Commit 15                                                         "
    "    ├ 77f5868 Commit 14                                                         "
    "    ├ 9593ccc Commit 13                                                         "
    "    ├ 0198648 Commit 12                                                         "
    "    ├ d4ec593 Commit 11                                                         "
    "    ├ b33b874 Commit 10                                                         "
    "    ├ 1817c0e Commit 9                                                          "
    "    ├ 444ad36 Commit 8                                                          "
    "    ├ 31b4d64 Commit 7                                                          "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press_ctrl(KeyCode::Char('d'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    ├ f00429c (main, tag: v1.0.0, feature) Commit 29                            "
    "    ├ d975d5b Commit 28                                                         "
    "    ├ d6fcefe Commit 27                                                         "
    "    ├ 4c8fe93 Commit 26                                                         "
    "    ├ 9feaeb4 Commit 25                                                         "
    "    ├ 2e50686 Commit 24                                                         "
    "    ├ cbbad51 Commit 23                                                         "
    "    ├ dc6b3d8 Commit 22                                                         "
    "    ├ 984d749 Commit 21                                                         "
    "    ├ 8aeb3df Commit 20                                                         "
    "    ├ 15df8f3 Commit 19                                                         "
    "  → ├ ca24928 Commit 18                                                         "
    "    ├ 50c9441 Commit 17                                                         "
    "    ├ b37126e Commit 16                                                         "
    "    ├ 3a1959a Commit 15                                                         "
    "    ├ 77f5868 Commit 14                                                         "
    "    ├ 9593ccc Commit 13                                                         "
    "    ├ 0198648 Commit 12                                                         "
    "    ├ d4ec593 Commit 11                                                         "
    "    ├ b33b874 Commit 10                                                         "
    "    ├ 1817c0e Commit 9                                                          "
    "    ├ 444ad36 Commit 8                                                          "
    "    ├ 31b4d64 Commit 7                                                          "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    ├ f00429c (main, tag: v1.0.0, feature) Commit 29                            "
    "    ├ d975d5b Commit 28                                                         "
    "    ├ d6fcefe Commit 27                                                         "
    "    ├ 4c8fe93 Commit 26                                                         "
    "    ├ 9feaeb4 Commit 25                                                         "
    "    ├ 2e50686 Commit 24                                                         "
    "    ├ cbbad51 Commit 23                                                         "
    "    ├ dc6b3d8 Commit 22                                                         "
    "    ├ 984d749 Commit 21                                                         "
    "    ├ 8aeb3df Commit 20                                                         "
    "    ├ 15df8f3 Commit 19                                                         "
    "    ├ ca24928 Commit 18                                                         "
    "  → ├ 50c9441 Commit 17                                                         "
    "    ├ b37126e Commit 16                                                         "
    "    ├ 3a1959a Commit 15                                                         "
    "    ├ 77f5868 Commit 14                                                         "
    "    ├ 9593ccc Commit 13                                                         "
    "    ├ 0198648 Commit 12                                                         "
    "    ├ d4ec593 Commit 11                                                         "
    "    ├ b33b874 Commit 10                                                         "
    "    ├ 1817c0e Commit 9                                                          "
    "    ├ 444ad36 Commit 8                                                          "
    "    ├ 31b4d64 Commit 7                                                          "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press_ctrl(KeyCode::Char('u'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    ├ f00429c (main, tag: v1.0.0, feature) Commit 29                            "
    "  → ├ d975d5b Commit 28                                                         "
    "    ├ d6fcefe Commit 27                                                         "
    "    ├ 4c8fe93 Commit 26                                                         "
    "    ├ 9feaeb4 Commit 25                                                         "
    "    ├ 2e50686 Commit 24                                                         "
    "    ├ cbbad51 Commit 23                                                         "
    "    ├ dc6b3d8 Commit 22                                                         "
    "    ├ 984d749 Commit 21                                                         "
    "    ├ 8aeb3df Commit 20                                                         "
    "    ├ 15df8f3 Commit 19                                                         "
    "    ├ ca24928 Commit 18                                                         "
    "    ├ 50c9441 Commit 17                                                         "
    "    ├ b37126e Commit 16                                                         "
    "    ├ 3a1959a Commit 15                                                         "
    "    ├ 77f5868 Commit 14                                                         "
    "    ├ 9593ccc Commit 13                                                         "
    "    ├ 0198648 Commit 12                                                         "
    "    ├ d4ec593 Commit 11                                                         "
    "    ├ b33b874 Commit 10                                                         "
    "    ├ 1817c0e Commit 9                                                          "
    "    ├ 444ad36 Commit 8                                                          "
    "    ├ 31b4d64 Commit 7                                                          "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}

#[test]
#[serial]
fn log_view_with_merge_graph() {
    reset_counter();

    let repo = TestRepo::new();
    repo.commit("Initial commit");
    repo.commit("Second commit");
    repo.create_branch("feature");
    repo.checkout("feature");
    repo.commit("Feature commit 1");
    repo.commit("Feature commit 2");
    repo.checkout("main");
    repo.commit("Main commit after branch");
    repo.merge("feature");

    let mut h = Harness::with_repo(&repo);

    insta::assert_snapshot!(h.snapshot(), @r###"
    "  → ├─╮ 872d8a4 (main) Merge feature                                            "
    "    ├ │ 8afef21 Main commit after branch                                        "
    "    │ ├ f1cb02f (feature) Feature commit 2                                      "
    "    ├ │ a6e1387 Second commit                                                   "
    "    │ ├ 30beeb9 Feature commit 1                                                "
    "    ├ │ b2c992c Initial commit                                                  "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}
