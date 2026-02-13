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
	var prevMergedHashes map[string]bool // Hashes that were merged in previous row

	for _, commit := range commits {
		// 1. Assign column for this commit
		col, newLanes := assignColumn(commit.Hash, lanes)
		lanes = newLanes

		// 2. Detect converging columns (other lanes waiting for this same commit)
		convergingColumns := detectConvergingColumns(col, commit.Hash, lanes)

		// 3. Clear converging columns so updateLanes can reuse them for merge parents
		// This must happen AFTER detection but BEFORE updateLanes
		for _, convergingCol := range convergingColumns {
			lanes[convergingCol] = ""
		}

		// 4. Check if this commit was merged in previous row (existing lane merge)
		// If so, it should show convergence (closing the merge bracket)
		wasInExistingLaneMerge := prevMergedHashes != nil && prevMergedHashes[commit.Hash]

		// 5. Save lanes state before update (to track merged hashes)
		lanesCopy := make([]string, len(lanes))
		copy(lanesCopy, lanes)

		// 6. Update lanes with parent information
		// Merge parents will naturally fill the cleared converging columns
		updateResult := updateLanes(col, commit.Parents, lanes)
		lanes = updateResult.Lanes

		// Override convergence if this commit was in a previous existing lane merge
		if wasInExistingLaneMerge && len(commit.Parents) == 1 {
			updateResult.ConvergesToParent = true
		}

		// 7. Track which hashes were merged in this row (for next iteration)
		currentMergedHashes := make(map[string]bool)
		for _, mergeCol := range updateResult.ExistingLanesMerge {
			// Get the hash that was in this column before updateLanes
			if mergeCol < len(lanesCopy) {
				mergedHash := lanesCopy[mergeCol]
				if mergedHash != "" {
					currentMergedHashes[mergedHash] = true
				}
			}
		}
		prevMergedHashes = currentMergedHashes

		// 4. Detect passing columns (lanes that continue through without interaction)
		passingColumns := detectPassingColumns(col, lanes, updateResult.MergeColumns, convergingColumns)

		// 5. Generate symbols for this row
		numCols := len(lanes)
		if numCols == 0 {
			numCols = 1 // At least the commit column
		}
		// Add extra column for merge corner when merging to existing lane
		if len(updateResult.ExistingLanesMerge) > 0 {
			numCols++
		}
		// Add extra column for convergence symbol when converging to parent
		if updateResult.ConvergesToParent {
			numCols++
		}
		row := generateRowSymbols(col, numCols, updateResult.MergeColumns, convergingColumns, passingColumns, updateResult.ExistingLanesMerge, updateResult.ConvergesToParent)
		rows = append(rows, row)

		// 6. Collapse trailing empty lanes
		lanes = collapseTrailingEmpty(lanes)
	}

	return &Layout{Rows: rows}
}
