package states

import (
	"strings"

	"github.com/oberprah/splice/internal/git"
	"github.com/oberprah/splice/internal/ui/styles"
)

// View renders the files state
func (s *FilesState) View(ctx Context) string {
	var b strings.Builder

	// Render header with commit info
	header := s.renderHeader()
	b.WriteString(header)

	// Render separator
	separator := strings.Repeat("─", min(ctx.Width(), 80))
	b.WriteString(styles.HeaderStyle.Render(separator))
	b.WriteString("\n")

	// Calculate available height for file list (subtract header lines)
	// Count actual header lines (including body if present)
	headerLines := strings.Count(header, "\n") + 1 // +1 for separator
	availableHeight := max(ctx.Height()-headerLines, 1)

	// Calculate the end of the viewport
	viewportEnd := min(s.ViewportStart+availableHeight, len(s.Files))

	// Render only visible files
	for i := s.ViewportStart; i < viewportEnd; i++ {
		file := s.Files[i]
		line := s.formatFileLine(file, i == s.Cursor, ctx.Width())
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader formats the commit information header
func (s *FilesState) renderHeader() string {
	// Format:
	// abc123d · John Doe committed 2 hours ago · 3 files · +45 -12
	//
	// Subject line
	//
	// Body paragraph 1...
	// Body paragraph 2...

	var b strings.Builder

	// First line: metadata
	b.WriteString(RenderCommitMetadata(s.Commit, s.Files))
	b.WriteString("\n\n")

	// Subject line
	b.WriteString(styles.MessageStyle.Render(s.Commit.Message))
	b.WriteString("\n")

	// Body (if exists)
	if s.Commit.Body != "" {
		b.WriteString("\n")
		b.WriteString(styles.MessageStyle.Render(s.Commit.Body))
		b.WriteString("\n")
	}

	return b.String()
}


// formatFileLine formats a single file line with proper styling
func (s *FilesState) formatFileLine(file git.FileChange, isSelected bool, width int) string {
	// Calculate dynamic widths based on all files
	maxAddWidth, maxDelWidth := CalculateMaxStatWidth(s.Files)

	return FormatFileLine(FormatFileLineParams{
		File:         file,
		IsSelected:   isSelected,
		Width:        width,
		MaxAddWidth:  maxAddWidth,
		MaxDelWidth:  maxDelWidth,
		ShowSelector: true,
	})
}
