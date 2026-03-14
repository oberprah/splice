mod builder;
pub mod layout;
mod types;

pub use builder::build_file_diff;
pub use layout::{build_rows, Cell, CellKind, HunkRange, ScreenRow};
pub use types::{DiffBlock, DiffLine, FileDiff, UnchangedLine};
