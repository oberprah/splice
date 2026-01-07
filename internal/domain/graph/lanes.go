package graph

// findInLanes searches for a hash in the active lanes.
// Returns the index where the hash was found, or -1 if not found.
func findInLanes(hash string, lanes []string) int {
	for i, h := range lanes {
		if h == hash {
			return i
		}
	}
	return -1
}

// findEmptyLane finds the first empty (empty string) slot in lanes.
// Returns the index of the empty slot, or -1 if no empty slots exist.
func findEmptyLane(lanes []string) int {
	for i, h := range lanes {
		if h == "" {
			return i
		}
	}
	return -1
}

// assignColumn determines which column a commit should occupy.
// It first checks if the commit's hash already exists in a lane (branch continuation),
// then looks for an empty slot, and finally appends a new column if needed.
// Returns the column index and the potentially expanded lanes slice.
func assignColumn(hash string, lanes []string) (int, []string) {
	// First, check if this hash is already in a lane (we're continuing a branch)
	if idx := findInLanes(hash, lanes); idx >= 0 {
		return idx, lanes
	}

	// Look for an empty slot to reuse
	if idx := findEmptyLane(lanes); idx >= 0 {
		lanes[idx] = hash
		return idx, lanes
	}

	// No existing slot, append a new column
	lanes = append(lanes, hash)
	return len(lanes) - 1, lanes
}

// UpdateResult contains the result of updating lanes after processing a commit.
type UpdateResult struct {
	Lanes              []string // Updated lanes state
	MergeColumns       []int    // Columns where merge parents were placed (empty for non-merge commits)
	ExistingLanesMerge []int    // Columns where merge parents already existed in lanes (need extra visual column for corner)
	ConvergesToParent  bool     // True if commit's first parent already exists in another lane (convergence)
}

// updateLanes updates the active lanes after processing a commit.
// It replaces the commit's position with its first parent and adds merge parents.
//
// Parameters:
//   - col: the column where this commit is located
//   - parents: the commit's parent hashes (first parent is primary branch)
//   - lanes: current lane state
//
// Returns UpdateResult with updated lanes and merge parent column positions.
func updateLanes(col int, parents []string, lanes []string) UpdateResult {
	result := UpdateResult{
		Lanes:              lanes,
		MergeColumns:       []int{},
		ExistingLanesMerge: []int{},
		ConvergesToParent:  false,
	}

	if len(parents) == 0 {
		// Root commit - clear this lane
		if col < len(result.Lanes) {
			result.Lanes[col] = ""
		}
		return result
	}

	// First parent replaces commit's position (branch continuation)
	if col < len(result.Lanes) {
		firstParent := parents[0]
		// Check if first parent already exists in another lane
		existingCol := findInLanes(firstParent, result.Lanes)
		if existingCol >= 0 && existingCol != col {
			// First parent exists elsewhere
			// For single-parent commits, decide if this is a convergence (lane ends)
			// or if we should propagate the parent (traditional convergence comes later)

			if len(parents) == 1 {
				// Check if this is forward convergence vs traditional convergence
				// Simulate placing the parent and see if it would be in multiple lanes

				// Count how many lanes would have this parent AFTER placement
				futureParentCount := 0
				for i, hash := range result.Lanes {
					if hash == firstParent {
						futureParentCount++ // Already exists
					} else if i == col {
						futureParentCount++ // We're about to place it here
					}
				}

				if futureParentCount > 1 {
					// Multiple lanes will have parent - traditional convergence
					// Keep parent in lane for multi-lane convergence detection at parent's row
					result.Lanes[col] = firstParent
				} else {
					// Only one lane will have parent - forward convergence (branch ending)
					// Clear lane and show convergence symbol here
					result.Lanes[col] = ""
					result.ConvergesToParent = true
				}
			} else {
				// Merge commit - keep first parent, merge join logic handles it
				result.Lanes[col] = firstParent
			}
		} else {
			result.Lanes[col] = firstParent
		}
	}

	// Additional parents (merge) get placed in available slots or appended
	for i := 1; i < len(parents); i++ {
		mergeParent := parents[i]

		// Check if this parent is already in a lane
		if existingCol := findInLanes(mergeParent, result.Lanes); existingCol >= 0 {
			// Parent already has a lane - this lane continues vertically
			// We need an extra visual column for the merge corner
			result.MergeColumns = append(result.MergeColumns, existingCol)
			result.ExistingLanesMerge = append(result.ExistingLanesMerge, existingCol)
			continue
		}

		// Find an empty slot (prefer slots to the right of commit column)
		placed := false
		for j := col + 1; j < len(result.Lanes); j++ {
			if result.Lanes[j] == "" {
				result.Lanes[j] = mergeParent
				result.MergeColumns = append(result.MergeColumns, j)
				placed = true
				break
			}
		}

		if !placed {
			// No empty slot found, append new column
			result.Lanes = append(result.Lanes, mergeParent)
			result.MergeColumns = append(result.MergeColumns, len(result.Lanes)-1)
		}
	}

	return result
}

// collapseTrailingEmpty removes trailing empty strings from lanes.
// This prevents unbounded width growth when branches complete.
func collapseTrailingEmpty(lanes []string) []string {
	// Find the last non-empty lane
	lastNonEmpty := -1
	for i := len(lanes) - 1; i >= 0; i-- {
		if lanes[i] != "" {
			lastNonEmpty = i
			break
		}
	}

	// Return slice up to and including last non-empty, or empty slice
	if lastNonEmpty < 0 {
		return []string{}
	}
	return lanes[:lastNonEmpty+1]
}
