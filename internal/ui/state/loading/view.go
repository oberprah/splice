package loading

import "github.com/oberprah/splice/internal/ui/state"

// View renders the loading message
func (s State) View(ctx state.Context) string {
	return "  Loading commits...\n"
}
