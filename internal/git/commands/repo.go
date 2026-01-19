package commands

import (
	"os/exec"
	"strings"
)

// GetRepositoryRoot executes git rev-parse --show-toplevel to get the absolute path of the repository root.
func GetRepositoryRoot() (string, error) {
	stdout, stderr, err := execGit("rev-parse", "--show-toplevel")
	if err != nil {
		return "", checkGitError(stderr, err, "git rev-parse")
	}

	return strings.TrimSpace(stdout), nil
}

// ValidateDiffHasChanges checks if a diff specification has any changes.
// For uncommitted changes, uses the provided git diff args.
// For commit ranges, checks if the range has any diff.
// Returns nil if there are changes, error if no changes or invalid spec.
func ValidateDiffHasChanges(args ...string) error {
	// Add --quiet flag to suppress output
	quietArgs := append([]string{}, args...)
	quietArgs = append(quietArgs, "--quiet")

	cmd := exec.Command("git", quietArgs...)
	err := cmd.Run()

	if err == nil {
		// Exit 0 = no changes
		return nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		if exitCode == 1 {
			// Exit 1 = has changes (this is what we want)
			return nil
		}
		// Exit 128+ = git error (invalid ref, etc.)
		return err
	}

	// Other errors
	return err
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
