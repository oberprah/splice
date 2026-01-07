package loading

import (
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/ui/components"
)

// View renders the loading message
func (s State) View(ctx app.Context) app.ViewRenderer {
	vb := components.NewViewBuilder()
	vb.AddLine("  Loading commits...")
	return vb
}
