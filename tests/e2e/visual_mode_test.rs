use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn enter_visual_mode() {
    reset_counter();

    let repo = TestRepo::new();
    repo.commit("First commit");
    repo.commit("Second commit");
    repo.commit("Third commit");

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(
        r#"
    "  → ├ f78c7bc (main) Third commit                                               "
    "    ├ 1b3e4a7 Second commit                                                     "
    "    ├ c5800ca First commit                                                      "
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
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );

    h.press(KeyCode::Char('v'));
    h.assert_snapshot(
        r#"
    "  █ ├ f78c7bc (main) Third commit                                               "
    "    ├ 1b3e4a7 Second commit                                                     "
    "    ├ c5800ca First commit                                                      "
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
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );
}

#[test]
#[serial]
fn visual_mode_selection() {
    reset_counter();

    let repo = TestRepo::new();
    repo.commit("First commit");
    repo.commit("Second commit");
    repo.commit("Third commit");
    repo.commit("Fourth commit");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Char('v'));
    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
    "  ▌ ├ 19f3962 (main) Fourth commit                                              "
    "  █ ├ f78c7bc Third commit                                                      "
    "    ├ 1b3e4a7 Second commit                                                     "
    "    ├ c5800ca First commit                                                      "
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
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
    "  ▌ ├ 19f3962 (main) Fourth commit                                              "
    "  ▌ ├ f78c7bc Third commit                                                      "
    "  █ ├ 1b3e4a7 Second commit                                                     "
    "    ├ c5800ca First commit                                                      "
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
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );
}

#[test]
#[serial]
fn exit_visual_mode() {
    reset_counter();

    let repo = TestRepo::new();
    repo.commit("First commit");
    repo.commit("Second commit");
    repo.commit("Third commit");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Char('v'));
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Char('v'));

    h.assert_snapshot(
        r#"
    "    ├ f78c7bc (main) Third commit                                               "
    "    ├ 1b3e4a7 Second commit                                                     "
    "  → ├ c5800ca First commit                                                      "
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
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );
}

#[test]
#[serial]
fn visual_mode_open_files() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("First commit");
    repo.add_file("src/lib.rs", "pub fn lib() {}\n");
    repo.commit("Second commit");
    repo.add_file("src/utils.rs", "pub fn utils() {}\n");
    repo.commit("Third commit");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Char('v'));
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
    "  7de2139..840b74f (2 commits)                                                  "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "  →├── src/                                                                     "
    "   │   ├── A +1 -0  lib.rs                                                      "
    "   │   └── A +1 -0  utils.rs                                                    "
    "   ├── A +1 -0  file_1.txt                                                      "
    "   └── A +1 -0  file_2.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
    "#,
    );
}
