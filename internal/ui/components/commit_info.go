package components

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/format"
	"github.com/oberprah/splice/internal/ui/styles"
)

// CommitInfoFromRange renders commit info for either a single commit or a range.
// For single commits, delegates to CommitInfo().
// For ranges, renders a header showing the range and commit count.
func CommitInfoFromRange(commitRange core.CommitRange, width int, bodyMaxLines int, ctx core.Context) []string {
	if commitRange.IsSingleCommit() {
		return CommitInfo(commitRange.End, width, bodyMaxLines, ctx)
	}

	// Range display - just header, no subject/body
	var lines []string

	// Format: abc123d..def456e (3 commits)
	startHash := format.ToShortHash(commitRange.Start.Hash)
	endHash := format.ToShortHash(commitRange.End.Hash)

	header := fmt.Sprintf("%s..%s (%d commits)", startHash, endHash, commitRange.Count)
	lines = append(lines, styles.HashStyle.Render(header))

	return lines
}

// CommitInfo renders complete commit information
// Returns lines to display (metadata, refs, blank, subject, blank, body with truncation indicator)
//
// Parameters:
//   - commit: The commit to display
//   - width: Panel width for wrapping/truncation
//   - bodyMaxLines: 0 for unlimited (files view), 5 for log detail view
//   - ctx: For time formatting
//
// Structure:
//   - Metadata line: {hash} · {author} committed {time}
//   - Refs line: main, origin/main, HEAD (if refs exist)
//   - (blank)
//   - Subject (wrapped if needed)
//   - (blank, if body exists)
//   - Body (wrapped, truncated to bodyMaxLines with indicator)
func CommitInfo(commit core.GitCommit, width int, bodyMaxLines int, ctx core.Context) []string {
	var lines []string

	// 1. Metadata line
	metadataLine := renderMetadataLine(commit, width, ctx)
	lines = append(lines, metadataLine)

	// 2. Refs line (if any refs exist)
	if len(commit.Refs) > 0 {
		refsLines := renderRefsLines(commit.Refs, width)
		lines = append(lines, refsLines...)
	}

	// 3. Blank line before subject
	lines = append(lines, "")

	// 4. Subject (wrapped if needed)
	subjectLines := wrapText(commit.Message, width)
	lines = append(lines, subjectLines...)

	// 5. Body (if exists)
	if commit.Body != "" {
		// Blank line before body
		lines = append(lines, "")

		// Wrap and truncate body
		bodyLines := renderBodyLines(commit.Body, width, bodyMaxLines)
		lines = append(lines, bodyLines...)
	}

	return lines
}

// renderMetadataLine renders the metadata line with smart truncation
// Format: {hash} · {author} committed {time}
// Priority: hash > time > author
func renderMetadataLine(commit core.GitCommit, width int, ctx core.Context) string {
	shortHash := format.ToShortHash(commit.Hash)
	relativeTime := format.ToRelativeTimeFrom(commit.Date, ctx.Now())
	author := commit.Author

	// Build the full metadata line without styling to measure plain text width
	separator := " · "
	committedText := " committed "

	// Try full format first: {hash} · {author} committed {time}
	fullText := shortHash + separator + author + committedText + relativeTime
	if utf8.RuneCountInString(fullText) <= width {
		// Render with styling
		var b strings.Builder
		b.WriteString(styles.HashStyle.Render(shortHash))
		b.WriteString(styles.HeaderStyle.Render(separator))
		b.WriteString(styles.AuthorStyle.Render(author))
		b.WriteString(styles.HeaderStyle.Render(committedText))
		b.WriteString(styles.TimeStyle.Render(relativeTime))
		return b.String()
	}

	// Level 2: Truncate author: {hash} · {auth…} committed {time}
	// Calculate available space for author
	fixedPartsLen := utf8.RuneCountInString(shortHash + separator + committedText + relativeTime)
	availableForAuthor := width - fixedPartsLen
	if availableForAuthor >= 4 { // Minimum: 3 chars + ellipsis
		truncatedAuthor := truncateWithEllipsis(author, availableForAuthor)
		var b strings.Builder
		b.WriteString(styles.HashStyle.Render(shortHash))
		b.WriteString(styles.HeaderStyle.Render(separator))
		b.WriteString(styles.AuthorStyle.Render(truncatedAuthor))
		b.WriteString(styles.HeaderStyle.Render(committedText))
		b.WriteString(styles.TimeStyle.Render(relativeTime))
		return b.String()
	}

	// Level 3: Drop "committed": {hash} · {auth…} {time}
	fixedPartsLen = utf8.RuneCountInString(shortHash + separator + " " + relativeTime)
	availableForAuthor = width - fixedPartsLen
	if availableForAuthor >= 4 {
		truncatedAuthor := truncateWithEllipsis(author, availableForAuthor)
		var b strings.Builder
		b.WriteString(styles.HashStyle.Render(shortHash))
		b.WriteString(styles.HeaderStyle.Render(separator))
		b.WriteString(styles.AuthorStyle.Render(truncatedAuthor))
		b.WriteString(styles.HeaderStyle.Render(" "))
		b.WriteString(styles.TimeStyle.Render(relativeTime))
		return b.String()
	}

	// Level 4: Drop author: {hash} · {time}
	var b strings.Builder
	b.WriteString(styles.HashStyle.Render(shortHash))
	b.WriteString(styles.HeaderStyle.Render(separator))
	b.WriteString(styles.TimeStyle.Render(relativeTime))
	return b.String()
}

// truncateWithEllipsis truncates text to maxLen, adding ellipsis if truncated
// Uses single ellipsis character "…" (U+2026)
func truncateWithEllipsis(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	textLen := utf8.RuneCountInString(text)
	if textLen <= maxLen {
		return text
	}
	if maxLen == 1 {
		return "…"
	}

	// Truncate to maxLen-1 runes and add ellipsis
	runes := []rune(text)
	return string(runes[:maxLen-1]) + "…"
}

// renderRefsLines renders ref names, wrapping if needed
// Format: main, origin/main, HEAD
func renderRefsLines(refs []core.RefInfo, width int) []string {
	// Build comma-separated ref names
	var refNames []string
	for _, ref := range refs {
		refNames = append(refNames, ref.Name)
	}
	refsText := strings.Join(refNames, ", ")

	// Wrap if needed
	wrappedLines := wrapText(refsText, width)

	// Apply monochrome styling (using TimeStyle for subtle appearance)
	var styledLines []string
	for _, line := range wrappedLines {
		styledLines = append(styledLines, styles.TimeStyle.Render(line))
	}

	return styledLines
}

// renderBodyLines renders the body text, wrapping and truncating if needed
// Returns body lines with truncation indicator if truncated
func renderBodyLines(body string, width int, maxLines int) []string {
	// Split body into lines and wrap each
	var wrappedLines []string
	bodyLines := strings.Split(body, "\n")

	for _, line := range bodyLines {
		if line == "" {
			wrappedLines = append(wrappedLines, "")
		} else {
			wrapped := wrapText(line, width)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	// Apply maxLines limit if specified
	if maxLines > 0 && len(wrappedLines) > maxLines {
		truncatedLines := wrappedLines[:maxLines]
		remainingCount := len(wrappedLines) - maxLines
		indicator := fmt.Sprintf("(... %d more lines)", remainingCount)
		truncatedLines = append(truncatedLines, styles.TimeStyle.Render(indicator))
		return truncatedLines
	}

	return wrappedLines
}

// wrapText wraps text to the specified width
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var currentLine strings.Builder
	for _, word := range words {
		// If adding this word would exceed width, start a new line
		if currentLine.Len() > 0 && currentLine.Len()+1+utf8.RuneCountInString(word) > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
		}

		// Add word to current line
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
