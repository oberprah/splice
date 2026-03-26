mod builder;
pub mod inline_diff;
pub mod layout;
mod types;

pub use builder::build_file_diff;
pub use layout::{build_rows, build_unified_rows, Cell, CellKind, ScreenRow, UnifiedRow};
pub use types::{DiffBlock, DiffLine, FileDiff, HunkRange, UnchangedLine};
