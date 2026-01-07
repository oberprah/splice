package core

import "fmt"

// CursorState represents the cursor position and mode in a list view.
// It is either CursorNormal (single position) or CursorVisual (position + anchor for selection).
type CursorState interface {
	Position() int
	cursorState() // unexported marker method
}

// CursorNormal represents a normal cursor at a single position.
type CursorNormal struct {
	Pos int
}

func (c CursorNormal) Position() int { return c.Pos }
func (c CursorNormal) cursorState()  {}

// CursorVisual represents a visual selection mode with a cursor position and anchor.
type CursorVisual struct {
	Pos    int
	Anchor int
}

func (c CursorVisual) Position() int { return c.Pos }
func (c CursorVisual) cursorState()  {}

// SelectionRange returns the ordered (min, max) indices of the selection.
// For CursorNormal, returns (Pos, Pos).
// For CursorVisual, returns (min(Pos, Anchor), max(Pos, Anchor)).
func SelectionRange(cursor CursorState) (int, int) {
	switch c := cursor.(type) {
	case CursorNormal:
		return c.Pos, c.Pos
	case CursorVisual:
		if c.Pos < c.Anchor {
			return c.Pos, c.Anchor
		}
		return c.Anchor, c.Pos
	default:
		panic(fmt.Sprintf("unhandled CursorState type: %T", cursor))
	}
}

// IsInSelection returns true if the given index is within the selection range.
func IsInSelection(cursor CursorState, index int) bool {
	min, max := SelectionRange(cursor)
	return index >= min && index <= max
}
