use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

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
    "  → ├ e2af8ce (main) Modify and add files                                       "
    "    ├ c500da6 Initial commit                                                    "
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
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );

    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  e2af8ce · Test committed 6 years ago                                          "
    "                                                                                "
    "  Modify and add files                                                          "
    "                                                                                "
    "  3 files · +5 -1                                                               "
    "  → A +1 -0  file_1.txt                                                         "
    "    M +3 -1  src/main.rs                                                        "
    "    A +1 -0  src/new.rs                                                         "
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
    "  j/k: navigate  Enter: open diff  q: back                                      "
    "#,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
    "  e2af8ce · Test committed 6 years ago                                          "
    "                                                                                "
    "  Modify and add files                                                          "
    "                                                                                "
    "  3 files · +5 -1                                                               "
    "    A +1 -0  file_1.txt                                                         "
    "  → M +3 -1  src/main.rs                                                        "
    "    A +1 -0  src/new.rs                                                         "
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
    "  j/k: navigate  Enter: open diff  q: back                                      "
    "#,
    );

    h.press(KeyCode::Char('q'));
    h.assert_snapshot(
        r#"
    "  → ├ e2af8ce (main) Modify and add files                                       "
    "    ├ c500da6 Initial commit                                                    "
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
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "#,
    );
}
