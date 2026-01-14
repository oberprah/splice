package diff

import (
	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// Block is a sealed interface for diff content blocks
type Block interface {
	block()         // Sealed marker
	LineCount() int // Display lines in this block
}

// UnchangedBlock contains consecutive lines identical on both sides
type UnchangedBlock struct {
	Lines []LinePair
}

func (UnchangedBlock) block()           {}
func (b UnchangedBlock) LineCount() int { return len(b.Lines) }

// LinePair holds matching lines from both file versions
type LinePair struct {
	LeftLineNo  int               // 1-indexed line number in old file
	RightLineNo int               // 1-indexed line number in new file
	Tokens      []highlight.Token // Shared tokens (content is identical)
}

// ChangeBlock contains consecutive changed lines (a "hunk")
type ChangeBlock struct {
	Lines []ChangeLine
}

func (ChangeBlock) block()           {}
func (b ChangeBlock) LineCount() int { return len(b.Lines) }

// ChangeLine is a sealed interface for individual changed lines
type ChangeLine interface {
	changeLine()
}

// ModifiedLine: line exists in both files but content differs
type ModifiedLine struct {
	LeftLineNo  int
	RightLineNo int
	LeftTokens  []highlight.Token
	RightTokens []highlight.Token
	InlineDiff  []diffmatchpatch.Diff
}

func (ModifiedLine) changeLine() {}

// RemovedLine: line exists only in old file
type RemovedLine struct {
	LeftLineNo int
	Tokens     []highlight.Token
}

func (RemovedLine) changeLine() {}

// AddedLine: line exists only in new file
type AddedLine struct {
	RightLineNo int
	Tokens      []highlight.Token
}

func (AddedLine) changeLine() {}

// FileDiff is the top-level structure for a parsed file diff
type FileDiff struct {
	Path   string // File path (for display)
	Blocks []Block
}

// TotalLineCount returns the total number of display lines across all blocks
func (fd *FileDiff) TotalLineCount() int {
	total := 0
	for _, b := range fd.Blocks {
		total += b.LineCount()
	}
	return total
}
