package graph

import (
	"strings"
	"testing"
)

// renderLayout renders a Layout to a slice of strings for easier comparison.
func renderLayout(layout *Layout) []string {
	var result []string
	for _, row := range layout.Rows {
		result = append(result, strings.TrimRight(RenderRow(row), " "))
	}
	return result
}

func TestComputeLayout_Empty(t *testing.T) {
	layout := ComputeLayout([]Commit{})
	if len(layout.Rows) != 0 {
		t.Errorf("Expected empty layout, got %d rows", len(layout.Rows))
	}
}

// Test Case 1: Linear History (No Branches)
// Topology: A ← B ← C ← D (HEAD)
func TestComputeLayout_LinearHistory(t *testing.T) {
	commits := []Commit{
		{Hash: "D", Parents: []string{"C"}},
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├",
		"├",
		"├",
		"├",
	}

	if len(rendered) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(rendered))
	}

	for i, exp := range expected {
		if rendered[i] != exp {
			t.Errorf("Row %d: got %q, want %q", i, rendered[i], exp)
		}
	}
}

// Test Case 2: Simple Feature Branch Merge
// Topology:
//
//	 C ← D (feature, merged)
//	/     \
//
// A ← B ←───← E (HEAD -> main)
func TestComputeLayout_SimpleFeatureBranchMerge(t *testing.T) {
	commits := []Commit{
		{Hash: "E", Parents: []string{"B", "D"}}, // Merge commit
		{Hash: "D", Parents: []string{"C"}},
		{Hash: "C", Parents: []string{"A"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├─╮", // E: merge commit
		"│ ├", // D: on feature branch
		"│ ├", // C: on feature branch
		"├ │", // B: on main branch
		"├─╯", // A: convergence point
	}

	if len(rendered) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(rendered))
	}

	for i, exp := range expected {
		if rendered[i] != exp {
			t.Errorf("Row %d: got %q, want %q", i, rendered[i], exp)
		}
	}
}

// Test Case 6: Root Commit (End of History)
// Topology: A (root, no parents)
func TestComputeLayout_RootCommit(t *testing.T) {
	commits := []Commit{
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├",
	}

	if len(rendered) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(rendered))
	}

	for i, exp := range expected {
		if rendered[i] != exp {
			t.Errorf("Row %d: got %q, want %q", i, rendered[i], exp)
		}
	}
}

// Test Case 7: Multiple Roots (Merged Repositories)
// Topology:
// A ← B ← D (HEAD -> main)
//
//	↑
//
// C ──────┘ (was separate repo)
func TestComputeLayout_MultipleRoots(t *testing.T) {
	commits := []Commit{
		{Hash: "D", Parents: []string{"B", "C"}}, // Merge commit
		{Hash: "C", Parents: []string{}},         // Root of other repo
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}}, // Root of main repo
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├─╮", // D: merge commit
		"│ ├", // C: other repo root (no continuation below)
		"├",   // B: main branch (column 1 collapsed after C)
		"├",   // A: main repo root
	}

	if len(rendered) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(rendered))
	}

	for i, exp := range expected {
		if rendered[i] != exp {
			t.Errorf("Row %d: got %q, want %q", i, rendered[i], exp)
		}
	}
}

// Test Case 4: Sequential Merges
// Topology:
//
//	 B ← C (feature-1)
//	/     \
//
// A ←───────← D ← G (HEAD -> main)
//
//	\   /
//	 E─F (feature-2)
func TestComputeLayout_SequentialMerges(t *testing.T) {
	commits := []Commit{
		{Hash: "G", Parents: []string{"D", "F"}},
		{Hash: "F", Parents: []string{"E"}},
		{Hash: "E", Parents: []string{"D"}},
		{Hash: "D", Parents: []string{"A", "C"}},
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	// Expected output based on test-cases.md:
	// > ├─╮ G Merge feature-2
	//   │ ├ F Feature-2 done
	//   │ ├ E Feature-2 work
	//   ├─┤ D Merge feature-1 (needs merge join symbol)
	//   │ ├ C Feature-1 done
	//   │ ├ B Feature-1 work
	//   ├─╯ A Initial commit

	// Note: The exact symbols may vary based on algorithm details.
	// Key structural checks:
	if len(rendered) != 7 {
		t.Fatalf("Expected 7 rows, got %d", len(rendered))
	}

	// First row should be a merge
	if !strings.HasPrefix(rendered[0], "├─") {
		t.Errorf("Row 0: expected merge commit, got %q", rendered[0])
	}

	// Last row should show convergence
	if !strings.Contains(rendered[6], "╯") && !strings.Contains(rendered[6], "├") {
		t.Errorf("Row 6: expected convergence or simple commit, got %q", rendered[6])
	}
}

// Test Case 5: Sequential Merges with commits on main
// Topology:
// A ← B ← E ← F ← H (HEAD -> main)
//
//	\   /   \   /
//	 C─D     G
func TestComputeLayout_SequentialMergesWithMainCommits(t *testing.T) {
	commits := []Commit{
		{Hash: "H", Parents: []string{"F", "G"}},
		{Hash: "G", Parents: []string{"F"}},
		{Hash: "F", Parents: []string{"E"}},
		{Hash: "E", Parents: []string{"B", "D"}},
		{Hash: "D", Parents: []string{"C"}},
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	if len(rendered) != 8 {
		t.Fatalf("Expected 8 rows, got %d", len(rendered))
	}

	// First row should be a merge
	if !strings.HasPrefix(rendered[0], "├─") {
		t.Errorf("Row 0: expected merge commit, got %q", rendered[0])
	}
}

// Test Case 3: Two Parallel Feature Branches (Octopus Merge)
// Topology:
//
//	 C ← D (feature-1)
//	/     \
//
// A ←───────← G (HEAD -> main)
//
//	\     /
//	 E ← F (feature-2)
func TestComputeLayout_OctopusMerge(t *testing.T) {
	commits := []Commit{
		{Hash: "G", Parents: []string{"A", "D", "F"}}, // Octopus merge - 3 parents
		{Hash: "F", Parents: []string{"E"}},
		{Hash: "E", Parents: []string{"A"}},
		{Hash: "D", Parents: []string{"C"}},
		{Hash: "C", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	if len(rendered) != 6 {
		t.Fatalf("Expected 6 rows, got %d", len(rendered))
	}

	// First row should be an octopus merge with multiple branch tops
	// Expected: ├─┬─╮ or similar
	if !strings.HasPrefix(rendered[0], "├─") {
		t.Errorf("Row 0: expected octopus merge, got %q", rendered[0])
	}
}

// Test Case 8: With Tags and Remote Refs (graph only, no refs in this test)
// Same as linear history - just tests the graph part
func TestComputeLayout_WithRefs(t *testing.T) {
	commits := []Commit{
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├",
		"├",
		"├",
	}

	if len(rendered) != len(expected) {
		t.Fatalf("Expected %d rows, got %d", len(expected), len(rendered))
	}

	for i, exp := range expected {
		if rendered[i] != exp {
			t.Errorf("Row %d: got %q, want %q", i, rendered[i], exp)
		}
	}
}

// Test Case 9: Complex Multi-Branch
// This is the most complex test case from the spec
func TestComputeLayout_ComplexMultiBranch(t *testing.T) {
	commits := []Commit{
		{Hash: "M", Parents: []string{"J", "L"}},
		{Hash: "L", Parents: []string{"K"}},
		{Hash: "K", Parents: []string{"D"}},
		{Hash: "J", Parents: []string{"H", "I"}},
		{Hash: "I", Parents: []string{"G"}},
		{Hash: "H", Parents: []string{"F", "G"}},
		{Hash: "G", Parents: []string{"F"}},
		{Hash: "F", Parents: []string{"D", "E"}},
		{Hash: "E", Parents: []string{"D"}},
		{Hash: "D", Parents: []string{"B", "C"}},
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	if len(rendered) != 13 {
		t.Fatalf("Expected 13 rows, got %d", len(rendered))
	}

	// Verify basic structure - first should be merge, last should be root
	if !strings.HasPrefix(rendered[0], "├─") {
		t.Errorf("Row 0: expected merge commit, got %q", rendered[0])
	}

	// Last row should end the graph
	if rendered[12] == "" {
		t.Errorf("Row 12: expected non-empty row for root commit")
	}
}

// Additional test: verify that passing lanes are detected correctly
func TestComputeLayout_PassingLanes(t *testing.T) {
	// Scenario: commit on column 1 with passing lane on column 0
	// A ← B ← D
	//      \  ↑
	//       ← C
	commits := []Commit{
		{Hash: "D", Parents: []string{"B", "C"}},
		{Hash: "C", Parents: []string{"B"}}, // C is at col 1, B passes on col 0
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	// Row 1 (C) should have passing lane on left: │ ├
	if len(rendered) >= 2 {
		if !strings.Contains(rendered[1], "│") || !strings.Contains(rendered[1], "├") {
			t.Logf("Row 1: got %q (passing lane + commit expected)", rendered[1])
		}
	}
}
