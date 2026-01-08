package diff

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Basic Cases
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_AllUnchanged(t *testing.T) {
	// Both files are identical - single UnchangedSegment covering all lines
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Context, Content: "line2", OldLineNo: 2, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 3},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Should produce a single UnchangedSegment
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	us, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("expected UnchangedSegment, got %T", segments[0])
	}

	if us.LeftStart != 0 {
		t.Errorf("expected LeftStart 0, got %d", us.LeftStart)
	}
	if us.RightStart != 0 {
		t.Errorf("expected RightStart 0, got %d", us.RightStart)
	}
	if us.Count != 3 {
		t.Errorf("expected Count 3, got %d", us.Count)
	}
}

func TestBuildSegments_PureAddition_InMiddle(t *testing.T) {
	// One line added in the middle
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("added"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Add, Content: "added", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "line3", OldLineNo: 2, NewLineNo: 3},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: UnchangedSegment (line1), HunkSegment (added), UnchangedSegment (line3)
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First: unchanged (line1)
	us1, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us1.LeftStart != 0 || us1.RightStart != 0 || us1.Count != 1 {
		t.Errorf("segment 0: expected (0, 0, 1), got (%d, %d, %d)", us1.LeftStart, us1.RightStart, us1.Count)
	}

	// Second: hunk (added line)
	hs, ok := segments[1].(HunkSegment)
	if !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}
	if len(hs.LeftLines) != 0 {
		t.Errorf("segment 1: expected 0 LeftLines, got %d", len(hs.LeftLines))
	}
	if len(hs.RightLines) != 1 {
		t.Errorf("segment 1: expected 1 RightLine, got %d", len(hs.RightLines))
	}
	if len(hs.RightLines) > 0 {
		if hs.RightLines[0].SourceIdx != 1 {
			t.Errorf("segment 1: expected RightLines[0].SourceIdx 1, got %d", hs.RightLines[0].SourceIdx)
		}
		if hs.RightLines[0].Type != HunkLineAdded {
			t.Errorf("segment 1: expected HunkLineAdded, got %d", hs.RightLines[0].Type)
		}
	}

	// Third: unchanged (line3)
	us2, ok := segments[2].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 2: expected UnchangedSegment, got %T", segments[2])
	}
	if us2.LeftStart != 1 || us2.RightStart != 2 || us2.Count != 1 {
		t.Errorf("segment 2: expected (1, 2, 1), got (%d, %d, %d)", us2.LeftStart, us2.RightStart, us2.Count)
	}
}

func TestBuildSegments_PureDeletion_InMiddle(t *testing.T) {
	// One line removed from the middle
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("removed"),
			makeAlignedLine("line3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "removed", OldLineNo: 2, NewLineNo: 0},
			{Type: Context, Content: "line3", OldLineNo: 3, NewLineNo: 2},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: UnchangedSegment (line1), HunkSegment (removed), UnchangedSegment (line3)
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First: unchanged (line1)
	us1, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us1.Count != 1 {
		t.Errorf("segment 0: expected Count 1, got %d", us1.Count)
	}

	// Second: hunk (removed line)
	hs, ok := segments[1].(HunkSegment)
	if !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}
	if len(hs.LeftLines) != 1 {
		t.Errorf("segment 1: expected 1 LeftLine, got %d", len(hs.LeftLines))
	}
	if len(hs.RightLines) != 0 {
		t.Errorf("segment 1: expected 0 RightLines, got %d", len(hs.RightLines))
	}
	if len(hs.LeftLines) > 0 {
		if hs.LeftLines[0].SourceIdx != 1 {
			t.Errorf("segment 1: expected LeftLines[0].SourceIdx 1, got %d", hs.LeftLines[0].SourceIdx)
		}
		if hs.LeftLines[0].Type != HunkLineRemoved {
			t.Errorf("segment 1: expected HunkLineRemoved, got %d", hs.LeftLines[0].Type)
		}
	}

	// Third: unchanged (line3)
	us2, ok := segments[2].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 2: expected UnchangedSegment, got %T", segments[2])
	}
	if us2.LeftStart != 2 || us2.RightStart != 1 || us2.Count != 1 {
		t.Errorf("segment 2: expected (2, 1, 1), got (%d, %d, %d)", us2.LeftStart, us2.RightStart, us2.Count)
	}
}

func TestBuildSegments_MixedChanges_InSameHunk(t *testing.T) {
	// Some lines removed and some added in the same hunk
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("old1"),
			makeAlignedLine("old2"),
			makeAlignedLine("unchanged2"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("new1"),
			makeAlignedLine("new2"),
			makeAlignedLine("new3"),
			makeAlignedLine("unchanged2"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "unchanged", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "old1", OldLineNo: 2, NewLineNo: 0},
			{Type: Remove, Content: "old2", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "new1", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "new2", OldLineNo: 0, NewLineNo: 3},
			{Type: Add, Content: "new3", OldLineNo: 0, NewLineNo: 4},
			{Type: Context, Content: "unchanged2", OldLineNo: 4, NewLineNo: 5},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: UnchangedSegment, HunkSegment, UnchangedSegment
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First: unchanged
	us1, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us1.Count != 1 {
		t.Errorf("segment 0: expected Count 1, got %d", us1.Count)
	}

	// Second: hunk
	hs, ok := segments[1].(HunkSegment)
	if !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}
	if len(hs.LeftLines) != 2 {
		t.Errorf("segment 1: expected 2 LeftLines, got %d", len(hs.LeftLines))
	}
	if len(hs.RightLines) != 3 {
		t.Errorf("segment 1: expected 3 RightLines, got %d", len(hs.RightLines))
	}

	// Third: unchanged
	us2, ok := segments[2].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 2: expected UnchangedSegment, got %T", segments[2])
	}
	if us2.LeftStart != 3 || us2.RightStart != 4 || us2.Count != 1 {
		t.Errorf("segment 2: expected (3, 4, 1), got (%d, %d, %d)", us2.LeftStart, us2.RightStart, us2.Count)
	}
}

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Multiple Hunks
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_MultipleHunks(t *testing.T) {
	// Two hunks separated by unchanged lines
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged1"),
			makeAlignedLine("old1"),
			makeAlignedLine("unchanged2"),
			makeAlignedLine("old2"),
			makeAlignedLine("unchanged3"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged1"),
			makeAlignedLine("new1"),
			makeAlignedLine("unchanged2"),
			makeAlignedLine("new2"),
			makeAlignedLine("unchanged3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "unchanged1", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "old1", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "new1", OldLineNo: 0, NewLineNo: 2},
			{Type: Context, Content: "unchanged2", OldLineNo: 3, NewLineNo: 3},
			{Type: Remove, Content: "old2", OldLineNo: 4, NewLineNo: 0},
			{Type: Add, Content: "new2", OldLineNo: 0, NewLineNo: 4},
			{Type: Context, Content: "unchanged3", OldLineNo: 5, NewLineNo: 5},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: Unchanged, Hunk, Unchanged, Hunk, Unchanged
	if len(segments) != 5 {
		t.Fatalf("expected 5 segments, got %d", len(segments))
	}

	// Verify alternating pattern
	for i := 0; i < len(segments); i++ {
		if i%2 == 0 {
			if _, ok := segments[i].(UnchangedSegment); !ok {
				t.Errorf("segment %d: expected UnchangedSegment, got %T", i, segments[i])
			}
		} else {
			if _, ok := segments[i].(HunkSegment); !ok {
				t.Errorf("segment %d: expected HunkSegment, got %T", i, segments[i])
			}
		}
	}
}

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Edge Cases
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_ChangeAtStart(t *testing.T) {
	// File starts with a change
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("old"),
			makeAlignedLine("unchanged"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("new"),
			makeAlignedLine("unchanged"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "old", OldLineNo: 1, NewLineNo: 0},
			{Type: Add, Content: "new", OldLineNo: 0, NewLineNo: 1},
			{Type: Context, Content: "unchanged", OldLineNo: 2, NewLineNo: 2},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: HunkSegment, UnchangedSegment
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	// First: hunk
	hs, ok := segments[0].(HunkSegment)
	if !ok {
		t.Fatalf("segment 0: expected HunkSegment, got %T", segments[0])
	}
	if len(hs.LeftLines) != 1 || len(hs.RightLines) != 1 {
		t.Errorf("segment 0: expected 1 LeftLine and 1 RightLine, got %d and %d",
			len(hs.LeftLines), len(hs.RightLines))
	}

	// Second: unchanged
	us, ok := segments[1].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 1: expected UnchangedSegment, got %T", segments[1])
	}
	if us.LeftStart != 1 || us.RightStart != 1 || us.Count != 1 {
		t.Errorf("segment 1: expected (1, 1, 1), got (%d, %d, %d)", us.LeftStart, us.RightStart, us.Count)
	}
}

func TestBuildSegments_ChangeAtEnd(t *testing.T) {
	// File ends with a change
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("old"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("unchanged"),
			makeAlignedLine("new"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "unchanged", OldLineNo: 1, NewLineNo: 1},
			{Type: Remove, Content: "old", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "new", OldLineNo: 0, NewLineNo: 2},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: UnchangedSegment, HunkSegment
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(segments))
	}

	// First: unchanged
	us, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us.Count != 1 {
		t.Errorf("segment 0: expected Count 1, got %d", us.Count)
	}

	// Second: hunk
	hs, ok := segments[1].(HunkSegment)
	if !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}
	if len(hs.LeftLines) != 1 || len(hs.RightLines) != 1 {
		t.Errorf("segment 1: expected 1 LeftLine and 1 RightLine, got %d and %d",
			len(hs.LeftLines), len(hs.RightLines))
	}
}

func TestBuildSegments_EmptyFiles(t *testing.T) {
	left := FileContent{Path: "test.txt", Lines: []AlignedLine{}}
	right := FileContent{Path: "test.txt", Lines: []AlignedLine{}}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines:   []Line{},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Empty files should produce no segments
	if len(segments) != 0 {
		t.Errorf("expected 0 segments for empty files, got %d", len(segments))
	}
}

func TestBuildSegments_NilParsedDiff(t *testing.T) {
	left := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{makeAlignedLine("line1")},
	}
	right := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{makeAlignedLine("line1")},
	}

	segments := BuildSegments(left, right, nil)

	// Nil diff should produce no segments
	if segments != nil {
		t.Errorf("expected nil segments for nil parsedDiff, got %v", segments)
	}
}

func TestBuildSegments_OnlyAdditions(t *testing.T) {
	// New file - only additions
	left := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Add, Content: "line1", OldLineNo: 0, NewLineNo: 1},
			{Type: Add, Content: "line2", OldLineNo: 0, NewLineNo: 2},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Should produce a single HunkSegment
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	hs, ok := segments[0].(HunkSegment)
	if !ok {
		t.Fatalf("expected HunkSegment, got %T", segments[0])
	}

	if len(hs.LeftLines) != 0 {
		t.Errorf("expected 0 LeftLines, got %d", len(hs.LeftLines))
	}
	if len(hs.RightLines) != 2 {
		t.Errorf("expected 2 RightLines, got %d", len(hs.RightLines))
	}
}

func TestBuildSegments_OnlyRemovals(t *testing.T) {
	// File deleted - only removals
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
		},
	}
	right := FileContent{
		Path:  "test.txt",
		Lines: []AlignedLine{},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "line1", OldLineNo: 1, NewLineNo: 0},
			{Type: Remove, Content: "line2", OldLineNo: 2, NewLineNo: 0},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Should produce a single HunkSegment
	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	hs, ok := segments[0].(HunkSegment)
	if !ok {
		t.Fatalf("expected HunkSegment, got %T", segments[0])
	}

	if len(hs.LeftLines) != 2 {
		t.Errorf("expected 2 LeftLines, got %d", len(hs.LeftLines))
	}
	if len(hs.RightLines) != 0 {
		t.Errorf("expected 0 RightLines, got %d", len(hs.RightLines))
	}
}

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Line Type Verification
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_VerifyLineTypes(t *testing.T) {
	// Verify that HunkLine types are correctly assigned
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("removed1"),
			makeAlignedLine("removed2"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("added1"),
			makeAlignedLine("added2"),
			makeAlignedLine("added3"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Remove, Content: "removed1", OldLineNo: 1, NewLineNo: 0},
			{Type: Remove, Content: "removed2", OldLineNo: 2, NewLineNo: 0},
			{Type: Add, Content: "added1", OldLineNo: 0, NewLineNo: 1},
			{Type: Add, Content: "added2", OldLineNo: 0, NewLineNo: 2},
			{Type: Add, Content: "added3", OldLineNo: 0, NewLineNo: 3},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	if len(segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segments))
	}

	hs, ok := segments[0].(HunkSegment)
	if !ok {
		t.Fatalf("expected HunkSegment, got %T", segments[0])
	}

	// Verify left lines are all HunkLineRemoved
	for i, line := range hs.LeftLines {
		if line.Type != HunkLineRemoved {
			t.Errorf("LeftLines[%d]: expected HunkLineRemoved, got %d", i, line.Type)
		}
		if line.SourceIdx != i {
			t.Errorf("LeftLines[%d]: expected SourceIdx %d, got %d", i, i, line.SourceIdx)
		}
	}

	// Verify right lines are all HunkLineAdded
	for i, line := range hs.RightLines {
		if line.Type != HunkLineAdded {
			t.Errorf("RightLines[%d]: expected HunkLineAdded, got %d", i, line.Type)
		}
		if line.SourceIdx != i {
			t.Errorf("RightLines[%d]: expected SourceIdx %d, got %d", i, i, line.SourceIdx)
		}
	}
}

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Consecutive Unchanged Lines
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_ConsecutiveUnchangedLines_MergedIntoOne(t *testing.T) {
	// Multiple consecutive unchanged lines should be in a single UnchangedSegment
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("line3"),
			makeAlignedLine("line4"),
			makeAlignedLine("line5"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("line1"),
			makeAlignedLine("line2"),
			makeAlignedLine("changed"),
			makeAlignedLine("line4"),
			makeAlignedLine("line5"),
		},
	}

	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "line1", OldLineNo: 1, NewLineNo: 1},
			{Type: Context, Content: "line2", OldLineNo: 2, NewLineNo: 2},
			{Type: Remove, Content: "line3", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "changed", OldLineNo: 0, NewLineNo: 3},
			{Type: Context, Content: "line4", OldLineNo: 4, NewLineNo: 4},
			{Type: Context, Content: "line5", OldLineNo: 5, NewLineNo: 5},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Expected: UnchangedSegment(2 lines), HunkSegment, UnchangedSegment(2 lines)
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First unchanged: 2 lines
	us1, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us1.Count != 2 {
		t.Errorf("segment 0: expected Count 2, got %d", us1.Count)
	}

	// Hunk
	if _, ok := segments[1].(HunkSegment); !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}

	// Second unchanged: 2 lines
	us2, ok := segments[2].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 2: expected UnchangedSegment, got %T", segments[2])
	}
	if us2.Count != 2 {
		t.Errorf("segment 2: expected Count 2, got %d", us2.Count)
	}
	if us2.LeftStart != 3 || us2.RightStart != 3 {
		t.Errorf("segment 2: expected LeftStart 3, RightStart 3, got %d, %d", us2.LeftStart, us2.RightStart)
	}
}

// ═══════════════════════════════════════════════════════════
// BuildSegments Tests - Files Outside Diff Context
// ═══════════════════════════════════════════════════════════

func TestBuildSegments_LinesOutsideDiffContext(t *testing.T) {
	// When diff only covers middle portion of file, lines outside are unchanged
	left := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("before1"),
			makeAlignedLine("before2"),
			makeAlignedLine("old"),
			makeAlignedLine("after1"),
			makeAlignedLine("after2"),
		},
	}
	right := FileContent{
		Path: "test.txt",
		Lines: []AlignedLine{
			makeAlignedLine("before1"),
			makeAlignedLine("before2"),
			makeAlignedLine("new"),
			makeAlignedLine("after1"),
			makeAlignedLine("after2"),
		},
	}

	// Diff only shows context around the change (typical git diff behavior)
	parsedDiff := &FileDiff{
		OldPath: "test.txt",
		NewPath: "test.txt",
		Lines: []Line{
			{Type: Context, Content: "before2", OldLineNo: 2, NewLineNo: 2},
			{Type: Remove, Content: "old", OldLineNo: 3, NewLineNo: 0},
			{Type: Add, Content: "new", OldLineNo: 0, NewLineNo: 3},
			{Type: Context, Content: "after1", OldLineNo: 4, NewLineNo: 4},
		},
	}

	segments := BuildSegments(left, right, parsedDiff)

	// Should handle lines before and after the diff context
	// Expected: Unchanged(before1, before2), Hunk, Unchanged(after1, after2)
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	// First unchanged: before1 and before2
	us1, ok := segments[0].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 0: expected UnchangedSegment, got %T", segments[0])
	}
	if us1.Count != 2 {
		t.Errorf("segment 0: expected Count 2, got %d", us1.Count)
	}

	// Hunk
	if _, ok := segments[1].(HunkSegment); !ok {
		t.Fatalf("segment 1: expected HunkSegment, got %T", segments[1])
	}

	// Last unchanged: after1 and after2
	us2, ok := segments[2].(UnchangedSegment)
	if !ok {
		t.Fatalf("segment 2: expected UnchangedSegment, got %T", segments[2])
	}
	if us2.Count != 2 {
		t.Errorf("segment 2: expected Count 2, got %d", us2.Count)
	}
}
