package graph

// generateRowSymbols generates the graph symbols for a single commit row.
//
// Parameters:
//   - commitCol: which column this commit is in
//   - numCols: total number of columns to generate
//   - mergeColumns: columns where merge parents are located (for merge lines)
//   - convergingColumns: columns that are converging into this commit (same hash in multiple lanes)
//   - passingColumns: columns with lanes that continue through this row (not commit, not converging)
func generateRowSymbols(commitCol int, numCols int, mergeColumns []int, convergingColumns []int, passingColumns []int) Row {
	if numCols == 0 {
		return Row{Symbols: []GraphSymbol{}}
	}

	symbols := make([]GraphSymbol, numCols)

	// Initialize all to empty
	for i := range symbols {
		symbols[i] = SymbolEmpty
	}

	// Find rightmost merge or converging column (for drawing merge line)
	rightmostMerge := -1
	for _, col := range mergeColumns {
		if col > rightmostMerge {
			rightmostMerge = col
		}
	}
	rightmostConverge := -1
	for _, col := range convergingColumns {
		if col > rightmostConverge {
			rightmostConverge = col
		}
	}

	// The rightmost column that needs a horizontal line from commit
	rightmostHorizontal := max(rightmostMerge, rightmostConverge)

	// Build sets for quick lookup
	mergeSet := make(map[int]bool)
	for _, col := range mergeColumns {
		mergeSet[col] = true
	}
	convergeSet := make(map[int]bool)
	for _, col := range convergingColumns {
		convergeSet[col] = true
	}
	passingSet := make(map[int]bool)
	for _, col := range passingColumns {
		passingSet[col] = true
	}

	// Generate symbols for each column
	for col := 0; col < numCols; col++ {
		if col == commitCol {
			// This is the commit column
			if rightmostHorizontal > commitCol {
				// There's a merge line going right
				symbols[col] = SymbolMergeCommit // ├─
			} else {
				symbols[col] = SymbolCommit // ├
			}
		} else if mergeSet[col] {
			// This is a merge parent column (new branch starting from merge)
			symbols[col] = SymbolBranchTop // ╮
		} else if convergeSet[col] {
			// This is a converging column (branch ending, joining commit)
			symbols[col] = SymbolBranchBottom // ╯
		} else if passingSet[col] {
			// This is a passing lane
			if col > commitCol && col < rightmostHorizontal {
				// Merge line crosses this vertical line
				symbols[col] = SymbolBranchCross // │─
			} else {
				symbols[col] = SymbolBranchPass // │
			}
		} else if col > commitCol && col < rightmostHorizontal {
			// Empty column in merge line path - need horizontal line only
			// This shouldn't normally happen, but handle it
			symbols[col] = SymbolEmpty
		}
	}

	return Row{Symbols: symbols}
}

// max returns the larger of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// detectConvergingColumns finds columns that have the same hash as the commit.
// This indicates branches that are converging at this commit (common ancestor).
//
// Parameters:
//   - commitCol: the column where the commit is located
//   - commitHash: the hash of the current commit
//   - lanes: current lane state (before this commit is processed)
//
// Returns indices of lanes (other than commitCol) that have commitHash.
func detectConvergingColumns(commitCol int, commitHash string, lanes []string) []int {
	var converging []int
	for i, hash := range lanes {
		if i != commitCol && hash == commitHash {
			converging = append(converging, i)
		}
	}
	return converging
}

// detectPassingColumns finds columns that have active lanes passing through
// (not the commit column, not merging, not converging).
//
// Parameters:
//   - commitCol: the column where the commit is located
//   - lanes: current lane state
//   - mergeColumns: columns that are merge parents
//   - convergingColumns: columns that are converging
//
// Returns indices of lanes that are just passing through.
func detectPassingColumns(commitCol int, lanes []string, mergeColumns []int, convergingColumns []int) []int {
	// Build exclusion set
	exclude := make(map[int]bool)
	exclude[commitCol] = true
	for _, col := range mergeColumns {
		exclude[col] = true
	}
	for _, col := range convergingColumns {
		exclude[col] = true
	}

	var passing []int
	for i, hash := range lanes {
		if !exclude[i] && hash != "" {
			passing = append(passing, i)
		}
	}
	return passing
}
