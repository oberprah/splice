package loading

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/components"
)

// View renders the loading message
func (s State) View(ctx core.Context) core.ViewRenderer {
	vb := components.NewViewBuilder()
	vb.AddLine("  Loading commits...")
	return vb
}
