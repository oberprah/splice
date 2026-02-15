use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum RefType {
    Branch,
    RemoteBranch,
    Tag,
    DetachedHead,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct RefInfo {
    pub name: String,
    pub ref_type: RefType,
    pub is_head: bool,
}

impl RefInfo {
    pub fn new(name: String, ref_type: RefType, is_head: bool) -> Self {
        Self {
            name,
            ref_type,
            is_head,
        }
    }

    pub fn branch(name: String) -> Self {
        Self::new(name, RefType::Branch, false)
    }

    pub fn remote_branch(name: String) -> Self {
        Self::new(name, RefType::RemoteBranch, false)
    }

    pub fn tag(name: String) -> Self {
        Self::new(name, RefType::Tag, false)
    }
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Commit {
    pub hash: String,
    pub parent_hashes: Vec<String>,
    pub refs: Vec<RefInfo>,
    pub message: String,
    pub author: String,
    pub date: DateTime<Utc>,
}

impl Commit {
    pub fn short_hash(&self) -> &str {
        if self.hash.len() >= 7 {
            &self.hash[..7]
        } else {
            &self.hash
        }
    }
}
