mod common;

use common::TestRepo;
use crossterm::event::KeyCode;

#[test]
fn test_log_view() {
    let repo = TestRepo::new()
        .commit("Initial commit")
        .create_branch("feature")
        .commit("Add feature")
        .create_tag("v1.0.0")
        .commit("Fix bug");

    let mut h = common::Harness::with_repo(&repo);

    // Initial view - shows commits with refs
    let snapshot = format_snapshot(h.snapshot());
    insta::assert_snapshot!(snapshot, @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  → aa6a3d2 (main) Fix bug                                                      "
    "    8c2a8a1 (tag: v1.0.0) Add feature                                           "
    "    570cb66 (feature) Initial commit                                            "
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
    "  j/k: navigate  g/G: jump  q: quit                                             "
    "                                                                                "
    "                                                                                "
    "###);

    // Navigate down
    h.press(KeyCode::Char('j'));

    let snapshot = format_snapshot(h.snapshot());
    insta::assert_snapshot!(snapshot, @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "    aa6a3d2 (main) Fix bug                                                      "
    "  → 8c2a8a1 (tag: v1.0.0) Add feature                                           "
    "    570cb66 (feature) Initial commit                                            "
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
    "  j/k: navigate  g/G: jump  q: quit                                             "
    "                                                                                "
    "                                                                                "
    "###);

    // Jump to last
    h.press(KeyCode::Char('G'));

    let snapshot = format_snapshot(h.snapshot());
    insta::assert_snapshot!(snapshot, @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "    aa6a3d2 (main) Fix bug                                                      "
    "    8c2a8a1 (tag: v1.0.0) Add feature                                           "
    "  → 570cb66 (feature) Initial commit                                            "
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
    "  j/k: navigate  g/G: jump  q: quit                                             "
    "                                                                                "
    "                                                                                "
    "###);

    // Jump to first
    h.press(KeyCode::Char('g'));

    let snapshot = format_snapshot(h.snapshot());
    insta::assert_snapshot!(snapshot, @r###"
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  → aa6a3d2 (main) Fix bug                                                      "
    "    8c2a8a1 (tag: v1.0.0) Add feature                                           "
    "    570cb66 (feature) Initial commit                                            "
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
    "  j/k: navigate  g/G: jump  q: quit                                             "
    "                                                                                "
    "                                                                                "
    "###);
}

fn format_snapshot(backend: &ratatui::backend::TestBackend) -> String {
    let buffer = backend.buffer();
    let mut lines = Vec::new();
    for y in 0..buffer.area.height {
        let mut line = String::new();
        for x in 0..buffer.area.width {
            let cell = buffer.cell((x, y)).unwrap();
            line.push_str(cell.symbol());
        }
        lines.push(format!("\"{}\"", line));
    }
    lines.join("\n")
}
