use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice::git::repository_root;

#[test]
#[serial]
fn repository_root_resolves_subdirectory_to_repo_root() {
    reset_counter();
    let repo = TestRepo::new();
    repo.add_file("src/foo.rs", "fn main() {}");
    repo.commit("add file");

    let subdir = repo.path().join("src");
    let result = repository_root(&subdir);

    assert!(result.is_ok());
    assert_eq!(
        result.unwrap().canonicalize().unwrap(),
        repo.path().canonicalize().unwrap()
    );
}
