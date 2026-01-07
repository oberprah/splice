package core

import (
	"github.com/oberprah/splice/internal/domain/diff"
)

// CommitsLoadedMsg is sent when commits have been loaded
type CommitsLoadedMsg struct {
	Commits []GitCommit
	Err     error
}

// FilesLoadedMsg is sent when files for a commit have been loaded
type FilesLoadedMsg struct {
	CommitRange CommitRange
	Files       []FileChange
	Err         error
}

// FilesPreviewLoadedMsg is sent when files for a preview panel have been loaded
type FilesPreviewLoadedMsg struct {
	ForHash string
	Files   []FileChange
	Err     error
}

// DiffLoadedMsg is sent when diff content for a file has been loaded
type DiffLoadedMsg struct {
	CommitRange   CommitRange
	File          FileChange
	Diff          *diff.AlignedFileDiff
	ChangeIndices []int // Indices of alignments that have changes (for navigation)
	Err           error
}
