use crate::common::{reset_counter, Harness, TestRepo};
use serial_test::serial;

#[test]
#[serial]
fn complex_graph() {
    reset_counter();

    let repo = TestRepo::new();

    // Build the graph according to the test case from go/internal/domain/graph/layout_test.go
    // Expected structure:
    // M ← J, L
    // L ← K
    // K ← D
    // J ← H, I
    // I ← G
    // H ← F, G
    // G ← F
    // F ← D, E
    // E ← D
    // D ← B, C
    // C ← B
    // B ← A
    // A ← (root)

    // 1. A: root commit on main
    repo.commit("A");

    // 2. B: next commit on main
    repo.commit("B");

    // 3. C: branch feature-1 from B
    repo.create_branch("feature-1");
    repo.checkout("feature-1");
    repo.commit("C");

    // 4. D: merge feature-1 into main
    repo.checkout("main");
    repo.merge("feature-1"); // D is the merge commit
    let d_hash = repo.rev_parse("HEAD");

    // 5. E: branch hotfix from D
    repo.create_branch("hotfix");
    repo.checkout("hotfix");
    repo.commit("E");

    // 6. F: merge hotfix into main
    repo.checkout("main");
    repo.merge("hotfix"); // F is the merge commit

    // 7. G: branch feature-2 from F
    repo.create_branch("feature-2");
    repo.checkout("feature-2");
    repo.commit("G");

    // Save G's hash for later branching (I needs to branch from G after G is merged)
    let g_hash = repo.rev_parse("HEAD");

    // 8. H: merge feature-2 into main (G is now merged)
    repo.checkout("main");
    repo.merge("feature-2"); // H is the merge commit

    // 9. I: branch from G (which is already merged into H)
    // This creates a commit that points back to G, but after H exists
    repo.checkout_hash(&g_hash);
    repo.create_branch("feature-2b");
    repo.checkout("feature-2b");
    repo.commit("I");

    // 10. J: merge feature-2b into main
    repo.checkout("main");
    repo.merge("feature-2b"); // J is the merge commit

    // 11. K: branch feature-3 from D (late branching from earlier commit)
    repo.checkout_hash(&d_hash);
    repo.create_branch("feature-3");
    repo.checkout("feature-3");
    repo.commit("K");

    // 12. L: continue on feature-3
    repo.commit("L");

    // 13. M: merge feature-3 into main
    repo.checkout("main");
    repo.merge("feature-3"); // M is the merge commit

    let mut h = Harness::with_repo(&repo);

    // Expected graph (from go/internal/domain/graph/layout_test.go):
    // ├─╮       M: merge commit
    // │ ├       L: feature-3
    // │ ├       K: feature-3
    // ├─│─╮     J: merge commit
    // │ │ ├     I: feature-2b
    // ├─│─│─╮   H: merge commit
    // │ │ ├─╯   G: feature-2 converges
    // ├─│─┤     F: merge commit (merge join)
    // │ │ ├     E: hotfix
    // ├─┼─╯     D: merge commit (cross + convergence)
    // │ ├       C: feature-1
    // ├─╯       B: convergence
    // ├         A: root

    insta::assert_snapshot!(h.snapshot(), @r###"
    "  → ├─╮ 45fe0e3 (main) Merge feature-3                                          "
    "    │ ├ 44eee0d (feature-3) L                                                   "
    "    │ ├ 4c75a9d K                                                               "
    "    ├─│─╮ 55fbe6d Merge feature-2b                                              "
    "    │ │ ├ 02f5e0a (feature-2b) I                                                "
    "    ├─│─│─╮ 6eab1a0 Merge feature-2                                             "
    "    │ │ ├─╯ 6385fe0 (feature-2) G                                               "
    "    ├─│─┤ 16b8252 Merge hotfix                                                  "
    "    │ │ ├ e5790ed (hotfix) E                                                    "
    "    ├─┼─╯ 3c1ac31 Merge feature-1                                               "
    "    │ ├ e8faeba (feature-1) C                                                   "
    "    ├─╯ cc4032c B                                                               "
    "    ├ fe76018 A                                                                 "
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
    "  j/k: navigate  Ctrl+d/u: half-page  q: quit                                   "
    "###);
}
