package operations

import (
	"fmt"
	"strings"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// FetchCommits executes git log and returns a slice of commits
func FetchCommits(limit int) ([]core.GitCommit, error) {
	return commands.FetchLog(limit)
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
		mergeBase, err := commands.FindMergeBase(startRef, endRef)
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
	count, err := commands.CountCommitsInRange(startRef, endRef)
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
	return commands.ResolveRefToCommit(ref)
}
