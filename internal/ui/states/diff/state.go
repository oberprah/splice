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

	// ScrollAccumulator tracks fractional scroll progress for the slower side
	// during differential scrolling. It is incremented each scroll step and
	// determines when the slower side should advance (when accumulator reaches ratio).
	ScrollAccumulator int
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

// isAtStart returns true if we're at the beginning of all segments.
func (s *State) isAtStart() bool {
	return s.SegmentIndex == 0 && s.LeftOffset == 0 && s.RightOffset == 0
}

// isAtEnd returns true if we've scrolled to the end of all segments.
// We're at the end when we're past the last segment or when in the last segment
// and both sides have been fully scrolled through (offsets >= line counts).
func (s *State) isAtEnd(viewportHeight int) bool {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return true
	}

	// Past all segments
	if s.SegmentIndex >= len(s.Diff.Segments) {
		return true
	}

	// Not at end if we're not in the last segment
	if s.SegmentIndex < len(s.Diff.Segments)-1 {
		return false
	}

	// In the last segment - check if both sides are exhausted
	leftCount := s.segmentLeftLineCount(s.SegmentIndex)
	rightCount := s.segmentRightLineCount(s.SegmentIndex)

	// We're at the end only when we've scrolled through all content
	// For hunks, we need to check if both sides are exhausted
	// For unchanged segments, left and right are always equal
	return s.LeftOffset >= leftCount && s.RightOffset >= rightCount
}

// isHunkCentered returns true if the current segment is a hunk and overlaps
// the center zone of the viewport. The center zone is the middle 40% of the viewport
// (from 30% to 70% of viewport height).
func (s *State) isHunkCentered(viewportHeight int) bool {
	if s.Diff == nil || s.SegmentIndex >= len(s.Diff.Segments) {
		return false
	}

	// Check if current segment is a hunk
	hunk, isHunk := s.Diff.Segments[s.SegmentIndex].(diff.HunkSegment)
	if !isHunk {
		return false
	}

	// Calculate where the hunk appears in the viewport
	// Since we're at SegmentIndex with LeftOffset/RightOffset, the hunk starts at row 0
	// (we're already inside the hunk at offset LeftOffset/RightOffset)
	// The hunk occupies rows 0 to max(leftRemaining, rightRemaining) - 1

	leftCount := len(hunk.LeftLines)
	rightCount := len(hunk.RightLines)
	leftRemaining := leftCount - s.LeftOffset
	rightRemaining := rightCount - s.RightOffset
	hunkVisibleLines := max(leftRemaining, rightRemaining)

	// Hunk top is at row 0 (since we're positioned at this segment)
	hunkTopRow := 0
	hunkBottomRow := hunkVisibleLines

	// Calculate center zone (30% to 70% of viewport)
	centerStart := viewportHeight * 30 / 100
	centerEnd := viewportHeight * 70 / 100

	// Hunk overlaps center zone if its extent intersects [centerStart, centerEnd)
	return hunkTopRow < centerEnd && hunkBottomRow > centerStart
}

// scrollDownSegment handles scrolling down by one line using segment-based position.
// Implements differential scrolling when a hunk is centered in the viewport.
func (s *State) scrollDownSegment(viewportHeight int) {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return
	}

	// Check if we're at the end
	if s.isAtEnd(viewportHeight) {
		return
	}

	seg := s.Diff.Segments[s.SegmentIndex]

	switch segment := seg.(type) {
	case diff.UnchangedSegment:
		// Unchanged segment: both panels scroll together
		s.LeftOffset++
		s.RightOffset++
		s.ScrollAccumulator = 0

		// Check if we've exhausted this segment
		if s.LeftOffset >= segment.Count {
			s.SegmentIndex++
			s.LeftOffset = 0
			s.RightOffset = 0
		}

	case diff.HunkSegment:
		leftCount := len(segment.LeftLines)
		rightCount := len(segment.RightLines)

		// Check if hunk is centered for differential scrolling
		if s.isHunkCentered(viewportHeight) && leftCount != rightCount {
			s.scrollHunkDifferential(leftCount, rightCount)
		} else {
			// Normal scrolling through hunk (both advance together when not centered)
			s.LeftOffset++
			s.RightOffset++
			s.ScrollAccumulator = 0
		}

		// Check if we've exhausted this hunk (both sides done)
		if s.LeftOffset >= leftCount && s.RightOffset >= rightCount {
			s.SegmentIndex++
			s.LeftOffset = 0
			s.RightOffset = 0
			s.ScrollAccumulator = 0
		}
	}
}

// scrollHunkDifferential applies differential scrolling rates.
// The side with more lines scrolls every step, the side with fewer lines
// scrolls only when the accumulator reaches the ratio.
func (s *State) scrollHunkDifferential(leftCount, rightCount int) {
	// Calculate ratio: larger / smaller
	var ratio int
	leftLarger := leftCount > rightCount

	if leftLarger {
		ratio = leftCount / rightCount
		if leftCount%rightCount != 0 {
			ratio++ // Round up for smoother distribution
		}
	} else {
		ratio = rightCount / leftCount
		if rightCount%leftCount != 0 {
			ratio++
		}
	}

	// Advance the larger side every step
	// Advance the smaller side every 'ratio' steps
	if leftLarger {
		// Left is larger: always advance left
		if s.LeftOffset < leftCount {
			s.LeftOffset++
		}
		// Right advances only when accumulator reaches ratio
		s.ScrollAccumulator++
		if s.ScrollAccumulator >= ratio {
			if s.RightOffset < rightCount {
				s.RightOffset++
			}
			s.ScrollAccumulator = 0
		}
	} else {
		// Right is larger: always advance right
		if s.RightOffset < rightCount {
			s.RightOffset++
		}
		// Left advances only when accumulator reaches ratio
		s.ScrollAccumulator++
		if s.ScrollAccumulator >= ratio {
			if s.LeftOffset < leftCount {
				s.LeftOffset++
			}
			s.ScrollAccumulator = 0
		}
	}
}

// scrollUpSegment handles scrolling up by one line using segment-based position.
// Implements differential scrolling when a hunk is centered in the viewport.
func (s *State) scrollUpSegment(viewportHeight int) {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return
	}

	// Check if we're at the start
	if s.isAtStart() {
		return
	}

	// If at the start of current segment, move to previous segment
	if s.LeftOffset == 0 && s.RightOffset == 0 {
		if s.SegmentIndex > 0 {
			s.SegmentIndex--
			// Set offsets to end of previous segment
			leftCount := s.segmentLeftLineCount(s.SegmentIndex)
			rightCount := s.segmentRightLineCount(s.SegmentIndex)
			s.LeftOffset = leftCount
			s.RightOffset = rightCount
			s.ScrollAccumulator = 0
		}
	}

	seg := s.Diff.Segments[s.SegmentIndex]

	switch segment := seg.(type) {
	case diff.UnchangedSegment:
		// Unchanged segment: both panels scroll together
		if s.LeftOffset > 0 {
			s.LeftOffset--
			s.RightOffset--
		}
		s.ScrollAccumulator = 0

	case diff.HunkSegment:
		leftCount := len(segment.LeftLines)
		rightCount := len(segment.RightLines)

		// Check if hunk is centered for differential scrolling
		if s.isHunkCentered(viewportHeight) && leftCount != rightCount {
			s.scrollHunkDifferentialUp(leftCount, rightCount)
		} else {
			// Normal scrolling through hunk (both retreat together when not centered)
			if s.LeftOffset > 0 {
				s.LeftOffset--
			}
			if s.RightOffset > 0 {
				s.RightOffset--
			}
			s.ScrollAccumulator = 0
		}
	}
}

// scrollHunkDifferentialUp applies differential scrolling rates for scrolling up.
// Symmetric to scrollHunkDifferential but in reverse.
func (s *State) scrollHunkDifferentialUp(leftCount, rightCount int) {
	// Calculate ratio: larger / smaller
	var ratio int
	leftLarger := leftCount > rightCount

	if leftLarger {
		ratio = leftCount / rightCount
		if leftCount%rightCount != 0 {
			ratio++
		}
	} else {
		ratio = rightCount / leftCount
		if rightCount%leftCount != 0 {
			ratio++
		}
	}

	// Retreat the larger side every step
	// Retreat the smaller side every 'ratio' steps
	if leftLarger {
		// Left is larger: always retreat left
		if s.LeftOffset > 0 {
			s.LeftOffset--
		}
		// Right retreats only when accumulator reaches ratio
		s.ScrollAccumulator++
		if s.ScrollAccumulator >= ratio {
			if s.RightOffset > 0 {
				s.RightOffset--
			}
			s.ScrollAccumulator = 0
		}
	} else {
		// Right is larger: always retreat right
		if s.RightOffset > 0 {
			s.RightOffset--
		}
		// Left retreats only when accumulator reaches ratio
		s.ScrollAccumulator++
		if s.ScrollAccumulator >= ratio {
			if s.LeftOffset > 0 {
				s.LeftOffset--
			}
			s.ScrollAccumulator = 0
		}
	}
}

// resetToStart resets scroll position to the beginning of all segments.
func (s *State) resetToStart() {
	s.SegmentIndex = 0
	s.LeftOffset = 0
	s.RightOffset = 0
	s.ScrollAccumulator = 0
}

// scrollToEnd positions the viewport at the end of all segments.
func (s *State) scrollToEnd(viewportHeight int) {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return
	}

	// Calculate total lines on left side
	totalLeft := s.totalLeftLines()

	// We need to position so that the last lines are visible
	// Work backwards from the end to find the right segment and offset
	targetLeftStart := totalLeft - viewportHeight
	if targetLeftStart < 0 {
		targetLeftStart = 0
	}

	// Find the segment that contains targetLeftStart
	leftAccum := 0
	for i, seg := range s.Diff.Segments {
		leftCount := s.segmentLeftLineCount(i)
		if leftAccum+leftCount > targetLeftStart {
			s.SegmentIndex = i
			s.LeftOffset = targetLeftStart - leftAccum
			// For unchanged segments, right offset equals left offset
			if _, isUnchanged := seg.(diff.UnchangedSegment); isUnchanged {
				s.RightOffset = s.LeftOffset
			} else {
				// For hunks, calculate right offset proportionally
				rightCount := s.segmentRightLineCount(i)
				if leftCount > 0 {
					s.RightOffset = s.LeftOffset * rightCount / leftCount
				} else {
					s.RightOffset = 0
				}
			}
			s.ScrollAccumulator = 0
			return
		}
		leftAccum += leftCount
	}

	// If we get here, start at the last segment
	s.SegmentIndex = len(s.Diff.Segments) - 1
	s.LeftOffset = s.segmentLeftLineCount(s.SegmentIndex)
	s.RightOffset = s.segmentRightLineCount(s.SegmentIndex)
	s.ScrollAccumulator = 0
}

// jumpToNextHunkSegment navigates to the next HunkSegment after current position.
// Resets offsets to 0 to position at the start of the hunk.
// If already at or past the last hunk, does nothing.
func (s *State) jumpToNextHunkSegment() {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return
	}

	// Search for next hunk starting from SegmentIndex + 1
	for i := s.SegmentIndex + 1; i < len(s.Diff.Segments); i++ {
		if _, isHunk := s.Diff.Segments[i].(diff.HunkSegment); isHunk {
			s.SegmentIndex = i
			s.LeftOffset = 0
			s.RightOffset = 0
			s.ScrollAccumulator = 0
			return
		}
	}
	// No next hunk found - stay at current position
}

// jumpToPreviousHunkSegment navigates to the previous HunkSegment before current position.
// Resets offsets to 0 to position at the start of the hunk.
// If already at or before the first hunk, does nothing.
func (s *State) jumpToPreviousHunkSegment() {
	if s.Diff == nil || len(s.Diff.Segments) == 0 {
		return
	}

	// Search for previous hunk starting from SegmentIndex - 1
	for i := s.SegmentIndex - 1; i >= 0; i-- {
		if _, isHunk := s.Diff.Segments[i].(diff.HunkSegment); isHunk {
			s.SegmentIndex = i
			s.LeftOffset = 0
			s.RightOffset = 0
			s.ScrollAccumulator = 0
			return
		}
	}
	// No previous hunk found - stay at current position
}
