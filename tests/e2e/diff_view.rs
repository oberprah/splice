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
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"    1   pub struct Calculator;        │   1   pub struct Calculator;            "
"    2                                 │   2                                     "
"    3   impl Calculator {             │   3   impl Calculator {                 "
"    4       pub fn new() -> Self {    │   4       pub fn new() -> Self {        "
"    5           Self                  │   5           Self                      "
"    6       }                         │   6       }                             "
"    7                                 │   7                                     "
"    8       pub fn add(&self, a: i32, │   8       pub fn add(&self, a: i32,     "
"    ↪   b: i32) -> i32 {              │   ↪   i32) -> i32 {                     "
"    9           a + b                 │   9           a + b                     "
"   10       }                         │  10       }                             "
"   11                                 │  11                                     "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"    ↪   b: i32) -> i32 {              │   ↪   i32) -> i32 {                     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"    ↪ - i32, b: i32) -> i32 {         │   ↪ + i32) -> i32 {                     "
"   17 -         a * b                 │  17 +                                   "
"                                      │   ↪ + a.checked_mul(b).unwrap_or(0)     "
"   18 -     }                         │                                         "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"   11                                 │  11                                     "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"    ↪   b: i32) -> i32 {              │   ↪   i32) -> i32 {                     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"    ↪ - i32, b: i32) -> i32 {         │   ↪ + i32) -> i32 {                     "
"   17 -         a * b                 │  17 +                                   "
"                                      │   ↪ + a.checked_mul(b).unwrap_or(0)     "
"   18 -     }                         │                                         "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a:   │                                         "
"    ↪ - i32, b: i32) -> Option<i32> { │                                         "
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

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a:   │                                         "
"    ↪ - i32, b: i32) -> Option<i32> { │                                         "
"   21 -         if b == 0 {           │                                         "
"   22 -             None              │                                         "
"   23 -         } else {              │                                         "
"   24 -             Some(a / b)       │                                         "
"   25 -         }                     │                                         "
"   26       }                         │  18       }                             "
"   27                                 │  19                                     "
"   28       pub fn sum(&self, values: │  20       pub fn sum(&self, values:     "
"    ↪   &[i32]) -> i32 {              │   ↪   &[i32]) -> i32 {                  "
"   29           values.iter().sum()   │  21           values.iter().sum()       "
"   30       }                         │  22       }                             "
"   31                                 │  23                                     "
"                                      │  24 +     pub fn pow(&self, base: i3    "
"                                      │   ↪ + exp: u32) -> i32 {                "
"                                      │  25 +         (0..exp).fold(1, |acc,    "
"                                      │   ↪ + _| acc * base)                    "
"                                      │  26 +     }                             "
"                                      │  27 +                                   "
"   32       pub fn format_result(&sel │  28       pub fn format_result(&self    "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press_ctrl(KeyCode::Char('u'));
    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"   12       pub fn sub(&self, a: i32, │  12       pub fn sub(&self, a: i32,     "
"    ↪   b: i32) -> i32 {              │   ↪   i32) -> i32 {                     "
"   13           a - b                 │  13           a - b                     "
"   14       }                         │  14       }                             "
"   15                                 │  15                                     "
"   16 -     pub fn multiply(&self, a: │  16 +     pub fn mul(&self, a: i32,     "
"    ↪ - i32, b: i32) -> i32 {         │   ↪ + i32) -> i32 {                     "
"   17 -         a * b                 │  17 +                                   "
"                                      │   ↪ + a.checked_mul(b).unwrap_or(0)     "
"   18 -     }                         │                                         "
"   19 -                               │                                         "
"   20 -     pub fn divide(&self, a:   │                                         "
"    ↪ - i32, b: i32) -> Option<i32> { │                                         "
"   21 -         if b == 0 {           │                                         "
"   22 -             None              │                                         "
"   23 -         } else {              │                                         "
"   24 -             Some(a / b)       │                                         "
"   25 -         }                     │                                         "
"   26       }                         │  18       }                             "
"   27                                 │  19                                     "
"   28       pub fn sum(&self, values: │  20       pub fn sum(&self, values:     "
"    ↪   &[i32]) -> i32 {              │   ↪   &[i32]) -> i32 {                  "
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
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  b0799a4 · src/ui/log.rs · +1 -0                                                                                       "
"    1   impl LineDisplayState {                           │   1   impl LineDisplayState {                               "
"    2       fn prefix(&self) -> &'static str {            │   2       fn prefix(&self) -> &'static str {                "
"    3           match self {                              │   3           match self {                                  "
"                                                          │   4 +             LineDisplayState::None => "  ",           "
"    4               LineDisplayState::Cursor => "→ ",     │   5               LineDisplayState::Cursor => "→ ",         "
"    5               LineDisplayState::Selected => "▌ ",   │   6               LineDisplayState::Selected => "▌ ",       "
"    6           }                                         │   7           }                                             "
"    7       }                                             │   8       }                                                 "
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
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"    1 - old line 1                    │   1 + new line 1                        "
"    2 - old line 2                    │   2 + new line 2                        "
"    3 - old line 3                    │   3 + new line 3                        "
"    4 - old line 4                    │   4 + new line 4                        "
"    5 - old line 5                    │   5 + new line 5                        "
"    6 - old line 6                    │   6 + new line 6                        "
"    7 - old line 7                    │   7 + new line 7                        "
"    8 - old line 8                    │   8 + new line 8                        "
"    9 - old line 9                    │   9 + new line 9                        "
"   10 - old line 10                   │  10 + new line 10                       "
"   11 - old line 11                   │  11 + new line 11                       "
"   12 - old line 12                   │  12 + new line 12                       "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"    7 - old line 7                    │   7 + new line 7                        "
"    8 - old line 8                    │   8 + new line 8                        "
"    9 - old line 9                    │   9 + new line 9                        "
"   10 - old line 10                   │  10 + new line 10                       "
"   11 - old line 11                   │  11 + new line 11                       "
"   12 - old line 12                   │  12 + new line 12                       "
"   13 - old line 13                   │  13 + new line 13                       "
"   14 - old line 14                   │  14 + new line 14                       "
"   15 - old line 15                   │  15 + new line 15                       "
"   16 - old line 16                   │  16 + new line 16                       "
"   17 - old line 17                   │  17 + new line 17                       "
"   18 - old line 18                   │  18 + new line 18                       "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · src/large.txt · +20 -20                                             "
"    9 - old line 9                    │   9 + new line 9                        "
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
"  331f884 · file_1.txt · +1 -0                                                  "
"                                      │   1 + content_1                         "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  331f884 · notes.md · +1 -1                                                    "
"    1   alpha                         │   1   alpha                             "
"    2 - beta                          │   2 + gamma                             "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('p'));
    h.assert_snapshot(
        r#"
"  331f884 · file_1.txt · +1 -0                                                  "
"                                      │   1 + content_1                         "
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
"   13   line 13                       │  13   line 13                           "
"   14   line 14                       │  14   line 14                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · a.txt · +3 -2                                                       "
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
"   13   line 13                       │  13   line 13                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · a.txt · +3 -2                                                       "
"    8   line 8                        │   8   line 8                            "
"    9   line 9                        │   9   line 9                            "
"   10   line 10                       │  10   line 10                           "
"   11   line 11                       │  11   line 11                           "
"   12   line 12                       │  12   line 12                           "
"   13   line 13                       │  13   line 13                           "
"   14   line 14                       │  14   line 14                           "
"   15   line 15                       │  15   line 15                           "
"   16   line 16                       │  16   line 16                           "
"   17   line 17                       │  17   line 17                           "
"   18   line 18                       │  18   line 18                           "
"                                      │  19 + line 19 added                     "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  bb76eb0 · b.txt · +1 -1                                                       "
"    1   one                           │   1   one                               "
"    2 - two                           │   2 + changed                           "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_soft_wrap_long_lines() {
    reset_counter();

    let repo = TestRepo::new();
    let original = r#"fn main() {
    let very_long_variable_name_that_definitely_exceeds_screen_width = "this is an extremely long string value that should wrap across multiple lines when displayed in the diff view";
    println!("{}", very_long_variable_name_that_definitely_exceeds_screen_width);
}
"#;
    let updated = r#"fn main() {
    let very_long_variable_name_that_definitely_exceeds_screen_width = "this is an updated extremely long string value that should wrap across multiple lines when displayed in the diff view";
    let another_long_variable_name_for_testing = "another very long string to test soft wrapping behavior with multiple wrapped segments in the diff view";
    println!("{}", very_long_variable_name_that_definitely_exceeds_screen_width);
    println!("{}", another_long_variable_name_for_testing);
}
"#;

    repo.add_file("src/main.rs", original);
    repo.commit("Initial commit");
    repo.modify_file("src/main.rs", updated);
    repo.commit("Update with long lines");

    let mut h = Harness::with_repo_and_screen_size(&repo, 60, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Enter);

    h.assert_snapshot(
        r#"
"  e65f008 · src/main.rs · +3 -1                             "
"    1   fn main() {         │   1   fn main() {             "
"    2 -     let             │   2 +     let                 "
"    ↪ - very_long_variable_ │   ↪ + very_long_variable_n    "
"    ↪ - me_that_definitely_ │   ↪ + e_that_definitely_ex    "
"    ↪ - ceeds_screen_width  │   ↪ + eds_screen_width =      "
"    ↪ - "this is an extreme │   ↪ + "this is an updated     "
"    ↪ - long string value   │   ↪ + extremely long strin    "
"    ↪ - that should wrap    │   ↪ + value that should wr    "
"    ↪ - across multiple lin │   ↪ + across multiple line    "
"    ↪ - when displayed in t │   ↪ + when displayed in th    "
"    ↪ - diff view";         │   ↪ + diff view";             "
"                            │   3 +     let                 "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back        "
"#,
    );

    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
"  e65f008 · src/main.rs · +3 -1                             "
"    2 -     let             │   2 +     let                 "
"    ↪ - very_long_variable_ │   ↪ + very_long_variable_n    "
"    ↪ - me_that_definitely_ │   ↪ + e_that_definitely_ex    "
"    ↪ - ceeds_screen_width  │   ↪ + eds_screen_width =      "
"    ↪ - "this is an extreme │   ↪ + "this is an updated     "
"    ↪ - long string value   │   ↪ + extremely long strin    "
"    ↪ - that should wrap    │   ↪ + value that should wr    "
"    ↪ - across multiple lin │   ↪ + across multiple line    "
"    ↪ - when displayed in t │   ↪ + when displayed in th    "
"    ↪ - diff view";         │   ↪ + diff view";             "
"                            │   3 +     let                 "
"                            │   ↪ + another_long_variabl    "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back        "
"#,
    );

    h.press(KeyCode::Char('k'));
    h.assert_snapshot(
        r#"
"  e65f008 · src/main.rs · +3 -1                             "
"    1   fn main() {         │   1   fn main() {             "
"    2 -     let             │   2 +     let                 "
"    ↪ - very_long_variable_ │   ↪ + very_long_variable_n    "
"    ↪ - me_that_definitely_ │   ↪ + e_that_definitely_ex    "
"    ↪ - ceeds_screen_width  │   ↪ + eds_screen_width =      "
"    ↪ - "this is an extreme │   ↪ + "this is an updated     "
"    ↪ - long string value   │   ↪ + extremely long strin    "
"    ↪ - that should wrap    │   ↪ + value that should wr    "
"    ↪ - across multiple lin │   ↪ + across multiple line    "
"    ↪ - when displayed in t │   ↪ + when displayed in th    "
"    ↪ - diff view";         │   ↪ + diff view";             "
"                            │   3 +     let                 "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back        "
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
"    9   line 9                        │   9   line 9                            "
"   10 - line 10                       │  10 + line 10 changed                   "
"   11   line 11                       │  11   line 11                           "
"   12   line 12                       │  12   line 12                           "
"   13   line 13                       │  13   line 13                           "
"   14   line 14                       │  14   line 14                           "
"   15   line 15                       │  15   line 15                           "
"   16   line 16                       │  16   line 16                           "
"   17   line 17                       │  17   line 17                           "
"   18   line 18                       │  18   line 18                           "
"   19   line 19                       │  19   line 19                           "
"   20   line 20                       │  20   line 20                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_n_does_not_skip_last_hunk_when_scroll_clamped() {
    reset_counter();

    let repo = TestRepo::new();
    let old_a: String = (1..=18)
        .map(|n| format!("line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\nline 19\nline 20\n";
    let new_a: String = (1..=18)
        .map(|n| format!("line {n}"))
        .collect::<Vec<_>>()
        .join("\n")
        + "\nline 19 changed\nline 20 changed\n";

    repo.add_file("a.txt", &old_a);
    repo.add_file("b.txt", "alpha\nbeta\n");
    repo.commit("Add fixtures");

    repo.modify_file("a.txt", &new_a);
    repo.modify_file("b.txt", "alpha\ngamma\n");
    repo.commit("Modify fixtures");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
"  159c9a8 · a.txt · +2 -2                                                       "
"    1   line 1                        │   1   line 1                            "
"    2   line 2                        │   2   line 2                            "
"    3   line 3                        │   3   line 3                            "
"    4   line 4                        │   4   line 4                            "
"    5   line 5                        │   5   line 5                            "
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

    for _ in 0..10 {
        h.press(KeyCode::Char('j'));
    }

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  159c9a8 · a.txt · +2 -2                                                       "
"    9   line 9                        │   9   line 9                            "
"   10   line 10                       │  10   line 10                           "
"   11   line 11                       │  11   line 11                           "
"   12   line 12                       │  12   line 12                           "
"   13   line 13                       │  13   line 13                           "
"   14   line 14                       │  14   line 14                           "
"   15   line 15                       │  15   line 15                           "
"   16   line 16                       │  16   line 16                           "
"   17   line 17                       │  17   line 17                           "
"   18   line 18                       │  18   line 18                           "
"   19 - line 19                       │  19 + line 19 changed                   "
"   20 - line 20                       │  20 + line 20 changed                   "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  159c9a8 · b.txt · +1 -1                                                       "
"    1   alpha                         │   1   alpha                             "
"    2 - beta                          │   2 + gamma                             "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_can_scroll_to_last_lines_of_purely_added_file() {
    reset_counter();

    let repo = TestRepo::new();
    let content: String = (1..=20).map(|n| format!("line {n}\n")).collect();

    repo.commit("Initial commit");
    repo.add_file("new_file.txt", &content);
    repo.commit("Add new file");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Char('j'));
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
"  0be46d7 · new_file.txt · +20 -0                                               "
"                                      │   1 + line 1                            "
"                                      │   2 + line 2                            "
"                                      │   3 + line 3                            "
"                                      │   4 + line 4                            "
"                                      │   5 + line 5                            "
"                                      │   6 + line 6                            "
"                                      │   7 + line 7                            "
"                                      │   8 + line 8                            "
"                                      │   9 + line 9                            "
"                                      │  10 + line 10                           "
"                                      │  11 + line 11                           "
"                                      │  12 + line 12                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    for _ in 0..20 {
        h.press(KeyCode::Char('j'));
    }
    h.assert_snapshot(
        r#"
"  0be46d7 · new_file.txt · +20 -0                                               "
"                                      │   9 + line 9                            "
"                                      │  10 + line 10                           "
"                                      │  11 + line 11                           "
"                                      │  12 + line 12                           "
"                                      │  13 + line 13                           "
"                                      │  14 + line 14                           "
"                                      │  15 + line 15                           "
"                                      │  16 + line 16                           "
"                                      │  17 + line 17                           "
"                                      │  18 + line 18                           "
"                                      │  19 + line 19                           "
"                                      │  20 + line 20                           "
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}

#[test]
#[serial]
fn diff_view_focused_hunk_set_when_crossing_to_new_file() {
    reset_counter();

    let repo = TestRepo::new();
    repo.add_file("a.txt", "one\ntwo\n");
    repo.add_file("b.txt", "alpha\nbeta\n");
    repo.add_file("c.txt", "x\ny\n");
    repo.commit("Add fixtures");

    repo.modify_file("a.txt", "CHANGED\ntwo\n");
    repo.modify_file("b.txt", "alpha\ngamma\n");
    repo.modify_file("c.txt", "x\nz\n");
    repo.commit("Modify fixtures");

    let mut h = Harness::with_repo_and_screen_size(&repo, 80, 14);
    h.press(KeyCode::Enter);
    h.press(KeyCode::Enter);
    h.assert_snapshot(
        r#"
"  d53285e · a.txt · +1 -1                                                       "
"    1 - one                           │   1 + CHANGED                           "
"    2   two                           │   2   two                               "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  d53285e · b.txt · +1 -1                                                       "
"    1   alpha                         │   1   alpha                             "
"    2 - beta                          │   2 + gamma                             "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );

    h.press(KeyCode::Char('n'));
    h.assert_snapshot(
        r#"
"  d53285e · c.txt · +1 -1                                                       "
"    1   x                             │   1   x                                 "
"    2 - y                             │   2 + z                                 "
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
"  j/k: scroll  n/p: next/prev diff  o: open  q: back                            "
"#,
    );
}
