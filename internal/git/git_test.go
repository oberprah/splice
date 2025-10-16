package git

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestParseGitLogOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []GitCommit
	}{
		{
			name:  "single commit",
			input: "abc123def456789012345678901234567890abcd|John Doe|2024-01-15T10:00:00Z|Fix memory leak",
			expected: []GitCommit{
				{
					Hash:    "abc123def456789012345678901234567890abcd",
					Author:  "John Doe",
					Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					Message: "Fix memory leak",
				},
			},
		},
		{
			name: "multiple commits",
			input: `hash1|Author One|2024-01-01T10:00:00Z|First commit
hash2|Author Two|2024-01-02T11:30:00Z|Second commit
hash3|Author Three|2024-01-03T15:45:00Z|Third commit`,
			expected: []GitCommit{
				{
					Hash:    "hash1",
					Author:  "Author One",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "First commit",
				},
				{
					Hash:    "hash2",
					Author:  "Author Two",
					Date:    time.Date(2024, 1, 2, 11, 30, 0, 0, time.UTC),
					Message: "Second commit",
				},
				{
					Hash:    "hash3",
					Author:  "Author Three",
					Date:    time.Date(2024, 1, 3, 15, 45, 0, 0, time.UTC),
					Message: "Third commit",
				},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []GitCommit{},
		},
		{
			name:     "whitespace only",
			input:    "   \n\n   ",
			expected: []GitCommit{},
		},
		{
			name:  "pipe in commit message preserved",
			input: "hash|Author|2024-01-01T10:00:00Z|Fix | handling in parser",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Fix | handling in parser",
				},
			},
		},
		{
			name:  "multiple pipes in message",
			input: "hash|Author|2024-01-01T10:00:00Z|Fix A | B | C issue",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Fix A | B | C issue",
				},
			},
		},
		{
			name: "empty lines between commits ignored",
			input: `hash1|Author|2024-01-01T10:00:00Z|First

hash2|Author|2024-01-02T10:00:00Z|Second


hash3|Author|2024-01-03T10:00:00Z|Third`,
			expected: []GitCommit{
				{
					Hash:    "hash1",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "First",
				},
				{
					Hash:    "hash2",
					Author:  "Author",
					Date:    time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
					Message: "Second",
				},
				{
					Hash:    "hash3",
					Author:  "Author",
					Date:    time.Date(2024, 1, 3, 10, 0, 0, 0, time.UTC),
					Message: "Third",
				},
			},
		},
		{
			name:  "author with special characters",
			input: "hash|José García-López|2024-01-01T10:00:00Z|Add feature",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "José García-López",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "Add feature",
				},
			},
		},
		{
			name:  "empty message",
			input: "hash|Author|2024-01-01T10:00:00Z|",
			expected: []GitCommit{
				{
					Hash:    "hash",
					Author:  "Author",
					Date:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
					Message: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits, err := ParseGitLogOutput(tt.input)

			if err != nil {
				t.Fatalf("ParseGitLogOutput() error = %v", err)
			}

			if !reflect.DeepEqual(commits, tt.expected) {
				t.Errorf("ParseGitLogOutput() mismatch:\ngot:  %+v\nwant: %+v", commits, tt.expected)
			}
		})
	}
}

func TestParseGitLogOutput_InvalidDate(t *testing.T) {
	input := "hash|Author|INVALID_DATE|Message"

	commits, err := ParseGitLogOutput(input)

	if err == nil {
		t.Fatal("ParseGitLogOutput() expected error for invalid date, got nil")
	}

	if commits != nil {
		t.Errorf("ParseGitLogOutput() expected nil commits on error, got %d commits", len(commits))
	}

	// Verify error message contains the invalid date
	expectedSubstring := "INVALID_DATE"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Error message %q should contain %q", err.Error(), expectedSubstring)
	}
}

func TestParseGitLogOutput_MalformedInput(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedErrSubstr string
	}{
		{
			name:              "line without pipes",
			input:             "MALFORMED_LINE_WITHOUT_PIPES",
			expectedErrSubstr: "malformed line",
		},
		{
			name: "line with only 3 parts",
			input: `hash1|Author|2024-01-01T10:00:00Z|Valid
incomplete|line|only`,
			expectedErrSubstr: "malformed line",
		},
		{
			name: "malformed line after valid commit",
			input: `hash1|Author|2024-01-01T10:00:00Z|Valid commit
MALFORMED_LINE_WITHOUT_PIPES`,
			expectedErrSubstr: "malformed line",
		},
		{
			name:              "line with only 1 field",
			input:             "onlyonefield",
			expectedErrSubstr: "malformed line",
		},
		{
			name:              "line with 2 fields",
			input:             "field1|field2",
			expectedErrSubstr: "malformed line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits, err := ParseGitLogOutput(tt.input)

			if err == nil {
				t.Fatal("ParseGitLogOutput() expected error for malformed input, got nil")
			}

			if commits != nil {
				t.Errorf("ParseGitLogOutput() expected nil commits on error, got %d commits", len(commits))
			}

			if !strings.Contains(err.Error(), tt.expectedErrSubstr) {
				t.Errorf("Error message %q should contain %q", err.Error(), tt.expectedErrSubstr)
			}
		})
	}
}
