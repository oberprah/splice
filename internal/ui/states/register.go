package states

import (
	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/ui/states/diff"
	stateserror "github.com/oberprah/splice/internal/ui/states/error"
	"github.com/oberprah/splice/internal/ui/states/files"
	"github.com/oberprah/splice/internal/ui/states/log"
)

func init() {
	// Register LogState factory
	app.RegisterStateFactory(app.LogScreen, func(data any) app.State {
		d := data.(app.LogScreenData)
		// Create LogState with initial preview loading for the first commit
		firstCommitHash := d.Commits[0].Hash
		return &log.State{
			Commits:       d.Commits,
			Cursor:        0,
			ViewportStart: 0,
			Preview:       log.PreviewLoading{ForHash: firstCommitHash},
			GraphLayout:   d.GraphLayout,
		}
	})

	// Register FilesState factory
	app.RegisterStateFactory(app.FilesScreen, func(data any) app.State {
		d := data.(app.FilesScreenData)
		return &files.State{
			Commit:        d.Commit,
			Files:         d.Files,
			Cursor:        0,
			ViewportStart: 0,
		}
	})

	// Register DiffState factory
	app.RegisterStateFactory(app.DiffScreen, func(data any) app.State {
		d := data.(app.DiffScreenData)
		// Calculate initial viewport position - scroll to first change
		viewportStart := 0
		if d.Diff != nil && len(d.ChangeIndices) > 0 {
			viewportStart = d.ChangeIndices[0]
		}
		return &diff.State{
			Commit:           d.Commit,
			File:             d.File,
			Diff:             d.Diff,
			ChangeIndices:    d.ChangeIndices,
			ViewportStart:    viewportStart,
			CurrentChangeIdx: 0,
		}
	})

	// Register ErrorState factory
	app.RegisterStateFactory(app.ErrorScreen, func(data any) app.State {
		d := data.(app.ErrorScreenData)
		return stateserror.State{
			Err: d.Err,
		}
	})
}
