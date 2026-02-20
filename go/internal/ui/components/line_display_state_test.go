package components

import "testing"

func TestLineDisplayState_String(t *testing.T) {
	tests := []struct {
		state LineDisplayState
		want  string
	}{
		{LineStateNone, "None"},
		{LineStateCursor, "Cursor"},
		{LineStateSelected, "Selected"},
		{LineStateVisualCursor, "VisualCursor"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("LineDisplayState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestLineDisplayState_String_InvalidState(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid LineDisplayState, but did not panic")
		}
	}()

	invalidState := LineDisplayState(999)
	_ = invalidState.String()
}

func TestLineDisplayState_SelectorString(t *testing.T) {
	tests := []struct {
		state LineDisplayState
		want  string
	}{
		{LineStateNone, "  "},
		{LineStateCursor, "→ "},
		{LineStateSelected, "▌ "},
		{LineStateVisualCursor, "█ "},
	}

	for _, tt := range tests {
		if got := tt.state.SelectorString(); got != tt.want {
			t.Errorf("LineDisplayState(%d).SelectorString() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestLineDisplayState_SelectorString_InvalidState(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid LineDisplayState, but did not panic")
		}
	}()

	invalidState := LineDisplayState(999)
	_ = invalidState.SelectorString()
}

func TestLineDisplayState_Values(t *testing.T) {
	// Verify the enum values are distinct
	values := []LineDisplayState{
		LineStateNone,
		LineStateCursor,
		LineStateSelected,
		LineStateVisualCursor,
	}

	seen := make(map[LineDisplayState]bool)
	for _, v := range values {
		if seen[v] {
			t.Errorf("Duplicate enum value: %d", v)
		}
		seen[v] = true
	}

	// Verify they start at 0 and increment by 1
	if LineStateNone != 0 {
		t.Errorf("LineStateNone = %d, want 0", LineStateNone)
	}
	if LineStateCursor != 1 {
		t.Errorf("LineStateCursor = %d, want 1", LineStateCursor)
	}
	if LineStateSelected != 2 {
		t.Errorf("LineStateSelected = %d, want 2", LineStateSelected)
	}
	if LineStateVisualCursor != 3 {
		t.Errorf("LineStateVisualCursor = %d, want 3", LineStateVisualCursor)
	}
}
