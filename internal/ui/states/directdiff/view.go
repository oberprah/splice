package directdiff

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/components"
)

// View renders the loading message based on the diff source type
func (s State) View(ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()

	// Generate appropriate loading message based on source type
	switch src := s.Source.(type) {
	case core.CommitRangeDiffSource:
		if src.Count == 1 {
			vb.AddLine("  Loading files for commit...")
		} else {
			vb.AddLine("  Loading files for commit range...")
		}
	case core.UncommittedChangesDiffSource:
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			vb.AddLine("  Loading unstaged changes...")
		case core.UncommittedTypeStaged:
			vb.AddLine("  Loading staged changes...")
		case core.UncommittedTypeAll:
			vb.AddLine("  Loading uncommitted changes...")
		default:
			vb.AddLine("  Loading changes...")
		}
	default:
		vb.AddLine("  Loading changes...")
	}

	return vb
}
