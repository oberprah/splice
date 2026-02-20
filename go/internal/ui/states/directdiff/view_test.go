package directdiff

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/components"
	"github.com/oberprah/splice/internal/ui/testutils"
)

var update = flag.Bool("update", false, "update golden files")

// assertDirectDiffViewGolden adds subdirectory prefix for golden files
func assertDirectDiffViewGolden(t *testing.T, output *components.ViewBuilder, filename string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", filename)
	testutils.AssertGolden(t, output.String(), goldenPath, *update)
}

func TestDirectDiffLoadingState_View_CommitRange_Single(t *testing.T) {
	commit := core.GitCommit{
		Hash:    "abc123",
		Message: "Test commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	commitRange := core.NewSingleCommitRange(commit)
	source := commitRange.ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertDirectDiffViewGolden(t, output.(*components.ViewBuilder), "loading_commit_single.golden")
}

func TestDirectDiffLoadingState_View_CommitRange_Multiple(t *testing.T) {
	start := core.GitCommit{
		Hash:    "abc123",
		Message: "Start commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
	end := core.GitCommit{
		Hash:    "def456",
		Message: "End commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	commitRange := core.CommitRange{Start: start, End: end, Count: 5}
	source := commitRange.ToDiffSource()

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertDirectDiffViewGolden(t, output.(*components.ViewBuilder), "loading_commit_range.golden")
}

func TestDirectDiffLoadingState_View_Unstaged(t *testing.T) {
	source := core.UncommittedChangesDiffSource{Type: core.UncommittedTypeUnstaged}

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertDirectDiffViewGolden(t, output.(*components.ViewBuilder), "loading_unstaged.golden")
}

func TestDirectDiffLoadingState_View_Staged(t *testing.T) {
	source := core.UncommittedChangesDiffSource{Type: core.UncommittedTypeStaged}

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertDirectDiffViewGolden(t, output.(*components.ViewBuilder), "loading_staged.golden")
}

func TestDirectDiffLoadingState_View_AllUncommitted(t *testing.T) {
	source := core.UncommittedChangesDiffSource{Type: core.UncommittedTypeAll}

	s := State{Source: source}
	ctx := testutils.MockContext{W: 80, H: 24}

	output := s.View(ctx)

	assertDirectDiffViewGolden(t, output.(*components.ViewBuilder), "loading_all_uncommitted.golden")
}
