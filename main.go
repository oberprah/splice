package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/oberprah/splice/internal/app"
	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/ui/states/directdiff"
	"github.com/oberprah/splice/internal/ui/states/loading"
)

// parseArgs parses command line arguments and returns the command and remaining args.
// If no arguments: returns "log", []
// If first arg is "diff": returns "diff", remaining args
// Otherwise: returns "log", [] (unknown commands default to log)
func parseArgs(args []string) (cmd string, remainingArgs []string) {
	if len(args) < 2 {
		return "log", []string{}
	}

	firstArg := args[1]
	if firstArg == "diff" {
		return "diff", args[2:]
	}

	// Unknown command defaults to log
	return "log", []string{}
}

// parseDiffSpec parses diff specification arguments.
// Returns either a raw spec string (for commit ranges) or an UncommittedType (for uncommitted changes).
// For commit ranges, returns the spec as-is and nil uncommittedType.
// For uncommitted changes, returns empty string and non-nil uncommittedType.
func parseDiffSpec(args []string) (rawSpec string, uncommittedType *core.UncommittedType, err error) {
	if len(args) == 0 {
		// No args = unstaged changes
		t := core.UncommittedTypeUnstaged
		return "", &t, nil
	}

	if len(args) > 1 {
		return "", nil, fmt.Errorf("unexpected arguments: %v", args[1:])
	}

	firstArg := args[0]

	// Check for special flags/keywords
	switch firstArg {
	case "--staged", "--cached":
		t := core.UncommittedTypeStaged
		return "", &t, nil
	case "HEAD":
		t := core.UncommittedTypeAll
		return "", &t, nil
	default:
		// Must be a commit spec - validate basic syntax
		if !isValidDiffSpec(firstArg) {
			return "", nil, fmt.Errorf("invalid diff spec: %q", firstArg)
		}
		return firstArg, nil, nil
	}
}

// isValidDiffSpec performs basic syntax validation on diff specs.
// Rejects obvious malformed input (spaces, shell metacharacters).
// Actual ref validation happens later via git commands.
func isValidDiffSpec(spec string) bool {
	// Disallow spaces and shell metacharacters
	return !strings.Contains(spec, " ") &&
		!strings.Contains(spec, ";") &&
		!strings.Contains(spec, "|") &&
		!strings.Contains(spec, "&") &&
		!strings.Contains(spec, ">") &&
		!strings.Contains(spec, "<") &&
		!strings.Contains(spec, "$") &&
		!strings.Contains(spec, "`")
}

// validateDiffSpec validates that a diff specification has changes.
// Uses git diff --quiet to check:
// - Exit 0 = no changes (error)
// - Exit 1 = has changes (success)
// - Exit 128+ = invalid spec (error)
func validateDiffSpec(rawSpec string, uncommittedType *core.UncommittedType) error {
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
		if uncommittedType != nil {
			return fmt.Errorf("error checking uncommitted changes: %v", err)
		}
		return fmt.Errorf("invalid diff specification %q: %v", rawSpec, err)
	}

	// Other errors
	return fmt.Errorf("error running git diff: %v", err)
}

// parseCommitRange parses a commit range spec (like "main..feature" or "HEAD~5..HEAD")
// into a CommitRangeDiffSource with actual GitCommit objects.
func parseCommitRange(spec string) (core.CommitRangeDiffSource, error) {
	// Parse the range specification to get start and end refs
	startRef := spec
	endRef := "HEAD"

	// Check if it's a range (contains ..)
	if strings.Contains(spec, "...") {
		parts := strings.SplitN(spec, "...", 2)
		startRef = parts[0]
		endRef = parts[1]
		// Three-dot range: find merge base
		mergeBaseCmd := exec.Command("git", "merge-base", startRef, endRef)
		mergeBaseOutput, err := mergeBaseCmd.Output()
		if err != nil {
			return core.CommitRangeDiffSource{}, fmt.Errorf("error finding merge base: %v", err)
		}
		startRef = strings.TrimSpace(string(mergeBaseOutput))
	} else if strings.Contains(spec, "..") {
		parts := strings.SplitN(spec, "..", 2)
		startRef = parts[0]
		endRef = parts[1]
	}

	// Resolve refs to commits
	startCommit, err := resolveCommit(startRef)
	if err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error resolving start ref %q: %v", startRef, err)
	}

	endCommit, err := resolveCommit(endRef)
	if err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error resolving end ref %q: %v", endRef, err)
	}

	// Count commits in range
	countCmd := exec.Command("git", "rev-list", "--count", startRef+".."+endRef)
	countOutput, err := countCmd.Output()
	if err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error counting commits: %v", err)
	}
	count := 0
	if _, err := fmt.Sscanf(string(countOutput), "%d", &count); err != nil {
		return core.CommitRangeDiffSource{}, fmt.Errorf("error parsing commit count: %v", err)
	}

	return core.CommitRangeDiffSource{
		Start: startCommit,
		End:   endCommit,
		Count: count,
	}, nil
}

// resolveCommit resolves a git ref (like "HEAD", "main", "abc123") to a GitCommit.
func resolveCommit(ref string) (core.GitCommit, error) {
	// Use git log to get commit info
	cmd := exec.Command("git", "log", "-1", "--format=%H%n%s%n%an%n%aI%n%P", ref)
	output, err := cmd.Output()
	if err != nil {
		return core.GitCommit{}, fmt.Errorf("error resolving ref: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
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

	commit := core.GitCommit{
		Hash:         lines[0],
		Message:      lines[1],
		Author:       lines[2],
		Date:         date,
		ParentHashes: parents,
		Refs:         []core.RefInfo{}, // We don't need refs for CLI parsing
	}

	return commit, nil
}

func main() {
	// Parse command line arguments
	cmd, args := parseArgs(os.Args)

	var initialState core.State

	if cmd == "diff" {
		// Parse and validate diff specification
		rawSpec, uncommittedType, err := parseDiffSpec(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Validate that the diff spec has changes
		if err := validateDiffSpec(rawSpec, uncommittedType); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Create DiffSource from parsed spec
		var diffSource core.DiffSource
		if uncommittedType != nil {
			// Uncommitted changes
			diffSource = core.UncommittedChangesDiffSource{Type: *uncommittedType}
		} else {
			// Commit range - parse into CommitRangeDiffSource
			commitRange, err := parseCommitRange(rawSpec)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			diffSource = commitRange
		}

		// Create DirectDiffLoadingState
		initialState = directdiff.New(diffSource)
	} else {
		// Default to log view
		initialState = loading.State{}
	}

	// Create the initial model with the appropriate state
	initialModel := app.NewModel(
		app.WithInitialState(initialState),
	)

	// Start the Bubbletea program with alternate screen (fullscreen mode)
	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
