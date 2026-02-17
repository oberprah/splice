use super::lanes::{
    assign_column, collapse_trailing_empty, detect_converging_columns, detect_passing_columns,
    update_lanes,
};
use super::types::{GraphCommit, Layout, Row};
use crate::domain::graph::generate::generate_row_symbols;
use std::collections::HashMap;

pub fn compute_layout(commits: &[GraphCommit]) -> Layout {
    if commits.is_empty() {
        return Layout { rows: vec![] };
    }

    let mut rows: Vec<Row> = Vec::new();
    let mut lanes: Vec<String> = Vec::new();
    let mut prev_merged_hashes: HashMap<String, bool> = HashMap::new();

    for commit in commits {
        let col = assign_column(&commit.hash, &mut lanes);

        let converging_columns = detect_converging_columns(col, &commit.hash, &lanes);

        for converging_col in &converging_columns {
            lanes[*converging_col].clear();
        }

        let was_in_existing_lane_merge = prev_merged_hashes.contains_key(&commit.hash);

        let lanes_copy = lanes.clone();

        let mut update_result = update_lanes(col, &commit.parents, lanes);

        if was_in_existing_lane_merge && commit.parents.len() == 1 {
            update_result.converges_to_parent = true;
        }

        let mut current_merged_hashes: HashMap<String, bool> = HashMap::new();
        for merge_col in &update_result.existing_lanes_merge {
            if *merge_col < lanes_copy.len() {
                let merged_hash = &lanes_copy[*merge_col];
                if !merged_hash.is_empty() {
                    current_merged_hashes.insert(merged_hash.clone(), true);
                }
            }
        }
        prev_merged_hashes = current_merged_hashes;

        let passing_columns = detect_passing_columns(
            col,
            &update_result.lanes,
            &update_result.merge_columns,
            &converging_columns,
        );

        let mut num_cols = update_result.lanes.len();
        if num_cols == 0 {
            num_cols = 1;
        }
        if !update_result.existing_lanes_merge.is_empty() {
            num_cols += 1;
        }
        if update_result.converges_to_parent {
            num_cols += 1;
        }

        let row = generate_row_symbols(
            col,
            num_cols,
            &update_result.merge_columns,
            &converging_columns,
            &passing_columns,
            &update_result.existing_lanes_merge,
            update_result.converges_to_parent,
        );
        rows.push(row);

        lanes = collapse_trailing_empty(update_result.lanes);
    }

    Layout { rows }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn render_layout(layout: &Layout) -> Vec<String> {
        layout
            .rows
            .iter()
            .map(|row| super::super::types::render_row(row).trim_end_matches(' ').to_string())
            .collect()
    }

    #[test]
    fn test_empty() {
        let layout = compute_layout(&[]);
        assert!(layout.rows.is_empty());
    }

    #[test]
    fn test_linear_history() {
        let commits = vec![
            GraphCommit { hash: "D".to_string(), parents: vec!["C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["B".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├", "├", "├", "├"]);
    }

    #[test]
    fn test_simple_merge() {
        let commits = vec![
            GraphCommit { hash: "E".to_string(), parents: vec!["B".to_string(), "D".to_string()] },
            GraphCommit { hash: "D".to_string(), parents: vec!["C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─╮", "│ ├", "│ ├", "├ │", "├─╯"]);
    }

    #[test]
    fn test_root_commit() {
        let commits = vec![
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├"]);
    }

    #[test]
    fn test_multiple_roots() {
        let commits = vec![
            GraphCommit { hash: "D".to_string(), parents: vec!["B".to_string(), "C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec![] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─╮", "│ ├", "├", "├"]);
    }

    #[test]
    fn test_octopus_merge() {
        let commits = vec![
            GraphCommit { hash: "G".to_string(), parents: vec!["A".to_string(), "D".to_string(), "F".to_string()] },
            GraphCommit { hash: "F".to_string(), parents: vec!["E".to_string()] },
            GraphCommit { hash: "E".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "D".to_string(), parents: vec!["C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─┬─╮", "│ │ ├", "│ │ ├", "│ ├ │", "│ ├ │", "├─┴─╯"]);
    }

    #[test]
    fn test_sequential_merges() {
        let commits = vec![
            GraphCommit { hash: "G".to_string(), parents: vec!["D".to_string(), "F".to_string()] },
            GraphCommit { hash: "F".to_string(), parents: vec!["E".to_string()] },
            GraphCommit { hash: "E".to_string(), parents: vec!["D".to_string()] },
            GraphCommit { hash: "D".to_string(), parents: vec!["A".to_string(), "C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["B".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─╮", "│ ├", "│ ├", "├─┤", "│ ├", "│ ├", "├─╯"]);
    }

    #[test]
    fn test_sequential_merges_with_main_commits() {
        let commits = vec![
            GraphCommit { hash: "H".to_string(), parents: vec!["F".to_string(), "G".to_string()] },
            GraphCommit { hash: "G".to_string(), parents: vec!["F".to_string()] },
            GraphCommit { hash: "F".to_string(), parents: vec!["E".to_string()] },
            GraphCommit { hash: "E".to_string(), parents: vec!["B".to_string(), "D".to_string()] },
            GraphCommit { hash: "D".to_string(), parents: vec!["C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["B".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─╮", "│ ├", "├─╯", "├─╮", "│ ├", "│ ├", "├─╯", "├"]);
    }

    #[test]
    fn test_passing_lanes() {
        let commits = vec![
            GraphCommit { hash: "D".to_string(), parents: vec!["B".to_string(), "C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["B".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec!["├─╮", "│ ├", "├─╯", "├"]);
    }

    #[test]
    fn test_complex_multi_branch() {
        let commits = vec![
            GraphCommit { hash: "M".to_string(), parents: vec!["J".to_string(), "L".to_string()] },
            GraphCommit { hash: "L".to_string(), parents: vec!["K".to_string()] },
            GraphCommit { hash: "K".to_string(), parents: vec!["D".to_string()] },
            GraphCommit { hash: "J".to_string(), parents: vec!["H".to_string(), "I".to_string()] },
            GraphCommit { hash: "I".to_string(), parents: vec!["G".to_string()] },
            GraphCommit { hash: "H".to_string(), parents: vec!["F".to_string(), "G".to_string()] },
            GraphCommit { hash: "G".to_string(), parents: vec!["F".to_string()] },
            GraphCommit { hash: "F".to_string(), parents: vec!["D".to_string(), "E".to_string()] },
            GraphCommit { hash: "E".to_string(), parents: vec!["D".to_string()] },
            GraphCommit { hash: "D".to_string(), parents: vec!["B".to_string(), "C".to_string()] },
            GraphCommit { hash: "C".to_string(), parents: vec!["B".to_string()] },
            GraphCommit { hash: "B".to_string(), parents: vec!["A".to_string()] },
            GraphCommit { hash: "A".to_string(), parents: vec![] },
        ];

        let layout = compute_layout(&commits);
        let rendered = render_layout(&layout);

        assert_eq!(rendered, vec![
            "├─╮",
            "│ ├",
            "│ ├",
            "├─│─╮",
            "│ │ ├",
            "├─│─│─╮",
            "│ │ ├─╯",
            "├─│─┤",
            "│ │ ├",
            "├─┼─╯",
            "│ ├",
            "├─╯",
            "├",
        ]);
    }
}
