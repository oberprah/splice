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
    "  e2af8ce Modify and add files                                                  "
    "                                                                                "
    "  3 files · +5 -1                                                               "
    "  →├── src/                                                                     "
    "   │   ├── M +3 -1  main.rs                                                     "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
    "#,
    );

    h.press(KeyCode::Char('j'));
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
    "#,
    );

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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: collapse/expand  q: back        "
    "#,
    );
}
