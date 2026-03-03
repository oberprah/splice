use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn diff_view_side_by_side() {
    reset_counter();

    let repo = TestRepo::new();
    let original = r#"pub struct Calculator;

impl Calculator {
    pub fn new() -> Self {
        Self
    }

    pub fn add(&self, a: i32, b: i32) -> i32 {
        a + b
    }

    pub fn sub(&self, a: i32, b: i32) -> i32 {
        a - b
    }

    pub fn multiply(&self, a: i32, b: i32) -> i32 {
        a * b
    }

    pub fn divide(&self, a: i32, b: i32) -> Option<i32> {
        if b == 0 {
            None
        } else {
            Some(a / b)
        }
    }

    pub fn sum(&self, values: &[i32]) -> i32 {
        values.iter().sum()
    }

    pub fn format_result(&self, value: i32) -> String {
        format!("Result: {}", value)
    }
}
"#;

    let updated = r#"pub struct Calculator;

impl Calculator {
    pub fn new() -> Self {
        Self
    }

    pub fn add(&self, a: i32, b: i32) -> i32 {
        a + b
    }

    pub fn sub(&self, a: i32, b: i32) -> i32 {
        a - b
    }

    pub fn mul(&self, a: i32, b: i32) -> i32 {
        a.checked_mul(b).unwrap_or(0)
    }

    pub fn sum(&self, values: &[i32]) -> i32 {
        values.iter().sum()
    }

    pub fn pow(&self, base: i32, exp: u32) -> i32 {
        (0..exp).fold(1, |acc, _| acc * base)
    }

    pub fn format_result(&self, value: i32) -> String {
        format!("Result: {value}")
    }
}
"#;

    repo.add_file("src/calculator.rs", &original);
    repo.add_file("README.md", "# Test\n");
    repo.commit("Initial commit");
    repo.modify_file("src/calculator.rs", &updated);
    repo.commit("Update calculator");

    let mut h = Harness::with_repo(&repo);

    h.press(KeyCode::Enter);
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"    1   pub struct Calculator;        │   1   pub struct Calculator;            "
"    2                                 │   2                                     "
"    3   impl Calculator {             │   3   impl Calculator {                 "
"    4       pub fn new() -> Self {    │   4       pub fn new() -> Self {        "
"    5           Self                  │   5           Self                      "
"    6       }                         │   6       }                             "
"    7                                 │   7                                     "
"    8       pub fn add(&self, a: i32, │   8       pub fn add(&self, a: i32,     "
"    9           a + b                 │   9           a + b                     "
"   10       }                         │  10       }                             "
"   11                                 │  11                                     "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
"    7                                 │   7                                     "
"    8       pub fn add(&self, a: i32, │   8       pub fn add(&self, a: i32,     "
"    9           a + b                 │   9           a + b                     "
"   10       }                         │  10       }                             "
"   11                                 │  11                                     "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"   17 -         a * b                 │  17 +         a.checked_mul(b).unwra    "
"   18 -     }                         │                                         "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a: i │                                         "
"   21 -         if b == 0 {           │                                         "
"   22 -             None              │                                         "
"   23 -         } else {              │                                         "
"   24 -             Some(a / b)       │                                         "
"   25 -         }                     │                                         "
"   26       }                         │  18       }                             "
"   27                                 │  19                                     "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
"   18 -     }                         │                                         "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a: i │                                         "
"   21 -         if b == 0 {           │                                         "
"   22 -             None              │                                         "
"   23 -         } else {              │                                         "
"   24 -             Some(a / b)       │                                         "
"   25 -         }                     │                                         "
"   26       }                         │  18       }                             "
"   27                                 │  19                                     "
"   28       pub fn sum(&self, values: │  20       pub fn sum(&self, values:     "
"   29           values.iter().sum()   │  21           values.iter().sum()       "
"   30       }                         │  22       }                             "
"   31                                 │  23                                     "
"                                      │  24 +     pub fn pow(&self, base: i3    "
"                                      │  25 +         (0..exp).fold(1, |acc,    "
"                                      │  26 +     }                             "
"                                      │  27 +                                   "
"   32       pub fn format_result(&sel │  28       pub fn format_result(&self    "
"   33 -         format!("Result: {}", │  29 +         format!("Result: {valu    "
"   34       }                         │  30       }                             "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press_ctrl(KeyCode::Char('u'));
    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
"    8       pub fn add(&self, a: i32, │   8       pub fn add(&self, a: i32,     "
"    9           a + b                 │   9           a + b                     "
"   10       }                         │  10       }                             "
"   11                                 │  11                                     "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"   17 -         a * b                 │  17 +         a.checked_mul(b).unwra    "
"   18 -     }                         │                                         "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a: i │                                         "
"   21 -         if b == 0 {           │                                         "
"   22 -             None              │                                         "
"   23 -         } else {              │                                         "
"   24 -             Some(a / b)       │                                         "
"   25 -         }                     │                                         "
"   26       }                         │  18       }                             "
"   27                                 │  19                                     "
"   28       pub fn sum(&self, values: │  20       pub fn sum(&self, values:     "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_separator_with_wide_glyphs() {
    reset_counter();

    let repo = TestRepo::new();
    let original = r#"impl LineDisplayState {
    fn prefix(&self) -> &'static str {
        match self {
            LineDisplayState::Cursor => "→ ",
            LineDisplayState::Selected => "▌ ",
        }
    }
}
"#;
    let updated = r#"impl LineDisplayState {
    fn prefix(&self) -> &'static str {
        match self {
            LineDisplayState::None => "  ",
            LineDisplayState::Cursor => "→ ",
            LineDisplayState::Selected => "▌ ",
        }
    }
}
"#;

    repo.add_file("src/ui/log.rs", original);
    repo.commit("Add line display state");
    repo.modify_file("src/ui/log.rs", updated);
    repo.commit("Add none state");

    let mut h = Harness::with_repo_and_screen_size(&repo, 120, 10);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Right);
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  b0799a4 · src/ui/log.rs · +1 -0                                                                                       "
"                                                                                                                        "
"                                                          │                                                             "
"                                                          │                                                             "
"    1   impl LineDisplayState {                           │   1   impl LineDisplayState {                               "
"    2       fn prefix(&self) -> &'static str {            │   2       fn prefix(&self) -> &'static str {                "
"    3           match self {                              │   3           match self {                                  "
"                                                          │   4 +             LineDisplayState::None => "  ",           "
"    4               LineDisplayState::Cursor => "→ ",     │   5               LineDisplayState::Cursor => "→ ",         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                                                                    "
"#,
    );
}

#[test]
#[serial]
fn diff_view_navigates_hunks_and_jumps_between_files() {
    reset_counter();

    let repo = TestRepo::new();
    let old_large = (1..=20)
        .map(|n| format!("old line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\n";
    let new_large = (1..=20)
        .map(|n| format!("new line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\n";

    repo.add_file("src/large.txt", &old_large);
    repo.add_file("notes.md", "alpha\nbeta\n");
    repo.commit("Add diff fixtures");

    repo.modify_file("src/large.txt", &new_large);
    repo.modify_file("notes.md", "alpha\ngamma\n");
    repo.commit("Update fixtures");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"                                                                                "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"    1 - old line 1                    │   1 + new line 1                        "
"    2 - old line 2                    │   2 + new line 2                        "
"    3 - old line 3                    │   3 + new line 3                        "
"    4 - old line 4                    │   4 + new line 4                        "
"    5 - old line 5                    │   5 + new line 5                        "
"    6 - old line 6                    │   6 + new line 6                        "
"    7 - old line 7                    │   7 + new line 7                        "
"    8 - old line 8                    │   8 + new line 8                        "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"                                                                                "
"    4 - old line 4                    │   4 + new line 4                        "
"    5 - old line 5                    │   5 + new line 5                        "
"    6 - old line 6                    │   6 + new line 6                        "
"    7 - old line 7                    │   7 + new line 7                        "
"    8 - old line 8                    │   8 + new line 8                        "
"    9 - old line 9                    │   9 + new line 9                        "
"   10 - old line 10                   │  10 + new line 10                       "
"   11 - old line 11                   │  11 + new line 11                       "
"   12 - old line 12                   │  12 + new line 12                       "
"   13 - old line 13                   │  13 + new line 13                       "
"   14 - old line 14                   │  14 + new line 14                       "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"                                                                                "
"   10 - old line 10                   │  10 + new line 10                       "
"   11 - old line 11                   │  11 + new line 11                       "
"   12 - old line 12                   │  12 + new line 12                       "
"   13 - old line 13                   │  13 + new line 13                       "
"   14 - old line 14                   │  14 + new line 14                       "
"   15 - old line 15                   │  15 + new line 15                       "
"   16 - old line 16                   │  16 + new line 16                       "
"   17 - old line 17                   │  17 + new line 17                       "
"   18 - old line 18                   │  18 + new line 18                       "
"   19 - old line 19                   │  19 + new line 19                       "
"   20 - old line 20                   │  20 + new line 20                       "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"                                                                                "
"   12 - old line 12                   │  12 + new line 12                       "
"   13 - old line 13                   │  13 + new line 13                       "
"   14 - old line 14                   │  14 + new line 14                       "
"   15 - old line 15                   │  15 + new line 15                       "
"   16 - old line 16                   │  16 + new line 16                       "
"   17 - old line 17                   │  17 + new line 17                       "
"   18 - old line 18                   │  18 + new line 18                       "
"   19 - old line 19                   │  19 + new line 19                       "
"   20 - old line 20                   │  20 + new line 20                       "
"                                      │                                         "
"                                      │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · file_1.txt · +1 -0                                                  "
"                                                                                "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │   1 + content_1                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('p'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"                                                                                "
"   12 - old line 12                   │  12 + new line 12                       "
"   13 - old line 13                   │  13 + new line 13                       "
"   14 - old line 14                   │  14 + new line 14                       "
"   15 - old line 15                   │  15 + new line 15                       "
"   16 - old line 16                   │  16 + new line 16                       "
"   17 - old line 17                   │  17 + new line 17                       "
"   18 - old line 18                   │  18 + new line 18                       "
"   19 - old line 19                   │  19 + new line 19                       "
"   20 - old line 20                   │  20 + new line 20                       "
"                                      │                                         "
"                                      │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_handles_close_hunks_and_bottom_additions() {
    reset_counter();

    let repo = TestRepo::new();
    let old_a = (1..=18)
        .map(|n| format!("line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\n";
    let new_a = vec![
        "line 1",
        "line 2",
        "line 3 changed",
        "line 4",
        "line 5 changed",
        "line 6",
        "line 7",
        "line 8",
        "line 9",
        "line 10",
        "line 11",
        "line 12",
        "line 13",
        "line 14",
        "line 15",
        "line 16",
        "line 17",
        "line 18",
        "line 19 added",
    ]
    .join("\n")
        + "\n";

    repo.add_file("a.txt", &old_a);
    repo.add_file("b.txt", "one\ntwo\n");
    repo.commit("Add edge case fixtures");

    repo.modify_file("a.txt", &new_a);
    repo.modify_file("b.txt", "one\nchanged\n");
    repo.commit("Update edge case fixtures");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Enter);

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · a.txt · +3 -2                                                       "
"                                                                                "
"                                      │                                         "
"    1   line 1                        │   1   line 1                            "
"    2   line 2                        │   2   line 2                            "
"    3 - line 3                        │   3 + line 3 changed                    "
"    4   line 4                        │   4   line 4                            "
"    5 - line 5                        │   5 + line 5 changed                    "
"    6   line 6                        │   6   line 6                            "
"    7   line 7                        │   7   line 7                            "
"    8   line 8                        │   8   line 8                            "
"    9   line 9                        │   9   line 9                            "
"   10   line 10                       │  10   line 10                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · a.txt · +3 -2                                                       "
"                                                                                "
"    2   line 2                        │   2   line 2                            "
"    3 - line 3                        │   3 + line 3 changed                    "
"    4   line 4                        │   4   line 4                            "
"    5 - line 5                        │   5 + line 5 changed                    "
"    6   line 6                        │   6   line 6                            "
"    7   line 7                        │   7   line 7                            "
"    8   line 8                        │   8   line 8                            "
"    9   line 9                        │   9   line 9                            "
"   10   line 10                       │  10   line 10                           "
"   11   line 11                       │  11   line 11                           "
"   12   line 12                       │  12   line 12                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · a.txt · +3 -2                                                       "
"                                                                                "
"   16   line 16                       │  16   line 16                           "
"   17   line 17                       │  17   line 17                           "
"   18   line 18                       │  18   line 18                           "
"                                      │  19 + line 19 added                     "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · b.txt · +1 -1                                                       "
"                                                                                "
"                                      │                                         "
"                                      │                                         "
"    1   one                           │   1   one                               "
"    2 - two                           │   2 + changed                           "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"                                      │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_prev_from_below_last_hunk_stays_in_file() {
    reset_counter();

    let repo = TestRepo::new();
    let old = (1..=20)
        .map(|n| format!("line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\n";
    let new = vec![
        "line 1",
        "line 2",
        "line 3 changed",
        "line 4",
        "line 5",
        "line 6",
        "line 7",
        "line 8",
        "line 9",
        "line 10 changed",
        "line 11",
        "line 12",
        "line 13",
        "line 14",
        "line 15",
        "line 16",
        "line 17",
        "line 18",
        "line 19",
        "line 20",
    ]
    .join("\n")
        + "\n";

    repo.add_file("a.txt", &old);
    repo.add_file("b.txt", "one\ntwo\n");
    repo.commit("Add p-navigation fixtures");

    repo.modify_file("a.txt", &new);
    repo.modify_file("b.txt", "one\nchanged\n");
    repo.commit("Update p-navigation fixtures");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Enter);

    for _ in 0..16 {
        h.press(KeyCode::Char('j'));
    }
    h.press(KeyCode::Char('p'));

    h.assert_snapshot(
        r#"
"  7dd25c4 · a.txt · +2 -2                                                       "
"                                                                                "
"    7   line 7                        │   7   line 7                            "
"    8   line 8                        │   8   line 8                            "
"    9   line 9                        │   9   line 9                            "
"   10 - line 10                       │  10 + line 10 changed                   "
"   11   line 11                       │  11   line 11                           "
"   12   line 12                       │  12   line 12                           "
"   13   line 13                       │  13   line 13                           "
"   14   line 14                       │  14   line 14                           "
"   15   line 15                       │  15   line 15                           "
"   16   line 16                       │  16   line 16                           "
"   17   line 17                       │  17   line 17                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}
