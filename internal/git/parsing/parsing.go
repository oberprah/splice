package parsing

import (
	"fmt"
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
