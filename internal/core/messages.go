package core

import (
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/git"
)

// CommitsLoadedMsg is sent when commits have been loaded
type CommitsLoadedMsg struct {
	Commits []git.GitCommit
	Err     error
}

// FilesLoadedMsg is sent when files for a commit have been loaded
type FilesLoadedMsg struct {
	Range CommitRange
	Files []git.FileChange
	Err   error
}

// FilesPreviewLoadedMsg is sent when files for a preview panel have been loaded
type FilesPreviewLoadedMsg struct {
	ForHash string
	Files   []git.FileChange
	Err     error
}

// DiffLoadedMsg is sent when diff content for a file has been loaded
type DiffLoadedMsg struct {
	Range         CommitRange
	File          git.FileChange
	Diff          *diff.AlignedFileDiff
	ChangeIndices []int // Indices of alignments that have changes (for navigation)
	Err           error
}
