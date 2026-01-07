package error

import (
	"fmt"
	"testing"

	"github.com/oberprah/splice/internal/ui/components"
)

// Per-file helper that adds subdirectory prefix
func assertErrorViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	assertGolden(t, output.String(), ""+filename, *update)
}

func TestErrorState_View(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		goldenFile string
	}{
		{
			name:       "simple error message",
			err:        fmt.Errorf("file not found"),
			goldenFile: "simple_error.golden",
		},
		{
			name:       "git error message",
			err:        fmt.Errorf("not a git repository"),
			goldenFile: "git_error.golden",
		},
		{
			name:       "empty commits error",
			err:        fmt.Errorf("no commits found in repository"),
			goldenFile: "empty_commits_error.golden",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := State{Err: tt.err}
			ctx := mockContext{width: 80, height: 24}

			output := s.View(ctx)

			assertErrorViewGolden(t, output.(*components.ViewBuilder), tt.goldenFile)
		})
	}
}
