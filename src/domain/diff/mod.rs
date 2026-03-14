mod builder;
mod types;

pub use builder::{build_file_diff, build_file_diff_full};
pub use types::{DiffBlock, DiffLine, FileDiff, UnchangedLine};
