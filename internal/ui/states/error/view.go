package error

import (
	"fmt"

	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/ui/components"
)

// View renders the error message
func (s State) View(ctx app.Context) app.ViewRenderer {
	vb := components.NewViewBuilder()
	vb.AddLine(fmt.Sprintf("  Error: %v", s.Err))
	return vb
}
