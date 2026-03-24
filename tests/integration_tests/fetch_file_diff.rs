use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice::git::fetch_file_diff;

#[test]
#[serial]
fn fetch_file_diff_returns_unified_diff_for_file() {
    reset_counter();
    let repo = TestRepo::new();

    repo.add_file(
        "src/calculator.rs",
        "pub fn add(a: i32, b: i32) -> i32 { a + b }\n",
    );
    repo.commit("Initial commit");

    repo.modify_file(
        "src/calculator.rs",
        "pub fn add(a: i32, b: i32) -> i32 { a + b }\n\npub fn sub(a: i32, b: i32) -> i32 { a - b }\n",
    );
    repo.commit("Add subtraction");

    let hash = repo.rev_parse("HEAD");
    let diff =
        fetch_file_diff(repo.path(), &hash, "src/calculator.rs").expect("diff should be returned");

    assert!(diff.contains("diff --git"));
    assert!(diff.contains("@@"));
    assert!(diff.contains("+pub fn sub"));
}
