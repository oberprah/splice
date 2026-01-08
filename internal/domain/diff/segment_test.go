package diff

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════
// HunkLineType Tests
// ═══════════════════════════════════════════════════════════

func TestHunkLineType_EnumValues_AreDistinct(t *testing.T) {
	// Verify all HunkLineType constants have distinct values
	types := []HunkLineType{HunkLineAdded, HunkLineRemoved, HunkLineModified}
	seen := make(map[HunkLineType]bool)

	for _, lt := range types {
		if seen[lt] {
			t.Errorf("duplicate HunkLineType value found: %d", lt)
		}
		seen[lt] = true
	}

	// Should have 3 distinct values
	if len(seen) != 3 {
		t.Errorf("expected 3 distinct HunkLineType values, got %d", len(seen))
	}
}

// ═══════════════════════════════════════════════════════════
// HunkLine Tests
// ═══════════════════════════════════════════════════════════

func TestHunkLine_Construction(t *testing.T) {
	tests := []struct {
		name      string
		sourceIdx int
		lineType  HunkLineType
	}{
		{"added line", 5, HunkLineAdded},
		{"removed line", 10, HunkLineRemoved},
		{"modified line", 0, HunkLineModified},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hl := HunkLine{
				SourceIdx: tt.sourceIdx,
				Type:      tt.lineType,
			}

			if hl.SourceIdx != tt.sourceIdx {
				t.Errorf("expected SourceIdx %d, got %d", tt.sourceIdx, hl.SourceIdx)
			}
			if hl.Type != tt.lineType {
				t.Errorf("expected Type %d, got %d", tt.lineType, hl.Type)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// UnchangedSegment Tests
// ═══════════════════════════════════════════════════════════

func TestUnchangedSegment_Construction(t *testing.T) {
	seg := UnchangedSegment{
		LeftStart:  0,
		RightStart: 0,
		Count:      10,
	}

	if seg.LeftStart != 0 {
		t.Errorf("expected LeftStart 0, got %d", seg.LeftStart)
	}
	if seg.RightStart != 0 {
		t.Errorf("expected RightStart 0, got %d", seg.RightStart)
	}
	if seg.Count != 10 {
		t.Errorf("expected Count 10, got %d", seg.Count)
	}
}

func TestUnchangedSegment_DifferentStartIndices(t *testing.T) {
	// After insertions/deletions, line indices can differ between sides
	seg := UnchangedSegment{
		LeftStart:  5,
		RightStart: 8,
		Count:      3,
	}

	if seg.LeftStart != 5 {
		t.Errorf("expected LeftStart 5, got %d", seg.LeftStart)
	}
	if seg.RightStart != 8 {
		t.Errorf("expected RightStart 8, got %d", seg.RightStart)
	}
	if seg.Count != 3 {
		t.Errorf("expected Count 3, got %d", seg.Count)
	}
}

func TestUnchangedSegment_ImplementsSegment(t *testing.T) {
	var seg Segment = UnchangedSegment{
		LeftStart:  0,
		RightStart: 0,
		Count:      5,
	}

	// Type assertion should succeed
	if _, ok := seg.(UnchangedSegment); !ok {
		t.Error("UnchangedSegment should implement Segment interface")
	}
}

// ═══════════════════════════════════════════════════════════
// HunkSegment Tests
// ═══════════════════════════════════════════════════════════

func TestHunkSegment_Construction_PureAddition(t *testing.T) {
	// Hunk with only additions (left side empty)
	seg := HunkSegment{
		LeftLines: []HunkLine{},
		RightLines: []HunkLine{
			{SourceIdx: 0, Type: HunkLineAdded},
			{SourceIdx: 1, Type: HunkLineAdded},
		},
	}

	if len(seg.LeftLines) != 0 {
		t.Errorf("expected 0 LeftLines, got %d", len(seg.LeftLines))
	}
	if len(seg.RightLines) != 2 {
		t.Errorf("expected 2 RightLines, got %d", len(seg.RightLines))
	}
}

func TestHunkSegment_Construction_PureRemoval(t *testing.T) {
	// Hunk with only removals (right side empty)
	seg := HunkSegment{
		LeftLines: []HunkLine{
			{SourceIdx: 3, Type: HunkLineRemoved},
			{SourceIdx: 4, Type: HunkLineRemoved},
			{SourceIdx: 5, Type: HunkLineRemoved},
		},
		RightLines: []HunkLine{},
	}

	if len(seg.LeftLines) != 3 {
		t.Errorf("expected 3 LeftLines, got %d", len(seg.LeftLines))
	}
	if len(seg.RightLines) != 0 {
		t.Errorf("expected 0 RightLines, got %d", len(seg.RightLines))
	}
}

func TestHunkSegment_Construction_MixedChanges(t *testing.T) {
	// Hunk with both removals on left and additions on right
	seg := HunkSegment{
		LeftLines: []HunkLine{
			{SourceIdx: 2, Type: HunkLineRemoved},
			{SourceIdx: 3, Type: HunkLineModified},
		},
		RightLines: []HunkLine{
			{SourceIdx: 2, Type: HunkLineModified},
			{SourceIdx: 3, Type: HunkLineAdded},
			{SourceIdx: 4, Type: HunkLineAdded},
		},
	}

	if len(seg.LeftLines) != 2 {
		t.Errorf("expected 2 LeftLines, got %d", len(seg.LeftLines))
	}
	if len(seg.RightLines) != 3 {
		t.Errorf("expected 3 RightLines, got %d", len(seg.RightLines))
	}

	// Verify individual line types
	if seg.LeftLines[0].Type != HunkLineRemoved {
		t.Errorf("expected first left line to be Removed")
	}
	if seg.LeftLines[1].Type != HunkLineModified {
		t.Errorf("expected second left line to be Modified")
	}
	if seg.RightLines[0].Type != HunkLineModified {
		t.Errorf("expected first right line to be Modified")
	}
}

func TestHunkSegment_ImplementsSegment(t *testing.T) {
	var seg Segment = HunkSegment{
		LeftLines:  []HunkLine{},
		RightLines: []HunkLine{},
	}

	// Type assertion should succeed
	if _, ok := seg.(HunkSegment); !ok {
		t.Error("HunkSegment should implement Segment interface")
	}
}

// ═══════════════════════════════════════════════════════════
// Segment Type Switch Tests
// ═══════════════════════════════════════════════════════════

func TestSegment_TypeSwitch(t *testing.T) {
	segments := []Segment{
		UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 5},
		HunkSegment{
			LeftLines:  []HunkLine{{SourceIdx: 5, Type: HunkLineRemoved}},
			RightLines: []HunkLine{{SourceIdx: 5, Type: HunkLineAdded}},
		},
		UnchangedSegment{LeftStart: 6, RightStart: 6, Count: 3},
	}

	unchangedCount := 0
	hunkCount := 0

	for _, seg := range segments {
		switch seg.(type) {
		case UnchangedSegment:
			unchangedCount++
		case HunkSegment:
			hunkCount++
		default:
			t.Errorf("unexpected segment type: %T", seg)
		}
	}

	if unchangedCount != 2 {
		t.Errorf("expected 2 UnchangedSegments, got %d", unchangedCount)
	}
	if hunkCount != 1 {
		t.Errorf("expected 1 HunkSegment, got %d", hunkCount)
	}
}

func TestSegment_TypeAssertionWithValue(t *testing.T) {
	var seg Segment = UnchangedSegment{LeftStart: 10, RightStart: 15, Count: 20}

	us, ok := seg.(UnchangedSegment)
	if !ok {
		t.Fatal("expected type assertion to succeed")
	}

	if us.LeftStart != 10 {
		t.Errorf("expected LeftStart 10, got %d", us.LeftStart)
	}
	if us.RightStart != 15 {
		t.Errorf("expected RightStart 15, got %d", us.RightStart)
	}
	if us.Count != 20 {
		t.Errorf("expected Count 20, got %d", us.Count)
	}
}

// ═══════════════════════════════════════════════════════════
// Segment Slice Tests
// ═══════════════════════════════════════════════════════════

func TestSegmentSlice_EmptyIsValid(t *testing.T) {
	segments := []Segment{}

	if len(segments) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(segments))
	}
}

func TestSegmentSlice_MixedTypes(t *testing.T) {
	// Typical diff structure: unchanged, hunk, unchanged, hunk, unchanged
	segments := []Segment{
		UnchangedSegment{LeftStart: 0, RightStart: 0, Count: 10},
		HunkSegment{
			LeftLines:  []HunkLine{{SourceIdx: 10, Type: HunkLineRemoved}},
			RightLines: []HunkLine{{SourceIdx: 10, Type: HunkLineAdded}, {SourceIdx: 11, Type: HunkLineAdded}},
		},
		UnchangedSegment{LeftStart: 11, RightStart: 12, Count: 30},
		HunkSegment{
			LeftLines:  []HunkLine{{SourceIdx: 41, Type: HunkLineModified}},
			RightLines: []HunkLine{{SourceIdx: 42, Type: HunkLineModified}},
		},
		UnchangedSegment{LeftStart: 42, RightStart: 43, Count: 5},
	}

	if len(segments) != 5 {
		t.Errorf("expected 5 segments, got %d", len(segments))
	}

	// Verify alternating types
	_, isUnchanged := segments[0].(UnchangedSegment)
	if !isUnchanged {
		t.Error("expected first segment to be UnchangedSegment")
	}

	_, isHunk := segments[1].(HunkSegment)
	if !isHunk {
		t.Error("expected second segment to be HunkSegment")
	}
}
