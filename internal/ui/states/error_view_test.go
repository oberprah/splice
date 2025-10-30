package states

import (
	"fmt"
	"testing"
)

func TestErrorState_View(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "simple error message",
			err:      fmt.Errorf("file not found"),
			expected: "  Error: file not found\n",
		},
		{
			name:     "git error message",
			err:      fmt.Errorf("not a git repository"),
			expected: "  Error: not a git repository\n",
		},
		{
			name:     "empty commits error",
			err:      fmt.Errorf("no commits found in repository"),
			expected: "  Error: no commits found in repository\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ErrorState{Err: tt.err}
			ctx := mockContext{width: 80, height: 24}

			result := s.View(ctx)

			if result != tt.expected {
				t.Errorf("View() = %q, want %q", result, tt.expected)
			}
		})
	}
}
