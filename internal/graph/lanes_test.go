package graph

import (
	"reflect"
	"testing"
)

func TestFindInLanes(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		lanes    []string
		expected int
	}{
		{
			name:     "empty lanes",
			hash:     "abc",
			lanes:    []string{},
			expected: -1,
		},
		{
			name:     "hash found at index 0",
			hash:     "abc",
			lanes:    []string{"abc", "def"},
			expected: 0,
		},
		{
			name:     "hash found at index 1",
			hash:     "def",
			lanes:    []string{"abc", "def"},
			expected: 1,
		},
		{
			name:     "hash not found",
			hash:     "xyz",
			lanes:    []string{"abc", "def"},
			expected: -1,
		},
		{
			name:     "hash found with empty slots",
			hash:     "def",
			lanes:    []string{"abc", "", "def"},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findInLanes(tt.hash, tt.lanes)
			if got != tt.expected {
				t.Errorf("findInLanes(%q, %v) = %d, want %d", tt.hash, tt.lanes, got, tt.expected)
			}
		})
	}
}

func TestFindEmptyLane(t *testing.T) {
	tests := []struct {
		name     string
		lanes    []string
		expected int
	}{
		{
			name:     "empty lanes slice",
			lanes:    []string{},
			expected: -1,
		},
		{
			name:     "no empty slots",
			lanes:    []string{"abc", "def"},
			expected: -1,
		},
		{
			name:     "empty slot at index 0",
			lanes:    []string{"", "def"},
			expected: 0,
		},
		{
			name:     "empty slot at index 1",
			lanes:    []string{"abc", ""},
			expected: 1,
		},
		{
			name:     "multiple empty slots returns first",
			lanes:    []string{"abc", "", "", "def"},
			expected: 1,
		},
		{
			name:     "all empty",
			lanes:    []string{"", "", ""},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findEmptyLane(tt.lanes)
			if got != tt.expected {
				t.Errorf("findEmptyLane(%v) = %d, want %d", tt.lanes, got, tt.expected)
			}
		})
	}
}

func TestAssignColumn(t *testing.T) {
	tests := []struct {
		name          string
		hash          string
		lanes         []string
		expectedCol   int
		expectedLanes []string
	}{
		{
			name:          "empty lanes - append new",
			hash:          "abc",
			lanes:         []string{},
			expectedCol:   0,
			expectedLanes: []string{"abc"},
		},
		{
			name:          "hash already in lane - reuse position",
			hash:          "abc",
			lanes:         []string{"abc", "def"},
			expectedCol:   0,
			expectedLanes: []string{"abc", "def"},
		},
		{
			name:          "hash not found, empty slot exists - use empty slot",
			hash:          "xyz",
			lanes:         []string{"abc", "", "def"},
			expectedCol:   1,
			expectedLanes: []string{"abc", "xyz", "def"},
		},
		{
			name:          "hash not found, no empty slots - append",
			hash:          "xyz",
			lanes:         []string{"abc", "def"},
			expectedCol:   2,
			expectedLanes: []string{"abc", "def", "xyz"},
		},
		{
			name:          "first commit",
			hash:          "initial",
			lanes:         []string{},
			expectedCol:   0,
			expectedLanes: []string{"initial"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			lanes := make([]string, len(tt.lanes))
			copy(lanes, tt.lanes)

			col, resultLanes := assignColumn(tt.hash, lanes)
			if col != tt.expectedCol {
				t.Errorf("assignColumn(%q, %v) col = %d, want %d", tt.hash, tt.lanes, col, tt.expectedCol)
			}
			if !reflect.DeepEqual(resultLanes, tt.expectedLanes) {
				t.Errorf("assignColumn(%q, %v) lanes = %v, want %v", tt.hash, tt.lanes, resultLanes, tt.expectedLanes)
			}
		})
	}
}

func TestCollapseTrailingEmpty(t *testing.T) {
	tests := []struct {
		name     string
		lanes    []string
		expected []string
	}{
		{
			name:     "empty slice",
			lanes:    []string{},
			expected: []string{},
		},
		{
			name:     "no trailing empty",
			lanes:    []string{"abc", "def"},
			expected: []string{"abc", "def"},
		},
		{
			name:     "one trailing empty",
			lanes:    []string{"abc", "def", ""},
			expected: []string{"abc", "def"},
		},
		{
			name:     "multiple trailing empty",
			lanes:    []string{"abc", "", "def", "", ""},
			expected: []string{"abc", "", "def"},
		},
		{
			name:     "all empty",
			lanes:    []string{"", "", ""},
			expected: []string{},
		},
		{
			name:     "empty in middle preserved",
			lanes:    []string{"abc", "", "def"},
			expected: []string{"abc", "", "def"},
		},
		{
			name:     "single non-empty",
			lanes:    []string{"abc"},
			expected: []string{"abc"},
		},
		{
			name:     "single empty",
			lanes:    []string{""},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			lanes := make([]string, len(tt.lanes))
			copy(lanes, tt.lanes)

			got := collapseTrailingEmpty(lanes)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("collapseTrailingEmpty(%v) = %v, want %v", tt.lanes, got, tt.expected)
			}
		})
	}
}

// TestUpdateLanes_ConvergingColumnReuse tests that converging columns are
// reused for merge parents. This behavior is achieved by clearing converging
// columns in ComputeLayout before calling updateLanes, so updateLanes naturally
// uses the empty slots.
func TestUpdateLanes_ConvergingColumnReuse(t *testing.T) {
	// Scenario: D is at col 0, col 1 was cleared (was converging)
	// D's parents are [A, C]
	// Expected: A goes to col 0, C goes into col 1 (the cleared converging column)
	col := 0
	parents := []string{"A", "C"}
	lanes := []string{"D", ""} // col 1 cleared by ComputeLayout before calling updateLanes

	result := updateLanes(col, parents, lanes)

	// C should go into col 1 (the empty slot)
	expectedLanes := []string{"A", "C"}
	expectedMergeColumns := []int{1}

	if len(result.Lanes) != len(expectedLanes) {
		t.Errorf("Lanes length: got %d, want %d", len(result.Lanes), len(expectedLanes))
	}

	for i, exp := range expectedLanes {
		if i < len(result.Lanes) && result.Lanes[i] != exp {
			t.Errorf("Lanes[%d]: got %q, want %q", i, result.Lanes[i], exp)
		}
	}

	if len(result.MergeColumns) != len(expectedMergeColumns) {
		t.Errorf("MergeColumns: got %v, want %v", result.MergeColumns, expectedMergeColumns)
	}
}

func TestUpdateLanes(t *testing.T) {
	tests := []struct {
		name                 string
		col                  int
		parents              []string
		lanes                []string
		expectedLanes        []string
		expectedMergeColumns []int
	}{
		{
			name:                 "root commit - clears lane",
			col:                  0,
			parents:              []string{},
			lanes:                []string{"A"},
			expectedLanes:        []string{""},
			expectedMergeColumns: []int{},
		},
		{
			name:                 "single parent - replaces position",
			col:                  0,
			parents:              []string{"parent"},
			lanes:                []string{"commit"},
			expectedLanes:        []string{"parent"},
			expectedMergeColumns: []int{},
		},
		{
			name:                 "merge commit - adds second parent",
			col:                  0,
			parents:              []string{"parent1", "parent2"},
			lanes:                []string{"commit"},
			expectedLanes:        []string{"parent1", "parent2"},
			expectedMergeColumns: []int{1},
		},
		{
			name:                 "merge commit - second parent in empty slot",
			col:                  0,
			parents:              []string{"parent1", "parent2"},
			lanes:                []string{"commit", ""},
			expectedLanes:        []string{"parent1", "parent2"},
			expectedMergeColumns: []int{1},
		},
		{
			name:                 "merge commit - second parent already in lane",
			col:                  0,
			parents:              []string{"parent1", "parent2"},
			lanes:                []string{"commit", "parent2"},
			expectedLanes:        []string{"parent1", "parent2"},
			expectedMergeColumns: []int{1},
		},
		{
			name:                 "octopus merge - three parents",
			col:                  0,
			parents:              []string{"p1", "p2", "p3"},
			lanes:                []string{"commit"},
			expectedLanes:        []string{"p1", "p2", "p3"},
			expectedMergeColumns: []int{1, 2},
		},
		{
			name:                 "merge in middle column",
			col:                  1,
			parents:              []string{"parent1", "parent2"},
			lanes:                []string{"other", "commit", ""},
			expectedLanes:        []string{"other", "parent1", "parent2"},
			expectedMergeColumns: []int{2},
		},
		{
			name:                 "merge parent prefers slot to the right",
			col:                  0,
			parents:              []string{"p1", "p2"},
			lanes:                []string{"commit", "occupied", ""},
			expectedLanes:        []string{"p1", "occupied", "p2"},
			expectedMergeColumns: []int{2},
		},
		{
			name:                 "single parent with multiple lanes",
			col:                  0,
			parents:              []string{"parent"},
			lanes:                []string{"commit", "other"},
			expectedLanes:        []string{"parent", "other"},
			expectedMergeColumns: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid modifying test data
			lanes := make([]string, len(tt.lanes))
			copy(lanes, tt.lanes)

			result := updateLanes(tt.col, tt.parents, lanes)

			if !reflect.DeepEqual(result.Lanes, tt.expectedLanes) {
				t.Errorf("updateLanes() lanes = %v, want %v", result.Lanes, tt.expectedLanes)
			}
			if !reflect.DeepEqual(result.MergeColumns, tt.expectedMergeColumns) {
				t.Errorf("updateLanes() mergeColumns = %v, want %v", result.MergeColumns, tt.expectedMergeColumns)
			}
		})
	}
}
