use crate::core::FileChange;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub struct FolderStats {
    pub file_count: usize,
    pub additions: u32,
    pub deletions: u32,
}

#[derive(Debug, Clone)]
pub struct FolderNode {
    pub name: String,
    pub depth: i32,
    pub children: Vec<TreeNode>,
    pub is_expanded: bool,
    pub stats: FolderStats,
}

#[derive(Debug, Clone)]
pub struct FileNode {
    pub name: String,
    pub depth: i32,
    pub file: FileChange,
}

#[derive(Debug, Clone)]
pub enum TreeNode {
    Folder(FolderNode),
    File(FileNode),
}

impl TreeNode {
    pub fn name(&self) -> &str {
        match self {
            TreeNode::Folder(f) => &f.name,
            TreeNode::File(f) => &f.name,
        }
    }

    pub fn depth(&self) -> i32 {
        match self {
            TreeNode::Folder(f) => f.depth,
            TreeNode::File(f) => f.depth,
        }
    }

    pub fn is_folder(&self) -> bool {
        matches!(self, TreeNode::Folder(_))
    }
}

#[derive(Debug, Clone)]
pub struct VisibleTreeItem {
    pub node: TreeNode,
    pub is_last_child: bool,
    pub parent_lines: Vec<bool>,
}

pub fn build_tree(files: &[FileChange]) -> TreeNode {
    let mut root = FolderNode {
        name: String::new(),
        depth: -1,
        children: Vec::new(),
        is_expanded: true,
        stats: FolderStats::default(),
    };

    for file in files {
        insert_file(&mut root, file.clone());
    }

    sort_children(&mut root);
    TreeNode::Folder(root)
}

fn insert_file(root: &mut FolderNode, file: FileChange) {
    let parts: Vec<&str> = file.path.split('/').collect();
    let mut current = root;

    for part in parts.iter().take(parts.len() - 1) {
        current = get_or_create_folder(current, part);
    }

    let filename = parts[parts.len() - 1].to_string();
    let file_node = FileNode {
        name: filename,
        depth: current.depth + 1,
        file,
    };
    current.children.push(TreeNode::File(file_node));
}

fn get_or_create_folder<'a>(parent: &'a mut FolderNode, name: &str) -> &'a mut FolderNode {
    let existing_index = parent.children.iter().position(|c| {
        if let TreeNode::Folder(f) = c {
            f.name == name
        } else {
            false
        }
    });

    if let Some(idx) = existing_index {
        if let TreeNode::Folder(ref mut f) = parent.children[idx] {
            return f;
        }
        unreachable!()
    }

    let folder = FolderNode {
        name: name.to_string(),
        depth: parent.depth + 1,
        children: Vec::new(),
        is_expanded: true,
        stats: FolderStats::default(),
    };
    parent.children.push(TreeNode::Folder(folder));

    let last_idx = parent.children.len() - 1;
    if let TreeNode::Folder(ref mut f) = parent.children[last_idx] {
        f
    } else {
        unreachable!()
    }
}

fn sort_children(node: &mut FolderNode) {
    node.children.sort_by(|a, b| {
        let a_is_folder = matches!(a, TreeNode::Folder(_));
        let b_is_folder = matches!(b, TreeNode::Folder(_));

        if a_is_folder != b_is_folder {
            return b_is_folder.cmp(&a_is_folder);
        }

        a.name().cmp(b.name())
    });

    for child in &mut node.children {
        if let TreeNode::Folder(ref mut f) = child {
            sort_children(f);
        }
    }
}

pub fn collapse_paths(root: TreeNode) -> TreeNode {
    let TreeNode::Folder(mut folder) = root else {
        return root;
    };
    collapse_folder(&mut folder);
    TreeNode::Folder(folder)
}

fn collapse_folder(folder: &mut FolderNode) {
    for child in &mut folder.children {
        if let TreeNode::Folder(ref mut f) = child {
            collapse_folder(f);
        }
    }

    for child in &mut folder.children {
        let TreeNode::Folder(ref mut child_folder) = child else {
            continue;
        };

        let (collapsed_name, collapsed_children, preserved_expanded) = collapse_chain(child_folder);

        *child_folder = FolderNode {
            name: collapsed_name,
            depth: child_folder.depth,
            children: collapsed_children,
            is_expanded: preserved_expanded,
            stats: FolderStats::default(),
        };

        adjust_depths_folder(child_folder, child_folder.depth);
    }
}

fn collapse_chain(folder: &FolderNode) -> (String, Vec<TreeNode>, bool) {
    let mut path_components = vec![folder.name.clone()];
    let mut current = folder;
    let is_expanded = folder.is_expanded;

    while current.children.len() == 1 {
        if let TreeNode::Folder(ref child) = current.children[0] {
            path_components.push(child.name.clone());
            current = child;
        } else {
            break;
        }
    }

    let name = path_components.join("/");
    let children = current.children.clone();

    (name, children, is_expanded)
}

fn adjust_depths_folder(folder: &mut FolderNode, new_depth: i32) {
    folder.depth = new_depth;
    for child in &mut folder.children {
        adjust_depths(child, new_depth + 1);
    }
}

fn adjust_depths(node: &mut TreeNode, new_depth: i32) {
    match node {
        TreeNode::Folder(f) => {
            f.depth = new_depth;
            for child in &mut f.children {
                adjust_depths(child, new_depth + 1);
            }
        }
        TreeNode::File(f) => {
            f.depth = new_depth;
        }
    }
}

pub fn compute_stats(node: &TreeNode) -> FolderStats {
    match node {
        TreeNode::File(f) => FolderStats {
            file_count: 1,
            additions: f.file.additions,
            deletions: f.file.deletions,
        },
        TreeNode::Folder(f) => {
            let mut total = FolderStats::default();
            for child in &f.children {
                let child_stats = compute_stats(child);
                total.file_count += child_stats.file_count;
                total.additions += child_stats.additions;
                total.deletions += child_stats.deletions;
            }
            total
        }
    }
}

pub fn apply_stats(node: &mut TreeNode) {
    let TreeNode::Folder(ref mut folder) = node else {
        return;
    };
    for child in &mut folder.children {
        apply_stats(child);
    }
    let stats = compute_stats_folder(folder);
    folder.stats = stats;
}

fn compute_stats_folder(folder: &FolderNode) -> FolderStats {
    let mut total = FolderStats::default();
    for child in &folder.children {
        let child_stats = compute_stats(child);
        total.file_count += child_stats.file_count;
        total.additions += child_stats.additions;
        total.deletions += child_stats.deletions;
    }
    total
}

pub fn flatten_visible(root: &TreeNode) -> Vec<VisibleTreeItem> {
    let mut result = Vec::new();

    let TreeNode::Folder(folder) = root else {
        return result;
    };

    for (i, child) in folder.children.iter().enumerate() {
        let is_last_child = i == folder.children.len() - 1;
        let parent_lines = Vec::new();
        walk(child, is_last_child, &parent_lines, &mut result);
    }

    result
}

fn walk(
    node: &TreeNode,
    is_last_child: bool,
    parent_lines: &[bool],
    result: &mut Vec<VisibleTreeItem>,
) {
    result.push(VisibleTreeItem {
        node: node.clone(),
        is_last_child,
        parent_lines: parent_lines.to_vec(),
    });

    let TreeNode::Folder(folder) = node else {
        return;
    };

    if !folder.is_expanded {
        return;
    }

    let mut child_parent_lines = parent_lines.to_vec();
    child_parent_lines.push(!is_last_child);

    for (i, child) in folder.children.iter().enumerate() {
        let child_is_last = i == folder.children.len() - 1;
        walk(child, child_is_last, &child_parent_lines, result);
    }
}

pub fn deep_copy(node: &TreeNode) -> TreeNode {
    node.clone()
}

pub fn toggle_folder_at_cursor(
    visible_items: &[VisibleTreeItem],
    cursor: usize,
    expand_only: bool,
    collapse_only: bool,
) -> Option<(Vec<VisibleTreeItem>, usize)> {
    let item = visible_items.get(cursor)?;
    let TreeNode::Folder(folder) = &item.node else {
        return None;
    };

    if expand_only && folder.is_expanded {
        return None;
    }
    if collapse_only && !folder.is_expanded {
        return None;
    }

    let mut new_visible = visible_items.to_vec();
    let target = new_visible.get_mut(cursor)?;
    let TreeNode::Folder(ref mut target_folder) = target.node else {
        return None;
    };

    let was_expanded = target_folder.is_expanded;
    target_folder.is_expanded = !was_expanded;

    if was_expanded {
        let depth = new_visible[cursor].parent_lines.len();
        let mut i = cursor + 1;
        while i < new_visible.len() {
            let item = &new_visible[i];
            if item.parent_lines.len() <= depth {
                break;
            }
            i += 1;
        }
        new_visible.drain(cursor + 1..i);
    } else {
        let TreeNode::Folder(ref folder) = new_visible[cursor].node else {
            unreachable!()
        };

        let mut insert_items = Vec::new();
        let mut child_parent_lines = new_visible[cursor].parent_lines.clone();
        child_parent_lines.push(!new_visible[cursor].is_last_child);

        for (i, child) in folder.children.iter().enumerate() {
            let child_is_last = i == folder.children.len() - 1;
            collect_visible_items_recursive(
                child,
                child_is_last,
                &child_parent_lines,
                &mut insert_items,
            );
        }

        new_visible.splice(cursor + 1..cursor + 1, insert_items);
    }

    let new_cursor = if cursor >= new_visible.len() {
        new_visible.len().saturating_sub(1)
    } else {
        cursor
    };

    Some((new_visible, new_cursor))
}

fn collect_visible_items_recursive(
    node: &TreeNode,
    is_last_child: bool,
    parent_lines: &[bool],
    result: &mut Vec<VisibleTreeItem>,
) {
    result.push(VisibleTreeItem {
        node: node.clone(),
        is_last_child,
        parent_lines: parent_lines.to_vec(),
    });

    if let TreeNode::Folder(folder) = node {
        if folder.is_expanded {
            let mut child_parent_lines = parent_lines.to_vec();
            child_parent_lines.push(!is_last_child);

            for (i, child) in folder.children.iter().enumerate() {
                let child_is_last = i == folder.children.len() - 1;
                collect_visible_items_recursive(child, child_is_last, &child_parent_lines, result);
            }
        }
    }
}

pub fn build_visible_tree(files: &[FileChange]) -> (TreeNode, Vec<VisibleTreeItem>) {
    let mut root = build_tree(files);
    root = collapse_paths(root);
    apply_stats(&mut root);
    let visible = flatten_visible(&root);
    (root, visible)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::core::FileStatus;

    fn make_file(path: &str, additions: u32, deletions: u32) -> FileChange {
        FileChange {
            path: path.to_string(),
            status: FileStatus::Modified,
            additions,
            deletions,
            is_binary: false,
        }
    }

    #[test]
    fn test_build_tree_single_file() {
        let files = vec![make_file("src/main.rs", 10, 5)];
        let root = build_tree(&files);

        if let TreeNode::Folder(f) = root {
            assert_eq!(f.depth, -1);
            assert_eq!(f.children.len(), 1);

            if let TreeNode::Folder(src) = &f.children[0] {
                assert_eq!(src.name, "src");
                assert_eq!(src.depth, 0);
                assert_eq!(src.children.len(), 1);

                if let TreeNode::File(file) = &src.children[0] {
                    assert_eq!(file.name, "main.rs");
                    assert_eq!(file.depth, 1);
                } else {
                    panic!("Expected file node");
                }
            } else {
                panic!("Expected folder node");
            }
        } else {
            panic!("Expected folder node");
        }
    }

    #[test]
    fn test_build_tree_multiple_files() {
        let files = vec![
            make_file("src/main.rs", 10, 5),
            make_file("src/lib.rs", 20, 3),
            make_file("README.md", 5, 0),
        ];
        let root = build_tree(&files);

        if let TreeNode::Folder(f) = root {
            assert_eq!(f.children.len(), 2);

            let first = &f.children[0];
            let second = &f.children[1];

            assert!(matches!(first, TreeNode::Folder(_)));
            assert!(matches!(second, TreeNode::File(_)));

            if let TreeNode::Folder(src) = first {
                assert_eq!(src.name, "src");
                assert_eq!(src.children.len(), 2);
            }
            if let TreeNode::File(readme) = second {
                assert_eq!(readme.name, "README.md");
            }
        }
    }

    #[test]
    fn test_flatten_visible() {
        let files = vec![
            make_file("src/main.rs", 10, 5),
            make_file("src/lib.rs", 20, 3),
            make_file("README.md", 5, 0),
        ];
        let root = build_tree(&files);
        let visible = flatten_visible(&root);

        assert_eq!(visible.len(), 4);

        assert!(matches!(visible[0].node, TreeNode::Folder(_)));
        assert!(!visible[0].is_last_child);

        assert!(matches!(visible[1].node, TreeNode::File(_)));
        assert!(!visible[1].is_last_child);

        assert!(matches!(visible[2].node, TreeNode::File(_)));
        assert!(visible[2].is_last_child);

        assert!(matches!(visible[3].node, TreeNode::File(_)));
        assert!(visible[3].is_last_child);
    }

    #[test]
    fn test_collapse_paths() {
        let files = vec![make_file("src/components/ui/Button.tsx", 10, 5)];
        let root = build_tree(&files);
        let collapsed = collapse_paths(root);

        if let TreeNode::Folder(f) = collapsed {
            assert_eq!(f.children.len(), 1);

            if let TreeNode::Folder(src) = &f.children[0] {
                assert_eq!(src.name, "src/components/ui");
                assert_eq!(src.depth, 0);
            }
        }
    }

    #[test]
    fn test_compute_stats() {
        let files = vec![
            make_file("src/main.rs", 10, 5),
            make_file("src/lib.rs", 20, 3),
        ];
        let mut root = build_tree(&files);
        apply_stats(&mut root);

        if let TreeNode::Folder(f) = root {
            assert_eq!(f.stats.file_count, 2);
            assert_eq!(f.stats.additions, 30);
            assert_eq!(f.stats.deletions, 8);

            if let TreeNode::Folder(src) = &f.children[0] {
                assert_eq!(src.stats.file_count, 2);
                assert_eq!(src.stats.additions, 30);
                assert_eq!(src.stats.deletions, 8);
            }
        }
    }

    #[test]
    fn test_build_visible_tree() {
        let files = vec![
            make_file("src/main.rs", 10, 5),
            make_file("README.md", 5, 0),
        ];
        let (root, visible) = build_visible_tree(&files);

        assert!(matches!(root, TreeNode::Folder(_)));
        assert_eq!(visible.len(), 3);
    }
}
