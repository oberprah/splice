package main

import (
	"testing"

	"github.com/oberprah/splice/internal/cli"
	"github.com/oberprah/splice/internal/core"
)

// Tests for CLI parsing functionality are in internal/cli/parser_test.go
// This file contains integration tests for main.go

func TestParseCommand_Integration(t *testing.T) {
	// Basic sanity check that cli.ParseCommand works from main package
	cmd, args := cli.ParseCommand([]string{"splice", "diff", "--staged"})

	if cmd != "diff" {
		t.Errorf("ParseCommand() cmd = %v, want diff", cmd)
	}
	if len(args) != 1 || args[0] != "--staged" {
		t.Errorf("ParseCommand() args = %v, want [--staged]", args)
	}
}

func TestParseDiffArgs_Integration(t *testing.T) {
	// Test that ParseDiffArgs returns correct types
	tests := []struct {
		name            string
		args            []string
		wantCommitRange bool
		wantType        *core.UncommittedType
	}{
		{
			name:            "no args returns unstaged",
			args:            []string{},
			wantCommitRange: false,
			wantType:        ptrUncommittedType(core.UncommittedTypeUnstaged),
		},
		{
			name:            "--staged returns staged",
			args:            []string{"--staged"},
			wantCommitRange: false,
			wantType:        ptrUncommittedType(core.UncommittedTypeStaged),
		},
		{
			name:            "HEAD returns all uncommitted",
			args:            []string{"HEAD"},
			wantCommitRange: false,
			wantType:        ptrUncommittedType(core.UncommittedTypeAll),
		},
		{
			name:            "commit range returns commit range",
			args:            []string{"main..feature"},
			wantCommitRange: true,
			wantType:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffArgs, err := cli.ParseDiffArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diffArgs.IsCommitRange() != tt.wantCommitRange {
				t.Errorf("IsCommitRange() = %v, want %v", diffArgs.IsCommitRange(), tt.wantCommitRange)
			}

			if tt.wantType != nil {
				if diffArgs.UncommittedType == nil {
					t.Fatalf("UncommittedType is nil, want %v", *tt.wantType)
				}
				if *diffArgs.UncommittedType != *tt.wantType {
					t.Errorf("UncommittedType = %v, want %v", *diffArgs.UncommittedType, *tt.wantType)
				}
			} else if diffArgs.UncommittedType != nil {
				t.Errorf("UncommittedType = %v, want nil", *diffArgs.UncommittedType)
			}
		})
	}
}

func ptrUncommittedType(t core.UncommittedType) *core.UncommittedType {
	return &t
}
