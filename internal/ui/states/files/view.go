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

	// Render tree section using TreeSection component
	// Note: TreeSection includes blank line separator and stats line
	treeSectionLines := components.TreeSection(s.VisibleItems, s.Cursor, ctx.Width())

	// Calculate available height for tree view (subtract commit info lines + tree section header)
	// commitInfoLines + blank line + stats line = total non-tree lines
	commitInfoLinesCount := len(commitInfoLines)
	treeSectionHeaderLines := 2 // blank line + stats line
	availableHeight := max(ctx.Height()-commitInfoLinesCount-treeSectionHeaderLines, 1)

	// Determine which tree lines to render based on viewport
	// The TreeSection returns: blank line, stats line, then all tree item lines
	// We need to render the header (blank + stats) then only visible items
	totalTreeLines := len(treeSectionLines) - treeSectionHeaderLines

	// Add the tree section header (blank line + stats line)
	for i := 0; i < treeSectionHeaderLines && i < len(treeSectionLines); i++ {
		vb.AddLine(treeSectionLines[i])
	}

	// Calculate viewport for tree items
	viewportEnd := min(s.ViewportStart+availableHeight, totalTreeLines)

	// Add only visible tree lines
	for i := s.ViewportStart; i < viewportEnd; i++ {
		lineIndex := treeSectionHeaderLines + i
		if lineIndex < len(treeSectionLines) {
			vb.AddLine(treeSectionLines[lineIndex])
		}
	}

	return vb
}
