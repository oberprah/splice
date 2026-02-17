use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn files_view_navigation() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}");
    repo.add_file("src/lib.rs", "pub fn lib() {}");
    repo.add_file("README.md", "# Test");
    repo.commit("Add initial files");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    insta::assert_snapshot!(h.snapshot(), @r#"
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
    "#);

    h.press(KeyCode::Char('j'));
    insta::assert_snapshot!(h.snapshot(), @r#"
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
    "#);

    h.press(KeyCode::Char('q'));
    insta::assert_snapshot!(h.snapshot(), @r#"
    "  → ├ ec332cd (main) Add initial files                                          "
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
    "#);
}
