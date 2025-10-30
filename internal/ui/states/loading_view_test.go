package states

import (
	"testing"
)

func TestLoadingState_View(t *testing.T) {
	s := LoadingState{}
	ctx := mockContext{width: 80, height: 24}

	result := s.View(ctx)
	expected := "  Loading commits...\n"

	if result != expected {
		t.Errorf("View() = %q, want %q", result, expected)
	}
}
