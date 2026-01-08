package diff

import (
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
)

// DiffState represents the state when viewing a file diff
type State struct {
	// Current diff data
	CommitRange core.CommitRange
	File        core.FileChange
	Diff        *diff.AlignedFileDiff

	// Viewport control (legacy - alignment-based)
	// Keep ViewportStart temporarily for compatibility during migration to segment-based scrolling.
	// Will be removed once all code uses segment-based position.
	ViewportStart    int
	CurrentChangeIdx int   // Index into ChangeIndices for navigation
	ChangeIndices    []int // Indices of alignments that have changes (for navigation)

	// Segment-based scroll position (for smart diff scrolling)
	// These fields enable independent scrolling of left and right panels.
	SegmentIndex int // Which segment the viewport starts in
	LeftOffset   int // Line offset for left panel within segment
	RightOffset  int // Line offset for right panel within segment
}

// New creates a new DiffState with viewport positioned at the first change.
func New(commitRange core.CommitRange, file core.FileChange, d *diff.AlignedFileDiff, changeIndices []int) *State {
	viewportStart := 0
	if d != nil && len(changeIndices) > 0 {
		viewportStart = changeIndices[0]
	}

	// Initialize segment-based position to start at first hunk (if any exists)
	segmentIndex := 0
	if d != nil {
		segmentIndex = findFirstHunkSegmentIndex(d.Segments)
	}

	return &State{
		CommitRange:      commitRange,
		File:             file,
		Diff:             d,
		ChangeIndices:    changeIndices,
		ViewportStart:    viewportStart,
		CurrentChangeIdx: 0,
		SegmentIndex:     segmentIndex,
		LeftOffset:       0,
		RightOffset:      0,
	}
}

// findFirstHunkSegmentIndex returns the index of the first HunkSegment,
// or 0 if there are no hunks (start at beginning of file).
func findFirstHunkSegmentIndex(segments []diff.Segment) int {
	for i, seg := range segments {
		if _, isHunk := seg.(diff.HunkSegment); isHunk {
			return i
		}
	}
	return 0
}

// segmentLeftLineCount returns the number of left lines in the given segment.
// For UnchangedSegment, this is the Count field.
// For HunkSegment, this is the length of LeftLines.
func (s *State) segmentLeftLineCount(segIdx int) int {
	if s.Diff == nil || segIdx < 0 || segIdx >= len(s.Diff.Segments) {
		return 0
	}
	switch seg := s.Diff.Segments[segIdx].(type) {
	case diff.UnchangedSegment:
		return seg.Count
	case diff.HunkSegment:
		return len(seg.LeftLines)
	default:
		return 0
	}
}

// segmentRightLineCount returns the number of right lines in the given segment.
// For UnchangedSegment, this is the Count field (same as left).
// For HunkSegment, this is the length of RightLines.
func (s *State) segmentRightLineCount(segIdx int) int {
	if s.Diff == nil || segIdx < 0 || segIdx >= len(s.Diff.Segments) {
		return 0
	}
	switch seg := s.Diff.Segments[segIdx].(type) {
	case diff.UnchangedSegment:
		return seg.Count
	case diff.HunkSegment:
		return len(seg.RightLines)
	default:
		return 0
	}
}

// totalLeftLines returns total lines on left side across all segments.
func (s *State) totalLeftLines() int {
	if s.Diff == nil {
		return 0
	}
	total := 0
	for i := range s.Diff.Segments {
		total += s.segmentLeftLineCount(i)
	}
	return total
}

// totalRightLines returns total lines on right side across all segments.
func (s *State) totalRightLines() int {
	if s.Diff == nil {
		return 0
	}
	total := 0
	for i := range s.Diff.Segments {
		total += s.segmentRightLineCount(i)
	}
	return total
}
