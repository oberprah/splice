use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn files_view_scroll_long_file_list() {
    reset_counter();

    let repo = TestRepo::new();
    // Add 20 files so that together with the auto-added file_0.txt the total (21)
    // exceeds the visible list area of 20 rows.
    for i in 1..=20 {
        repo.add_file(&format!("b_{:02}.txt", i), "x\n");
    }
    repo.commit("Add many files");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    // Trigger a render so the viewport height is initialised before navigating.
    h.snapshot();

    // Navigate to the last file (index 20, file_0.txt).
    for _ in 0..20 {
        h.press(KeyCode::Char('j'));
    }

    // With the correct scroll behaviour the view must shift by one row so that
    // file_0.txt (index 20) is visible with the → cursor on the last list line.
    // The buggy implementation keeps scroll_offset at 0 (viewport_height is set
    // to the full area height minus 1 rather than the actual list height), so
    // file_0.txt ends up off-screen and no cursor is visible.
    h.assert_snapshot(
        r#"
        "  539d384 Add many files                                                        "
        "                                                                                "
        "  21 files · +21 -0                                                             "
        "   ├── A +1 -0  b_02.txt                                                        "
        "   ├── A +1 -0  b_03.txt                                                        "
        "   ├── A +1 -0  b_04.txt                                                        "
        "   ├── A +1 -0  b_05.txt                                                        "
        "   ├── A +1 -0  b_06.txt                                                        "
        "   ├── A +1 -0  b_07.txt                                                        "
        "   ├── A +1 -0  b_08.txt                                                        "
        "   ├── A +1 -0  b_09.txt                                                        "
        "   ├── A +1 -0  b_10.txt                                                        "
        "   ├── A +1 -0  b_11.txt                                                        "
        "   ├── A +1 -0  b_12.txt                                                        "
        "   ├── A +1 -0  b_13.txt                                                        "
        "   ├── A +1 -0  b_14.txt                                                        "
        "   ├── A +1 -0  b_15.txt                                                        "
        "   ├── A +1 -0  b_16.txt                                                        "
        "   ├── A +1 -0  b_17.txt                                                        "
        "   ├── A +1 -0  b_18.txt                                                        "
        "   ├── A +1 -0  b_19.txt                                                        "
        "   ├── A +1 -0  b_20.txt                                                        "
        "  →└── A +1 -0  file_0.txt                                                      "
        "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
        "#,
    );
}

#[test]
#[serial]
fn files_view_navigation_with_modifications() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.add_file("README.md", "# Test\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}\n");
    repo.commit("Modify and add files");

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ e2af8ce (main) Modify and add files · 2d ago                              "
    "    ├ c500da6 Initial commit · 2d ago                                           "
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

    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  e2af8ce Modify and add files                                                  "
    "                                                                                "
    "  3 files · +5 -1                                                               "
    "   ├── src/                                                                     "
    "  →│   ├── M +3 -1  main.rs                                                     "
    "   │   └── A +1 -0  new.rs                                                      "
    "   └── A +1 -0  file_1.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
    "  e2af8ce Modify and add files                                                  "
    "                                                                                "
    "  3 files · +5 -1                                                               "
    "   ├── src/                                                                     "
    "   │   ├── M +3 -1  main.rs                                                     "
    "  →│   └── A +1 -0  new.rs                                                      "
    "   └── A +1 -0  file_1.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );

    h.press(KeyCode::Char('q'));
    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ e2af8ce (main) Modify and add files · 2d ago                              "
    "    ├ c500da6 Initial commit · 2d ago                                           "
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
fn files_view_folder_collapse_expand() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/components/Button.tsx", "export {}\n");
    repo.add_file("src/components/Input.tsx", "export {}\n");
    repo.add_file("src/utils/helper.ts", "export {}\n");
    repo.commit("Add nested files");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  01e0c9d Add nested files                                                      "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "   ├── src/                                                                     "
    "   │   ├── components/                                                          "
    "  →│   │   ├── A +1 -0  Button.tsx                                              "
    "   │   │   └── A +1 -0  Input.tsx                                               "
    "   │   └── utils/                                                               "
    "   │       └── A +1 -0  helper.ts                                               "
    "   └── A +1 -0  file_0.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );

    // Navigate up to the src/ folder so we can test collapse/expand
    h.press(KeyCode::Char('k'));
    h.press(KeyCode::Char('k'));
    h.press(KeyCode::Left);
    h.assert_snapshot(
        r#"
    "  01e0c9d Add nested files                                                      "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "  →├── src/ +3 -0 (3 files)                                                     "
    "   └── A +1 -0  file_0.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );

    h.press(KeyCode::Right);
    h.assert_snapshot(
        r#"
    "  01e0c9d Add nested files                                                      "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "  →├── src/                                                                     "
    "   │   ├── components/                                                          "
    "   │   │   ├── A +1 -0  Button.tsx                                              "
    "   │   │   └── A +1 -0  Input.tsx                                               "
    "   │   └── utils/                                                               "
    "   │       └── A +1 -0  helper.ts                                               "
    "   └── A +1 -0  file_0.txt                                                      "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );
}
