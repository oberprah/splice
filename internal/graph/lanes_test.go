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
