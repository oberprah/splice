package commands

import (
	"strings"
	"testing"
)

func TestCheckGitError(t *testing.T) {
	tests := []struct {
		name         string
		stderr       string
		err          error
		context      string
		expectedErr  string
		expectNilErr bool
	}{
		{
			name:         "no error",
			stderr:       "",
			err:          nil,
			context:      "git log",
			expectNilErr: true,
		},
		{
			name:        "not a git repository",
			stderr:      "fatal: not a git repository",
			err:         &testError{},
			context:     "git log",
			expectedErr: "not a git repository",
		},
		{
			name:        "unknown revision",
			stderr:      "fatal: unknown revision or path not in the working tree",
			err:         &testError{},
			context:     "git show",
			expectedErr: "invalid revision in git show",
		},
		{
			name:        "bad revision",
			stderr:      "fatal: bad revision 'nonexistent'",
			err:         &testError{},
			context:     "git diff",
			expectedErr: "invalid revision in git diff",
		},
		{
			name:        "generic error",
			stderr:      "some other error",
			err:         &testError{},
			context:     "git command",
			expectedErr: "git command failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkGitError(tt.stderr, tt.err, tt.context)

			if tt.expectNilErr {
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("error %q should contain %q", err.Error(), tt.expectedErr)
			}
		})
	}
}

// testError is a simple error implementation for testing
type testError struct{}

func (e *testError) Error() string {
	return "test error"
}
