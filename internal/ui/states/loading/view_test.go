package loading

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/testutils"
)

var update = flag.Bool("update", false, "update golden files")

// Per-file helper that adds subdirectory prefix
func assertLoadingViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", filename)
	testutils.AssertGolden(t, output.String(), goldenPath, *update)
}

func TestLoadingState_View(t *testing.T) {
	s := State{}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertLoadingViewGolden(t, output.(*components.ViewBuilder), "loading_message.golden")
}
