package states

// View renders the loading message
func (s LoadingState) View(ctx Context) string {
	return "  Loading commits...\n"
}
