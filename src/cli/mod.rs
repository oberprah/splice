use crate::core::UncommittedType;

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Command {
    Log,
    Diff(DiffSpec),
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct DiffSpec {
    pub raw: Option<String>,
    pub uncommitted_type: Option<UncommittedType>,
}

pub fn parse_args(args: &[String]) -> Command {
    if args.len() < 2 {
        return Command::Log;
    }

    if args[1] == "diff" {
        let diff_args = parse_diff_args(&args[2..]).unwrap_or(DiffSpec {
            raw: None,
            uncommitted_type: Some(UncommittedType::Unstaged),
        });
        Command::Diff(diff_args)
    } else {
        Command::Log
    }
}

pub fn parse_diff_args(args: &[String]) -> Result<DiffSpec, String> {
    if args.is_empty() {
        return Ok(DiffSpec {
            raw: None,
            uncommitted_type: Some(UncommittedType::Unstaged),
        });
    }

    if args.len() > 1 {
        return Err(format!("unexpected arguments: {:?}", &args[1..]));
    }

    let first_arg = &args[0];

    match first_arg.as_str() {
        "--staged" | "--cached" => Ok(DiffSpec {
            raw: None,
            uncommitted_type: Some(UncommittedType::Staged),
        }),
        "HEAD" => Ok(DiffSpec {
            raw: None,
            uncommitted_type: Some(UncommittedType::All),
        }),
        _ => {
            if !is_valid_diff_spec(first_arg) {
                return Err(format!("invalid diff spec: {:?}", first_arg));
            }
            Ok(DiffSpec {
                raw: Some(first_arg.clone()),
                uncommitted_type: None,
            })
        }
    }
}

pub fn is_valid_diff_spec(spec: &str) -> bool {
    let forbidden = [' ', ';', '|', '&', '>', '<', '$', '`'];
    !spec.chars().any(|c| forbidden.contains(&c))
}

#[cfg(test)]
mod tests {
    use super::*;

    fn args(input: &[&str]) -> Vec<String> {
        input.iter().map(|s| s.to_string()).collect()
    }

    mod parse_args {
        use super::*;

        #[test]
        fn no_args_returns_log() {
            assert_eq!(parse_args(&args(&["splice"])), Command::Log);
        }

        #[test]
        fn unknown_command_returns_log() {
            assert_eq!(parse_args(&args(&["splice", "unknown"])), Command::Log);
        }

        #[test]
        fn diff_command_returns_diff_with_unstaged() {
            let result = parse_args(&args(&["splice", "diff"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Unstaged),
                })
            );
        }

        #[test]
        fn diff_with_staged_flag() {
            let result = parse_args(&args(&["splice", "diff", "--staged"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Staged),
                })
            );
        }

        #[test]
        fn diff_with_cached_flag() {
            let result = parse_args(&args(&["splice", "diff", "--cached"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Staged),
                })
            );
        }

        #[test]
        fn diff_with_head() {
            let result = parse_args(&args(&["splice", "diff", "HEAD"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::All),
                })
            );
        }

        #[test]
        fn diff_with_commit_range() {
            let result = parse_args(&args(&["splice", "diff", "main..feature"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: Some("main..feature".to_string()),
                    uncommitted_type: None,
                })
            );
        }

        #[test]
        fn diff_with_invalid_spec_falls_back_to_unstaged() {
            let result = parse_args(&args(&["splice", "diff", "invalid;spec"]));
            assert_eq!(
                result,
                Command::Diff(DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Unstaged),
                })
            );
        }
    }

    mod parse_diff_args {
        use super::*;

        #[test]
        fn no_args_returns_unstaged() {
            let result = parse_diff_args(&[]).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Unstaged),
                }
            );
        }

        #[test]
        fn staged_flag() {
            let result = parse_diff_args(&args(&["--staged"])).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Staged),
                }
            );
        }

        #[test]
        fn cached_flag() {
            let result = parse_diff_args(&args(&["--cached"])).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::Staged),
                }
            );
        }

        #[test]
        fn head_keyword() {
            let result = parse_diff_args(&args(&["HEAD"])).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: None,
                    uncommitted_type: Some(UncommittedType::All),
                }
            );
        }

        #[test]
        fn commit_range_spec() {
            let result = parse_diff_args(&args(&["main..feature"])).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: Some("main..feature".to_string()),
                    uncommitted_type: None,
                }
            );
        }

        #[test]
        fn single_commit_hash() {
            let result = parse_diff_args(&args(&["abc123"])).unwrap();
            assert_eq!(
                result,
                DiffSpec {
                    raw: Some("abc123".to_string()),
                    uncommitted_type: None,
                }
            );
        }

        #[test]
        fn too_many_args_returns_error() {
            let result = parse_diff_args(&args(&["main", "feature"]));
            assert!(result.is_err());
            assert!(result.unwrap_err().contains("unexpected arguments"));
        }

        #[test]
        fn invalid_spec_with_space_returns_error() {
            let result = parse_diff_args(&args(&["main feature"]));
            assert!(result.is_err());
            assert!(result.unwrap_err().contains("invalid diff spec"));
        }

        #[test]
        fn invalid_spec_with_semicolon_returns_error() {
            let result = parse_diff_args(&args(&["main;rm"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_pipe_returns_error() {
            let result = parse_diff_args(&args(&["main|cat"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_ampersand_returns_error() {
            let result = parse_diff_args(&args(&["main&&ls"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_greater_than_returns_error() {
            let result = parse_diff_args(&args(&["main>file"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_less_than_returns_error() {
            let result = parse_diff_args(&args(&["main<file"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_dollar_returns_error() {
            let result = parse_diff_args(&args(&["$HOME"]));
            assert!(result.is_err());
        }

        #[test]
        fn invalid_spec_with_backtick_returns_error() {
            let result = parse_diff_args(&args(&["`cmd`"]));
            assert!(result.is_err());
        }
    }

    mod is_valid_diff_spec {
        use super::*;

        #[test]
        fn valid_simple_spec() {
            assert!(is_valid_diff_spec("main"));
        }

        #[test]
        fn valid_commit_range() {
            assert!(is_valid_diff_spec("main..feature"));
        }

        #[test]
        fn valid_commit_hash() {
            assert!(is_valid_diff_spec("abc123def456"));
        }

        #[test]
        fn rejects_space() {
            assert!(!is_valid_diff_spec("main feature"));
        }

        #[test]
        fn rejects_semicolon() {
            assert!(!is_valid_diff_spec("main;rm"));
        }

        #[test]
        fn rejects_pipe() {
            assert!(!is_valid_diff_spec("main|cat"));
        }

        #[test]
        fn rejects_ampersand() {
            assert!(!is_valid_diff_spec("main&&ls"));
        }

        #[test]
        fn rejects_greater_than() {
            assert!(!is_valid_diff_spec("main>file"));
        }

        #[test]
        fn rejects_less_than() {
            assert!(!is_valid_diff_spec("main<file"));
        }

        #[test]
        fn rejects_dollar() {
            assert!(!is_valid_diff_spec("$HOME"));
        }

        #[test]
        fn rejects_backtick() {
            assert!(!is_valid_diff_spec("`cmd`"));
        }

        #[test]
        fn empty_spec_is_valid() {
            assert!(is_valid_diff_spec(""));
        }
    }
}
