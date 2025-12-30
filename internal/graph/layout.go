package graph

// ComputeLayout computes the graph layout for a sequence of commits.
// Commits should be in display order (most recent first).
//
// The algorithm processes each commit and:
// 1. Assigns it to a column (existing lane or new column)
// 2. Generates symbols based on merge lines, convergence, and passing lanes
// 3. Updates lane state with parent information
// 4. Collapses trailing empty lanes
func ComputeLayout(commits []Commit) *Layout {
	if len(commits) == 0 {
		return &Layout{Rows: []Row{}}
	}

	var rows []Row
	var lanes []string

	for _, commit := range commits {
		// 1. Assign column for this commit
		col, newLanes := assignColumn(commit.Hash, lanes)
		lanes = newLanes

		// 2. Detect converging columns (other lanes waiting for this same commit)
		convergingColumns := detectConvergingColumns(col, commit.Hash, lanes)

		// 3. Update lanes with parent information (before symbol generation)
		// This tells us where merge parents will go
		updateResult := updateLanes(col, commit.Parents, lanes)
		lanes = updateResult.Lanes

		// 4. Detect passing columns (lanes that continue through without interaction)
		passingColumns := detectPassingColumns(col, lanes, updateResult.MergeColumns, convergingColumns)

		// 5. Generate symbols for this row
		numCols := len(lanes)
		if numCols == 0 {
			numCols = 1 // At least the commit column
		}
		row := generateRowSymbols(col, numCols, updateResult.MergeColumns, convergingColumns, passingColumns)
		rows = append(rows, row)

		// 6. Collapse trailing empty lanes
		lanes = collapseTrailingEmpty(lanes)
	}

	return &Layout{Rows: rows}
}
