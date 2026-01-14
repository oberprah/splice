package diff

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/domain/diff"
	"github.com/oberprah/splice/internal/git"
)

// Update handles messages for the diff state
func (s *State) Update(msg tea.Msg, ctx core.Context) (core.State, tea.Cmd) {
	switch msg := msg.(type) {
	case core.DiffLoadedMsg:
		// Handle async diff loading result (from file navigation)
		if msg.Err != nil {
			// On error, stay in current state - could push error screen in future
			return s, nil
		}
		// Update state with new file's diff
		s.Source = msg.Source
		s.Files = msg.Files
		s.FileIndex = msg.FileIndex
		s.File = msg.File
		s.Diff = msg.Diff
		// Position at first change block
		s.positionAtFirstChange()
		return s, nil

	case EditorFinishedMsg:
		// Handle editor completion
		if msg.err != nil {
			// Push error screen to show the error
			return s, func() tea.Msg {
				return core.PushErrorScreenMsg{Err: msg.err}
			}
		}
		// Success case - editor exited cleanly, just resume
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			// Go back to the previous state using navigation pattern
			return s, func() tea.Msg {
				return core.PopScreenMsg{}
			}

		case "ctrl+c", "Q":
			return s, tea.Quit

		case "j", "down":
			// Scroll down
			if s.Diff != nil && len(s.Diff.Blocks) > 0 {
				maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
				if s.ViewportStart < maxViewportStart {
					s.ViewportStart++
				}
			}
			return s, nil

		case "k", "up":
			// Scroll up
			if s.ViewportStart > 0 {
				s.ViewportStart--
			}
			return s, nil

		case "ctrl+d":
			// Scroll down half page
			if s.Diff != nil && len(s.Diff.Blocks) > 0 {
				headerLines := 2
				availableHeight := max(ctx.Height()-headerLines, 1)
				halfPage := availableHeight / 2
				maxViewportStart := s.calculateMaxViewportStart(ctx.Height())
				s.ViewportStart = min(s.ViewportStart+halfPage, maxViewportStart)
			}
			return s, nil

		case "ctrl+u":
			// Scroll up half page
			headerLines := 2
			availableHeight := max(ctx.Height()-headerLines, 1)
			halfPage := availableHeight / 2
			s.ViewportStart = max(s.ViewportStart-halfPage, 0)
			return s, nil

		case "g":
			// Jump to top
			s.ViewportStart = 0
			return s, nil

		case "G":
			// Jump to bottom
			if s.Diff != nil && len(s.Diff.Blocks) > 0 {
				s.ViewportStart = s.calculateMaxViewportStart(ctx.Height())
			}
			return s, nil

		case "n":
			// Smart next change navigation
			return s.navigateToNextChange(ctx)

		case "p":
			// Smart previous change navigation
			return s.navigateToPrevChange(ctx)

		case "o":
			// Open file in editor
			return s, s.openFileInEditor()

		case "]":
			// Navigate to next file
			return s.navigateToNextFile(ctx)

		case "[":
			// Navigate to previous file
			return s.navigateToPrevFile(ctx)
		}
	}

	return s, nil
}

// navigateToNextChange implements smart next change navigation.
// If in a ChangeBlock that extends below viewport, scrolls down half page.
// Otherwise jumps to the next ChangeBlock.
// If at the last change block in the file, navigates to the next file.
func (s *State) navigateToNextChange(ctx core.Context) (*State, tea.Cmd) {
	if s.Diff == nil || len(s.Diff.Blocks) == 0 {
		return s, nil
	}

	height := ctx.Height()
	headerLines := 2
	availableHeight := max(height-headerLines, 1)
	halfPage := availableHeight / 2

	// Find which block we're currently in
	currentBlockIdx, _ := s.getBlockAtPosition(s.ViewportStart)

	// Check if we're in a ChangeBlock
	if currentBlockIdx >= 0 && currentBlockIdx < len(s.Diff.Blocks) {
		if _, isChange := s.Diff.Blocks[currentBlockIdx].(diff.ChangeBlock); isChange {
			// We're in a change block - check if we need to scroll through it
			if !s.isChangeBlockEndVisible(currentBlockIdx, availableHeight) {
				// Scroll down half page to show more of current change
				maxViewport := s.calculateMaxViewportStart(height)
				s.ViewportStart = min(s.ViewportStart+halfPage, maxViewport)
				return s, nil
			}
		}
	}

	// Find next change block
	nextChangeIdx := s.findNextChangeBlock(currentBlockIdx)
	if nextChangeIdx != -1 {
		// Jump to next change block
		s.CurrentBlockIdx = nextChangeIdx
		s.ViewportStart = s.getBlockStartPosition(nextChangeIdx)
		maxViewport := s.calculateMaxViewportStart(height)
		if s.ViewportStart > maxViewport {
			s.ViewportStart = maxViewport
		}
		return s, nil
	}

	// No more change blocks in this file - navigate to next file
	if s.FileIndex < len(s.Files)-1 {
		return s.navigateToNextFile(ctx)
	}
	// At last file, stay in place
	return s, nil
}

// navigateToPrevChange implements smart previous change navigation.
// If in a ChangeBlock that extends above viewport, scrolls up half page.
// Otherwise jumps to the previous ChangeBlock.
// If at the first change block in the file, navigates to the previous file.
func (s *State) navigateToPrevChange(ctx core.Context) (*State, tea.Cmd) {
	if s.Diff == nil || len(s.Diff.Blocks) == 0 {
		return s, nil
	}

	height := ctx.Height()
	headerLines := 2
	availableHeight := max(height-headerLines, 1)
	halfPage := availableHeight / 2

	// Find which block we're currently in
	currentBlockIdx, _ := s.getBlockAtPosition(s.ViewportStart)

	// Check if we're in a ChangeBlock
	if currentBlockIdx >= 0 && currentBlockIdx < len(s.Diff.Blocks) {
		if _, isChange := s.Diff.Blocks[currentBlockIdx].(diff.ChangeBlock); isChange {
			blockStart := s.getBlockStartPosition(currentBlockIdx)

			// Check if we're not at the start of this change block
			if s.ViewportStart > blockStart {
				// Check if start is above viewport
				if !s.isChangeBlockStartVisible(currentBlockIdx) {
					// Scroll up half page to show more of current change
					s.ViewportStart = max(s.ViewportStart-halfPage, 0)
					return s, nil
				}
				// Jump to start of this change block
				s.ViewportStart = blockStart
				return s, nil
			}
		}
	}

	// Find previous change block
	prevChangeIdx := s.findPrevChangeBlock(currentBlockIdx)
	if prevChangeIdx != -1 {
		// Jump to start of previous change block
		s.CurrentBlockIdx = prevChangeIdx
		s.ViewportStart = s.getBlockStartPosition(prevChangeIdx)
		return s, nil
	}

	// No previous change blocks in this file - navigate to previous file
	if s.FileIndex > 0 {
		return s.navigateToPrevFile(ctx)
	}
	// At first file, stay in place
	return s, nil
}

// calculateMaxViewportStart returns the maximum valid viewport start position
func (s *State) calculateMaxViewportStart(height int) int {
	if s.Diff == nil {
		return 0
	}

	// Account for header lines
	headerLines := 2 // header + separator
	availableHeight := max(height-headerLines, 1)

	totalLines := s.Diff.TotalLineCount()
	maxStart := totalLines - availableHeight
	if maxStart < 0 {
		maxStart = 0
	}
	return maxStart
}

// getBlockAtPosition returns the block index and line offset within that block
// for a given global line position
func (s *State) getBlockAtPosition(linePos int) (blockIdx int, lineInBlock int) {
	currentLine := 0
	for i, block := range s.Diff.Blocks {
		blockLineCount := block.LineCount()
		if linePos < currentLine+blockLineCount {
			return i, linePos - currentLine
		}
		currentLine += blockLineCount
	}
	// Past end, return last position
	if len(s.Diff.Blocks) > 0 {
		lastBlock := len(s.Diff.Blocks) - 1
		return lastBlock, s.Diff.Blocks[lastBlock].LineCount()
	}
	return -1, 0
}

// getBlockStartPosition returns the global line position where a block starts
func (s *State) getBlockStartPosition(blockIdx int) int {
	pos := 0
	for i := 0; i < blockIdx && i < len(s.Diff.Blocks); i++ {
		pos += s.Diff.Blocks[i].LineCount()
	}
	return pos
}

// getBlockEndPosition returns the global line position where a block ends (exclusive)
func (s *State) getBlockEndPosition(blockIdx int) int {
	return s.getBlockStartPosition(blockIdx) + s.Diff.Blocks[blockIdx].LineCount()
}

// findNextChangeBlock finds the next ChangeBlock starting from (but not including) fromBlock
// Returns -1 if no more change blocks
func (s *State) findNextChangeBlock(fromBlock int) int {
	for i := fromBlock + 1; i < len(s.Diff.Blocks); i++ {
		if _, isChange := s.Diff.Blocks[i].(diff.ChangeBlock); isChange {
			return i
		}
	}
	return -1
}

// findPrevChangeBlock finds the previous ChangeBlock before fromBlock
// Returns -1 if no previous change blocks
func (s *State) findPrevChangeBlock(fromBlock int) int {
	for i := fromBlock - 1; i >= 0; i-- {
		if _, isChange := s.Diff.Blocks[i].(diff.ChangeBlock); isChange {
			return i
		}
	}
	return -1
}

// isChangeBlockEndVisible returns true if the end of the change block is visible
func (s *State) isChangeBlockEndVisible(blockIdx, availableHeight int) bool {
	blockEnd := s.getBlockEndPosition(blockIdx)
	viewportEnd := s.ViewportStart + availableHeight
	return blockEnd <= viewportEnd
}

// isChangeBlockStartVisible returns true if the start of the change block is visible
func (s *State) isChangeBlockStartVisible(blockIdx int) bool {
	blockStart := s.getBlockStartPosition(blockIdx)
	return blockStart >= s.ViewportStart
}

// getCurrentFileLineNumber maps the current viewport position to a file line number.
// It handles all block and line types, returning a 1-indexed line number suitable for
// opening in an editor. For RemovedLine (deleted lines), it searches forward
// to find the next line with a RightLineNo, falling back to line 1 if none found.
func (s *State) getCurrentFileLineNumber() (int, error) {
	if s.Diff == nil {
		return 0, fmt.Errorf("no diff available")
	}

	if len(s.Diff.Blocks) == 0 {
		return 0, fmt.Errorf("diff has no blocks")
	}

	totalLines := s.Diff.TotalLineCount()
	if totalLines == 0 {
		return 0, fmt.Errorf("diff has no lines")
	}

	if s.ViewportStart >= totalLines {
		return 0, fmt.Errorf("viewport position out of range")
	}

	// Find the line at ViewportStart
	currentLine := 0
	for _, block := range s.Diff.Blocks {
		switch b := block.(type) {
		case diff.UnchangedBlock:
			for _, lp := range b.Lines {
				if currentLine == s.ViewportStart {
					return lp.RightLineNo, nil
				}
				currentLine++
			}
		case diff.ChangeBlock:
			for _, cl := range b.Lines {
				if currentLine == s.ViewportStart {
					switch line := cl.(type) {
					case diff.ModifiedLine:
						return line.RightLineNo, nil
					case diff.AddedLine:
						return line.RightLineNo, nil
					case diff.RemovedLine:
						// RemovedLine has no RightLineNo (deleted line doesn't exist in new file)
						// Search forward for the next line with a RightLineNo
						return s.findNextRightLineNo(currentLine + 1)
					}
				}
				currentLine++
			}
		}
	}

	return 0, fmt.Errorf("viewport position out of range")
}

// findNextRightLineNo searches forward from startLine to find a line with a RightLineNo.
// Returns 1 if no such line is found.
func (s *State) findNextRightLineNo(startLine int) (int, error) {
	currentLine := 0
	for _, block := range s.Diff.Blocks {
		switch b := block.(type) {
		case diff.UnchangedBlock:
			for _, lp := range b.Lines {
				if currentLine >= startLine {
					return lp.RightLineNo, nil
				}
				currentLine++
			}
		case diff.ChangeBlock:
			for _, cl := range b.Lines {
				if currentLine >= startLine {
					switch line := cl.(type) {
					case diff.ModifiedLine:
						return line.RightLineNo, nil
					case diff.AddedLine:
						return line.RightLineNo, nil
					case diff.RemovedLine:
						// Keep searching
					}
				}
				currentLine++
			}
		}
	}
	// No line with RightLineNo found - fall back to line 1
	return 1, nil
}

// EditorFinishedMsg is returned when the editor finishes execution
type EditorFinishedMsg struct {
	err error
}

// getEditor returns the configured editor from environment variables.
// It checks $EDITOR first, then $VISUAL. Returns an error if neither is set.
func getEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor, nil
	}

	visual := os.Getenv("VISUAL")
	if visual != "" {
		return visual, nil
	}

	return "", fmt.Errorf("no editor configured (set $EDITOR or $VISUAL)")
}

// openFileInEditor validates preconditions and launches the editor with tea.ExecProcess.
// Returns a tea.Cmd that will eventually produce an EditorFinishedMsg.
func (s *State) openFileInEditor() tea.Cmd {
	return func() tea.Msg {
		// Get editor
		editor, err := getEditor()
		if err != nil {
			return EditorFinishedMsg{err: err}
		}

		// Validate: not binary
		if s.File.IsBinary {
			return EditorFinishedMsg{err: fmt.Errorf("cannot open binary file in editor")}
		}

		// Validate: not deleted
		if s.File.Status == "D" {
			return EditorFinishedMsg{err: fmt.Errorf("cannot open: file has been deleted")}
		}

		// Get line number
		lineNo, err := s.getCurrentFileLineNumber()
		if err != nil {
			return EditorFinishedMsg{err: fmt.Errorf("failed to determine line number: %w", err)}
		}

		// Get repository root
		repoRoot, err := git.GetRepositoryRoot()
		if err != nil {
			return EditorFinishedMsg{err: fmt.Errorf("failed to determine repository root: %w", err)}
		}

		// Resolve absolute path
		absolutePath := filepath.Join(repoRoot, s.File.Path)

		// Check file exists
		if _, err := os.Stat(absolutePath); err != nil {
			if os.IsNotExist(err) {
				return EditorFinishedMsg{err: fmt.Errorf("cannot open: file not found")}
			}
			return EditorFinishedMsg{err: fmt.Errorf("failed to access file: %w", err)}
		}

		// Build command
		// Use "+lineNo | normal! zt" to position the line at the top of the screen
		// instead of centering it (vim's default behavior)
		cmd := exec.Command(editor, fmt.Sprintf("+%d | normal! zt", lineNo), absolutePath)

		// Use tea.ExecProcess to suspend TUI and run editor
		return tea.ExecProcess(cmd, func(err error) tea.Msg {
			return EditorFinishedMsg{err: err}
		})()
	}
}

// navigateToNextFile creates a command to load and navigate to the next file
func (s *State) navigateToNextFile(ctx core.Context) (*State, tea.Cmd) {
	if s.FileIndex >= len(s.Files)-1 {
		// Already at last file, stay in place
		return s, nil
	}

	nextFileIndex := s.FileIndex + 1
	nextFile := s.Files[nextFileIndex]

	return s, s.loadDiffForFile(nextFile, nextFileIndex, ctx)
}

// navigateToPrevFile creates a command to load and navigate to the previous file
func (s *State) navigateToPrevFile(ctx core.Context) (*State, tea.Cmd) {
	if s.FileIndex <= 0 {
		// Already at first file, stay in place
		return s, nil
	}

	prevFileIndex := s.FileIndex - 1
	prevFile := s.Files[prevFileIndex]

	return s, s.loadDiffForFile(prevFile, prevFileIndex, ctx)
}

// loadDiffForFile creates a command to fetch and parse the diff for a file
func (s *State) loadDiffForFile(file core.FileChange, fileIndex int, ctx core.Context) tea.Cmd {
	return func() tea.Msg {
		// Fetch full file content and diff based on DiffSource type
		fullDiffResult, err := fetchFileDiffForSource(s.Source, file, ctx.FetchFullFileDiff())
		if err != nil {
			return core.DiffLoadedMsg{
				Source:    s.Source,
				Files:     s.Files,
				FileIndex: fileIndex,
				File:      file,
				Err:       err,
			}
		}

		// Build the block-based diff
		fileDiff, err := diff.BuildFileDiff(
			file.Path,
			fullDiffResult.OldContent,
			fullDiffResult.NewContent,
			fullDiffResult.DiffOutput,
		)
		if err != nil {
			return core.DiffLoadedMsg{
				Source:    s.Source,
				Files:     s.Files,
				FileIndex: fileIndex,
				File:      file,
				Err:       err,
			}
		}

		return core.DiffLoadedMsg{
			Source:    s.Source,
			Files:     s.Files,
			FileIndex: fileIndex,
			File:      file,
			Diff:      fileDiff,
		}
	}
}

// fetchFileDiffForSource fetches the full file diff based on the DiffSource type.
// Uses type switch to call the appropriate git function for each source type.
func fetchFileDiffForSource(source core.DiffSource, file core.FileChange, fetchFullFileDiff core.FetchFullFileDiffFunc) (*core.FullFileDiffResult, error) {
	switch src := source.(type) {
	case core.CommitRangeDiffSource:
		// For commit ranges, use the injected fetchFullFileDiff function
		commitRange := src.ToCommitRange()
		return fetchFullFileDiff(commitRange, file)

	case core.UncommittedChangesDiffSource:
		// For uncommitted changes, type switch on Type field to call appropriate git function
		switch src.Type {
		case core.UncommittedTypeUnstaged:
			return git.FetchUnstagedFileDiff(file)
		case core.UncommittedTypeStaged:
			return git.FetchStagedFileDiff(file)
		case core.UncommittedTypeAll:
			return git.FetchAllUncommittedFileDiff(file)
		default:
			return nil, fmt.Errorf("unknown uncommitted type: %v", src.Type)
		}

	default:
		return nil, fmt.Errorf("unknown diff source type: %T", source)
	}
}

// positionAtFirstChange positions the viewport at the first change block
func (s *State) positionAtFirstChange() {
	s.ViewportStart = 0
	s.CurrentBlockIdx = -1

	if s.Diff == nil {
		return
	}

	lineOffset := 0
	for i, block := range s.Diff.Blocks {
		if _, isChange := block.(diff.ChangeBlock); isChange {
			s.ViewportStart = lineOffset
			s.CurrentBlockIdx = i
			return
		}
		lineOffset += block.LineCount()
	}
}
