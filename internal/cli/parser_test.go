package cli

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

// ═══════════════════════════════════════════════════════════
// ParseCommand Tests
// ═══════════════════════════════════════════════════════════

func TestParseCommand_NoArgs(t *testing.T) {
	cmd, remaining := ParseCommand([]string{"splice"})

	if cmd != "log" {
		t.Errorf("cmd = %q, want %q", cmd, "log")
	}
	if len(remaining) != 0 {
		t.Errorf("remaining = %v, want empty slice", remaining)
	}
}

func TestParseCommand_DiffCommand(t *testing.T) {
	cmd, remaining := ParseCommand([]string{"splice", "diff"})

	if cmd != "diff" {
		t.Errorf("cmd = %q, want %q", cmd, "diff")
	}
	if len(remaining) != 0 {
		t.Errorf("remaining = %v, want empty slice", remaining)
	}
}

func TestParseCommand_DiffWithArgs(t *testing.T) {
	cmd, remaining := ParseCommand([]string{"splice", "diff", "--staged"})

	if cmd != "diff" {
		t.Errorf("cmd = %q, want %q", cmd, "diff")
	}
	if len(remaining) != 1 || remaining[0] != "--staged" {
		t.Errorf("remaining = %v, want [--staged]", remaining)
	}
}

func TestParseCommand_DiffWithMultipleArgs(t *testing.T) {
	cmd, remaining := ParseCommand([]string{"splice", "diff", "main..feature"})

	if cmd != "diff" {
		t.Errorf("cmd = %q, want %q", cmd, "diff")
	}
	if len(remaining) != 1 || remaining[0] != "main..feature" {
		t.Errorf("remaining = %v, want [main..feature]", remaining)
	}
}

func TestParseCommand_UnknownCommand(t *testing.T) {
	cmd, remaining := ParseCommand([]string{"splice", "unknown"})

	if cmd != "log" {
		t.Errorf("cmd = %q, want %q (unknown commands should default to log)", cmd, "log")
	}
	if len(remaining) != 0 {
		t.Errorf("remaining = %v, want empty slice", remaining)
	}
}

// ═══════════════════════════════════════════════════════════
// ParseDiffArgs Tests - Uncommitted Changes
// ═══════════════════════════════════════════════════════════

func TestParseDiffArgs_NoArgs_UnstagedChanges(t *testing.T) {
	args, err := ParseDiffArgs([]string{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.RawSpec != "" {
		t.Errorf("RawSpec = %q, want empty", args.RawSpec)
	}
	if args.UncommittedType == nil {
		t.Fatal("UncommittedType is nil, want non-nil")
	}
	if *args.UncommittedType != core.UncommittedTypeUnstaged {
		t.Errorf("UncommittedType = %v, want UncommittedTypeUnstaged", *args.UncommittedType)
	}
	if args.IsCommitRange() {
		t.Error("IsCommitRange() = true, want false")
	}
}

func TestParseDiffArgs_Staged(t *testing.T) {
	testCases := []string{"--staged", "--cached"}

	for _, arg := range testCases {
		t.Run(arg, func(t *testing.T) {
			args, err := ParseDiffArgs([]string{arg})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if args.RawSpec != "" {
				t.Errorf("RawSpec = %q, want empty", args.RawSpec)
			}
			if args.UncommittedType == nil {
				t.Fatal("UncommittedType is nil, want non-nil")
			}
			if *args.UncommittedType != core.UncommittedTypeStaged {
				t.Errorf("UncommittedType = %v, want UncommittedTypeStaged", *args.UncommittedType)
			}
		})
	}
}

func TestParseDiffArgs_HEAD(t *testing.T) {
	args, err := ParseDiffArgs([]string{"HEAD"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.RawSpec != "" {
		t.Errorf("RawSpec = %q, want empty", args.RawSpec)
	}
	if args.UncommittedType == nil {
		t.Fatal("UncommittedType is nil, want non-nil")
	}
	if *args.UncommittedType != core.UncommittedTypeAll {
		t.Errorf("UncommittedType = %v, want UncommittedTypeAll", *args.UncommittedType)
	}
}

// ═══════════════════════════════════════════════════════════
// ParseDiffArgs Tests - Commit Ranges
// ═══════════════════════════════════════════════════════════

func TestParseDiffArgs_CommitRange(t *testing.T) {
	args, err := ParseDiffArgs([]string{"main..feature"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.RawSpec != "main..feature" {
		t.Errorf("RawSpec = %q, want %q", args.RawSpec, "main..feature")
	}
	if args.UncommittedType != nil {
		t.Errorf("UncommittedType = %v, want nil", args.UncommittedType)
	}
	if !args.IsCommitRange() {
		t.Error("IsCommitRange() = false, want true")
	}
}

func TestParseDiffArgs_ThreeDotRange(t *testing.T) {
	args, err := ParseDiffArgs([]string{"main...feature"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.RawSpec != "main...feature" {
		t.Errorf("RawSpec = %q, want %q", args.RawSpec, "main...feature")
	}
}

func TestParseDiffArgs_SingleRef(t *testing.T) {
	args, err := ParseDiffArgs([]string{"abc123"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args.RawSpec != "abc123" {
		t.Errorf("RawSpec = %q, want %q", args.RawSpec, "abc123")
	}
}

// ═══════════════════════════════════════════════════════════
// ParseDiffArgs Tests - Validation Errors
// ═══════════════════════════════════════════════════════════

func TestParseDiffArgs_TooManyArgs(t *testing.T) {
	_, err := ParseDiffArgs([]string{"arg1", "arg2"})

	if err == nil {
		t.Error("expected error for too many arguments")
	}
}

func TestParseDiffArgs_InvalidSpec_Spaces(t *testing.T) {
	_, err := ParseDiffArgs([]string{"main feature"})

	if err == nil {
		t.Error("expected error for spec with spaces")
	}
}

func TestParseDiffArgs_InvalidSpec_ShellMetachars(t *testing.T) {
	invalidSpecs := []string{
		"main;rm -rf /",
		"main|cat /etc/passwd",
		"main&background",
		"main>output",
		"main<input",
		"main$VAR",
		"main`cmd`",
	}

	for _, spec := range invalidSpecs {
		t.Run(spec, func(t *testing.T) {
			_, err := ParseDiffArgs([]string{spec})
			if err == nil {
				t.Errorf("expected error for spec %q with shell metacharacters", spec)
			}
		})
	}
}

// ═══════════════════════════════════════════════════════════
// IsValidDiffSpec Tests
// ═══════════════════════════════════════════════════════════

func TestIsValidDiffSpec_ValidSpecs(t *testing.T) {
	validSpecs := []string{
		"main",
		"main..feature",
		"HEAD~5",
		"abc123",
		"v1.0.0",
		"origin/main",
		"refs/heads/main",
	}

	for _, spec := range validSpecs {
		t.Run(spec, func(t *testing.T) {
			if !IsValidDiffSpec(spec) {
				t.Errorf("expected %q to be valid", spec)
			}
		})
	}
}

func TestIsValidDiffSpec_InvalidSpecs(t *testing.T) {
	invalidSpecs := []string{
		"main feature",
		"main;echo",
		"main|grep",
		"main&",
		"main>file",
		"main<file",
		"$HOME",
		"`whoami`",
	}

	for _, spec := range invalidSpecs {
		t.Run(spec, func(t *testing.T) {
			if IsValidDiffSpec(spec) {
				t.Errorf("expected %q to be invalid", spec)
			}
		})
	}
}
