use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn log_view_complex_graph_navigation() {
    reset_counter();

    let repo = TestRepo::new();

    repo.commit("A");
    repo.commit("B");
    repo.create_branch("feature-1");
    repo.checkout("feature-1");
    repo.commit("C");

    repo.checkout("main");
    repo.merge("feature-1");
    let d_hash = repo.rev_parse("HEAD");

    repo.create_branch("hotfix");
    repo.checkout("hotfix");
    repo.commit("E");

    repo.checkout("main");
    repo.merge("hotfix");

    repo.create_branch("feature-2");
    repo.checkout("feature-2");
    repo.commit("G");
    let g_hash = repo.rev_parse("HEAD");

    repo.checkout("main");
    repo.merge("feature-2");

    repo.checkout_hash(&g_hash);
    repo.create_branch("feature-2b");
    repo.checkout("feature-2b");
    repo.commit("I");

    repo.checkout("main");
    repo.merge("feature-2b");

    repo.checkout_hash(&d_hash);
    repo.create_branch("feature-3");
    repo.checkout("feature-3");
    repo.commit("K");
    repo.commit("L");

    repo.checkout("main");
    repo.merge("feature-3");

    for i in 0..12 {
        repo.commit(&format!("Linear {i}"));
    }

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(r###"
    "  → ├ d91fc04 (main) Linear 11                                                  "
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "    ├ 73f72bc Linear 0                                                          "
    "    ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(r###"
    "    ├ d91fc04 (main) Linear 11                                                  "
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "  → ├ 73f72bc Linear 0                                                          "
    "    ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(r###"
    "    ├ d91fc04 (main) Linear 11                                                  "
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "    ├ 73f72bc Linear 0                                                          "
    "    ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "  → │ ├ e8faeba (feature-1) C                                                   "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(r###"
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "    ├ 73f72bc Linear 0                                                          "
    "    ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "  → ├─╯ cc4032c B                                                               "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press_ctrl(KeyCode::Char('u'));
    h.assert_snapshot(r###"
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "    ├ 73f72bc Linear 0                                                          "
    "  → ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "    ├─╯ cc4032c B                                                               "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);

    h.press(KeyCode::Char('k'));
    h.assert_snapshot(r###"
    "    ├ 32e3e44 Linear 10                                                         "
    "    ├ a561433 Linear 9                                                          "
    "    ├ b670e03 Linear 8                                                          "
    "    ├ b091d37 Linear 7                                                          "
    "    ├ 4aaaccb Linear 6                                                          "
    "    ├ 4651237 Linear 5                                                          "
    "    ├ d1ad270 Linear 4                                                          "
    "    ├ 2c74952 Linear 3                                                          "
    "    ├ adcfad5 Linear 2                                                          "
    "    ├ 2a31f86 Linear 1                                                          "
    "  → ├ 73f72bc Linear 0                                                          "
    "    ├─╮ 45fe0e3 Merge feature-3                                                 "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "    ├─╯ cc4032c B                                                               "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}
