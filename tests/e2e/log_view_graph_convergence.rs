use crate::common::{reset_counter, Harness, TestRepo};
use serial_test::serial;

// Reproduces a graph rendering bug: when a lane converges into a commit across
// columns whose lanes have already ended, the horizontal connector must be drawn
// through the empty columns. Previously the converging ╯ was left floating,
// disconnected from its commit (e.g. "├─╮   ╯" instead of "├─┬───╯").
#[test]
#[serial]
fn log_view_convergence_across_empty_lanes() {
    reset_counter();

    let repo = TestRepo::new();

    // BASE, and DD on a side branch merged back => M11
    repo.commit("Base");
    repo.create_branch("deploy");
    repo.checkout("deploy");
    repo.commit("Deploy fix");
    repo.checkout("main");
    repo.merge("deploy");

    // Long-lived branch: NF (merged much later, so its lane runs far right)
    repo.create_branch("new-font");
    repo.checkout("new-font");
    repo.commit("Replace font");
    repo.checkout("main");

    // X directly on main; two feature branches fork from it
    repo.commit("Remove hover animations");
    repo.create_branch("ai-page");
    repo.create_branch("news-links");
    repo.checkout("ai-page");
    repo.commit("AI page");
    repo.checkout("news-links");
    repo.commit("News links");

    // Merge everything back in sequence: NF first, then the forks of X
    repo.checkout("main");
    repo.merge("new-font");
    repo.merge("ai-page");
    repo.merge("news-links");

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(
        r###"
    "        Working tree clean                                                      "
    "  → ├─╮ cb45873 (main) Merge news-links · Test · 2d ago                         "
    "    │ ├ 93eb83f (news-links) News links · Test · 2d ago                         "
    "    ├─│─╮ 072ed6a Merge ai-page · Test · 2d ago                                 "
    "    │ │ ├ ad417ff (ai-page) AI page · Test · 2d ago                             "
    "    ├─│─│─╮ 365e5f7 Merge new-font · Test · 2d ago                              "
    "    │ │ │ ├ 07a4c46 (new-font) Replace font · Test · 2d ago                     "
    "    ├─┴─╯ │ 2f2355d Remove hover animations · Test · 2d ago                     "
    "    ├─┬───╯ 573bb7b Merge deploy · Test · 2d ago                                "
    "    │ ├ 100d811 (deploy) Deploy fix · Test · 2d ago                             "
    "    ├─╯ e6a693a Base · Test · 2d ago                                            "
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
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );
}
