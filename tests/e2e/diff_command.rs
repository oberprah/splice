use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;
use splice_rust::core::{DiffRef, UncommittedType};
use splice_rust::git;

#[test]
#[serial]
fn diff_command_commit_range_and_quit() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.add_file("README.md", "# Test\n");
    repo.commit("Initial commit");
    repo.add_file("src/extra.rs", "pub fn extra() {}\n");
    repo.commit("Second commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}\n");
    repo.commit("Third commit");
    repo.modify_file(
        "src/main.rs",
        "fn main() {\n    println!(\"hello again\");\n}\n",
    );
    repo.commit("Fourth commit");

    let range = git::resolve_commit_range(repo.path(), "HEAD~2..HEAD").unwrap();
    let range = DiffRef::CommitRange(range);
    let mut h = Harness::with_diff_source_and_screen_size(&repo, range, 80, 14).unwrap();

    h.assert_snapshot(
        r#"
"  a19fbff..3f5a73d (2 commits)                                                  "
"                                                                                "
"  6 files · +8 -1                                                               "
"   ├── src/                                                                     "
"  →│   ├── A +1 -0  extra.rs                                                    "
"   │   ├── M +3 -1  main.rs                                                     "
"   │   └── A +1 -0  new.rs                                                      "
"   ├── A +1 -0  file_1.txt                                                      "
"   ├── A +1 -0  file_2.txt                                                      "
"   └── A +1 -0  file_3.txt                                                      "
"                                                                                "
"                                                                                "
"                                                                                "
"  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
"#,
    );

    h.press(KeyCode::Char('q'));
    assert!(h.should_exit());
}

#[test]
#[serial]
fn diff_command_uncommitted_views() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");

    std::fs::write(
        repo.path().join("src/main.rs"),
        "fn main() {\n    println!(\"hello\");\n}\n",
    )
    .unwrap();

    let unstaged = DiffRef::Uncommitted(UncommittedType::Unstaged);
    let mut unstaged_h =
        Harness::with_diff_source_and_screen_size(&repo, unstaged, 80, 14).unwrap();

    unstaged_h.assert_snapshot(
        r#"
"  Unstaged changes                                                              "
"                                                                                "
"  1 files · +3 -1                                                               "
"   └── src/                                                                     "
"  →    └── M +3 -1  main.rs                                                     "
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

    unstaged_h.press(KeyCode::Enter);

    unstaged_h.assert_snapshot(
        r#"
"  Unstaged changes · src/main.rs · +3 -1                                        "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    repo.stage_file("src/main.rs");

    let staged = DiffRef::Uncommitted(UncommittedType::Staged);
    let mut staged_h = Harness::with_diff_source_and_screen_size(&repo, staged, 80, 14).unwrap();

    staged_h.assert_snapshot(
        r#"
"  Staged changes                                                                "
"                                                                                "
"  1 files · +3 -1                                                               "
"   └── src/                                                                     "
"  →    └── M +3 -1  main.rs                                                     "
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
