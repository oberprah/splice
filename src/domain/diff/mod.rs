mod builder;
mod types;

pub use builder::{build_file_diff, build_file_diff_full};
pub use types::{ChangeBlock, DiffBlock, DiffMeta, FileDiff, UnchangedBlock};
