package states

// View renders the loading message
func (s LoadingState) View(ctx Context) *ViewBuilder {
	vb := NewViewBuilder()
	vb.AddLine("  Loading commits...")
	return vb
}
