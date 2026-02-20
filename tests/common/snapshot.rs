pub fn assert_snapshot(actual: &str, expected: &str) {
    let expected = normalize_snapshot(expected);
    let actual = normalize_actual(actual);

    if actual == expected {
        return;
    }

    panic!(
        "Snapshot mismatch\n\nExpected:\n{}\n\nActual:\n{}\n\nIf this is correct, update the inline snapshot.",
        expected,
        actual
    );
}

fn normalize_actual(raw: &str) -> String {
    raw.lines()
        .map(|line| line.trim_end())
        .collect::<Vec<_>>()
        .join("\n")
}

fn normalize_snapshot(raw: &str) -> String {
    let lines: Vec<&str> = raw.lines().collect();
    if lines.is_empty() {
        return String::new();
    }

    let mut start = 0;
    let mut end = lines.len();

    if lines[start].trim().is_empty() {
        start += 1;
    }

    while end > start && lines[end - 1].trim().is_empty() {
        end -= 1;
    }

    let trimmed = &lines[start..end];
    let indent = trimmed
        .iter()
        .filter(|line| !line.trim().is_empty())
        .map(|line| line.chars().take_while(|c| c.is_whitespace()).count())
        .min()
        .unwrap_or(0);

    trimmed
        .iter()
        .map(|line| {
            let line = if line.len() >= indent {
                &line[indent..]
            } else {
                *line
            };
            line.trim_end()
        })
        .collect::<Vec<_>>()
        .join("\n")
}
