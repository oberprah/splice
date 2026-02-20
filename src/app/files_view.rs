use crate::core::{DiffSource, FileChange};
use crate::domain::filetree::{self, TreeNode, VisibleTreeItem};

pub struct FilesView {
    pub source: DiffSource,
    pub files: Vec<FileChange>,
    pub root: TreeNode,
    pub visible_items: Vec<VisibleTreeItem>,
    pub selected: usize,
    pub scroll_offset: usize,
    pub viewport_height: usize,
}

impl FilesView {
    pub fn new(source: DiffSource, files: Vec<FileChange>) -> Self {
        let (root, visible_items) = filetree::build_visible_tree(&files);

        Self {
            source,
            files,
            root,
            visible_items,
            selected: 0,
            scroll_offset: 0,
            viewport_height: 0,
        }
    }

    pub fn set_viewport_height(&mut self, height: usize) {
        self.viewport_height = height;
        self.clamp_scroll_offset();
    }

    fn clamp_scroll_offset(&mut self) {
        if self.visible_items.is_empty() {
            self.selected = 0;
            self.scroll_offset = 0;
            return;
        }
        if self.selected < self.scroll_offset {
            self.scroll_offset = self.selected;
        } else if self.viewport_height > 0
            && self.selected >= self.scroll_offset + self.viewport_height
        {
            self.scroll_offset = self.selected - self.viewport_height + 1;
        }
    }

    pub fn move_down(&mut self, amount: usize) {
        if self.visible_items.is_empty() {
            return;
        }
        let last = self.visible_items.len().saturating_sub(1);
        self.selected = self.selected.saturating_add(amount).min(last);
        self.clamp_scroll_offset();
    }

    pub fn move_up(&mut self, amount: usize) {
        if self.visible_items.is_empty() {
            return;
        }
        self.selected = self.selected.saturating_sub(amount);
        self.clamp_scroll_offset();
    }

    pub fn page_step(&self) -> usize {
        (self.viewport_height / 2).max(1)
    }

    pub fn selected_file(&self) -> Option<&FileChange> {
        let item = self.visible_items.get(self.selected)?;
        if let TreeNode::File(file_node) = &item.node {
            Some(&file_node.file)
        } else {
            None
        }
    }

    pub fn selected_is_folder(&self) -> bool {
        let Some(item) = self.visible_items.get(self.selected) else {
            return false;
        };
        matches!(item.node, TreeNode::Folder(_))
    }

    pub fn toggle_folder(&mut self, expand_only: bool, collapse_only: bool) -> bool {
        let Some((new_visible, new_cursor)) = filetree::toggle_folder_at_cursor(
            &self.visible_items,
            self.selected,
            expand_only,
            collapse_only,
        ) else {
            return false;
        };

        self.visible_items = new_visible;
        self.selected = new_cursor;
        self.clamp_scroll_offset();
        true
    }

    pub fn total_additions(&self) -> u32 {
        self.files.iter().map(|f| f.additions).sum()
    }

    pub fn total_deletions(&self) -> u32 {
        self.files.iter().map(|f| f.deletions).sum()
    }
}
