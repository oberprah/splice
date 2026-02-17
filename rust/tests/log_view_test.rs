mod common;

use common::{reset_counter, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn test_log_view() {
    reset_counter();

    let repo = TestRepo::new();
    for i in 0..30 {
        repo.commit(&format!("Commit {i}"));
    }
    repo.create_branch("feature");
    repo.create_tag("v1.0.0");

    let mut h = common::Harness::with_repo(&repo);

    // Initial view shows first commit selected with branch and tag refs
    insta::assert_snapshot!(h.snapshot(), @r###"
    "  → f00429c (main, tag: v1.0.0, feature) Commit 29                              "
    "    d975d5b Commit 28                                                           "
    "    d6fcefe Commit 27                                                           "
    "    4c8fe93 Commit 26                                                           "
    "    9feaeb4 Commit 25                                                           "
    "    2e50686 Commit 24                                                           "
    "    cbbad51 Commit 23                                                           "
    "    dc6b3d8 Commit 22                                                           "
    "    984d749 Commit 21                                                           "
    "    8aeb3df Commit 20                                                           "
    "    15df8f3 Commit 19                                                           "
    "    ca24928 Commit 18                                                           "
    "    50c9441 Commit 17                                                           "
    "    b37126e Commit 16                                                           "
    "    3a1959a Commit 15                                                           "
    "    77f5868 Commit 14                                                           "
    "    9593ccc Commit 13                                                           "
    "    0198648 Commit 12                                                           "
    "    d4ec593 Commit 11                                                           "
    "    b33b874 Commit 10                                                           "
    "    1817c0e Commit 9                                                            "
    "    444ad36 Commit 8                                                            "
    "    31b4d64 Commit 7                                                            "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    // Ctrl+d moves down half a page
    h.press_ctrl(KeyCode::Char('d'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    f00429c (main, tag: v1.0.0, feature) Commit 29                              "
    "    d975d5b Commit 28                                                           "
    "    d6fcefe Commit 27                                                           "
    "    4c8fe93 Commit 26                                                           "
    "    9feaeb4 Commit 25                                                           "
    "    2e50686 Commit 24                                                           "
    "    cbbad51 Commit 23                                                           "
    "    dc6b3d8 Commit 22                                                           "
    "    984d749 Commit 21                                                           "
    "    8aeb3df Commit 20                                                           "
    "    15df8f3 Commit 19                                                           "
    "  → ca24928 Commit 18                                                           "
    "    50c9441 Commit 17                                                           "
    "    b37126e Commit 16                                                           "
    "    3a1959a Commit 15                                                           "
    "    77f5868 Commit 14                                                           "
    "    9593ccc Commit 13                                                           "
    "    0198648 Commit 12                                                           "
    "    d4ec593 Commit 11                                                           "
    "    b33b874 Commit 10                                                           "
    "    1817c0e Commit 9                                                            "
    "    444ad36 Commit 8                                                            "
    "    31b4d64 Commit 7                                                            "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    // j moves down one commit
    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    f00429c (main, tag: v1.0.0, feature) Commit 29                              "
    "    d975d5b Commit 28                                                           "
    "    d6fcefe Commit 27                                                           "
    "    4c8fe93 Commit 26                                                           "
    "    9feaeb4 Commit 25                                                           "
    "    2e50686 Commit 24                                                           "
    "    cbbad51 Commit 23                                                           "
    "    dc6b3d8 Commit 22                                                           "
    "    984d749 Commit 21                                                           "
    "    8aeb3df Commit 20                                                           "
    "    15df8f3 Commit 19                                                           "
    "    ca24928 Commit 18                                                           "
    "  → 50c9441 Commit 17                                                           "
    "    b37126e Commit 16                                                           "
    "    3a1959a Commit 15                                                           "
    "    77f5868 Commit 14                                                           "
    "    9593ccc Commit 13                                                           "
    "    0198648 Commit 12                                                           "
    "    d4ec593 Commit 11                                                           "
    "    b33b874 Commit 10                                                           "
    "    1817c0e Commit 9                                                            "
    "    444ad36 Commit 8                                                            "
    "    31b4d64 Commit 7                                                            "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    // Ctrl+u moves up half a page
    h.press_ctrl(KeyCode::Char('u'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "    f00429c (main, tag: v1.0.0, feature) Commit 29                              "
    "  → d975d5b Commit 28                                                           "
    "    d6fcefe Commit 27                                                           "
    "    4c8fe93 Commit 26                                                           "
    "    9feaeb4 Commit 25                                                           "
    "    2e50686 Commit 24                                                           "
    "    cbbad51 Commit 23                                                           "
    "    dc6b3d8 Commit 22                                                           "
    "    984d749 Commit 21                                                           "
    "    8aeb3df Commit 20                                                           "
    "    15df8f3 Commit 19                                                           "
    "    ca24928 Commit 18                                                           "
    "    50c9441 Commit 17                                                           "
    "    b37126e Commit 16                                                           "
    "    3a1959a Commit 15                                                           "
    "    77f5868 Commit 14                                                           "
    "    9593ccc Commit 13                                                           "
    "    0198648 Commit 12                                                           "
    "    d4ec593 Commit 11                                                           "
    "    b33b874 Commit 10                                                           "
    "    1817c0e Commit 9                                                            "
    "    444ad36 Commit 8                                                            "
    "    31b4d64 Commit 7                                                            "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}
