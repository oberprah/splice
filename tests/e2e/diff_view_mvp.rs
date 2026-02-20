use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn diff_view_mvp_side_by_side() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.add_file("README.md", "# Test\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}\n");
    repo.commit("Modify and add files");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
        "  e2af8ce · src/main.rs · +3 -1                                                 "
        "                                                                                "
        "    1 - fn main() {}                  │   1 + fn main() {                       "
        "                                      │   2 +     println!("hello");            "
        "                                      │   3 + }                                 "
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
        "  j/k: scroll  q: back                                                          "
        "#,
    );
}
