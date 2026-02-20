#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CursorState {
    Normal { pos: usize },
    Visual { pos: usize, anchor: usize },
}

impl CursorState {
    pub fn position(&self) -> usize {
        match self {
            CursorState::Normal { pos } => *pos,
            CursorState::Visual { pos, .. } => *pos,
        }
    }
}

pub fn selection_range(cursor: &CursorState) -> (usize, usize) {
    match cursor {
        CursorState::Normal { pos } => (*pos, *pos),
        CursorState::Visual { pos, anchor } => (*pos.min(anchor), *pos.max(anchor)),
    }
}

pub fn is_in_selection(cursor: &CursorState, index: usize) -> bool {
    let (min, max) = selection_range(cursor);
    index >= min && index <= max
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn position_returns_pos_for_normal() {
        let cursor = CursorState::Normal { pos: 5 };
        assert_eq!(cursor.position(), 5);
    }

    #[test]
    fn position_returns_pos_for_visual() {
        let cursor = CursorState::Visual { pos: 3, anchor: 7 };
        assert_eq!(cursor.position(), 3);
    }

    #[test]
    fn selection_range_normal_returns_pos_pos() {
        let cursor = CursorState::Normal { pos: 5 };
        assert_eq!(selection_range(&cursor), (5, 5));
    }

    #[test]
    fn selection_range_visual_returns_ordered_when_pos_less_than_anchor() {
        let cursor = CursorState::Visual { pos: 2, anchor: 8 };
        assert_eq!(selection_range(&cursor), (2, 8));
    }

    #[test]
    fn selection_range_visual_returns_ordered_when_pos_greater_than_anchor() {
        let cursor = CursorState::Visual { pos: 8, anchor: 2 };
        assert_eq!(selection_range(&cursor), (2, 8));
    }

    #[test]
    fn selection_range_visual_returns_ordered_when_pos_equals_anchor() {
        let cursor = CursorState::Visual { pos: 5, anchor: 5 };
        assert_eq!(selection_range(&cursor), (5, 5));
    }

    #[test]
    fn is_in_selection_normal_true_at_pos() {
        let cursor = CursorState::Normal { pos: 5 };
        assert!(is_in_selection(&cursor, 5));
    }

    #[test]
    fn is_in_selection_normal_false_not_at_pos() {
        let cursor = CursorState::Normal { pos: 5 };
        assert!(!is_in_selection(&cursor, 4));
        assert!(!is_in_selection(&cursor, 6));
    }

    #[test]
    fn is_in_selection_visual_true_within_range() {
        let cursor = CursorState::Visual { pos: 2, anchor: 5 };
        assert!(is_in_selection(&cursor, 2));
        assert!(is_in_selection(&cursor, 3));
        assert!(is_in_selection(&cursor, 4));
        assert!(is_in_selection(&cursor, 5));
    }

    #[test]
    fn is_in_selection_visual_false_outside_range() {
        let cursor = CursorState::Visual { pos: 2, anchor: 5 };
        assert!(!is_in_selection(&cursor, 1));
        assert!(!is_in_selection(&cursor, 6));
    }

    #[test]
    fn is_in_selection_visual_works_with_reversed_pos_anchor() {
        let cursor = CursorState::Visual { pos: 5, anchor: 2 };
        assert!(is_in_selection(&cursor, 2));
        assert!(is_in_selection(&cursor, 3));
        assert!(is_in_selection(&cursor, 4));
        assert!(is_in_selection(&cursor, 5));
        assert!(!is_in_selection(&cursor, 1));
        assert!(!is_in_selection(&cursor, 6));
    }
}
