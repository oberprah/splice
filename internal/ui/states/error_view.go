package states

import (
	"fmt"
)

// View renders the error message
func (s ErrorState) View(ctx Context) *ViewBuilder {
	vb := NewViewBuilder()
	vb.AddLine(fmt.Sprintf("  Error: %v", s.Err))
	return vb
}
