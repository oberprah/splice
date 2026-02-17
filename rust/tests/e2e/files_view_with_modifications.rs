use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn files_view_displays_modified_and_added_files() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}");
    repo.commit("Modify and add files");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    insta::assert_snapshot!(h.snapshot(), @r#"
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
    "#);
}
