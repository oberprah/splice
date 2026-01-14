package diff

import (
	"fmt"

	"github.com/oberprah/splice/internal/domain/highlight"
	"github.com/sergi/go-diff/diffmatchpatch"
)

// BuildFileDiff is the main entry point for building a block-based file diff.
// It takes file path, old content, new content, and diff output, returning
// a FileDiff with blocks grouped for navigation.
//
// The function groups consecutive unchanged lines into UnchangedBlock and
// consecutive changed lines into ChangeBlock, enabling efficient navigation
// through changes.
//
// Parameters:
//   - filePath: Path to the file (used for syntax highlighting lexer selection)
//   - oldContent: Content of the file before changes
//   - newContent: Content of the file after changes
//   - diffOutput: Raw unified diff output from git
//
// Returns:
//   - *FileDiff: Complete diff structure with blocks for navigation
//   - error: Any error that occurred during processing
func BuildFileDiff(
	filePath string,
	oldContent string,
	newContent string,
	diffOutput string,
) (*FileDiff, error) {
	// Step 1: Parse the unified diff output
	parsedDiff, err := ParseUnifiedDiff(diffOutput)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	// Step 2: Build file content with syntax highlighting for old version
	leftContent, err := BuildFileContent(filePath, oldContent)
	if err != nil {
		return nil, fmt.Errorf("failed to build left content: %w", err)
	}

	// Step 3: Build file content with syntax highlighting for new version
	rightContent, err := BuildFileContent(filePath, newContent)
	if err != nil {
		return nil, fmt.Errorf("failed to build right content: %w", err)
	}

	// Step 4: Build blocks from the content and diff
	blocks := buildBlocks(leftContent, rightContent, &parsedDiff)

	return &FileDiff{
		Path:   filePath,
		Blocks: blocks,
	}, nil
}

// buildBlocks constructs blocks from file content and parsed diff.
// It groups consecutive unchanged lines into UnchangedBlock and
// consecutive changed lines into ChangeBlock.
func buildBlocks(left, right FileContent, parsedDiff *ParsedFileDiff) []Block {
	if parsedDiff == nil {
		return nil
	}

	// Build maps from line numbers to diff information
	// Key: 1-indexed line number, Value: line type
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

	var blocks []Block
	var currentUnchanged []LinePair
	var currentChanged []ChangeLine

	// Helper to flush accumulated unchanged lines
	flushUnchanged := func() {
		if len(currentUnchanged) > 0 {
			blocks = append(blocks, UnchangedBlock{Lines: currentUnchanged})
			currentUnchanged = nil
		}
	}

	// Helper to flush accumulated changed lines
	flushChanged := func() {
		if len(currentChanged) > 0 {
			blocks = append(blocks, ChangeBlock{Lines: currentChanged})
			currentChanged = nil
		}
	}

	// Variables to collect changes within a hunk
	var hunkRemoved []int // Indices into left.Lines
	var hunkAdded []int   // Indices into right.Lines

	// flushHunk processes collected removed/added lines and emits ChangeLine entries
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

		// Emit ModifiedLine for each pair with inline diffs
		for _, pair := range pairs {
			removedIdx := hunkRemoved[pair[0]]
			addedIdx := hunkAdded[pair[1]]

			// Compute inline character-level diff
			dmp := diffmatchpatch.New()
			leftText := left.Lines[removedIdx].Text()
			rightText := right.Lines[addedIdx].Text()
			diffs := dmp.DiffMain(leftText, rightText, false)

			currentChanged = append(currentChanged, ModifiedLine{
				LeftLineNo:  removedIdx + 1, // Convert to 1-indexed
				RightLineNo: addedIdx + 1,   // Convert to 1-indexed
				LeftTokens:  left.Lines[removedIdx].Tokens,
				RightTokens: right.Lines[addedIdx].Tokens,
				InlineDiff:  diffs,
			})
		}

		// Emit RemovedLine for unpaired removed lines
		for _, idx := range unpairedRemoved {
			leftIdx := hunkRemoved[idx]
			currentChanged = append(currentChanged, RemovedLine{
				LeftLineNo: leftIdx + 1, // Convert to 1-indexed
				Tokens:     left.Lines[leftIdx].Tokens,
			})
		}

		// Emit AddedLine for unpaired added lines
		for _, idx := range unpairedAdded {
			rightIdx := hunkAdded[idx]
			currentChanged = append(currentChanged, AddedLine{
				RightLineNo: rightIdx + 1, // Convert to 1-indexed
				Tokens:      right.Lines[rightIdx].Tokens,
			})
		}

		// Clear the hunk buffers
		hunkRemoved = nil
		hunkAdded = nil
	}

	leftIdx := 0  // Current position in left file (0-indexed)
	rightIdx := 0 // Current position in right file (0-indexed)

	// Walk through both files using indices
	for leftIdx < len(left.Lines) || rightIdx < len(right.Lines) {
		leftLineNo := leftIdx + 1   // 1-indexed line number for left file
		rightLineNo := rightIdx + 1 // 1-indexed line number for right file

		// Check the diff status of current lines
		leftType, leftInDiff := leftDiffMap[leftLineNo]
		rightType, rightInDiff := rightDiffMap[rightLineNo]

		// Determine the action based on diff status
		leftIsUnchanged := !leftInDiff || leftType == Context
		rightIsUnchanged := !rightInDiff || rightType == Context

		// Case 1: Both lines are unchanged -> emit LinePair and advance both
		if leftIdx < len(left.Lines) && rightIdx < len(right.Lines) &&
			leftIsUnchanged && rightIsUnchanged {
			// Flush any pending hunk and changed lines before adding unchanged
			flushHunk()
			flushChanged()

			currentUnchanged = append(currentUnchanged, LinePair{
				LeftLineNo:  leftLineNo,
				RightLineNo: rightLineNo,
				Tokens:      left.Lines[leftIdx].Tokens, // Use left tokens (content is identical)
			})
			leftIdx++
			rightIdx++
			continue
		}

		// Flush any pending unchanged lines before collecting changes
		flushUnchanged()

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

	// Flush any remaining hunk and blocks at the end
	flushHunk()
	flushChanged()
	flushUnchanged()

	return blocks
}

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
