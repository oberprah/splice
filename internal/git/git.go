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
		// Range: compare Start commit's parent with End commit
		fromHash = commitRange.Start.Hash + "^"
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
		// Range: compare Start commit's parent with End commit
		fromHash = commitRange.Start.Hash + "^"
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
