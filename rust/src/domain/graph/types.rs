pub struct GraphCommit {
    pub hash: String,
    pub parents: Vec<String>,
}

pub struct Layout {
    pub rows: Vec<Row>,
}

pub struct Row {
    pub symbols: Vec<GraphSymbol>,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum GraphSymbol {
    Empty,
    BranchPass,
    BranchCross,
    Commit,
    MergeCommit,
    BranchTop,
    BranchBottom,
    MergeJoin,
    Octopus,
    Diverge,
    MergeCross,
}

impl GraphSymbol {
    pub fn as_str(&self) -> &'static str {
        match self {
            GraphSymbol::Empty => "  ",
            GraphSymbol::BranchPass => "│ ",
            GraphSymbol::BranchCross => "│─",
            GraphSymbol::Commit => "├ ",
            GraphSymbol::MergeCommit => "├─",
            GraphSymbol::BranchTop => "╮ ",
            GraphSymbol::BranchBottom => "╯ ",
            GraphSymbol::MergeJoin => "┤ ",
            GraphSymbol::Octopus => "┬─",
            GraphSymbol::Diverge => "┴─",
            GraphSymbol::MergeCross => "┼─",
        }
    }
}

pub fn render_row(row: &Row) -> String {
    row.symbols.iter().map(|s| s.as_str()).collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_symbol_strings() {
        assert_eq!(GraphSymbol::Empty.as_str(), "  ");
        assert_eq!(GraphSymbol::BranchPass.as_str(), "│ ");
        assert_eq!(GraphSymbol::BranchCross.as_str(), "│─");
        assert_eq!(GraphSymbol::Commit.as_str(), "├ ");
        assert_eq!(GraphSymbol::MergeCommit.as_str(), "├─");
        assert_eq!(GraphSymbol::BranchTop.as_str(), "╮ ");
        assert_eq!(GraphSymbol::BranchBottom.as_str(), "╯ ");
        assert_eq!(GraphSymbol::MergeJoin.as_str(), "┤ ");
        assert_eq!(GraphSymbol::Octopus.as_str(), "┬─");
        assert_eq!(GraphSymbol::Diverge.as_str(), "┴─");
        assert_eq!(GraphSymbol::MergeCross.as_str(), "┼─");
    }

    #[test]
    fn test_render_row() {
        let row = Row {
            symbols: vec![GraphSymbol::Commit, GraphSymbol::BranchTop],
        };
        assert_eq!(render_row(&row), "├ ╮ ");
    }
}
