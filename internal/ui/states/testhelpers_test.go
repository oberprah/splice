package states

// mockContext is a test helper that implements the Context interface
type mockContext struct {
	width  int
	height int
}

func (m mockContext) Width() int {
	return m.width
}

func (m mockContext) Height() int {
	return m.height
}
