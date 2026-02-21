#[derive(Debug, Clone, PartialEq, Eq)]
pub enum LogSpec {
    Head,
    All,
    Rev(String),
}
