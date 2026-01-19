package operations

import (
	"testing"

	"github.com/oberprah/splice/internal/core"
)

func TestUncommittedDiffSources(t *testing.T) {
	tests := []struct {
		name            string
		uncommittedType core.UncommittedType
		wantOldSource   contentSource
		wantNewSource   contentSource
		wantDiffFlags   []string
	}{
		{
			name:            "unstaged compares index to working tree",
			uncommittedType: core.UncommittedTypeUnstaged,
			wantOldSource:   sourceIndex,
			wantNewSource:   sourceWorkingTree,
			wantDiffFlags:   nil,
		},
		{
			name:            "staged compares HEAD to index",
			uncommittedType: core.UncommittedTypeStaged,
			wantOldSource:   sourceHEAD,
			wantNewSource:   sourceIndex,
			wantDiffFlags:   []string{"--staged"},
		},
		{
			name:            "all compares HEAD to working tree",
			uncommittedType: core.UncommittedTypeAll,
			wantOldSource:   sourceHEAD,
			wantNewSource:   sourceWorkingTree,
			wantDiffFlags:   []string{"HEAD"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old, new, flags := uncommittedDiffSources(tt.uncommittedType)

			if old != tt.wantOldSource {
				t.Errorf("oldSource = %v, want %v", old, tt.wantOldSource)
			}
			if new != tt.wantNewSource {
				t.Errorf("newSource = %v, want %v", new, tt.wantNewSource)
			}
			if len(flags) != len(tt.wantDiffFlags) {
				t.Errorf("diffFlags length = %d, want %d", len(flags), len(tt.wantDiffFlags))
			}
			for i, flag := range flags {
				if i < len(tt.wantDiffFlags) && flag != tt.wantDiffFlags[i] {
					t.Errorf("diffFlags[%d] = %q, want %q", i, flag, tt.wantDiffFlags[i])
				}
			}
		})
	}
}

// uncommittedDiffSources returns the content sources for a given uncommitted type.
// This is a test helper that documents the expected mapping.
func uncommittedDiffSources(t core.UncommittedType) (oldSource, newSource contentSource, diffFlags []string) {
	switch t {
	case core.UncommittedTypeUnstaged:
		return sourceIndex, sourceWorkingTree, nil
	case core.UncommittedTypeStaged:
		return sourceHEAD, sourceIndex, []string{"--staged"}
	case core.UncommittedTypeAll:
		return sourceHEAD, sourceWorkingTree, []string{"HEAD"}
	default:
		return 0, 0, nil
	}
}
