use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;
use splice_rust::core::{DiffSource, UncommittedType};
use splice_rust::git;

#[test]
#[serial]
fn diff_command_single_commit() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.add_file("README.md", "# Test\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.add_file("src/new.rs", "pub fn new() {}\n");
    repo.commit("Modify and add files");

    let hash = repo.rev_parse("HEAD");
    let source = git::resolve_commit_range(repo.path(), &hash).unwrap();
    let source = DiffSource::CommitRange(source);

    let mut h = Harness::with_diff_source(&repo, source).unwrap();

    let snapshot = h.snapshot();
    assert!(snapshot.contains("e2af8ce Modify and add files"));
    assert!(snapshot.contains("3 files"));
    assert!(snapshot.contains("M +3 -1  main.rs"));
    assert!(snapshot.contains("A +1 -0  new.rs"));
}

#[test]
#[serial]
fn diff_command_commit_range() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.commit("Second commit");
    repo.add_file("src/new.rs", "pub fn new() {}\n");
    repo.commit("Third commit");

    let source = git::resolve_commit_range(repo.path(), "HEAD~1").unwrap();
    let source = DiffSource::CommitRange(source);

    let mut h = Harness::with_diff_source(&repo, source).unwrap();

    let snapshot = h.snapshot();
    assert!(snapshot.contains("Third commit"));
    assert!(snapshot.contains("2 files"));
    assert!(snapshot.contains("A +1 -0  new.rs"));
}

#[test]
#[serial]
fn diff_command_unstaged_changes() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");

    std::fs::write(
        repo.path().join("src/main.rs"),
        "fn main() {\n    println!(\"hello\");\n}\n",
    )
    .unwrap();

    let source = DiffSource::Uncommitted(UncommittedType::Unstaged);
    let mut h = Harness::with_diff_source(&repo, source).unwrap();

    let snapshot = h.snapshot();
    assert!(snapshot.contains("Unstaged changes"));
    assert!(snapshot.contains("1 files"));
    assert!(snapshot.contains("M +3 -1  main.rs"));
}

#[test]
#[serial]
fn diff_command_staged_changes() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");

    std::fs::write(
        repo.path().join("src/main.rs"),
        "fn main() {\n    println!(\"hello\");\n}\n",
    )
    .unwrap();
    repo.stage_file("src/main.rs");

    let source = DiffSource::Uncommitted(UncommittedType::Staged);
    let mut h = Harness::with_diff_source(&repo, source).unwrap();

    let snapshot = h.snapshot();
    assert!(snapshot.contains("Staged changes"));
    assert!(snapshot.contains("1 files"));
    assert!(snapshot.contains("M +3 -1  main.rs"));
}

#[test]
#[serial]
fn diff_command_invalid_ref_error() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");

    let result = git::resolve_commit_range(repo.path(), "nonexistent_ref");
    assert!(result.is_err());
    assert!(result.unwrap_err().contains("error resolving"));
}

#[test]
#[serial]
fn diff_command_empty_diff_returns_empty_files() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");

    let files =
        git::fetch_uncommitted_file_changes(repo.path(), UncommittedType::Unstaged).unwrap();
    assert!(files.is_empty());
}

#[test]
#[serial]
fn diff_command_quits_to_exit() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/main.rs", "fn main() {}\n");
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", "fn main() {\n    println!(\"hello\");\n}\n");
    repo.commit("Modify");

    let hash = repo.rev_parse("HEAD");
    let source = git::resolve_commit_range(repo.path(), &hash).unwrap();
    let source = DiffSource::CommitRange(source);

    let mut h = Harness::with_diff_source(&repo, source).unwrap();

    h.press(KeyCode::Char('q'));
    assert!(h.should_exit());
}
