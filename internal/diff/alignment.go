package diff

import (
	"strings"

	"github.com/oberprah/splice/internal/highlight"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// ═══════════════════════════════════════════════════════════
// CONTENT: File lines with syntax highlighting (no layout concerns)
// ═══════════════════════════════════════════════════════════

// AlignedLine represents a single line of file content with syntax highlighting tokens.
// This is distinct from the Line type in parse.go which represents a parsed diff line.
type AlignedLine struct {
	Tokens []highlight.Token
}

// Text returns the raw text content of the line by concatenating all token values.
// This is used for similarity matching and inline diff computation.
func (l *AlignedLine) Text() string {
	var b strings.Builder
	for _, t := range l.Tokens {
		b.WriteString(t.Value)
	}
	return b.String()
}

// FileContent represents the complete content of one side of a diff (old or new file).
// It contains the file path and all lines with their syntax highlighting tokens.
type FileContent struct {
	Path  string
	Lines []AlignedLine
}

// LineNo returns the 1-indexed line number for display purposes.
// The idx parameter is the 0-based index into the Lines slice.
func (fc *FileContent) LineNo(idx int) int {
	return idx + 1
}

// ═══════════════════════════════════════════════════════════
// ALIGNMENT: Sum type representing how lines relate
// ═══════════════════════════════════════════════════════════

// Alignment is a sum type representing the relationship between lines in the old
// and new versions of a file. Each concrete type represents one of four possible
// relationships: unchanged, modified, removed, or added.
//
// This is a sealed interface - only the four concrete types below can implement it.
// Use a type switch to handle all cases in rendering logic.
type Alignment interface {
	alignment() // unexported marker method prevents external implementation
}

// UnchangedAlignment represents a line that exists in both files with identical content.
// Both sides should be rendered with neutral styling.
type UnchangedAlignment struct {
	LeftIdx  int // Index into left FileContent.Lines
	RightIdx int // Index into right FileContent.Lines
}

// ModifiedAlignment represents a paired line where the left (old) line was modified
// to become the right (new) line. The InlineDiff field contains character-level
// differences for highlighting specific changes within the line.
type ModifiedAlignment struct {
	LeftIdx    int                   // Index into left FileContent.Lines
	RightIdx   int                   // Index into right FileContent.Lines
	InlineDiff []diffmatchpatch.Diff // Character-level diff for inline highlighting
}

// RemovedAlignment represents a line that exists only in the old file.
// The left side should be rendered with deletion styling, and the right side
// should show a filler to maintain visual alignment.
type RemovedAlignment struct {
	LeftIdx int // Index into left FileContent.Lines
	// Right side is implicitly a filler - no field needed
}

// AddedAlignment represents a line that exists only in the new file.
// The left side should show a filler, and the right side should be rendered
// with addition styling.
type AddedAlignment struct {
	RightIdx int // Index into right FileContent.Lines
	// Left side is implicitly a filler - no field needed
}

// Marker method implementations - seal the Alignment interface
func (UnchangedAlignment) alignment() {}
func (ModifiedAlignment) alignment()  {}
func (RemovedAlignment) alignment()   {}
func (AddedAlignment) alignment()     {}

// ═══════════════════════════════════════════════════════════
// TOP-LEVEL: Combines content + alignment
// ═══════════════════════════════════════════════════════════

// AlignedFileDiff represents a complete diff between two versions of a file.
// It combines the content from both sides with an alignment sequence that
// describes how lines relate and should be displayed.
// This is distinct from the FileDiff type in parse.go which represents a parsed unified diff.
type AlignedFileDiff struct {
	Left       FileContent // Old version of the file
	Right      FileContent // New version of the file
	Alignments []Alignment // One entry per display row, describes line relationships
}
