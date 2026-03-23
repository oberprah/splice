mod builder;
pub mod inline_diff;
pub mod layout;
mod types;

pub use builder::build_file_diff;
pub use layout::{build_rows, Cell, CellKind, ScreenRow};
pub use types::{DiffBlock, DiffLine, FileDiff, HunkRange, UnchangedLine};
