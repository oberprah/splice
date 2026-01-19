package operations

import (
	"fmt"

	"github.com/oberprah/splice/internal/core"
	"github.com/oberprah/splice/internal/git/commands"
)

// GetRepositoryRoot executes git rev-parse --show-toplevel to get the absolute path of the repository root.
func GetRepositoryRoot() (string, error) {
	return commands.GetRepositoryRoot()
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
			args = []string{"diff"}
		case core.UncommittedTypeStaged:
			args = []string{"diff", "--staged"}
		case core.UncommittedTypeAll:
			args = []string{"diff", "HEAD"}
		default:
			return fmt.Errorf("unknown uncommitted type: %v", *uncommittedType)
		}
	} else {
		// Commit range
		args = []string{"diff", rawSpec}
	}

	err := commands.ValidateDiffHasChanges(args...)
	if err != nil {
		// Command returns nil if changes exist, error if no changes
		// We want to return error if no changes
		if uncommittedType != nil {
			return fmt.Errorf("no uncommitted changes found")
		}
		return fmt.Errorf("no changes found in %q", rawSpec)
	}

	return nil
}
