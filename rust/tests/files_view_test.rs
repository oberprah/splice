mod common;

use common::{reset_counter, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn test_files_view_navigation() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}");
    repo.add_file("src/lib.rs", "pub fn lib() {}");
    repo.add_file("README.md", "# Test");
    repo.commit("Add initial files");

    let mut h = common::Harness::with_repo(&repo);

    // Press Enter to open files view
    h.press(KeyCode::Enter);
    insta::assert_snapshot!(h.snapshot(), @r###"
    "  ec332cd · Test committed 6 years ago                                          "
    "                                                                                "
    "  Add initial files                                                             "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "  → A +1 -0  README.md                                                          "
    "    A +1 -0  file_0.txt                                                         "
    "    A +1 -0  src/lib.rs                                                         "
    "    A +1 -0  src/main.rs                                                        "
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
    "###);

    // Navigate down
    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "  ec332cd · Test committed 6 years ago                                          "
    "                                                                                "
    "  Add initial files                                                             "
    "                                                                                "
    "  4 files · +4 -0                                                               "
    "    A +1 -0  README.md                                                          "
    "  → A +1 -0  file_0.txt                                                         "
    "    A +1 -0  src/lib.rs                                                         "
    "    A +1 -0  src/main.rs                                                        "
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
    "###);

    // Press q to go back to log view
    h.press(KeyCode::Char('q'));
    insta::assert_snapshot!(h.snapshot(), @r###"
    "  → ec332cd (main) Add initial files                                            "
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
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}

#[test]
#[serial]
fn test_files_view_with_modifications() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}");
    repo.commit("Modify and add files");

    let mut h = common::Harness::with_repo(&repo);

    // Press Enter to open files view
    h.press(KeyCode::Enter);
    insta::assert_snapshot!(h.snapshot(), @r###"
    "  ae67c8c · Test committed 6 years ago                                          "
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
    "###);
}
