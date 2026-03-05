use crate::common::{reset_counter, TestRepo};
use serial_test::serial;
use splice_rust::core::FileStatus;
use splice_rust::git::{fetch_file_changes_for_source, resolve_diff_source, DiffSpec};

#[test]
#[serial]
fn fetch_file_changes_three_dot_excludes_merge_base_changes() {
    reset_counter();
    let repo = TestRepo::new();

    repo.add_file("base.txt", "base");
    repo.commit("Base commit");

    repo.create_branch("feature");
    repo.checkout("feature");
    repo.add_file("feature.txt", "feature");
    repo.commit("Feature commit");

    repo.checkout("main");
    repo.add_file("main.txt", "main");
    repo.commit("Main commit");

    let source = resolve_diff_source(
        repo.path(),
        DiffSpec {
            raw: Some("main...feature".to_string()),
            uncommitted_type: None,
        },
    )
    .unwrap();

    let changes = fetch_file_changes_for_source(repo.path(), &source).unwrap();
    let paths: Vec<&str> = changes.iter().map(|change| change.path.as_str()).collect();

    assert!(paths.contains(&"feature.txt"));
    assert!(!paths.contains(&"base.txt"));
    assert!(!paths.contains(&"main.txt"));
}

#[test]
#[serial]
fn fetch_file_changes_single_commit_tracks_renamed_file_edits_in_moved_folder() {
    reset_counter();
    let repo = TestRepo::new();

    repo.add_file(
        "src/components/Button.tsx",
        "export const Button = {}\nconst X = 1;\nconst Y = 2;\n",
    );
    repo.add_file("src/components/Input.tsx", "export const Input = {}\n");
    repo.commit("Add components");

    std::fs::create_dir_all(repo.path().join("ui")).unwrap();
    repo.move_folder("src/components", "ui/components");
    repo.modify_file(
        "ui/components/Button.tsx",
        "export const Button = {}\nconst X = 1;\nconst Y = 2;\nconst Z = 3;\n",
    );
    repo.commit("Move components and edit button");

    let source = resolve_diff_source(
        repo.path(),
        DiffSpec {
            raw: Some("HEAD".to_string()),
            uncommitted_type: None,
        },
    )
    .unwrap();

    let changes = fetch_file_changes_for_source(repo.path(), &source).unwrap();
    let button = changes
        .iter()
        .find(|c| c.path == "ui/components/Button.tsx")
        .expect("renamed button file should exist");

    assert_eq!(button.status, FileStatus::Renamed);
    assert_eq!(
        button.old_path.as_deref(),
        Some("src/components/Button.tsx")
    );
    assert_eq!(button.additions, 1);
    assert_eq!(button.deletions, 0);
}
