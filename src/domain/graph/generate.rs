use super::types::{GraphSymbol, Row};

pub fn generate_row_symbols(
    commit_col: usize,
    num_cols: usize,
    merge_columns: &[usize],
    converging_columns: &[usize],
    passing_columns: &[usize],
    existing_lanes_merge: &[usize],
    converges_to_parent: bool,
) -> Row {
    if num_cols == 0 {
        return Row { symbols: vec![] };
    }

    let mut symbols = vec![GraphSymbol::Empty; num_cols];

    let rightmost_merge = merge_columns.iter().copied().max();
    let rightmost_converge = converging_columns.iter().copied().max();

    let mut rightmost_horizontal = match (rightmost_merge, rightmost_converge) {
        (Some(m), Some(c)) => std::cmp::max(m, c),
        (Some(m), None) => m,
        (None, Some(c)) => c,
        (None, None) => commit_col,
    };

    if !existing_lanes_merge.is_empty() {
        rightmost_horizontal += 1;
    }

    let convergence_symbol_col = if converges_to_parent {
        Some(num_cols - 1)
    } else {
        None
    };

    if let Some(conv_col) = convergence_symbol_col {
        if commit_col < conv_col {
            rightmost_horizontal = std::cmp::max(rightmost_horizontal, conv_col);
        }
    }

    let merge_set: std::collections::HashSet<usize> = merge_columns.iter().copied().collect();
    let converge_set: std::collections::HashSet<usize> =
        converging_columns.iter().copied().collect();
    let passing_set: std::collections::HashSet<usize> = passing_columns.iter().copied().collect();
    let existing_lane_merge_set: std::collections::HashSet<usize> =
        existing_lanes_merge.iter().copied().collect();

    for (col, symbol) in symbols.iter_mut().enumerate() {
        if col == commit_col {
            if converges_to_parent || rightmost_horizontal > commit_col {
                *symbol = GraphSymbol::MergeCommit;
            } else {
                *symbol = GraphSymbol::Commit;
            }
        } else if merge_set.contains(&col)
            && converge_set.contains(&col)
            && !existing_lane_merge_set.contains(&col)
        {
            if col < rightmost_horizontal {
                *symbol = GraphSymbol::MergeCross;
            } else {
                *symbol = GraphSymbol::MergeJoin;
            }
        } else if merge_set.contains(&col) && !existing_lane_merge_set.contains(&col) {
            if let Some(rm) = rightmost_merge {
                if col < rm {
                    *symbol = GraphSymbol::Octopus;
                } else {
                    *symbol = GraphSymbol::BranchTop;
                }
            } else {
                *symbol = GraphSymbol::BranchTop;
            }
        } else if converge_set.contains(&col) {
            if let Some(rc) = rightmost_converge {
                if col < rc {
                    *symbol = GraphSymbol::Diverge;
                } else {
                    *symbol = GraphSymbol::BranchBottom;
                }
            } else {
                *symbol = GraphSymbol::BranchBottom;
            }
        } else if passing_set.contains(&col) || existing_lane_merge_set.contains(&col) {
            if col > commit_col && col < rightmost_horizontal {
                *symbol = GraphSymbol::BranchCross;
            } else {
                *symbol = GraphSymbol::BranchPass;
            }
        }

        if col == rightmost_horizontal && !existing_lanes_merge.is_empty() {
            *symbol = GraphSymbol::BranchTop;
        }

        if Some(col) == convergence_symbol_col {
            *symbol = GraphSymbol::BranchBottom;
        }
    }

    Row { symbols }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_row_commit_only() {
        let row = generate_row_symbols(0, 1, &[], &[], &[], &[], false);
        assert_eq!(row.symbols, vec![GraphSymbol::Commit]);
    }

    #[test]
    fn test_generate_row_with_merge() {
        let row = generate_row_symbols(0, 2, &[1], &[], &[], &[], false);
        assert_eq!(
            row.symbols,
            vec![GraphSymbol::MergeCommit, GraphSymbol::BranchTop]
        );
    }

    #[test]
    fn test_generate_row_with_convergence() {
        let row = generate_row_symbols(0, 2, &[], &[1], &[], &[], false);
        assert_eq!(
            row.symbols,
            vec![GraphSymbol::MergeCommit, GraphSymbol::BranchBottom]
        );
    }
}
