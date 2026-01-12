package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"time"

	"github.com/oberprah/splice/internal/core"
)

// parseDate parses a date string in RFC3339 format
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse(time.RFC3339, dateStr)
}

// parseRefDecorations parses git's %d output into structured RefInfo data.
// Git's %d format examples:
//
//	" (HEAD -> main)" - HEAD pointing to branch
//	" (HEAD -> main, origin/main)" - HEAD and remote
//	" (tag: v1.0)" - tag
//	" (main)" - local branch
//	" (origin/main)" - remote branch
//	"" - no refs
func parseRefDecorations(refsStr string) []core.RefInfo {
	// Trim the leading space and surrounding parentheses
	refsStr = strings.TrimSpace(refsStr)
	if refsStr == "" {
		return []core.RefInfo{}
	}

	// Remove the surrounding parentheses
	if strings.HasPrefix(refsStr, "(") && strings.HasSuffix(refsStr, ")") {
		refsStr = refsStr[1 : len(refsStr)-1]
	} else {
		// No parentheses means no refs
		return []core.RefInfo{}
	}

	// Split by comma to get individual refs
	refParts := strings.Split(refsStr, ",")
	refs := make([]core.RefInfo, 0, len(refParts))

	headBranch := "" // Track which branch HEAD points to

	// First pass: find HEAD pointer
	for _, part := range refParts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "HEAD -> ") {
			headBranch = strings.TrimPrefix(part, "HEAD -> ")
			break
		}
	}

	// Second pass: parse all refs
	for _, part := range refParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		var ref core.RefInfo

		// Handle "HEAD -> branch" specially
		if strings.HasPrefix(part, "HEAD -> ") {
			branchName := strings.TrimPrefix(part, "HEAD -> ")
			ref = core.RefInfo{
				Name:   branchName,
				Type:   core.RefTypeBranch,
				IsHead: true,
			}
		} else if strings.HasPrefix(part, "tag: ") {
			// Handle tags
			tagName := strings.TrimPrefix(part, "tag: ")
			ref = core.RefInfo{
				Name:   tagName,
				Type:   core.RefTypeTag,
				IsHead: false,
			}
		} else if strings.Contains(part, "/") {
			// Remote branch (contains /)
			ref = core.RefInfo{
				Name:   part,
				Type:   core.RefTypeRemoteBranch,
				IsHead: false,
			}
		} else {
			// Local branch
			ref = core.RefInfo{
				Name:   part,
				Type:   core.RefTypeBranch,
				IsHead: part == headBranch,
			}
		}

		refs = append(refs, ref)
	}

	return refs
}

// ParseGitLogOutput parses git log output into GitCommit structs.
// Input format: "hash\0parents\0refs\0author\0date\0subject\0body\x1e" (NULL-separated fields, record separator between commits).
func ParseGitLogOutput(output string) ([]core.GitCommit, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []core.GitCommit{}, nil
	}

	// Split by record separator to get individual commits
	commitRecords := strings.Split(output, "\x1e")

	commits := make([]core.GitCommit, 0, len(commitRecords))

	for _, record := range commitRecords {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		// Split each commit by single NULL to get fields
		fields := strings.SplitN(record, "\x00", 7)
		if len(fields) != 7 {
			continue // Skip malformed records
		}

		hash := fields[0]
		parentsStr := fields[1]
		refsStr := fields[2]
		author := fields[3]
		dateStr := fields[4]
		message := fields[5]
		body := strings.TrimSpace(fields[6])

		// Skip empty commits
		if hash == "" {
			continue
		}

		// Parse parent hashes (space-separated, empty string for root commits)
		var parentHashes []string
		if parentsStr != "" {
			parentHashes = strings.Split(parentsStr, " ")
		} else {
			parentHashes = []string{}
		}

		// Parse ref decorations
		refs := parseRefDecorations(refsStr)

		// Parse the date
		date, err := parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date %q: %w", dateStr, err)
		}

		commit := core.GitCommit{
			Hash:         hash,
			ParentHashes: parentHashes,
			Refs:         refs,
			Message:      message,
			Body:         body,
			Author:       author,
			Date:         date,
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// FetchCommits executes git log and returns a slice of commits
func FetchCommits(limit int) ([]core.GitCommit, error) {
	// Use git log with custom format using NULL separator: hash\0parents\0refs\0author\0date\0subject\0body
	// NULL character is used as field delimiter since it won't appear in commit messages
	// ASCII Record Separator (0x1e) is used as commit record separator
	// %P outputs parent hashes (space-separated for merges, empty for root commits)
	// %d outputs ref decorations (e.g., " (HEAD -> main, tag: v1.0)")
	cmd := exec.Command("git", "log",
		"--pretty=format:%H%x00%P%x00%d%x00%an%x00%ad%x00%s%x00%b%x1e",
		"--date=iso-strict",
		fmt.Sprintf("-n %d", limit))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if this is not a git repository
		if strings.Contains(stderr.String(), "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git log failed: %v - %s", err, stderr.String())
	}

	return ParseGitLogOutput(out.String())
}

// ParseFileChangesOutput parses git diff output into FileChange structs.
// Input format: "additions\tdeletions\tfilepath" (one file per line).
func ParseFileChangesOutput(output string) ([]core.FileChange, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	changes := make([]core.FileChange, 0, len(lines))

	for i, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("malformed line %d: expected 3 tab-separated fields, got %d: %q", i+1, len(parts), line)
		}

		additionsStr := parts[0]
		deletionsStr := parts[1]
		path := parts[2]

		var additions, deletions int
		var isBinary bool

		// Check if this is a binary file (git shows "-" for both additions and deletions)
		if additionsStr == "-" && deletionsStr == "-" {
			isBinary = true
			additions = 0
			deletions = 0
		} else {
			// Parse additions
			var err error
			_, err = fmt.Sscanf(additionsStr, "%d", &additions)
			if err != nil {
				return nil, fmt.Errorf("invalid additions count on line %d: %q", i+1, additionsStr)
			}

			// Parse deletions
			_, err = fmt.Sscanf(deletionsStr, "%d", &deletions)
			if err != nil {
				return nil, fmt.Errorf("invalid deletions count on line %d: %q", i+1, deletionsStr)
			}
		}

		change := core.FileChange{
			Path:      path,
			Additions: additions,
			Deletions: deletions,
			IsBinary:  isBinary,
		}

		changes = append(changes, change)
	}

	return changes, nil
}

// FetchFileChanges executes git diff and returns a slice of file changes for a commit range.
func FetchFileChanges(commitRange core.CommitRange) ([]core.FileChange, error) {
	// Determine the from and to hashes based on whether this is a single commit or range
	var fromHash, toHash string
	if commitRange.IsSingleCommit() {
		// Single commit: compare commit with its parent
		fromHash = commitRange.End.Hash + "^"
		toHash = commitRange.End.Hash
	} else {
		// Range: compare Start commit with End commit
		fromHash = commitRange.Start.Hash
		toHash = commitRange.End.Hash
	}

	rangeSpec := fromHash + ".." + toHash

	// First, get file statuses (A/M/D/R)
	// Note: diff-tree doesn't work well with ranges, so we use git diff for status
	statusCmd := exec.Command("git", "diff", "--name-status", rangeSpec)
	var statusOut bytes.Buffer
	var statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr

	err := statusCmd.Run()
	if err != nil {
		stderrStr := statusErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return nil, fmt.Errorf("invalid commit range: %s..%s", fromHash, toHash)
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse status information into a map
	statusMap := make(map[string]string)
	statusLines := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			status := parts[0]
			path := parts[1]
			statusMap[path] = status
		}
	}

	// Now get numstat information
	cmd := exec.Command("git", "diff", "--numstat", rangeSpec)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return nil, fmt.Errorf("invalid commit range: %s..%s", fromHash, toHash)
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse file changes and add status
	changes, err := ParseFileChangesOutput(out.String())
	if err != nil {
		return nil, err
	}

	// Add status to each change
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}

	return changes, nil
}

// FetchFileContent retrieves the content of a file at a specific commit.
// Returns empty string without error if the file doesn't exist at that commit.
func FetchFileContent(commitHash, filePath string) (string, error) {
	cmd := exec.Command("git", "show", commitHash+":"+filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		// Check if this is a "file not found" error - return empty string, no error
		if strings.Contains(stderrStr, "does not exist") ||
			strings.Contains(stderrStr, "exists on disk, but not in") ||
			strings.Contains(stderrStr, "fatal: path") {
			return "", nil
		}
		// Check for invalid commit
		if strings.Contains(stderrStr, "unknown revision") ||
			strings.Contains(stderrStr, "bad revision") ||
			strings.Contains(stderrStr, "not a valid object name") {
			return "", fmt.Errorf("invalid commit: %s", commitHash)
		}
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		return "", fmt.Errorf("git show failed: %v - %s", err, stderrStr)
	}

	return out.String(), nil
}

// FetchFullFileDiff fetches the complete file content before and after a change,
// along with the diff output. This enables showing the full file with changes highlighted.
func FetchFullFileDiff(commitRange core.CommitRange, change core.FileChange) (*core.FullFileDiffResult, error) {
	// Determine the from and to hashes based on whether this is a single commit or range
	var fromHash, toHash string
	if commitRange.IsSingleCommit() {
		// Single commit: compare commit with its parent
		fromHash = commitRange.End.Hash + "^"
		toHash = commitRange.End.Hash
	} else {
		// Range: compare Start commit with End commit
		fromHash = commitRange.Start.Hash
		toHash = commitRange.End.Hash
	}

	result := &core.FullFileDiffResult{
		NewPath: change.Path,
		OldPath: change.Path,
	}

	// TODO: Handle renames properly (status starts with "R")
	// For renames, the path contains "old -> new" format, but we get OldPath from git
	// In our FileChange struct, we only have Path (the new path)
	// We need to handle this differently - for now assume same path

	// Fetch new content (at toHash)
	switch change.Status {
	case "D": // Deleted file - no new content
		result.NewContent = ""
	default:
		newContent, err := FetchFileContent(toHash, change.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch old content (at fromHash)
	switch change.Status {
	case "A": // Added file - no old content
		result.OldContent = ""
	default:
		oldContent, err := FetchFileContent(fromHash, result.OldPath)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch the diff
	rangeSpec := fromHash + ".." + toHash
	diffOutput, err := FetchFileDiffRange(rangeSpec, change.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch diff: %w", err)
	}
	result.DiffOutput = diffOutput

	return result, nil
}

// FetchFileDiff retrieves the unified diff for a specific file in a commit.
// The filePath should be relative to the repository root.
func FetchFileDiff(commitHash, filePath string) (string, error) {
	// Use :(top) pathspec to ensure path is relative to repo root regardless of cwd
	cmd := exec.Command("git", "show", commitHash, "--format=", "--", ":(top)"+filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return "", fmt.Errorf("invalid commit: %s", commitHash)
		}
		return "", fmt.Errorf("git show failed: %v - %s", err, stderrStr)
	}

	return out.String(), nil
}

// FetchFileDiffRange retrieves the unified diff for a specific file in a commit range.
// The rangeSpec should be in the format "fromHash..toHash".
// The filePath should be relative to the repository root.
func FetchFileDiffRange(rangeSpec, filePath string) (string, error) {
	// Use :(top) pathspec to ensure path is relative to repo root regardless of cwd
	cmd := exec.Command("git", "diff", rangeSpec, "--", ":(top)"+filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return "", fmt.Errorf("invalid commit range: %s", rangeSpec)
		}
		return "", fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	return out.String(), nil
}

// FetchUnstagedFileChanges executes git diff and returns a slice of file changes
// for unstaged changes (working tree vs index).
func FetchUnstagedFileChanges() ([]core.FileChange, error) {
	// Get file statuses (A/M/D/R)
	statusCmd := exec.Command("git", "diff", "--name-status")
	var statusOut bytes.Buffer
	var statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr

	err := statusCmd.Run()
	if err != nil {
		stderrStr := statusErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse status information into a map
	statusMap := make(map[string]string)
	statusLines := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			status := parts[0]
			path := parts[1]
			statusMap[path] = status
		}
	}

	// Get numstat information
	cmd := exec.Command("git", "diff", "--numstat")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse file changes and add status
	changes, err := ParseFileChangesOutput(out.String())
	if err != nil {
		return nil, err
	}

	// Add status to each change
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}

	return changes, nil
}

// FetchStagedFileChanges executes git diff --staged and returns a slice of file changes
// for staged changes (index vs HEAD).
func FetchStagedFileChanges() ([]core.FileChange, error) {
	// Get file statuses (A/M/D/R)
	statusCmd := exec.Command("git", "diff", "--staged", "--name-status")
	var statusOut bytes.Buffer
	var statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr

	err := statusCmd.Run()
	if err != nil {
		stderrStr := statusErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse status information into a map
	statusMap := make(map[string]string)
	statusLines := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			status := parts[0]
			path := parts[1]
			statusMap[path] = status
		}
	}

	// Get numstat information
	cmd := exec.Command("git", "diff", "--staged", "--numstat")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse file changes and add status
	changes, err := ParseFileChangesOutput(out.String())
	if err != nil {
		return nil, err
	}

	// Add status to each change
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}

	return changes, nil
}

// FetchAllUncommittedFileChanges executes git diff HEAD and returns a slice of file changes
// for all uncommitted changes (working tree vs HEAD).
func FetchAllUncommittedFileChanges() ([]core.FileChange, error) {
	// Get file statuses (A/M/D/R)
	statusCmd := exec.Command("git", "diff", "HEAD", "--name-status")
	var statusOut bytes.Buffer
	var statusErr bytes.Buffer
	statusCmd.Stdout = &statusOut
	statusCmd.Stderr = &statusErr

	err := statusCmd.Run()
	if err != nil {
		stderrStr := statusErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse status information into a map
	statusMap := make(map[string]string)
	statusLines := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	for _, line := range statusLines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 {
			status := parts[0]
			path := parts[1]
			statusMap[path] = status
		}
	}

	// Get numstat information
	cmd := exec.Command("git", "diff", "HEAD", "--numstat")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}

	// Parse file changes and add status
	changes, err := ParseFileChangesOutput(out.String())
	if err != nil {
		return nil, err
	}

	// Add status to each change
	for i := range changes {
		if status, ok := statusMap[changes[i].Path]; ok {
			changes[i].Status = status
		} else {
			changes[i].Status = "M" // Default to modified if not found
		}
	}

	return changes, nil
}

// FetchIndexFileContent retrieves the content of a file from the index (staging area).
// Returns empty string without error if the file doesn't exist in the index.
func FetchIndexFileContent(filePath string) (string, error) {
	cmd := exec.Command("git", "show", ":"+filePath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		// Check if this is a "file not found" error - return empty string, no error
		if strings.Contains(stderrStr, "does not exist") ||
			strings.Contains(stderrStr, "exists on disk, but not in") ||
			strings.Contains(stderrStr, "fatal: path") {
			return "", nil
		}
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		return "", fmt.Errorf("git show failed: %v - %s", err, stderrStr)
	}

	return out.String(), nil
}

// FetchWorkingTreeFileContent retrieves the content of a file from the working tree.
// Returns empty string without error if the file doesn't exist.
func FetchWorkingTreeFileContent(filePath string) (string, error) {
	content, err := exec.Command("cat", filePath).Output()
	if err != nil {
		// File doesn't exist or can't be read - return empty string, no error
		return "", nil
	}
	return string(content), nil
}

// FetchUnstagedFileDiff fetches the complete file content before and after an unstaged change,
// along with the diff output. This enables showing the full file with changes highlighted.
// Unstaged changes compare the index (old) with the working tree (new).
func FetchUnstagedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from index)
	switch file.Status {
	case "A": // Added file - no old content in index
		result.OldContent = ""
	default:
		oldContent, err := FetchIndexFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from working tree)
	switch file.Status {
	case "D": // Deleted file - no new content in working tree
		result.NewContent = ""
	default:
		newContent, err := FetchWorkingTreeFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffCmd := exec.Command("git", "diff", "--", file.Path)
	var diffOut bytes.Buffer
	var diffErr bytes.Buffer
	diffCmd.Stdout = &diffOut
	diffCmd.Stderr = &diffErr

	err := diffCmd.Run()
	if err != nil {
		stderrStr := diffErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}
	result.DiffOutput = diffOut.String()

	return result, nil
}

// FetchStagedFileDiff fetches the complete file content before and after a staged change,
// along with the diff output. This enables showing the full file with changes highlighted.
// Staged changes compare HEAD (old) with the index (new).
func FetchStagedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from HEAD)
	switch file.Status {
	case "A": // Added file - no old content in HEAD
		result.OldContent = ""
	default:
		oldContent, err := FetchFileContent("HEAD", file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from index)
	switch file.Status {
	case "D": // Deleted file - no new content in index
		result.NewContent = ""
	default:
		newContent, err := FetchIndexFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffCmd := exec.Command("git", "diff", "--staged", "--", file.Path)
	var diffOut bytes.Buffer
	var diffErr bytes.Buffer
	diffCmd.Stdout = &diffOut
	diffCmd.Stderr = &diffErr

	err := diffCmd.Run()
	if err != nil {
		stderrStr := diffErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}
	result.DiffOutput = diffOut.String()

	return result, nil
}

// FetchAllUncommittedFileDiff fetches the complete file content before and after all uncommitted changes,
// along with the diff output. This enables showing the full file with changes highlighted.
// All uncommitted changes compare HEAD (old) with the working tree (new).
func FetchAllUncommittedFileDiff(file core.FileChange) (*core.FullFileDiffResult, error) {
	result := &core.FullFileDiffResult{
		NewPath: file.Path,
		OldPath: file.Path,
	}

	// Fetch old content (from HEAD)
	switch file.Status {
	case "A": // Added file - no old content in HEAD
		result.OldContent = ""
	default:
		oldContent, err := FetchFileContent("HEAD", file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch old content: %w", err)
		}
		result.OldContent = oldContent
	}

	// Fetch new content (from working tree)
	switch file.Status {
	case "D": // Deleted file - no new content in working tree
		result.NewContent = ""
	default:
		newContent, err := FetchWorkingTreeFileContent(file.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch new content: %w", err)
		}
		result.NewContent = newContent
	}

	// Fetch the unified diff
	diffCmd := exec.Command("git", "diff", "HEAD", "--", file.Path)
	var diffOut bytes.Buffer
	var diffErr bytes.Buffer
	diffCmd.Stdout = &diffOut
	diffCmd.Stderr = &diffErr

	err := diffCmd.Run()
	if err != nil {
		stderrStr := diffErr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return nil, fmt.Errorf("not a git repository")
		}
		return nil, fmt.Errorf("git diff failed: %v - %s", err, stderrStr)
	}
	result.DiffOutput = diffOut.String()

	return result, nil
}

// ValidateDiffHasChanges checks if a diff specification has any changes.
// For uncommitted changes, checks the appropriate git diff.
// For commit ranges, checks if the range has any diff.
// Returns nil if there are changes, error if no changes or invalid spec.
func ValidateDiffHasChanges(rawSpec string, uncommittedType *core.UncommittedType) error {
	var args []string

	if uncommittedType != nil {
		// Uncommitted changes
		switch *uncommittedType {
		case core.UncommittedTypeUnstaged:
			args = []string{"diff", "--quiet"}
		case core.UncommittedTypeStaged:
			args = []string{"diff", "--quiet", "--staged"}
		case core.UncommittedTypeAll:
			args = []string{"diff", "--quiet", "HEAD"}
		default:
			return fmt.Errorf("unknown uncommitted type: %v", *uncommittedType)
		}
	} else {
		// Commit range
		args = []string{"diff", "--quiet", rawSpec}
	}

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err == nil {
		// Exit 0 = no changes
		if uncommittedType != nil {
			return fmt.Errorf("no uncommitted changes found")
		}
		return fmt.Errorf("no changes found in %q", rawSpec)
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		if exitCode == 1 {
			// Exit 1 = has changes (this is what we want)
			return nil
		}
		// Exit 128+ = git error (invalid ref, etc.)
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return fmt.Errorf("not a git repository")
		}
		if uncommittedType != nil {
			return fmt.Errorf("error checking uncommitted changes: %v", err)
		}
		return fmt.Errorf("invalid diff specification %q: %v", rawSpec, err)
	}

	// Other errors
	return fmt.Errorf("error running git diff: %v", err)
}

// ResolveCommitRange parses a commit range spec (like "main..feature" or "HEAD~5")
// and resolves refs to GitCommit objects.
func ResolveCommitRange(spec string) (core.CommitRangeDiffSource, error) {
	// Parse the range specification to get start and end refs
	startRef := spec
	endRef := "HEAD"

	// Check if it's a range (contains ..)
	if strings.Contains(spec, "...") {
		parts := strings.SplitN(spec, "...", 2)
		startRef = parts[0]
		endRef = parts[1]
		// Three-dot range: find merge base
		mergeBase, err := findMergeBase(startRef, endRef)
		if err != nil {
			return core.CommitRangeDiffSource{}, err
		}
		startRef = mergeBase
	} else if strings.Contains(spec, "..") {
		parts := strings.SplitN(spec, "..", 2)
		startRef = parts[0]
		endRef = parts[1]
	}

	// Resolve refs to commits
	startCommit, err := ResolveRef(startRef)
	if err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error resolving start ref %q: %v", startRef, err)
	}

	endCommit, err := ResolveRef(endRef)
	if err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error resolving end ref %q: %v", endRef, err)
	}

	// Count commits in range
	count, err := countCommitsInRange(startRef, endRef)
	if err != nil {
		return core.CommitRangeDiffSource{}, err
	}

	return core.CommitRangeDiffSource{
		Start: startCommit,
		End:   endCommit,
		Count: count,
	}, nil
}

// ResolveRef resolves a git ref (like "HEAD", "main", "abc123") to a GitCommit.
func ResolveRef(ref string) (core.GitCommit, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H%n%s%n%an%n%aI%n%P", ref)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return core.GitCommit{}, fmt.Errorf("not a git repository")
		}
		if strings.Contains(stderrStr, "unknown revision") || strings.Contains(stderrStr, "bad revision") {
			return core.GitCommit{}, fmt.Errorf("unknown revision: %s", ref)
		}
		return core.GitCommit{}, fmt.Errorf("error resolving ref: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 4 {
		return core.GitCommit{}, fmt.Errorf("unexpected git log output")
	}

	// Parse date
	date, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[3]))
	if err != nil {
		return core.GitCommit{}, fmt.Errorf("error parsing date: %v", err)
	}

	// Parse parent hashes
	var parents []string
	if len(lines) >= 5 && lines[4] != "" {
		parents = strings.Fields(lines[4])
	}

	return core.GitCommit{
		Hash:         lines[0],
		Message:      lines[1],
		Author:       lines[2],
		Date:         date,
		ParentHashes: parents,
		Refs:         []core.RefInfo{},
	}, nil
}

// findMergeBase finds the merge base of two refs.
func findMergeBase(ref1, ref2 string) (string, error) {
	cmd := exec.Command("git", "merge-base", ref1, ref2)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		return "", fmt.Errorf("error finding merge base: %v", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// countCommitsInRange counts commits between two refs.
func countCommitsInRange(startRef, endRef string) (int, error) {
	cmd := exec.Command("git", "rev-list", "--count", startRef+".."+endRef)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return 0, fmt.Errorf("not a git repository")
		}
		return 0, fmt.Errorf("error counting commits: %v", err)
	}

	var count int
	if _, err := fmt.Sscanf(out.String(), "%d", &count); err != nil {
		return 0, fmt.Errorf("error parsing commit count: %v", err)
	}
	return count, nil
}

// GetRepositoryRoot executes git rev-parse --show-toplevel to get the absolute path of the repository root.
func GetRepositoryRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "not a git repository") {
			return "", fmt.Errorf("not a git repository")
		}
		return "", fmt.Errorf("git rev-parse failed: %v - %s", err, stderrStr)
	}

	return strings.TrimSpace(out.String()), nil
}

