use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn renamed_file_displays_as_renamed() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("old_name.txt", "initial content\n");
    repo.commit("Add file");
    repo.rename_file("old_name.txt", "new_name.txt");
    repo.commit("Rename file");

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(
        r#"
    "      Working tree clean                                                        "
    "  → ├ 70e9905 (main) Rename file · 2d ago                                       "
    "    ├ 62f6d2f Add file · 2d ago                                                 "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
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
    "  70e9905 Rename file                                                           "
    "                                                                                "
    "  2 files · +1 -0                                                               "
    "  →├── A +1 -0  file_1.txt                                                      "
    "   └── R +0 -0  new_name.txt (moved from old_name.txt)                          "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
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

#[test]
#[serial]
fn moved_folder_displays_renamed_files() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/components/Button.tsx", "export const Button = {}\n");
    repo.add_file("src/components/Input.tsx", "export const Input = {}\n");
    repo.commit("Add components");

    std::fs::create_dir_all(repo.path().join("ui")).unwrap();
    repo.move_folder("src/components", "ui/components");
    repo.commit("Move components folder");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  cb6373d Move components folder                                                "
    "                                                                                "
    "  3 files · +1 -0                                                               "
    "  →├── ui/components/                                                           "
    "   │   ├── R +0 -0  Button.tsx (moved from src/components/Button.tsx)           "
    "   │   └── R +0 -0  Input.tsx (moved from src/components/Input.tsx)             "
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
}
