use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;
use splice_rust::git;

fn short_body() -> &'static str {
    "This commit adds the feature.\nSee ticket #42 for details."
}

fn long_body() -> String {
    (1..=12)
        .map(|i| format!("Body line {}.", i))
        .collect::<Vec<_>>()
        .join("\n")
}

#[test]
#[serial]
fn files_view_shows_short_body_always() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() { println!(\"hi\"); }\n");
    repo.commit_with_body("Add greeting", short_body());

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  10a104c Add greeting                                                          "
    "                                                                                "
    "    This commit adds the feature.                                               "
    "    See ticket #42 for details.                                                 "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── src/                                                                     "
    "   │   └── M +1 -1  main.rs                                                     "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );
}

#[test]
#[serial]
fn files_view_no_body_layout_unchanged() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("README.md", "# hello\n");
    repo.commit("Initial commit");
    repo.modify_file("README.md", "# hi\n");
    repo.commit("Update README");

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  6765376 Update README                                                         "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── M +1 -1  README.md                                                       "
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
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );
}

#[test]
#[serial]
fn files_view_long_body_truncated_with_expand_hint() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("lib.rs", "// lib\n");
    repo.commit("Initial commit");
    repo.modify_file("lib.rs", "// updated lib\n");
    repo.commit_with_body("Big refactor", &long_body());

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
    "  9902ee0 Big refactor                                                          "
    "                                                                                "
    "    Body line 1.                                                                "
    "    Body line 2.                                                                "
    "    ↓ 10 more lines  (m: expand)                                                "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── A +1 -0  file_1.txt                                                      "
    "   └── M +1 -1  lib.rs                                                          "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  m: expand  q: back        "
    "#,
    );
}

#[test]
#[serial]
fn files_view_m_expands_and_collapses_long_body() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("lib.rs", "// lib\n");
    repo.commit("Initial commit");
    repo.modify_file("lib.rs", "// updated lib\n");
    repo.commit_with_body("Big refactor", &long_body());

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);

    h.press(KeyCode::Char('m'));
    h.assert_snapshot(
        r#"
    "  9902ee0 Big refactor                                                          "
    "                                                                                "
    "    Body line 1.                                                                "
    "    Body line 2.                                                                "
    "    Body line 3.                                                                "
    "    Body line 4.                                                                "
    "    Body line 5.                                                                "
    "    Body line 6.                                                                "
    "    Body line 7.                                                                "
    "    Body line 8.                                                                "
    "    Body line 9.                                                                "
    "    Body line 10.                                                               "
    "    Body line 11.                                                               "
    "    Body line 12.                                                               "
    "    ↑ show less (m: collapse)                                                   "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── A +1 -0  file_1.txt                                                      "
    "   └── M +1 -1  lib.rs                                                          "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  m: collapse  q: back      "
    "#,
    );

    h.press(KeyCode::Char('m'));
    h.assert_snapshot(
        r#"
    "  9902ee0 Big refactor                                                          "
    "                                                                                "
    "    Body line 1.                                                                "
    "    Body line 2.                                                                "
    "    ↓ 10 more lines  (m: expand)                                                "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── A +1 -0  file_1.txt                                                      "
    "   └── M +1 -1  lib.rs                                                          "
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
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  m: expand  q: back        "
    "#,
    );
}

#[test]
#[serial]
fn files_view_m_does_nothing_when_no_body() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("a.txt", "a\n");
    repo.commit("No body commit");

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);
    let before = h.snapshot();
    h.press(KeyCode::Char('m'));
    let after = h.snapshot();
    assert_eq!(before, after, "m key should have no effect when no body");
}

#[test]
#[serial]
fn files_view_m_does_nothing_for_short_body() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() { println!(\"hi\"); }\n");
    repo.commit_with_body("Add greeting", "Line one.\nLine two.");

    let mut h = Harness::with_repo(&repo);
    h.press(KeyCode::Enter);
    let before = h.snapshot();
    h.press(KeyCode::Char('m'));
    let after = h.snapshot();
    assert_eq!(before, after, "m key should have no effect for short body");
}

#[test]
#[serial]
fn files_view_no_body_panel_for_multi_commit_range() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() { println!(\"hi\"); }\n");
    repo.commit_with_body(
        "Add greeting",
        "This is a body.\nLine two.\nLine three.\nLine four.",
    );

    // Resolve the single-commit range and reconstruct it with count=2 to simulate
    // a multi-commit range (e.g. the kind produced when navigating a merge row).
    // The body panel must not appear for multi-commit ranges.
    let single_range = git::resolve_commit_range(repo.path(), "HEAD~1..HEAD").unwrap();
    let multi_source = splice_rust::DiffRef::CommitRange(splice_rust::core::CommitRange {
        start: single_range.start.clone(),
        end: single_range.end.clone(),
        count: 2,
        include_start: false,
    });

    let mut h = Harness::with_diff_source(&repo, multi_source).unwrap();
    h.assert_snapshot(
        r#"
    "  62ab883..e133e1f (2 commits)                                                  "
    "                                                                                "
    "  2 files · +2 -1                                                               "
    "  →├── src/                                                                     "
    "   │   └── M +1 -1  main.rs                                                     "
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
    "                                                                                "
    "  j/k: navigate  Enter/space: toggle/open  ←/→: fold  q: back                   "
    "#,
    );
}
