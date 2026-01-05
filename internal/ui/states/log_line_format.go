package states

// This file contains commit line formatting and truncation logic.
// The main entry point is formatCommitLine() which applies a 10-level
// progressive truncation strategy to fit commit lines within terminal width.

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/styles"
)

// RefsLevel represents the truncation level for refs display
type RefsLevel int

const (
	RefsLevelFull RefsLevel = iota
	RefsLevelShortenIndividual
	RefsLevelFirstPlusCount
	RefsLevelCountOnly
)

// CommitLineComponents holds all pre-computed components for a commit line.
// This struct enables formatCommitLine to be a pure function.
type CommitLineComponents struct {
	IsSelected bool
	Graph      string
	Hash       string
	Refs       []git.RefInfo
	Message    string
	Author     string
	Time       string
}

// capMessage truncates a message to maxLen characters with "..." suffix.
// Returns the original message if it fits within maxLen.
func capMessage(message string, maxLen int) string {
	if utf8.RuneCountInString(message) <= maxLen {
		return message
	}
	if maxLen < 3 {
		return ""
	}
	// Convert to runes to properly truncate multi-byte characters
	runes := []rune(message)
	return string(runes[:maxLen-3]) + "..."
}

// truncateAuthor truncates an author name to maxLen characters with "..." suffix.
// Returns the original author if it fits within maxLen.
// Returns empty string if maxLen < 3.
func truncateAuthor(author string, maxLen int) string {
	if utf8.RuneCountInString(author) <= maxLen {
		return author
	}
	if maxLen < 3 {
		return ""
	}
	// Convert to runes to properly truncate multi-byte characters
	runes := []rune(author)
	return string(runes[:maxLen-3]) + "..."
}

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// truncateEntireLine hard-truncates an assembled line to maxWidth characters.
// Uses "..." suffix if maxWidth >= 3, otherwise truncates to available space.
// Note: This function works on plain strings only. It is currently unused but kept
// for potential future use cases where entire line truncation is needed.
func truncateEntireLine(line string, maxWidth int) string {
	if utf8.RuneCountInString(line) <= maxWidth {
		return line
	}
	if maxWidth <= 0 {
		return ""
	}
	// Convert to runes to properly truncate multi-byte characters
	runes := []rune(line)
	if maxWidth < 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}

// formatRefsFull formats all refs with their full names.
// Returns formatted refs like "(HEAD -> main, tag: v1.0)" with trailing space.
func formatRefsFull(refs []git.RefInfo) string {
	if len(refs) == 0 {
		return ""
	}

	var parts []string
	for _, ref := range refs {
		var formatted string
		switch ref.Type {
		case git.RefTypeTag:
			formatted = fmt.Sprintf("tag: %s", ref.Name)
		default:
			// For branches, just use the name
			formatted = ref.Name
		}
		parts = append(parts, formatted)
	}

	return fmt.Sprintf("(%s) ", strings.Join(parts, ", "))
}

// formatRefsShortenedIndividual formats refs with individual names truncated to maxLen.
// Uses "…" (single ellipsis char) for truncation to save space.
// Note: maxLen is treated as character count (rune count).
func formatRefsShortenedIndividual(refs []git.RefInfo, maxLen int) string {
	if len(refs) == 0 {
		return ""
	}

	var parts []string
	for _, ref := range refs {
		var formatted string
		switch ref.Type {
		case git.RefTypeTag:
			name := ref.Name
			if utf8.RuneCountInString(name) > maxLen {
				if maxLen < 3 {
					name = ""
				} else {
					runes := []rune(name)
					name = string(runes[:maxLen-3]) + "…"
				}
			}
			formatted = fmt.Sprintf("tag: %s", name)
		default:
			// For branches, truncate the name
			name := ref.Name
			if utf8.RuneCountInString(name) > maxLen {
				if maxLen < 3 {
					name = ""
				} else {
					runes := []rune(name)
					name = string(runes[:maxLen-3]) + "…"
				}
			}
			formatted = name
		}
		parts = append(parts, formatted)
	}

	return fmt.Sprintf("(%s) ", strings.Join(parts, ", "))
}

// formatRefsFirstPlusCount formats refs showing only the first ref plus a count.
// Prefers showing the current branch (HEAD ref) if present.
// First ref is still truncated if needed with "…".
// Note: maxLen is treated as character count (rune count).
func formatRefsFirstPlusCount(refs []git.RefInfo, maxLen int) string {
	if len(refs) == 0 {
		return ""
	}

	// Find the HEAD ref (current branch) if it exists
	var firstRef git.RefInfo
	foundHead := false
	for _, ref := range refs {
		if ref.IsHead {
			firstRef = ref
			foundHead = true
			break
		}
	}

	// If no HEAD ref, use the first ref
	if !foundHead {
		firstRef = refs[0]
	}

	// Format the first ref
	var formatted string
	switch firstRef.Type {
	case git.RefTypeTag:
		name := firstRef.Name
		if utf8.RuneCountInString(name) > maxLen {
			if maxLen < 3 {
				name = ""
			} else {
				runes := []rune(name)
				name = string(runes[:maxLen-3]) + "…"
			}
		}
		formatted = fmt.Sprintf("tag: %s", name)
	default:
		name := firstRef.Name
		if utf8.RuneCountInString(name) > maxLen {
			if maxLen < 3 {
				name = ""
			} else {
				runes := []rune(name)
				name = string(runes[:maxLen-3]) + "…"
			}
		}
		formatted = name
	}

	// Calculate remaining refs count
	remaining := len(refs) - 1
	if remaining > 0 {
		return fmt.Sprintf("(%s +%d more) ", formatted, remaining)
	}

	return fmt.Sprintf("(%s) ", formatted)
}

// buildRefs builds the refs string at the specified truncation level.
// Returns empty string if no refs, otherwise returns formatted string with trailing space.
func buildRefs(refs []git.RefInfo, level RefsLevel) string {
	if len(refs) == 0 {
		return ""
	}

	switch level {
	case RefsLevelFull:
		return formatRefsFull(refs)
	case RefsLevelShortenIndividual:
		return formatRefsShortenedIndividual(refs, 30)
	case RefsLevelFirstPlusCount:
		return formatRefsFirstPlusCount(refs, 30)
	case RefsLevelCountOnly:
		return fmt.Sprintf("(%d refs) ", len(refs))
	default:
		return formatRefsFull(refs)
	}
}

// measureLineWidth calculates the total width of a commit line.
// Accounts for all components and spacing: selector + graph + hash + space + refs + message + separator + author + space + time.
// Note: refs already includes trailing space if non-empty.
func measureLineWidth(selector, graph, hash, refs, message, author, time string) int {
	width := utf8.RuneCountInString(selector) + utf8.RuneCountInString(graph) + utf8.RuneCountInString(hash)

	if refs != "" {
		width += 1 + utf8.RuneCountInString(refs) // space before refs + refs (which includes trailing space)
	} else {
		width += 1 // space after hash when no refs
	}

	width += utf8.RuneCountInString(message)

	if author != "" {
		width += 3 + utf8.RuneCountInString(author) // " - " + author
	}

	if time != "" {
		width += 1 + utf8.RuneCountInString(time) // space + time
	}

	return width
}

// assembleLine assembles the final commit line with proper spacing, separators, and styling.
// This is a pure function that builds the styled string from plain components.
func assembleLine(selector, graph, hash, refs, message, author, time string, isSelected bool) string {
	var line strings.Builder

	// Add selector and graph (no styling)
	line.WriteString(selector)
	line.WriteString(graph)

	// Choose styles based on selection
	var hashStyle, messageStyle, authorStyle, timeStyle lipgloss.Style
	if isSelected {
		hashStyle = styles.SelectedHashStyle
		messageStyle = styles.SelectedMessageStyle
		authorStyle = styles.SelectedAuthorStyle
		timeStyle = styles.SelectedTimeStyle
	} else {
		hashStyle = styles.HashStyle
		messageStyle = styles.MessageStyle
		authorStyle = styles.AuthorStyle
		timeStyle = styles.TimeStyle
	}

	// Add hash with space
	line.WriteString(hashStyle.Render(hash))
	line.WriteString(" ")

	// Add refs (with space) if present
	if refs != "" {
		line.WriteString(timeStyle.Render(refs)) // Use time style for refs (dim)
	}

	// Add message
	line.WriteString(messageStyle.Render(message))

	// Add separator + author if both present
	if author != "" && message != "" {
		line.WriteString(" - ")
		line.WriteString(authorStyle.Render(author))
	}

	// Add time (with space prefix) if present
	if time != "" {
		line.WriteString(" ")
		line.WriteString(timeStyle.Render(time))
	}

	return line.String()
}

// formatCommitLine applies progressive truncation to fit a commit line within available width.
// Pure function - all inputs provided via CommitLineComponents struct, no side effects.
func formatCommitLine(components CommitLineComponents, availableWidth int) string {
	// 1. Build selector based on selection state
	selector := "  "
	if components.IsSelected {
		selector = "> "
	}

	// 2. Extract components (already computed by caller)
	graph := components.Graph
	hash := components.Hash
	message := components.Message
	author := components.Author
	time := components.Time

	// Build refs at full level initially
	refs := buildRefs(components.Refs, RefsLevelFull)

	// 3. Apply truncation levels sequentially until line fits
	level := 0
	for measureLineWidth(selector, graph, hash, refs, message, author, time) > availableWidth && level < 10 {
		switch level {
		case 0:
			// Level 0: Cap message at 72 chars
			message = capMessage(message, 72)
		case 1:
			// Level 1: Truncate author to 25 chars
			author = truncateAuthor(author, 25)
		case 2:
			// Level 2: Shorten refs Level 1 - Truncate individual ref names
			refs = buildRefs(components.Refs, RefsLevelShortenIndividual)
		case 3:
			// Level 3: Shorten refs Level 2 - Show first ref + count
			refs = buildRefs(components.Refs, RefsLevelFirstPlusCount)
		case 4:
			// Level 4: Shorten refs Level 3 - Show total count only
			refs = buildRefs(components.Refs, RefsLevelCountOnly)
		case 5:
			// Level 5: Truncate author to 5 chars
			author = truncateAuthor(author, 5)
		case 6:
			// Level 6: Drop time
			time = ""
		case 7:
			// Level 7: Shorten message to 40 chars
			message = capMessage(message, 40)
		case 8:
			// Level 8: Drop author
			author = ""
		case 9:
			// Level 9: Drop refs, truncate message if needed BEFORE assembling
			refs = ""

			// Measure what we'd have without truncation (plain strings)
			plainLine := selector + graph + hash + " " + message
			if author != "" {
				plainLine += " - " + author
			}
			if time != "" {
				plainLine += " " + time
			}

			visualWidth := utf8.RuneCountInString(plainLine)

			// If too long, truncate the message (plain) to fit
			if visualWidth > availableWidth {
				excess := visualWidth - availableWidth
				targetMsgLen := max(utf8.RuneCountInString(message)-excess, 5)
				message = capMessage(message, targetMsgLen)
			}

			// NOW assemble with styling - all components fit
			return assembleLine(selector, graph, hash, refs, message, author, time, components.IsSelected)
		}
		level++
	}

	// 4. Assemble and style the line
	return assembleLine(selector, graph, hash, refs, message, author, time, components.IsSelected)
}
