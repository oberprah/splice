use crate::common::{reset_counter, Harness, TestRepo};
use crossterm::event::KeyCode;
use serial_test::serial;

#[test]
#[serial]
fn diff_view_mvp_side_by_side() {
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
"   29           values.iter().sum()   │  21           values.iter().sum()       "
"   30       }                         │  22       }                             "
"   31                                 │  23                                     "
"                                      │  24 +     pub fn pow(&self, base: i3    "
"                                      │  25 +         (0..exp).fold(1, |acc,    "
"  j/k: scroll  q: back                                                          "
"#,
    );

    h.press_ctrl(KeyCode::Char('d'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
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
"   29           values.iter().sum()   │  21           values.iter().sum()       "
"   30       }                         │  22       }                             "
"   31                                 │  23                                     "
"                                      │  24 +     pub fn pow(&self, base: i3    "
"                                      │  25 +         (0..exp).fold(1, |acc,    "
"                                      │  26 +     }                             "
"                                      │  27 +                                   "
"   32       pub fn format_result(&sel │  28       pub fn format_result(&self    "
"   33 -         format!("Result: {}", │  29 +         format!("Result: {valu    "
"  j/k: scroll  q: back                                                          "
"#,
    );

    h.press_ctrl(KeyCode::Char('u'));
    h.press(KeyCode::Char('j'));
    h.assert_snapshot(
        r#"
"  0fdee5c · src/calculator.rs · +7 -11                                          "
"                                                                                "
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
"   29           values.iter().sum()   │  21           values.iter().sum()       "
"   30       }                         │  22       }                             "
"   31                                 │  23                                     "
"                                      │  24 +     pub fn pow(&self, base: i3    "
"                                      │  25 +         (0..exp).fold(1, |acc,    "
"                                      │  26 +     }                             "
"  j/k: scroll  q: back                                                          "
"#,
    );
}
