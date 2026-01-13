package core

import (
	"testing"
	"time"
)

// ═══════════════════════════════════════════════════════════
// Test Helpers
// ═══════════════════════════════════════════════════════════

func makeTestCommitForDiffSource(hash string) GitCommit {
	return GitCommit{
		Hash:    hash,
		Message: "Test commit",
		Author:  "Test Author",
		Date:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

// ═══════════════════════════════════════════════════════════
// CommitRangeDiffSource Tests
// ═══════════════════════════════════════════════════════════

func TestCommitRangeDiffSource_Construction(t *testing.T) {
	start := makeTestCommitForDiffSource("abc123")
	end := makeTestCommitForDiffSource("def456")

	source := CommitRangeDiffSource{
		Start: start,
		End:   end,
		Count: 5,
	}

	if source.Start.Hash != "abc123" {
		t.Errorf("Start.Hash = %s, want abc123", source.Start.Hash)
	}
	if source.End.Hash != "def456" {
		t.Errorf("End.Hash = %s, want def456", source.End.Hash)
	}
	if source.Count != 5 {
		t.Errorf("Count = %d, want 5", source.Count)
	}
}

func TestCommitRangeDiffSource_SealedInterface(t *testing.T) {
	start := makeTestCommitForDiffSource("abc123")
	end := makeTestCommitForDiffSource("def456")

	var source DiffSource = CommitRangeDiffSource{
		Start: start,
		End:   end,
		Count: 5,
	}

	// Type assertion should work
	commitRange, ok := source.(CommitRangeDiffSource)
	if !ok {
		t.Error("Failed to type assert DiffSource to CommitRangeDiffSource")
	}
	if commitRange.Start.Hash != "abc123" {
		t.Errorf("Type asserted Start.Hash = %s, want abc123", commitRange.Start.Hash)
	}
}

// ═══════════════════════════════════════════════════════════
// UncommittedChangesDiffSource Tests
// ═══════════════════════════════════════════════════════════

func TestUncommittedChangesDiffSource_Construction_Unstaged(t *testing.T) {
	source := UncommittedChangesDiffSource{
		Type: UncommittedTypeUnstaged,
	}

	if source.Type != UncommittedTypeUnstaged {
		t.Errorf("Type = %d, want UncommittedTypeUnstaged", source.Type)
	}
}

func TestUncommittedChangesDiffSource_Construction_Staged(t *testing.T) {
	source := UncommittedChangesDiffSource{
		Type: UncommittedTypeStaged,
	}

	if source.Type != UncommittedTypeStaged {
		t.Errorf("Type = %d, want UncommittedTypeStaged", source.Type)
	}
}

func TestUncommittedChangesDiffSource_Construction_All(t *testing.T) {
	source := UncommittedChangesDiffSource{
		Type: UncommittedTypeAll,
	}

	if source.Type != UncommittedTypeAll {
		t.Errorf("Type = %d, want UncommittedTypeAll", source.Type)
	}
}

func TestUncommittedChangesDiffSource_SealedInterface(t *testing.T) {
	var source DiffSource = UncommittedChangesDiffSource{
		Type: UncommittedTypeStaged,
	}

	// Type assertion should work
	uncommitted, ok := source.(UncommittedChangesDiffSource)
	if !ok {
		t.Error("Failed to type assert DiffSource to UncommittedChangesDiffSource")
	}
	if uncommitted.Type != UncommittedTypeStaged {
		t.Errorf("Type asserted Type = %d, want UncommittedTypeStaged", uncommitted.Type)
	}
}

// ═══════════════════════════════════════════════════════════
// UncommittedType Enum Tests
// ═══════════════════════════════════════════════════════════

func TestUncommittedType_EnumValues(t *testing.T) {
	// Test that enum values have expected distinct values
	if UncommittedTypeUnstaged == UncommittedTypeStaged {
		t.Error("UncommittedTypeUnstaged and UncommittedTypeStaged should have different values")
	}
	if UncommittedTypeStaged == UncommittedTypeAll {
		t.Error("UncommittedTypeStaged and UncommittedTypeAll should have different values")
	}
	if UncommittedTypeUnstaged == UncommittedTypeAll {
		t.Error("UncommittedTypeUnstaged and UncommittedTypeAll should have different values")
	}
}

func TestUncommittedType_StartValue(t *testing.T) {
	// Test that iota starts at 0 as expected
	if UncommittedTypeUnstaged != 0 {
		t.Errorf("UncommittedTypeUnstaged = %d, want 0", UncommittedTypeUnstaged)
	}
}

func TestUncommittedType_SequentialValues(t *testing.T) {
	// Test that values are sequential
	if int(UncommittedTypeStaged) != int(UncommittedTypeUnstaged)+1 {
		t.Error("UncommittedTypeStaged should be UncommittedTypeUnstaged + 1")
	}
	if int(UncommittedTypeAll) != int(UncommittedTypeStaged)+1 {
		t.Error("UncommittedTypeAll should be UncommittedTypeStaged + 1")
	}
}

// ═══════════════════════════════════════════════════════════
// DiffSource Interface Tests
// ═══════════════════════════════════════════════════════════

func TestDiffSource_TypeSwitch(t *testing.T) {
	// Test that we can use type switches with DiffSource
	testCases := []struct {
		name     string
		source   DiffSource
		expected string
	}{
		{
			name: "CommitRangeDiffSource",
			source: CommitRangeDiffSource{
				Start: makeTestCommitForDiffSource("abc123"),
				End:   makeTestCommitForDiffSource("def456"),
				Count: 5,
			},
			expected: "commit_range",
		},
		{
			name: "UncommittedChangesDiffSource_Unstaged",
			source: UncommittedChangesDiffSource{
				Type: UncommittedTypeUnstaged,
			},
			expected: "uncommitted_unstaged",
		},
		{
			name: "UncommittedChangesDiffSource_Staged",
			source: UncommittedChangesDiffSource{
				Type: UncommittedTypeStaged,
			},
			expected: "uncommitted_staged",
		},
		{
			name: "UncommittedChangesDiffSource_All",
			source: UncommittedChangesDiffSource{
				Type: UncommittedTypeAll,
			},
			expected: "uncommitted_all",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifyDiffSource(tc.source)
			if result != tc.expected {
				t.Errorf("classifyDiffSource(%v) = %s, want %s", tc.name, result, tc.expected)
			}
		})
	}
}

// Helper function to test type switches work correctly
func classifyDiffSource(source DiffSource) string {
	switch s := source.(type) {
	case CommitRangeDiffSource:
		return "commit_range"
	case UncommittedChangesDiffSource:
		switch s.Type {
		case UncommittedTypeUnstaged:
			return "uncommitted_unstaged"
		case UncommittedTypeStaged:
			return "uncommitted_staged"
		case UncommittedTypeAll:
			return "uncommitted_all"
		}
	}
	return "unknown"
}

// ═══════════════════════════════════════════════════════════
// UncommittedType.String() Tests
// ═══════════════════════════════════════════════════════════

func TestUncommittedType_String_ValidValues(t *testing.T) {
	testCases := []struct {
		value    UncommittedType
		expected string
	}{
		{UncommittedTypeUnstaged, "UncommittedTypeUnstaged"},
		{UncommittedTypeStaged, "UncommittedTypeStaged"},
		{UncommittedTypeAll, "UncommittedTypeAll"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.value.String()
			if result != tc.expected {
				t.Errorf("String() = %s, want %s", result, tc.expected)
			}
		})
	}
}

func TestUncommittedType_String_InvalidValue_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid UncommittedType, but did not panic")
		}
	}()

	invalidType := UncommittedType(999)
	_ = invalidType.String() // Should panic
}
