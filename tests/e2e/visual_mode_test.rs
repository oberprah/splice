use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn visual_mode() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("src/a.txt", "a\n");
    repo.commit("First");
    repo.add_file("src/b.txt", "b\n");
    repo.commit("Second");
    repo.add_file("src/c.txt", "c\n");
    repo.commit("Third");
    repo.add_file("src/d.txt", "d\n");
    repo.commit("Fourth");

    let mut h = Harness::with_repo(&repo);

    let log = h.snapshot();
    assert!(log.contains("→"), "should show normal cursor");

    h.press(KeyCode::Char('v'));
    let visual = h.snapshot();
    assert!(
        visual.contains("█"),
        "should show visual cursor after pressing v"
    );

    h.press(KeyCode::Char('j'));
    let selected = h.snapshot();
    assert!(selected.contains("▌"), "should show selected marker");
    assert!(selected.contains("█"), "should still show visual cursor");

    h.press(KeyCode::Char('q'));
    let exited = h.snapshot();
    assert!(exited.contains("→"), "should show normal cursor after exit");
    assert!(
        !exited.contains("█"),
        "should not show visual cursor after exit"
    );

    h.press(KeyCode::Char('v'));
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);
    let files = h.snapshot();
    assert!(
        files.contains("(2 commits)"),
        "should show commit range: {}",
        files
    );
}
