package diff

// BuildSegments creates a segment-based representation of the diff.
// It walks through both files using the diff information to identify
// unchanged regions and hunks.
//
// The algorithm builds maps from line numbers to diff types, then walks through
// both files using two pointers. Consecutive unchanged lines are accumulated
// into UnchangedSegments, while consecutive changed lines (removed/added) are
// accumulated into HunkSegments.
//
// For this initial implementation, all removed/added lines are treated as
// HunkLineRemoved/HunkLineAdded respectively. Inline diff pairing (which would
// mark some lines as HunkLineModified) is deferred to a future enhancement.
func BuildSegments(left, right FileContent, parsedDiff *FileDiff) []Segment {
	if parsedDiff == nil {
		return nil
	}

	// Build maps from 1-indexed line numbers to diff line types
	leftDiffMap := make(map[int]LineType)
	rightDiffMap := make(map[int]LineType)

	for _, line := range parsedDiff.Lines {
		if line.OldLineNo > 0 {
			leftDiffMap[line.OldLineNo] = line.Type
		}
		if line.NewLineNo > 0 {
			rightDiffMap[line.NewLineNo] = line.Type
		}
	}

	var segments []Segment
	leftIdx := 0  // Current position in left file (0-indexed)
	rightIdx := 0 // Current position in right file (0-indexed)

	// Accumulators for building segments
	var unchangedStart *unchangedAccum
	var hunkLeft []HunkLine
	var hunkRight []HunkLine

	// Helper to flush accumulated unchanged lines as a segment
	flushUnchanged := func() {
		if unchangedStart != nil && unchangedStart.count > 0 {
			segments = append(segments, UnchangedSegment{
				LeftStart:  unchangedStart.leftStart,
				RightStart: unchangedStart.rightStart,
				Count:      unchangedStart.count,
			})
			unchangedStart = nil
		}
	}

	// Helper to flush accumulated hunk lines as a segment
	flushHunk := func() {
		if len(hunkLeft) > 0 || len(hunkRight) > 0 {
			segments = append(segments, HunkSegment{
				LeftLines:  hunkLeft,
				RightLines: hunkRight,
			})
			hunkLeft = nil
			hunkRight = nil
		}
	}

	// Walk through both files
	for leftIdx < len(left.Lines) || rightIdx < len(right.Lines) {
		leftLineNo := leftIdx + 1   // 1-indexed
		rightLineNo := rightIdx + 1 // 1-indexed

		// Determine if current lines are changed or unchanged
		leftType, leftInDiff := leftDiffMap[leftLineNo]
		rightType, rightInDiff := rightDiffMap[rightLineNo]

		leftIsUnchanged := !leftInDiff || leftType == Context
		rightIsUnchanged := !rightInDiff || rightType == Context

		// Case 1: Both lines are unchanged -> accumulate into unchanged segment
		if leftIdx < len(left.Lines) && rightIdx < len(right.Lines) &&
			leftIsUnchanged && rightIsUnchanged {
			// Flush any pending hunk before starting/continuing unchanged
			flushHunk()

			if unchangedStart == nil {
				unchangedStart = &unchangedAccum{
					leftStart:  leftIdx,
					rightStart: rightIdx,
					count:      0,
				}
			}
			unchangedStart.count++
			leftIdx++
			rightIdx++
			continue
		}

		// Case 2: Left line is removed -> accumulate into hunk
		if leftIdx < len(left.Lines) && leftInDiff && leftType == Remove {
			// Flush any pending unchanged before starting/continuing hunk
			flushUnchanged()

			hunkLeft = append(hunkLeft, HunkLine{
				SourceIdx: leftIdx,
				Type:      HunkLineRemoved,
			})
			leftIdx++
			continue
		}

		// Case 3: Right line is added -> accumulate into hunk
		if rightIdx < len(right.Lines) && rightInDiff && rightType == Add {
			// Flush any pending unchanged before starting/continuing hunk
			flushUnchanged()

			hunkRight = append(hunkRight, HunkLine{
				SourceIdx: rightIdx,
				Type:      HunkLineAdded,
			})
			rightIdx++
			continue
		}

		// Case 4: One side finished but other has remaining lines
		if leftIdx >= len(left.Lines) && rightIdx < len(right.Lines) {
			// Only right lines remaining
			if rightInDiff && rightType == Add {
				flushUnchanged()
				hunkRight = append(hunkRight, HunkLine{
					SourceIdx: rightIdx,
					Type:      HunkLineAdded,
				})
			}
			rightIdx++
			continue
		}

		if rightIdx >= len(right.Lines) && leftIdx < len(left.Lines) {
			// Only left lines remaining
			if leftInDiff && leftType == Remove {
				flushUnchanged()
				hunkLeft = append(hunkLeft, HunkLine{
					SourceIdx: leftIdx,
					Type:      HunkLineRemoved,
				})
			}
			leftIdx++
			continue
		}

		// Fallback: advance both pointers (shouldn't normally reach here)
		leftIdx++
		rightIdx++
	}

	// Flush any remaining accumulated lines
	flushUnchanged()
	flushHunk()

	return segments
}

// unchangedAccum tracks the start indices and count of consecutive unchanged lines
type unchangedAccum struct {
	leftStart  int
	rightStart int
	count      int
}
