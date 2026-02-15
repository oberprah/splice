mod common;

use common::TestRepo;
use splice_rust::git::fetch_commits;
use std::path::PathBuf;

#[test]
fn test_fetch_commits_empty_repo() {
    let repo = TestRepo::new();
    let result = fetch_commits(repo.path());
    assert!(result.is_err());
}

#[test]
fn test_fetch_commits_single_commit() {
    let repo = TestRepo::new().commit("Initial commit");
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    assert_eq!(commits[0].message, "Initial commit");
    assert!(!commits[0].hash.is_empty());
    assert!(commits[0].parent_hashes.is_empty());
}

#[test]
fn test_fetch_commits_multiple_commits() {
    let repo = TestRepo::new()
        .commit("First commit")
        .commit("Second commit")
        .commit("Third commit");
    
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 3);
    assert_eq!(commits[0].message, "Third commit");
    assert_eq!(commits[1].message, "Second commit");
    assert_eq!(commits[2].message, "First commit");
}

#[test]
fn test_fetch_commits_has_parent_hashes() {
    let repo = TestRepo::new()
        .commit("First")
        .commit("Second");
    
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 2);
    assert!(commits[0].parent_hashes.len() >= 1);
    assert_eq!(commits[0].parent_hashes[0], commits[1].hash);
}

#[test]
fn test_fetch_commits_with_branch() {
    let repo = TestRepo::new()
        .commit("Initial commit")
        .create_branch("feature");
    
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    let branch_refs: Vec<_> = commits[0].refs.iter().filter(|r| r.name == "feature").collect();
    assert_eq!(branch_refs.len(), 1);
}

#[test]
fn test_fetch_commits_with_tag() {
    let repo = TestRepo::new()
        .commit("Initial commit")
        .create_tag("v1.0.0");
    
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    let tag_refs: Vec<_> = commits[0].refs.iter().filter(|r| r.name == "v1.0.0").collect();
    assert_eq!(tag_refs.len(), 1);
    assert!(tag_refs.iter().all(|r| r.ref_type == splice_rust::core::RefType::Tag));
}

#[test]
fn test_fetch_commits_head_branch() {
    let repo = TestRepo::new().commit("Initial commit");
    
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    let head_refs: Vec<_> = commits[0].refs.iter().filter(|r| r.is_head).collect();
    assert_eq!(head_refs.len(), 1);
}

#[test]
fn test_fetch_commits_invalid_path() {
    let invalid_path = PathBuf::from("/nonexistent/path/that/does/not/exist");
    let result = fetch_commits(&invalid_path);
    
    assert!(result.is_err());
}

#[test]
fn test_fetch_commits_not_a_git_repo() {
    let temp_dir = tempfile::TempDir::new().unwrap();
    let result = fetch_commits(temp_dir.path());
    
    assert!(result.is_err());
}

#[test]
fn test_fetch_commits_author_info() {
    let repo = TestRepo::new().commit("Test commit");
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    assert!(!commits[0].author.is_empty());
}

#[test]
fn test_fetch_commits_hash_format() {
    let repo = TestRepo::new().commit("Test commit");
    let commits = fetch_commits(repo.path()).unwrap();
    
    assert_eq!(commits.len(), 1);
    assert_eq!(commits[0].hash.len(), 40);
    assert!(commits[0].hash.chars().all(|c| c.is_ascii_hexdigit()));
}
