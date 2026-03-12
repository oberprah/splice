use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn ctrl_u_scroll_up_with_uncommitted_changes_resets_scroll_offset() {
    reset_counter();

    let repo = TestRepo::new();
    // Create 20 commits to fill more than one screen
    for i in 0..20 {
        repo.commit(&format!("Commit {}", i));
    }
    // Add uncommitted changes
    std::fs::write(repo.path().join("file_0.txt"), "modified").unwrap();

    // Small screen: viewport_height = 10, page_step = 5
    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 12);

    // Scroll down 3 times to accumulate scroll_offset
    h.press_ctrl(KeyCode::Char('d'));
    h.press_ctrl(KeyCode::Char('d'));
    h.press_ctrl(KeyCode::Char('d'));

    // Scroll back up 3 times — should land back on pos=0 (uncommitted row)
    h.press_ctrl(KeyCode::Char('u'));
    h.press_ctrl(KeyCode::Char('u'));
    h.press_ctrl(KeyCode::Char('u'));

    // BUG: without fix, scroll_offset stays > 0 and commits start from Commit 15
    // instead of Commit 19. The uncommitted row is selected but the list is scrolled.
    h.assert_snapshot(
        r#"
    "  →   Unstaged changes · 1 file                                                 "
    "    ├ 15df8f3 (main) Commit 19 · 2d ago                                         "
    "    ├ ca24928 Commit 18 · 2d ago                                                "
    "    ├ 50c9441 Commit 17 · 2d ago                                                "
    "    ├ b37126e Commit 16 · 2d ago                                                "
    "    ├ 3a1959a Commit 15 · 2d ago                                                "
    "    ├ 77f5868 Commit 14 · 2d ago                                                "
    "    ├ 9593ccc Commit 13 · 2d ago                                                "
    "    ├ 0198648 Commit 12 · 2d ago                                                "
    "    ├ d4ec593 Commit 11 · 2d ago                                                "
    "    ├ b33b874 Commit 10 · 2d ago                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
        "#,
    );
}

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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );
}
