package operations

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestUncommittedFileChangesFlags(t *testing.T) {
	tests := []struct {
		name            string
		uncommittedType core.UncommittedType
		wantFlags       []string
	}{
		{
			name:            "unstaged uses git diff",
			uncommittedType: core.UncommittedTypeUnstaged,
			wantFlags:       []string{"diff"},
		},
		{
			name:            "staged uses git diff --staged",
			uncommittedType: core.UncommittedTypeStaged,
			wantFlags:       []string{"diff", "--staged"},
		},
		{
			name:            "all uses git diff HEAD",
			uncommittedType: core.UncommittedTypeAll,
			wantFlags:       []string{"diff", "HEAD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags := uncommittedFileChangesFlags(tt.uncommittedType)

			if len(flags) != len(tt.wantFlags) {
				t.Errorf("flags length = %d, want %d", len(flags), len(tt.wantFlags))
				return
			}
			for i, flag := range flags {
				if flag != tt.wantFlags[i] {
					t.Errorf("flags[%d] = %q, want %q", i, flag, tt.wantFlags[i])
				}
			}
		})
	}
}

// uncommittedFileChangesFlags returns the git diff flags for a given uncommitted type.
func uncommittedFileChangesFlags(t core.UncommittedType) []string {
	switch t {
	case core.UncommittedTypeUnstaged:
		return []string{"diff"}
	case core.UncommittedTypeStaged:
		return []string{"diff", "--staged"}
	case core.UncommittedTypeAll:
		return []string{"diff", "HEAD"}
	default:
		return nil
	}
}
