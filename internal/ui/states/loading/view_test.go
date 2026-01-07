package loading

import (
	"testing"

	"github.com/oberprah/splice/internal/ui/components"
)

// Per-file helper that adds subdirectory prefix
func assertLoadingViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), ""+filename, *update)
}

func TestLoadingState_View(t *testing.T) {
	s := State{}
	ctx := mockContext{width: 80, height: 24}

	output := s.View(ctx)

	assertLoadingViewGolden(t, output.(*components.ViewBuilder), "loading_message.golden")
}
