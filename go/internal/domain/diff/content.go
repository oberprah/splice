package diff

import (
	"strings"

	"github.com/oberprah/splice/internal/domain/highlight"
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
