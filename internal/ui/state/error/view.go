package error

import (
	"fmt"

	"github.com/oberprah/splice/internal/ui/state"
)

// View renders the error message
func (s State) View(ctx state.Context) string {
	return fmt.Sprintf("  Error: %v\n", s.Err)
}
