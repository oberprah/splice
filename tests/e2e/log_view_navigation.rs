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

    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "  → ├ d91fc04 (main) Linear 11 · 2d ago                                         "
    "    ├ 32e3e44 Linear 10 · 2d ago                                                "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "    ├ 73f72bc Linear 0 · 2d ago                                                 "
    "    ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "    ├ d91fc04 (main) Linear 11 · 2d ago                                         "
    "    ├ 32e3e44 Linear 10 · 2d ago                                                "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "  → ├ 73f72bc Linear 0 · 2d ago                                                 "
    "    ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "    ├ 32e3e44 Linear 10 · 2d ago                                                "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "    ├ 73f72bc Linear 0 · 2d ago                                                 "
    "    ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "  → │ ├ e8faeba (feature-1) C · 2d ago                                          "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "    ├ 73f72bc Linear 0 · 2d ago                                                 "
    "    ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "    │ ├ e8faeba (feature-1) C · 2d ago                                          "
    "  → ├─╯ cc4032c B · 2d ago                                                      "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );

    h.press_ctrl(KeyCode::Char('u'));
    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "    ├ 73f72bc Linear 0 · 2d ago                                                 "
    "  → ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "    │ ├ e8faeba (feature-1) C · 2d ago                                          "
    "    ├─╯ cc4032c B · 2d ago                                                      "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );

    h.press(KeyCode::Char('k'));
    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "    ├ a561433 Linear 9 · 2d ago                                                 "
    "    ├ b670e03 Linear 8 · 2d ago                                                 "
    "    ├ b091d37 Linear 7 · 2d ago                                                 "
    "    ├ 4aaaccb Linear 6 · 2d ago                                                 "
    "    ├ 4651237 Linear 5 · 2d ago                                                 "
    "    ├ d1ad270 Linear 4 · 2d ago                                                 "
    "    ├ 2c74952 Linear 3 · 2d ago                                                 "
    "    ├ adcfad5 Linear 2 · 2d ago                                                 "
    "    ├ 2a31f86 Linear 1 · 2d ago                                                 "
    "  → ├ 73f72bc Linear 0 · 2d ago                                                 "
    "    ├─╮ 45fe0e3 Merge feature-3 · 2d ago                                        "
    "    │ ├ 44eee0d (feature-3) L · 2d ago                                          "
    "    │ ├ 4c75a9d K · 2d ago                                                      "
    "    ├─│─╮ 55fbe6d Merge feature-2b · 2d ago                                     "
    "    │ │ ├ 02f5e0a (feature-2b) I · 2d ago                                       "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2 · 2d ago                                    "
    "    │ │ ├─╯ 6385fe0 (feature-2) G · 2d ago                                      "
    "    ├─│─┤ 16b8252 Merge hotfix · 2d ago                                         "
    "    │ │ ├ e5790ed (hotfix) E · 2d ago                                           "
    "    ├─┼─╯ 3c1ac31 Merge feature-1 · 2d ago                                      "
    "    │ ├ e8faeba (feature-1) C · 2d ago                                          "
    "    ├─╯ cc4032c B · 2d ago                                                      "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );
}
