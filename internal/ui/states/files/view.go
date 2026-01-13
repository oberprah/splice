package files

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/components"
)

// View renders the files state
func (s *State) View(ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Render commit info using shared component
	// For now, we only support CommitRangeDiffSource for display
	// TODO: Add support for uncommitted changes display in a future step
	var commitInfoLines []string
	switch src := s.Source.(type) {
	case core.CommitRangeDiffSource:
		commitInfoLines = components.CommitInfoFromRange(src.ToCommitRange(), ctx.Width(), 0, ctx) // 0 = unlimited body lines
	case core.UncommittedChangesDiffSource:
		// TODO: Display uncommitted changes header
		// For now, show a simple label
		var label string
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			label = "Uncommitted changes (unstaged)"
		case core.UncommittedTypeStaged:
			label = "Uncommitted changes (staged)"
		case core.UncommittedTypeAll:
			label = "Uncommitted changes"
		}
		commitInfoLines = []string{label}
	}
	for _, line := range commitInfoLines {
		vb.AddLine(line)
	}

	// Render file section using shared component
	// Note: FileSection includes blank line separator and stats line
	fileSectionLines := components.FileSection(s.Files, ctx.Width(), &s.Cursor)

	// Calculate available height for file list (subtract commit info lines + file section header)
	// commitInfoLines + blank line + stats line = total non-file lines
	commitInfoLinesCount := len(commitInfoLines)
	fileSectionHeaderLines := 2 // blank line + stats line
	availableHeight := max(ctx.Height()-commitInfoLinesCount-fileSectionHeaderLines, 1)

	// Determine which file lines to render based on viewport
	// The FileSection returns: blank line, stats line, then all file lines
	// We need to render the header (blank + stats) then only visible files
	totalFileLines := len(fileSectionLines) - fileSectionHeaderLines

	// Add the file section header (blank line + stats line)
	for i := 0; i < fileSectionHeaderLines && i < len(fileSectionLines); i++ {
		vb.AddLine(fileSectionLines[i])
	}

	// Calculate viewport for files
	viewportEnd := min(s.ViewportStart+availableHeight, totalFileLines)

	// Add only visible file lines
	for i := s.ViewportStart; i < viewportEnd; i++ {
		lineIndex := fileSectionHeaderLines + i
		if lineIndex < len(fileSectionLines) {
			vb.AddLine(fileSectionLines[lineIndex])
		}
	}

	return vb
}
