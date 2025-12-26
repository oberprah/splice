package diff

import (
	"github.com/oberprah/splice/internal/highlight"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// BuildFileContent takes raw file content and a file path, applies syntax highlighting,
// and returns a FileContent struct containing lines with syntax tokens.
//
// The path is used to determine the appropriate lexer for syntax highlighting.
// Content is split by newlines, and each line is tokenized independently.
//
// Returns a FileContent with:
//   - Path: the provided file path
//   - Lines: slice of AlignedLine structs, each containing syntax tokens
//
// Example:
//
//	content := "package main\n\nfunc hello() {\n}"
//	fc, err := BuildFileContent("main.go", content)
//	// fc.Lines[0] contains tokens for "package main"
//	// fc.Lines[1] contains tokens for empty line
//	// fc.Lines[2] contains tokens for "func hello() {"
func BuildFileContent(path string, content string) (FileContent, error) {
	// Tokenize the entire file using Chroma syntax highlighter
	// Returns [][]Token where each inner slice represents one line
	lineTokens := highlight.TokenizeFile(content, path)

	// Convert from [][]highlight.Token to []AlignedLine
	lines := make([]AlignedLine, len(lineTokens))
	for i, tokens := range lineTokens {
		lines[i] = AlignedLine{Tokens: tokens}
	}

	return FileContent{
		Path:  path,
		Lines: lines,
	}, nil
}

// BuildAlignments takes file content for both sides and parsed diff hunks,
// then produces an alignment sequence that describes how lines relate.
//
// The algorithm walks through both files line by line, emitting:
//   - UnchangedAlignment for lines that exist in both files
//   - ModifiedAlignment for paired changed lines (with inline diffs)
//   - RemovedAlignment for unpaired removed lines
//   - AddedAlignment for unpaired added lines
//
// Processing operates at the "hunk" level - contiguous regions of changes
// bounded by unchanged lines. Within each hunk:
//  1. Collect all removed and added line indices
//  2. Use pairLines() to match them based on similarity
//  3. Compute inline character-level diffs for paired lines
//  4. Emit alignments in the correct order
//
// Parameters:
//   - left: FileContent for the old version of the file
//   - right: FileContent for the new version of the file
//   - parsedDiff: FileDiff from ParseUnifiedDiff containing the hunks
//
// Returns:
//   - []Alignment: one entry per display row
//
// Example:
//
//	Left file:           Right file:         Result:
//	Line 1: unchanged    Line 1: unchanged   UnchangedAlignment{0, 0}
//	Line 2: old text     Line 2: new text    ModifiedAlignment{1, 1, diffs}
//	Line 3: removed                          RemovedAlignment{2}
//	                     Line 3: added       AddedAlignment{2}
func BuildAlignments(left, right FileContent, parsedDiff *FileDiff) []Alignment {
	if parsedDiff == nil {
		return nil
	}

	// Build maps from line numbers to diff information
	// Key: 1-indexed line number, Value: diff line with type and content
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

	var alignments []Alignment
	leftIdx := 0  // Current position in left file (0-indexed)
	rightIdx := 0 // Current position in right file (0-indexed)

	// Variables to collect changes within a hunk
	var hunkRemoved []int // Indices into left.Lines
	var hunkAdded []int   // Indices into right.Lines

	// flushHunk processes collected removed/added lines and emits alignments
	flushHunk := func() {
		if len(hunkRemoved) == 0 && len(hunkAdded) == 0 {
			return
		}

		// Extract the lines for pairing
		removedLines := make([]AlignedLine, len(hunkRemoved))
		addedLines := make([]AlignedLine, len(hunkAdded))
		for i, idx := range hunkRemoved {
			removedLines[i] = left.Lines[idx]
		}
		for i, idx := range hunkAdded {
			addedLines[i] = right.Lines[idx]
		}

		// Pair the lines based on similarity
		pairs, unpairedRemoved, unpairedAdded := pairLines(removedLines, addedLines)

		// Emit ModifiedAlignment for each pair with inline diffs
		for _, pair := range pairs {
			removedIdx := hunkRemoved[pair[0]]
			addedIdx := hunkAdded[pair[1]]

			// Compute inline character-level diff
			dmp := diffmatchpatch.New()
			leftText := left.Lines[removedIdx].Text()
			rightText := right.Lines[addedIdx].Text()
			diffs := dmp.DiffMain(leftText, rightText, false)

			alignments = append(alignments, ModifiedAlignment{
				LeftIdx:    removedIdx,
				RightIdx:   addedIdx,
				InlineDiff: diffs,
			})
		}

		// Emit RemovedAlignment for unpaired removed lines
		for _, idx := range unpairedRemoved {
			alignments = append(alignments, RemovedAlignment{
				LeftIdx: hunkRemoved[idx],
			})
		}

		// Emit AddedAlignment for unpaired added lines
		for _, idx := range unpairedAdded {
			alignments = append(alignments, AddedAlignment{
				RightIdx: hunkAdded[idx],
			})
		}

		// Clear the hunk buffers
		hunkRemoved = nil
		hunkAdded = nil
	}

	// Walk through both files using indices
	for leftIdx < len(left.Lines) || rightIdx < len(right.Lines) {
		leftLineNo := leftIdx + 1  // 1-indexed line number for left file
		rightLineNo := rightIdx + 1 // 1-indexed line number for right file

		// Check the diff status of current lines
		leftType, leftInDiff := leftDiffMap[leftLineNo]
		rightType, rightInDiff := rightDiffMap[rightLineNo]

		// Determine the action based on diff status
		leftIsUnchanged := !leftInDiff || leftType == Context
		rightIsUnchanged := !rightInDiff || rightType == Context

		// Case 1: Both lines are unchanged -> emit UnchangedAlignment and advance both
		if leftIdx < len(left.Lines) && rightIdx < len(right.Lines) &&
			leftIsUnchanged && rightIsUnchanged {
			// Flush any pending hunk before emitting unchanged
			flushHunk()

			alignments = append(alignments, UnchangedAlignment{
				LeftIdx:  leftIdx,
				RightIdx: rightIdx,
			})
			leftIdx++
			rightIdx++
			continue
		}

		// Case 2: Left line is removed -> collect in hunk
		if leftIdx < len(left.Lines) && leftInDiff && leftType == Remove {
			hunkRemoved = append(hunkRemoved, leftIdx)
			leftIdx++
			continue
		}

		// Case 3: Right line is added -> collect in hunk
		if rightIdx < len(right.Lines) && rightInDiff && rightType == Add {
			hunkAdded = append(hunkAdded, rightIdx)
			rightIdx++
			continue
		}

		// Case 4: One side finished but other has remaining lines
		if leftIdx >= len(left.Lines) && rightIdx < len(right.Lines) {
			// Only right lines remaining
			if rightInDiff && rightType == Add {
				hunkAdded = append(hunkAdded, rightIdx)
			}
			rightIdx++
			continue
		}

		if rightIdx >= len(right.Lines) && leftIdx < len(left.Lines) {
			// Only left lines remaining
			if leftInDiff && leftType == Remove {
				hunkRemoved = append(hunkRemoved, leftIdx)
			}
			leftIdx++
			continue
		}

		// Fallback: advance both pointers (shouldn't normally reach here)
		leftIdx++
		rightIdx++
	}

	// Flush any remaining hunk at the end
	flushHunk()

	return alignments
}
