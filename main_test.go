package main

import (
	"strings"
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "no arguments defaults to log",
			args:     []string{"splice"},
			wantCmd:  "log",
			wantArgs: []string{},
		},
		{
			name:     "diff with no args",
			args:     []string{"splice", "diff"},
			wantCmd:  "diff",
			wantArgs: []string{},
		},
		{
			name:     "diff with --staged flag",
			args:     []string{"splice", "diff", "--staged"},
			wantCmd:  "diff",
			wantArgs: []string{"--staged"},
		},
		{
			name:     "diff with --cached flag",
			args:     []string{"splice", "diff", "--cached"},
			wantCmd:  "diff",
			wantArgs: []string{"--cached"},
		},
		{
			name:     "diff with HEAD",
			args:     []string{"splice", "diff", "HEAD"},
			wantCmd:  "diff",
			wantArgs: []string{"HEAD"},
		},
		{
			name:     "diff with commit range",
			args:     []string{"splice", "diff", "main..feature"},
			wantCmd:  "diff",
			wantArgs: []string{"main..feature"},
		},
		{
			name:     "diff with three-dot range",
			args:     []string{"splice", "diff", "main...feature"},
			wantCmd:  "diff",
			wantArgs: []string{"main...feature"},
		},
		{
			name:     "diff with multiple arguments",
			args:     []string{"splice", "diff", "HEAD~5..HEAD", "extra"},
			wantCmd:  "diff",
			wantArgs: []string{"HEAD~5..HEAD", "extra"},
		},
		{
			name:     "unknown command defaults to log",
			args:     []string{"splice", "unknown"},
			wantCmd:  "log",
			wantArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCmd, gotArgs := parseArgs(tt.args)
			if gotCmd != tt.wantCmd {
				t.Errorf("parseArgs() cmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("parseArgs() args length = %v, want %v", len(gotArgs), len(tt.wantArgs))
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("parseArgs() args[%d] = %v, want %v", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestParseDiffSpec(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantRawSpec    string
		wantUncommited *core.UncommittedType
		wantErr        bool
	}{
		{
			name:           "no args means unstaged changes",
			args:           []string{},
			wantUncommited: ptrUncommittedType(core.UncommittedTypeUnstaged),
			wantErr:        false,
		},
		{
			name:           "--staged flag",
			args:           []string{"--staged"},
			wantUncommited: ptrUncommittedType(core.UncommittedTypeStaged),
			wantErr:        false,
		},
		{
			name:           "--cached flag",
			args:           []string{"--cached"},
			wantUncommited: ptrUncommittedType(core.UncommittedTypeStaged),
			wantErr:        false,
		},
		{
			name:           "HEAD means all uncommitted",
			args:           []string{"HEAD"},
			wantUncommited: ptrUncommittedType(core.UncommittedTypeAll),
			wantErr:        false,
		},
		{
			name:        "commit range two-dot",
			args:        []string{"main..feature"},
			wantRawSpec: "main..feature",
			wantErr:     false,
		},
		{
			name:        "commit range three-dot",
			args:        []string{"main...feature"},
			wantRawSpec: "main...feature",
			wantErr:     false,
		},
		{
			name:        "commit range with tildes",
			args:        []string{"HEAD~5..HEAD"},
			wantRawSpec: "HEAD~5..HEAD",
			wantErr:     false,
		},
		{
			name:        "single commit",
			args:        []string{"abc123"},
			wantRawSpec: "abc123",
			wantErr:     false,
		},
		{
			name:    "extra args after --staged",
			args:    []string{"--staged", "extra"},
			wantErr: true,
		},
		{
			name:    "extra args after HEAD",
			args:    []string{"HEAD", "extra"},
			wantErr: true,
		},
		{
			name:    "extra args after range",
			args:    []string{"main..feature", "extra"},
			wantErr: true,
		},
		{
			name:    "invalid spec with spaces",
			args:    []string{"invalid spec"},
			wantErr: true,
		},
		{
			name:    "invalid spec with semicolon",
			args:    []string{"main;feature"},
			wantErr: true,
		},
		{
			name:    "invalid spec with pipe",
			args:    []string{"main|feature"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRawSpec, gotUncommitted, err := parseDiffSpec(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDiffSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tt.wantUncommited != nil {
				if gotUncommitted == nil {
					t.Errorf("parseDiffSpec() gotUncommitted = nil, want %v", *tt.wantUncommited)
					return
				}
				if *gotUncommitted != *tt.wantUncommited {
					t.Errorf("parseDiffSpec() gotUncommitted = %v, want %v", *gotUncommitted, *tt.wantUncommited)
				}
			} else {
				if gotUncommitted != nil {
					t.Errorf("parseDiffSpec() gotUncommitted = %v, want nil", *gotUncommitted)
				}
				if gotRawSpec != tt.wantRawSpec {
					t.Errorf("parseDiffSpec() gotRawSpec = %v, want %v", gotRawSpec, tt.wantRawSpec)
				}
			}
		})
	}
}

func ptrUncommittedType(t core.UncommittedType) *core.UncommittedType {
	return &t
}

func TestValidateDiffSpec(t *testing.T) {
	tests := []struct {
		name        string
		rawSpec     string
		uncommitted *core.UncommittedType
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid uncommitted unstaged with changes",
			uncommitted: ptrUncommittedType(core.UncommittedTypeUnstaged),
			wantErr:     false,
		},
		{
			name:        "valid uncommitted staged with changes",
			uncommitted: ptrUncommittedType(core.UncommittedTypeStaged),
			wantErr:     false,
		},
		{
			name:        "valid uncommitted all with changes",
			uncommitted: ptrUncommittedType(core.UncommittedTypeAll),
			wantErr:     false,
		},
		{
			name:        "uncommitted unstaged with no changes",
			uncommitted: ptrUncommittedType(core.UncommittedTypeUnstaged),
			wantErr:     true,
			errContains: "no changes",
		},
		{
			name:    "valid commit range",
			rawSpec: "HEAD~5..HEAD",
			wantErr: false,
		},
		{
			name:        "invalid ref",
			rawSpec:     "nonexistent..HEAD",
			wantErr:     true,
			errContains: "invalid",
		},
		{
			name:        "commit range with no changes",
			rawSpec:     "HEAD..HEAD",
			wantErr:     true,
			errContains: "no changes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that require specific git state
			// These are integration tests that would need a test repo
			t.Skip("Integration test - requires specific git state")

			err := validateDiffSpec(tt.rawSpec, tt.uncommitted)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDiffSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validateDiffSpec() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestIsValidDiffSpec(t *testing.T) {
	tests := []struct {
		name  string
		spec  string
		valid bool
	}{
		{
			name:  "simple ref",
			spec:  "main",
			valid: true,
		},
		{
			name:  "two-dot range",
			spec:  "main..feature",
			valid: true,
		},
		{
			name:  "three-dot range",
			spec:  "main...feature",
			valid: true,
		},
		{
			name:  "tilde notation",
			spec:  "HEAD~5..HEAD",
			valid: true,
		},
		{
			name:  "caret notation",
			spec:  "HEAD^..HEAD",
			valid: true,
		},
		{
			name:  "hash-like",
			spec:  "abc123def456",
			valid: true,
		},
		{
			name:  "with slash (remote branch)",
			spec:  "origin/main..feature",
			valid: true,
		},
		{
			name:  "invalid with space",
			spec:  "main feature",
			valid: false,
		},
		{
			name:  "invalid with semicolon",
			spec:  "main;feature",
			valid: false,
		},
		{
			name:  "invalid with pipe",
			spec:  "main|feature",
			valid: false,
		},
		{
			name:  "invalid with ampersand",
			spec:  "main&feature",
			valid: false,
		},
		{
			name:  "invalid with redirect",
			spec:  "main>file",
			valid: false,
		},
		{
			name:  "invalid with dollar",
			spec:  "main$feature",
			valid: false,
		},
		{
			name:  "invalid with backtick",
			spec:  "main`cmd`",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidDiffSpec(tt.spec)
			if got != tt.valid {
				t.Errorf("isValidDiffSpec(%q) = %v, want %v", tt.spec, got, tt.valid)
			}
		})
	}
}
