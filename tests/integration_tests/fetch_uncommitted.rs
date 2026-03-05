use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice_rust::core::UncommittedType;
use splice_rust::git::fetch_uncommitted_file_changes;

#[test]
#[serial]
fn fetch_uncommitted_returns_empty_when_no_changes() {
    reset_counter();
    let repo = TestRepo::new();
    repo.commit("Initial commit");

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::All).unwrap();
    assert!(changes.is_empty());
}

#[test]
#[serial]
fn fetch_uncommitted_unstaged_returns_modified_file() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("test.txt", "initial content");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("test.txt"), "modified content").unwrap();

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::Unstaged).unwrap();

    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].path, "test.txt");
}

#[test]
#[serial]
fn fetch_uncommitted_staged_returns_staged_file() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("test.txt", "initial content");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("test.txt"), "modified content").unwrap();
    repo.stage_file("test.txt");

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::Staged).unwrap();

    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].path, "test.txt");
}

#[test]
#[serial]
fn fetch_uncommitted_all_returns_both_staged_and_unstaged() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("staged.txt", "initial");
    repo.add_file("unstaged.txt", "initial");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("staged.txt"), "modified staged").unwrap();
    repo.stage_file("staged.txt");

    std::fs::write(repo.path().join("unstaged.txt"), "modified unstaged").unwrap();

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::All).unwrap();

    assert_eq!(changes.len(), 2);
    let paths: Vec<&str> = changes.iter().map(|c| c.path.as_str()).collect();
    assert!(paths.contains(&"staged.txt"));
    assert!(paths.contains(&"unstaged.txt"));
}

#[test]
#[serial]
fn fetch_uncommitted_staged_excludes_unstaged_changes() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("test.txt", "initial content");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("test.txt"), "unstaged modification").unwrap();

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::Staged).unwrap();

    assert!(changes.is_empty());
}

#[test]
#[serial]
fn fetch_uncommitted_unstaged_excludes_staged_changes() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("test.txt", "initial content");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("test.txt"), "staged modification").unwrap();
    repo.stage_file("test.txt");
    std::fs::write(repo.path().join("test.txt"), "additional unstaged").unwrap();

    let unstaged_changes =
        fetch_uncommitted_file_changes(repo.path(), UncommittedType::Unstaged).unwrap();
    let staged_changes =
        fetch_uncommitted_file_changes(repo.path(), UncommittedType::Staged).unwrap();

    assert_eq!(unstaged_changes.len(), 1);
    assert_eq!(staged_changes.len(), 1);
}

#[test]
#[serial]
fn fetch_uncommitted_unstaged_file_name_with_arrow_keeps_stats() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("docs/a => b.md", "line1\n");
    repo.commit("Initial commit");

    std::fs::write(repo.path().join("docs/a => b.md"), "line1\nline2\nline3\n").unwrap();

    let changes = fetch_uncommitted_file_changes(repo.path(), UncommittedType::Unstaged).unwrap();

    assert_eq!(changes.len(), 1);
    assert_eq!(changes[0].path, "docs/a => b.md");
    assert_eq!(changes[0].additions, 2);
    assert_eq!(changes[0].deletions, 0);
}
