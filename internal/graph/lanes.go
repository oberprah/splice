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
	Lanes        []string // Updated lanes state
	MergeColumns []int    // Columns where merge parents were placed (empty for non-merge commits)
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
		Lanes:        lanes,
		MergeColumns: []int{},
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
		result.Lanes[col] = parents[0]
	}

	// Additional parents (merge) get placed in available slots or appended
	for i := 1; i < len(parents); i++ {
		mergeParent := parents[i]

		// Check if this parent is already in a lane
		if existingCol := findInLanes(mergeParent, result.Lanes); existingCol >= 0 {
			// Parent already has a lane - record where it is
			result.MergeColumns = append(result.MergeColumns, existingCol)
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
