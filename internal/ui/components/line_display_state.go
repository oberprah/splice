package components

// LineDisplayState represents the visual state of a commit line in the log view.
type LineDisplayState int

const (
	LineStateNone         LineDisplayState = iota // Not selected, not cursor
	LineStateCursor                               // Normal mode cursor (→)
	LineStateSelected                             // Visual mode selected (▌)
	LineStateVisualCursor                         // Visual mode cursor (█)
)

// String returns the string representation of the state.
func (s LineDisplayState) String() string {
	switch s {
	case LineStateNone:
		return "None"
	case LineStateCursor:
		return "Cursor"
	case LineStateSelected:
		return "Selected"
	case LineStateVisualCursor:
		return "VisualCursor"
	default:
		return "Unknown"
	}
}

// SelectorString returns the selector character(s) for this state.
func (s LineDisplayState) SelectorString() string {
	switch s {
	case LineStateNone:
		return "  "
	case LineStateCursor:
		return "→ "
	case LineStateSelected:
		return "▌ "
	case LineStateVisualCursor:
		return "█ "
	default:
		return "  "
	}
}
