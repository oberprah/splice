fn find_in_lanes(hash: &str, lanes: &[String]) -> Option<usize> {
    lanes.iter().position(|h| h == hash)
}

fn find_empty_lane(lanes: &[String]) -> Option<usize> {
    lanes.iter().position(|h| h.is_empty())
}

pub fn assign_column(hash: &str, lanes: &mut Vec<String>) -> usize {
    if let Some(idx) = find_in_lanes(hash, lanes) {
        return idx;
    }

    if let Some(idx) = find_empty_lane(lanes) {
        lanes[idx] = hash.to_string();
        return idx;
    }

    lanes.push(hash.to_string());
    lanes.len() - 1
}

pub struct UpdateResult {
    pub lanes: Vec<String>,
    pub merge_columns: Vec<usize>,
    pub existing_lanes_merge: Vec<usize>,
    pub converges_to_parent: bool,
}

pub fn update_lanes(col: usize, parents: &[String], lanes: Vec<String>) -> UpdateResult {
    let mut lanes = lanes;
    let mut merge_columns = Vec::new();
    let mut existing_lanes_merge = Vec::new();
    let mut converges_to_parent = false;

    if parents.is_empty() {
        if col < lanes.len() {
            lanes[col].clear();
        }
        return UpdateResult {
            lanes,
            merge_columns,
            existing_lanes_merge,
            converges_to_parent,
        };
    }

    let first_parent = &parents[0];
    if col < lanes.len() {
        if let Some(existing_col) = find_in_lanes(first_parent, &lanes) {
            if existing_col != col {
                if parents.len() == 1 {
                    let mut future_parent_count = 0;
                    for (i, hash) in lanes.iter().enumerate() {
                        if hash == first_parent {
                            future_parent_count += 1;
                        } else if i == col {
                            future_parent_count += 1;
                        }
                    }
                    if future_parent_count > 1 {
                        lanes[col] = first_parent.clone();
                    } else {
                        lanes[col].clear();
                        converges_to_parent = true;
                    }
                } else {
                    lanes[col] = first_parent.clone();
                }
            } else {
                lanes[col] = first_parent.clone();
            }
        } else {
            lanes[col] = first_parent.clone();
        }
    }

    for i in 1..parents.len() {
        let merge_parent = &parents[i];

        if let Some(existing_col) = find_in_lanes(merge_parent, &lanes) {
            merge_columns.push(existing_col);
            existing_lanes_merge.push(existing_col);
            continue;
        }

        let mut placed = false;
        for j in (col + 1)..lanes.len() {
            if lanes[j].is_empty() {
                lanes[j] = merge_parent.clone();
                merge_columns.push(j);
                placed = true;
                break;
            }
        }

        if !placed {
            lanes.push(merge_parent.clone());
            merge_columns.push(lanes.len() - 1);
        }
    }

    UpdateResult {
        lanes,
        merge_columns,
        existing_lanes_merge,
        converges_to_parent,
    }
}

pub fn collapse_trailing_empty(lanes: Vec<String>) -> Vec<String> {
    let last_non_empty = lanes.iter().rposition(|h| !h.is_empty());
    
    match last_non_empty {
        Some(idx) => lanes.into_iter().take(idx + 1).collect(),
        None => Vec::new(),
    }
}

pub fn detect_converging_columns(commit_col: usize, commit_hash: &str, lanes: &[String]) -> Vec<usize> {
    lanes.iter()
        .enumerate()
        .filter(|(i, hash)| *i != commit_col && *hash == commit_hash)
        .map(|(i, _)| i)
        .collect()
}

pub fn detect_passing_columns(
    commit_col: usize,
    lanes: &[String],
    merge_columns: &[usize],
    converging_columns: &[usize],
) -> Vec<usize> {
    let exclude: std::collections::HashSet<usize> = 
        [commit_col].iter()
            .chain(merge_columns.iter())
            .chain(converging_columns.iter())
            .copied()
            .collect();

    lanes.iter()
        .enumerate()
        .filter(|(i, hash)| !exclude.contains(i) && !hash.is_empty())
        .map(|(i, _)| i)
        .collect()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_find_in_lanes() {
        let lanes = vec!["A".to_string(), "B".to_string(), "".to_string()];
        assert_eq!(find_in_lanes("A", &lanes), Some(0));
        assert_eq!(find_in_lanes("B", &lanes), Some(1));
        assert_eq!(find_in_lanes("C", &lanes), None);
    }

    #[test]
    fn test_find_empty_lane() {
        let lanes = vec!["A".to_string(), "".to_string(), "B".to_string()];
        assert_eq!(find_empty_lane(&lanes), Some(1));
        
        let full_lanes = vec!["A".to_string(), "B".to_string()];
        assert_eq!(find_empty_lane(&full_lanes), None);
    }

    #[test]
    fn test_assign_column_existing() {
        let mut lanes = vec!["A".to_string(), "B".to_string()];
        let col = assign_column("A", &mut lanes);
        assert_eq!(col, 0);
        assert_eq!(lanes, vec!["A", "B"]);
    }

    #[test]
    fn test_assign_column_empty_slot() {
        let mut lanes = vec!["A".to_string(), "".to_string(), "B".to_string()];
        let col = assign_column("C", &mut lanes);
        assert_eq!(col, 1);
        assert_eq!(lanes, vec!["A", "C", "B"]);
    }

    #[test]
    fn test_assign_column_new() {
        let mut lanes = vec!["A".to_string()];
        let col = assign_column("B", &mut lanes);
        assert_eq!(col, 1);
        assert_eq!(lanes, vec!["A", "B"]);
    }

    #[test]
    fn test_update_lanes_root_commit() {
        let lanes = vec!["A".to_string()];
        let result = update_lanes(0, &[], lanes);
        assert_eq!(result.lanes, vec![""]);
        assert!(result.merge_columns.is_empty());
        assert!(!result.converges_to_parent);
    }

    #[test]
    fn test_collapse_trailing_empty() {
        let lanes = vec!["A".to_string(), "".to_string(), "".to_string()];
        assert_eq!(collapse_trailing_empty(lanes), vec!["A"]);

        let lanes = vec!["A".to_string(), "B".to_string()];
        assert_eq!(collapse_trailing_empty(lanes), vec!["A", "B"]);

        let lanes: Vec<String> = vec!["".to_string(), "".to_string()];
        assert!(collapse_trailing_empty(lanes).is_empty());
    }
}
