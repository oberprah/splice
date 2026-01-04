package states

import (
	"testing"
)

// Per-file helper that adds subdirectory prefix
func assertLoadingViewGolden(t *testing.T, output *ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), "loading_view/"+filename, *update)
}

func TestLoadingState_View(t *testing.T) {
	s := LoadingState{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLoadingViewGolden(t, output, "loading_message.golden")
}
