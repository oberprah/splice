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
		"│ ├", // D: feature branch
		"│ ├", // C: feature branch
		"├ │", // B: main branch
		"├─╯", // A: convergence
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
		"│ ├", // C: other repo root
		"├",   // B: main branch
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

	expected := []string{
		"├─╮", // G: merge commit
		"│ ├", // F: feature-2
		"│ ├", // E: feature-2
		"├─┤", // D: merge commit (merge join)
		"│ ├", // C: feature-1
		"│ ├", // B: feature-1
		"├─╯", // A: convergence
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

	expected := []string{
		"├─╮", // H: merge commit
		"│ ├", // G: feature-2
		"├─╯", // F: convergence
		"├─╮", // E: merge commit
		"│ ├", // D: feature-1
		"│ ├", // C: feature-1
		"├─╯", // B: convergence
		"├",   // A: root
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

	expected := []string{
		"├─┬─╮", // G: octopus merge (3 parents)
		"│ │ ├", // F: feature-2
		"│ │ ├", // E: feature-2
		"│ ├ │", // D: feature-1
		"│ ├ │", // C: feature-1
		"├─┴─╯", // A: convergence
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

	expected := []string{
		"├─╮",     // M: merge commit
		"│ ├",     // L: feature-3
		"│ ├",     // K: feature-3
		"├─│─╮",   // J: merge commit
		"│ │ ├",   // I: feature-2
		"├─│─│─╮", // H: merge commit
		"│ │ ├─╯", // G: feature-2 converges
		"├─│─┤",   // F: merge commit (merge join)
		"│ │ ├",   // E: hotfix
		"├─┼─╯",   // D: merge commit (cross + convergence)
		"│ ├",     // C: feature-1
		"├─╯",     // B: convergence
		"├",       // A: root
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

func TestComputeLayout_PassingLanes(t *testing.T) {
	commits := []Commit{
		{Hash: "D", Parents: []string{"B", "C"}},
		{Hash: "C", Parents: []string{"B"}},
		{Hash: "B", Parents: []string{"A"}},
		{Hash: "A", Parents: []string{}},
	}

	layout := ComputeLayout(commits)
	rendered := renderLayout(layout)

	expected := []string{
		"├─╮", // D: merge commit
		"│ ├", // C: with passing lane
		"├─╯", // B: convergence
		"├",   // A: root
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
