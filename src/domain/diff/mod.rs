mod builder;
mod types;

pub use builder::build_file_diff;
pub use types::{DiffBlock, DiffLine, FileDiff, UnchangedLine};
