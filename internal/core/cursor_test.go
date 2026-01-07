package core

import "testing"

func TestCursorNormal_Position(t *testing.T) {
	cursor := CursorNormal{Pos: 5}
	if got := cursor.Position(); got != 5 {
		t.Errorf("Position() = %d, want 5", got)
	}
}

func TestCursorVisual_Position(t *testing.T) {
	cursor := CursorVisual{Pos: 10, Anchor: 3}
	if got := cursor.Position(); got != 10 {
		t.Errorf("Position() = %d, want 10", got)
	}
}

func TestSelectionRange_CursorNormal(t *testing.T) {
	cursor := CursorNormal{Pos: 7}
	min, max := SelectionRange(cursor)
	if min != 7 || max != 7 {
		t.Errorf("SelectionRange() = (%d, %d), want (7, 7)", min, max)
	}
}

func TestSelectionRange_CursorVisual_PosGreaterThanAnchor(t *testing.T) {
	cursor := CursorVisual{Pos: 10, Anchor: 3}
	min, max := SelectionRange(cursor)
	if min != 3 || max != 10 {
		t.Errorf("SelectionRange() = (%d, %d), want (3, 10)", min, max)
	}
}

func TestSelectionRange_CursorVisual_PosLessThanAnchor(t *testing.T) {
	cursor := CursorVisual{Pos: 2, Anchor: 8}
	min, max := SelectionRange(cursor)
	if min != 2 || max != 8 {
		t.Errorf("SelectionRange() = (%d, %d), want (2, 8)", min, max)
	}
}

func TestSelectionRange_CursorVisual_PosEqualsAnchor(t *testing.T) {
	cursor := CursorVisual{Pos: 5, Anchor: 5}
	min, max := SelectionRange(cursor)
	if min != 5 || max != 5 {
		t.Errorf("SelectionRange() = (%d, %d), want (5, 5)", min, max)
	}
}

func TestIsInSelection_CursorNormal(t *testing.T) {
	cursor := CursorNormal{Pos: 5}

	tests := []struct {
		index int
		want  bool
	}{
		{4, false},
		{5, true},
		{6, false},
	}

	for _, tt := range tests {
		if got := IsInSelection(cursor, tt.index); got != tt.want {
			t.Errorf("IsInSelection(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}
}

func TestIsInSelection_CursorVisual(t *testing.T) {
	cursor := CursorVisual{Pos: 10, Anchor: 3}

	tests := []struct {
		index int
		want  bool
	}{
		{2, false},
		{3, true},
		{5, true},
		{10, true},
		{11, false},
	}

	for _, tt := range tests {
		if got := IsInSelection(cursor, tt.index); got != tt.want {
			t.Errorf("IsInSelection(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}
}

func TestIsInSelection_CursorVisual_ReverseOrder(t *testing.T) {
	cursor := CursorVisual{Pos: 2, Anchor: 9}

	tests := []struct {
		index int
		want  bool
	}{
		{1, false},
		{2, true},
		{5, true},
		{9, true},
		{10, false},
	}

	for _, tt := range tests {
		if got := IsInSelection(cursor, tt.index); got != tt.want {
			t.Errorf("IsInSelection(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}
}

func TestCursorState_Interface(t *testing.T) {
	// Verify both types implement the CursorState interface
	var _ CursorState = CursorNormal{}
	var _ CursorState = CursorVisual{}
}
