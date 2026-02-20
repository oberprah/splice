package graph

import (
	"reflect"
	"testing"
)

// TestGenerateRowSymbols_MergeJoin tests that when a column is BOTH converging
// AND receiving a merge parent, it should show ┤ (merge join), not ╯ (just convergence).
//
// This is the case in SequentialMerges at commit D:
// - D is at col 0, merging C which should go into col 1
// - Col 1 was converging (had D from a different branch)
// - The symbol at col 1 should be ┤ (merge line joins into continuing branch)
func TestGenerateRowSymbols_MergeJoin(t *testing.T) {
	// Scenario: commit at col 0, col 1 is both converging AND a merge target
	commitCol := 0
	numCols := 2
	mergeColumns := []int{1}      // C is being merged into col 1
	convergingColumns := []int{1} // col 1 was also converging
	passingColumns := []int{}

	row := generateRowSymbols(commitCol, numCols, mergeColumns, convergingColumns, passingColumns, []int{}, false)

	// Expected: ├─┤ (commit with merge line, merge join symbol)
	expected := []GraphSymbol{SymbolMergeCommit, SymbolMergeJoin}

	if len(row.Symbols) != len(expected) {
		t.Fatalf("Length: got %d, want %d", len(row.Symbols), len(expected))
	}

	for i, exp := range expected {
		if row.Symbols[i] != exp {
			t.Errorf("Symbols[%d]: got %v (%s), want %v (%s)",
				i, row.Symbols[i], row.Symbols[i].String(), exp, exp.String())
		}
	}
}

func TestGenerateRowSymbols(t *testing.T) {
	tests := []struct {
		name              string
		commitCol         int
		numCols           int
		mergeColumns      []int
		convergingColumns []int
		passingColumns    []int
		expected          []GraphSymbol
	}{
		{
			name:              "single commit - linear history",
			commitCol:         0,
			numCols:           1,
			mergeColumns:      nil,
			convergingColumns: nil,
			passingColumns:    nil,
			expected:          []GraphSymbol{SymbolCommit}, // ├
		},
		{
			name:              "merge commit - two parents",
			commitCol:         0,
			numCols:           2,
			mergeColumns:      []int{1},
			convergingColumns: nil,
			passingColumns:    nil,
			expected:          []GraphSymbol{SymbolMergeCommit, SymbolBranchTop}, // ├─╮
		},
		{
			name:              "convergence - branches join",
			commitCol:         0,
			numCols:           2,
			mergeColumns:      nil,
			convergingColumns: []int{1},
			passingColumns:    nil,
			expected:          []GraphSymbol{SymbolMergeCommit, SymbolBranchBottom}, // ├─╯
		},
		{
			name:              "commit with passing lane",
			commitCol:         0,
			numCols:           2,
			mergeColumns:      nil,
			convergingColumns: nil,
			passingColumns:    []int{1},
			expected:          []GraphSymbol{SymbolCommit, SymbolBranchPass}, // ├ │
		},
		{
			name:              "commit on right with passing lane on left",
			commitCol:         1,
			numCols:           2,
			mergeColumns:      nil,
			convergingColumns: nil,
			passingColumns:    []int{0},
			expected:          []GraphSymbol{SymbolBranchPass, SymbolCommit}, // │ ├
		},
		{
			name:              "merge with crossing",
			commitCol:         0,
			numCols:           3,
			mergeColumns:      []int{2},
			convergingColumns: nil,
			passingColumns:    []int{1},
			expected:          []GraphSymbol{SymbolMergeCommit, SymbolBranchCross, SymbolBranchTop}, // ├─│─╮
		},
		{
			name:              "octopus merge - three parents",
			commitCol:         0,
			numCols:           3,
			mergeColumns:      []int{1, 2},
			convergingColumns: nil,
			passingColumns:    nil,
			expected:          []GraphSymbol{SymbolMergeCommit, SymbolOctopus, SymbolBranchTop}, // ├─┬─╮
		},
		{
			name:              "empty row",
			commitCol:         0,
			numCols:           0,
			mergeColumns:      nil,
			convergingColumns: nil,
			passingColumns:    nil,
			expected:          []GraphSymbol{},
		},
		{
			name:              "convergence with passing lane",
			commitCol:         0,
			numCols:           3,
			mergeColumns:      nil,
			convergingColumns: []int{2},
			passingColumns:    []int{1},
			expected:          []GraphSymbol{SymbolMergeCommit, SymbolBranchCross, SymbolBranchBottom}, // ├─│─╯
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateRowSymbols(tt.commitCol, tt.numCols, tt.mergeColumns, tt.convergingColumns, tt.passingColumns, []int{}, false)
			if !reflect.DeepEqual(got.Symbols, tt.expected) {
				// Print rendered versions for easier debugging
				gotStr := RenderRow(got)
				expectedStr := RenderRow(Row{Symbols: tt.expected})
				t.Errorf("generateRowSymbols() = %v (%q), want %v (%q)",
					got.Symbols, gotStr, tt.expected, expectedStr)
			}
		})
	}
}

func TestDetectConvergingColumns(t *testing.T) {
	tests := []struct {
		name       string
		commitCol  int
		commitHash string
		lanes      []string
		expected   []int
	}{
		{
			name:       "no convergence",
			commitCol:  0,
			commitHash: "A",
			lanes:      []string{"A", "B"},
			expected:   nil,
		},
		{
			name:       "one converging column",
			commitCol:  0,
			commitHash: "A",
			lanes:      []string{"A", "A"},
			expected:   []int{1},
		},
		{
			name:       "multiple converging columns",
			commitCol:  0,
			commitHash: "A",
			lanes:      []string{"A", "A", "B", "A"},
			expected:   []int{1, 3},
		},
		{
			name:       "commit in middle column",
			commitCol:  1,
			commitHash: "A",
			lanes:      []string{"A", "A", "A"},
			expected:   []int{0, 2},
		},
		{
			name:       "empty lanes",
			commitCol:  0,
			commitHash: "A",
			lanes:      []string{},
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectConvergingColumns(tt.commitCol, tt.commitHash, tt.lanes)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("detectConvergingColumns() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDetectPassingColumns(t *testing.T) {
	tests := []struct {
		name              string
		commitCol         int
		lanes             []string
		mergeColumns      []int
		convergingColumns []int
		expected          []int
	}{
		{
			name:              "one passing lane",
			commitCol:         0,
			lanes:             []string{"A", "B"},
			mergeColumns:      nil,
			convergingColumns: nil,
			expected:          []int{1},
		},
		{
			name:              "multiple passing lanes",
			commitCol:         1,
			lanes:             []string{"A", "B", "C"},
			mergeColumns:      nil,
			convergingColumns: nil,
			expected:          []int{0, 2},
		},
		{
			name:              "exclude merge column",
			commitCol:         0,
			lanes:             []string{"A", "B", "C"},
			mergeColumns:      []int{1},
			convergingColumns: nil,
			expected:          []int{2},
		},
		{
			name:              "exclude converging column",
			commitCol:         0,
			lanes:             []string{"A", "B", "C"},
			mergeColumns:      nil,
			convergingColumns: []int{2},
			expected:          []int{1},
		},
		{
			name:              "empty lane not included",
			commitCol:         0,
			lanes:             []string{"A", "", "C"},
			mergeColumns:      nil,
			convergingColumns: nil,
			expected:          []int{2},
		},
		{
			name:              "no passing lanes",
			commitCol:         0,
			lanes:             []string{"A"},
			mergeColumns:      nil,
			convergingColumns: nil,
			expected:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectPassingColumns(tt.commitCol, tt.lanes, tt.mergeColumns, tt.convergingColumns)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("detectPassingColumns() = %v, want %v", got, tt.expected)
			}
		})
	}
}
