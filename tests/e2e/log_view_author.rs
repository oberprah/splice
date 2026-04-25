use crate::common::{reset_counter, Harness, TestRepo};
use serial_test::serial;

#[test]
#[serial]
fn log_view_renders_long_author_name() {
    reset_counter();

    let repo = TestRepo::new();

    repo.commit_with_author(
        "Refactor rendering pipeline for clarity",
        "Hannes Oberprantacher",
    );
    repo.commit_with_author(
        "Introduce viewport-aware scroll offset logic",
        "Hannes Oberprantacher",
    );
    repo.commit_with_author(
        "Tighten commit hash formatting in log view",
        "Hannes Oberprantacher",
    );

    let mut h = Harness::with_repo(&repo);

    h.assert_snapshot(
        r###"
    "      Working tree clean                                                        "
    "  → ├ ac89f53 (main) Tighten commit hash formatting in log view · Hannes Oberp  "
    "    ├ 11f4b97 Introduce viewport-aware scroll offset logic · Hannes Oberpranta  "
    "    ├ 449727f Refactor rendering pipeline for clarity · Hannes Oberprantacher   "
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
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "                                                                                "
    "  j/k: navigate  Ctrl+d/u: half-page  y: copy hash  q: quit                     "
    "###,
    );
}
