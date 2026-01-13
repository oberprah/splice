package cli

import (
	"fmt"
	"strings"

	"github.com/oberprah/splice/internal/core"
)

// ParseCommand parses command line arguments and returns the command and remaining args.
// If no arguments: returns "log", []
// If first arg is "diff": returns "diff", remaining args
// Otherwise: returns "log", [] (unknown commands default to log)
func ParseCommand(args []string) (cmd string, remainingArgs []string) {
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

// DiffArgs represents parsed diff command arguments.
// Either RawSpec is set (for commit ranges) or UncommittedType is set (for uncommitted changes).
type DiffArgs struct {
	RawSpec         string                // Non-empty for commit range specs like "main..feature"
	UncommittedType *core.UncommittedType // Non-nil for uncommitted changes (--staged, HEAD, or default)
}

// IsCommitRange returns true if this represents a commit range diff.
func (d DiffArgs) IsCommitRange() bool {
	return d.RawSpec != ""
}

// ParseDiffArgs parses diff specification arguments.
// This is pure parsing - no git validation is performed.
//
// Examples:
//   - [] → UncommittedTypeUnstaged (working tree vs index)
//   - ["--staged"] or ["--cached"] → UncommittedTypeStaged (index vs HEAD)
//   - ["HEAD"] → UncommittedTypeAll (working tree vs HEAD)
//   - ["main..feature"] → RawSpec="main..feature"
func ParseDiffArgs(args []string) (DiffArgs, error) {
	if len(args) == 0 {
		// No args = unstaged changes
		t := core.UncommittedTypeUnstaged
		return DiffArgs{UncommittedType: &t}, nil
	}

	if len(args) > 1 {
		return DiffArgs{}, fmt.Errorf("unexpected arguments: %v", args[1:])
	}

	firstArg := args[0]

	// Check for special flags/keywords
	switch firstArg {
	case "--staged", "--cached":
		t := core.UncommittedTypeStaged
		return DiffArgs{UncommittedType: &t}, nil
	case "HEAD":
		t := core.UncommittedTypeAll
		return DiffArgs{UncommittedType: &t}, nil
	default:
		// Must be a commit spec - validate basic syntax
		if !IsValidDiffSpec(firstArg) {
			return DiffArgs{}, fmt.Errorf("invalid diff spec: %q", firstArg)
		}
		return DiffArgs{RawSpec: firstArg}, nil
	}
}

// IsValidDiffSpec performs basic syntax validation on diff specs.
// Rejects obvious malformed input (spaces, shell metacharacters).
// Actual ref validation happens later via git commands.
func IsValidDiffSpec(spec string) bool {
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
