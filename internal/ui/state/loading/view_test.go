package loading

import (
	"testing"
)

type mockContext struct {
	width, height int
}

func (m mockContext) Width() int  { return m.width }
func (m mockContext) Height() int { return m.height }

func TestLoadingState_View(t *testing.T) {
	s := State{}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)
	expected := "  Loading commits...\n"

	if result != expected {
		t.Errorf("View() = %q, want %q", result, expected)
	}
}
