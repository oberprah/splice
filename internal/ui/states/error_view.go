package states

import (
	"fmt"
)

// View renders the error message
func (s ErrorState) View(ctx Context) string {
	return fmt.Sprintf("  Error: %v\n", s.Err)
}
