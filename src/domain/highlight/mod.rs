use tree_sitter_highlight::{Highlight, HighlightConfiguration, HighlightEvent, Highlighter};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum HighlightKind {
    Keyword,
    String,
    Comment,
    Type,
    Function,
    Constant,
    Variable,
    Number,
    Operator,
    Property,
    Punctuation,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct TokenSpan {
    pub start_col: usize,
    pub end_col: usize,
    pub kind: HighlightKind,
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct HighlightedFile {
    pub lines: Vec<Vec<TokenSpan>>,
}

impl HighlightedFile {
    pub fn line_tokens(&self, line_number: u32) -> Option<&[TokenSpan]> {
        if line_number == 0 {
            return None;
        }
        self.lines
            .get((line_number - 1) as usize)
            .map(Vec::as_slice)
    }
}

#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct DiffHighlights {
    pub old: HighlightedFile,
    pub new: HighlightedFile,
}

pub fn highlight_diff_sides(path: &str, old_content: &str, new_content: &str) -> DiffHighlights {
    let Some(language) = detect_language(path) else {
        return DiffHighlights::default();
    };

    let old = highlight_file(language, old_content).unwrap_or_default();
    let new = highlight_file(language, new_content).unwrap_or_default();

    DiffHighlights { old, new }
}

#[derive(Debug, Clone, Copy)]
enum SupportedLanguage {
    Rust,
    JavaScript,
    TypeScript,
    Tsx,
    Python,
    Go,
    Json,
    Yaml,
    Toml,
    C,
    Cpp,
    CSharp,
    Java,
    Kotlin,
    Swift,
    Php,
    Ruby,
    Lua,
    Bash,
    Html,
    Css,
    Scss,
    Vue,
    Svelte,
    Hcl,
    Makefile,
    Markdown,
}

fn detect_language(path: &str) -> Option<SupportedLanguage> {
    let file_name = path.rsplit('/').next().unwrap_or(path);
    if file_name.eq_ignore_ascii_case("dockerfile") {
        return Some(SupportedLanguage::Bash);
    }
    if file_name.eq_ignore_ascii_case("makefile") {
        return Some(SupportedLanguage::Makefile);
    }

    let ext = path.rsplit('.').next()?.to_ascii_lowercase();
    match ext.as_str() {
        "rs" => Some(SupportedLanguage::Rust),
        "c" | "h" => Some(SupportedLanguage::C),
        "cc" | "cpp" | "cxx" | "hpp" | "hh" | "hxx" => Some(SupportedLanguage::Cpp),
        "cs" => Some(SupportedLanguage::CSharp),
        "java" => Some(SupportedLanguage::Java),
        "kt" | "kts" => Some(SupportedLanguage::Kotlin),
        "swift" => Some(SupportedLanguage::Swift),
        "js" | "mjs" | "cjs" => Some(SupportedLanguage::JavaScript),
        "ts" => Some(SupportedLanguage::TypeScript),
        "tsx" => Some(SupportedLanguage::Tsx),
        "php" => Some(SupportedLanguage::Php),
        "rb" => Some(SupportedLanguage::Ruby),
        "lua" => Some(SupportedLanguage::Lua),
        "sh" | "bash" | "zsh" => Some(SupportedLanguage::Bash),
        "html" | "htm" => Some(SupportedLanguage::Html),
        "css" => Some(SupportedLanguage::Css),
        "scss" => Some(SupportedLanguage::Scss),
        "vue" => Some(SupportedLanguage::Vue),
        "svelte" => Some(SupportedLanguage::Svelte),
        "hcl" | "tf" | "tfvars" => Some(SupportedLanguage::Hcl),
        "mk" | "mak" => Some(SupportedLanguage::Makefile),
        "md" | "markdown" => Some(SupportedLanguage::Markdown),
        "py" => Some(SupportedLanguage::Python),
        "go" => Some(SupportedLanguage::Go),
        "json" => Some(SupportedLanguage::Json),
        "yaml" | "yml" => Some(SupportedLanguage::Yaml),
        "toml" => Some(SupportedLanguage::Toml),
        _ => None,
    }
}

fn highlight_file(language: SupportedLanguage, content: &str) -> Result<HighlightedFile, String> {
    let mut config = language.highlight_config()?;
    config.configure(HIGHLIGHT_NAMES);

    let mut highlighter = Highlighter::new();
    let mut line_spans: Vec<Vec<TokenSpan>> = vec![Vec::new(); content.lines().count().max(1)];
    let mut stack: Vec<Highlight> = Vec::new();

    let line_starts = line_start_bytes(content);
    let events = highlighter
        .highlight(&config, content.as_bytes(), None, |_| None)
        .map_err(|e| format!("Failed to highlight: {e}"))?;

    for event in events {
        match event.map_err(|e| format!("Failed to process highlight event: {e}"))? {
            HighlightEvent::HighlightStart(highlight) => stack.push(highlight),
            HighlightEvent::HighlightEnd => {
                stack.pop();
            }
            HighlightEvent::Source { start, end } => {
                let Some(active) = stack.last().copied() else {
                    continue;
                };
                let Some(kind) = capture_name_to_kind(active.0) else {
                    continue;
                };

                for (line_idx, start_col, end_col) in
                    split_byte_range_by_line(content, &line_starts, start, end)
                {
                    if start_col < end_col {
                        if let Some(line) = line_spans.get_mut(line_idx) {
                            line.push(TokenSpan {
                                start_col,
                                end_col,
                                kind,
                            });
                        }
                    }
                }
            }
        }
    }

    Ok(HighlightedFile { lines: line_spans })
}

fn line_start_bytes(content: &str) -> Vec<usize> {
    let mut starts = vec![0];
    for (idx, byte) in content.as_bytes().iter().enumerate() {
        if *byte == b'\n' {
            starts.push(idx + 1);
        }
    }
    starts
}

fn split_byte_range_by_line(
    content: &str,
    line_starts: &[usize],
    start: usize,
    end: usize,
) -> Vec<(usize, usize, usize)> {
    let mut result = Vec::new();
    if start >= end || content.is_empty() {
        return result;
    }

    let mut current_start = start;
    while current_start < end {
        let line_idx = line_index_for_offset(line_starts, current_start);
        let line_start = line_starts[line_idx];
        let line_end_exclusive = line_starts
            .get(line_idx + 1)
            .copied()
            .unwrap_or(content.len());
        let line_content_end = if line_end_exclusive > line_start
            && content.as_bytes().get(line_end_exclusive - 1) == Some(&b'\n')
        {
            line_end_exclusive - 1
        } else {
            line_end_exclusive
        };

        let current_end = end.min(line_content_end);
        if current_start < current_end {
            let line_text = &content[line_start..line_content_end];
            let start_col = byte_offset_to_char_index(line_text, current_start - line_start);
            let end_col = byte_offset_to_char_index(line_text, current_end - line_start);
            result.push((line_idx, start_col, end_col));
        }

        current_start = line_end_exclusive;
    }

    result
}

fn line_index_for_offset(line_starts: &[usize], offset: usize) -> usize {
    let idx = line_starts.partition_point(|value| *value <= offset);
    idx.saturating_sub(1)
}

fn byte_offset_to_char_index(line: &str, byte_offset: usize) -> usize {
    let mut chars = 0;
    for (idx, _) in line.char_indices() {
        if idx >= byte_offset {
            break;
        }
        chars += 1;
    }
    if byte_offset >= line.len() {
        line.chars().count()
    } else {
        chars
    }
}

fn capture_name_to_kind(index: usize) -> Option<HighlightKind> {
    let name = *HIGHLIGHT_NAMES.get(index)?;
    let kind = if name.starts_with("keyword") {
        HighlightKind::Keyword
    } else if name.starts_with("string") {
        HighlightKind::String
    } else if name.starts_with("comment") {
        HighlightKind::Comment
    } else if name.starts_with("type") || name.starts_with("tag") {
        HighlightKind::Type
    } else if name.starts_with("function") || name.starts_with("constructor") {
        HighlightKind::Function
    } else if name.starts_with("constant") || name == "boolean" {
        HighlightKind::Constant
    } else if name.starts_with("variable") || name.starts_with("parameter") {
        HighlightKind::Variable
    } else if name.starts_with("number") || name == "float" {
        HighlightKind::Number
    } else if name.starts_with("operator") {
        HighlightKind::Operator
    } else if name.starts_with("property") {
        HighlightKind::Property
    } else if name.starts_with("punctuation") {
        HighlightKind::Punctuation
    } else {
        return None;
    };
    Some(kind)
}

impl SupportedLanguage {
    fn highlight_config(self) -> Result<HighlightConfiguration, String> {
        match self {
            SupportedLanguage::Rust => HighlightConfiguration::new(
                tree_sitter_rust::LANGUAGE.into(),
                "rust",
                tree_sitter_rust::HIGHLIGHTS_QUERY,
                tree_sitter_rust::INJECTIONS_QUERY,
                "",
            )
            .map_err(|e| format!("Failed to create rust highlight config: {e}")),
            SupportedLanguage::JavaScript => HighlightConfiguration::new(
                tree_sitter_javascript::LANGUAGE.into(),
                "javascript",
                tree_sitter_javascript::HIGHLIGHT_QUERY,
                tree_sitter_javascript::INJECTIONS_QUERY,
                tree_sitter_javascript::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create javascript highlight config: {e}")),
            SupportedLanguage::TypeScript => HighlightConfiguration::new(
                tree_sitter_typescript::LANGUAGE_TYPESCRIPT.into(),
                "typescript",
                tree_sitter_typescript::HIGHLIGHTS_QUERY,
                "",
                tree_sitter_typescript::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create typescript highlight config: {e}")),
            SupportedLanguage::Tsx => HighlightConfiguration::new(
                tree_sitter_typescript::LANGUAGE_TSX.into(),
                "tsx",
                tree_sitter_typescript::HIGHLIGHTS_QUERY,
                "",
                tree_sitter_typescript::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create tsx highlight config: {e}")),
            SupportedLanguage::Python => HighlightConfiguration::new(
                tree_sitter_python::LANGUAGE.into(),
                "python",
                tree_sitter_python::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create python highlight config: {e}")),
            SupportedLanguage::Go => HighlightConfiguration::new(
                tree_sitter_go::LANGUAGE.into(),
                "go",
                tree_sitter_go::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create go highlight config: {e}")),
            SupportedLanguage::Json => HighlightConfiguration::new(
                tree_sitter_json::LANGUAGE.into(),
                "json",
                tree_sitter_json::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create json highlight config: {e}")),
            SupportedLanguage::Yaml => HighlightConfiguration::new(
                tree_sitter_yaml::LANGUAGE.into(),
                "yaml",
                tree_sitter_yaml::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create yaml highlight config: {e}")),
            SupportedLanguage::Toml => HighlightConfiguration::new(
                tree_sitter_toml_ng::LANGUAGE.into(),
                "toml",
                tree_sitter_toml_ng::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create toml highlight config: {e}")),
            SupportedLanguage::C => HighlightConfiguration::new(
                tree_sitter_c::LANGUAGE.into(),
                "c",
                tree_sitter_c::HIGHLIGHT_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create c highlight config: {e}")),
            SupportedLanguage::Cpp => HighlightConfiguration::new(
                tree_sitter_cpp::LANGUAGE.into(),
                "cpp",
                tree_sitter_cpp::HIGHLIGHT_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create cpp highlight config: {e}")),
            SupportedLanguage::CSharp => HighlightConfiguration::new(
                tree_sitter_c_sharp::LANGUAGE.into(),
                "c_sharp",
                "",
                "",
                "",
            )
            .map_err(|e| format!("Failed to create csharp highlight config: {e}")),
            SupportedLanguage::Java => HighlightConfiguration::new(
                tree_sitter_java::LANGUAGE.into(),
                "java",
                tree_sitter_java::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create java highlight config: {e}")),
            SupportedLanguage::Kotlin => HighlightConfiguration::new(
                tree_sitter_kotlin_sg::LANGUAGE.into(),
                "kotlin",
                tree_sitter_kotlin_sg::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create kotlin highlight config: {e}")),
            SupportedLanguage::Swift => HighlightConfiguration::new(
                tree_sitter_swift::LANGUAGE.into(),
                "swift",
                tree_sitter_swift::HIGHLIGHTS_QUERY,
                tree_sitter_swift::INJECTIONS_QUERY,
                tree_sitter_swift::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create swift highlight config: {e}")),
            SupportedLanguage::Php => HighlightConfiguration::new(
                tree_sitter_php::LANGUAGE_PHP.into(),
                "php",
                tree_sitter_php::HIGHLIGHTS_QUERY,
                tree_sitter_php::INJECTIONS_QUERY,
                "",
            )
            .map_err(|e| format!("Failed to create php highlight config: {e}")),
            SupportedLanguage::Ruby => HighlightConfiguration::new(
                tree_sitter_ruby::LANGUAGE.into(),
                "ruby",
                tree_sitter_ruby::HIGHLIGHTS_QUERY,
                "",
                tree_sitter_ruby::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create ruby highlight config: {e}")),
            SupportedLanguage::Lua => HighlightConfiguration::new(
                tree_sitter_lua::LANGUAGE.into(),
                "lua",
                tree_sitter_lua::HIGHLIGHTS_QUERY,
                tree_sitter_lua::INJECTIONS_QUERY,
                tree_sitter_lua::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create lua highlight config: {e}")),
            SupportedLanguage::Bash => HighlightConfiguration::new(
                tree_sitter_bash::LANGUAGE.into(),
                "bash",
                tree_sitter_bash::HIGHLIGHT_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create bash highlight config: {e}")),
            SupportedLanguage::Html => HighlightConfiguration::new(
                tree_sitter_html::LANGUAGE.into(),
                "html",
                tree_sitter_html::HIGHLIGHTS_QUERY,
                tree_sitter_html::INJECTIONS_QUERY,
                "",
            )
            .map_err(|e| format!("Failed to create html highlight config: {e}")),
            SupportedLanguage::Css | SupportedLanguage::Scss => HighlightConfiguration::new(
                tree_sitter_css::LANGUAGE.into(),
                "css",
                tree_sitter_css::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create css highlight config: {e}")),
            SupportedLanguage::Vue => HighlightConfiguration::new(
                tree_sitter_vue3::LANGUAGE.into(),
                "vue",
                tree_sitter_vue3::HIGHLIGHTS_QUERY,
                tree_sitter_vue3::INJECTIONS_QUERY,
                "",
            )
            .map_err(|e| format!("Failed to create vue highlight config: {e}")),
            SupportedLanguage::Svelte => HighlightConfiguration::new(
                tree_sitter_svelte_next::LANGUAGE.into(),
                "svelte",
                tree_sitter_svelte_next::HIGHLIGHTS_QUERY,
                tree_sitter_svelte_next::INJECTIONS_QUERY,
                tree_sitter_svelte_next::LOCALS_QUERY,
            )
            .map_err(|e| format!("Failed to create svelte highlight config: {e}")),
            SupportedLanguage::Hcl => {
                HighlightConfiguration::new(tree_sitter_hcl::LANGUAGE.into(), "hcl", "", "", "")
                    .map_err(|e| format!("Failed to create hcl highlight config: {e}"))
            }
            SupportedLanguage::Makefile => HighlightConfiguration::new(
                tree_sitter_make::LANGUAGE.into(),
                "make",
                tree_sitter_make::HIGHLIGHTS_QUERY,
                "",
                "",
            )
            .map_err(|e| format!("Failed to create make highlight config: {e}")),
            SupportedLanguage::Markdown => HighlightConfiguration::new(
                tree_sitter_md::LANGUAGE.into(),
                "markdown",
                tree_sitter_md::HIGHLIGHT_QUERY_BLOCK,
                tree_sitter_md::INJECTION_QUERY_BLOCK,
                "",
            )
            .map_err(|e| format!("Failed to create markdown highlight config: {e}")),
        }
    }
}

const HIGHLIGHT_NAMES: &[&str] = &[
    "attribute",
    "boolean",
    "comment",
    "constant",
    "constant.builtin",
    "constructor",
    "embedded",
    "function",
    "function.builtin",
    "function.call",
    "keyword",
    "number",
    "operator",
    "property",
    "punctuation",
    "punctuation.bracket",
    "punctuation.delimiter",
    "string",
    "string.escape",
    "string.special",
    "tag",
    "type",
    "type.builtin",
    "variable",
    "variable.builtin",
    "variable.member",
    "variable.parameter",
];

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn highlights_rust_keywords_and_strings() {
        let highlights = highlight_diff_sides(
            "src/lib.rs",
            "fn old() { let a = 1; }",
            "fn new_fn() { let s = \"x\"; }",
        );

        let line_tokens = highlights.new.line_tokens(1).unwrap_or(&[]);
        assert!(
            line_tokens
                .iter()
                .any(|span| span.kind == HighlightKind::Keyword),
            "Expected keyword token in highlighted rust line"
        );
        assert!(
            line_tokens
                .iter()
                .any(|span| span.kind == HighlightKind::String),
            "Expected string token in highlighted rust line"
        );
    }

    #[test]
    fn returns_empty_for_unknown_extension() {
        let highlights = highlight_diff_sides("notes.txt", "a", "b");
        assert!(highlights.old.lines.is_empty());
        assert!(highlights.new.lines.is_empty());
    }
}
