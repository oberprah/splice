use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn visual_mode() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/a.txt", "a\n");
    repo.commit("First");
    repo.add_file("src/b.txt", "b\n");
    repo.commit("Second");
    repo.add_file("src/c.txt", "c\n");
    repo.commit("Third");
    repo.add_file("src/d.txt", "d\n");
    repo.commit("Fourth");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 18);

    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ 2ff7a36 (main) Fourth                                                     "
    "    ├ 57e06a3 Third                                                             "
    "    ├ 8341314 Second                                                            "
    "    ├ e253ff5 First                                                             "
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
    "#,
    );

    h.press(KeyCode::Char('v'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  █ ├ 2ff7a36 (main) Fourth                                                     "
    "    ├ 57e06a3 Third                                                             "
    "    ├ 8341314 Second                                                            "
    "    ├ e253ff5 First                                                             "
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
    "#,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  ▌ ├ 2ff7a36 (main) Fourth                                                     "
    "  █ ├ 57e06a3 Third                                                             "
    "    ├ 8341314 Second                                                            "
    "    ├ e253ff5 First                                                             "
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
    "#,
    );

    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  57e06a3..2ff7a36 (2 commits)                                                  "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "  →├── src/                                                                     "
    "   │   ├── A +1 -0  c.txt                                                       "
    "   │   └── A +1 -0  d.txt                                                       "
    "   ├── A +1 -0  file_2.txt                                                      "
    "   └── A +1 -0  file_3.txt                                                      "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
    "#,
    );

    h.press(KeyCode::Char('q'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  ▌ ├ 2ff7a36 (main) Fourth                                                     "
    "  █ ├ 57e06a3 Third                                                             "
    "    ├ 8341314 Second                                                            "
    "    ├ e253ff5 First                                                             "
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
    "#,
    );

    h.press(KeyCode::Char('q'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "    ├ 2ff7a36 (main) Fourth                                                     "
    "  → ├ 57e06a3 Third                                                             "
    "    ├ 8341314 Second                                                            "
    "    ├ e253ff5 First                                                             "
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
    "#,
    );
}
