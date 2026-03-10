use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn log_view_clean_summary_not_selectable() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("note.txt", "one\n");
    repo.commit("Initial commit");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 8);

    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ be4f0b7 (main) Initial commit · 2d ago                                    "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );

    h.press(KeyCode::Char('k'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ be4f0b7 (main) Initial commit · 2d ago                                    "
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
fn log_view_uncommitted_summary_opens_files_view() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("note.txt", "one\n");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("note.txt"), "one\ntwo\n").unwrap();

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);

    h.assert_snapshot(
        r#"
    "  →   Unstaged changes · 1 file                                                 "
    "    ├ be4f0b7 (main) Initial commit · 2d ago                                    "
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
    "  Unstaged changes                                                              "
    "                                                                                "
    "  1 files · +1 -0                                                               "
    "  →└── M +1 -0  note.txt                                                        "
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
